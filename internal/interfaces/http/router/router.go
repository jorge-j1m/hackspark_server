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
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/handler/projects"
	"github.com/jorge-j1m/hackspark_server/internal/interfaces/http/handler/tags"
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
	usersHandler := users.NewUsersHandler(client)
	projectsHandler := projects.NewProjectsHandler(client)
	tagsHandler := tags.NewTagsHandler(client)

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
				r.Get("/{username}", usersHandler.GetUserProfile)
				r.Get("/{username}/technologies", usersHandler.GetUserTechnologies)

				r.Group(func(r chi.Router) {
					r.Use(authMiddleware.Authenticate)
					r.Get("/me", usersHandler.Me)
					r.Post("/technologies", usersHandler.AddUserTechnology)
					r.Put("/technologies/{slug}", usersHandler.UpdateUserTechnology)
					r.Delete("/technologies/{slug}", usersHandler.RemoveUserTechnology)
				})
			})

			// Project routes
			r.Route("/projects", func(r chi.Router) {
				r.Get("/", projectsHandler.ListProjects)
				r.Get("/{id}", projectsHandler.GetProject)
				r.Get("/{id}/likes", projectsHandler.GetProjectLikes)

				r.Group(func(r chi.Router) {
					r.Use(authMiddleware.Authenticate)
					r.Post("/", projectsHandler.CreateProject)
					r.Put("/{id}", projectsHandler.UpdateProject)
					r.Delete("/{id}", projectsHandler.DeleteProject)
					r.Post("/{id}/like", projectsHandler.LikeProject)
					r.Delete("/{id}/like", projectsHandler.UnlikeProject)
					r.Get("/{id}/liked", projectsHandler.CheckProjectLiked)
				})
			})

			// Tag routes
			r.Route("/tags", func(r chi.Router) {
				r.Get("/", tagsHandler.ListTags)
				r.Get("/trending", tagsHandler.GetTrendingTags)
				r.Get("/{slug}", tagsHandler.GetTag)
				r.Get("/{slug}/projects", tagsHandler.GetTagProjects)
				r.Get("/{slug}/users", tagsHandler.GetTagUsers)
			})
		})
	})

	return r
}
