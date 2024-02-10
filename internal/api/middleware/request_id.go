package middleware

import (
	"context"
	"log/slog"
	"net/http"
)

const RequestIDContextKey ContextKey = "request-id"

func RequestID(generate func() string) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rid := generate()
			ctx := context.WithValue(r.Context(), RequestIDContextKey, rid)

			logger, _ := r.Context().Value(LoggerContextKey).(*slog.Logger)
			logger = logger.With("rid", rid)
			ctx = context.WithValue(ctx, LoggerContextKey, logger)

			w.Header().Add("Request-ID", rid)
			r = r.WithContext(ctx)

			h(w, r)
		}
	}
}
