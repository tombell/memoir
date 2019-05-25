package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/matryer/way"

	"github.com/tombell/memoir/services"
)

const perPage = 10

func (s *Server) handleGetTracklists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		// TODO: clean up query params parsing.

		params := r.URL.Query()
		page := params.Get("page")

		if page == "" {
			page = "1"
		}

		npage, err := strconv.Atoi(page)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tracklists, err := s.services.GetTracklists()
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		paged := services.NewPagedTracklists(tracklists, npage, perPage)

		resp, err := json.Marshal(paged)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}

func (s *Server) handleGetTracklist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		id := way.Param(r.Context(), "id")

		tracklist, err := s.services.GetTracklist(id)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if tracklist == nil {
			s.logger.Printf("rid=%s error=tracklist not found", rid)
			http.Error(w, "tracklist not found", http.StatusNotFound)
			return
		}

		resp, err := json.Marshal(tracklist)
		if err != nil {
			s.logger.Printf("rid=%s error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}
