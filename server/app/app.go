// Package app provides the main application HTTP handlers
package app

import (
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/db"
	"github.com/jrschumacher/dis.quest/internal/middleware"
	"github.com/jrschumacher/dis.quest/internal/pds"
	"github.com/jrschumacher/dis.quest/internal/svrlib"
)

// Router handles application-specific HTTP routes
type Router struct {
	*svrlib.Router
	dbService  *db.Service  
	pdsService pds.Service
}

// RegisterRoutes registers all application routes and returns a Router
func RegisterRoutes(mux *http.ServeMux, _ string, cfg *config.Config, dbService *db.Service, pdsService pds.Service) *Router {
	router := &Router{
		Router:     svrlib.NewRouter(mux, "/", cfg),
		dbService:  dbService,
		pdsService: pdsService,
	}

	// Create route groups
	public := middleware.PublicGroup(mux, cfg.AppEnv)
	protectedPages := middleware.ProtectedPageGroup(mux, cfg.AppEnv)
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

	// Protected API routes (require authentication)
	protectedAPI := middleware.ProtectedAPIGroup(mux)
	protectedAPI.HandleFunc("/api/topics", router.TopicsAPIHandler)
	protectedAPI.HandleFunc("/api/topics/{id}/messages", router.MessagesAPIHandler)
	protectedAPI.HandleFunc("/api/messages/{id}/like", router.LikeMessageHandler)
	protectedAPI.HandleFunc("/api/messages/reply", router.ReplyMessageHandler)
	protectedAPI.HandleFunc("/api/pds/topics", router.PDSTopicsHandler)
	protectedAPI.HandleFunc("/api/pds/record", router.PDSRecordHandler)

	// Raw routes (no middleware)
	raw.HandleFunc("/stream/topics", router.TopicsStreamHandler)
	
	// Development routes (only in development)
	if cfg.AppEnv == "development" {
		pagesWithContext.HandleFunc("/dev/pds", router.DevPDSHandler)
		raw.HandleFunc("/dev/pds/test", router.DevPDSTestHandler)
	}

	return router
}
