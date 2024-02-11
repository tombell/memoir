package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/tombell/memoir/internal/api/middleware"
	"github.com/tombell/memoir/internal/artworkstore"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/trackliststore"
	"github.com/tombell/memoir/internal/trackstore"
)

func routes(
	logger *slog.Logger,
	router *http.ServeMux,
	config *config.Config,
	tracklistStore *trackliststore.Store,
	trackStore *trackstore.Store,
	artworkStore *artworkstore.Store,
) {
	api := middleware.Use(
		middleware.CORS(),
		middleware.RequestLogger(),
		middleware.RequestID(uuid.NewString),
		middleware.Logger(logger),
	)

	authorized := middleware.Use(
		middleware.Authorize(config.API.Token),
		api,
	)

	router.Handle("GET /tracklists", api(handleGetTracklists(tracklistStore)))
	router.Handle("GET /tracklists/{id}", api(handleGetTracklist(tracklistStore)))
	router.Handle("POST /tracklists", authorized(handleAddTracklist(tracklistStore)))
	router.Handle("PATCH /tracklists/{id}", authorized(handleUpdateTracklist(tracklistStore)))

	router.Handle("GET /tracks/mostplayed", api(handleGetMostPlayedTracks(trackStore)))
	router.Handle("GET /tracks/search", api(handleSearchTracks(trackStore)))

	router.Handle("GET /tracks/{id}", api(handleGetTrack(trackStore)))
	router.Handle("GET /tracks/{id}/tracklists", api(handleGetTracklistsByTrack(trackStore, tracklistStore)))

	router.Handle("POST /artwork", api(handlePostArtwork(artworkStore)))

	router.Handle("OPTIONS /{path...}", api(handlePreflight()))
	router.Handle("/{path...}", api(handleNotFound()))
}
