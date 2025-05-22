package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Burmuley/ovoo/internal/applications/rest/middleware"
	"github.com/Burmuley/ovoo/internal/entities"
)

// readBody reads and unmarshals JSON from an HTTP request body into the provided data structure.
// It takes an io.ReadCloser (typically request.Body) and a destination interface{} to unmarshal the JSON into.
// Returns an error if reading the body or unmarshaling the JSON fails.
func readBody(body io.ReadCloser, data any) error {
	rawBody, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(rawBody, data); err != nil {
		return err
	}

	return nil
}

// userFromContext extracts the User entity from the HTTP request context.
// This function is designed to work with the middleware.UserContextKey to retrieve
// the authenticated user information that was previously stored in the request context.
// Returns the user entity and nil if successful, or an empty user and error if the user
// cannot be found in the context.
func userFromContext(r *http.Request) (entities.User, error) {
	user := r.Context().Value(middleware.UserContextKey("user"))
	if user == nil {
		return entities.User{}, fmt.Errorf("%w: no user info in headers", entities.ErrNotAuthorized)
	}

	return user.(entities.User), nil
}

// statusFErr determines the appropriate HTTP status code based on the given error.
// It maps specific error types to corresponding HTTP status codes.
//
// Parameters:
//   - err: The error to be evaluated.
//
// Returns:
//
//	An integer representing the HTTP status code.
//
// The function checks for the following error types:
//   - entities.ErrNotFound: Returns http.StatusNotFound (404)
//   - entities.ErrValidation: Returns http.StatusBadRequest (400)
//   - entities.ErrDuplicateEntry: Returns http.StatusBadRequest (400)
//
// For any other error types, it returns http.StatusInternalServerError (500).
func statusFErr(err error) int {
	if errors.Is(err, entities.ErrNotFound) {
		return http.StatusNotFound
	}

	if errors.Is(err, entities.ErrValidation) {
		return http.StatusBadRequest
	}

	if errors.Is(err, entities.ErrDuplicateEntry) {
		return http.StatusBadRequest
	}

	if errors.Is(err, entities.ErrNotAuthorized) {
		return http.StatusForbidden
	}

	return http.StatusInternalServerError
}
