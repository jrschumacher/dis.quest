// Package oauth provides OAuth provider abstractions for ATProtocol authentication
package oauth

import (
	"context"
	"crypto/ecdsa"
)

// TokenResult represents the result of a successful OAuth token exchange
type TokenResult struct {
	AccessToken  string
	RefreshToken string
	DPoPKey      *ecdsa.PrivateKey
	UserDID      string
	ExpiresIn    int64
}

// XRPCClient represents an authenticated XRPC client for ATProtocol operations
type XRPCClient interface {
	CreateRecord(ctx context.Context, repo, collection, rkey string, record any) error
	GetRecord(ctx context.Context, repo, collection, rkey string, result any) error
	PutRecord(ctx context.Context, repo, collection, rkey string, record any) error
	DeleteRecord(ctx context.Context, repo, collection, rkey string) error
}

// Provider defines the interface for ATProtocol OAuth implementations
type Provider interface {
	// GetAuthURL generates the OAuth authorization URL with PKCE
	GetAuthURL(state, codeChallenge string) string
	
	// ExchangeToken exchanges authorization code for access token with DPoP binding
	// The context should contain the HTTP request for accessing cookies/session data
	ExchangeToken(ctx context.Context, code, codeVerifier string) (*TokenResult, error)
	
	// CreateAuthorizedClient creates an XRPC client with the given token
	CreateAuthorizedClient(token *TokenResult) (XRPCClient, error)
	
	// RefreshToken refreshes an expired access token
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error)
	
	// GetProviderName returns the name of this provider implementation
	GetProviderName() string
}

// Config holds common configuration for OAuth providers
type Config struct {
	ClientID        string
	ClientURI       string
	RedirectURI     string
	PDSEndpoint     string
	JWKSPrivateKey  string
	JWKSPublicKey   string
	Scope           string
}