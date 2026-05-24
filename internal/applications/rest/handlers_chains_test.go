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

// --- getChainByHash ---

func TestGetChainByHash_NoUser(t *testing.T) {
	ta := newTestApp(t)
	hash := entities.NewHash("a@b.com", "c@d.com")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chains/"+hash.String(), nil)
	req.SetPathValue("hash", hash.String())
	w := httptest.NewRecorder()
	ta.app.getChainByHash(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetChainByHash_UnauthorizedUser(t *testing.T) {
	// RegularUser cannot get chains; only AdminUser or MilterUser are allowed.
	ta := newTestApp(t)
	regularUser := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
	hash := entities.NewHash("a@b.com", "c@d.com")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chains/"+hash.String(), nil)
	req.SetPathValue("hash", hash.String())
	req = withUser(req, regularUser)
	w := httptest.NewRecorder()
	ta.app.getChainByHash(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetChainByHash_InvalidHash(t *testing.T) {
	ta := newTestApp(t)
	// hash is too short → fails Hash.Validate()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chains/short", nil)
	req.SetPathValue("hash", "short")
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.getChainByHash(w, req)
	// service returns ErrValidation → 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetChainByHash_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	hash := entities.NewHash("from@example.com", "to@example.com")

	ta.chainRepo.On("GetByHash", mock.Anything, hash).Return(entities.Chain{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chains/"+hash.String(), nil)
	req.SetPathValue("hash", hash.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.getChainByHash(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.chainRepo.AssertExpectations(t)
}

func TestGetChainByHash_InactiveDestination(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	chain := testChain()
	chain.OrigToAddress.Active = false // inactive alias → chain service returns ErrNotFound

	ta.chainRepo.On("GetByHash", mock.Anything, chain.Hash).Return(chain, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chains/"+chain.Hash.String(), nil)
	req.SetPathValue("hash", chain.Hash.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.getChainByHash(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.chainRepo.AssertExpectations(t)
}

func TestGetChainByHash_InactiveOwner(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	chain := testChain()
	chain.OrigToAddress.Owner.Active = false // owner inactive → ErrNotFound

	ta.chainRepo.On("GetByHash", mock.Anything, chain.Hash).Return(chain, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chains/"+chain.Hash.String(), nil)
	req.SetPathValue("hash", chain.Hash.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.getChainByHash(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.chainRepo.AssertExpectations(t)
}

func TestGetChainByHash_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	chain := testChain()

	ta.chainRepo.On("GetByHash", mock.Anything, chain.Hash).Return(chain, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chains/"+chain.Hash.String(), nil)
	req.SetPathValue("hash", chain.Hash.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.getChainByHash(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	ta.chainRepo.AssertExpectations(t)
}

// --- CreateChain ---

func TestGetChainByHash_MilterUserSuccess(t *testing.T) {
	// MilterUser is also allowed to get chains (admin and milter only).
	ta := newTestApp(t)
	milterUser := entities.User{ID: entities.NewId(), Type: entities.MilterUser}
	chain := testChain()

	ta.chainRepo.On("GetByHash", mock.Anything, chain.Hash).Return(chain, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chains/"+chain.Hash.String(), nil)
	req.SetPathValue("hash", chain.Hash.String())
	req = withUser(req, milterUser)
	w := httptest.NewRecorder()
	ta.app.getChainByHash(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ta.chainRepo.AssertExpectations(t)
}

func TestCreateChain_NoUser(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{"from_email": "from@example.com", "to_email": "to@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chains", body)
	w := httptest.NewRecorder()
	ta.app.CreateChain(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateChain_UnauthorizedUser(t *testing.T) {
	// RegularUser cannot create chains.
	ta := newTestApp(t)
	regularUser := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
	body := bytes.NewBufferString(`{"from_email": "from@example.com", "to_email": "to@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chains", body)
	req = withUser(req, regularUser)
	w := httptest.NewRecorder()
	ta.app.CreateChain(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateChain_InvalidBody(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{bad json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chains", body)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.CreateChain(w, req)
	// not wrapped in ErrValidation → 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateChain_ExistingChain(t *testing.T) {
	// If a chain already exists for the hash, the service returns it directly.
	ta := newTestApp(t)
	user := testUser()
	chain := testChain()
	fromEmail := string(chain.OrigFromAddress.Email)
	toEmail := string(chain.OrigToAddress.Email)
	hash := entities.NewHash(fromEmail, toEmail)

	ta.chainRepo.On("GetByHash", mock.Anything, hash).Return(chain, nil)

	body := bytes.NewBufferString(`{"from_email": "` + fromEmail + `", "to_email": "` + toEmail + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chains", body)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	require.NotPanics(t, func() { ta.app.CreateChain(w, req) })

	assert.Equal(t, http.StatusCreated, w.Code)
	ta.chainRepo.AssertExpectations(t)
}

func TestCreateChain_DestinationNotFound(t *testing.T) {
	// No existing chain and no alias with toEmail → ErrNotFound → 404
	ta := newTestApp(t)
	user := testUser()
	fromEmail := "from@external.com"
	toEmail := "nonexistent@test.com"
	hash := entities.NewHash(fromEmail, toEmail)

	ta.chainRepo.On("GetByHash", mock.Anything, hash).Return(entities.Chain{}, entities.ErrNotFound)
	ta.addrRepo.On("GetByEmail", mock.Anything, entities.Email(toEmail)).Return([]entities.Address{}, nil)

	body := bytes.NewBufferString(`{"from_email": "` + fromEmail + `", "to_email": "` + toEmail + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chains", body)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.CreateChain(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.chainRepo.AssertExpectations(t)
	ta.addrRepo.AssertExpectations(t)
}

// --- DeleteChain ---

func TestDeleteChain_NoUser(t *testing.T) {
	ta := newTestApp(t)
	hash := entities.NewHash("a@b.com", "c@d.com")
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/chains/"+hash.String(), nil)
	req.SetPathValue("hash", hash.String())
	w := httptest.NewRecorder()
	ta.app.DeleteChain(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteChain_UnauthorizedUser(t *testing.T) {
	ta := newTestApp(t)
	regularUser := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
	hash := entities.NewHash("a@b.com", "c@d.com")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/chains/"+hash.String(), nil)
	req.SetPathValue("hash", hash.String())
	req = withUser(req, regularUser)
	w := httptest.NewRecorder()
	ta.app.DeleteChain(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteChain_InvalidHash(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/chains/tooshort", nil)
	req.SetPathValue("hash", "tooshort")
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.DeleteChain(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteChain_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	hash := entities.NewHash("from@example.com", "to@example.com")

	ta.chainRepo.On("Delete", mock.Anything, mock.Anything, hash).Return(entities.Chain{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/chains/"+hash.String(), nil)
	req.SetPathValue("hash", hash.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.DeleteChain(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.chainRepo.AssertExpectations(t)
}

func TestDeleteChain_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	hash := entities.NewHash("from@example.com", "to@example.com")

	ta.chainRepo.On("Delete", mock.Anything, mock.Anything, hash).Return(entities.Chain{}, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/chains/"+hash.String(), nil)
	req.SetPathValue("hash", hash.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.DeleteChain(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	ta.chainRepo.AssertExpectations(t)
}
