package middleware

import (
	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/auth"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/middle/ware"
	"net/http"
)

func Session(service *auth.Service, cfg config.AuthConfig, verified bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.SessionCookieName())
			if err != nil {
				payload.WriteError(ware.LoggerFromContext(r.Context()), w, errors.E("session", http.StatusUnauthorized))
				return
			}
			user, err := service.Lookup(r.Context(), cookie.Value)
			if err != nil {
				payload.WriteError(ware.LoggerFromContext(r.Context()), w, err)
				return
			}
			if verified && !user.EmailVerified {
				payload.WriteError(ware.LoggerFromContext(r.Context()), w, errors.E("session", http.StatusForbidden, errors.M{"message": {"email verification required"}}))
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			next.ServeHTTP(w, r.WithContext(auth.WithUser(r.Context(), user)))
		})
	}
}
