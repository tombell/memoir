package api

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
)

type contextKey string

const requestIDKey contextKey = "requestid"

func (s *Server) json(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		h(w, r)
	}
}

func (s *Server) requestID(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := uuid.NewV4()

		w.Header().Add("X-Request-ID", id.String())

		ctx := context.WithValue(r.Context(), requestIDKey, id.String())
		r.WithContext(ctx)

		h(w, r)
	}
}
