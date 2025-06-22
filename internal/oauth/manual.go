package oauth

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/config"
	"golang.org/x/oauth2"
)

// ManualProvider implements Provider using the current manual DPoP implementation
type ManualProvider struct {
	config   *Config
	metadata *auth.AuthorizationServerMetadata
}

// NewManualProvider creates a new ManualProvider
func NewManualProvider(config *Config) *ManualProvider {
	return &ManualProvider{
		config: config,
	}
}

// GetProviderName returns the name of this provider
func (m *ManualProvider) GetProviderName() string {
	return "manual"
}

// GetAuthURL generates the OAuth authorization URL with PKCE
func (m *ManualProvider) GetAuthURL(state, codeChallenge string) string {
	// Load metadata if not already loaded
	if m.metadata == nil {
		var err error
		m.metadata, err = auth.DiscoverAuthorizationServer(m.config.PDSEndpoint)
		if err != nil {
			// Return empty string on error - caller should handle
			return ""
		}
	}

	conf := m.oauth2Config()
	authURL := conf.AuthCodeURL(state, 
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	
	return authURL
}

// ExchangeToken exchanges authorization code for access token with DPoP binding
func (m *ManualProvider) ExchangeToken(ctx context.Context, code, codeVerifier string) (*TokenResult, error) {
	// Load metadata if not already loaded
	if m.metadata == nil {
		var err error
		m.metadata, err = auth.DiscoverAuthorizationServer(m.config.PDSEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to discover authorization server metadata: %w", err)
		}
	}

	// Generate DPoP key
	dpopKeyPair, err := auth.GenerateDPoPKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate DPoP key: %w", err)
	}
	dpopKey := dpopKeyPair.PrivateKey

	// Convert to legacy config format for compatibility
	legacyConfig := &config.Config{
		OAuthClientID:    m.config.ClientID,
		OAuthRedirectURL: m.config.RedirectURI,
		PDSEndpoint:      m.config.PDSEndpoint,
		JWKSPrivate:      m.config.JWKSPrivateKey,
		JWKSPublic:       m.config.JWKSPublicKey,
		PublicDomain:     m.config.ClientURI,
	}

	// Exchange code for token with DPoP using existing implementation
	token, err := auth.ExchangeCodeForTokenWithDPoP(ctx, m.metadata, code, codeVerifier, dpopKey, legacyConfig)
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
		DPoPKey:      dpopKey,
		UserDID:      userDID,
		ExpiresIn:    int64(token.Expiry.Unix()),
	}, nil
}

// CreateAuthorizedClient creates an XRPC client with the given token
func (m *ManualProvider) CreateAuthorizedClient(token *TokenResult) (XRPCClient, error) {
	// Integration with existing XRPC client - for now return a wrapper
	return &manualXRPCClient{
		accessToken: token.AccessToken,
		dpopKey:     token.DPoPKey,
		pdsEndpoint: m.config.PDSEndpoint,
	}, nil
}

// RefreshToken refreshes an expired access token
func (m *ManualProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error) {
	// Load metadata if not already loaded
	if m.metadata == nil {
		var err error
		m.metadata, err = auth.DiscoverAuthorizationServer(m.config.PDSEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to discover authorization server metadata: %w", err)
		}
	}

	// Create OAuth2 token source for refresh
	conf := m.oauth2Config()
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}
	
	tokenSource := conf.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Generate new DPoP key for the refreshed token
	dpopKeyPair, err := auth.GenerateDPoPKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate DPoP key: %w", err)
	}
	dpopKey := dpopKeyPair.PrivateKey

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

// oauth2Config creates an OAuth2 configuration
func (m *ManualProvider) oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     m.config.ClientID,
		ClientSecret: "", // Not required for public clients
		RedirectURL:  m.config.RedirectURI,
		Scopes:       []string{"atproto", "transition:generic", "transition:chat.bsky"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  m.metadata.AuthorizationEndpoint,
			TokenURL: m.metadata.TokenEndpoint,
		},
	}
}

// manualXRPCClient is a placeholder XRPC client implementation that uses existing XRPC code
type manualXRPCClient struct {
	accessToken string
	dpopKey     *ecdsa.PrivateKey
	pdsEndpoint string
}

func (c *manualXRPCClient) CreateRecord(ctx context.Context, repo, collection, rkey string, record any) error {
	// TODO: Integrate with existing XRPC implementation
	return fmt.Errorf("manual XRPC client integration pending")
}

func (c *manualXRPCClient) GetRecord(ctx context.Context, repo, collection, rkey string, result any) error {
	// TODO: Integrate with existing XRPC implementation
	return fmt.Errorf("manual XRPC client integration pending")
}

func (c *manualXRPCClient) PutRecord(ctx context.Context, repo, collection, rkey string, record any) error {
	// TODO: Integrate with existing XRPC implementation
	return fmt.Errorf("manual XRPC client integration pending")
}

func (c *manualXRPCClient) DeleteRecord(ctx context.Context, repo, collection, rkey string) error {
	// TODO: Integrate with existing XRPC implementation
	return fmt.Errorf("manual XRPC client integration pending")
}