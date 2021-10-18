package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/tombell/memoir/pkg/services"
)

func (s *Server) handleGetTracklists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)
		page := s.pageQueryParam(rid, w, r)

		tracklists, total, err := s.services.GetTracklists(rid, page, perPageTracklists)
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var tl services.TracklistAdd
		if err := json.Unmarshal(body, &tl); err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tracklist, err := s.services.AddTracklist(rid, &tl)
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.writeJSON(rid, w, tracklist)
	}
}

func (s *Server) handleGetTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)
		id := s.idRouteParam(rid, w, r)

		tracklist, err := s.services.GetTracklist(rid, id)
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if tracklist == nil {
			s.services.Logger.Printf("[%s] error=tracklist not found", rid)
			http.Error(w, "tracklist not found", http.StatusNotFound)
			return
		}

		s.writeJSON(rid, w, tracklist)
	}
}

func (s *Server) handlePatchTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)
		id := s.idRouteParam(rid, w, r)

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var tracklistUpdate services.TracklistUpdate
		if err := json.Unmarshal(body, &tracklistUpdate); err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tracklist, err := s.services.UpdateTracklist(rid, id, &tracklistUpdate)
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.writeJSON(rid, w, tracklist)
	}
}
