// This file verifies that the session management system compiles correctly
package main

import (
	"context"
	"log"

	"github.com/jrschumacher/dis.quest/pkg/atproto"
	"github.com/jrschumacher/dis.quest/pkg/atproto/session"
)

func main() {
	// Test that all the interfaces compile correctly
	
	// 1. Test session storage creation
	storage := session.NewMemoryStorage()
	log.Printf("Memory storage created: %T", storage)
	
	// 2. Test client configuration
	config := atproto.Config{
		ClientID:       "test-client",
		RedirectURI:    "https://example.com/callback",
		PDSEndpoint:    "https://bsky.social",
		JWKSPrivateKey: `{"keys":[]}`,
		JWKSPublicKey:  `{"keys":[]}`,
		SessionStorage: storage,
	}
	
	// 3. Test client creation
	client, err := atproto.New(config)
	if err != nil {
		log.Printf("Client creation would fail with: %v", err)
	} else {
		log.Printf("Client created successfully: %T", client)
		defer client.Close()
	}
	
	// 4. Test session manager access
	if client != nil {
		sessionManager := client.GetSessionManager()
		log.Printf("Session manager: %T", sessionManager)
	}
	
	// 5. Test direct session manager creation
	sessionConfig := session.Config{
		EncryptionKey: []byte("test-key-32-bytes-long-padding!!"),
	}
	
	directManager := session.NewManager(storage, sessionConfig, nil, nil)
	log.Printf("Direct session manager: %T", directManager)
	defer directManager.Close()
	
	// 6. Test token result creation
	tokenResult := &session.TokenResult{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		UserDID:      "did:plc:test",
		ExpiresIn:    3600,
	}
	
	ctx := context.Background()
	sess, err := directManager.CreateSession(ctx, tokenResult)
	if err != nil {
		log.Printf("Session creation would fail with: %v", err)
	} else {
		log.Printf("Session created: %s", sess.GetSessionID())
		
		// Test session interface methods
		log.Printf("User DID: %s", sess.GetUserDID())
		log.Printf("Access Token: %s", sess.GetAccessToken())
		log.Printf("Refresh Token: %s", sess.GetRefreshToken())
		log.Printf("DPoP Key available: %t", sess.GetDPoPKey() != nil)
	}
	
	log.Println("✅ All interfaces compile successfully!")
}