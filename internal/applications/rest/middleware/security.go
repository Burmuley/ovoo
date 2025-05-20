package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

// SecurityHeaders returns an Adapter that adds security-related HTTP headers to the response.
// The following headers are added:
// - Content-Security-Policy: Restricts the sources from which content can be loaded. CSP directives include:
//   - default-src 'none': Blocks all content by default
//   - frame-ancestors 'none': Prevents embedding in iframes
//   - form-action 'self' + OAuth URLs: Restricts form submissions
//   - style-src 'self' 'unsafe-inline': Allows inline styles and local CSS
//   - connect-src 'self': Restricts API calls to same origin
//   - script-src 'self' + Redoc CDN: Allows local scripts and Redoc documentation
//   - worker-src 'self' blob:: Allows web workers
//   - img-src 'self' 'unsafe-inline' + Redoc CDN: Allows images from same origin and Redoc
//
// - X-Frame-Options: Prevents the page from being displayed in an iframe.
// - X-Content-Type-Options: Prevents MIME type sniffing.
// - Cache-Control: Prevents caching of the response.
// - Referrer-Policy: Controls what information is included in the Referer header.
// - Strict-Transport-Security: Enforces HTTPS connections with 1 year max age.
func SecurityHeaders() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// format and define Content Security Policy (CSP)
			formActionUrls := []string{}
			if oidcConfigs != nil {
				for _, prov := range oidcConfigs {
					formActionUrls = append(formActionUrls, prov.OAuth2Config.Endpoint.AuthURL)
				}
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
				fmt.Sprintf("script-src 'self' 'unsafe-inline' %s", strings.Join(scriptSrcUrls, " ")),
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
