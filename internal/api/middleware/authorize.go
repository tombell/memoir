package middleware

import (
	"log/slog"
	"net/http"

	"github.com/tombell/middle/ware"

	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/errors"
)

// Authorize is a middleware function that checks an API-Token from the request
// headers against the given API token.
func Authorize(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := ware.LoggerFromContext(r.Context())
			if logger == nil {
				logger = slog.Default()
			}

			key := r.Header.Get("API-Token")

			if key == "" {
				logger.Info("token authorization failed", "reason", "missing")
				payload.WriteError(logger, w, errors.E("middleware[authorize]", http.StatusUnauthorized))
				return
			}

			if key != token {
				logger.Info("token authorization failed", "reason", "invalid")
				payload.WriteError(logger, w, errors.E("middleware[authorize]", http.StatusForbidden))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
