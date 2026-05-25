package milter

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// roundTripFn is a function-based http.RoundTripper for injecting controlled responses.
type roundTripFn func(*http.Request) (*http.Response, error)

func (f roundTripFn) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// closeSpy wraps a Reader and records whether Close() was called.
type closeSpy struct {
	io.Reader
	closed bool
}

func (c *closeSpy) Close() error { c.closed = true; return nil }

// ovooCLIWith returns an OvooClient that uses the provided RoundTripper.
func ovooCLIWith(rt http.RoundTripper) OvooClient {
	return OvooClient{
		client:  &http.Client{Transport: rt},
		server:  "http://example.com",
		token:   "test-token",
		domains: []string{"ovoo.com"},
	}
}

// --- NewClient ---

func TestNewClient_EmptyDomain(t *testing.T) {
	_, err := NewClient("http://localhost", "tok", false, []string{}, 5*time.Second, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "domain")
}

func TestNewClient_ValidParams(t *testing.T) {
	cli, err := NewClient("http://localhost", "mytoken", false, []string{"ovoo.com"}, 3*time.Second, "Ovoo Mail")
	require.NoError(t, err)
	assert.Equal(t, "http://localhost", cli.server)
	assert.Equal(t, "mytoken", cli.token)
	assert.Equal(t, []string{"ovoo.com"}, cli.domains)
	assert.Equal(t, 3*time.Second, cli.client.Timeout)
}

// --- createRequest ---

func TestCreateRequest_QueryParams(t *testing.T) {
	cli := ovooCLIWith(nil)
	req, err := cli.createRequest(context.Background(), "http://example.com", "/path",
		http.MethodGet, nil, nil, map[string]string{"foo": "bar"})
	require.NoError(t, err)
	assert.Equal(t, "foo=bar", req.URL.RawQuery)
}

func TestCreateRequest_MultipleQueryParams(t *testing.T) {
	cli := ovooCLIWith(nil)
	req, err := cli.createRequest(context.Background(), "http://example.com", "/path",
		http.MethodGet, nil, nil, map[string]string{"a": "1", "b": "2"})
	require.NoError(t, err)
	// url.Values.Encode sorts keys alphabetically, so order is deterministic.
	assert.Equal(t, "a=1&b=2", req.URL.RawQuery)
}

func TestCreateRequest_NoQueryParams(t *testing.T) {
	cli := ovooCLIWith(nil)
	req, err := cli.createRequest(context.Background(), "http://example.com", "/path",
		http.MethodGet, nil, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, req.URL.RawQuery)
}

func TestCreateRequest_Headers(t *testing.T) {
	cli := ovooCLIWith(nil)
	req, err := cli.createRequest(context.Background(), "http://example.com", "/path",
		http.MethodPost, nil, map[string]string{"X-Custom": "value"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "value", req.Header.Get("X-Custom"))
}

// --- CreateChain ---

// When the HTTP client itself fails (nil response), CreateChain must return an error
// without panicking on a nil response pointer.
func TestCreateChain_NetworkError(t *testing.T) {
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	}))

	result, err := cli.CreateChain(context.Background(), "sender@ext.com", "alias@ovoo.com")
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestCreateChain_Success(t *testing.T) {
	body := `{"hash":"h1","from_email":"reply@ovoo.com","to_email":"user@gmail.com",` +
		`"orig_from_address":{"email":"sender@ext.com","type":"external"},` +
		`"orig_to_address":{"email":"alias@ovoo.com","type":"alias"}}`

	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	}))

	result, err := cli.CreateChain(context.Background(), "sender@ext.com", "alias@ovoo.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "h1", result.Hash)
	assert.Equal(t, "reply@ovoo.com", result.FromEmail)
	assert.Equal(t, "user@gmail.com", result.ToEmail)
	assert.Equal(t, "alias", result.OrigToAddress.Type)
}

// Response body must be closed even on a successful 201 response.
func TestCreateChain_Success_BodyClosed(t *testing.T) {
	spy := &closeSpy{Reader: strings.NewReader(
		`{"hash":"h","from_email":"f@o.com","to_email":"t@g.com",` +
			`"orig_from_address":{"email":"","type":""},"orig_to_address":{"email":"","type":""}}`)}

	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusCreated, Body: spy}, nil
	}))

	_, err := cli.CreateChain(context.Background(), "a@b.com", "c@ovoo.com")
	require.NoError(t, err)
	assert.True(t, spy.closed, "response body must be closed on success")
}

// Response body must be closed even when the server returns a non-201 status.
func TestCreateChain_ErrorResponse_BodyClosed(t *testing.T) {
	spy := &closeSpy{Reader: strings.NewReader(`{"Error":[{"status":"error","detail":"not found"}]}`)}

	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNotFound, Body: spy}, nil
	}))

	result, err := cli.CreateChain(context.Background(), "a@b.com", "c@ovoo.com")
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, spy.closed, "response body must be closed on error response")
}

func TestCreateChain_InvalidJSON(t *testing.T) {
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader("not-json{")),
		}, nil
	}))

	result, err := cli.CreateChain(context.Background(), "a@b.com", "c@ovoo.com")
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestCreateChain_AuthorizationHeader(t *testing.T) {
	var gotAuth string
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body: io.NopCloser(strings.NewReader(
				`{"hash":"","from_email":"","to_email":"",` +
					`"orig_from_address":{"email":"","type":""},` +
					`"orig_to_address":{"email":"","type":""}}`)),
		}, nil
	}))
	cli.token = "my-secret-token"

	_, _ = cli.CreateChain(context.Background(), "a@b.com", "c@ovoo.com")
	assert.Equal(t, "Bearer my-secret-token", gotAuth)
}

func TestCreateChain_ContentTypeHeader(t *testing.T) {
	var gotCT string
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		gotCT = r.Header.Get("Content-Type")
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body: io.NopCloser(strings.NewReader(
				`{"hash":"","from_email":"","to_email":"",` +
					`"orig_from_address":{"email":"","type":""},` +
					`"orig_to_address":{"email":"","type":""}}`)),
		}, nil
	}))

	_, _ = cli.CreateChain(context.Background(), "a@b.com", "c@ovoo.com")
	assert.Equal(t, "application/json", gotCT)
}
