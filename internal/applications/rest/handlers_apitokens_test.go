package rest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/entities"
)

// --- GetApiTokens ---

func TestGetApiTokens_NoUser(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tokens", nil)
	w := httptest.NewRecorder()
	ta.app.GetApiTokens(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetApiTokens_ServiceError(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	ta.tokensRepo.On("GetAll", mock.Anything, mock.Anything).Return(nil, entities.ErrDatabase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tokens", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetApiTokens(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

func TestGetApiTokens_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	token := testToken(user.ID)
	ta.tokensRepo.On("GetAll", mock.Anything, mock.Anything).Return([]entities.ApiToken{token}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tokens", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetApiTokens(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	ta.tokensRepo.AssertExpectations(t)
}

func TestGetApiTokens_Empty(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	ta.tokensRepo.On("GetAll", mock.Anything, mock.Anything).Return([]entities.ApiToken{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tokens", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetApiTokens(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

// --- GetApiTokenById ---

func TestGetApiTokenById_NoUser(t *testing.T) {
	ta := newTestApp(t)
	id := entities.NewId()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tokens/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	w := httptest.NewRecorder()
	ta.app.GetApiTokenById(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetApiTokenById_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	ta.tokensRepo.On("GetById", mock.Anything, id).Return(entities.ApiToken{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tokens/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetApiTokenById(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

func TestGetApiTokenById_OtherUserToken(t *testing.T) {
	ta := newTestApp(t)
	user := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
	token := testToken(entities.NewId()) // owned by a different user

	ta.tokensRepo.On("GetById", mock.Anything, token.ID).Return(token, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tokens/"+token.ID.String(), nil)
	req.SetPathValue("id", token.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetApiTokenById(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

func TestGetApiTokenById_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	token := testToken(user.ID)

	ta.tokensRepo.On("GetById", mock.Anything, token.ID).Return(token, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tokens/"+token.ID.String(), nil)
	req.SetPathValue("id", token.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetApiTokenById(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

// --- CreateApiToken ---
// NOTE: this handler lacks return after userFromContext error.

func TestCreateApiToken_NoUser(t *testing.T) {
	// Use invalid JSON body to short-circuit at body parsing (which has a proper return).
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{bad json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tokens", body)
	w := httptest.NewRecorder()
	require.NotPanics(t, func() { ta.app.CreateApiToken(w, req) })
	// first WriteHeader(403) wins
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateApiToken_InvalidBody(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tokens", body)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.CreateApiToken(w, req)
	// wrapped in ErrValidation → 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateApiToken_EmptyName(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{"name": "", "expire_in": 1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tokens", body)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.CreateApiToken(w, req)
	// service validates name → ErrValidation → 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateApiToken_ZeroExpireIn(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{"name": "mytoken", "expire_in": 0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tokens", body)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.CreateApiToken(w, req)
	// service rejects expire_in < 1 → ErrValidation → 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateApiToken_WithDescription(t *testing.T) {
	ta := newTestApp(t)
	user := testUserFull()
	ta.tokensRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"name": "mytoken", "expire_in": 30, "description": "for CI"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tokens", body)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.CreateApiToken(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

func TestCreateApiToken_Success(t *testing.T) {
	ta := newTestApp(t)
	// testUserFull() includes Login, required for token.Validate() → Owner.Validate()
	user := testUserFull()
	ta.tokensRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"name": "mytoken", "expire_in": 30}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tokens", body)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.CreateApiToken(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

// --- UpdateApiToken ---

func TestUpdateApiToken_NoUser(t *testing.T) {
	ta := newTestApp(t)
	id := entities.NewId()
	body := bytes.NewBufferString(`{"name": "updated"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tokens/"+id.String(), body)
	req.SetPathValue("id", id.String())
	w := httptest.NewRecorder()
	ta.app.UpdateApiToken(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateApiToken_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	ta.tokensRepo.On("GetById", mock.Anything, id).Return(entities.ApiToken{}, entities.ErrNotFound)

	body := bytes.NewBufferString(`{"name": "updated"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tokens/"+id.String(), body)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.UpdateApiToken(w, req)

	// service wraps fmt.Errorf("%w: %w", ErrValidation, ErrNotFound); statusFErr matches ErrNotFound first → 404
	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

func TestUpdateApiToken_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	token := testToken(user.ID)
	updated := token
	updated.Name = "renamed"

	ta.tokensRepo.On("GetById", mock.Anything, token.ID).Return(token, nil)
	ta.tokensRepo.On("Update", mock.Anything, mock.Anything).Return(updated, nil)

	body := bytes.NewBufferString(`{"name": "renamed"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tokens/"+token.ID.String(), body)
	req.SetPathValue("id", token.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.UpdateApiToken(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

func TestUpdateApiToken_Deactivate(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	token := testToken(user.ID)
	token.Active = true
	deactivated := token
	deactivated.Active = false

	ta.tokensRepo.On("GetById", mock.Anything, token.ID).Return(token, nil)
	ta.tokensRepo.On("Update", mock.Anything, mock.Anything).Return(deactivated, nil)

	body := bytes.NewBufferString(`{"active": false}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tokens/"+token.ID.String(), body)
	req.SetPathValue("id", token.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.UpdateApiToken(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

func TestUpdateApiToken_ReactivationForbidden(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	token := testToken(user.ID)
	token.Active = false // inactive token

	ta.tokensRepo.On("GetById", mock.Anything, token.ID).Return(token, nil)

	active := true
	body := bytes.NewBufferString(`{"active": true}`)
	_ = active
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tokens/"+token.ID.String(), body)
	req.SetPathValue("id", token.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.UpdateApiToken(w, req)

	// service forbids reactivating inactive tokens → ErrValidation → 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

// --- DeleteApiToken ---
// NOTE: this handler lacks return after userFromContext error.

func TestDeleteApiToken_NoUser(t *testing.T) {
	// DeleteApiToken lacks return after userFromContext error; it proceeds to call the service.
	// Provide a valid token ID so the service calls GetById; mock it to avoid panic.
	ta := newTestApp(t)
	id := entities.NewId()
	ta.tokensRepo.On("GetById", mock.Anything, id).Return(entities.ApiToken{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tokens/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	w := httptest.NewRecorder()
	require.NotPanics(t, func() { ta.app.DeleteApiToken(w, req) })
	// first WriteHeader(403) wins despite subsequent errorLogNResponse calls
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteApiToken_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	ta.tokensRepo.On("GetById", mock.Anything, id).Return(entities.ApiToken{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tokens/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.DeleteApiToken(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

func TestDeleteApiToken_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	token := testToken(user.ID)

	ta.tokensRepo.On("GetById", mock.Anything, token.ID).Return(token, nil)
	ta.tokensRepo.On("Delete", mock.Anything, mock.Anything, token.ID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tokens/"+token.ID.String(), nil)
	req.SetPathValue("id", token.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	require.NotPanics(t, func() { ta.app.DeleteApiToken(w, req) })

	assert.Equal(t, http.StatusOK, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}

func TestDeleteApiToken_OtherUserToken(t *testing.T) {
	ta := newTestApp(t)
	user := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
	token := testToken(entities.NewId()) // owned by a different user

	ta.tokensRepo.On("GetById", mock.Anything, token.ID).Return(token, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tokens/"+token.ID.String(), nil)
	req.SetPathValue("id", token.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.DeleteApiToken(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	ta.tokensRepo.AssertExpectations(t)
}
