package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/stores/artworkstore"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
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

func (s *Server) Run() error {
	if err := s.server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			return err
		}
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
