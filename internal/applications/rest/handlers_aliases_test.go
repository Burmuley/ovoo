package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/entities"
)

// --- GetAliases ---

func TestGetAliases_NoUser(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/aliases", nil)
	w := httptest.NewRecorder()
	ta.app.GetAliases(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAliases_InvalidFilter(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/aliases?page=notanumber", nil)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.GetAliases(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAliases_ServiceError(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	ta.addrRepo.On("GetAll", mock.Anything, mock.Anything).
		Return([]entities.Address{}, entities.PaginationMetadata{}, fmt.Errorf("db failure"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aliases", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetAliases(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestGetAliases_EmptyList(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	ta.addrRepo.On("GetAll", mock.Anything, mock.Anything).
		Return([]entities.Address{}, entities.PaginationMetadata{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aliases", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetAliases(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestGetAliases_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	alias := testAlias(user.ID)

	ta.addrRepo.On("GetAll", mock.Anything, mock.Anything).
		Return([]entities.Address{alias}, entities.PaginationMetadata{TotalRecords: 1}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aliases", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetAliases(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	ta.addrRepo.AssertExpectations(t)
}

func TestGetAliases_ResponseBody(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	alias := testAlias(user.ID)

	ta.addrRepo.On("GetAll", mock.Anything, mock.Anything).
		Return([]entities.Address{alias}, entities.PaginationMetadata{TotalRecords: 1, PageSize: 5}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aliases", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetAliases(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body GetAliasesResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Aliases, 1)
	assert.Equal(t, alias.ID.String(), body.Aliases[0].Id)
	assert.Equal(t, string(alias.Email), string(body.Aliases[0].Email))
	assert.Equal(t, 1, body.PaginationMetadata.TotalRecords)
}

// --- GetAliaseById ---

func TestGetAliaseById_NoUser(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/aliases/someid", nil)
	req.SetPathValue("id", "someid")
	w := httptest.NewRecorder()
	ta.app.GetAliaseById(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAliaseById_InvalidID(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/aliases/not-a-ulid", nil)
	req.SetPathValue("id", "not-a-ulid")
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.GetAliaseById(w, req)
	// invalid ULID → service returns ErrValidation → 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAliaseById_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	ta.addrRepo.On("GetById", mock.Anything, id).Return(entities.Address{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aliases/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetAliaseById(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestGetAliaseById_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	alias := testAlias(user.ID)

	ta.addrRepo.On("GetById", mock.Anything, alias.ID).Return(alias, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aliases/"+alias.ID.String(), nil)
	req.SetPathValue("id", alias.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetAliaseById(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

// --- CreateAlias ---

func TestCreateAlias_NoUser(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{"protected_address_id": "someid", "metadata": {}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/aliases", body)
	w := httptest.NewRecorder()
	ta.app.CreateAlias(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateAlias_InvalidBody(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{invalid_json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/aliases", body)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.CreateAlias(w, req)
	// raw JSON error, not wrapped in ErrValidation → 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateAlias_ProtectedAddrNotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	prAddrID := entities.NewId()

	ta.addrRepo.On("GetById", mock.Anything, prAddrID).Return(entities.Address{}, entities.ErrNotFound)

	body := bytes.NewBufferString(`{"protected_address_id": "` + prAddrID.String() + `", "metadata": {}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/aliases", body)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.CreateAlias(w, req)

	// service wraps fmt.Errorf("%w: %w", ErrValidation, ErrNotFound); statusFErr checks ErrNotFound first → 404
	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestCreateAlias_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	prAddr := testProtectedAddr(user.ID)
	domainId := entities.NewId()

	ta.addrRepo.On("GetById", mock.Anything, prAddr.ID).Return(prAddr, nil)
	ta.addrRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"protected_address_id": "` + prAddr.ID.String() + `", "domain_id": "` + domainId.String() + `", "metadata": {}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/aliases", body)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.CreateAlias(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestCreateAlias_Success_WithMetadata(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	prAddr := testProtectedAddr(user.ID)
	domainId := entities.NewId()

	ta.addrRepo.On("GetById", mock.Anything, prAddr.ID).Return(prAddr, nil)
	ta.addrRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"protected_address_id": "` + prAddr.ID.String() + `", "domain_id": "` + domainId.String() + `", "metadata": {"comment": "test", "service_name": "svc"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/aliases", body)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.CreateAlias(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

// --- DeleteAlias ---

func TestDeleteAlias_NoUser(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/aliases/someid", nil)
	req.SetPathValue("id", "someid")
	w := httptest.NewRecorder()
	ta.app.DeleteAlias(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteAlias_InvalidID(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/aliases/not-valid", nil)
	req.SetPathValue("id", "not-valid")
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.DeleteAlias(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteAlias_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	ta.addrRepo.On("GetById", mock.Anything, id).Return(entities.Address{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/aliases/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.DeleteAlias(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestDeleteAlias_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	alias := testAlias(user.ID)

	ta.addrRepo.On("GetById", mock.Anything, alias.ID).Return(alias, nil)
	// deleteChainsForAliasIds: two GetByFilters calls, then BatchDelete, then BatchDeleteById for reply aliases
	ta.chainRepo.On("GetByFilters", mock.Anything, mock.Anything).Return(nil, nil).Twice()
	ta.chainRepo.On("BatchDelete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// BatchDeleteById called twice: once for empty chain addrs, once for the alias itself
	ta.addrRepo.On("BatchDeleteById", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/aliases/"+alias.ID.String(), nil)
	req.SetPathValue("id", alias.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	require.NotPanics(t, func() { ta.app.DeleteAlias(w, req) })

	assert.Equal(t, http.StatusNoContent, w.Code)
	ta.addrRepo.AssertExpectations(t)
	ta.chainRepo.AssertExpectations(t)
}

// --- UpdateAlias (additional corner cases) ---

func TestUpdateAlias_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	ta.addrRepo.On("GetById", mock.Anything, id).Return(entities.Address{}, entities.ErrNotFound)

	body := bytes.NewBufferString(`{"active": true}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/aliases/"+id.String(), body)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.UpdateAlias(w, req)

	// service wraps fmt.Errorf("%w: %w", ErrValidation, ErrNotFound); statusFErr matches ErrNotFound first → 404
	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestUpdateAlias_Unauthorized(t *testing.T) {
	ta := newTestApp(t)
	// MilterUser cannot update any alias
	milterUser := entities.User{ID: entities.NewId(), Type: entities.MilterUser}
	alias := testAlias(entities.NewId()) // owned by someone else

	ta.addrRepo.On("GetById", mock.Anything, alias.ID).Return(alias, nil)

	body := bytes.NewBufferString(`{"active": true}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/aliases/"+alias.ID.String(), body)
	req.SetPathValue("id", alias.ID.String())
	req = withUser(req, milterUser)
	w := httptest.NewRecorder()
	ta.app.UpdateAlias(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestUpdateAlias_InvalidBody(t *testing.T) {
	ta := newTestApp(t)
	id := entities.NewId()
	body := bytes.NewBufferString(`{not valid json`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/aliases/"+id.String(), body)
	req.SetPathValue("id", id.String())
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.UpdateAlias(w, req)
	// raw JSON error, not wrapped → 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
