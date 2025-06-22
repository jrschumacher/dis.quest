// Package web provides HTTP-specific session management
package web

import (
	"context"
	"crypto/ecdsa"
	"net/http"

	"github.com/jrschumacher/dis.quest/pkg/atproto/jwt"
	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
	"github.com/jrschumacher/dis.quest/pkg/atproto/session"
)

// SessionBridge converts between HTTP cookies and ATProtocol sessions
type SessionBridge struct {
	manager session.Manager
}

// NewSessionBridge creates a new session bridge for HTTP requests
func NewSessionBridge(manager session.Manager) *SessionBridge {
	return &SessionBridge{
		manager: manager,
	}
}

// CreateSessionFromTokens creates a new ATProtocol session and saves to cookies
func (sb *SessionBridge) CreateSessionFromTokens(
	ctx context.Context,
	w http.ResponseWriter,
	accessToken, refreshToken, userDID, handle string,
	dpopKey *ecdsa.PrivateKey,
	isDev bool,
) (session.Session, error) {
	// Create token result
	tokenResult := &session.TokenResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserDID:      userDID,
		Handle:       handle,
		DPoPKey:      dpopKey,
	}

	// Create session using the manager
	sess, err := sb.manager.CreateSession(ctx, tokenResult)
	if err != nil {
		return nil, err
	}

	// Save session data to cookies for web compatibility
	if err := sb.SaveSessionToCookies(w, sess, isDev); err != nil {
		return nil, err
	}

	return sess, nil
}

// LoadSessionFromCookies loads an ATProtocol session from HTTP cookies
func (sb *SessionBridge) LoadSessionFromCookies(ctx context.Context, r *http.Request) (session.Session, error) {
	// Get access token from cookie
	accessToken, err := GetSessionCookie(r)
	if err != nil {
		return nil, err
	}

	// Extract user DID for session ID
	userDID, err := jwt.ExtractDID(accessToken)
	if err != nil {
		return nil, err
	}

	// Try to load existing session
	sess, err := sb.manager.LoadSession(ctx, userDID)
	if err != nil {
		// If session doesn't exist, create a minimal one from cookies
		refreshToken, _ := GetRefreshTokenCookie(r)
		dpopKey, _ := oauth.GetDPoPKeyFromCookie(r)
		
		tokenResult := &session.TokenResult{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			UserDID:      userDID,
			DPoPKey:      dpopKey,
		}

		sess, err = sb.manager.CreateSession(ctx, tokenResult)
		if err != nil {
			return nil, err
		}
	}

	return sess, nil
}

// SaveSessionToCookies saves session data to HTTP cookies
func (sb *SessionBridge) SaveSessionToCookies(w http.ResponseWriter, sess session.Session, isDev bool) error {
	// Use cookie functions
	SetSessionCookieWithEnv(w, sess.GetAccessToken(), []string{sess.GetRefreshToken()}, isDev)
	
	// Save DPoP key if available
	if sess.GetDPoPKey() != nil {
		if err := oauth.SetDPoPKeyCookie(w, sess.GetDPoPKey(), isDev); err != nil {
			return err
		}
	}
	
	return nil
}

// RefreshSessionFromCookies refreshes an expired session and updates cookies
func (sb *SessionBridge) RefreshSessionFromCookies(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	isDev bool,
) (session.Session, error) {
	// Load current session
	sess, err := sb.LoadSessionFromCookies(ctx, r)
	if err != nil {
		return nil, err
	}

	// Refresh the session
	if err := sess.Refresh(ctx); err != nil {
		return nil, err
	}

	// Update cookies with new tokens
	if err := sb.SaveSessionToCookies(w, sess, isDev); err != nil {
		return nil, err
	}

	return sess, nil
}

// SimpleSessionData represents basic session information for web apps
// This is a lightweight alternative to full ATProtocol sessions when you just need tokens
type SimpleSessionData struct {
	AccessToken  string
	RefreshToken string
	UserDID      string
	DPoPKey      *ecdsa.PrivateKey
}

// LoadSimpleSessionFromCookies loads basic session data from cookies without ATProtocol session manager
func LoadSimpleSessionFromCookies(r *http.Request) (*SimpleSessionData, error) {
	accessToken, err := GetSessionCookie(r)
	if err != nil {
		return nil, err
	}

	refreshToken, _ := GetRefreshTokenCookie(r)
	dpopKey, _ := oauth.GetDPoPKeyFromCookie(r)

	// Extract user DID from access token
	userDID, err := jwt.ExtractDID(accessToken)
	if err != nil {
		return nil, err
	}

	return &SimpleSessionData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserDID:      userDID,
		DPoPKey:      dpopKey,
	}, nil
}