package rest

import (
	"encoding/json"
	"net/http"
)

// successResponse writes a JSON-encoded success response to the http.ResponseWriter.
// It sets the Content-Type header to application/json, marshals the provided data,
// and writes the response with the specified HTTP status code.
func (c *Controller) successResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	dataJson, err := json.Marshal(data)
	if err != nil {
		c.logger.Error("rendering response", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(status)
	w.Write(dataJson)
}
