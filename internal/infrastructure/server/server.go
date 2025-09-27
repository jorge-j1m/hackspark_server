package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jorge-j1m/hackspark_server/internal/infrastructure/config"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/router"

	"github.com/rs/zerolog/log"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	config *config.Config
}

// New creates a new server instance
func New(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
	}
}

// Start initializes and starts the server
func (s *Server) Start() error {
	// Initialize router
	r := router.New(s.config)

	// Configure HTTP server
	s.server = &http.Server{
		Addr:    ":" + s.config.Port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("address", s.server.Addr).Msg("Starting HTTP server")
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info().Msg("Shutting down server...")

	// Shutdown HTTP server
	if err := s.server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
	}

	log.Info().Msg("Server shutdown complete")
}

// WaitForShutdown waits for termination signals and triggers graceful shutdown
func (s *Server) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info().Str("signal", sig.String()).Msg("Received termination signal")
	s.Shutdown()
}
