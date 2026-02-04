package middleware

import "net/http"

func logout(w http.ResponseWriter, r *http.Request) {
	// clear all auth cookies by setting MaxAge to -1
	setSecureCookie(w, r, authCookieName, "", -1, "/")
	setSecureCookie(w, r, stateCookieName, "", -1, "")
	setSecureCookie(w, r, nonceCookieName, "", -1, "")

	// redirect to root
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
