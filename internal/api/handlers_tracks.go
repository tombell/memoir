package api

import (
	"net/http"
)

func (s *Server) handleGetTrack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := s.idParam(w, r)

		if id == "" {
			return
		}

		track, err := s.GetTrack(id)
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}
		if track == nil {
			s.writeNotFound(w, r)
			return
		}

		s.writeJSON(w, track)
	}
}

func (s *Server) handleGetTracklistsByTrack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := s.idParam(w, r)
		page := s.pageParam(w, r)

		if id == "" {
			return
		}

		tracklists, total, err := s.GetTracklistsByTrack(id, page, perPageTracklists)
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}

		s.addPaginationHeaders(w, perPageTracklists, page, total)
		s.writeJSON(w, tracklists)
	}
}

func (s *Server) handleGetMostPlayedTracks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracks, err := s.GetMostPlayedTracks(mostPlayedTracksLimit)
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}

		s.writeJSON(w, tracks)
	}
}

func (s *Server) handleSearchTracks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := searchParam(r)

		tracks, err := s.SearchTracks(q, searchResultsLimit)
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}

		s.writeJSON(w, tracks)
	}
}
