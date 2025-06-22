// Package auth - SessionWrapper for backward compatibility with web application
package auth

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jrschumacher/dis.quest/pkg/atproto"
	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
	"github.com/jrschumacher/dis.quest/pkg/atproto/session"
)

// SessionWrapper wraps session.Session while maintaining cookie-based session management
// This provides enhanced functionality while preserving existing session interface
type SessionWrapper struct {
	atprotoSession session.Session  // Use the new session interface
	accessToken    string
	refreshToken   string
	userDID        string
	dpopKey        *ecdsa.PrivateKey
}

// NewSessionWrapper creates a new session wrapper from authentication tokens
func NewSessionWrapper(accessToken, refreshToken, userDID string, dpopKey *ecdsa.PrivateKey, atprotoClient *atproto.Client) (*SessionWrapper, error) {
	// Create the internal atproto.Session
	// Note: We'll need to create this through the client's exchange process or mock it for compatibility
	// For now, we create a wrapper that manages both interfaces
	wrapper := &SessionWrapper{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		userDID:      userDID,
		dpopKey:      dpopKey,
	}
	
	// If we have an atproto client, we can create a proper session
	// This will be populated during OAuth flow
	if atprotoClient != nil {
		// The atproto.Session would typically be created during ExchangeCode
		// For now, we store the client reference for later use
		wrapper.atprotoSession = nil // Will be set during proper OAuth flow
	}
	
	return wrapper, nil
}

// GetAccessToken returns the current access token
func (sw *SessionWrapper) GetAccessToken() string {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.GetAccessToken()
	}
	return sw.accessToken
}

// GetRefreshToken returns the current refresh token
func (sw *SessionWrapper) GetRefreshToken() string {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.GetRefreshToken()
	}
	return sw.refreshToken
}

// GetUserDID returns the user's DID
func (sw *SessionWrapper) GetUserDID() string {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.GetUserDID()
	}
	return sw.userDID
}

// GetDPoPKey returns the DPoP private key
func (sw *SessionWrapper) GetDPoPKey() *ecdsa.PrivateKey {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.GetDPoPKey()
	}
	return sw.dpopKey
}

// CreateRecord creates a record using the internal session if available
func (sw *SessionWrapper) CreateRecord(collection, rkey string, record interface{}) (*session.RecordResult, error) {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.CreateRecord(context.Background(), collection, rkey, record)
	}
	return nil, fmt.Errorf("atproto session not available for CreateRecord")
}

// GetRecord retrieves a record using the internal session if available
func (sw *SessionWrapper) GetRecord(collection, rkey string, result interface{}) error {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.GetRecord(context.Background(), collection, rkey, result)
	}
	return fmt.Errorf("atproto session not available for GetRecord")
}

// ListRecords lists records using the internal session if available
func (sw *SessionWrapper) ListRecords(collection string, limit int, cursor string) (*session.ListRecordsResult, error) {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.ListRecords(context.Background(), collection, limit, cursor)
	}
	return nil, fmt.Errorf("atproto session not available for ListRecords")
}

// UpdateRecord updates a record using the internal session if available
func (sw *SessionWrapper) UpdateRecord(collection, rkey string, record interface{}) (*session.RecordResult, error) {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.UpdateRecord(context.Background(), collection, rkey, record)
	}
	return nil, fmt.Errorf("atproto session not available for UpdateRecord")
}

// DeleteRecord deletes a record using the internal session if available
func (sw *SessionWrapper) DeleteRecord(collection, rkey string) error {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.DeleteRecord(context.Background(), collection, rkey)
	}
	return fmt.Errorf("atproto session not available for DeleteRecord")
}

// SetAtprotoSession sets the internal session (used during OAuth flow)
func (sw *SessionWrapper) SetAtprotoSession(sess session.Session) {
	sw.atprotoSession = sess
}

// GetAtprotoSession returns the internal session (if available)
func (sw *SessionWrapper) GetAtprotoSession() session.Session {
	return sw.atprotoSession
}

// SaveToCookies saves session data to HTTP cookies
func (sw *SessionWrapper) SaveToCookies(w http.ResponseWriter, isDev bool) error {
	// Use existing cookie management functions
	SetSessionCookieWithEnv(w, sw.GetAccessToken(), []string{sw.GetRefreshToken()}, isDev)
	
	// Save DPoP key if available
	if sw.GetDPoPKey() != nil {
		if err := oauth.SetDPoPKeyCookie(w, sw.GetDPoPKey(), isDev); err != nil {
			return fmt.Errorf("failed to save DPoP key to cookie: %w", err)
		}
	}
	
	return nil
}

// LoadSessionFromCookies creates session wrapper from HTTP cookies  
func LoadSessionFromCookies(r *http.Request) (*SessionWrapper, error) {
	// Get access token from cookie
	accessToken, err := GetSessionCookie(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token from cookie: %w", err)
	}
	
	// Get refresh token (optional)
	refreshToken, _ := GetRefreshTokenCookie(r)
	
	// Get DPoP key (optional)
	dpopKey, _ := oauth.GetDPoPKeyFromCookie(r)
	
	// Extract user DID from access token (JWT parsing)
	userDID := ""
	// Simple JWT parsing to get subject
	if accessToken != "" {
		parts := strings.Split(accessToken, ".")
		if len(parts) >= 2 {
			// Decode payload (add padding if needed)
			payload := parts[1]
			for len(payload)%4 != 0 {
				payload += "="
			}
			
			if decoded, err := base64.StdEncoding.DecodeString(payload); err == nil {
				var claims map[string]interface{}
				if err := json.Unmarshal(decoded, &claims); err == nil {
					if sub, ok := claims["sub"].(string); ok {
						userDID = sub
					}
				}
			}
		}
	}
	
	return NewSessionWrapper(accessToken, refreshToken, userDID, dpopKey, nil)
}