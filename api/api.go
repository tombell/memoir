package api

import (
	"log"
	"net/http"

	"github.com/matryer/way"

	"github.com/tombell/memoir/services"
)

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

	return http.ListenAndServe(addr, s.router)
}

// NewServer ...
func NewServer(cfg *Config) *Server {
	return &Server{
		logger:   cfg.Logger,
		router:   way.NewRouter(),
		services: cfg.Services,
	}
}

func getRequestID(r *http.Request) string {
	return r.Context().Value(requestIDKey).(string)
}
