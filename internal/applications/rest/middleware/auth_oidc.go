package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	refreshTokenMaxAge = 30 * 24 * 3600 // 30 days in seconds
	userInfoCacheTTL   = 60 * time.Second
)

type cachedUserInfo struct {
	email     string
	expiresAt time.Time
}

var userInfoCache sync.Map

var (
	oidcLoginUriReg    = regexp.MustCompile(`/auth/(\w+)/login`)
	oidcCallbackUriReg = regexp.MustCompile(`/auth/(\w+)/callback`)
	oidcRefreshUriReg  = regexp.MustCompile(`/auth/(\w+)/refresh`)
)

type OIDCProvider struct {
	OAuth2Config   *oauth2.Config
	OIDCProvider   *oidc.Provider
	OIDCConfig     *oidc.Config
	Issuer         string
	ExtraScopes    []string
	ExtraURLParams map[string]string
}

// SetOIDCConfigs stores the OIDC provider configurations and populates the list
// of provider names used for the /auth/providers endpoint.
//
// Parameters:
//   - configs: map of provider name to OIDCProvider configuration
func SetOIDCConfigs(configs map[string]OIDCProvider) {
	oidcConfigs = configs
	oidcProviderNames = make([]string, 0, len(oidcConfigs))
	for name := range oidcConfigs {
		oidcProviderNames = append(oidcProviderNames, name)
	}
}

// validateAccessTokenViaUserInfo validates an access token by calling the OIDC provider's
// UserInfo endpoint and returns the user's email address. Results are cached for
// userInfoCacheTTL seconds to avoid per-request network calls to the provider.
//
// Parameters:
//   - ctx: context for the UserInfo request
//   - accessToken: the OAuth2 access token to validate
//   - prov: the OIDCProvider configuration to use for the UserInfo call
//
// Returns:
//   - string: the email address from the UserInfo response
//   - error: an error if the UserInfo call fails or the response contains no email
func validateAccessTokenViaUserInfo(ctx context.Context, accessToken string, prov OIDCProvider) (string, error) {
	if v, ok := userInfoCache.Load(accessToken); ok {
		if entry := v.(cachedUserInfo); time.Now().Before(entry.expiresAt) {
			return entry.email, nil
		}
	}

	userInfo, err := prov.OIDCProvider.UserInfo(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}))
	if err != nil {
		return "", err
	}

	if userInfo.Email == "" {
		return "", errors.New("userinfo response contains no email")
	}

	userInfoCache.Store(accessToken, cachedUserInfo{
		email:     userInfo.Email,
		expiresAt: time.Now().Add(userInfoCacheTTL),
	})

	return userInfo.Email, nil
}

// getBearerToken extracts an OAuth2 access token from the Authorization header.
// Returns an empty string if the header is absent or if the token belongs to an
// API key (identified by the ApiTokenPrefix).
//
// Parameters:
//   - r: the HTTP request to read the Authorization header from
//
// Returns:
//   - string: the extracted bearer token, or an empty string if not present
func getBearerToken(r *http.Request) string {
	token, ok := strings.CutPrefix(r.Header.Get(authorizationHeader), "Bearer ")
	if !ok {
		return ""
	}

	if strings.HasPrefix(token, entities.ApiTokenPrefix) {
		return ""
	}

	return token
}

// resolveProviderForAccessToken identifies the OIDC provider for a given access token.
// Resolution is attempted in the following order:
// 1. Parse the token as a JWT and match the "iss" claim against known providers.
// 2. Read the ovoo_provider cookie and look up the named provider.
// 3. If exactly one provider is configured, use it without any token inspection.
//
// Parameters:
//   - r: the HTTP request, used to read the ovoo_provider cookie
//   - token: the access token string to identify the provider for
//
// Returns:
//   - OIDCProvider: the matching provider configuration
//   - error: an error if no provider can be determined
func resolveProviderForAccessToken(r *http.Request, token string) (OIDCProvider, error) {
	if prov, err := getJWTTokenProvider(token); err == nil {
		return prov, nil
	}

	if provCookie, err := r.Cookie(providerCookieName); err == nil {
		if prov, ok := oidcConfigs[provCookie.Value]; ok {
			return prov, nil
		}
	}

	if len(oidcConfigs) == 1 {
		for _, prov := range oidcConfigs {
			return prov, nil
		}
	}

	return OIDCProvider{}, errors.New("cannot determine OIDC provider for access token")
}

// refreshAccessToken uses a refresh token to obtain a new access token from the provider.
//
// Parameters:
//   - ctx: context for the token endpoint request
//   - prov: the OIDCProvider whose token endpoint will be called
//   - refreshToken: the refresh token string to exchange
//
// Returns:
//   - *oauth2.Token: the new token containing at least a fresh access token
//   - error: an error if the token endpoint request fails or returns an error response
func refreshAccessToken(ctx context.Context, prov OIDCProvider, refreshToken string) (*oauth2.Token, error) {
	return prov.OAuth2Config.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).Token()
}

