package rest

import (
	"context"
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
	rw.ResponseWriter.WriteHeader(s)
	rw.status = s
}

// withLogging is a middleware function that wraps an http.Handler with logging functionality.
// It logs information about each HTTP request, including method, URI, status code, response size, and duration.
// The log level is set to Error for non-successful responses (status < 200 or > 399), and Info otherwise.
func withLogging(ctx context.Context, h http.Handler, logger *slog.Logger) http.Handler {
	loggingFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := loggingResponseWriter{ResponseWriter: w, status: http.StatusOK, size: 0}
		h.ServeHTTP(&lrw, r)
		duration := time.Since(start)
		logLevel := slog.LevelInfo
		if lrw.status < http.StatusOK || lrw.status > 399 {
			logLevel = slog.LevelError
		}

		logger.Log(ctx, logLevel, r.RequestURI,
			"method", fmt.Sprintf("%s", r.Method),
			"status", lrw.status,
			"size", lrw.size,
			"duration", duration,
		)
	}

	return http.HandlerFunc(loggingFn)
}
