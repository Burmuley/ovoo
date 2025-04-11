package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

// SecurityHeaders returns an Adapter that adds security-related HTTP headers to the response.
// The following headers are added:
// - Content-Security-Policy: Restricts the sources from which content can be loaded.
// - X-Frame-Options: Prevents the page from being displayed in an iframe.
// - X-Content-Type-Options: Prevents MIME type sniffing.
// - Cache-Control: Prevents caching of the response.
// - Referrer-Policy: Controls what information is included in the Referer header.
// - Strict-Transport-Security: Enforces HTTPS connections.
func SecurityHeaders() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cspUrls := []string{}
			if providerConfig != nil {
				cspUrls = append(cspUrls, providerConfig.OAuth2Config.Endpoint.AuthURL)
			}

			w.Header().Add(
				"Content-Security-Policy",
				fmt.Sprintf(
					"default-src 'none'; frame-ancestors 'none'; form-action 'self' %s",
					strings.Join(cspUrls, " "),
				),
			)
			w.Header().Add("X-Frame-Options", "DENY")
			w.Header().Add("X-Content-Type-Options", "nosniff")
			w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Add("Referrer-Policy", "no-referrer")
			w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			h.ServeHTTP(w, r)
		})
	}
}
