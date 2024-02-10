package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/tombell/memoir/internal/api/middleware"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/services"
)

func routes(
	logger *slog.Logger,
	router *http.ServeMux,
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

	router.HandleFunc("OPTIONS /{path...}", middleware.Use(handlePreflight(), middleware.CORS()))

	router.HandleFunc("GET /tracklists", middleware.Use(handleGetTracklists(services), api...))
	router.HandleFunc("GET /tracklists/{id}", middleware.Use(handleGetTracklist(services), api...))
	router.HandleFunc("POST /tracklists", middleware.Use(handleAddTracklist(services), apiAuth...))
	router.HandleFunc("PATCH /tracklists/{id}", middleware.Use(handleUpdateTracklist(services), apiAuth...))

	router.HandleFunc("GET /tracks/mostplayed", middleware.Use(handleGetMostPlayedTracks(services), api...))
	router.HandleFunc("GET /tracks/search", middleware.Use(handleSearchTracks(services), api...))

	router.HandleFunc("GET /tracks/{id}", middleware.Use(handleGetTrack(services), api...))
	router.HandleFunc("GET /tracks/{id}/tracklists", middleware.Use(handleGetTracklistsByTrack(services), api...))

	router.HandleFunc("POST /artwork", middleware.Use(handlePostArtwork(services), apiAuth...))

	router.HandleFunc("/{path...}", middleware.Use(handleNotFound(), api...))
}
