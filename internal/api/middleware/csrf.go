package middleware

import (
	"crypto/subtle"
	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/auth"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/middle/ware"
	"net/http"
)

// CSRF uses a host-only double-submit cookie and an exact allowed origin.
// Production cookies use the __Host- prefix to prevent subdomain injection.
func CSRF(cfg config.AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := r.Cookie(cfg.CSRFCookieName())
			header := r.Header.Get("X-CSRF-Token")
			if r.Header.Get("Origin") != cfg.Origin || err != nil || !auth.ValidToken(header) || subtle.ConstantTimeCompare([]byte(header), []byte(cookie.Value)) != 1 {
				payload.WriteError(ware.LoggerFromContext(r.Context()), w, errors.E("csrf", http.StatusForbidden))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
