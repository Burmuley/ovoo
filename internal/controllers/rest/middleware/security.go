package middleware

import (
	"net/http"
)

func SecurityHeaders() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; form-action 'none'")
			w.Header().Add("X-Frame-Options", "DENY")
			w.Header().Add("X-Content-Type-Options", "nosniff")
			w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Add("Referrer-Policy", "no-referrer")

			h.ServeHTTP(w, r)
		})
	}
}
