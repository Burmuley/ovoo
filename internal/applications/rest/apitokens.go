package rest

import (
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
)

// GetApiTokens retrieves all API tokens associated with the authenticated user.
// It returns a list of API tokens in the response.
//
// Returns:
//   - 200 OK with list of API tokens on success
//   - Error response if user authentication or token retrieval fails
func (a *Application) GetApiTokens(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting api tokens: identifying user", err)
	}

	tokens, err := a.svcGw.Tokens.GetAll(r.Context(), user.ID)
	if err != nil {
		a.errorLogNResponse(w, "getting aliases: parsing id", err)
		return
	}

	resp := make(GetApiTokensResponse, 0, len(tokens))
	for _, v := range tokens {
		resp = append(resp, tokenTApiTokenData(v))
	}

	a.successResponse(w, resp, http.StatusOK)
}

// GetApiTokenById retrieves a specific API token by its ID for the authenticated user.
// It requires the token ID as a path parameter.
//
// Returns:
//   - 200 OK with token details on success
//   - Error response if user authentication or token retrieval fails
func (a *Application) GetApiTokenById(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting api token by id: identifying user", err)
	}

	token, err := a.svcGw.Tokens.GetById(r.Context(), entities.Id(r.PathValue("id")), user.ID)
	if err != nil {
		a.errorLogNResponse(w, "getting aliases: parsing id", err)
		return
	}

	resp := GetApiTokenDetailsResponse(tokenTApiTokenData(token))
	a.successResponse(w, resp, http.StatusOK)
}

// CreateApiToken creates a new API token for the authenticated user.
// It accepts a request containing token name, description, and expiration time.
//
// Request Body:
//   - Name: Name of the token (required)
//   - Description: Description of the token (optional)
//   - ExpireIn: Token expiration time in seconds
//
// Returns:
//   - 200 OK with the created token details on success
//   - Error response if user authentication, request parsing, or token creation fails
func (a *Application) CreateApiToken(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "creating new api token: identifying user", err)
	}

	req := CreateApiToken{}
	if err := readBody(r.Body, &req); err != nil {
		a.errorLogNResponse(w, "creating new api token: parsing request", err)
	}

	token, err := a.svcGw.Tokens.Create(r.Context(), user, req.Name, *req.Description, int(req.ExpireIn))
	if err != nil {
		a.errorLogNResponse(w, "creating new api token", err)
	}

	resp := CreateApiTokenResponse(tokenTApiTokenDataOnCreate(token))
	a.successResponse(w, resp, http.StatusOK)
}

// UpdateApiToken updates an existing API token.
// This function is not yet implemented.
//
// TODO: Implement token update functionality
func (a *Application) UpdateApiToken(w http.ResponseWriter, r *http.Request) {
	panic("implement me!")
}

// DeleteApiToken deletes an API token.
// This function is not yet implemented.
//
// TODO: Implement token deletion functionality
func (a *Application) DeleteApiToken(w http.ResponseWriter, r *http.Request) {
	panic("implement me!")
}
