package server

import (
	"log"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/flightpath-dev/flightpath-server/internal/config"
	"github.com/flightpath-dev/flightpath-server/internal/middleware"
)

// Server represents the Flightpath server
type Server struct {
	config       *config.Config
	dependencies *Dependencies
	mux          *http.ServeMux
	logger       *log.Logger
}

// New creates a new Server instance
func New(cfg *config.Config) *Server {
	deps := NewDependencies(cfg)

	return &Server{
		config:       cfg,
		dependencies: deps,
		mux:          http.NewServeMux(),
		logger:       deps.GetLogger(),
	}
}

// RegisterService registers a Connect service handler
func (s *Server) RegisterService(path string, handler http.Handler) {
	s.logger.Printf("Registering service: %s", path)
	s.mux.Handle(path, handler)
}

// buildHandler builds the final HTTP handler with all middleware
func (s *Server) buildHandler() http.Handler {
	// Start with the mux
	handler := http.Handler(s.mux)

	// Add middleware in reverse order (last applied first)
	handler = middleware.CORS(s.config.Server.CORSOrigins)(handler)
	handler = middleware.Logging(s.logger)(handler)
	handler = middleware.Recovery(s.logger)(handler)

	// Wrap with h2c (HTTP/2 Cleartext) for Connect protocol
	return h2c.NewHandler(handler, &http2.Server{})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := s.config.ServerAddr()
	handler := s.buildHandler()

	s.logger.Printf("🚀 Flightpath server starting on %s", addr)
	s.logger.Printf("📡 Ready to accept Connect protocol requests")

	return http.ListenAndServe(addr, handler)
}

// GetDependencies returns the shared dependencies
func (s *Server) GetDependencies() *Dependencies {
	return s.dependencies
}
