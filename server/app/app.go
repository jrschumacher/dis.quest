// Package app provides the main application HTTP handlers
package app

import (
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/db"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/middleware"
	"github.com/jrschumacher/dis.quest/pkg/atproto"
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
func RegisterRoutes(mux *http.ServeMux, _ string, cfg *config.Config, dbService *db.Service, pdsService pds.Service, atprotoClient *atproto.Client) *Router {
	router := &Router{
		Router:     svrlib.NewRouter(mux, "/", cfg),
		dbService:  dbService,
		pdsService: pdsService,
	}

	// Use passed atproto client for token refresh
	if atprotoClient == nil {
		logger.Error("ATProtocol client is nil - token refresh will be disabled")
	}

	// Create route groups
	public := middleware.PublicGroup(mux, cfg.AppEnv)
	protectedPages := middleware.ProtectedPageGroup(mux, cfg.AppEnv)
	raw := middleware.RawGroup(mux)

	// Create protected routes with automatic token refresh
	var protectedWithRefresh *middleware.RouteGroup
	if atprotoClient != nil {
		protectedWithRefresh = middleware.AutoRefreshGroup(mux, atprotoClient,
			middleware.UserContextMiddleware,
			middleware.AuthMiddleware,
			middleware.LayoutMiddleware(cfg.AppEnv),
		)
	} else {
		// Fallback to regular protected routes if OAuth service creation failed
		protectedWithRefresh = protectedPages
	}

	// Public routes
	public.HandleFunc("/", router.LandingHandler)
	public.HandleFunc("/login", router.LoginHandler)

	// Protected pages with auto token refresh
	protectedWithRefresh.HandleFunc("/discussion", router.DiscussionHandler)

	// Semi-public pages (user context but no auth required)
	pagesWithContext := middleware.NewRouteGroup(mux, 
		middleware.UserContextMiddleware,
		middleware.LayoutMiddleware(cfg.AppEnv),
	)
	pagesWithContext.HandleFunc("/topics", router.TopicsHandler)

	// Protected API routes with auto token refresh
	var protectedAPIWithRefresh *middleware.RouteGroup
	if atprotoClient != nil {
		protectedAPIWithRefresh = middleware.AutoRefreshGroup(mux, atprotoClient,
			middleware.UserContextMiddleware,
			middleware.AuthMiddleware,
		)
	} else {
		protectedAPIWithRefresh = middleware.ProtectedAPIGroup(mux)
	}
	protectedAPIWithRefresh.HandleFunc("/api/topics", router.TopicsAPIHandler)
	protectedAPIWithRefresh.HandleFunc("/api/topics/{id}/messages", router.MessagesAPIHandler)
	protectedAPIWithRefresh.HandleFunc("/api/messages/{id}/like", router.LikeMessageHandler)
	protectedAPIWithRefresh.HandleFunc("/api/messages/reply", router.ReplyMessageHandler)
	protectedAPIWithRefresh.HandleFunc("/api/pds/topics", router.PDSTopicsHandler)
	protectedAPIWithRefresh.HandleFunc("/api/pds/record", router.PDSRecordHandler)

	// Raw routes (no middleware)
	raw.HandleFunc("/stream/topics", router.TopicsStreamHandler)
	
	// Development routes (only in development) with auto token refresh
	if cfg.AppEnv == "development" {
		pagesWithContext.HandleFunc("/dev/pds", router.DevPDSHandler)
		if atprotoClient != nil {
			// Use auto-refresh for PDS test operations since they use tokens
			devAPIWithRefresh := middleware.AutoRefreshGroup(mux, atprotoClient)
			devAPIWithRefresh.HandleFunc("/dev/pds/test", router.DevPDSTestHandler)
		} else {
			raw.HandleFunc("/dev/pds/test", router.DevPDSTestHandler)
		}
	}

	return router
}