// setNewOIDCCookies writes ovoo_access and ovoo_provider cookies from the given token,
// and ovoo_refresh if the token contains a refresh token. ovoo_provider is always
// written so the middleware can identify the provider for subsequent UserInfo calls
// even when no refresh token is present. When a refresh token is present, both
// ovoo_refresh and ovoo_provider are written with a refreshTokenMaxAge TTL.
//
// Parameters:
//   - w: HTTP response writer to set the cookies on
//   - r: HTTP request used to determine the Secure flag and domain
//   - token: the OAuth2 token to persist; RefreshToken may be empty
//   - providerName: the provider key to store in the ovoo_provider cookie
func setNewOIDCCookies(w http.ResponseWriter, r *http.Request, token *oauth2.Token, providerName string) {
	accessTTL := 3600
	if !token.Expiry.IsZero() {
		if secs := int(time.Until(token.Expiry).Seconds()); secs > 0 {
			accessTTL = secs
		}
	}

	setSecureCookie(w, r, accessCookieName, token.AccessToken, accessTTL, "/")
	setSecureCookie(w, r, providerCookieName, providerName, accessTTL, "/")

	if token.RefreshToken != "" {
		setSecureCookie(w, r, refreshCookieName, token.RefreshToken, refreshTokenMaxAge, "/")
		// extend provider cookie lifetime to match the longer-lived refresh token
		setSecureCookie(w, r, providerCookieName, providerName, refreshTokenMaxAge, "/")
	}
}

// clearOIDCCookies expires the ovoo_access, ovoo_refresh, and ovoo_provider cookies
// by setting their MaxAge to -1.
//
// Parameters:
//   - w: HTTP response writer to set the expired cookies on
//   - r: HTTP request used to determine the Secure flag and domain
func clearOIDCCookies(w http.ResponseWriter, r *http.Request) {
	setSecureCookie(w, r, accessCookieName, "", -1, "/")
	setSecureCookie(w, r, refreshCookieName, "", -1, "/")
	setSecureCookie(w, r, providerCookieName, "", -1, "/")
}

// handleOIDCCallback completes the OAuth2 authorization code exchange initiated by
// handleOIDCLogin. The ID token is verified only to validate the nonce; it is not
// stored. The access token and refresh token are stored in HttpOnly cookies.
//
// The function performs the following steps:
// 1. Verifies the state query parameter against the ovoo_state cookie.
// 2. Exchanges the authorization code for an OAuth2 token.
// 3. Verifies the ID token signature and nonce for replay protection.
// 4. Stores the access token, refresh token, and provider name in HttpOnly cookies.
// 5. Redirects the browser to RootPageURI.
//
// Parameters:
//   - w: HTTP response writer used to set cookies and issue the redirect
//   - r: HTTP request containing the state and code query parameters
//   - prov: the OIDCProvider configuration to use for token exchange and verification
//   - providerName: the provider key stored in the ovoo_provider cookie
//
// The function handles errors by writing appropriate HTTP error responses.
func handleOIDCCallback(w http.ResponseWriter, r *http.Request, prov OIDCProvider, providerName string) {
	state, err := r.Cookie(stateCookieName)
	if err != nil {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "state does not match", http.StatusBadRequest)
		return
	}

	redirectUrl := formatRedirectURL(r, prov)
	oauth2Token, err := prov.OAuth2Config.Exchange(
		r.Context(),
		r.URL.Query().Get("code"),
		oauth2.SetAuthURLParam("redirect_uri", redirectUrl),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to exchange token: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// verify the ID token only to validate the nonce -- it is not stored
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token field in oauth2 token", http.StatusBadRequest)
		return
	}

	idToken, err := prov.OIDCProvider.Verifier(prov.OIDCConfig).Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to verify id token: %s", err.Error()), http.StatusBadRequest)
		return
	}

	nonce, err := r.Cookie(nonceCookieName)
	if err != nil {
		http.Error(w, "nonce not found", http.StatusBadRequest)
		return
	}

	if idToken.Nonce != nonce.Value {
		http.Error(w, "nonce does not match", http.StatusBadRequest)
		return
	}

	if oauth2Token.RefreshToken == "" {
		logger.Warn("OIDC provider returned no refresh token")
	}

	setNewOIDCCookies(w, r, oauth2Token, providerName)
	http.Redirect(w, r, RootPageURI, http.StatusFound)
}

// handleOIDCLogin initiates the OIDC authorization code flow by generating security
// parameters, setting cookies, and redirecting the browser to the provider.
//
// The function performs the following steps:
// 1. Generates a cryptographically random nonce.
// 2. Reads the optional return_url query parameter (defaults to RootPageURI).
// 3. Encodes the return URL as the OAuth2 state parameter.
// 4. Sets ovoo_state and ovoo_nonce cookies with a one-hour expiry.
// 5. Redirects the browser to the provider's authorization endpoint.
//
// Parameters:
//   - w: HTTP response writer used to set cookies and issue the redirect
//   - r: HTTP request containing the optional return_url query parameter
//   - prov: the OIDCProvider configuration to use for the authorization URL
func handleOIDCLogin(w http.ResponseWriter, r *http.Request, prov OIDCProvider) {
	nonce, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	return_url := r.URL.Query().Get("return_url")
	if return_url == "" {
		return_url = RootPageURI
	}

	state := base64.URLEncoding.EncodeToString([]byte(return_url))
	setSecureCookie(w, r, stateCookieName, state, int(time.Hour.Seconds()), "")
	setSecureCookie(w, r, nonceCookieName, nonce, int(time.Hour.Seconds()), "")

	redirectUrl := formatRedirectURL(r, prov)
	opts := []oauth2.AuthCodeOption{
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("redirect_uri", redirectUrl),
	}
	if len(prov.ExtraURLParams) > 0 {
		for key, val := range prov.ExtraURLParams {
			opts = append(opts, oauth2.SetAuthURLParam(key, val))
		}
	}
	http.Redirect(w, r,
		prov.OAuth2Config.AuthCodeURL(state, opts...),
		http.StatusFound,
	)
}

