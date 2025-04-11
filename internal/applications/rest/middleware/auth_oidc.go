package middleware

import (
	"context"
	"errors"
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	OAuth2Config *oauth2.Config
	OIDCProvider *oidc.Provider
	OIDCConfig   *oidc.Config
	Issuer       string
}

const (
	OIDCLoginUri     = "/auth/oidc"
	OIDCCallbackUri  = "/auth/callback"
	OIDCLoginPageUri = "/"

	stateCookie = "ovoo_state"
	nonceCookie = "ovoo_nonce"
	authCookie  = "ovoo_auth"
)

var providerConfig *OIDCProvider

func SetOIDCProvider(provider *OIDCProvider) {
	providerConfig = provider
	// fmt.Println(*providerConfig)
}

// validateOIDCToken validates the OIDC token with the provider and retrieves the associated user.
// It first verifies the token with the OIDC provider to extract the user's email,
// then fetches the corresponding user from the system using the service gateway.
//
// Parameters:
//   - ctx: The context for the operation
//   - token: The OIDC token to validate
//   - svcGw: The service gateway used to access user services
//
// Returns:
//   - entities.User: The user associated with the validated token
//   - error: An error if validation fails or the user cannot be found
func validateOIDCToken(ctx context.Context, token string, svcGw *services.ServiceGateway) (entities.User, error) {
	var user entities.User
	var userEmail string
	var err error
	if userEmail, err = validateWithProvider(ctx, token); err != nil {
		return entities.User{}, fmt.Errorf("validating with provider: %w", err)
	}

	user, err = svcGw.Users.GetByLogin(ctx, entities.Email(userEmail))
	if err != nil {
		return entities.User{}, fmt.Errorf("validating user in the system: %w", err)
	}

	return user, nil
}

// validateWithProvider verifies the provided token with the OIDC provider and extracts the email from the token claims.
// It creates a verifier using the configured OIDC provider and client ID, then verifies the token and extracts the email claim.
//
// Parameters:
//   - ctx: The context for the verification operation
//   - token: The OIDC token string to be verified
//
// Returns:
//   - string: The email address extracted from the token claims
//   - error: An error if token verification fails, if claims extraction fails, or if the email claim is empty
func validateWithProvider(ctx context.Context, token string) (string, error) {
	verifier := providerConfig.OIDCProvider.Verifier(&oidc.Config{ClientID: providerConfig.OAuth2Config.ClientID})
	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		return "", err
	}
	claims := struct {
		Email string
	}{}
	if err := idToken.Claims(&claims); err != nil {
		return "", err
	}
	if claims.Email == "" {
		return "", errors.New("token claims got no email")
	}
	return claims.Email, nil
}
