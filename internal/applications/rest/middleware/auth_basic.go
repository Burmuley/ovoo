package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
)

var errAccountLocked = errors.New("account is temporarily locked due to failed login attempts")

// validateBasicAuth validates a user's login credentials against the database.
//
// It attempts to retrieve a user with the provided login (email) and then validates
// the provided password against the stored password hash. If either the user lookup
// fails or the password doesn't match, an error is returned.
//
// A user with an empty stored password hash is rejected outright: such accounts are
// authenticated through OIDC only and can never be reached with a password.
//
// Failed password attempts are counted per account: after services.MaxFailedLogins consecutive
// failures the account is locked for a growing period, during which even the correct password
// is rejected. A successful login resets the counter.
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

	if user.IsLockedOut(time.Now()) {
		return entities.User{}, errAccountLocked
	}

	if !entities.ValidPassword(password, user.PasswordHash) {
		if err := svcGw.Users.RegisterLoginFailure(ctx, user); err != nil {
			slog.Error("failed to register failed login attempt", "user_id", user.ID, "error", err.Error())
		}
		return entities.User{}, fmt.Errorf("invalid password")
	}

	if err := svcGw.Users.ResetLoginFailures(ctx, user); err != nil {
		slog.Error("failed to reset failed login attempts", "user_id", user.ID, "error", err.Error())
	}

	if !user.Active {
		return entities.User{}, fmt.Errorf("inactive user")
	}

	return user, nil
}
