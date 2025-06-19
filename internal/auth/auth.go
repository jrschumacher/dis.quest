// Package auth provides authentication utilities for OAuth2 and session management
package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/jrschumacher/dis.quest/internal/config"
	"golang.org/x/oauth2"
)

// PKCETransport adds PKCE code_verifier to token exchange requests
type PKCETransport struct {
	Base         http.RoundTripper
	CodeVerifier string
}

// RoundTrip implements http.RoundTripper
func (t *PKCETransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only modify token endpoint requests (POST with grant_type=authorization_code)
	if req.Method == "POST" && req.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if req.Body != nil {
			// Read the existing body
			body, err := io.ReadAll(req.Body)
			if err == nil {
				_ = req.Body.Close()
				
				// Parse form values
				values, err := url.ParseQuery(string(body))
				if err == nil && values.Get("grant_type") == "authorization_code" {
					// Add code_verifier
					values.Set("code_verifier", t.CodeVerifier)
					newBody := values.Encode()
					req.Body = io.NopCloser(strings.NewReader(newBody))
					req.ContentLength = int64(len(newBody))
				}
			}
		}
	}
	
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// DPoPPKCETransport adds both PKCE code_verifier and DPoP header to token exchange requests
type DPoPPKCETransport struct {
	Base         http.RoundTripper
	CodeVerifier string
	DPoPKey      *ecdsa.PrivateKey
	TargetURL    string
}

// RoundTrip implements http.RoundTripper for DPoP + PKCE with nonce retry
func (t *DPoPPKCETransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request body since we might need to retry
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		if err := req.Body.Close(); err != nil {
			// Ignore close error, common in HTTP clients
			_ = err
		}
	}
	
	// Helper to create request with DPoP and PKCE
	makeRequest := func(nonce string) (*http.Request, error) {
		// Clone the original request
		newReq := req.Clone(req.Context())
		
		// Create DPoP JWT with optional nonce
		dpopJWT, err := CreateDPoPJWTWithNonce(t.DPoPKey, req.Method, t.TargetURL, nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to create DPoP JWT: %w", err)
		}
		
		// Add DPoP header
		newReq.Header.Set("DPoP", dpopJWT)
		
		// Restore and modify body for PKCE
		if bodyBytes != nil {
			// Parse form values
			values, err := url.ParseQuery(string(bodyBytes))
			if err == nil && values.Get("grant_type") == "authorization_code" {
				// Add code_verifier
				values.Set("code_verifier", t.CodeVerifier)
				newBody := values.Encode()
				newReq.Body = io.NopCloser(strings.NewReader(newBody))
				newReq.ContentLength = int64(len(newBody))
			} else {
				newReq.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
				newReq.ContentLength = int64(len(bodyBytes))
			}
		}
		
		return newReq, nil
	}
	
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	
	// First attempt without nonce
	firstReq, err := makeRequest("")
	if err != nil {
		return nil, err
	}
	
	resp, err := base.RoundTrip(firstReq)
	if err != nil {
		return resp, err
	}
	
	// Check if we got a use_dpop_nonce error
	if resp.StatusCode == 400 {
		// Read response to check for nonce requirement
		respBody, err := io.ReadAll(resp.Body)
		if err := resp.Body.Close(); err != nil {
			// Ignore close error, common in HTTP clients
			_ = err
		}
		if err == nil && strings.Contains(string(respBody), "use_dpop_nonce") {
			// Extract nonce from DPoP-Nonce header
			if nonce := resp.Header.Get("DPoP-Nonce"); nonce != "" {
				// Retry with nonce
				retryReq, err := makeRequest(nonce)
				if err != nil {
					return nil, err
				}
				return base.RoundTrip(retryReq)
			}
		}
		// Restore the response body for the original error
		resp.Body = io.NopCloser(strings.NewReader(string(respBody)))
	}
	
	return resp, err
}

// GeneratePKCE generates a PKCE code verifier and challenge for OAuth2 flows
func GeneratePKCE() (codeVerifier, codeChallenge string, err error) {
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	codeVerifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	codeChallenge = base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return
}

// OAuth2Config creates an OAuth2 configuration for Bluesky/ATProto using authorization server metadata
func OAuth2Config(metadata *AuthorizationServerMetadata, cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.OAuthClientID,
		ClientSecret: "", // Not required for public clients
		RedirectURL:  cfg.OAuthRedirectURL,
		Scopes:       []string{"atproto", "transition:generic"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  metadata.AuthorizationEndpoint,
			TokenURL: metadata.TokenEndpoint,
		},
	}
}

// Session utilities
const (
	sessionCookieName      = "dsq_session"
	refreshTokenCookieName = "dsq_refresh"
)

// SetSessionCookieWithEnv sets session cookies with environment-specific security settings
func SetSessionCookieWithEnv(w http.ResponseWriter, accessToken string, refreshToken []string, isDev bool) {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
	})
	if len(refreshToken) > 0 && refreshToken[0] != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     refreshTokenCookieName,
			Value:    refreshToken[0],
			Path:     "/",
			HttpOnly: true,
			Secure:   secure,
		})
	}
}

// SetSessionCookie sets session cookies with default production security settings
func SetSessionCookie(w http.ResponseWriter, accessToken string, refreshToken ...string) {
	// Default to production (secure) if not using the new function
	SetSessionCookieWithEnv(w, accessToken, refreshToken, false)
}

// ClearSessionCookieWithEnv clears session cookies with environment-specific settings
func ClearSessionCookieWithEnv(w http.ResponseWriter, isDev bool) {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
	})
}

// ClearSessionCookie clears session cookies with default production settings
func ClearSessionCookie(w http.ResponseWriter) {
	// Default to production (secure) if not using the new function
	ClearSessionCookieWithEnv(w, false)
}

// GetSessionCookie retrieves the session cookie value from the request
func GetSessionCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetRefreshTokenCookie retrieves the refresh token cookie value from the request
func GetRefreshTokenCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ExchangeCodeForToken exchanges an authorization code for an access token
func ExchangeCodeForToken(ctx context.Context, metadata *AuthorizationServerMetadata, code, codeVerifier string, cfg *config.Config) (*oauth2.Token, error) {
	conf := OAuth2Config(metadata, cfg)
	// For PKCE, we need to include the code_verifier in the token exchange request
	// The oauth2 library doesn't directly support this, so we need to add it via context
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: &PKCETransport{
			Base: http.DefaultTransport,
			CodeVerifier: codeVerifier,
		},
	})
	return conf.Exchange(ctx, code)
}

// ExchangeCodeForTokenWithDPoP exchanges an authorization code for an access token using DPoP
func ExchangeCodeForTokenWithDPoP(ctx context.Context, metadata *AuthorizationServerMetadata, code, codeVerifier string, dpopKey *ecdsa.PrivateKey, cfg *config.Config) (*oauth2.Token, error) {
	conf := OAuth2Config(metadata, cfg)
	
	// Use a custom transport that adds both PKCE and DPoP with nonce retry
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: &DPoPPKCETransport{
			Base:         http.DefaultTransport,
			CodeVerifier: codeVerifier,
			DPoPKey:      dpopKey,
			TargetURL:    metadata.TokenEndpoint,
		},
	})
	return conf.Exchange(ctx, code)
}
