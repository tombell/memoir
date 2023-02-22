package api

import (
	"github.com/gofrs/uuid"

	"github.com/tombell/memoir/internal/api/middleware"
)

func requestID() string {
	id, _ := uuid.NewV4()
	return id.String()
}

func (s *Server) routes() {
	api := []middleware.Middleware{
		middleware.CORS(),
		middleware.Logging(s.Logger),
		middleware.RequestID(requestID, s.Logger),
	}

	apiAuth := []middleware.Middleware{
		middleware.CORS(),
		middleware.APIToken(s.Logger, s.Config.API.Token),
		middleware.Logging(s.Logger),
		middleware.RequestID(requestID, s.Logger),
	}

	s.router.Handle("OPTIONS", "/...", middleware.Use(s.handlePreflight(), middleware.CORS()))

	s.router.Handle("GET", "/tracklists", middleware.Use(s.handleGetTracklists(), api...))
	s.router.Handle("POST", "/tracklists", middleware.Use(s.handlePostTracklists(), apiAuth...))
	s.router.Handle("GET", "/tracklists/:id", middleware.Use(s.handleGetTracklist(), api...))
	s.router.Handle("PATCH", "/tracklists/:id", middleware.Use(s.handlePatchTracklist(), apiAuth...))

	s.router.Handle("GET", "/tracks/mostplayed", middleware.Use(s.handleGetMostPlayedTracks(), api...))
	s.router.Handle("GET", "/tracks/search", middleware.Use(s.handleSearchTracks(), api...))

	s.router.Handle("GET", "/tracks/:id/tracklists", middleware.Use(s.handleGetTracklistsByTrack(), api...))
	s.router.Handle("GET", "/tracks/:id", middleware.Use(s.handleGetTrack(), api...))

	s.router.Handle("POST", "/artwork", middleware.Use(s.handlePostArtwork(), apiAuth...))

	s.router.NotFound = middleware.Use(s.handleNotFOund(), api...)
}
