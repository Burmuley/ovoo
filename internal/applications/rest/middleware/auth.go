package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"github.com/Burmuley/ovoo/internal/services"
)

const (
	authorizationHeader = "Authorization"
	RootPageURI         = "/"

	// cookie names
	stateCookieName    = "ovoo_state"
	nonceCookieName    = "ovoo_nonce"
	accessCookieName   = "ovoo_access"   // access token, HttpOnly, short-lived
	refreshCookieName  = "ovoo_refresh"  // refresh token, HttpOnly, long-lived
	providerCookieName = "ovoo_provider" // OIDC provider name, HttpOnly, same TTL as refresh

	// ApiKey constants
	apiTokenCookieName = "ovoo_key"
	authProvidersUrl   = "/auth/providers"
	authLogoutUrl      = "/auth/logout"
)

var oidcConfigs map[string]OIDCProvider
var oidcProviderNames []string

type ContextKey string

const UserContextKey ContextKey = "user"

// Authentication creates a middleware adapter for handling user authentication.
// It supports multiple authentication methods and tries them in the following order:
//   - OIDC/OAuth2 via access token in the Authorization header
//   - OIDC/OAuth2 via access token and refresh token in HttpOnly cookies
//   - Basic authentication (username/password)
//   - API token via Authorization header or cookie
//
// OIDC cookie flow:
//  1. ovoo_access + ovoo_provider cookies present -> validate via UserInfo endpoint.
//  2. Access token invalid -> ovoo_refresh cookie present -> refresh -> set new cookies -> validate.
//
// OIDC Bearer flow:
//  1. Authorization: Bearer {access_token} -> resolve provider -> validate via UserInfo.
//
// Parameters:
//   - skipUris: URI path prefixes that bypass authentication entirely
//   - svcGw: service gateway providing user and token validation
//
// The middleware stores the authenticated user in the request context under
// UserContextKey and calls the next handler on success. Returns 401 for all
// requests that do not match a supported authentication method.
func Authentication(skipUris []string, svcGw *services.ServiceGateway) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if slices.ContainsFunc(skipUris, func(s string) bool {
				return strings.HasPrefix(r.URL.Path, s)
			}) {
				h.ServeHTTP(w, r)
				return
			}

			if r.URL.Path == authLogoutUrl {
				logout(w, r)
				return
			}

			if r.URL.Path == authProvidersUrl {
				resp, err := json.Marshal(oidcProviderNames)
				if err != nil {
					logger.Error("could render providers list", "src", r.RemoteAddr, "error", err.Error())
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write(resp); err != nil {
					logger.Error("authentication response write", "err", err.Error())
				}
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

				handleOIDCCallback(w, r, provider, match[1])
				return
			}

			if match := oidcRefreshUriReg.FindStringSubmatch(r.URL.Path); match != nil {
				provider, ok := oidcConfigs[match[1]]
				if !ok {
					logger.Error("unknown OIDC provider refresh URI", "src", r.RemoteAddr, "value", match[1])
					http.Error(w, "unknown OIDC provider refresh URI", http.StatusBadRequest)
					return
				}

				handleOIDCRefresh(w, r, provider, match[1])
				return
			}

			// Process Basic authentication (username/password) header
			if username, password, ok := r.BasicAuth(); ok {
				user, err := validateBasicAuth(r.Context(), username, password, svcGw)
				if err != nil {
					logger.Error("invalid basic authentication credentials", "src", r.RemoteAddr, "msg", err.Error())
					http.Error(w, "invalid basic credentials", http.StatusUnauthorized)
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), UserContextKey, user))
				h.ServeHTTP(w, r)
				return
			}

			// Process ApiToken authorization header
			apiToken := getApiToken(r)
			if apiToken != "" {
				user, err := validateApiToken(r.Context(), svcGw, apiToken)
				if err != nil {
					logger.Error("invalid api token", "src", r.RemoteAddr, "error", err.Error())
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), UserContextKey, user))
				h.ServeHTTP(w, r)
				return
			}

			if oidcConfigs != nil {
				// PATH 1: Authorization Bearer header (for API clients)
				if bearerToken := getBearerToken(r); bearerToken != "" {
					prov, err := resolveProviderForAccessToken(r, bearerToken)
					if err != nil {
						logger.Error("cannot identify OIDC provider for access token", "src", r.RemoteAddr, "error", err.Error())
						http.Error(w, "invalid OAuth2 credentials provided", http.StatusUnauthorized)
						return
					}

					email, err := validateAccessTokenViaUserInfo(r.Context(), bearerToken, prov)
					if err != nil {
						logger.Error("access token validation failed", "src", r.RemoteAddr, "error", err.Error())
						http.Error(w, "invalid OAuth2 credentials provided", http.StatusUnauthorized)
						return
					}

					user, err := svcGw.Users.GetByLogin(r.Context(), email)
					if err != nil {
						logger.Error("user from OAuth2 token not found in database", "src", r.RemoteAddr, "error", err.Error())
						http.Error(w, "invalid OAuth2 credentials provided", http.StatusUnauthorized)
						return
					}

					r = r.WithContext(context.WithValue(r.Context(), UserContextKey, user))
					h.ServeHTTP(w, r)
					return
				}

				// PATH 2: Cookie-based flow (for browser/SPA)
				if providerCookie, err := r.Cookie(providerCookieName); err == nil {
					prov, ok := oidcConfigs[providerCookie.Value]
					if !ok {
						logger.Error("invalid OIDC provider in cookie", "src", r.RemoteAddr, "provider", providerCookie.Value)
						clearOIDCCookies(w, r)
						http.Error(w, "invalid OAuth2 credentials provided", http.StatusUnauthorized)
						return
					}

					// try current access token
					if accessCookie, err := r.Cookie(accessCookieName); err == nil {
						if email, err := validateAccessTokenViaUserInfo(r.Context(), accessCookie.Value, prov); err == nil {
							user, err := svcGw.Users.GetByLogin(r.Context(), email)
							if err != nil {
								logger.Error("user from OAuth2 token not found in database", "src", r.RemoteAddr, "error", err.Error())
								http.Error(w, "invalid OAuth2 credentials provided", http.StatusUnauthorized)
								return
							}

							r = r.WithContext(context.WithValue(r.Context(), UserContextKey, user))
							h.ServeHTTP(w, r)
							return
						}
						// access token invalid or expired — fall through to refresh
					}

					// try refresh token
					if refreshCookie, err := r.Cookie(refreshCookieName); err == nil {
						newToken, err := refreshAccessToken(r.Context(), prov, refreshCookie.Value)
						if err != nil {
							logger.Error("failed to refresh OIDC access token", "src", r.RemoteAddr, "error", err.Error())
							clearOIDCCookies(w, r)
							http.Error(w, "session expired", http.StatusUnauthorized)
							return
						}

						// providers that don't rotate refresh tokens omit it from the response;
						// carry the existing token forward so setNewOIDCCookies resets its TTL
						if newToken.RefreshToken == "" {
							newToken.RefreshToken = refreshCookie.Value
						}

						setNewOIDCCookies(w, r, newToken, providerCookie.Value)
						w.Header().Set("X-Access-Token", newToken.AccessToken)

						email, err := validateAccessTokenViaUserInfo(r.Context(), newToken.AccessToken, prov)
						if err != nil {
							logger.Error("refreshed access token validation failed", "src", r.RemoteAddr, "error", err.Error())
							clearOIDCCookies(w, r)
							http.Error(w, "invalid OAuth2 credentials provided", http.StatusUnauthorized)
							return
						}

						user, err := svcGw.Users.GetByLogin(r.Context(), email)
						if err != nil {
							logger.Error("user from refreshed OAuth2 token not found in database", "src", r.RemoteAddr, "error", err.Error())
							http.Error(w, "invalid OAuth2 credentials provided", http.StatusUnauthorized)
							return
						}

						r = r.WithContext(context.WithValue(r.Context(), UserContextKey, user))
						h.ServeHTTP(w, r)
						return
					}
				}
			}

			// if authentication info not found still pass to the root webpage
			if r.URL.Path == RootPageURI {
				h.ServeHTTP(w, r)
				return
			}

			http.Error(w, "missing valid authentication headers", http.StatusUnauthorized)
		})
	}
}
