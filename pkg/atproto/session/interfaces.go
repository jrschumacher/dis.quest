// Package session provides universal session management for ATProtocol applications.
// It separates ATProtocol session semantics from storage/transport concerns,
// enabling reuse across web apps, CLI tools, mobile apps, and production services.
package session

import (
	"context"
	"crypto/ecdsa"
	"time"
)

// Storage defines the interface for session persistence backends.
// Implementations can use cookies, files, databases, Redis, etc.
type Storage interface {
	// Store saves session data with the given key
	Store(ctx context.Context, key string, data *Data) error
	
	// Load retrieves session data by key
	Load(ctx context.Context, key string) (*Data, error)
	
	// Delete removes session data
	Delete(ctx context.Context, key string) error
	
	// Cleanup removes expired sessions (called periodically)
	Cleanup(ctx context.Context) error
	
	// Close cleans up storage resources
	Close() error
}

// Manager handles ATProtocol session lifecycle using pluggable storage.
type Manager interface {
	// CreateSession creates a new session from OAuth token result
	CreateSession(ctx context.Context, tokenResult *TokenResult) (Session, error)
	
	// LoadSession retrieves an existing session by ID
	LoadSession(ctx context.Context, sessionID string) (Session, error)
	
	// SaveSession persists session to storage
	SaveSession(ctx context.Context, session Session) error
	
	// RefreshSession refreshes an expired session using refresh token
	RefreshSession(ctx context.Context, session Session) error
	
	// DeleteSession removes session from storage
	DeleteSession(ctx context.Context, sessionID string) error
	
	// ValidateToken parses and validates JWT tokens
	ValidateToken(token string) (*TokenClaims, error)
	
	// IsTokenExpired checks if token expires within threshold
	IsTokenExpired(token string, threshold time.Duration) bool
	
	// GenerateSessionID creates a unique session identifier
	GenerateSessionID() string
	
	// Close cleans up manager resources
	Close() error
}

// Data represents session information stored in the backend.
type Data struct {
	// Session identification
	SessionID string    `json:"session_id"`
	UserDID   string    `json:"user_did"`
	Handle    string    `json:"handle,omitempty"`
	
	// OAuth tokens
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type,omitempty"`
	
	// DPoP key for token binding (encrypted in storage)
	DPoPKey *ecdsa.PrivateKey `json:"-"` // Never serialize directly
	DPoPKeyEncrypted []byte   `json:"dpop_key_encrypted,omitempty"`
	
	// Timing information
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Optional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TokenResult represents the result of OAuth token exchange.
type TokenResult struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
	UserDID      string
	Handle       string
	DPoPKey      *ecdsa.PrivateKey
}

// TokenClaims represents parsed JWT token claims.
type TokenClaims struct {
	Subject   string                 `json:"sub"`
	Issuer    string                 `json:"iss"`
	Audience  string                 `json:"aud"`
	ExpiresAt int64                  `json:"exp"`
	IssuedAt  int64                  `json:"iat"`
	NotBefore int64                  `json:"nbf,omitempty"`
	Custom    map[string]interface{} `json:"-"` // Additional claims
}

// Config contains session manager configuration.
type Config struct {
	// Token expiration checking
	TokenExpiryThreshold time.Duration
	
	// Session ID generation
	SessionIDGenerator func() string
	
	// Encryption key for sensitive data (DPoP keys)
	EncryptionKey []byte
	
	// Cleanup settings
	CleanupInterval time.Duration
	MaxSessionAge   time.Duration
	
	// Custom validation
	TokenValidator func(token string) (*TokenClaims, error)
}

// Session represents an active ATProtocol session.
type Session interface {
	// Identity
	GetSessionID() string
	GetUserDID() string
	GetHandle() string
	
	// Tokens
	GetAccessToken() string
	GetRefreshToken() string
	GetDPoPKey() *ecdsa.PrivateKey
	
	// State management
	IsExpired() bool
	Refresh(ctx context.Context) error
	UpdateTokens(accessToken, refreshToken string, expiresIn int64) error
	
	// Persistence
	Save(ctx context.Context) error
	Delete(ctx context.Context) error
	
	// ATProtocol operations (delegates to XRPC client)
	CreateRecord(ctx context.Context, collection, rkey string, record interface{}) (*RecordResult, error)
	GetRecord(ctx context.Context, collection, rkey string, result interface{}) error
	ListRecords(ctx context.Context, collection string, limit int, cursor string) (*ListRecordsResult, error)
	UpdateRecord(ctx context.Context, collection, rkey string, record interface{}) (*RecordResult, error)
	DeleteRecord(ctx context.Context, collection, rkey string) error
	
	// Session data access
	GetData() *Data
	GetMetadata(key string) interface{}
	SetMetadata(key string, value interface{})
}

// RecordResult represents the result of record operations.
type RecordResult struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

// ListRecordsResult represents the result of listing records.
type ListRecordsResult struct {
	Records []Record `json:"records"`
	Cursor  string   `json:"cursor,omitempty"`
}

// Record represents a single ATProtocol record.
type Record struct {
	URI   string      `json:"uri"`
	CID   string      `json:"cid"`
	Value interface{} `json:"value"`
}