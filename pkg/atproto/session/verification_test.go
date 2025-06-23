package session

import (
	"context"
	"testing"
	"time"
)

// TestInterfaceCompilation verifies that all interfaces compile correctly.
func TestInterfaceCompilation(t *testing.T) {
	// Test storage interface
	var storage Storage = NewMemoryStorage()
	if storage == nil {
		t.Fatal("Memory storage should not be nil")
	}

	// Test manager interface
	config := Config{
		TokenExpiryThreshold: 5 * time.Minute,
		EncryptionKey:       []byte("test-key-32-bytes-long-padding!!"),
	}
	
	var manager Manager = NewManager(storage, config, nil, nil)
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	// Test session creation (should compile without errors)
	tokenResult := &TokenResult{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		UserDID:      "did:plc:test123",
		Handle:       "test.bsky.social",
		ExpiresIn:    3600,
	}

	ctx := context.Background()
	session, err := manager.CreateSession(ctx, tokenResult)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test session interface methods
	if session.GetUserDID() != "did:plc:test123" {
		t.Errorf("Expected DID 'did:plc:test123', got '%s'", session.GetUserDID())
	}

	if session.GetAccessToken() != "test-access-token" {
		t.Errorf("Expected access token 'test-access-token', got '%s'", session.GetAccessToken())
	}

	// Test session operations
	err = session.Save(ctx)
	if err != nil {
		t.Errorf("Failed to save session: %v", err)
	}

	// Test manager operations
	loadedSession, err := manager.LoadSession(ctx, session.GetSessionID())
	if err != nil {
		t.Errorf("Failed to load session: %v", err)
	}

	if loadedSession.GetUserDID() != session.GetUserDID() {
		t.Errorf("Loaded session DID mismatch")
	}

	// Clean up
	err = manager.DeleteSession(ctx, session.GetSessionID())
	if err != nil {
		t.Errorf("Failed to delete session: %v", err)
	}

	manager.Close()
}

// TestStorageBackends verifies all storage backends implement the interface correctly.
func TestStorageBackends(t *testing.T) {
	encryptionKey := []byte("test-encryption-key-32-bytes!!")

	// Test all storage backends implement Storage interface
	var storage Storage

	// Memory storage
	storage = NewMemoryStorage()
	if storage == nil {
		t.Error("Memory storage should implement Storage interface")
	}

	// File storage  
	storage = NewFileStorage("/tmp/test-sessions", encryptionKey)
	if storage == nil {
		t.Error("File storage should implement Storage interface")
	}

	// Cookie storage
	cookieConfig := CookieConfig{
		SessionCookieName: "test_session",
		MaxAge:           3600,
	}
	storage = NewCookieStorage(encryptionKey, cookieConfig)
	if storage == nil {
		t.Error("Cookie storage should implement Storage interface")
	}
}