package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
	"github.com/golang-jwt/jwt/v4"
)

type OAuth2Provider struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	AuthURL      string `yaml:"auth_url"`
	TokenURL     string `yaml:"token_url"`
	UserInfoURL  string `yaml:"userinfo_url"`
	Issuer       string `yaml:"issuer"`
}

var providers map[string]*OAuth2Provider

func LoadOAuth2Providers(cfgProviders map[string]*OAuth2Provider) {
	providers = cfgProviders
}

func validateOAuth2Token(ctx context.Context, token string, svcGw *services.ServiceGateway) (entities.User, error) {
	if len(providers) == 0 {
		return entities.User{}, fmt.Errorf("no oauth2 providers configured")
	}

	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return entities.User{}, fmt.Errorf("parsing jwt token: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return entities.User{}, fmt.Errorf("parsing jwt token: invalid claims")
	}

	issuer, exists := claims["iss"].(string)
	if !exists {
		return entities.User{}, fmt.Errorf("parsing jwt token: no issuer found in claims")
	}

	exp, exists := claims["exp"].(float64)
	if !exists || int64(exp) < time.Now().Unix() {
		return entities.User{}, fmt.Errorf("parsing jwt token: token expired")
	}

	var user entities.User
	var userEmail string
	for _, provider := range providers {
		if provider.Issuer == issuer {
			if err := validateWithProvider(token, provider.UserInfoURL); err != nil {
				return entities.User{}, fmt.Errorf("validating with provider: %w", err)
			}

			email, exists := claims["email"].(string)
			if !exists || email == "" {
				return entities.User{}, fmt.Errorf("invalid email in token")
			}
			userEmail = email
		}
	}

	user, err = svcGw.Users.GetByLogin(ctx, entities.Email(userEmail))
	if err != nil {
		return entities.User{}, fmt.Errorf("validating user in the system: %w", err)
	}

	return user, nil
}

func validateWithProvider(token, userInfoURL string) error {
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("validation with provider failed")
	}

	return nil
}
