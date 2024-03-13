package web

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/tombell/middle"
	"github.com/tombell/middle/ware"

	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/stores/artworkstore"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
	"github.com/tombell/memoir/internal/web/handlers"
	"github.com/tombell/memoir/internal/web/middleware"
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
		ware.CORS(ware.CORSOptions{
			AllowedMethods: []string{"GET", "POST", "PATCH"},
			AllowedHeaders: []string{"API-Token"},
		}),
		ware.RequestLogging(),
		ware.RequestID(uuid.NewString),
		ware.Logger(logger),
	)

	authorized := middle.Use(
		middleware.Authorize(config.API.Token),
		api,
	)

	router.Handle("GET /api/tracklists", api(handlers.GetTracklists(tracklistStore)))
	router.Handle("GET /api/tracklists/{id}", api(handlers.GetTracklist(tracklistStore)))
	router.Handle("POST /api/tracklists", authorized(handlers.PostTracklist(tracklistStore)))
	router.Handle("PATCH /api/tracklists/{id}", authorized(handlers.PatchTracklist(tracklistStore)))

	router.Handle("GET /api/tracks/mostplayed", api(handlers.GetMostPlayedTracks(trackStore)))
	router.Handle("GET /api/tracks/search", api(handlers.GetTrackSearch(trackStore)))
	router.Handle("GET /api/tracks/{id}", api(handlers.GetTrack(trackStore)))
	router.Handle("GET /api/tracks/{id}/tracklists", api(handlers.GetTracklistsByTrack(trackStore, tracklistStore)))

	router.Handle("POST /api/artwork", authorized(handlers.PostArtwork(artworkStore)))

	router.Handle("/{path...}", api(handlers.NotFound()))
}
