package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

func GetTracklistsByTrack(
	trackStore *trackstore.Store,
	tracklistStore *trackliststore.Store,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		page := pageParam(r)

		track, err := trackStore.GetTrack(ctx, r.PathValue("id"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if track == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		tracklists, total, err := tracklistStore.GetTracklistsByTrack(ctx, track.ID, page, tracklistsPerPage)
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
