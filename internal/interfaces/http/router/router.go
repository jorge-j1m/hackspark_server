package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/jorge-j1m/hackspark_server/ent"
	"github.com/jorge-j1m/hackspark_server/internal/infrastructure/config"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/handler"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/handler/auth"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/handler/users"
	cMiddleware "github.com/jorge-j1m/hackspark_server/internal/interfaces/http/middleware"
)

// New creates a new router with all routes and middleware
func New(cfg *config.Config, client *ent.Client) http.Handler {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Heartbeat("/ping"))

	// Custom middleware
	r.Use(cMiddleware.RequestID)
	r.Use(cMiddleware.Logger)
	r.Use(cMiddleware.SecurityHeaders)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Auth middleware
	authMiddleware := cMiddleware.NewAuthMiddleware(client)

	// Health check endpoint
	healthHandler := handler.NewHealthHandler(cfg)
	authHandler := auth.NewAuthHandler(client)
	usersHandler := users.NewUsersHandler()

	r.Get("/health", healthHandler.Handle)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// v1 API routes
		r.Route("/v1", func(r chi.Router) {
			// Auth routes
			r.Route("/auth", func(r chi.Router) {
				r.Post("/signup", authHandler.SignUp)
				r.Post("/login", authHandler.Login)
				r.Post("/logout", authHandler.Logout)
			})

			// User routes
			r.Route("/users", func(r chi.Router) {
				r.Use(authMiddleware.Authenticate)
				r.Get("/me", usersHandler.Me)
			})
		})
	})

	return r
}
