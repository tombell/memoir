package api

import (
	"net/http"

	"github.com/tombell/memoir/internal/services"
)

func handleGetTrack(services *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := idParam(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		track, err := services.GetTrack(id)
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

func handleGetTracklistsByTrack(services *services.Services) http.HandlerFunc {
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

		track, err := services.GetTrack(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if track == nil {
			writeNotFound(w, r)
			return
		}

		tracklists, total, err := services.GetTracklistsByTrack(track.ID, page, perPageTracklists)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		addPaginationHeaders(w, perPageTracklists, page, total)
		writeJSON(w, tracklists, http.StatusOK)
	}
}

func handleGetMostPlayedTracks(services *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracks, err := services.GetMostPlayedTracks(mostPlayedTracksLimit)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		writeJSON(w, tracks, http.StatusOK)
	}
}

func handleSearchTracks(services *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")

		tracks, err := services.SearchTracks(q, searchResultsLimit)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		writeJSON(w, tracks, http.StatusOK)
	}
}
