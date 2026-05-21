package rest

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/entities"
)

// --- GetAllPrAddrs ---

func TestGetAllPrAddrs_NoUser(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/praddrs", nil)
	w := httptest.NewRecorder()
	ta.app.GetAllPrAddrs(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAllPrAddrs_InvalidFilter(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/praddrs?page=bad", nil)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.GetAllPrAddrs(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAllPrAddrs_ServiceError(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	ta.addrRepo.On("GetAll", mock.Anything, mock.Anything).
		Return([]entities.Address{}, entities.PaginationMetadata{}, fmt.Errorf("db failure"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/praddrs", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetAllPrAddrs(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestGetAllPrAddrs_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	prAddr := testProtectedAddr(user.ID)

	ta.addrRepo.On("GetAll", mock.Anything, mock.Anything).
		Return([]entities.Address{prAddr}, entities.PaginationMetadata{TotalRecords: 1}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/praddrs", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetAllPrAddrs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	ta.addrRepo.AssertExpectations(t)
}

func TestGetAllPrAddrs_EmptyList(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	ta.addrRepo.On("GetAll", mock.Anything, mock.Anything).
		Return([]entities.Address{}, entities.PaginationMetadata{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/praddrs", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetAllPrAddrs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

// --- GetPrAddrById ---
// NOTE: this handler incorrectly delegates to a.svcGw.Aliases.GetById instead of PrAddrs.GetById.
// Tests reflect the actual (buggy) behavior.

func TestGetPrAddrById_NoUser(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/praddrs/someid", nil)
	req.SetPathValue("id", "someid")
	w := httptest.NewRecorder()
	ta.app.GetPrAddrById(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetPrAddrById_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	// delegates to AliasesService.GetById which calls addrRepo.GetById
	ta.addrRepo.On("GetById", mock.Anything, id).Return(entities.Address{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/praddrs/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetPrAddrById(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestGetPrAddrById_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	prAddr := testProtectedAddr(user.ID)

	ta.addrRepo.On("GetById", mock.Anything, prAddr.ID).Return(prAddr, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/praddrs/"+prAddr.ID.String(), nil)
	req.SetPathValue("id", prAddr.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetPrAddrById(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

// --- CreatePrAddr ---

func TestCreatePrAddr_NoUser(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{"email": "new@example.com", "metadata": {}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/praddrs", body)
	w := httptest.NewRecorder()
	ta.app.CreatePrAddr(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreatePrAddr_InvalidBody(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{invalid_json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/praddrs", body)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.CreatePrAddr(w, req)
	// not wrapped in ErrValidation → 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreatePrAddr_InvalidEmail(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{"email": "not-an-email", "metadata": {}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/praddrs", body)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.CreatePrAddr(w, req)
	// service validates email → ErrValidation → 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePrAddr_DuplicateEntry(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	existing := testProtectedAddr(user.ID)

	// service checks by email first
	ta.addrRepo.On("GetByEmail", mock.Anything, entities.Email("protected@example.com")).
		Return([]entities.Address{existing}, nil)

	body := bytes.NewBufferString(`{"email": "protected@example.com", "metadata": {}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/praddrs", body)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.CreatePrAddr(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestCreatePrAddr_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()

	ta.addrRepo.On("GetByEmail", mock.Anything, entities.Email("new@example.com")).
		Return([]entities.Address{}, nil)
	ta.addrRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"email": "new@example.com", "metadata": {}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/praddrs", body)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.CreatePrAddr(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

// --- DeletePrAddr ---

func TestDeletePrAddr_NoUser(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/praddrs/someid", nil)
	req.SetPathValue("id", "someid")
	w := httptest.NewRecorder()
	ta.app.DeletePrAddr(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeletePrAddr_InvalidID(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/praddrs/bad-id", nil)
	req.SetPathValue("id", "bad-id")
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.DeletePrAddr(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePrAddr_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	ta.addrRepo.On("GetById", mock.Anything, id).Return(entities.Address{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/praddrs/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.DeletePrAddr(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestDeletePrAddr_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	prAddr := testProtectedAddr(user.ID)

	ta.addrRepo.On("GetById", mock.Anything, prAddr.ID).Return(prAddr, nil)
	// deleteAliasesForPrAddr: GetAll for aliases returns empty
	ta.addrRepo.On("GetAll", mock.Anything, mock.Anything).
		Return([]entities.Address{}, entities.PaginationMetadata{}, nil)
	// DeleteById for the protected address itself
	ta.addrRepo.On("DeleteById", mock.Anything, prAddr.ID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/praddrs/"+prAddr.ID.String(), nil)
	req.SetPathValue("id", prAddr.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	require.NotPanics(t, func() { ta.app.DeletePrAddr(w, req) })

	assert.Equal(t, http.StatusNoContent, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

// --- UpdatePrAddr (additional corner cases) ---

func TestUpdatePrAddr_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	ta.addrRepo.On("GetById", mock.Anything, id).Return(entities.Address{}, entities.ErrNotFound)

	body := bytes.NewBufferString(`{"active": false}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/praddrs/"+id.String(), body)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.UpdatePrAddr(w, req)

	// service wraps fmt.Errorf("%w: %w", ErrValidation, ErrNotFound); statusFErr matches ErrNotFound first → 404
	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestUpdatePrAddr_Unauthorized(t *testing.T) {
	ta := newTestApp(t)
	milterUser := entities.User{ID: entities.NewId(), Type: entities.MilterUser}
	prAddr := testProtectedAddr(entities.NewId()) // owned by someone else

	ta.addrRepo.On("GetById", mock.Anything, prAddr.ID).Return(prAddr, nil)

	body := bytes.NewBufferString(`{"active": false}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/praddrs/"+prAddr.ID.String(), body)
	req.SetPathValue("id", prAddr.ID.String())
	req = withUser(req, milterUser)
	w := httptest.NewRecorder()
	ta.app.UpdatePrAddr(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	ta.addrRepo.AssertExpectations(t)
}

func TestUpdatePrAddr_InvalidBody(t *testing.T) {
	ta := newTestApp(t)
	id := entities.NewId()
	body := bytes.NewBufferString(`{not json`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/praddrs/"+id.String(), body)
	req.SetPathValue("id", id.String())
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.UpdatePrAddr(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
