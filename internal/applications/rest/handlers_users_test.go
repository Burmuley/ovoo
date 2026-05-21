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

// --- GetUserProfile ---
// NOTE: this handler does NOT return after userFromContext error; it writes a 403 error
// response and then continues to write a success response with the zero-value user.

func TestGetUserProfile_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetUserProfile(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserProfile_NoUser(t *testing.T) {
	// Bug: handler lacks return after userFromContext error; 403 is written first
	// but execution continues and writes a second response with the zero-value user.
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	w := httptest.NewRecorder()
	require.NotPanics(t, func() { ta.app.GetUserProfile(w, req) })
	// first WriteHeader(403) wins
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- GetUserById ---
// NOTE: this handler also lacks return after userFromContext error.

func TestGetUserById_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	target := entities.User{
		ID:    entities.NewId(),
		Login: "other@example.com",
		Type:  entities.RegularUser,
	}
	ta.usersRepo.On("GetById", mock.Anything, target.ID).Return(target, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+target.ID.String(), nil)
	req.SetPathValue("id", target.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetUserById(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ta.usersRepo.AssertExpectations(t)
}

func TestGetUserById_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	ta.usersRepo.On("GetById", mock.Anything, id).Return(entities.User{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetUserById(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.usersRepo.AssertExpectations(t)
}

func TestGetUserById_NoUser(t *testing.T) {
	// Bug: handler lacks return after userFromContext error.
	// With empty user, GetById validates ID via service before calling repo.
	ta := newTestApp(t)
	id := entities.NewId()
	ta.usersRepo.On("GetById", mock.Anything, id).Return(entities.User{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	w := httptest.NewRecorder()
	require.NotPanics(t, func() { ta.app.GetUserById(w, req) })
	// first WriteHeader(403) wins
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- GetUsers ---

func TestGetUsers_ServiceError(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	ta.usersRepo.On("GetAll", mock.Anything, mock.Anything).
		Return([]entities.User{}, entities.PaginationMetadata{}, entities.ErrDatabase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetUsers(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	ta.usersRepo.AssertExpectations(t)
}

func TestGetUsers_NoUser(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	ta.app.GetUsers(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetUsers_InvalidFilter(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=xyz", nil)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.GetUsers(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUsers_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	other := entities.User{ID: entities.NewId(), Login: "x@example.com", Type: entities.RegularUser}

	ta.usersRepo.On("GetAll", mock.Anything, mock.Anything).
		Return([]entities.User{other}, entities.PaginationMetadata{TotalRecords: 1}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.GetUsers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	ta.usersRepo.AssertExpectations(t)
}

// --- CreateUser ---

func TestCreateUser_NoUser(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{"login": "new@example.com", "type": "regular", "first_name": "A", "last_name": "B"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	w := httptest.NewRecorder()
	ta.app.CreateUser(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateUser_InvalidBody(t *testing.T) {
	ta := newTestApp(t)
	body := bytes.NewBufferString(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.CreateUser(w, req)
	// wrapped in ErrValidation → 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateUser_InvalidType(t *testing.T) {
	ta := newTestApp(t)
	// "superadmin" is not a valid user type → userTypeFStr returns 99 → User.Validate fails
	body := bytes.NewBufferString(`{"login": "new@example.com", "type": "superadmin", "first_name": "A", "last_name": "B"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	req = withUser(req, testUser())
	w := httptest.NewRecorder()
	ta.app.CreateUser(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateUser_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	ta.usersRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"login": "newuser@example.com", "type": "regular", "first_name": "Alice", "last_name": "Smith"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.CreateUser(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	ta.usersRepo.AssertExpectations(t)
}

// --- UpdateUser ---
// NOTE: this handler lacks return after userFromContext error.

func TestUpdateUser_NoUser(t *testing.T) {
	ta := newTestApp(t)
	id := entities.NewId()
	// Invalid JSON body causes early return before repo is called
	body := bytes.NewBufferString(`{bad json`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+id.String(), body)
	req.SetPathValue("id", id.String())
	w := httptest.NewRecorder()
	require.NotPanics(t, func() { ta.app.UpdateUser(w, req) })
	// first WriteHeader(403) wins
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateUser_WithType(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	target := entities.User{
		ID:    entities.NewId(),
		Login: "other@example.com",
		Type:  entities.RegularUser,
	}
	ta.usersRepo.On("GetById", mock.Anything, target.ID).Return(target, nil)
	ta.usersRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"type": "regular"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+target.ID.String(), body)
	req.SetPathValue("id", target.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.UpdateUser(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	ta.usersRepo.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	target := entities.User{
		ID:    entities.NewId(),
		Login: "other@example.com",
		Type:  entities.RegularUser,
	}
	ta.usersRepo.On("GetById", mock.Anything, target.ID).Return(target, nil)
	ta.usersRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"first_name": "Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+target.ID.String(), body)
	req.SetPathValue("id", target.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.UpdateUser(w, req)

	// handler uses 201 for successful update (code quirk)
	assert.Equal(t, http.StatusCreated, w.Code)
	ta.usersRepo.AssertExpectations(t)
}

func TestUpdateUser_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	ta.usersRepo.On("GetById", mock.Anything, id).Return(entities.User{}, entities.ErrNotFound)

	body := bytes.NewBufferString(`{"first_name": "Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+id.String(), body)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.UpdateUser(w, req)

	// service wraps fmt.Errorf("%w: %w", ErrValidation, ErrNotFound); statusFErr matches ErrNotFound first → 404
	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.usersRepo.AssertExpectations(t)
}

// --- DeleteUser ---
// NOTE: this handler lacks return after userFromContext error.

func TestDeleteUser_NoUser(t *testing.T) {
	// With no user, DeleteUser writes 403 first, then tries to delete.
	// Empty ID "" fails ULID validation, causing ErrValidation → errorLogNResponse(400).
	// But 403 was already written.
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/", nil)
	w := httptest.NewRecorder()
	require.NotPanics(t, func() { ta.app.DeleteUser(w, req) })
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteUser_NotFound(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	id := entities.NewId()

	ta.usersRepo.On("GetById", mock.Anything, id).Return(entities.User{}, entities.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.DeleteUser(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ta.usersRepo.AssertExpectations(t)
}

func TestDeleteUser_CannotDeleteSelf(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()

	ta.usersRepo.On("GetById", mock.Anything, user.ID).Return(user, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+user.ID.String(), nil)
	req.SetPathValue("id", user.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	ta.app.DeleteUser(w, req)

	// canDeleteUser returns false if cuser deletes themselves → ErrNotAuthorized → 403
	assert.Equal(t, http.StatusForbidden, w.Code)
	ta.usersRepo.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	ta := newTestApp(t)
	user := testUser()
	target := entities.User{
		ID:    entities.NewId(),
		Login: "target@example.com",
		Type:  entities.RegularUser,
	}

	ta.usersRepo.On("GetById", mock.Anything, target.ID).Return(target, nil)
	// deletePrAddrsForUser: GetAll returns empty
	ta.addrRepo.On("GetAll", mock.Anything, mock.Anything).
		Return([]entities.Address{}, entities.PaginationMetadata{}, nil)
	ta.addrRepo.On("BatchDeleteById", mock.Anything, mock.Anything).Return(nil)
	ta.tokensRepo.On("BatchDeleteForUser", mock.Anything, target.ID).Return(nil)
	ta.usersRepo.On("Delete", mock.Anything, target.ID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+target.ID.String(), nil)
	req.SetPathValue("id", target.ID.String())
	req = withUser(req, user)
	w := httptest.NewRecorder()
	require.NotPanics(t, func() { ta.app.DeleteUser(w, req) })

	// handler uses 201 for successful delete (code quirk)
	assert.Equal(t, http.StatusCreated, w.Code)
	ta.usersRepo.AssertExpectations(t)
	ta.addrRepo.AssertExpectations(t)
	ta.tokensRepo.AssertExpectations(t)
}
