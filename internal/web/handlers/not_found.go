package handlers

import (
	"net/http"
)

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}
}
