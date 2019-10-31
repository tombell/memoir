package api

import (
	"context"
	"net/http"
	"time"

	"github.com/matryer/way"

	"github.com/tombell/memoir/services"
)

const perPageTracklists = 10

// Server ...
type Server struct {
	services *services.Services
	router   *way.Router
	server   *http.Server
}

// New ...
func New(services *services.Services) *Server {
	return &Server{
		services: services,
		router:   way.NewRouter(),
	}
}

// Start ...
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

// Shutdown ...
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.services.Logger.Println("shutting down api server...")

	return s.server.Shutdown(ctx)
}
