package middleware

import (
	"context"
	"net/http"

	"github.com/charmbracelet/log"
)

type RequestIDGenerator func() string

const requestIDContextKey ContextKey = "request-id"

func RequestID(generator RequestIDGenerator, logger log.Logger) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := generator()

			w.Header().Add("Request-ID", id)

			ctx := context.WithValue(r.Context(), requestIDContextKey, id)
			r = r.WithContext(ctx)

			logger.SetPrefix(id)

			h(w, r)

			logger.SetPrefix("")
		}
	}
}

func FindRequestID(r *http.Request) string {
	return r.Context().Value(requestIDContextKey).(string)
}
