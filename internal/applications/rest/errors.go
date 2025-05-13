package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
)

const (
	ERR_CODE_NOT_FOUND = iota + 10
	ERR_CODE_DUPLICATE_ENTRY
	ERR_CODE_BACKEND_ERROR
	ERR_CODE_UNAUTHORIZED
	ERR_CODE_UNAUTHENTICATED
	ERR_CODE_BAD_REQUEST
	ERR_CODE_UNPROCESSABLE
)

var (
	ErrUnauthenticated     = ErrorResponse{Id: ERR_CODE_UNAUTHENTICATED}
	ErrUnauthorized        = ErrorResponse{Id: ERR_CODE_UNAUTHORIZED}
	ErrNotFound            = ErrorResponse{Id: ERR_CODE_NOT_FOUND}
	ErrBadRequest          = ErrorResponse{Id: ERR_CODE_BAD_REQUEST}
	ErrDuplicateEntry      = ErrorResponse{Id: ERR_CODE_DUPLICATE_ENTRY, Msg: "duplicate entry"}
	ErrBackendError        = ErrorResponse{Id: ERR_CODE_BACKEND_ERROR}
	ErrUnprocessableEntity = ErrorResponse{Id: ERR_CODE_UNPROCESSABLE}
)

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

// errorLogNResponse logs an error, creates an error response, and writes it to the HTTP response writer.
// It takes the following parameters:
//   - w: The http.ResponseWriter to write the error response to.
//   - operation: A string describing the operation that caused the error.
//   - opErr: The error that occurred during the operation.
//   - logger: The slog.Logger to use for logging the error.
//
// This function performs the following steps:
//  1. Logs the error using slog.Error.
//  2. Determines the appropriate HTTP status code based on the error.
//  3. Creates an ErrorResponse struct with the status code and error message.
//  4. Marshals the ErrorResponse to JSON.
//  5. Writes the HTTP status code and JSON error response to the ResponseWriter.
func (c *Application) errorLogNResponse(w http.ResponseWriter, operation string, opErr error) {
	w.Header().Set("Content-Type", "application/json")
	c.logger.Error(operation, "error", opErr.Error())
	st_code := statusFErr(opErr)
	err_response := ErrorResponse{
		Id:  float32(st_code),
		Msg: opErr.Error(),
	}
	errBytes, _ := json.Marshal(err_response)
	w.WriteHeader(st_code)
	w.Write(errBytes)
}
