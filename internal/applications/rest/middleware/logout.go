package middleware

import "net/http"

// logout clears all OIDC and session cookies and redirects the browser to the root page.
//
// Parameters:
//   - w: HTTP response writer used to set the expired cookies and issue the redirect
//   - r: HTTP request used to determine the Secure flag and domain for cookie expiry
func logout(w http.ResponseWriter, r *http.Request) {
	clearOIDCCookies(w, r)
	setSecureCookie(w, r, stateCookieName, "", -1, "")
	setSecureCookie(w, r, nonceCookieName, "", -1, "")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
