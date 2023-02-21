package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/tombell/memoir/internal/services/models"
)

func (s *Server) handleGetTracklists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := s.pageParam(w, r)

		tracklists, total, err := s.GetTracklists(page, perPageTracklists)
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}

		s.addPaginationHeaders(w, perPageTracklists, page, total)
		s.writeJSON(w, tracklists)
	}
}

func (s *Server) handlePostTracklists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}

		var tl models.TracklistAdd
		if err = json.Unmarshal(body, &tl); err != nil {
			s.writeInternalServerError(w, err)
			return
		}

		tracklist, err := s.AddTracklist(&tl)
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		s.writeJSON(w, tracklist)
	}
}

func (s *Server) handleGetTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := s.idParam(w, r)

		if id == "" {
			return
		}

		tracklist, err := s.GetTracklist(id)
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}
		if tracklist == nil {
			s.writeNotFound(w, r)
			return
		}

		s.writeJSON(w, tracklist)
	}
}

func (s *Server) handlePatchTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := s.idParam(w, r)

		if id == "" {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}

		var tracklistUpdate models.TracklistUpdate
		if err = json.Unmarshal(body, &tracklistUpdate); err != nil {
			s.writeInternalServerError(w, err)
			return
		}

		tracklist, err := s.UpdateTracklist(id, &tracklistUpdate)
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}

		s.writeJSON(w, tracklist)
	}
}
