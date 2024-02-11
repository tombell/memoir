package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
)

const (
	tracklistsPerPage = 20

	maxMostPlayedTracks = 10
	maxSearchResults    = 10
)

func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("json encode failed: %w", err)
	}

	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("json decode failed: %w", err)
	}

	return v, nil
}

func addPaginationHeaders(w http.ResponseWriter, limit, page int32, total int64) {
	pages := int(math.Ceil(float64(total) / float64(limit)))

	w.Header().Add("Current-Page", strconv.FormatInt(int64(page), 10))
	w.Header().Add("Total-Pages", strconv.Itoa(pages))
}

func pageParam(r *http.Request) int32 {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	n, err := strconv.Atoi(page)
	if err != nil || n < 1 {
		n = 1
	}

	return int32(n)
}
