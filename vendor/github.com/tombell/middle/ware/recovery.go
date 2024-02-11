package ware

import (
	"log/slog"
	"net/http"
)

// Recovery is a middleware function that recovers from panics and logs out the
// error using the slog.Logger from the request context.
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if logger, ok := r.Context().Value(LoggerContextKey).(*slog.Logger); ok {
						logger.Error("recovered from panic", "err", err)
					}

					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
