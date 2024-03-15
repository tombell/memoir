package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tombell/memoir/internal/stores/artworkstore"
)

func PostArtwork(store *artworkstore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		file, header, err := r.FormFile("artwork")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		uploaded, exists, err := store.Upload(ctx, file, header.Filename)
		if err != nil {
			fmt.Println(err)
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
