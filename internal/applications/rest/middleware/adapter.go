package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
)

// logger is the shared logger instance used across middleware components.
// It must be initialized using SetLogger before using any middleware that requires logging.
var logger *slog.Logger

// SetLogger sets the logger instance for the middleware package.
// It returns an error if the provided logger is nil.
// This logger will be used by middleware components that require logging capabilities.
func SetLogger(l *slog.Logger) error {
	if l == nil {
		return fmt.Errorf("middleware logger can not be nil")
	}

	logger = l
	return nil
}

// Adapter is a function type that transforms http.Handler into another http.Handler.
// It allows middleware to wrap HTTP handlers with additional functionality.
type Adapter func(http.Handler) http.Handler

// Adapt applies a chain of middleware adapters to a http.Handler.
// The adapters are applied in reverse order, so the first adapter in the list
// will be the outermost wrapper around the handler. This means the request flows
// through the adapters in the order they are provided, while the response flows
// through them in reverse.
//
// Example usage:
//
//	handler := http.HandlerFunc(myHandlerFunc)
//	adaptedHandler := Adapt(handler, LoggingAdapter, AuthAdapter)
//	// Request flow:  LoggingAdapter -> AuthAdapter -> handler
//	// Response flow: handler -> AuthAdapter -> LoggingAdapter
//	http.Handle("/path", adaptedHandler)
func Adapt(handler http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range slices.Backward(adapters) {
		handler = adapter(handler)
	}
	return handler
}
