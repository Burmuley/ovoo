package rest

import (
	"fmt"
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
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting api tokens: identifying user", err)
		return
	}

	tokens, err := a.svcGw.Tokens.GetAll(r.Context(), cuser)
	if err != nil {
		a.errorLogNResponse(w, "getting tokens: parsing id", err)
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
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting api token by id: identifying user", err)
		return
	}

	token, err := a.svcGw.Tokens.GetByIdCurUser(r.Context(), cuser, entities.Id(r.PathValue("id")))
	if err != nil {
		a.errorLogNResponse(w, "getting token by id", err)
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
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "creating new api token: identifying user", err)
	}

	req := CreateApiToken{}
	if err := readBody(r.Body, &req); err != nil {
		a.errorLogNResponse(w, "creating new api token: parsing request", fmt.Errorf("%w: %w", entities.ErrValidation, err))
		return
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}
	token, err := a.svcGw.Tokens.Create(r.Context(), cuser, req.Name, description, int(req.ExpireIn))
	if err != nil {
		a.errorLogNResponse(w, "creating new api token", err)
		return
	}

	resp := CreateApiTokenResponse(tokenTApiTokenDataOnCreate(token))
	a.successResponse(w, resp, http.StatusCreated)
}

// UpdateApiToken updates an existing API token for the authenticated user.
// It accepts a request containing updated token name, description, and active status.
//
// Request Body:
//   - Name: Updated name of the token (optional)
//   - Description: Updated description of the token (optional)
//   - Active: Updated status of the token (optional)
//
// Returns:
//   - 200 OK with the updated token details on success
//   - Error response if user authentication, request parsing, or token update fails
func (a *Application) UpdateApiToken(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting api token by id: identifying user", err)
		return
	}

	req := UpdateApiToken{}
	if err := readBody(r.Body, &req); err != nil {
		a.errorLogNResponse(w, "updating token by id: parsing request", err)
		return
	}

	tokenId := entities.Id(r.PathValue("id"))
	if _, err := a.svcGw.Tokens.GetByIdCurUser(r.Context(), cuser, tokenId); err != nil {
		a.errorLogNResponse(w, "updating token by id: validating token id", err)
		return
	}

	updToken, err := a.svcGw.Tokens.Update(r.Context(), cuser, tokenId, req.Name, req.Description, req.Active)
	if err != nil {
		a.errorLogNResponse(w, "updating token by id", err)
		return
	}

	resp := UpdateApiTokenResponse(tokenTApiTokenData(updToken))
	a.successResponse(w, resp, http.StatusOK)
}

// DeleteApiToken deletes an API token for the authenticated user.
// It requires the token ID as a path parameter.
//
// Returns:
//   - 200 OK with the deleted token details on success
//   - Error response if user authentication, token validation, or deletion fails
func (a *Application) DeleteApiToken(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "deleting api token by id: identifying user", err)
	}

	tokenId := entities.Id(r.PathValue("id"))
	token, err := a.svcGw.Tokens.GetByIdCurUser(r.Context(), cuser, tokenId)
	if err != nil {
		a.errorLogNResponse(w, "deleting token by id: validating token id", err)
		return
	}

	if err := a.svcGw.Tokens.Delete(r.Context(), cuser, tokenId); err != nil {
		a.errorLogNResponse(w, "deleting token by id", err)
		return
	}

	resp := DeleteApiTokenResponse(tokenTApiTokenData(token))
	a.successResponse(w, resp, http.StatusOK)
}
