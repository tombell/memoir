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

		resp, err := json.Marshal(tracklists)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}
