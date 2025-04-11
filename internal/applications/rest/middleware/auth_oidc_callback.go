package middleware

import (
	"fmt"
	"net/http"
	"time"
)

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
	state, err := r.Cookie(stateCookie)
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

	nonce, err := r.Cookie(nonceCookie)
	if err != nil {
		http.Error(w, "nonce not found", http.StatusBadRequest)
		return
	}

	if idToken.Nonce != nonce.Value {
		http.Error(w, "nonce does not match", http.StatusBadRequest)
		return
	}

	cookieLife := int(idToken.Expiry.Sub(time.Now()).Seconds())
	setSecureCookie(w, r, authCookie, rawIDToken, cookieLife, "/")
	http.Redirect(w, r, OIDCLoginPageUri, http.StatusFound)
}
