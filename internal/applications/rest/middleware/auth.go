package middleware

import (
	"context"
	"net/http"
	"slices"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
)

const (
	authorizationHeader = "Authorization"
	RootPageURI         = "/"

	// cookies names
	stateCookieName = "ovoo_state"
	nonceCookieName = "ovoo_nonce"
	authCookieName  = "ovoo_auth"

	// ApiKey constants
	apiTokenCookieName = "ovoo_key"
)

// var providerConfig *OIDCProvider
var oidcConfigs map[string]OIDCProvider

type UserContextKey string

// Authentication creates a middleware adapter for handling user authentication.
// It supports multiple authentication methods:
// - OIDC/OAuth2 via cookies (browser sessions) with multiple providers through URL-based registration
// - Basic authentication (username/password)
// - Bearer token authentication (OIDC/OAuth2)
// - API token authentication via cookie or header
//
// Parameters:
//   - skipUris: a slice of URI paths that bypass authentication entirely
//   - svcGw: service gateway providing user/token validation capabilities
//
// The middleware adds authenticated user information to the request context
// with the UserContextKey("user") key when authentication succeeds.
//
// Authentication flow:
// 1. Skip authentication for whitelisted URIs in skipUris
// 2. Check for provider-specific OIDC login/callback URIs using regex matching
// 3. Try Basic authentication header (username/password combo)
// 4. Try API token from cookie or header
// 5. Try Bearer token authentication (OIDC/OAuth2) if providers configured
// 6. Allow unauthenticated access to root page "/"
// 7. Return 401 Unauthorized for all other cases
func Authentication(skipUris []string, svcGw *services.ServiceGateway) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for whitelisted URIs
			if slices.Contains(skipUris, r.URL.Path) {
				h.ServeHTTP(w, r)
				return
			}

			if match := oidcLoginUriReg.FindStringSubmatch(r.URL.Path); match != nil {
				provider, ok := oidcConfigs[match[1]]
				if !ok {
					logger.Error("unknown OIDC provider login URI", "src", r.RemoteAddr, "value", match[1])
					http.Error(w, "unknown OIDC provider login URI", http.StatusBadRequest)
					return
				}

				handleOIDCLogin(w, r, provider)
				return
			}

			if match := oidcCallbackUriReg.FindStringSubmatch(r.URL.Path); match != nil {
				provider, ok := oidcConfigs[match[1]]
				if !ok {
					logger.Error("unknown OIDC provider callback URI", "src", r.RemoteAddr, "value", match[1])
					http.Error(w, "unknown OIDC provider callback URI", http.StatusBadRequest)
					return
				}

				handleOIDCCallback(w, r, provider)
				return
			}

			// Process Basic authentication (username/password)
			username, password, ok := r.BasicAuth()
			if ok {
				user, err := validateBasicAuth(r.Context(), username, password, svcGw)
				if err != nil {
					logger.Error("invalid basic authentication credentials", "src", r.RemoteAddr, "msg", err.Error())
					http.Error(w, "invalid basic credentials", http.StatusUnauthorized)
				}

				r = r.WithContext(context.WithValue(r.Context(), UserContextKey("user"), user))
				h.ServeHTTP(w, r)
				return
			}

			apiToken := getApiToken(r)
			if apiToken != "" {
				user, err := validateApiToken(r.Context(), svcGw, apiToken)
				if err != nil {
					logger.Error("invalid api token", "src", r.RemoteAddr, "error", err.Error())
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), UserContextKey("user"), user))
				h.ServeHTTP(w, r)
				return
			}

			// Process Bearer token authentication (OIDC/OAuth2)
			if oidcConfigs != nil {
				oidcToken := getOIDCToken(r)
				if oidcToken != "" {
					prov, err := getJWTTokenProvider(oidcToken)
					if err != nil {
						logger.Error("invalid OAuth2 credentials", "src", r.RemoteAddr, "error", err.Error())
						http.Error(w, "invalid OAuth2 credentials provided", http.StatusUnauthorized)
						return
					}

					userEmail, err := validateOIDCToken(r.Context(), oidcToken, prov)
					if err != nil && r.URL.Path != RootPageURI {
						logger.Error("invalid OAuth2 credentials", "src", r.RemoteAddr, "error", err.Error())
						http.Error(w, "invalid OAuth2 credentials provided", http.StatusUnauthorized)
						return
					}

					user, err := svcGw.Users.GetByLogin(r.Context(), entities.Email(userEmail))
					if err != nil {
						logger.Error("user from OAuth2 token not found in database", "src", r.RemoteAddr, "error", err.Error())
						http.Error(w, "invalid OAuth2 credentials provided", http.StatusUnauthorized)
						return
					}

					r = r.WithContext(context.WithValue(r.Context(), UserContextKey("user"), user))
					h.ServeHTTP(w, r)
					return
				}
			}

			// if authentication info not found still pass to the root webpage
			if r.URL.Path == RootPageURI {
				h.ServeHTTP(w, r)
				return
			}

			http.Error(w, "missing correct authentication data", http.StatusUnauthorized)
		})
	}
}
