package api

import (
	"github.com/google/uuid"

	"github.com/tombell/memoir/internal/api/middleware"
)

func (s *Server) routes() {
	api := []middleware.Middleware{
		middleware.CORS(),
		middleware.Logging(s.services.Logger),
		middleware.RequestID(uuid.NewString, s.services.Logger),
	}

	apiAuth := append(api, middleware.APIToken(s.services.Config.API.Token, s.services.Logger))

	s.router.Handle("OPTIONS", "/...", middleware.Use(s.handlePreflight(), middleware.CORS()))

	s.router.Handle("GET", "/tracklists", middleware.Use(s.handleGetTracklists(), api...))
	s.router.Handle("GET", "/tracklists/:id", middleware.Use(s.handleGetTracklist(), api...))
	s.router.Handle("POST", "/tracklists", middleware.Use(s.handleAddTracklist(), apiAuth...))
	s.router.Handle("PATCH", "/tracklists/:id", middleware.Use(s.handleUpdateTracklist(), apiAuth...))

	s.router.Handle("GET", "/tracks/mostplayed", middleware.Use(s.handleGetMostPlayedTracks(), api...))
	s.router.Handle("GET", "/tracks/search", middleware.Use(s.handleSearchTracks(), api...))

	s.router.Handle("GET", "/tracks/:id", middleware.Use(s.handleGetTrack(), api...))
	s.router.Handle("GET", "/tracks/:id/tracklists", middleware.Use(s.handleGetTracklistsByTrack(), api...))

	s.router.Handle("POST", "/artwork", middleware.Use(s.handlePostArtwork(), apiAuth...))

	s.router.NotFound = middleware.Use(s.handleNotFOund(), api...)
}
