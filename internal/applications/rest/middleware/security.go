package middleware

import (
	"fmt"
	"mime"
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
			for _, prov := range oidcConfigs {
				formActionUrls = append(formActionUrls, prov.OAuth2Config.Endpoint.AuthURL)
			}

			scriptSrcUrls := []string{
				"https://cdn.redoc.ly/redoc/v3.x/bundles/redoc.standalone.js",
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

func MaxBodyBytes(size int64) Adapter {
	return func(h http.Handler) http.Handler {
		return http.MaxBytesHandler(h, size)
	}
}

// CrossOriginProtection rejects cross-origin state-changing browser requests (CSRF).
// Requests with a non-safe method (anything but GET, HEAD, OPTIONS) are checked with
// Sec-Fetch-Site, falling back to comparing Origin with Host. Requests carrying neither
// header are not from browsers and pass, so API clients and tokens keep working.
func CrossOriginProtection() Adapter {
	cop := http.NewCrossOriginProtection()
	cop.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "cross-origin request rejected", http.StatusForbidden)
	}))

	return cop.Handler
}

// RequireJSONBody rejects requests with a body unless it is declared as application/json.
// Cross-site HTML forms can only send form or text/plain bodies without a CORS preflight
func RequireJSONBody() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength != 0 && r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
				mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil || mt != "application/json" {
					http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
					return
				}
			}

			h.ServeHTTP(w, r)
		})
	}
}
