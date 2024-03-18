package middleware

import (
	"log/slog"
	"net/http"

	"github.com/tombell/middle/ware"
)

// Authorize is a middleware function that checks an API-Token from the request
// headers against the given API token.
func Authorize(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger, _ := r.Context().Value(ware.LoggerContextKey).(*slog.Logger)

			key := r.Header.Get("API-Token")

			if key == "" {
				logger.Info("authorization failed", "reason", "missing")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if key != token {
				logger.Info("authorization failed", "reason", "invalid")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
