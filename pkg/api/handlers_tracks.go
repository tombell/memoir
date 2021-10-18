package api

import (
	"net/http"

	"github.com/tombell/memoir/pkg/services"
)

func (s *Server) handleGetTrack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)
		id := s.idRouteParam(rid, w, r)

		track, err := s.services.GetTrack(rid, id)
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.writeJSON(rid, w, track)
	}
}

func (s *Server) handleGetTracklistsByTrack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)
		id := s.idRouteParam(rid, w, r)
		page := s.pageQueryParam(rid, w, r)

		tracklists, total, err := s.services.GetTracklistsByTrack(rid, id, page, perPageTracklists)
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.addPaginationHeaders(w, perPageTracklists, page, total)
		s.writeJSON(rid, w, tracklists)
	}
}

func (s *Server) handleGetMostPlayedTracks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		tracks, err := s.services.GetMostPlayedTracks(rid, mostPlayedTracksLimit)
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.writeJSON(rid, w, tracks)
	}
}

func (s *Server) handleSearchTracks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)
		q := searchQueryParam(r)
		page := s.pageQueryParam(rid, w, r)

		tracks, err := s.services.SearchTracks(rid, q)
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		paged := services.NewPagedTracks(tracks, page, perPageTracks)

		s.writeJSON(rid, w, paged)
	}
}
