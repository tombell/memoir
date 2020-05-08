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

		page, err := pageQueryParam(r)
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tracklists, err := s.services.GetTracklists()
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		paged := services.NewPagedTracklists(tracklists, page, perPageTracklists)

		resp, err := json.Marshal(paged)
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}

func (s *Server) handleGetTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		id, err := idRouteParam(r)
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		tracklist, err := s.services.GetTracklist(id)
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if tracklist == nil {
			s.services.Logger.Printf("rid=%s error=tracklist not found", rid)
			http.Error(w, "tracklist not found", http.StatusNotFound)
			return
		}

		resp, err := json.Marshal(tracklist)
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}

func (s *Server) handleImportTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var tracklistImport services.TracklistImport
		if err := json.Unmarshal(body, &tracklistImport); err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tracklist, err := s.services.ImportTracklist(&tracklistImport)
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(tracklist)
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}

func (s *Server) handleUpdateTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		id, err := idRouteParam(r)
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var tracklistUpdate services.TracklistUpdate
		if err := json.Unmarshal(body, &tracklistUpdate); err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tracklist, err := s.services.UpdateTracklist(id, &tracklistUpdate)
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(tracklist)
		if err != nil {
			s.services.Logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}
