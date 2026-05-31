package milter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"testing"
	"time"

	"github.com/d--j/go-milter/mailfilter"
	"github.com/d--j/go-milter/mailfilter/addr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chainServer creates an httptest.Server that responds to GetDomains (GET /api/v1/domains)
// with ovoo.com and to CreateChain with 201 + the given chain data.
func chainServer(t *testing.T, chain OvooChainData) OvooClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/domains" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(OvooGetDomainsResponse{
				Domains: []OvooDomainData{{Id: "1", Name: "ovoo.com"}},
			})
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(chain)
	}))
	t.Cleanup(srv.Close)
	cli, err := NewClient(srv.URL, "test-token", false, 5*time.Second, "Mail Display Name")
	require.NoError(t, err)
	return cli
}

// errorServer creates an httptest.Server that responds 200 to GetDomains and 500 to everything else.
func errorServer(t *testing.T) OvooClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/domains" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(OvooGetDomainsResponse{
				Domains: []OvooDomainData{{Id: "1", Name: "ovoo.com"}},
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"status":"error","detail":"internal error"}]}`))
	}))
	t.Cleanup(srv.Close)
	cli, err := NewClient(srv.URL, "test-token", false, 5*time.Second, "Mail Display Name")
	require.NoError(t, err)
	return cli
}

// domainsServer creates an httptest.Server that returns a fixed domain list (ovoo.com) for GetDomains.
func domainsServer(t *testing.T) OvooClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(OvooGetDomainsResponse{
			Domains: []OvooDomainData{{Id: "1", Name: "ovoo.com"}},
		})
	}))
	t.Cleanup(srv.Close)
	cli, err := NewClient(srv.URL, "test-token", false, 5*time.Second, "Mail Display Name")
	require.NoError(t, err)
	return cli
}

// stubClient returns an OvooClient with no HTTP transport, suitable for tests that never
// reach any HTTP method (e.g. fail at getHeaderAddr before GetDomains).
func stubClient() OvooClient {
	return OvooClient{}
}

// --- getHeaderAddr ---

func TestGetHeaderAddr_Valid(t *testing.T) {
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com")
	got, err := getHeaderAddr("from", trx)
	require.NoError(t, err)
	assert.Equal(t, "Sender", got.Name)
	assert.Equal(t, "sender@ext.com", got.Address)
}

func TestGetHeaderAddr_Missing(t *testing.T) {
	// Text("from") returns ("", nil) when header is absent → ParseAddress("") fails.
	trx := newMockTrx("", "sender@ext.com")
	_, err := getHeaderAddr("from", trx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from")
}

func TestGetHeaderAddr_ParseError(t *testing.T) {
	trx := newMockTrx("not-an-address!!!", "sender@ext.com")
	_, err := getHeaderAddr("from", trx)
	assert.Error(t, err)
}

// --- AddressRewriter: early exit paths ---

// The From header is parsed before the recipient loop, so a bad From header rejects
// even messages with no Ovoo alias recipients.
func TestAddressRewriter_InvalidFromHeader_WithOvooRcpt(t *testing.T) {
	cli := stubClient()
	trx := newMockTrx("", "sender@ext.com", addr.NewRcptTo("alias@ovoo.com", "", ""))

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	assert.True(t, mailfilter.Reject.Equal(decision))
	assert.Error(t, err)
}

func TestAddressRewriter_InvalidFromHeader_NoOvooRcpt(t *testing.T) {
	// Malformed From is detected before the no-match early-Accept, so the message is
	// still rejected even when no Ovoo alias is involved.
	cli := stubClient()
	trx := newMockTrx("", "sender@ext.com", addr.NewRcptTo("user@external.com", "", ""))

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	assert.True(t, mailfilter.Reject.Equal(decision))
	assert.Error(t, err)
}

func TestAddressRewriter_EmptyRcptList(t *testing.T) {
	cli := domainsServer(t)
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com") // no recipients

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	assert.True(t, mailfilter.Accept.Equal(decision))
	assert.NoError(t, err)
	assert.Empty(t, trx.changeMailFromCalls)
}

func TestAddressRewriter_NoMatchingRecipients(t *testing.T) {
	cli := domainsServer(t)
	rcpt := addr.NewRcptTo("user@external.com", "", "")
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com", rcpt)

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	assert.True(t, mailfilter.Accept.Equal(decision))
	assert.NoError(t, err)
	assert.Empty(t, trx.changeMailFromCalls)
	assert.Empty(t, trx.delRcptToCalls)
}

// --- AddressRewriter: too many recipients ---

func TestAddressRewriter_MultipleMatchingRecipients(t *testing.T) {
	cli := domainsServer(t)
	rcpt1 := addr.NewRcptTo("alias1@ovoo.com", "", "")
	rcpt2 := addr.NewRcptTo("alias2@ovoo.com", "", "")
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com", rcpt1, rcpt2)

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	expected := mailfilter.CustomErrorResponse(522, "5.5.3 Too many recipients")
	assert.True(t, expected.Equal(decision))
	assert.Error(t, err)
	assert.Empty(t, trx.changeMailFromCalls)
}

// --- AddressRewriter: CreateChain failure ---

func TestAddressRewriter_CreateChainError(t *testing.T) {
	cli := errorServer(t)
	rcpt := addr.NewRcptTo("alias@ovoo.com", "", "")
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com", rcpt)

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	assert.True(t, mailfilter.Reject.Equal(decision))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating chain")
}

// --- AddressRewriter: forward chain (OrigToAddress.Type != "reply_alias") ---

func TestAddressRewriter_ForwardChain(t *testing.T) {
	chain := OvooChainData{
		FromEmail:     "reply@ovoo.com",
		ToEmail:       "user@gmail.com",
		OrigToAddress: OvooChainAddressData{Email: "alias@ovoo.com", Type: "alias"},
	}
	cli := chainServer(t, chain)

	rcpt := addr.NewRcptTo("alias@ovoo.com", "SIZE=1000", "")
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com", rcpt)

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	require.NoError(t, err)
	assert.True(t, mailfilter.Accept.Equal(decision))

	// Envelope: original alias removed, chain target added, From rewritten.
	require.Len(t, trx.changeMailFromCalls, 1)
	assert.Equal(t, "reply@ovoo.com", trx.changeMailFromCalls[0].from)
	assert.Equal(t, "", trx.changeMailFromCalls[0].args) // MailFrom.Args passed through
	assert.Equal(t, []string{"alias@ovoo.com"}, trx.delRcptToCalls)
	require.Len(t, trx.addRcptToCalls, 1)
	assert.Equal(t, "user@gmail.com", trx.addRcptToCalls[0].rcptTo)
	assert.Equal(t, "SIZE=1000", trx.addRcptToCalls[0].args) // rcpt.Args passed through

	// Headers: From preserves sender display name; To is the original alias address.
	expectedFrom := (&mail.Address{Name: "Sender", Address: "reply@ovoo.com"}).String()
	expectedTo := (&mail.Address{Name: "Ovoo Hidden Mail", Address: "alias@ovoo.com"}).String()
	assert.True(t, trx.headers.hasSet("from", expectedFrom))
	assert.True(t, trx.headers.hasSet("to", expectedTo))
	assert.True(t, trx.headers.hasSet("dkim-signature", ""))
	assert.True(t, trx.headers.hasSet("x-google-dkim-signature", ""))
}

// When the From header has no display name, the rewritten From header is a bare address.
func TestAddressRewriter_ForwardChain_SenderNoName(t *testing.T) {
	chain := OvooChainData{
		FromEmail:     "reply@ovoo.com",
		ToEmail:       "user@gmail.com",
		OrigToAddress: OvooChainAddressData{Type: "alias"},
	}
	cli := chainServer(t, chain)

	rcpt := addr.NewRcptTo("alias@ovoo.com", "", "")
	trx := newMockTrx("sender@ext.com", "sender@ext.com", rcpt) // bare address, no name

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	require.NoError(t, err)
	assert.True(t, mailfilter.Accept.Equal(decision))

	expectedFrom := (&mail.Address{Name: "", Address: "reply@ovoo.com"}).String()
	assert.True(t, trx.headers.hasSet("from", expectedFrom))
}

// MailFrom ESMTP args must be passed through to ChangeMailFrom unchanged.
func TestAddressRewriter_MailFromArgsPassthrough(t *testing.T) {
	chain := OvooChainData{
		FromEmail:     "reply@ovoo.com",
		ToEmail:       "user@gmail.com",
		OrigToAddress: OvooChainAddressData{Type: "alias"},
	}
	cli := chainServer(t, chain)

	rcpt := addr.NewRcptTo("alias@ovoo.com", "", "")
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com", rcpt)
	mf := addr.NewMailFrom("sender@ext.com", "SIZE=500 BODY=8BITMIME", "", "", "")
	trx.mailFrom = &mf

	_, err := AddressRewriter(cli)(context.Background(), trx)
	require.NoError(t, err)

	require.Len(t, trx.changeMailFromCalls, 1)
	assert.Equal(t, "SIZE=500 BODY=8BITMIME", trx.changeMailFromCalls[0].args)
}

// --- AddressRewriter: reply chain (OrigToAddress.Type == "reply_alias") ---

func TestAddressRewriter_ReplyChain(t *testing.T) {
	chain := OvooChainData{
		FromEmail:     "alias@ovoo.com",
		ToEmail:       "ext1@external.com",
		OrigToAddress: OvooChainAddressData{Email: "reply-alias@ovoo.com", Type: "reply_alias"},
	}
	cli := chainServer(t, chain)

	rcpt := addr.NewRcptTo("reply-alias@ovoo.com", "", "")
	trx := newMockTrx("User <user@gmail.com>", "user@gmail.com", rcpt)

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	require.NoError(t, err)
	assert.True(t, mailfilter.Accept.Equal(decision))

	// Reply chain: both From and To headers have no display name.
	expectedFrom := (&mail.Address{Address: "alias@ovoo.com"}).String()
	expectedTo := (&mail.Address{Address: "ext1@external.com"}).String()
	assert.True(t, trx.headers.hasSet("from", expectedFrom))
	assert.True(t, trx.headers.hasSet("to", expectedTo))
	assert.True(t, trx.headers.hasSet("dkim-signature", ""))
	assert.True(t, trx.headers.hasSet("x-google-dkim-signature", ""))
}

// --- AddressRewriter: reply-to header handling ---

// When reply-to is present and non-empty it must be overwritten with the new From value.
func TestAddressRewriter_ReplyToPresent(t *testing.T) {
	chain := OvooChainData{
		FromEmail:     "reply@ovoo.com",
		ToEmail:       "user@gmail.com",
		OrigToAddress: OvooChainAddressData{Type: "alias"},
	}
	cli := chainServer(t, chain)

	rcpt := addr.NewRcptTo("alias@ovoo.com", "", "")
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com", rcpt)
	trx.headers.textValues["reply-to"] = "Sender <sender@ext.com>"

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	require.NoError(t, err)
	assert.True(t, mailfilter.Accept.Equal(decision))

	expectedFrom := (&mail.Address{Name: "Sender", Address: "reply@ovoo.com"}).String()
	assert.True(t, trx.headers.hasSet("reply-to", expectedFrom), "reply-to must be overwritten when present")
}

// When reply-to is absent (Text returns "", nil) it must not be touched.
func TestAddressRewriter_ReplyToAbsent(t *testing.T) {
	chain := OvooChainData{
		FromEmail:     "reply@ovoo.com",
		ToEmail:       "user@gmail.com",
		OrigToAddress: OvooChainAddressData{Type: "alias"},
	}
	cli := chainServer(t, chain)

	rcpt := addr.NewRcptTo("alias@ovoo.com", "", "")
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com", rcpt)
	// textValues["reply-to"] not set → Text("reply-to") returns ("", nil)

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	require.NoError(t, err)
	assert.True(t, mailfilter.Accept.Equal(decision))

	for _, c := range trx.headers.setCalls {
		assert.NotEqual(t, "reply-to", c.key, "reply-to must not be touched when absent")
	}
}

// When Text("reply-to") returns an error the reply-to header must not be touched.
// This validates the bug fix: the old condition (err != nil || len != 0) would have
// triggered Set on a parse error, incorrectly overwriting the header.
func TestAddressRewriter_ReplyToParseError(t *testing.T) {
	chain := OvooChainData{
		FromEmail:     "reply@ovoo.com",
		ToEmail:       "user@gmail.com",
		OrigToAddress: OvooChainAddressData{Type: "alias"},
	}
	cli := chainServer(t, chain)

	rcpt := addr.NewRcptTo("alias@ovoo.com", "", "")
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com", rcpt)
	trx.headers.textErrors["reply-to"] = errors.New("charset decode error")

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	require.NoError(t, err)
	assert.True(t, mailfilter.Accept.Equal(decision))

	for _, c := range trx.headers.setCalls {
		assert.NotEqual(t, "reply-to", c.key, "reply-to must not be touched when Text returns an error")
	}
}

// --- AddressRewriter: mixed recipient lists ---

// One Ovoo alias and one external recipient: only the alias is processed; the external
// recipient is neither deleted nor re-added.
func TestAddressRewriter_MixedRecipients_OneDomainMatch(t *testing.T) {
	chain := OvooChainData{
		FromEmail:     "reply@ovoo.com",
		ToEmail:       "user@gmail.com",
		OrigToAddress: OvooChainAddressData{Type: "alias"},
	}
	cli := chainServer(t, chain)

	ovooRcpt := addr.NewRcptTo("alias@ovoo.com", "", "")
	extRcpt := addr.NewRcptTo("other@external.com", "", "")
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com", ovooRcpt, extRcpt)

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	require.NoError(t, err)
	assert.True(t, mailfilter.Accept.Equal(decision))

	// Only the Ovoo alias is deleted and replaced; external is untouched.
	assert.Equal(t, []string{"alias@ovoo.com"}, trx.delRcptToCalls)
	require.Len(t, trx.addRcptToCalls, 1)
	assert.Equal(t, "user@gmail.com", trx.addRcptToCalls[0].rcptTo)
}

// Two Ovoo aliases alongside an external recipient still triggers the multi-recipient rejection.
func TestAddressRewriter_MixedRecipients_TwoDomainMatches(t *testing.T) {
	cli := domainsServer(t)
	rcpt1 := addr.NewRcptTo("alias1@ovoo.com", "", "")
	rcpt2 := addr.NewRcptTo("alias2@ovoo.com", "", "")
	ext := addr.NewRcptTo("user@external.com", "", "")
	trx := newMockTrx("Sender <sender@ext.com>", "sender@ext.com", rcpt1, ext, rcpt2)

	decision, err := AddressRewriter(cli)(context.Background(), trx)

	expected := mailfilter.CustomErrorResponse(522, "5.5.3 Too many recipients")
	assert.True(t, expected.Equal(decision))
	assert.Error(t, err)
}
