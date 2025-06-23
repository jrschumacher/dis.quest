// Package atproto provides a production-ready ATProtocol client library for Go developers.
// Built on proven components from dis.quest with complete OAuth+DPoP authentication
// and custom lexicon support.
package atproto

import (
	"context"
	"fmt"
	"time"

	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
	"github.com/jrschumacher/dis.quest/pkg/atproto/session"
	"github.com/jrschumacher/dis.quest/pkg/atproto/xrpc"
)


// Client represents the main ATProtocol client
type Client struct {
	config         *Config
	provider       oauth.Provider
	sessionManager session.Manager
	xrpcClient     *xrpc.Client
}

// Config contains the configuration for the ATProtocol client
type Config struct {
	// OAuth configuration
	ClientID       string
	ClientURI      string
	RedirectURI    string
	PDSEndpoint    string
	JWKSPrivateKey string
	JWKSPublicKey  string
	Scope          string
	
	// Session management configuration (optional)
	SessionStorage session.Storage
	SessionConfig  session.Config
}

// New creates a new ATProtocol client with the given configuration
func New(config Config) (*Client, error) {
	if config.ClientID == "" {
		return nil, fmt.Errorf("ClientID is required")
	}
	if config.RedirectURI == "" {
		return nil, fmt.Errorf("RedirectURI is required")
	}
	if config.JWKSPrivateKey == "" {
		return nil, fmt.Errorf("JWKSPrivateKey is required")
	}
	if config.PDSEndpoint == "" {
		return nil, fmt.Errorf("PDSEndpoint is required")
	}

	// Set default scope if not provided
	if config.Scope == "" {
		config.Scope = "atproto transition:generic"
	}

	providerConfig := &oauth.ProviderConfig{
		ClientID:       config.ClientID,
		ClientURI:      config.ClientURI,
		RedirectURI:    config.RedirectURI,
		PDSEndpoint:    config.PDSEndpoint,
		JWKSPrivateKey: config.JWKSPrivateKey,
		JWKSPublicKey:  config.JWKSPublicKey,
		Scope:          config.Scope,
	}

	// Create OAuth provider
	provider := oauth.NewProvider(providerConfig)

	// Set up session storage defaults
	if config.SessionStorage == nil {
		// Default to memory storage for development
		config.SessionStorage = session.NewMemoryStorage()
	}
	
	// Set up session config defaults
	if config.SessionConfig.TokenExpiryThreshold == 0 {
		config.SessionConfig.TokenExpiryThreshold = 5 * time.Minute
	}
	if config.SessionConfig.CleanupInterval == 0 {
		config.SessionConfig.CleanupInterval = 1 * time.Hour
	}
	if config.SessionConfig.MaxSessionAge == 0 {
		config.SessionConfig.MaxSessionAge = 24 * time.Hour
	}
	
	// Create XRPC client
	xrpcClient := xrpc.NewClient()
	
	// Create session manager
	sessionManager := session.NewManager(config.SessionStorage, config.SessionConfig, provider, xrpcClient)

	return &Client{
		config:         &config,
		provider:       provider,
		sessionManager: sessionManager,
		xrpcClient:     xrpcClient,
	}, nil
}

// GetAuthURL generates the OAuth authorization URL for user authentication
func (c *Client) GetAuthURL(state, codeChallenge string) string {
	return c.provider.GetAuthURL(state, codeChallenge)
}

// LoadSession retrieves an existing session by ID
func (c *Client) LoadSession(ctx context.Context, sessionID string) (session.Session, error) {
	return c.sessionManager.LoadSession(ctx, sessionID)
}

// DeleteSession removes a session by ID
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error {
	return c.sessionManager.DeleteSession(ctx, sessionID)
}

// GetSessionManager returns the session manager for advanced usage
func (c *Client) GetSessionManager() session.Manager {
	return c.sessionManager
}

// Close cleans up client resources
func (c *Client) Close() error {
	return c.sessionManager.Close()
}

// ExchangeCode exchanges the authorization code for an authenticated session
func (c *Client) ExchangeCode(ctx context.Context, code, codeVerifier string) (session.Session, error) {
	tokenResult, err := c.provider.ExchangeToken(ctx, code, codeVerifier)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Convert OAuth TokenResult to session TokenResult
	sessionTokenResult := &session.TokenResult{
		AccessToken:  tokenResult.AccessToken,
		RefreshToken: tokenResult.RefreshToken,
		TokenType:    "DPoP", // ATProtocol uses DPoP tokens
		ExpiresIn:    tokenResult.ExpiresIn,
		UserDID:      tokenResult.UserDID,
		DPoPKey:      tokenResult.DPoPKey,
	}

	// Create session using session manager
	return c.sessionManager.CreateSession(ctx, sessionTokenResult)
}

// Legacy session type aliases for backward compatibility
// These delegate to the new session system but maintain the old interface
type Session = session.Session
type RecordResult = session.RecordResult
type ListRecordsResult = session.ListRecordsResult
type Record = session.Record

// NewXRPCClient creates a new XRPC client for lower-level operations
func (c *Client) NewXRPCClient() *xrpc.Client {
	return c.xrpcClient
}