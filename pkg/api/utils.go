package api

import (
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/matryer/way"
)

func pageQueryParam(r *http.Request) (int, error) {
	params := r.URL.Query()

	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	return strconv.Atoi(page)
}

func idRouteParam(r *http.Request) (string, error) {
	id := way.Param(r.Context(), "id")

	if _, err := uuid.FromString(id); err != nil {
		return "", err
	}

	return id, nil
}
