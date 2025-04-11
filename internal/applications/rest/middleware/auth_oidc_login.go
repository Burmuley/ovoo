package middleware

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

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
	setSecureCookie(w, r, stateCookie, state, int(time.Hour.Seconds()), "")
	setSecureCookie(w, r, nonceCookie, nonce, int(time.Hour.Seconds()), "")

	//redirect to the provider for authentication
	http.Redirect(w, r,
		providerConfig.OAuth2Config.AuthCodeURL(state, oidc.Nonce(nonce)),
		http.StatusFound,
	)
}
