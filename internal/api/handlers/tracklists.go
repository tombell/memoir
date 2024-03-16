package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/tombell/memoir/internal/stores/trackliststore"
)

func GetTracklist(tracklistStore *trackliststore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		tracklist, err := tracklistStore.GetTracklist(ctx, r.PathValue("id"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if tracklist == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err := encode(w, r, http.StatusOK, tracklist); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func PostTracklist(tracklistStore *trackliststore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		params, err := decode[trackliststore.AddTracklistParams](r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tracklist, err := tracklistStore.AddTracklist(ctx, &params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if tracklist == nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if err := encode(w, r, http.StatusOK, tracklist); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func PatchTracklist(tracklistStore *trackliststore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		params, err := decode[trackliststore.UpdateTracklistParams](r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tracklist, err := tracklistStore.UpdateTracklist(ctx, r.PathValue("id"), &params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := encode(w, r, http.StatusOK, tracklist); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
