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
			// format and define Content Security Policy (CSP)
			formActionUrls := []string{}
			if providerConfig != nil {
				formActionUrls = append(formActionUrls, providerConfig.OAuth2Config.Endpoint.AuthURL)
			}

			scriptSrcUrls := []string{
				"https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js",
			}
			cspElems := []string{
				"default-src 'none'",
				"frame-ancestors 'none'",
				fmt.Sprintf("form-action 'self' %s", strings.Join(formActionUrls, " ")),
				"style-src 'self' 'unsafe-inline'",
				"connect-src 'self'",
				fmt.Sprintf("script-src 'self' %s", strings.Join(scriptSrcUrls, " ")),
				"worker-src 'self' blob:",
				"img-src 'self' 'unsafe-inline' https://cdn.redoc.ly/redoc/logo-mini.svg",
			}

			w.Header().Add("Content-Security-Policy", strings.Join(cspElems, ";"))

			// define other security headers
			w.Header().Add("X-Frame-Options", "DENY")
			w.Header().Add("X-Content-Type-Options", "nosniff")
			w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Add("Referrer-Policy", "no-referrer")
			w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			h.ServeHTTP(w, r)
		})
	}
}
