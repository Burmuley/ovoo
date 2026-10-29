package middleware

import "net/http"

// logout clears all OIDC and session cookies and responds with 204 No Content.
// It must only be called for POST requests so a cross-site link or image can not log the user out.
//
// Parameters:
//   - w: HTTP response writer used to set the expired cookies
//   - r: HTTP request used to determine the Secure flag and domain for cookie expiry
func logout(w http.ResponseWriter, r *http.Request) {
	clearOIDCCookies(w, r)
	setSecureCookie(w, r, stateCookieName, "", -1, "")
	setSecureCookie(w, r, nonceCookieName, "", -1, "")
	w.WriteHeader(http.StatusNoContent)
}
