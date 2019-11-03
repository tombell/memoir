package api

import (
	"encoding/json"
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
