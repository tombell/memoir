package api

import (
	"net/http"

	"github.com/tombell/memoir/internal/trackliststore"
	"github.com/tombell/memoir/internal/trackstore"
)

func handleGetTrack(trackStore *trackstore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := idParam(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		track, err := trackStore.GetTrack(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if track == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err := encode(w, r, http.StatusOK, track); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func handleGetTracklistsByTrack(
	trackStore *trackstore.Store,
	tracklistStore *trackliststore.Store,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := idParam(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		page, err := pageParam(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		track, err := trackStore.GetTrack(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if track == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		tracklists, total, err := tracklistStore.GetTracklistsByTrack(track.ID, page, tracklistsPerPage)
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

func handleGetMostPlayedTracks(trackStore *trackstore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracks, err := trackStore.GetMostPlayedTracks(maxMostPlayedTracks)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := encode(w, r, http.StatusOK, tracks); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func handleSearchTracks(trackStore *trackstore.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")

		tracks, err := trackStore.SearchTracks(q, maxSearchResults)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := encode(w, r, http.StatusOK, tracks); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
