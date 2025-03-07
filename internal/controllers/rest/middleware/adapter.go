package middleware

import "net/http"

// Adapter is a function type that transforms http.Handler into another http.Handler.
// It allows middleware to wrap HTTP handlers with additional functionality.
type Adapter func(http.Handler) http.Handler

// Adapt applies a chain of middleware adapters to a http.Handler.
// The adapters are applied in the order they are provided, with each adapter
// wrapping the handler returned by the previous adapter.
//
// Example usage:
//
//	handler := http.HandlerFunc(myHandlerFunc)
//	adaptedHandler := Adapt(handler, LoggingAdapter, AuthAdapter)
//	http.Handle("/path", adaptedHandler)
func Adapt(handler http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		handler = adapter(handler)
	}
	return handler
}
