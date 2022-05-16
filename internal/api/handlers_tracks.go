package api

import (
	"fmt"
	"net/http"

	"github.com/tombell/mw"
)

func (s *Server) handleGetTrack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := mw.FindRequestID(r)
		id := s.idParam(rid, w, r)

		if id == "" {
			return
		}

		track, err := s.Services.GetTrack(rid, id)
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}
		if track == nil {
			s.writeNotFound(rid, w, fmt.Sprintf("track with id: %s", id))
			return
		}

		s.writeJSON(rid, w, track)
	}
}

func (s *Server) handleGetTracklistsByTrack() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := mw.FindRequestID(r)
		id := s.idParam(rid, w, r)
		page := s.pageParam(rid, w, r)

		if id == "" {
			return
		}

		tracklists, total, err := s.Services.GetTracklistsByTrack(rid, id, page, perPageTracklists)
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		s.addPaginationHeaders(w, perPageTracklists, page, total)
		s.writeJSON(rid, w, tracklists)
	}
}

func (s *Server) handleGetMostPlayedTracks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := mw.FindRequestID(r)

		tracks, err := s.Services.GetMostPlayedTracks(rid, mostPlayedTracksLimit)
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		s.writeJSON(rid, w, tracks)
	}
}

func (s *Server) handleSearchTracks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := mw.FindRequestID(r)
		q := searchParam(r)

		tracks, err := s.Services.SearchTracks(rid, q, searchResultsLimit)
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		s.writeJSON(rid, w, tracks)
	}
}
