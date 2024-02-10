package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/matryer/way"

	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/services"
)

const (
	perPageTracklists = 20

	mostPlayedTracksLimit = 10
	searchResultsLimit    = 10
)

type Server struct {
	router *way.Router
	server *http.Server
}

func New(
	logger *slog.Logger,
	config *config.Config,
	services *services.Services,
) *Server {
	router := way.NewRouter()
	server := &Server{router: router}

	server.server = &http.Server{
		Addr:         config.Address,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	routes(logger, router, services.Config, services)

	return server
}

func (s *Server) Start(logger *slog.Logger) error {
	logger.Info("starting api server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context, logger *slog.Logger) error {
	logger.Info("shutting down api server")
	return s.server.Shutdown(ctx)
}
