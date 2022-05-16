package api

import (
	"github.com/gofrs/uuid"
	"github.com/tombell/mw"
)

func requestID() string {
	id, _ := uuid.NewV4()
	return id.String()
}

func (s *Server) routes() {
	api := []mw.Middleware{
		mw.CORS(),
		mw.Logging(s.services.Logger),
		mw.RequestID(requestID),
	}

	apiAuth := []mw.Middleware{
		mw.CORS(),
		mw.Auth(s.services.Logger, s.services.Config.API.Token),
		mw.Logging(s.services.Logger),
		mw.RequestID(requestID),
	}

	s.router.Handle("OPTIONS", "/...", mw.Use(s.handlePreflight(), mw.CORS()))

	s.router.Handle("GET", "/tracklists", mw.Use(s.handleGetTracklists(), api...))
	s.router.Handle("POST", "/tracklists", mw.Use(s.handlePostTracklists(), apiAuth...))
	s.router.Handle("GET", "/tracklists/:id", mw.Use(s.handleGetTracklist(), api...))
	s.router.Handle("PATCH", "/tracklists/:id", mw.Use(s.handlePatchTracklist(), apiAuth...))

	s.router.Handle("GET", "/tracks/mostplayed", mw.Use(s.handleGetMostPlayedTracks(), api...))
	s.router.Handle("GET", "/tracks/search", mw.Use(s.handleSearchTracks(), api...))

	s.router.Handle("GET", "/tracks/:id/tracklists", mw.Use(s.handleGetTracklistsByTrack(), api...))
	s.router.Handle("GET", "/tracks/:id", mw.Use(s.handleGetTrack(), api...))

	s.router.Handle("POST", "/artwork", mw.Use(s.handlePostArtwork(), apiAuth...))
}
