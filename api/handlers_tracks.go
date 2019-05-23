package api

import (
	"encoding/json"
	"net/http"

	"github.com/matryer/way"
)

func (s *Server) handleGetTracklistsByTrack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		id := way.Param(r.Context(), "id")

		tracklists, err := s.services.GetTracklistsByTrack(id)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(tracklists)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}
