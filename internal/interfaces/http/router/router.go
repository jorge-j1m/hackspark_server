package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/jorge-j1m/hackspark_server/internal/infrastructure/config"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/handler"
)

// New creates a new router with all routes and middleware
func New(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Heartbeat("/ping"))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	healthHandler := handler.NewHealthHandler(cfg)

	r.Get("/health", healthHandler.Handle)

	return r
}
