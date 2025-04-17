package middleware

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	OAuth2Config *oauth2.Config
	OIDCProvider *oidc.Provider
	OIDCConfig   *oidc.Config
	Issuer       string
}

var providerConfig *OIDCProvider

func SetOIDCProvider(provider *OIDCProvider) {
	providerConfig = provider
}

// validateOIDCToken validates an OIDC token and extracts the user's email from its claims.
// It creates a verifier using the provider configuration, verifies the token, and extracts
// the email claim from the token. If the token is invalid, verification fails, or the token
// doesn't contain an email claim, an error is returned.
//
// Parameters:
//   - ctx: Context for the verification operation
//   - token: The OIDC token string to validate
//
// Returns:
//   - string: The email address extracted from the token claims
//   - error: An error if token verification fails or no email claim is present
func validateOIDCToken(ctx context.Context, token string) (string, error) {
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

// HandleOIDCCallback handles the callback from the OIDC provider after user authentication.
// It validates the state and nonce parameters, exchanges the authorization code for an
// OAuth2 token, verifies the ID token, and sets an authentication cookie.
//
// The function performs the following steps:
// 1. Verifies that the state parameter matches the value stored in the cookie
// 2. Exchanges the authorization code for an OAuth2 token
// 3. Extracts the ID token from the OAuth2 token
// 4. Verifies the ID token with the OIDC provider
// 5. Checks that the nonce in the ID token matches the value stored in the cookie
// 6. Sets an authentication cookie with the ID token and redirects to the login page
//
// If any step fails, an appropriate HTTP error is returned.
func HandleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie(stateCookieName)
	if err != nil {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "state does not match", http.StatusBadRequest)
		return
	}

	oauth2Token, err := providerConfig.OAuth2Config.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to exchange token: %s", err.Error()), http.StatusBadRequest)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token field in oauth2 token", http.StatusBadRequest)
		return
	}

	idToken, err := providerConfig.OIDCProvider.Verifier(providerConfig.OIDCConfig).Verify(r.Context(), rawIDToken)
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
	http.Redirect(w, r, OIDCLoginPageUri, http.StatusFound)
}

// HandleOIDCLogin handles the OIDC login process by generating security parameters,
// setting appropriate cookies, and redirecting the user to the OIDC provider.
//
// It performs the following steps:
// 1. Generates a random nonce for security
// 2. Extracts the return_url from query parameters (defaults to OIDCLoginPageUri)
// 3. Encodes the return_url as state parameter
// 4. Sets secure cookies for state and nonce with 1-hour expiration
// 5. Redirects the user to the OIDC provider's authorization endpoint
//
// Parameters:
//   - w: HTTP response writer used to set cookies and redirect
//   - r: HTTP request containing query parameters including optional return_url
func HandleOIDCLogin(w http.ResponseWriter, r *http.Request) {
	// generate random nonce
	nonce, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// generate state from the return url
	return_url := r.URL.Query().Get("return_url")
	if return_url == "" {
		return_url = OIDCLoginPageUri
	}

	state := base64.URLEncoding.EncodeToString([]byte(return_url))

	// set security cookies
	setSecureCookie(w, r, stateCookieName, state, int(time.Hour.Seconds()), "")
	setSecureCookie(w, r, nonceCookieName, nonce, int(time.Hour.Seconds()), "")

	//redirect to the provider for authentication
	http.Redirect(w, r,
		providerConfig.OAuth2Config.AuthCodeURL(state, oidc.Nonce(nonce)),
		http.StatusFound,
	)
}

// getOIDCToken extracts an OIDC token from the HTTP request.
// It checks for a Bearer token in the Authorization header first,
// and then falls back to checking for an authentication cookie.
//
// The function performs the following steps:
// 1. Extracts the Authorization header from the request
// 2. If the header contains a Bearer token, trims the prefix
// 3. Validates that the token is not an API token (by checking prefix)
// 4. If no valid Bearer token is found, attempts to retrieve the token from a cookie
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
		if !strings.HasPrefix(token, apiTokenPrefix) {
			return token
		}

		return ""
	}

	if authCookie, err := r.Cookie(authCookieName); err == nil && authCookie != nil {
		return authCookie.Value
	}

	return ""
}
