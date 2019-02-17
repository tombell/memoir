package api

import (
	"context"
	"net/http"
	"time"

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
		r = r.WithContext(ctx)

		h(w, r)
	}
}

func (s *Server) instruments(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		start := time.Now().UTC()
		method := r.Method
		path := r.URL.Path
		addr := r.RemoteAddr

		s.logger.Printf("at=start rid=%s method=%s path=%s ip=%s\n", rid, method, path, addr)

		h(w, r)

		dur := time.Since(start)

		s.logger.Printf("at=end rid=%s method=%s path=%s ip=%s time=%s\n", rid, method, path, addr, dur)
	}
}
