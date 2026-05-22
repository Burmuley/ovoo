package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// randString generates a cryptographically secure random string of specified byte length.
// It uses crypto/rand to generate random bytes which are then encoded using base64 URL-safe encoding.
//
// Parameters:
//   - nByte: the number of random bytes to generate.
//
// Returns:
//   - string: the base64 URL-safe encoded random string.
//   - error: an error if random byte generation fails.
func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// setSecureCookie sets a secure HTTP cookie in the response.
// It configures the cookie with security best practices like HttpOnly flag
// and automatically sets the Secure flag when the request is over HTTPS.
//
// Parameters:
//   - w: HTTP response writer to set the cookie on.
//   - r: HTTP request providing context like TLS and hostname information.
//   - name: the name of the cookie.
//   - value: the value to store in the cookie.
//   - maxAge: maximum age of the cookie in seconds. Zero or negative means session cookie.
//   - path: the URL path for which the cookie is valid.
func setSecureCookie(w http.ResponseWriter, r *http.Request, name, value string, maxAge int, path string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Secure:   r.TLS != nil,
		HttpOnly: false,
		Path:     path,
		Domain:   r.URL.Hostname(),
	}
	http.SetCookie(w, c)
}

// formatRedirectURL resolves the OAuth2 redirect URL for the given OIDC provider.
// If the configured RedirectURL is a relative path (no "http" prefix), it constructs
// an absolute URL using the scheme derived from the request's TLS state and the request Host.
func formatRedirectURL(r *http.Request, prov OIDCProvider) string {
	redirectUrl := prov.OAuth2Config.RedirectURL
	if !strings.HasPrefix(redirectUrl, "http") {
		scheme := "https"
		if r.TLS == nil {
			scheme = "http"
		}
		redirectUrl, _ = url.JoinPath(fmt.Sprintf("%s://%s", scheme, r.Host), prov.OAuth2Config.RedirectURL)
	}

	return redirectUrl
}
