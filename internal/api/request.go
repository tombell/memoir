package api

import (
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/matryer/way"
)

func (s *Server) pageParam(w http.ResponseWriter, r *http.Request) (int, error) {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	n, err := strconv.Atoi(page)
	if err != nil {
		return -1, err
	}

	if n < 1 {
		n = 1
	}

	return n, nil
}

func (s *Server) idParam(w http.ResponseWriter, r *http.Request) (string, error) {
	id := way.Param(r.Context(), "id")

	if _, err := uuid.FromString(id); err != nil {
		return "", err
	}

	return id, nil
}

func searchParam(r *http.Request) string {
	return r.URL.Query().Get("q")
}
