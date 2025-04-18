package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
)

// getApiToken extracts an API token from the HTTP request.
// It first checks the Authorization header for a "Bearer" token with the format "Bearer {prefix}_{token}".
// If not found in the header, it attempts to retrieve the token from a cookie named by apiTokenCookieName.
// Returns the extracted token or an empty string if no token is found.
func getApiToken(r *http.Request) string {
	authHeader := r.Header.Get(authorizationHeader)
	prefix := fmt.Sprintf("Bearer %s_", entities.ApiTokenPrefix)
	if strings.HasPrefix(authHeader, prefix) {
		return strings.TrimPrefix(authHeader, prefix)
	}

	if apiKeyCookie, err := r.Cookie(apiTokenCookieName); err == nil && apiKeyCookie != nil {
		return apiKeyCookie.Value
	}

	return ""
}

// validateApiToken validates the provided API token and returns the associated user.
// The token is expected to be in the format "tokenId_tokenBody".
// It retrieves the token record from the database, computes a hash of the token body using the stored salt,
// and compares it with the stored hash. It also checks if the token has expired.
// Returns the user associated with the token if valid, otherwise returns an error.
//
// Parameters:
//   - ctx: The context for the operation
//   - svcGw: Service gateway providing access to token services
//   - apiToken: The API token to validate in the format "tokenId_tokenBody"
//
// Returns:
//   - entities.User: The user associated with the valid token
//   - error: An error if the token is invalid, expired, or not found
func validateApiToken(ctx context.Context, svcGw *services.ServiceGateway, apiToken string) (entities.User, error) {
	tokenParts := strings.SplitN(apiToken, "_", 2)
	if len(tokenParts) < 2 {
		return entities.User{}, errors.New("error parsing token body")
	}
	tokenId, err := entities.Base62Decode(tokenParts[0])
	if err != nil {
		return entities.User{}, err
	}
	tokenBody := tokenParts[1]
	token, err := svcGw.Tokens.GetByIdNoValidation(ctx, entities.Id(tokenId))
	if err != nil {
		return entities.User{}, err
	}

	hash := entities.HashApiToken(token.Salt, tokenBody)
	if hash != token.TokenHash || token.Expired() {
		return entities.User{}, errors.New("invalid or expired token")
	}

	return token.Owner, nil
}
