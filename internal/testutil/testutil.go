package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/db"
)

// TestDatabase creates an in-memory SQLite database for testing
func TestDatabase(t *testing.T) *db.Service {
	t.Helper()

	cfg := &config.Config{
		DatabaseURL: ":memory:",
		AppEnv:     "test",
	}

	dbService, err := db.NewService(cfg)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test schema
	if err := CreateTestSchema(dbService.DB()); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Cleanup function
	t.Cleanup(func() {
		if err := dbService.Close(); err != nil {
			t.Errorf("Failed to close test database: %v", err)
		}
	})

	return dbService
}

// TestServer creates a test HTTP server - mux should be set up by the test
func TestServer(t *testing.T, mux *http.ServeMux) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(mux)
	t.Cleanup(func() {
		server.Close()
	})

	return server
}

// TestRequest represents a test HTTP request
type TestRequest struct {
	Method      string
	Path        string
	Body        string
	Headers     map[string]string
	AuthToken   string // JWT token for authenticated requests
}

// TestResponse represents the expected response
type TestResponse struct {
	StatusCode int
	BodyContains []string
	HeaderContains map[string]string
	JSONPath     map[string]interface{} // For JSON response validation
}

// APITestCase represents a complete API test scenario
type APITestCase struct {
	Name        string
	Description string
	Setup       func(t *testing.T, dbService *db.Service) // Database setup
	Request     TestRequest
	Expected    TestResponse
	Cleanup     func(t *testing.T, dbService *db.Service) // Optional cleanup
}

// RunAPITest executes an API test case
// TODO: Implement when needed for more complex test scenarios

// CreateTestUser creates a test user and returns a valid JWT token
func CreateTestUser(t *testing.T, dbService *db.Service) string {
	t.Helper()
	
	// This would create a test user and return a valid JWT
	// Implementation depends on your auth system
	return "test-jwt-token"
}

// CreateTestTopic creates a test topic in the database
func CreateTestTopic(t *testing.T, dbService *db.Service, userDID string) db.Topic {
	t.Helper()

	ctx := context.Background()
	params := db.CreateTopicParams{
		Did:            userDID,
		Rkey:           fmt.Sprintf("test-topic-%s", t.Name()),
		Subject:        "Test Topic",
		InitialMessage: "This is a test topic",
		Category:       sql.NullString{String: "test", Valid: true},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	topic, err := dbService.Queries().CreateTopic(ctx, params)
	if err != nil {
		t.Fatalf("Failed to create test topic: %v", err)
	}

	return topic
}

