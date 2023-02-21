package api

import (
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/matryer/way"
)

func (s *Server) pageParam(w http.ResponseWriter, r *http.Request) int {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	n, err := strconv.Atoi(page)
	if err != nil {
		s.Logger.Error("converting page from string to int failed", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return -1
	}

	if n < 1 {
		n = 1
	}

	return n
}

func (s *Server) idParam(w http.ResponseWriter, r *http.Request) string {
	id := way.Param(r.Context(), "id")

	if _, err := uuid.FromString(id); err != nil {
		s.writeNotFound(w, r)
		return ""
	}

	return id
}

func searchParam(r *http.Request) string {
	return r.URL.Query().Get("q")
}
