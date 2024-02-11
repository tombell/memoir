package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/tombell/middle"
	"github.com/tombell/middle/ware"

	"github.com/tombell/memoir/internal/api/handlers"
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
	api := middle.Use(
		ware.Recovery(),
		middleware.CORS(),
		ware.RequestLogging(),
		ware.RequestID(uuid.NewString),
		ware.Logger(logger),
	)

	authorized := middle.Use(
		middleware.Authorize(config.API.Token),
		api,
	)

	router.Handle("GET /tracklists", api(handlers.GetTracklists(tracklistStore)))
	router.Handle("GET /tracklists/{id}", api(handlers.GetTracklist(tracklistStore)))
	router.Handle("POST /tracklists", authorized(handlers.PostTracklist(tracklistStore)))
	router.Handle("PATCH /tracklists/{id}", authorized(handlers.PatchTracklist(tracklistStore)))

	router.Handle("GET /tracks/mostplayed", api(handlers.GetMostPlayedTracks(trackStore)))
	router.Handle("GET /tracks/search", api(handlers.GetTrackSearch(trackStore)))
	router.Handle("GET /tracks/{id}", api(handlers.GetTrack(trackStore)))
	router.Handle("GET /tracks/{id}/tracklists", api(handlers.GetTracklistsByTrack(trackStore, tracklistStore)))

	router.Handle("POST /artwork", api(handlers.PostArtwork(artworkStore)))

	router.Handle("OPTIONS /{path...}", api(handlers.Preflight()))
	router.Handle("/{path...}", api(handlers.NotFound()))
}
