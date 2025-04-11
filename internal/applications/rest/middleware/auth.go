package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/Burmuley/ovoo/internal/services"
)

type UserContextKey string

// Authentication creates a middleware adapter for handling user authentication.
// It supports multiple authentication methods:
// - OIDC/OAuth2 via cookies (browser sessions)
// - Basic authentication (username/password)
// - Bearer token authentication (OIDC/OAuth2)
//
// Parameters:
//   - skipUris: a slice of URI paths that bypass authentication
//   - svcGw: service gateway for authentication operations
//   - logger: structured logger for recording authentication events
//
// The middleware adds authenticated user information to the request context
// with the UserContextKey("user") key when authentication succeeds.
func Authentication(skipUris []string, svcGw *services.ServiceGateway, logger *slog.Logger) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for whitelisted URIs
			if slices.Contains(skipUris, r.URL.Path) {
				h.ServeHTTP(w, r)
				return
			}

			// First try to authenticate using cookies (for browser sessions)
			authCookieValue, _ := r.Cookie(authCookie)
			if authCookieValue != nil {
				cookieToken := authCookieValue.Value
				user, err := validateOIDCToken(r.Context(), cookieToken, svcGw)
				if err != nil && r.URL.Path != OIDCLoginPageUri {
					logger.Error("invalid oauth2 credentials", "src", r.RemoteAddr, "error", err.Error())
					http.Error(w, "invalid oauth2 credentials provided", http.StatusUnauthorized)
					return
				} else if err == nil {
					r = r.WithContext(context.WithValue(r.Context(), UserContextKey("user"), user))
					h.ServeHTTP(w, r)
					return
				}
			}

			// Then check for Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" && r.URL.Path != OIDCLoginPageUri {
				http.Redirect(w, r, OIDCLoginUri, http.StatusFound)
				return
			}

			// Parse and validate the Authorization header
			tokenType, token, err := validateAuthHeader(authHeader)
			if err != nil && r.URL.Path != OIDCLoginPageUri {
				logger.Error(err.Error(), "src", r.RemoteAddr)
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}

			// Process Basic authentication (username/password)
			if tokenType == "Basic" {
				user, err := validateBasicAuth(r, svcGw)
				if err != nil {
					logger.Error("invalid basic authentication credentials", "src", r.RemoteAddr, "msg", err.Error())
					http.Error(w, "invalid credentials", http.StatusUnauthorized)
				}

				r = r.WithContext(context.WithValue(r.Context(), UserContextKey("user"), user))
				h.ServeHTTP(w, r)
				return
			}

			// Process Bearer token authentication (OIDC/OAuth2)
			if tokenType == "Bearer" {
				user, err := validateOIDCToken(r.Context(), token, svcGw)
				if err != nil && r.URL.Path != OIDCLoginPageUri {
					logger.Error("invalid oauth2 credentials", "src", r.RemoteAddr, "error", err.Error())
					http.Error(w, "invalid oauth2 credentials provided", http.StatusUnauthorized)
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), UserContextKey("user"), user))
				h.ServeHTTP(w, r)
				return
			}

			if r.URL.Path == OIDCLoginPageUri {
				h.ServeHTTP(w, r)
				return
			}

			http.Error(w, "missing correct authentication data", http.StatusUnauthorized)

		})
	}
}

func validateAuthHeader(header string) (string, string, error) {
	tokenParts := strings.Split(header, " ")
	if len(tokenParts) != 2 {
		return "", "", errors.New("invalid authentication token")
	}
	tokenType, token := tokenParts[0], tokenParts[1]
	if !slices.Contains([]string{"Basic", "Bearer"}, tokenType) {
		return "", "", errors.New("invalid authentication token")
	}

	return tokenType, token, nil
}
