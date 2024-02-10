package api

import (
	"net/http"

	"github.com/tombell/memoir/internal/artworkstore"
)

func handlePostArtwork(store *artworkstore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("artwork")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		uploaded, exists, err := store.Upload(file, header.Filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		status := http.StatusCreated
		if exists {
			status = http.StatusOK
		}

		if err := encode(w, r, status, uploaded); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
