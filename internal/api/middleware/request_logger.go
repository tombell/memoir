package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// RequestLogger is a middleware function that uses the slog.Logger from the
// request context to log before and after a request.
func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger, _ := r.Context().Value(LoggerContextKey).(*slog.Logger)

			start := time.Now().UTC()

			logger.Info(
				"http:started",
				"method",
				r.Method,
				"path",
				r.URL.Path,
			)

			rw := &responseWriter{ResponseWriter: w}

			next.ServeHTTP(rw, r)

			logger.Info(
				"http:finished",
				"method",
				r.Method,
				"path",
				r.URL.Path,
				"status",
				rw.status,
				"size",
				rw.size,
				"elapsed",
				time.Since(start),
			)
		})
	}
}
