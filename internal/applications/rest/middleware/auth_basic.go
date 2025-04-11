package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
)

// validateBasicAuth validates a user's login credentials against the database.
//
// It attempts to retrieve a user with the provided login (email) and then validates
// the provided password against the stored password hash. If either the user lookup
// fails or the password doesn't match, an error is returned.
//
// Parameters:
//   - ctx: The context for the authentication request
//   - login: The user's email address used as login
//   - password: The plaintext password to verify
//   - svcGw: Service gateway providing access to user services
//
// Returns:
//   - entities.User: The authenticated user if successful
//   - error: An error if authentication fails (user not found or invalid password)
func validateBasicAuth(req *http.Request, svcGw *services.ServiceGateway) (entities.User, error) {
	login, password, ok := req.BasicAuth()
	if !ok {
		return entities.User{}, errors.New("invalid authentication token")
	}

	user, err := svcGw.Users.GetByLogin(req.Context(), entities.Email(login))
	if err != nil {
		return entities.User{}, err
	}

	if !entities.ValidPassword(password, user.PasswordHash) {
		return entities.User{}, fmt.Errorf("invalid password")
	}

	return user, nil
}
