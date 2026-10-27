package middleware

import (
	"context"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
)

// validateBasicAuth validates a user's login credentials against the database.
//
// It attempts to retrieve a user with the provided login (email) and then validates
// the provided password against the stored password hash. If either the user lookup
// fails or the password doesn't match, an error is returned.
//
// A user with an empty stored password hash is rejected outright: such accounts are
// authenticated through OIDC only and can never be reached with a password.
//
// Parameters:
//   - ctx: The context for the authentication request
//   - username: The user's email address used as login
//   - password: The plaintext password to verify
//   - svcGw: Service gateway providing access to user services
//
// Returns:
//   - entities.User: The authenticated user if successful
//   - error: An error if authentication fails (user not found or invalid password)
func validateBasicAuth(ctx context.Context, username, password string, svcGw *services.ServiceGateway) (entities.User, error) {
	user, err := svcGw.Users.GetByLogin(ctx, username)
	if err != nil {
		return entities.User{}, err
	}

	// an empty stored hash means password authentication is not enabled for this account
	if len(user.PasswordHash) == 0 {
		return entities.User{}, fmt.Errorf("password authentication is not enabled for the user")
	}

	if !entities.ValidPassword(password, user.PasswordHash) {
		return entities.User{}, fmt.Errorf("invalid password")
	}

	if !user.Active {
		return entities.User{}, fmt.Errorf("inactive user")
	}

	return user, nil
}
