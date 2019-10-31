package api

import (
	"log"
	"net/http"

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

	services *services.Services
}

// Start ...
func (s *Server) Start(addr string) error {
	s.routes()

	srv := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	return srv.ListenAndServe()
}

// NewServer ...
func NewServer(cfg *Config) *Server {
	return &Server{
		logger:   cfg.Logger,
		router:   way.NewRouter(),
		services: cfg.Services,
	}
}
