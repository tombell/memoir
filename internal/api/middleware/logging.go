package middleware

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
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

func Logging(logger log.Logger) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UTC()

			logger.Info(
				"http",
				"method",
				r.Method,
				"path",
				r.URL.Path,
			)

			rw := &responseWriter{ResponseWriter: w}

			h(rw, r)

			logger.Info(
				"http",
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
		}
	}
}