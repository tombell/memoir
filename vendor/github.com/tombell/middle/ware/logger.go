package ware

import (
	"context"
	"log/slog"
	"net/http"
)

// LoggerContextKey is the key to retrieve the slog.Logger instance from the
// request context.
const LoggerContextKey ContextKey = "logger"

// Logger is a middleware function that adds a slog.Logger instance to the
// request context. This logger can then be used to scope logging to just this
// request. For example the RequestID middleware function.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), LoggerContextKey, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
