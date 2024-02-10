package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/matryer/way"

	"github.com/tombell/memoir/internal/api/middleware"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/services"
)

func routes(
	logger *slog.Logger,
	router *way.Router,
	config *config.Config,
	services *services.Services,
) {
	api := []func(http.HandlerFunc) http.HandlerFunc{
		middleware.CORS(),
		middleware.RequestLogger(),
		middleware.RequestID(uuid.NewString),
		middleware.Logger(logger),
	}

	apiAuth := append(api, middleware.APIToken(config.API.Token, logger))

	router.Handle("OPTIONS", "/...", middleware.Use(handlePreflight(), middleware.CORS()))

	router.Handle("GET", "/tracklists", middleware.Use(handleGetTracklists(services), api...))
	router.Handle("GET", "/tracklists/:id", middleware.Use(handleGetTracklist(services), api...))
	router.Handle("POST", "/tracklists", middleware.Use(handleAddTracklist(services), apiAuth...))
	router.Handle("PATCH", "/tracklists/:id", middleware.Use(handleUpdateTracklist(services), apiAuth...))

	router.Handle("GET", "/tracks/mostplayed", middleware.Use(handleGetMostPlayedTracks(services), api...))
	router.Handle("GET", "/tracks/search", middleware.Use(handleSearchTracks(services), api...))

	router.Handle("GET", "/tracks/:id", middleware.Use(handleGetTrack(services), api...))
	router.Handle("GET", "/tracks/:id/tracklists", middleware.Use(handleGetTracklistsByTrack(services), api...))

	router.Handle("POST", "/artwork", middleware.Use(handlePostArtwork(services), apiAuth...))

	router.NotFound = middleware.Use(handleNotFound(), api...)
}
