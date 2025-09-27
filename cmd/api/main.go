package main

import (
	"github.com/jorge-j1m/hackspark_server/internal/infrastructure/config"
	"github.com/jorge-j1m/hackspark_server/internal/infrastructure/logger"
	"github.com/jorge-j1m/hackspark_server/internal/infrastructure/server"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger
	if err := logger.Setup(cfg.LogLevel, cfg.Environment); err != nil {
		log.Fatal().Err(err).Msg("Failed to setup logger")
	}

	// Create and start server
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}

	// Wait for shutdown signal
	srv.WaitForShutdown()
}
