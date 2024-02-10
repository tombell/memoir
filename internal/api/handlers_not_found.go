package api

import (
	"net/http"
)

func handleNotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeNotFound(w, r)
	}
}
