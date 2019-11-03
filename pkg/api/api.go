package api

import (
	"context"
	"net/http"
	"time"

	"github.com/matryer/way"

	"github.com/tombell/memoir/pkg/services"
)

const perPageTracklists = 10

// Server represents an API server, with a router and service dependencies.
type Server struct {
	services *services.Services
	router   *way.Router
	server   *http.Server
}

// New returns an initialised API server with the given services.
func New(services *services.Services) *Server {
	return &Server{
		services: services,
		router:   way.NewRouter(),
	}
}

// Start initialises the API server, and begins listening on the configured
// network address.
func (s *Server) Start() error {
	s.routes()

	s.server = &http.Server{
		Addr:         s.services.Config.Address,
		Handler:      s.router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	s.services.Logger.Printf("starting api server (%s) ...", s.services.Config.Address)

	return s.server.ListenAndServe()
}

// Shutdown shuts the running API server down.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.services.Logger.Println("shutting down api server...")

	return s.server.Shutdown(ctx)
}
