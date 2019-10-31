package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/matryer/way"

	"github.com/tombell/memoir/services"
)

const perPageTracklists = 10

// Config ...
type Config struct {
	Logger   *log.Logger
	Services *services.Services
}

// Server ...
type Server struct {
	logger *log.Logger
	router *way.Router
	server *http.Server

	services *services.Services
}

// Start ...
func (s *Server) Start(addr string) error {
	s.routes()

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return s.server.ListenAndServe()
}

// Shutdown ...
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	return s.server.Shutdown(ctx)
}

// NewServer ...
func NewServer(cfg *Config) *Server {
	return &Server{
		logger:   cfg.Logger,
		router:   way.NewRouter(),
		services: cfg.Services,
	}
}
