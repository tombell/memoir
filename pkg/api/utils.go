package api

import (
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/matryer/way"
)

func pageQueryParam(r *http.Request) (int, error) {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	return strconv.Atoi(page)
}

func searchQueryParam(r *http.Request) string {
	return r.URL.Query().Get("q")
}

func idRouteParam(r *http.Request) (string, error) {
	id := way.Param(r.Context(), "id")

	if _, err := uuid.FromString(id); err != nil {
		return "", err
	}

	return id, nil
}
