package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
)

type middleware func(http.HandlerFunc) http.HandlerFunc

type contextKey string

const requestIDKey contextKey = "requestid"

func use(h http.HandlerFunc, middlewares ...middleware) http.HandlerFunc {
	for _, m := range middlewares {
		h = m(h)
	}

	return h
}

func (s *Server) json(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		h(w, r)
	}
}

func (s *Server) cors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
		w.Header().Add("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")

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

func getRequestID(r *http.Request) string {
	return r.Context().Value(requestIDKey).(string)
}

func (s *Server) instruments(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		start := time.Now().UTC()
		method := r.Method
		path := r.URL.Path
		addr := r.RemoteAddr

		s.logger.Printf("rid=%s method=%s path=%s ip=%s\n", rid, method, path, addr)
		h(w, r)
		s.logger.Printf("rid=%s method=%s path=%s ip=%s elapsed=%s\n", rid, method, path, addr, time.Since(start))
	}
}
