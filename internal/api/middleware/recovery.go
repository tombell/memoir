package middleware

import (
	"log/slog"
	"net/http"

	"github.com/tombell/middle/ware"

	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/errors"
)

// Recovery logs panics and writes the API error envelope if the response has
// not already been committed.
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := &responseWriter{ResponseWriter: w}
			defer func() {
				if value := recover(); value != nil {
					if value == http.ErrAbortHandler {
						panic(value)
					}

					logger := ware.LoggerFromContext(r.Context())
					if logger == nil {
						logger = slog.Default()
					}
					logger.Error("recovered from panic", "err", value)
					if !response.ResponseCommitted() {
						payload.WriteError(logger, response, errors.E("middleware[recovery]"))
					}
				}
			}()

			next.ServeHTTP(response, r)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	committed bool
}

func (w *responseWriter) ResponseCommitted() bool {
	return w.committed
}

func (w *responseWriter) WriteHeader(status int) {
	if w.committed {
		return
	}
	if status >= 200 || status == http.StatusSwitchingProtocols {
		w.committed = true
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(body []byte) (int, error) {
	if !w.committed {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriter) FlushError() error {
	if !w.committed {
		w.WriteHeader(http.StatusOK)
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}
