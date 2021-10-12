package api

import (
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/matryer/way"
)

func (s *Server) pageQueryParam(rid string, w http.ResponseWriter, r *http.Request) int {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	n, err := strconv.Atoi(page)
	if err != nil {
		s.services.Logger.Printf("[%s] error=%s\n", rid, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return -1
	}

	return n
}

func (s *Server) idRouteParam(rid string, w http.ResponseWriter, r *http.Request) string {
	id := way.Param(r.Context(), "id")

	if _, err := uuid.FromString(id); err != nil {
		s.services.Logger.Printf("[%s] error=%s\n", rid, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return ""
	}

	return id
}

func searchQueryParam(r *http.Request) string {
	return r.URL.Query().Get("q")
}
