package app

import (
	"net/http"
	"testing"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/db"
	"github.com/jrschumacher/dis.quest/internal/middleware"
	"github.com/jrschumacher/dis.quest/internal/pds"
)

// RegisterTestRoutes registers routes with test middleware for testing
func RegisterTestRoutes(mux *http.ServeMux, _ string, _ *config.Config, dbService *db.Service, testUserDID string) *Router {
	// Create a mock PDS service for integration tests
	pdsService := pds.NewMockService()
	
	router := &Router{
		Router:     nil, // We don't need the full router for tests
		dbService:  dbService,
		pdsService: pdsService,
	}

	// Public routes (same as production)
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("test home"))
	}))
	mux.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("test login"))
	}))
	
	// Protected routes with test middleware
	testChain := middleware.TestProtectedChain(testUserDID)
	
	mux.Handle("/discussion", testChain.ThenFunc(router.DiscussionHandler))
	mux.Handle("/topics", testChain.ThenFunc(router.TopicsHandler))
	mux.Handle("/api/topics", testChain.ThenFunc(router.TopicsAPIHandler))
	mux.Handle("/api/topics/{id}/messages", testChain.ThenFunc(router.MessagesAPIHandler))

	return router
}

// CreateTestServer creates a test server with test routes
func CreateTestServer(t *testing.T, dbService *db.Service, testUserDID string) *http.ServeMux {
	t.Helper()

	cfg := &config.Config{
		AppEnv:      "test",
		DatabaseURL: ":memory:",
	}

	mux := http.NewServeMux()
	RegisterTestRoutes(mux, "/", cfg, dbService, testUserDID)
	
	return mux
}