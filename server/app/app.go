// Package app provides the main application HTTP handlers
package app

import (
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/db"
	"github.com/jrschumacher/dis.quest/internal/middleware"
	"github.com/jrschumacher/dis.quest/internal/svrlib"
)

// Router handles application-specific HTTP routes
type Router struct {
	*svrlib.Router
	dbService *db.Service
}

// RegisterRoutes registers all application routes and returns a Router
func RegisterRoutes(mux *http.ServeMux, _ string, cfg *config.Config, dbService *db.Service) *Router {
	router := &Router{
		Router:    svrlib.NewRouter(mux, "/", cfg),
		dbService: dbService,
	}

	// Create route groups
	public := middleware.PublicGroup(mux, cfg.AppEnv)
	protectedPages := middleware.ProtectedPageGroup(mux, cfg.AppEnv)
	api := middleware.APIGroup(mux)
	raw := middleware.RawGroup(mux)

	// Public routes
	public.HandleFunc("/", router.LandingHandler)
	public.HandleFunc("/login", router.LoginHandler)

	// Protected pages
	protectedPages.HandleFunc("/discussion", router.DiscussionHandler)

	// Semi-public pages (user context but no auth required)
	pagesWithContext := middleware.NewRouteGroup(mux, 
		middleware.UserContextMiddleware,
		middleware.LayoutMiddleware(cfg.AppEnv),
	)
	pagesWithContext.HandleFunc("/topics", router.TopicsHandler)

	// API routes
	api.HandleFunc("/api/topics", router.TopicsAPIHandler)
	api.HandleFunc("/api/topics/{id}/messages", router.MessagesAPIHandler)
	api.HandleFunc("/api/messages/{id}/like", router.LikeMessageHandler)
	api.HandleFunc("/api/messages/reply", router.ReplyMessageHandler)

	// Raw routes (no middleware)
	raw.HandleFunc("/stream/topics", router.TopicsStreamHandler)

	return router
}
