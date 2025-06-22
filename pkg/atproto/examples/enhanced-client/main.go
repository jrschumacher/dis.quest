// Package main demonstrates the enhanced ATProtocol client with session management.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jrschumacher/dis.quest/pkg/atproto"
	"github.com/jrschumacher/dis.quest/pkg/atproto/session"
)

func main() {
	// Example: Memory storage for development
	memoryExample()
	
	// Example: File storage for CLI applications
	fileExample()
	
	// Example: Custom session configuration
	customConfigExample()
}

func memoryExample() {
	fmt.Println("=== Memory Storage Example ===")
	
	// Create client with memory storage (default)
	config := atproto.Config{
		ClientID:       "https://myapp.example.com/client-metadata.json",
		RedirectURI:    "https://myapp.example.com/auth/callback",
		PDSEndpoint:    "https://bsky.social",
		JWKSPrivateKey: `{"keys":[...]}`, // Your JWKS private key
		JWKSPublicKey:  `{"keys":[...]}`, // Your JWKS public key
		Scope:          "atproto transition:generic",
		// SessionStorage defaults to memory storage
	}
	
	client, err := atproto.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	
	// OAuth flow would happen here...
	// authURL := client.GetAuthURL(state, codeChallenge)
	// After OAuth callback:
	// session, err := client.ExchangeCode(ctx, code, codeVerifier)
	
	fmt.Println("Memory storage client created successfully")
}

func fileExample() {
	fmt.Println("\n=== File Storage Example ===")
	
	// Create client with file storage for persistent sessions
	encryptionKey := []byte("cli-app-32-byte-encryption-key!!")
	fileStorage := session.NewFileStorage("~/.myapp/sessions", encryptionKey)
	
	config := atproto.Config{
		ClientID:       "https://myapp.example.com/client-metadata.json",
		RedirectURI:    "https://myapp.example.com/auth/callback",
		PDSEndpoint:    "https://bsky.social",
		JWKSPrivateKey: `{"keys":[...]}`,
		JWKSPublicKey:  `{"keys":[...]}`,
		Scope:          "atproto transition:generic",
		SessionStorage: fileStorage, // Use file storage
		SessionConfig: session.Config{
			TokenExpiryThreshold: 10 * time.Minute,
			CleanupInterval:     2 * time.Hour,
			MaxSessionAge:       7 * 24 * time.Hour, // 1 week
			EncryptionKey:       encryptionKey,
		},
	}
	
	client, err := atproto.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	
	fmt.Println("File storage client created successfully")
	
	// Load existing session example
	ctx := context.Background()
	sessionID := "previous-session-id"
	
	existingSession, err := client.LoadSession(ctx, sessionID)
	if err != nil {
		fmt.Printf("No existing session found: %v\n", err)
	} else {
		fmt.Printf("Loaded existing session for: %s\n", existingSession.GetUserDID())
		
		// Use the session for ATProtocol operations
		// records, err := existingSession.ListRecords(ctx, "quest.dis.topic", 10, "")
	}
}

func customConfigExample() {
	fmt.Println("\n=== Custom Configuration Example ===")
	
	// Custom session configuration with specific timeouts
	sessionConfig := session.Config{
		TokenExpiryThreshold: 2 * time.Minute,  // Refresh tokens 2 min before expiry
		CleanupInterval:     15 * time.Minute,   // Clean up expired sessions every 15 min
		MaxSessionAge:       12 * time.Hour,     // Sessions expire after 12 hours
		EncryptionKey:       []byte("my-app-specific-encryption-key!!"),
		SessionIDGenerator: func() string {
			// Custom session ID generation
			return fmt.Sprintf("myapp_%d", time.Now().UnixNano())
		},
	}
	
	config := atproto.Config{
		ClientID:       "https://myapp.example.com/client-metadata.json",
		RedirectURI:    "https://myapp.example.com/auth/callback",
		PDSEndpoint:    "https://bsky.social",
		JWKSPrivateKey: `{"keys":[...]}`,
		JWKSPublicKey:  `{"keys":[...]}`,
		SessionConfig:  sessionConfig,
		// SessionStorage defaults to memory storage
	}
	
	client, err := atproto.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	
	fmt.Println("Custom configuration client created successfully")
	
	// Access session manager for advanced operations
	sessionManager := client.GetSessionManager()
	
	// Custom token validation
	token := "example.jwt.token"
	claims, err := sessionManager.ValidateToken(token)
	if err != nil {
		fmt.Printf("Token validation failed: %v\n", err)
	} else {
		fmt.Printf("Token is valid for user: %s\n", claims.Subject)
	}
}

// Example: Web application integration (pseudo-code)
func webIntegrationExample() {
	fmt.Println("\n=== Web Integration Example ===")
	
	// For web applications, you might use cookie storage
	encryptionKey := []byte("web-app-32-byte-encryption-key!!")
	
	// This would be integrated into your HTTP handlers
	cookieConfig := session.CookieConfig{
		SessionCookieName: "myapp_session",
		DPoPCookieName:    "myapp_dpop",
		MaxAge:           3600, // 1 hour
		SecureInProd:     true,
		Path:             "/",
	}
	
	cookieStorage := session.NewCookieStorage(encryptionKey, cookieConfig)
	
	config := atproto.Config{
		ClientID:       "https://myapp.example.com/client-metadata.json",
		RedirectURI:    "https://myapp.example.com/auth/callback",
		PDSEndpoint:    "https://bsky.social",
		JWKSPrivateKey: `{"keys":[...]}`,
		JWKSPublicKey:  `{"keys":[...]}`,
		SessionStorage: cookieStorage,
	}
	
	client, err := atproto.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	
	fmt.Println("Web integration client configured")
	
	// In HTTP handlers, you would:
	// 1. Use session.WithHTTPContext(ctx, request, responseWriter) for cookie operations
	// 2. Call client.ExchangeCode() in OAuth callback
	// 3. Sessions automatically save/load from cookies
}