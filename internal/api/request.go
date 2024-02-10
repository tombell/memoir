package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/matryer/way"
)

func pageParam(w http.ResponseWriter, r *http.Request) (int32, error) {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	n, err := strconv.Atoi(page)
	if err != nil {
		return -1, errors.New("invalid page parameter")
	}

	if n < 1 {
		n = 1
	}

	return int32(n), nil
}

func idParam(w http.ResponseWriter, r *http.Request) (string, error) {
	id := way.Param(r.Context(), "id")

	if _, err := uuid.Parse(id); err != nil {
		return "", errors.New("invalid id parameter")
	}

	return id, nil
}
