package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jorge-j1m/hackspark_server/ent/runtime"
	_ "github.com/lib/pq"

	"github.com/jorge-j1m/hackspark_server/ent"
	"github.com/jorge-j1m/hackspark_server/internal/infrastructure/config"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/router"

	"github.com/rs/zerolog/log"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	client *ent.Client
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
	// Initialize database connection
	client, err := ent.Open("postgres", s.config.DatabaseString)
	if err != nil {
		log.Fatal().Err(err).Msg("failed opening connection to postgres")
	}
	s.client = client

	ctx := context.Background()
	// Run the auto migration tool.
	// NOTE: In a production environment, it's recommended to use
	// versioned migrations instead of auto-migration.
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed creating schema resources")
	}

	// Initialize router
	r := router.New(s.config, client)

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

	// Close database connection
	if s.client != nil {
		// Blocks until all connections are returned to the pool. i.e. all transactions are committed.
		s.client.Close()
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
