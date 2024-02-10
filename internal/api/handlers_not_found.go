package api

import (
	"net/http"
)

func (s *Server) handleNotFOund() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.writeNotFound(w, r)
	}
}
