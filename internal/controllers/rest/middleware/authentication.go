package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/Burmuley/ovoo/internal/services"
)

type UserContextKey string

// Authentication returns an Adapter middleware that enforces authentication
// requirements for HTTP requests.
//
// It accepts a list of URIs that should bypass authentication checks (skipUris).
// For any request where the URI is not in skipUris, it verifies the presence of an
// Authorization header. If the header is missing, the user is redirected to the
// login page at "/api/v1/users/login".
//
// Parameters:
//   - skipUris: A slice of string URIs that will bypass authentication checks
//
// Returns:
//   - An Adapter function that wraps an http.Handler with authentication logic
func Authentication(skipUris []string, svcGw *services.ServiceGateway, logger *slog.Logger) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if slices.Contains(skipUris, r.RequestURI) {
				h.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Redirect(w, r, "/api/v1/users/login", http.StatusFound)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 {
				logger.Error("invalid authentication token", "src", r.RemoteAddr)
				http.Error(w, "invalid authentication token", http.StatusUnauthorized)
			}
			tokenType, token := tokenParts[0], tokenParts[1]

			if !slices.Contains([]string{"Basic", "Bearer"}, tokenType) {
				logger.Error("invalid authentication token", "src", r.RemoteAddr)
				http.Error(w, "invalid authentication token", http.StatusUnauthorized)
				return
			}

			if tokenType == "Basic" {
				login, password, ok := r.BasicAuth()
				if !ok {
					logger.Error("invalid basic authentication token", "src", r.RemoteAddr)
					http.Error(w, "invalid basic auth token", http.StatusUnauthorized)
					return
				}

				user, err := validateBasicAuth(r.Context(), login, password, svcGw)
				if err != nil {
					logger.Error("invalid basic authentication credentials", "src", r.RemoteAddr, "msg", err.Error())
					http.Error(w, "invalid credentials provided", http.StatusUnauthorized)
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), UserContextKey("user"), user))
			}

			if tokenType == "Bearer" {
				user, err := validateOAuth2Token(r.Context(), token, svcGw)
				if err != nil {
					logger.Error("invalid oauth2 credentials", "src", r.RemoteAddr, "msg", err.Error())
					http.Error(w, "invalid oauth2 credentials provided", http.StatusUnauthorized)
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), UserContextKey("user"), user))
			}

			h.ServeHTTP(w, r)
		})
	}
}
