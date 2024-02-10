package api

import (
	"net/http"

	"github.com/tombell/memoir/internal/services"
)

func handlePostArtwork(services *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("artwork")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		uploaded, exists, err := services.UploadArtwork(file, header.Filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		status := http.StatusCreated
		if exists {
			status = http.StatusOK
		}

		writeJSON(w, uploaded, status)
	}
}
