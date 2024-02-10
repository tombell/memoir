package middleware

import (
	"context"
	"log/slog"
	"net/http"
)

const LoggerContextKey ContextKey = "logger"

func Logger(logger *slog.Logger) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), LoggerContextKey, logger)
			r = r.WithContext(ctx)
			h(w, r)
		}
	}
}
