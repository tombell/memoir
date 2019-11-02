package api

import (
	"net/http"
)

func (s *Server) json(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		h(w, r)
	}
}
