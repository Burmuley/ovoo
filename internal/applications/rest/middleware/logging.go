package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// loggingResponseWriter is a custom http.ResponseWriter that tracks the status code and response size.
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// Write writes the data to the connection as part of an HTTP reply.
// It updates the size of the response.
func (rw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// WriteHeader sends an HTTP response header with the provided status code.
// It updates the status of the response.
func (rw *loggingResponseWriter) WriteHeader(s int) {
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}

// Logging is a middleware adapter that logs information about incoming HTTP requests.
// It captures details such as request method, URI, response status, size, and duration,
// and logs them using the provided slog.Logger.
//
// The log level is determined by the HTTP status code:
// - For 2xx-3xx status codes, information is logged at the INFO level.
// - For status codes < 200 or >= 400, information is logged at the ERROR level.
//
// Parameters:
//   - logger: The structured logger used to record the request information.
//
// Returns:
//   - An Adapter function that wraps an http.Handler with logging functionality.
func Logging(logger *slog.Logger) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := loggingResponseWriter{ResponseWriter: w, status: http.StatusOK, size: 0}
			h.ServeHTTP(&lrw, r)
			duration := time.Since(start)
			logLevel := slog.LevelInfo
			if lrw.status < http.StatusOK || lrw.status > 399 {
				logLevel = slog.LevelError
			}

			logger.Log(r.Context(), logLevel, r.RequestURI,
				"method", fmt.Sprintf("%s", r.Method),
				"status", lrw.status,
				"size", lrw.size,
				"duration", duration,
			)
		})
	}
}
