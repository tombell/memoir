package api

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
)

const (
	requestIDContextKey contextKey = "requestid"
)

func (s *Server) requestID(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := uuid.NewV4()

		w.Header().Add("X-Request-ID", id.String())

		ctx := context.WithValue(r.Context(), requestIDContextKey, id.String())
		r = r.WithContext(ctx)

		h(w, r)
	}
}

func getRequestID(r *http.Request) string {
	return r.Context().Value(requestIDContextKey).(string)
}
