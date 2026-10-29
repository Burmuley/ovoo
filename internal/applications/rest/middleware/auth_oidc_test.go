package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestOIDCProvider starts a minimal OIDC server whose UserInfo returns the given claims.
func newTestOIDCProvider(t *testing.T, claims map[string]any) OIDCProvider {
	t.Helper()

	var srv *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                 srv.URL,
			"authorization_endpoint": srv.URL + "/auth",
			"token_endpoint":         srv.URL + "/token",
			"jwks_uri":               srv.URL + "/keys",
			"userinfo_endpoint":      srv.URL + "/userinfo",
		})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(claims)
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	p, err := oidc.NewProvider(context.Background(), srv.URL)
	require.NoError(t, err)
	return OIDCProvider{OIDCProvider: p, Issuer: srv.URL}
}

func TestValidateAccessTokenViaUserInfo_EmailVerified(t *testing.T) {
	tests := []struct {
		name    string
		claims  map[string]any
		allow   bool
		wantErr bool
	}{
		{"verified email", map[string]any{"sub": "1", "email": "u@example.com", "email_verified": true}, false, false},
		{"unverified email rejected", map[string]any{"sub": "1", "email": "u@example.com", "email_verified": false}, false, true},
		{"missing claim rejected", map[string]any{"sub": "1", "email": "u@example.com"}, false, true},
		{"unverified email allowed by provider option", map[string]any{"sub": "1", "email": "u@example.com", "email_verified": false}, true, false},
		{"missing claim allowed by provider option", map[string]any{"sub": "1", "email": "u@example.com"}, true, false},
		{"no email always rejected", map[string]any{"sub": "1", "email_verified": true}, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prov := newTestOIDCProvider(t, tt.claims)
			prov.AllowUnverifiedEmail = tt.allow

			// unique token per case: successful lookups are cached by access token
			email, err := validateAccessTokenViaUserInfo(context.Background(), "token-"+tt.name, prov)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "u@example.com", email)
		})
	}
}
