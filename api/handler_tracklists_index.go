package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleTracklistsIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracklists, err := s.services.RecentTracklists()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(tracklists); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
