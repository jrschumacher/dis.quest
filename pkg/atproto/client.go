// Package atproto provides a production-ready ATProtocol client library for Go developers.
// Built on proven components from dis.quest with complete OAuth+DPoP authentication
// and custom lexicon support.
package atproto

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
	"github.com/jrschumacher/dis.quest/pkg/atproto/xrpc"
)

// Client represents the main ATProtocol client
type Client struct {
	config *Config
	oauth  *oauth.OAuthClient
}

// Config contains the configuration for the ATProtocol client
type Config struct {
	ClientID       string
	ClientURI      string
	RedirectURI    string
	JWKSPrivateKey string
	JWKSPublicKey  string
	Scope          string
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

	// Set default scope if not provided
	if config.Scope == "" {
		config.Scope = "atproto transition:generic"
	}

	oauthConfig := &oauth.Config{
		ClientID:       config.ClientID,
		ClientURI:      config.ClientURI,
		RedirectURI:    config.RedirectURI,
		JWKSPrivateKey: config.JWKSPrivateKey,
		JWKSPublicKey:  config.JWKSPublicKey,
		Scope:          config.Scope,
	}

	oauthClient, err := oauth.NewOAuthClient(oauthConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth client: %w", err)
	}

	return &Client{
		config: &config,
		oauth:  oauthClient,
	}, nil
}

// GetAuthURL generates the OAuth authorization URL for user authentication
func (c *Client) GetAuthURL(state, codeChallenge string) string {
	return c.oauth.GetAuthURL(state, codeChallenge)
}

// ExchangeCode exchanges the authorization code for an authenticated session
// This is a simplified interface - in production, the DPoP key, nonce, and auth server
// would be managed through the OAuth flow
func (c *Client) ExchangeCode(ctx context.Context, code, codeVerifier string, dpopKey *ecdsa.PrivateKey, dpopNonce, authServerIssuer string) (*Session, error) {
	tokenResult, err := c.oauth.ExchangeCode(ctx, code, codeVerifier, dpopKey, dpopNonce, authServerIssuer)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return &Session{
		client:       c,
		accessToken:  tokenResult.AccessToken,
		refreshToken: tokenResult.RefreshToken,
		userDID:      tokenResult.UserDID,
		dpopKey:      tokenResult.DPoPKey,
		expiresIn:    tokenResult.ExpiresIn,
	}, nil
}

// Session represents an authenticated session with a Personal Data Server
type Session struct {
	client       *Client
	accessToken  string
	refreshToken string
	userDID      string
	dpopKey      *ecdsa.PrivateKey
	expiresIn    int64
}

// GetUserDID returns the authenticated user's DID
func (s *Session) GetUserDID() string {
	return s.userDID
}

// CreateRecord creates a new record in the user's PDS
func (s *Session) CreateRecord(collection, rkey string, record interface{}) (*RecordResult, error) {
	client := s.client.NewXRPCClient()
	resp, err := client.CreateRecord(context.Background(), s.userDID, collection, rkey, record, s.accessToken, s.dpopKey)
	if err != nil {
		return nil, err
	}
	return &RecordResult{URI: resp.URI, CID: resp.CID}, nil
}

// GetRecord retrieves a record from the user's PDS
func (s *Session) GetRecord(collection, rkey string, result interface{}) error {
	client := s.client.NewXRPCClient()
	return client.GetRecord(context.Background(), s.userDID, collection, rkey, result, s.accessToken, s.dpopKey)
}

// ListRecords lists records from a collection in the user's PDS
func (s *Session) ListRecords(collection string, limit int, cursor string) (*ListRecordsResult, error) {
	client := s.client.NewXRPCClient()
	resp, err := client.ListRecords(context.Background(), s.userDID, collection, limit, cursor, s.accessToken, s.dpopKey)
	if err != nil {
		return nil, err
	}
	
	records := make([]Record, len(resp.Records))
	for i, r := range resp.Records {
		records[i] = Record{URI: r.URI, CID: r.CID, Value: r.Value}
	}
	
	return &ListRecordsResult{Records: records, Cursor: resp.Cursor}, nil
}

// UpdateRecord updates an existing record in the user's PDS
func (s *Session) UpdateRecord(collection, rkey string, record interface{}) (*RecordResult, error) {
	client := s.client.NewXRPCClient()
	resp, err := client.UpdateRecord(context.Background(), s.userDID, collection, rkey, record, s.accessToken, s.dpopKey)
	if err != nil {
		return nil, err
	}
	return &RecordResult{URI: resp.URI, CID: resp.CID}, nil
}

// DeleteRecord deletes a record from the user's PDS
func (s *Session) DeleteRecord(collection, rkey string) error {
	client := s.client.NewXRPCClient()
	return client.DeleteRecord(context.Background(), s.userDID, collection, rkey, s.accessToken, s.dpopKey)
}

// NewXRPCClient creates a new XRPC client for lower-level operations
func (c *Client) NewXRPCClient() *xrpc.Client {
	return xrpc.NewClient()
}

// RecordResult represents the result of a record operation
type RecordResult struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

// ListRecordsResult represents the result of a list records operation
type ListRecordsResult struct {
	Records []Record `json:"records"`
	Cursor  string   `json:"cursor,omitempty"`
}

// Record represents an ATProtocol record
type Record struct {
	URI   string      `json:"uri"`
	CID   string      `json:"cid"`
	Value interface{} `json:"value"`
}