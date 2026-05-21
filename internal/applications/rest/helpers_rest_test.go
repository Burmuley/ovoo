package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/entities"
)

func TestStatusFErr_NotFound(t *testing.T) {
	assert.Equal(t, http.StatusNotFound, statusFErr(entities.ErrNotFound))
}

func TestStatusFErr_NotFoundWrapped(t *testing.T) {
	err := fmt.Errorf("entity missing: %w", entities.ErrNotFound)
	assert.Equal(t, http.StatusNotFound, statusFErr(err))
}

func TestStatusFErr_Validation(t *testing.T) {
	assert.Equal(t, http.StatusBadRequest, statusFErr(entities.ErrValidation))
}

func TestStatusFErr_ValidationWrapped(t *testing.T) {
	err := fmt.Errorf("%w: invalid id", entities.ErrValidation)
	assert.Equal(t, http.StatusBadRequest, statusFErr(err))
}

func TestStatusFErr_DuplicateEntry(t *testing.T) {
	assert.Equal(t, http.StatusBadRequest, statusFErr(entities.ErrDuplicateEntry))
}

func TestStatusFErr_DuplicateEntryWrapped(t *testing.T) {
	err := fmt.Errorf("create failed: %w: email exists", entities.ErrDuplicateEntry)
	assert.Equal(t, http.StatusBadRequest, statusFErr(err))
}

func TestStatusFErr_NotAuthorized(t *testing.T) {
	assert.Equal(t, http.StatusForbidden, statusFErr(entities.ErrNotAuthorized))
}

func TestStatusFErr_NotAuthorizedWrapped(t *testing.T) {
	err := fmt.Errorf("operation denied: %w", entities.ErrNotAuthorized)
	assert.Equal(t, http.StatusForbidden, statusFErr(err))
}

func TestStatusFErr_GenericError(t *testing.T) {
	assert.Equal(t, http.StatusInternalServerError, statusFErr(errors.New("unexpected error")))
}

func TestStatusFErr_DatabaseError(t *testing.T) {
	// ErrDatabase has no special mapping, falls through to 500
	assert.Equal(t, http.StatusInternalServerError, statusFErr(entities.ErrDatabase))
}

func TestUserFromContext_Present(t *testing.T) {
	user := entities.User{ID: entities.NewId(), Type: entities.AdminUser}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withUser(req, user)

	result, err := userFromContext(req)
	require.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.Type, result.Type)
}

func TestUserFromContext_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := userFromContext(req)
	require.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
}

func TestReadBody_ValidJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	body := io.NopCloser(bytes.NewBufferString(`{"name": "hello"}`))
	var result payload
	err := readBody(body, &result)
	require.NoError(t, err)
	assert.Equal(t, "hello", result.Name)
}

func TestReadBody_ValidJSON_EmptyObject(t *testing.T) {
	type payload struct {
		Name string `json:"name,omitempty"`
	}
	body := io.NopCloser(bytes.NewBufferString(`{}`))
	var result payload
	err := readBody(body, &result)
	require.NoError(t, err)
	assert.Equal(t, "", result.Name)
}

func TestReadBody_InvalidJSON(t *testing.T) {
	body := io.NopCloser(bytes.NewBufferString(`{invalid json`))
	var result map[string]any
	err := readBody(body, &result)
	require.Error(t, err)
}

func TestReadBody_EmptyBody(t *testing.T) {
	body := io.NopCloser(bytes.NewReader(nil))
	var result map[string]any
	err := readBody(body, &result)
	require.Error(t, err)
}

func TestErrorLogNResponse_Body(t *testing.T) {
	app := &Application{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	w := httptest.NewRecorder()
	app.errorLogNResponse(w, "test op", entities.ErrNotFound)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var body ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Errors, 1)
	assert.Equal(t, "404", body.Errors[0].Status)
	assert.NotEmpty(t, body.Errors[0].Detail)
}

func TestReadBody_ArrayJSON(t *testing.T) {
	body := io.NopCloser(bytes.NewBufferString(`[1,2,3]`))
	var result []int
	err := readBody(body, &result)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, result)
}
