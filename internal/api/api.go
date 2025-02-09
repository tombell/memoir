package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/stores/artworkstore"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

// Server wraps an HTTP server and multiplexer.
type Server struct {
	router *http.ServeMux
	server *http.Server
}

// New returns a new instance of Server. It configures the routes of the
// application.
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
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
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

// Run starts the HTTP server listening and serving requests.
func (s *Server) Run() error {
	if err := s.server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			return fmt.Errorf("could not listen and serve: %w", err)
		}
	}

	return nil
}

// Shutdown attempts to gracefully shutdown the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
