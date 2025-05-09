package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	oidcLoginUriReg    = regexp.MustCompile(`/auth/(\w+)/login`)
	oidcCallbackUriReg = regexp.MustCompile(`/auth/(\w+)/callback`)
)

type OIDCProvider struct {
	OAuth2Config *oauth2.Config
	OIDCProvider *oidc.Provider
	OIDCConfig   *oidc.Config
	Issuer       string
}

func SetOIDCConfigs(configs map[string]OIDCProvider) {
	oidcConfigs = configs
}

// validateOIDCToken validates an OIDC token and extracts the user's email from its claims.
// It creates a verifier using the provider configuration, verifies the token, and extracts
// the email claim from the token. If the token is invalid, verification fails, or the token
// doesn't contain an email claim, an error is returned.
//
// Parameters:
//   - ctx: Context for the verification operation
//   - token: The OIDC token string to validate
//   - prov: The OIDCProvider configuration to use for validation
//
// Returns:
//   - string: The email address extracted from the token claims
//   - error: An error if token verification fails or no email claim is present
func validateOIDCToken(ctx context.Context, token string, prov OIDCProvider) (string, error) {
	verifier := prov.OIDCProvider.Verifier(&oidc.Config{ClientID: prov.OAuth2Config.ClientID})
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

// handleOIDCCallback handles the callback from the OIDC provider after user authentication.
// It validates the state and nonce parameters, exchanges the authorization code for an
// OAuth2 token, verifies the ID token, and sets an authentication cookie.
//
// The function performs the following steps:
// 1. Verifies that the state parameter matches the value stored in the cookie
// 2. Exchanges the authorization code for an OAuth2 token
// 3. Extracts the ID token from the OAuth2 token
// 4. Verifies the ID token with the OIDC provider
// 5. Checks that the nonce in the ID token matches the value stored in the cookie
// 6. Sets an authentication cookie with the ID token and redirects to the root page
//
// Parameters:
//   - w: HTTP response writer to write response and set cookies
//   - r: HTTP request containing callback parameters
//   - prov: The OIDCProvider configuration to use for token exchange
//
// The function handles errors by returning appropriate HTTP error responses
func handleOIDCCallback(w http.ResponseWriter, r *http.Request, prov OIDCProvider) {
	state, err := r.Cookie(stateCookieName)
	if err != nil {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "state does not match", http.StatusBadRequest)
		return
	}

	oauth2Token, err := prov.OAuth2Config.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to exchange token: %s", err.Error()), http.StatusBadRequest)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token field in oauth2 token", http.StatusBadRequest)
		return
	}

	idToken, err := prov.OIDCProvider.Verifier(prov.OIDCConfig).Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to verify id token: %s", err.Error()), http.StatusBadRequest)
		return
	}

	nonce, err := r.Cookie(nonceCookieName)
	if err != nil {
		http.Error(w, "nonce not found", http.StatusBadRequest)
		return
	}

	if idToken.Nonce != nonce.Value {
		http.Error(w, "nonce does not match", http.StatusBadRequest)
		return
	}

	cookieLife := int(idToken.Expiry.Sub(time.Now()).Seconds())
	setSecureCookie(w, r, authCookieName, rawIDToken, cookieLife, "/")
	http.Redirect(w, r, RootPageURI, http.StatusFound)
}

// handleOIDCLogin handles the OIDC login process by generating security parameters,
// setting appropriate cookies, and redirecting the user to the OIDC provider.
//
// It performs the following steps:
// 1. Generates a random nonce for security
// 2. Extracts the return_url from query parameters (defaults to RootPageURI)
// 3. Encodes the return_url as state parameter
// 4. Sets secure cookies for state and nonce with 1-hour expiration
// 5. Redirects the user to the OIDC provider's authorization endpoint
//
// Parameters:
//   - w: HTTP response writer used to set cookies and redirect
//   - r: HTTP request containing query parameters including optional return_url
//   - prov: The OIDCProvider configuration to use for the login flow
func handleOIDCLogin(w http.ResponseWriter, r *http.Request, prov OIDCProvider) {
	// generate random nonce
	nonce, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// generate state from the return url
	return_url := r.URL.Query().Get("return_url")
	if return_url == "" {
		return_url = RootPageURI
	}

	state := base64.URLEncoding.EncodeToString([]byte(return_url))

	// set security cookies
	setSecureCookie(w, r, stateCookieName, state, int(time.Hour.Seconds()), "")
	setSecureCookie(w, r, nonceCookieName, nonce, int(time.Hour.Seconds()), "")

	//redirect to the provider for authentication
	if !strings.HasPrefix(prov.OAuth2Config.RedirectURL, "http") {
		scheme := "https"
		if r.TLS == nil {
			scheme = "http"
		}
		prov.OAuth2Config.RedirectURL, _ = url.JoinPath(fmt.Sprintf("%s://%s", scheme, r.Host), prov.OAuth2Config.RedirectURL)
	}
	http.Redirect(w, r,
		prov.OAuth2Config.AuthCodeURL(state, oidc.Nonce(nonce)),
		http.StatusFound,
	)
}

// getOIDCToken extracts an OIDC token from the HTTP request.
// It checks for a Bearer token in the Authorization header first,
// and then falls back to checking for an authentication cookie.
//
// Parameters:
//   - r: The HTTP request to extract the token from
//
// Returns:
//   - string: The extracted OIDC token, or an empty string if no valid token is found
func getOIDCToken(r *http.Request) string {
	authHeader := r.Header.Get(authorizationHeader)
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if !strings.HasPrefix(token, entities.ApiTokenPrefix) {
			return token
		}

		return ""
	}

	if authCookie, err := r.Cookie(authCookieName); err == nil && authCookie != nil {
		return authCookie.Value
	}

	return ""
}

// getOIDCTokenIssuer extracts the issuer claim from a JWT token.
//
// Parameters:
//   - token: The JWT token string to parse
//
// Returns:
//   - string: The issuer claim value
//   - error: An error if the token is malformed or issuer claim is missing
func getOIDCTokenIssuer(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("malformed JWT token: should have 3 parts")
	}

	payloadBytes, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("error decoding JWT payload: %w", err)
	}

	payload := make(map[string]any)
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", fmt.Errorf("error unmarshaling JWT payload: %w", err)
	}

	iss := payload["iss"].(string)
	if iss == "" {
		return "", errors.New("issuer not defined in the JWT payload")
	}

	return iss, nil
}

// getProviderByIssuer looks up an OIDC provider configuration by issuer URL.
//
// Parameters:
//   - iss: The issuer URL to look up
//
// Returns:
//   - OIDCProvider: The matching provider configuration if found
//   - bool: Whether a matching provider was found
func getProviderByIssuer(iss string) (OIDCProvider, bool) {
	for _, provCfg := range oidcConfigs {
		if provCfg.Issuer == iss {
			return provCfg, true
		}
	}

	return OIDCProvider{}, false
}

// getJWTTokenProvider extracts the token issuer and returns the corresponding OIDC provider.
// It first gets the issuer from the JWT token claims, then looks up the matching provider
// configuration.
//
// Parameters:
//   - token: The JWT token string to parse
//
// Returns:
//   - OIDCProvider: The OIDC provider configuration for the token issuer
//   - error: An error if the token is invalid or no matching provider is found
func getJWTTokenProvider(token string) (OIDCProvider, error) {
	iss, err := getOIDCTokenIssuer(token)
	if err != nil {
		return OIDCProvider{}, err
	}

	prov, ok := getProviderByIssuer(iss)
	if !ok {
		return OIDCProvider{}, errors.New("token issued by unknown provider")
	}

	return prov, nil
}
