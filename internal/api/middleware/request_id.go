package middleware

import (
	"context"
	"log/slog"
	"net/http"
)

const RequestIDContextKey ContextKey = "request-id"

func RequestID(generator func() string, logger *slog.Logger) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := generator()

			w.Header().Add("Request-ID", id)

			ctx := context.WithValue(r.Context(), RequestIDContextKey, id)
			r = r.WithContext(ctx)

			// TODO: figure out best way to use .With and update the logger for
			// the entire request
			// logger = logger.With("request-id", id)

			h(w, r)
		}
	}
}
