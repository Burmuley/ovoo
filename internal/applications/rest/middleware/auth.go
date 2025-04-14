package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"slices"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
)

const (
	authorizationHeader = "Authorization"
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

			// Process Bearer token authentication (OIDC/OAuth2)
			oidcToken := getOIDCToken(r)
			if oidcToken != "" {
				userEmail, err := validateOIDCToken(r.Context(), oidcToken)
				if err != nil && r.URL.Path != OIDCLoginPageUri {
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

			// if authentication info not found still pass to the login webpage
			if r.URL.Path == OIDCLoginPageUri {
				h.ServeHTTP(w, r)
				return
			}

			http.Error(w, "missing correct authentication data", http.StatusUnauthorized)
		})
	}
}
