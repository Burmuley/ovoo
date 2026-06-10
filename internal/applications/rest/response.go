package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// successResponse writes a JSON-encoded success response to the http.ResponseWriter.
// It sets the Content-Type header to application/json, marshals the provided data,
// and writes the response with the specified HTTP status code.
func (a *Application) successResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	dataJson, err := json.Marshal(data)
	if err != nil {
		a.logger.Error("rendering response", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(dataJson); err != nil {
		a.logger.Error("writing response", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
		Errors: []Error{
			{
				Status: strconv.Itoa(st_code),
				Detail: opErr.Error(),
			},
		},
	}
	errBytes, _ := json.Marshal(err_response)
	w.WriteHeader(st_code)
	if _, err := w.Write(errBytes); err != nil {
		c.logger.Error("writing error response", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
