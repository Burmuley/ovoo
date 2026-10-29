package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxBodyBytes(t *testing.T) {
	h := MaxBodyBytes(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		}
	}))

	small := httptest.NewRecorder()
	h.ServeHTTP(small, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("12345")))
	assert.Equal(t, http.StatusOK, small.Code)

	big := httptest.NewRecorder()
	h.ServeHTTP(big, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 100))))
	assert.Equal(t, http.StatusRequestEntityTooLarge, big.Code)
}

func TestValidateApiToken_TooLong(t *testing.T) {
	_, err := validateApiToken(context.Background(), nil, strings.Repeat("a", maxInputTokenSize+1))
	assert.EqualError(t, err, "token too long")
}

func TestValidateApiToken_InvalidFormat(t *testing.T) {
	_, err := validateApiToken(context.Background(), nil, "notoken")
	assert.EqualError(t, err, "invalid token format")
}

func TestSetSecureCookie_SameSiteLax(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	setSecureCookie(w, r, "n", "v", 60, "/")
	cookies := w.Result().Cookies()
	if assert.Len(t, cookies, 1) {
		assert.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
	}
}

func csrfHandler() http.Handler {
	return Adapt(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), CrossOriginProtection(), RequireJSONBody())
}

func doCSRF(method, body string, headers map[string]string) int {
	r := httptest.NewRequest(method, "http://ovoo.example/api/v1/users", strings.NewReader(body))
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	csrfHandler().ServeHTTP(w, r)
	return w.Code
}

func TestCrossOriginProtection(t *testing.T) {
	js := map[string]string{"Content-Type": "application/json"}
	with := func(k, v string) map[string]string {
		return map[string]string{"Content-Type": "application/json", k: v}
	}

	assert.Equal(t, http.StatusForbidden, doCSRF(http.MethodPost, "{}", with("Sec-Fetch-Site", "cross-site")))
	assert.Equal(t, http.StatusForbidden, doCSRF(http.MethodPost, "{}", with("Sec-Fetch-Site", "same-site")))
	assert.Equal(t, http.StatusForbidden, doCSRF(http.MethodDelete, "", with("Sec-Fetch-Site", "cross-site")))
	assert.Equal(t, http.StatusForbidden, doCSRF(http.MethodPost, "{}", with("Origin", "https://evil.example")))

	assert.Equal(t, http.StatusOK, doCSRF(http.MethodPost, "{}", with("Sec-Fetch-Site", "same-origin")))
	assert.Equal(t, http.StatusOK, doCSRF(http.MethodPost, "{}", with("Origin", "http://ovoo.example")))
	assert.Equal(t, http.StatusOK, doCSRF(http.MethodPost, "{}", js), "non-browser client without Origin")
	assert.Equal(t, http.StatusOK, doCSRF(http.MethodGet, "", with("Sec-Fetch-Site", "cross-site")), "safe method")
}

func TestRequireJSONBody(t *testing.T) {
	assert.Equal(t, http.StatusUnsupportedMediaType, doCSRF(http.MethodPost, `{"a":1}`, map[string]string{"Content-Type": "text/plain"}))
	assert.Equal(t, http.StatusUnsupportedMediaType, doCSRF(http.MethodPost, `{"a":1}`, map[string]string{"Content-Type": "application/x-www-form-urlencoded"}))
	assert.Equal(t, http.StatusUnsupportedMediaType, doCSRF(http.MethodPatch, `{"a":1}`, nil))
	assert.Equal(t, http.StatusOK, doCSRF(http.MethodPost, `{"a":1}`, map[string]string{"Content-Type": "application/json; charset=utf-8"}))
	assert.Equal(t, http.StatusOK, doCSRF(http.MethodPost, "", nil), "bodyless POST")
	assert.Equal(t, http.StatusOK, doCSRF(http.MethodGet, "", nil))
}

func TestLogout_PostOnly(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Fatal("next handler must not be called") })
	h := Authentication(nil, nil)(next)

	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(method, authLogoutUrl, nil))
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code, method)
		assert.Equal(t, http.MethodPost, w.Header().Get("Allow"), method)
		assert.Empty(t, w.Result().Cookies(), method, "cookies must not be touched")
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, authLogoutUrl, nil))
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NotEmpty(t, w.Result().Cookies(), "session cookies must be expired")
	for _, c := range w.Result().Cookies() {
		assert.Negative(t, c.MaxAge, c.Name)
	}
}
