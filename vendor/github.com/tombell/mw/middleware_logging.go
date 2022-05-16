package mw

import (
	"log"
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
	if rw.status == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func Logging(logger *log.Logger) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rid := FindRequestID(r)
			start := time.Now().UTC()

			logger.Printf(
				"[%s] method=%s path=%s",
				rid,
				r.Method,
				r.URL.Path,
			)

			rw := &responseWriter{ResponseWriter: w}

			h(rw, r)

			logger.Printf(
				"[%s] method=%s path=%s status=%d size=%d elapsed=%s",
				rid,
				r.Method,
				r.URL.Path,
				rw.status,
				rw.size,
				time.Since(start),
			)
		}
	}
}
