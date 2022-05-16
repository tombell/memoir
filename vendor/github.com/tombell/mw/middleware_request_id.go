package mw

import (
	"context"
	"net/http"
)

type RequestIDGenerator func() string

const requestIDContextKey ContextKey = "request-id"

func RequestID(generator RequestIDGenerator) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := generator()

			w.Header().Add("Request-ID", id)

			ctx := context.WithValue(r.Context(), requestIDContextKey, id)
			r = r.WithContext(ctx)

			h(w, r)
		}
	}
}

func FindRequestID(r *http.Request) string {
	return r.Context().Value(requestIDContextKey).(string)
}
