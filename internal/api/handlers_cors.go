package api

import (
	"net/http"
)

func (s *Server) handlePreflight() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}