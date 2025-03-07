package middleware

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/Burmuley/ovoo/internal/entities"
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
func Authentication(skipUris []string, svcGw *services.ServiceGateway) Adapter {
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

			if strings.HasPrefix(authHeader, "Basic") {
				login, password, ok := r.BasicAuth()
				if !ok {
					http.Error(w, "invalid basic auth token", http.StatusUnauthorized)
					return
				}

				user, err := basicAuthenticator(r.Context(), login, password, svcGw)
				if err != nil {
					http.Error(w, "invalid credentials provided", http.StatusUnauthorized)
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), UserContextKey("user"), user))
			}

			h.ServeHTTP(w, r)
		})
	}
}

func basicAuthenticator(ctx context.Context, login, password string, svcGw *services.ServiceGateway) (entities.User, error) {
	user, err := svcGw.Users.GetByLogin(ctx, entities.Email(login))
	if err != nil {
		return entities.User{}, err
	}

	if !entities.ValidPassword(password, user.PasswordHash) {
		return entities.User{}, fmt.Errorf("invalid password")
	}

	return user, nil
}
