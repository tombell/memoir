package handlers

import (
	"net/http"

	"github.com/tombell/memoir/internal/trackliststore"
)

func GetTracklists(tracklistStore *trackliststore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := pageParam(r)

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

func GetTracklist(tracklistStore *trackliststore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracklist, err := tracklistStore.GetTracklist(r.PathValue("id"))
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

func PatchTracklist(tracklistStore *trackliststore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params, err := decode[trackliststore.UpdateTracklistParams](r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tracklist, err := tracklistStore.UpdateTracklist(r.PathValue("id"), &params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := encode(w, r, http.StatusOK, tracklist); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
