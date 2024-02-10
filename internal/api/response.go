package api

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
)

type Error struct {
	Err string `json:"err"`
	Msg string `json:"msg"`
}

func addPaginationHeaders(w http.ResponseWriter, limit, page int32, total int64) {
	pages := int(math.Ceil(float64(total) / float64(limit)))

	w.Header().Add("Current-Page", strconv.FormatInt(int64(page), 10))
	w.Header().Add("Total-Pages", strconv.Itoa(pages))
}

func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("json encode failed: %w", err)
	}

	return nil
}
