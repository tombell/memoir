package api

import (
	"encoding/json"
	"net/http"

	"github.com/tombell/memoir/services"
)

func (s *Server) handleGetTracklistsByTrack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		id, err := idRouteParam(r)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		page, err := pageQueryParam(r)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tracklists, err := s.services.GetTracklistsByTrack(id)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		paged := services.NewPagedTracklists(tracklists, page, perPageTracklists)

		resp, err := json.Marshal(paged)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}

func (s *Server) handleGetMostPlayedTracks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		tracks, err := s.services.GetMostPlayedTracks(10)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(tracks)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}

func (s *Server) handleSearchTracks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		tracks, err := s.services.SearchTracks(r.URL.Query().Get("q"))
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(tracks)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}
