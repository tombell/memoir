package api

import (
	"context"
	"net/http"
	"time"

	"github.com/matryer/way"

	"github.com/tombell/memoir/internal/services"
)

const (
	perPageTracklists     = 20
	perPageTracks         = 10
	mostPlayedTracksLimit = 10
	searchResultsLimit    = 10
)

type Server struct {
	*services.Services

	router *way.Router
	server *http.Server
}

func New(services *services.Services) *Server {
	return &Server{
		Services: services,
		router:   way.NewRouter(),
	}
}

func (s *Server) Start() error {
	s.routes()

	s.server = &http.Server{
		Addr:         s.Config.Address,
		Handler:      s.router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	s.Logger.Info("starting api server", "addr", s.Config.Address)

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.Logger.Info("shutting down api server")

	return s.server.Shutdown(ctx)
}
