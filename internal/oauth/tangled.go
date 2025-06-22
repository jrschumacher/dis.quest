package oauth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/config"
	"golang.org/x/oauth2"
)

// TangledOAuthProvider implements OAuth using the native auth functions
// This provider replaces the tangled-sh dependency while maintaining the same interface
// and focusing on token exchange to fix the "Bad token scope" error
type TangledOAuthProvider struct {
	config   *Config
	metadata *auth.AuthorizationServerMetadata
}

// NewTangledOAuthProvider creates a new tangled OAuth provider
func NewTangledOAuthProvider(config *Config) *TangledOAuthProvider {
	return &TangledOAuthProvider{
		config: config,
		// metadata will be loaded on first use
	}
}

// GetAuthURL generates the OAuth authorization URL with PKCE
func (t *TangledOAuthProvider) GetAuthURL(state, codeChallenge string) string {
	// Load metadata if not already loaded
	if t.metadata == nil {
		var err error
		t.metadata, err = auth.DiscoverAuthorizationServer(t.config.PDSEndpoint)
		if err != nil {
			log.Printf("[TangledOAuthProvider] Failed to discover authorization server: %v", err)
			// Return empty string on error - caller should handle
			return ""
		}
	}

	// Build authorization URL using discovered metadata
	return fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s&code_challenge=%s&code_challenge_method=S256",
		t.metadata.AuthorizationEndpoint, t.config.ClientID, t.config.RedirectURI, t.config.Scope, state, codeChallenge)
}

// ExchangeToken exchanges authorization code for access token using native auth functions
func (t *TangledOAuthProvider) ExchangeToken(ctx context.Context, code, codeVerifier string) (*TokenResult, error) {
	log.Printf("[TangledOAuthProvider] Starting token exchange with code: %s", code[:8]+"...")

	// Load metadata if not already loaded
	if t.metadata == nil {
		var err error
		t.metadata, err = auth.DiscoverAuthorizationServer(t.config.PDSEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to discover authorization server metadata: %w", err)
		}
	}

	// Extract HTTP request from context to access cookies
	req, ok := ctx.Value("http_request").(*http.Request)
	if !ok {
		return nil, fmt.Errorf("http request not found in context")
	}

	// Get DPoP key from cookies (as ECDSA private key)
	dpopKey, err := auth.GetDPoPKeyFromCookie(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get DPoP key: %w", err)
	}

	// Convert to legacy config format for compatibility
	legacyConfig := &config.Config{
		OAuthClientID:    t.config.ClientID,
		OAuthRedirectURL: t.config.RedirectURI,
		PDSEndpoint:      t.config.PDSEndpoint,
		JWKSPrivate:      t.config.JWKSPrivateKey,
		JWKSPublic:       t.config.JWKSPublicKey,
		PublicDomain:     t.config.ClientURI,
	}

	// Exchange code for token with DPoP using native implementation
	token, err := auth.ExchangeCodeForTokenWithDPoP(ctx, t.metadata, code, codeVerifier, dpopKey, legacyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	log.Printf("[TangledOAuthProvider] Token exchange successful, got access token")

	// Extract user DID from token (if available in extra fields)
	userDID := ""
	if extra := token.Extra("sub"); extra != nil {
		if did, ok := extra.(string); ok {
			userDID = did
		}
	}

	return &TokenResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		DPoPKey:      dpopKey,
		UserDID:      userDID,
		ExpiresIn:    int64(token.Expiry.Unix()),
	}, nil
}

// CreateAuthorizedClient creates an XRPC client with the given token
func (t *TangledOAuthProvider) CreateAuthorizedClient(token *TokenResult) (XRPCClient, error) {
	// Integration with existing XRPC client - for now return a wrapper
	return &tangledXRPCClient{
		accessToken: token.AccessToken,
		dpopKey:     token.DPoPKey,
		pdsEndpoint: t.config.PDSEndpoint,
	}, nil
}

// RefreshToken refreshes an expired access token
func (t *TangledOAuthProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error) {
	// Load metadata if not already loaded
	if t.metadata == nil {
		var err error
		t.metadata, err = auth.DiscoverAuthorizationServer(t.config.PDSEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to discover authorization server metadata: %w", err)
		}
	}

	// Get HTTP request from context for session access
	req, ok := ctx.Value("http_request").(*http.Request)
	if !ok {
		return nil, fmt.Errorf("HTTP request not found in context")
	}
	
	log.Printf("[TangledOAuthProvider] Starting token refresh")
	
	// Get DPoP key from cookies (as ECDSA private key)
	dpopKey, err := auth.GetDPoPKeyFromCookie(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get DPoP key: %w", err)
	}

	// Convert to legacy config format for compatibility
	legacyConfig := &config.Config{
		OAuthClientID:    t.config.ClientID,
		OAuthRedirectURL: t.config.RedirectURI,
		PDSEndpoint:      t.config.PDSEndpoint,
		JWKSPrivate:      t.config.JWKSPrivateKey,
		JWKSPublic:       t.config.JWKSPublicKey,
		PublicDomain:     t.config.ClientURI,
	}

	// Create OAuth2 token for refresh
	conf := auth.OAuth2Config(t.metadata, legacyConfig)
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}
	
	// Use DPoP transport for refresh
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: &auth.DPoPPKCETransport{
			Base:         http.DefaultTransport,
			CodeVerifier: "", // Not needed for refresh
			DPoPKey:      dpopKey,
			TargetURL:    t.metadata.TokenEndpoint,
			Config:       legacyConfig,
			Metadata:     t.metadata,
		},
	})
	
	tokenSource := conf.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	
	log.Printf("[TangledOAuthProvider] Token refresh successful")
	
	// Extract user DID from token (if available in extra fields)
	userDID := ""
	if extra := newToken.Extra("sub"); extra != nil {
		if did, ok := extra.(string); ok {
			userDID = did
		}
	}

	return &TokenResult{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		DPoPKey:      dpopKey,
		UserDID:      userDID,
		ExpiresIn:    int64(newToken.Expiry.Unix()),
	}, nil
}

// GetProviderName returns the name of this provider implementation
func (t *TangledOAuthProvider) GetProviderName() string {
	return "tangled"
}

// tangledXRPCClient is a placeholder XRPC client implementation that uses existing XRPC code
type tangledXRPCClient struct {
	accessToken string
	dpopKey     *ecdsa.PrivateKey
	pdsEndpoint string
}

func (c *tangledXRPCClient) CreateRecord(ctx context.Context, repo, collection, rkey string, record any) error {
	// TODO: Integrate with existing XRPC implementation
	return fmt.Errorf("tangled XRPC client integration pending")
}

func (c *tangledXRPCClient) GetRecord(ctx context.Context, repo, collection, rkey string, result any) error {
	// TODO: Integrate with existing XRPC implementation
	return fmt.Errorf("tangled XRPC client integration pending")
}

func (c *tangledXRPCClient) PutRecord(ctx context.Context, repo, collection, rkey string, record any) error {
	// TODO: Integrate with existing XRPC implementation
	return fmt.Errorf("tangled XRPC client integration pending")
}

func (c *tangledXRPCClient) DeleteRecord(ctx context.Context, repo, collection, rkey string) error {
	// TODO: Integrate with existing XRPC implementation
	return fmt.Errorf("tangled XRPC client integration pending")
}