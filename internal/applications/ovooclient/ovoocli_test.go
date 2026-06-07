package ovooclient

import (
	"context"
	"errors"
	"fmt"
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
func ovooCLIWith(rt http.RoundTripper) Client {
	return Client{
		client: &http.Client{Transport: rt},
		server: "http://example.com",
		token:  "test-token",
	}
}

// --- NewClient ---

func TestNewClient_ValidParams(t *testing.T) {
	cli, err := NewClient("http://localhost", "mytoken", false, 3*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost", cli.server)
	assert.Equal(t, "mytoken", cli.token)
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

// --- helpers for domain tests ---

// clearDomainCache resets the package-level cache between tests.
func clearDomainCache() {
	domainCache.Range(func(k, _ any) bool {
		domainCache.Delete(k)
		return true
	})
}

// domainsBody builds a GetDomainsResponse JSON string.
func domainsBody(names []string, currentPage, lastPage int) string {
	parts := make([]string, 0, len(names))
	for _, n := range names {
		parts = append(parts, fmt.Sprintf(`{"id":"id-%s","name":%q}`, n, n))
	}
	return fmt.Sprintf(
		`{"domains":[%s],"pagination_metadata":{"current_page":%d,"first_page":1,"last_page":%d,"page_size":1,"total_records":%d}}`,
		strings.Join(parts, ","), currentPage, lastPage, len(names),
	)
}

// --- getDomainsNetwork ---

func TestGetDomainsNetwork_SinglePage(t *testing.T) {
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{"example.com"}, 1, 1))),
		}, nil
	}))
	domains, err := cli.getDomainsNetwork(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, []string{"example.com"}, domains)
}

func TestGetDomainsNetwork_MultiPage(t *testing.T) {
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		var body string
		switch r.URL.Query().Get("page") {
		case "1":
			body = domainsBody([]string{"a.com"}, 1, 3)
		case "2":
			body = domainsBody([]string{"b.com"}, 2, 3)
		default:
			body = domainsBody([]string{"c.com"}, 3, 3)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	}))
	domains, err := cli.getDomainsNetwork(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, []string{"a.com", "b.com", "c.com"}, domains)
}

func TestGetDomainsNetwork_EmptyFirstPage(t *testing.T) {
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{}, 1, 1))),
		}, nil
	}))
	domains, err := cli.getDomainsNetwork(context.Background(), "")
	require.NoError(t, err)
	assert.Empty(t, domains)
}

func TestGetDomainsNetwork_NetworkError(t *testing.T) {
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	}))
	domains, err := cli.getDomainsNetwork(context.Background(), "")
	assert.Nil(t, domains)
	assert.Error(t, err)
}

func TestGetDomainsNetwork_HTTPError(t *testing.T) {
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(strings.NewReader(`{"errors":[{"status":"forbidden","detail":"token invalid"}]}`)),
		}, nil
	}))
	domains, err := cli.getDomainsNetwork(context.Background(), "")
	assert.Nil(t, domains)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ovoo api errors")
}

func TestGetDomainsNetwork_InvalidJSON(t *testing.T) {
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("not-json{")),
		}, nil
	}))
	domains, err := cli.getDomainsNetwork(context.Background(), "")
	assert.Nil(t, domains)
	assert.Error(t, err)
}

func TestGetDomainsNetwork_BodyClosed(t *testing.T) {
	spy := &closeSpy{Reader: strings.NewReader(domainsBody([]string{"example.com"}, 1, 1))}
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: spy}, nil
	}))
	_, err := cli.getDomainsNetwork(context.Background(), "")
	require.NoError(t, err)
	assert.True(t, spy.closed, "response body must be closed")
}

func TestGetDomainsNetwork_QueryParams_Fixed(t *testing.T) {
	var capturedReq *http.Request
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		capturedReq = r
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{"example.com"}, 1, 1))),
		}, nil
	}))
	_, err := cli.getDomainsNetwork(context.Background(), "")
	require.NoError(t, err)
	q := capturedReq.URL.Query()
	assert.Equal(t, "true", q.Get("active"))
	assert.Equal(t, "true", q.Get("verified"))
	assert.Equal(t, "1", q.Get("page_size"))
	assert.Equal(t, "1", q.Get("page"))
}

func TestGetDomainsNetwork_WithDomainNameFilter(t *testing.T) {
	var capturedReq *http.Request
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		capturedReq = r
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{"acme.io"}, 1, 1))),
		}, nil
	}))
	domains, err := cli.getDomainsNetwork(context.Background(), "acme.io")
	require.NoError(t, err)
	assert.Equal(t, []string{"acme.io"}, domains)
	assert.Equal(t, "acme.io", capturedReq.URL.Query().Get("domain_name"))
}

func TestGetDomainsNetwork_NoDomainNameFilter(t *testing.T) {
	var capturedReq *http.Request
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		capturedReq = r
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{"example.com"}, 1, 1))),
		}, nil
	}))
	_, err := cli.getDomainsNetwork(context.Background(), "")
	require.NoError(t, err)
	assert.Empty(t, capturedReq.URL.Query().Get("domain_name"))
}

// --- GetDomains ---

func TestGetDomains_CacheHit(t *testing.T) {
	clearDomainCache()
	domainCache.Store("domains", cachedActiveDomainsInfo{
		domains:   []string{"cached.com"},
		expiresAt: time.Now().Add(5 * time.Minute),
	})
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		t.Fatal("network should not be called on cache hit")
		return nil, nil
	}))
	domains, err := cli.GetDomains(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"cached.com"}, domains)
}

