package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/tombell/memoir/internal/trackliststore"
)

func handleGetTracklists(tracklistStore *trackliststore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := pageParam(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tracklists, total, err := tracklistStore.GetTracklists(page, perPageTracklists)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		addPaginationHeaders(w, perPageTracklists, page, total)
		writeJSON(w, tracklists, http.StatusOK)
	}
}

func handleGetTracklist(tracklistStore *trackliststore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := idParam(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tracklist, err := tracklistStore.GetTracklist(id)
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

func handleAddTracklist(tracklistStore *trackliststore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var tl trackliststore.AddTracklistParams
		if err = json.Unmarshal(body, &tl); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tracklist, err := tracklistStore.AddTracklist(&tl)
		if err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if tracklist == nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		writeJSON(w, tracklist, http.StatusCreated)
	}
}

func handleUpdateTracklist(tracklistStore *trackliststore.Store) http.HandlerFunc {
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

		var tracklistUpdate trackliststore.UpdateTracklistParams
		if err = json.Unmarshal(body, &tracklistUpdate); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tracklist, err := tracklistStore.UpdateTracklist(id, &tracklistUpdate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		writeJSON(w, tracklist, http.StatusOK)
	}
}
