package middleware

import (
	"log/slog"
	"net/http"
)

func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger, _ := r.Context().Value(LoggerContextKey).(*slog.Logger)

			defer func() {
				if err := recover(); err != nil {
					logger.Error("recovered from panic", "err", err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