func TestGetDomains_CacheExpired(t *testing.T) {
	clearDomainCache()
	domainCache.Store("domains", cachedActiveDomainsInfo{
		domains:   []string{"stale.com"},
		expiresAt: time.Now().Add(-time.Second),
	})
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{"fresh.com"}, 1, 1))),
		}, nil
	}))
	domains, err := cli.GetDomains(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"fresh.com"}, domains)
	val, ok := domainCache.Load("domains")
	require.True(t, ok)
	entry := val.(cachedActiveDomainsInfo)
	assert.Equal(t, []string{"fresh.com"}, entry.domains)
	assert.True(t, entry.expiresAt.After(time.Now()))
}

func TestGetDomains_CacheMiss(t *testing.T) {
	clearDomainCache()
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{"new.com"}, 1, 1))),
		}, nil
	}))
	domains, err := cli.GetDomains(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"new.com"}, domains)
	val, ok := domainCache.Load("domains")
	require.True(t, ok)
	entry := val.(cachedActiveDomainsInfo)
	assert.Equal(t, []string{"new.com"}, entry.domains)
	assert.True(t, entry.expiresAt.After(time.Now()))
}

func TestGetDomains_NetworkError(t *testing.T) {
	clearDomainCache()
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("timeout")
	}))
	result, err := cli.GetDomains(context.Background())
	assert.Nil(t, result)
	assert.Error(t, err)
	_, ok := domainCache.Load("domains")
	assert.False(t, ok)
}

func TestGetDomains_NoDomainNameParam(t *testing.T) {
	clearDomainCache()
	var capturedReq *http.Request
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		capturedReq = r
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{"example.com"}, 1, 1))),
		}, nil
	}))
	_, err := cli.GetDomains(context.Background())
	require.NoError(t, err)
	assert.Empty(t, capturedReq.URL.Query().Get("domain_name"))
}

// --- GetDomainByName ---

func TestGetDomainByName_EmptyString(t *testing.T) {
	clearDomainCache()
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		t.Fatal("network should not be called for empty domain name")
		return nil, nil
	}))
	assert.False(t, cli.GetDomainByName(context.Background(), ""))
}

func TestGetDomainByName_WhitespaceOnly(t *testing.T) {
	clearDomainCache()
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		t.Fatal("network should not be called for whitespace-only domain name")
		return nil, nil
	}))
	assert.False(t, cli.GetDomainByName(context.Background(), "  \t\n"))
}

func TestGetDomainByName_CacheHit(t *testing.T) {
	clearDomainCache()
	domainCache.Store("domain_name:hit.com", cachedDomainNameInfo{
		domain_name: "hit.com",
		expiresAt:   time.Now().Add(5 * time.Minute),
	})
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		t.Fatal("network should not be called on cache hit")
		return nil, nil
	}))
	assert.True(t, cli.GetDomainByName(context.Background(), "hit.com"))
}

func TestGetDomainByName_CacheExpired(t *testing.T) {
	clearDomainCache()
	domainCache.Store("domain_name:old.com", cachedDomainNameInfo{
		domain_name: "old.com",
		expiresAt:   time.Now().Add(-time.Second),
	})
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{"old.com"}, 1, 1))),
		}, nil
	}))
	assert.True(t, cli.GetDomainByName(context.Background(), "old.com"))
	val, ok := domainCache.Load("domain_name:old.com")
	require.True(t, ok)
	entry := val.(cachedDomainNameInfo)
	assert.True(t, entry.expiresAt.After(time.Now()))
}

func TestGetDomainByName_FoundViaNetwork(t *testing.T) {
	clearDomainCache()
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{"found.com"}, 1, 1))),
		}, nil
	}))
	assert.True(t, cli.GetDomainByName(context.Background(), "found.com"))
	val, ok := domainCache.Load("domain_name:found.com")
	require.True(t, ok)
	entry := val.(cachedDomainNameInfo)
	assert.Equal(t, "found.com", entry.domain_name)
	assert.True(t, entry.expiresAt.After(time.Now()))
}

func TestGetDomainByName_NotFoundViaNetwork(t *testing.T) {
	clearDomainCache()
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{}, 1, 1))),
		}, nil
	}))
	assert.False(t, cli.GetDomainByName(context.Background(), "missing.com"))
	_, ok := domainCache.Load("domain_name:missing.com")
	assert.False(t, ok)
}

func TestGetDomainByName_NetworkError(t *testing.T) {
	clearDomainCache()
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("refused")
	}))
	assert.False(t, cli.GetDomainByName(context.Background(), "err.com"))
	_, ok := domainCache.Load("domain_name:err.com")
	assert.False(t, ok)
}

func TestGetDomainByName_DomainNameSentAsQueryParam(t *testing.T) {
	clearDomainCache()
	var capturedReq *http.Request
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		capturedReq = r
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{"q.com"}, 1, 1))),
		}, nil
	}))
	assert.True(t, cli.GetDomainByName(context.Background(), "q.com"))
	assert.Equal(t, "q.com", capturedReq.URL.Query().Get("domain_name"))
}

func TestGetDomainByName_TrimsWhitespaceBeforeLookup(t *testing.T) {
	clearDomainCache()
	var capturedReq *http.Request
	cli := ovooCLIWith(roundTripFn(func(r *http.Request) (*http.Response, error) {
		capturedReq = r
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(domainsBody([]string{"trim.com"}, 1, 1))),
		}, nil
	}))
	assert.True(t, cli.GetDomainByName(context.Background(), "  trim.com  "))
	assert.Equal(t, "trim.com", capturedReq.URL.Query().Get("domain_name"))
	_, ok := domainCache.Load("domain_name:trim.com")
	assert.True(t, ok)
	_, ok2 := domainCache.Load("domain_name:  trim.com  ")
	assert.False(t, ok2)
}
