// Package oauth provides OAuth2 providers for ATProtocol authentication
package oauth

import (
	"context"
)

// Provider defines the interface for OAuth providers
type Provider interface {
	GetProviderName() string
	GetAuthURL(state, codeChallenge string) string
	ExchangeToken(ctx context.Context, code, codeVerifier string) (*TokenResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error)
	CreateAuthorizedClient(token *TokenResult) (XRPCClient, error)
}

// XRPCClient defines the interface for XRPC operations
type XRPCClient interface {
	CreateRecord(ctx context.Context, repo, collection, rkey string, record any) error
	GetRecord(ctx context.Context, repo, collection, rkey string, result any) error
	PutRecord(ctx context.Context, repo, collection, rkey string, record any) error
	DeleteRecord(ctx context.Context, repo, collection, rkey string) error
}

// ProviderConfig contains OAuth configuration for providers
type ProviderConfig struct {
	ClientID       string
	ClientURI      string
	RedirectURI    string
	PDSEndpoint    string
	JWKSPrivateKey string
	JWKSPublicKey  string
	Scope          string
}

// Legacy provider constructors for backward compatibility
func NewManualProvider(config *ProviderConfig) Provider {
	return NewProvider(config)
}

func NewTangledProvider(config *ProviderConfig) Provider {
	return NewProvider(config)
}