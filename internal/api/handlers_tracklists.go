package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/tombell/memoir/internal/services"
)

func (s *Server) handleGetTracklists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := s.pageParam(w, r)
		if err != nil {
			s.writeBadRequest(w, err)
			return
		}

		tracklists, total, err := s.services.GetTracklists(page, perPageTracklists)
		if err != nil {
			s.writeInternalServerError(w)
			return
		}

		s.addPaginationHeaders(w, perPageTracklists, page, total)
		s.writeJSON(w, tracklists, http.StatusOK)
	}
}

func (s *Server) handleGetTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := s.idParam(w, r)
		if err != nil {
			s.writeBadRequest(w, err)
			return
		}

		tracklist, err := s.services.GetTracklist(id)
		if err != nil {
			s.writeInternalServerError(w)
			return
		}
		if tracklist == nil {
			s.writeNotFound(w, r)
			return
		}

		s.writeJSON(w, tracklist, http.StatusOK)
	}
}

func (s *Server) handleAddTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			s.services.Logger.Error("handle-add-tracklist:error", "io read all failed", "err", err)
			s.writeInternalServerError(w)
			return
		}

		var tl services.TracklistAdd
		if err = json.Unmarshal(body, &tl); err != nil {
			s.services.Logger.Error("handle-add-tracklist:error", "json unmarshal failed", "err", err)
			s.writeInternalServerError(w)
			return
		}

		tracklist, err := s.services.AddTracklist(&tl)
		if err != nil {
			if errors.Is(err, services.ErrTracklistExists) {
				s.writeUnprocessableContent(w, err)
				return
			}

			s.writeInternalServerError(w)
			return
		}

		s.writeJSON(w, tracklist, http.StatusCreated)
	}
}

func (s *Server) handleUpdateTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := s.idParam(w, r)
		if err != nil {
			s.writeBadRequest(w, err)
			return
		}

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			s.writeInternalServerError(w)
			return
		}

		var tracklistUpdate services.TracklistUpdate
		if err = json.Unmarshal(body, &tracklistUpdate); err != nil {
			s.writeBadRequest(w, err)
			return
		}

		tracklist, err := s.services.UpdateTracklist(id, &tracklistUpdate)
		if err != nil {
			s.writeInternalServerError(w)
			return
		}

		s.writeJSON(w, tracklist, http.StatusOK)
	}
}
