package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Burmuley/ovoo/internal/controllers/rest/middleware"
	"github.com/Burmuley/ovoo/internal/entities"
)

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

func getUserFromContext(r *http.Request) (entities.User, error) {
	userraw := r.Context().Value(middleware.UserContextKey("user"))
	if userraw == nil {
		return entities.User{}, errors.New("unable to get user")
	}

	return userraw.(entities.User), nil
}
