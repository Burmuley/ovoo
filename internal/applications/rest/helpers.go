package rest

import (
	"encoding/json"
	"errors"
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
	userraw := r.Context().Value(middleware.UserContextKey("user"))
	if userraw == nil {
		return entities.User{}, errors.New("unable to get user")
	}

	return userraw.(entities.User), nil
}

// mapKeys extracts and returns all keys from a map as a slice.
// It takes a map with comparable keys and any values, and returns a slice containing all the keys.
// This is useful when you need to process or iterate over just the keys of a map.
func mapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	return keys
}
