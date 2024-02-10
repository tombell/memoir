package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("json decode failed: %w", err)
	}

	return v, nil
}

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
	id := r.PathValue("id")

	if _, err := uuid.Parse(id); err != nil {
		return "", errors.New("invalid id parameter")
	}

	return id, nil
}
