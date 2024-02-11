package api

import (
	"net/http"

	"github.com/tombell/memoir/internal/trackliststore"
)

func handleGetTracklists(tracklistStore *trackliststore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := pageParam(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tracklists, total, err := tracklistStore.GetTracklists(page, tracklistsPerPage)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		addPaginationHeaders(w, tracklistsPerPage, page, total)

		if err := encode(w, r, http.StatusOK, tracklists); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func handleGetTracklist(tracklistStore *trackliststore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err := encode(w, r, http.StatusOK, tracklist); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func handleAddTracklist(tracklistStore *trackliststore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params, err := decode[trackliststore.AddTracklistParams](r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tracklist, err := tracklistStore.AddTracklist(&params)
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

func handleUpdateTracklist(tracklistStore *trackliststore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := idParam(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		params, err := decode[trackliststore.UpdateTracklistParams](r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tracklist, err := tracklistStore.UpdateTracklist(id, &params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := encode(w, r, http.StatusOK, tracklist); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
