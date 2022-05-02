package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tombell/memoir/pkg/services"
)

func (s *Server) handleGetTracklists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)
		page := s.pageParam(rid, w, r)

		tracklists, total, err := s.services.GetTracklists(rid, page, perPageTracklists)
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		s.addPaginationHeaders(w, perPageTracklists, page, total)
		s.writeJSON(rid, w, tracklists)
	}
}

func (s *Server) handlePostTracklists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		var tl services.TracklistAdd
		if err = json.Unmarshal(body, &tl); err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		tracklist, err := s.services.AddTracklist(rid, &tl)
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		s.writeJSON(rid, w, tracklist)
	}
}

func (s *Server) handleGetTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)
		id := s.idParam(rid, w, r)

		if id == "" {
			return
		}

		tracklist, err := s.services.GetTracklist(rid, id)
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}
		if tracklist == nil {
			s.writeNotFound(rid, w, fmt.Sprintf("find tracklist with id: %s", id))
			return
		}

		s.writeJSON(rid, w, tracklist)
	}
}

func (s *Server) handlePatchTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)
		id := s.idParam(rid, w, r)

		if id == "" {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		var tracklistUpdate services.TracklistUpdate
		if err = json.Unmarshal(body, &tracklistUpdate); err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		tracklist, err := s.services.UpdateTracklist(rid, id, &tracklistUpdate)
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		s.writeJSON(rid, w, tracklist)
	}
}
