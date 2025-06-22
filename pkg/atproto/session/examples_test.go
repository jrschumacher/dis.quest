package session_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jrschumacher/dis.quest/pkg/atproto/session"
)

// Example: Memory storage for development/testing
func ExampleMemoryStorage() {
	// Create memory storage
	storage := session.NewMemoryStorage()
	
	// Create session data
	dpopKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	data := &session.Data{
		SessionID:    "test-session-123",
		UserDID:      "did:plc:example123",
		Handle:       "user.bsky.social",
		AccessToken:  "access-token-here",
		RefreshToken: "refresh-token-here",
		DPoPKey:      dpopKey,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	ctx := context.Background()
	
	// Store session
	if err := storage.Store(ctx, "test-session-123", data); err != nil {
		log.Fatal(err)
	}
	
	// Load session
	loadedData, err := storage.Load(ctx, "test-session-123")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Loaded session for user: %s\n", loadedData.Handle)
	// Output: Loaded session for user: user.bsky.social
}

// Example: File storage for CLI applications
func ExampleFileStorage() {
	// Create file storage in a temporary directory
	encryptionKey := []byte("example-32-byte-key-for-encryption")
	storage := session.NewFileStorage("/tmp/sessions", encryptionKey)
	
	// Create session data
	dpopKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	data := &session.Data{
		SessionID:    "cli-session-456",
		UserDID:      "did:plc:cliuser789",
		Handle:       "cliuser.bsky.social",
		AccessToken:  "cli-access-token",
		RefreshToken: "cli-refresh-token",
		DPoPKey:      dpopKey,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	ctx := context.Background()
	
	// Store session (will create file)
	if err := storage.Store(ctx, "cli-session-456", data); err != nil {
		log.Fatal(err)
	}
	
	// Load session (reads from file)
	loadedData, err := storage.Load(ctx, "cli-session-456")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("CLI session for: %s\n", loadedData.Handle)
	// Output: CLI session for: cliuser.bsky.social
}

// Example: Cookie storage for web applications
func ExampleCookieStorage() {
	// Create cookie storage
	encryptionKey := []byte("web-app-32-byte-encryption-key!!")
	cookieConfig := session.CookieConfig{
		SessionCookieName: "myapp_session",
		DPoPCookieName:    "myapp_dpop",
		MaxAge:           3600, // 1 hour
		SecureInProd:     true,
		SameSite:         http.SameSiteLaxMode,
		Path:             "/",
	}
	storage := session.NewCookieStorage(encryptionKey, cookieConfig)
	
	// Create session data
	dpopKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	data := &session.Data{
		SessionID:    "web-session-789",
		UserDID:      "did:plc:webuser456",
		Handle:       "webuser.bsky.social",
		AccessToken:  "web-access-token",
		RefreshToken: "web-refresh-token",
		DPoPKey:      dpopKey,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	// For cookie storage, you need HTTP context
	// In real usage, this would be in an HTTP handler
	ctx := context.Background()
	// ctx = session.WithHTTPContext(ctx, request, responseWriter)
	
	fmt.Printf("Cookie storage configured for: %s\n", cookieConfig.SessionCookieName)
	// Output: Cookie storage configured for: myapp_session
}

// Example: Session manager usage
func ExampleSessionManager() {
	// Create storage and config
	storage := session.NewMemoryStorage()
	config := session.Config{
		TokenExpiryThreshold: 5 * time.Minute,
		EncryptionKey:       []byte("session-manager-encryption-key!"),
		CleanupInterval:     30 * time.Minute,
		MaxSessionAge:       24 * time.Hour,
	}
	
	// Create manager (would normally pass a real OAuth provider and XRPC client)
	manager := session.NewManager(storage, config, nil, nil)
	defer manager.Close()
	
	// Create session from OAuth token result
	tokenResult := &session.TokenResult{
		AccessToken:  "oauth-access-token",
		RefreshToken: "oauth-refresh-token",
		UserDID:      "did:plc:manager123",
		Handle:       "manager.bsky.social",
		ExpiresIn:    3600,
	}
	
	ctx := context.Background()
	sessionInstance, err := manager.CreateSession(ctx, tokenResult)
	if err != nil {
		log.Fatal(err)
	}
	
	// Use session for ATProtocol operations
	fmt.Printf("Created session for: %s\n", sessionInstance.GetHandle())
	fmt.Printf("Session ID: %s\n", sessionInstance.GetSessionID())
	
	// Example output:
	// Created session for: manager.bsky.social
	// Session ID: <generated-session-id>
}

// Example: Different storage backends for different use cases
func ExampleStorageBackends() {
	encryptionKey := []byte("universal-encryption-key-32-byte")
	
	// Memory storage - for development/testing
	memStorage := session.NewMemoryStorage()
	fmt.Printf("Memory storage: %T\n", memStorage)
	
	// File storage - for CLI applications
	fileStorage := session.NewFileStorage("~/.myapp/sessions", encryptionKey)
	fmt.Printf("File storage: %T\n", fileStorage)
	
	// Cookie storage - for web applications
	cookieConfig := session.CookieConfig{
		SessionCookieName: "app_session",
		MaxAge:           3600,
		SecureInProd:     true,
	}
	cookieStorage := session.NewCookieStorage(encryptionKey, cookieConfig)
	fmt.Printf("Cookie storage: %T\n", cookieStorage)
	
	// All implement the same Storage interface
	var storage session.Storage
	storage = memStorage    // ✓
	storage = fileStorage   // ✓ 
	storage = cookieStorage // ✓
	
	fmt.Printf("All implement session.Storage: %T\n", storage)
	
	// Output:
	// Memory storage: *session.MemoryStorage
	// File storage: *session.FileStorage
	// Cookie storage: *session.CookieStorage
	// All implement session.Storage: *session.CookieStorage
}