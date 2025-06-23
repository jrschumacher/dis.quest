// Package oauth provides OAuth2 authentication for ATProtocol
package oauth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// DefaultProvider implements OAuth using the proven ATProtocol authentication pattern.
// This consolidates the working approach from the previous tangled implementation.
type DefaultProvider struct {
	config   *ProviderConfig
	metadata *AuthorizationServerMetadata
}

// NewProvider creates a new OAuth provider
func NewProvider(config *ProviderConfig) Provider {
	return &DefaultProvider{
		config: config,
		// metadata will be loaded on first use
	}
}

// GetProviderName returns a descriptive name for this provider
func (p *DefaultProvider) GetProviderName() string {
	return "ATProtocol OAuth Provider"
}

// GetAuthURL generates the OAuth authorization URL with PKCE
func (p *DefaultProvider) GetAuthURL(state, codeChallenge string) string {
	// Load metadata if not already loaded
	if p.metadata == nil {
		var err error
		p.metadata, err = DiscoverAuthorizationServer(p.config.PDSEndpoint)
		if err != nil {
			// Return empty string on error - caller should handle
			return ""
		}
	}

	// Build authorization URL using discovered metadata
	return fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s&code_challenge=%s&code_challenge_method=S256",
		p.metadata.AuthorizationEndpoint, p.config.ClientID, p.config.RedirectURI, p.config.Scope, state, codeChallenge)
}

// ExchangeToken exchanges authorization code for access token using the proven DPoP pattern
func (p *DefaultProvider) ExchangeToken(ctx context.Context, code, codeVerifier string) (*TokenResult, error) {
	// Load metadata if not already loaded
	if p.metadata == nil {
		var err error
		p.metadata, err = DiscoverAuthorizationServer(p.config.PDSEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to discover authorization server metadata: %w", err)
		}
	}

	// Extract HTTP request from context to access cookies
	req, ok := ctx.Value("http_request").(*http.Request)
	if !ok {
		return nil, fmt.Errorf("http request not found in context")
	}

	// Get DPoP key from cookies (stored during auth flow)
	dpopKey, err := GetDPoPKeyFromCookie(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get DPoP key: %w", err)
	}

	// Create DPoP key pair wrapper
	dpopKeyPair := &DPoPKeyPair{PrivateKey: dpopKey}

	// Exchange code for token with DPoP using proven implementation
	token, err := ExchangeCodeForTokenWithDPoP(ctx, p.metadata, code, codeVerifier, *dpopKeyPair, p.config)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

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
		DPoPKey:      dpopKeyPair.PrivateKey,
		UserDID:      userDID,
		ExpiresIn:    int64(token.Expiry.Unix()),
	}, nil
}

// RefreshToken refreshes an expired access token
func (p *DefaultProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error) {
	// Load metadata if not already loaded
	if p.metadata == nil {
		var err error
		p.metadata, err = DiscoverAuthorizationServer(p.config.PDSEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to discover authorization server metadata: %w", err)
		}
	}

	// Create OAuth2 config for token refresh
	conf := &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: "", // ATProtocol uses client credentials differently
		Endpoint: oauth2.Endpoint{
			AuthURL:  p.metadata.AuthorizationEndpoint,
			TokenURL: p.metadata.TokenEndpoint,
		},
		RedirectURL: p.config.RedirectURI,
		Scopes:      []string{p.config.Scope},
	}

	// Create OAuth2 token source for refresh
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}
	
	tokenSource := conf.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// For token refresh, we should preserve the existing DPoP key from the session
	// The key will be provided by the session management system when calling refresh
	// For now, we'll return a nil DPoP key and let the session manager handle it
	var dpopKey *ecdsa.PrivateKey = nil

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

// CreateAuthorizedClient creates an XRPC client with the given token
func (p *DefaultProvider) CreateAuthorizedClient(token *TokenResult) (XRPCClient, error) {
	// Integration with existing XRPC client
	return &defaultXRPCClient{
		accessToken: token.AccessToken,
		dpopKey:     token.DPoPKey,
		pdsEndpoint: p.config.PDSEndpoint,
	}, nil
}

// defaultXRPCClient implements XRPCClient interface
type defaultXRPCClient struct {
	accessToken string
	dpopKey     *ecdsa.PrivateKey
	pdsEndpoint string
}

// Placeholder XRPC client methods - will be properly integrated with pkg/atproto/xrpc
func (c *defaultXRPCClient) CreateRecord(ctx context.Context, repo, collection, rkey string, record any) error {
	return fmt.Errorf("XRPC operations will be handled by session interface")
}

func (c *defaultXRPCClient) GetRecord(ctx context.Context, repo, collection, rkey string, result any) error {
	return fmt.Errorf("XRPC operations will be handled by session interface")  
}

func (c *defaultXRPCClient) PutRecord(ctx context.Context, repo, collection, rkey string, record any) error {
	return fmt.Errorf("XRPC operations will be handled by session interface")
}

func (c *defaultXRPCClient) DeleteRecord(ctx context.Context, repo, collection, rkey string) error {
	return fmt.Errorf("XRPC operations will be handled by session interface")
}