// handleOIDCRefresh exchanges the ovoo_refresh cookie for a new access token and
// returns it as a JSON response. On success, ovoo_access and ovoo_refresh cookies
// are updated. If the provider does not rotate the refresh token, the existing
// refresh token is re-written to reset its cookie TTL.
//
// The function performs the following steps:
// 1. Reads the ovoo_refresh cookie; returns 400 if absent or empty.
// 2. Calls the provider token endpoint with the refresh token grant.
// 3. On failure, clears all OIDC cookies and returns 401.
// 4. On success, updates cookies and writes {"access_token","expires_in"} as JSON.
//
// Parameters:
//   - w: HTTP response writer used to set cookies and write the JSON response
//   - r: HTTP request used to read the ovoo_refresh cookie
//   - prov: the OIDCProvider configuration to use for token refresh
//   - providerName: the provider key passed to setNewOIDCCookies
func handleOIDCRefresh(w http.ResponseWriter, r *http.Request, prov OIDCProvider, providerName string) {
	refreshCookie, err := r.Cookie(refreshCookieName)
	if err != nil || refreshCookie.Value == "" {
		http.Error(w, "refresh token not found", http.StatusBadRequest)
		return
	}

	newToken, err := refreshAccessToken(r.Context(), prov, refreshCookie.Value)
	if err != nil {
		logger.Error("failed to refresh access token", "provider", providerName, "error", err.Error())
		clearOIDCCookies(w, r)
		http.Error(w, "session expired", http.StatusUnauthorized)
		return
	}

	// providers that don't rotate refresh tokens omit it from the response;
	// carry the existing token forward so setNewOIDCCookies resets its TTL
	if newToken.RefreshToken == "" {
		newToken.RefreshToken = refreshCookie.Value
	}

	setNewOIDCCookies(w, r, newToken, providerName)

	expiresIn := 3600
	if !newToken.Expiry.IsZero() {
		if secs := int(time.Until(newToken.Expiry).Seconds()); secs > 0 {
			expiresIn = secs
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}{
		AccessToken: newToken.AccessToken,
		ExpiresIn:   expiresIn,
	}); err != nil {
		logger.Error("failed to encode response", "err", err.Error())
		clearOIDCCookies(w, r)
		http.Error(w, "session expired", http.StatusUnauthorized)
	}
}

// getOIDCTokenIssuer extracts the issuer ("iss") claim from a JWT token payload
// without verifying the token's signature.
//
// Parameters:
//   - token: the JWT token string (three base64url-encoded segments joined by dots)
//
// Returns:
//   - string: the issuer claim value
//   - error: an error if the token is malformed or the issuer claim is absent
func getOIDCTokenIssuer(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("malformed JWT token: should have 3 parts")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("error decoding JWT payload: %w", err)
	}

	payload := make(map[string]any)
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", fmt.Errorf("error unmarshaling JWT payload: %w", err)
	}

	issRaw, ok := payload["iss"]
	iss, okType := issRaw.(string)
	if !ok || !okType || iss == "" {
		return "", errors.New("issuer not defined in the JWT payload")
	}

	return iss, nil
}

// getProviderByIssuer looks up an OIDC provider configuration by its issuer URL.
//
// Parameters:
//   - iss: the issuer URL to search for
//
// Returns:
//   - OIDCProvider: the matching provider configuration
//   - bool: true if a matching provider was found, false otherwise
func getProviderByIssuer(iss string) (OIDCProvider, bool) {
	for _, provCfg := range oidcConfigs {
		if provCfg.Issuer == iss {
			return provCfg, true
		}
	}

	return OIDCProvider{}, false
}

// getJWTTokenProvider extracts the issuer from a JWT token and returns the
// corresponding OIDC provider configuration.
//
// Parameters:
//   - token: the JWT token string to parse
//
// Returns:
//   - OIDCProvider: the OIDC provider configuration for the token's issuer
//   - error: an error if the token is malformed or no matching provider is found
func getJWTTokenProvider(token string) (OIDCProvider, error) {
	iss, err := getOIDCTokenIssuer(token)
	if err != nil {
		return OIDCProvider{}, err
	}

	prov, ok := getProviderByIssuer(iss)
	if !ok {
		return OIDCProvider{}, errors.New("token issued by unknown provider")
	}

	return prov, nil
}
