package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/tombell/memoir/internal/services"
)

func handleGetTracklists(services *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := pageParam(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tracklists, total, err := services.GetTracklists(page, perPageTracklists)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		addPaginationHeaders(w, perPageTracklists, page, total)
		writeJSON(w, tracklists, http.StatusOK)
	}
}

func handleGetTracklist(services *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := idParam(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tracklist, err := services.GetTracklist(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if tracklist == nil {
			writeNotFound(w, r)
			return
		}

		writeJSON(w, tracklist, http.StatusOK)
	}
}

func handleAddTracklist(svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var tl services.TracklistAdd
		if err = json.Unmarshal(body, &tl); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tracklist, err := svc.AddTracklist(&tl)
		if err != nil {
			if errors.Is(err, services.ErrTracklistExists) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		writeJSON(w, tracklist, http.StatusCreated)
	}
}

func handleUpdateTracklist(svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := idParam(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var tracklistUpdate services.TracklistUpdate
		if err = json.Unmarshal(body, &tracklistUpdate); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tracklist, err := svc.UpdateTracklist(id, &tracklistUpdate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		writeJSON(w, tracklist, http.StatusOK)
	}
}
