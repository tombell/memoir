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
	api := []func(http.HandlerFunc) http.HandlerFunc{
		middleware.CORS(),
		middleware.RequestLogger(),
		middleware.RequestID(uuid.NewString),
		middleware.Logger(logger),
	}

	apiAuth := append(api, middleware.APIToken(config.API.Token, logger))

	router.HandleFunc("OPTIONS /{path...}", middleware.Use(handlePreflight(), middleware.CORS()))

	router.HandleFunc("GET /tracklists", middleware.Use(handleGetTracklists(tracklistStore), api...))
	router.HandleFunc("GET /tracklists/{id}", middleware.Use(handleGetTracklist(tracklistStore), api...))
	router.HandleFunc("POST /tracklists", middleware.Use(handleAddTracklist(tracklistStore), apiAuth...))
	router.HandleFunc("PATCH /tracklists/{id}", middleware.Use(handleUpdateTracklist(tracklistStore), apiAuth...))

	router.HandleFunc("GET /tracks/mostplayed", middleware.Use(handleGetMostPlayedTracks(trackStore), api...))
	router.HandleFunc("GET /tracks/search", middleware.Use(handleSearchTracks(trackStore), api...))

	router.HandleFunc("GET /tracks/{id}", middleware.Use(handleGetTrack(trackStore), api...))
	router.HandleFunc("GET /tracks/{id}/tracklists", middleware.Use(handleGetTracklistsByTrack(trackStore, tracklistStore), api...))

	router.HandleFunc("POST /artwork", middleware.Use(handlePostArtwork(artworkStore), apiAuth...))

	router.HandleFunc("/{path...}", middleware.Use(handleNotFound(), api...))
}
