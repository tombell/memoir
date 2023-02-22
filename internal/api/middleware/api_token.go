package middleware

import (
	"net/http"

	"github.com/charmbracelet/log"
)

func APIToken(logger log.Logger, token string) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("API-Token")

			if key == "" {
				logger.Error("api-token", "reason", "missing")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if key != token {
				logger.Error("api-token", "reason", "invalid")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			h(w, r)
		}
	}
}
