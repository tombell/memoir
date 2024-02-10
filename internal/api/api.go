package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/tombell/memoir/internal/artworkstore"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/trackliststore"
	"github.com/tombell/memoir/internal/trackstore"
)

const (
	perPageTracklists = 20

	mostPlayedTracksLimit = 10
	searchResultsLimit    = 10
)

type Server struct {
	router *http.ServeMux
	server *http.Server
}

func New(
	logger *slog.Logger,
	config *config.Config,
	tracklistStore *trackliststore.Store,
	trackStore *trackstore.Store,
	artworkStore *artworkstore.Store,
) *Server {
	router := http.NewServeMux()
	server := &Server{router: router}

	server.server = &http.Server{
		Addr:         config.Address,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	routes(
		logger,
		router,
		config,
		tracklistStore,
		trackStore,
		artworkStore,
	)

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
