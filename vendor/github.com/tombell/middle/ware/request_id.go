package ware

import (
	"context"
	"log/slog"
	"net/http"
)

// RequestIDContextKey is the key to retrieve the request ID from the request
// context.
const RequestIDContextKey ContextKey = "request-id"

// RequestID is a middleware function that uses the given function to generate a
// unique ID for the request. This ID is then added to the request context, and
// as a key/value attribute to the slog.Logger from the request context. A
// Request-ID header is added to the response headers.
func RequestID(generate func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := generate()
			ctx := context.WithValue(r.Context(), RequestIDContextKey, rid)

			if logger, ok := r.Context().Value(LoggerContextKey).(*slog.Logger); ok {
				logger = logger.With("rid", rid)
				ctx = context.WithValue(ctx, LoggerContextKey, logger)
			}

			w.Header().Add("Request-ID", rid)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
