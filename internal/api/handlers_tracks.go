package api

import (
	"net/http"

	"github.com/tombell/memoir/internal/trackliststore"
	"github.com/tombell/memoir/internal/trackstore"
)

func handleGetTrack(trackStore *trackstore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			writeNotFound(w, r)
			return
		}

		writeJSON(w, track, http.StatusOK)
	}
}

func handleGetTracklistsByTrack(
	trackStore *trackstore.Store,
	tracklistStore *trackliststore.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			writeNotFound(w, r)
			return
		}

		tracklists, total, err := tracklistStore.GetTracklistsByTrack(track.ID, page, perPageTracklists)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		addPaginationHeaders(w, perPageTracklists, page, total)
		writeJSON(w, tracklists, http.StatusOK)
	}
}

func handleGetMostPlayedTracks(trackStore *trackstore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracks, err := trackStore.GetMostPlayedTracks(mostPlayedTracksLimit)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		writeJSON(w, tracks, http.StatusOK)
	}
}

func handleSearchTracks(trackStore *trackstore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")

		tracks, err := trackStore.SearchTracks(q, searchResultsLimit)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		writeJSON(w, tracks, http.StatusOK)
	}
}
