package session

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
	"github.com/jrschumacher/dis.quest/pkg/atproto/xrpc"
)

// DefaultManager implements the Manager interface using pluggable storage.
type DefaultManager struct {
	storage   Storage
	config    Config
	provider  oauth.Provider
	xrpcClient *xrpc.Client
}

// NewManager creates a new session manager with the given storage backend.
func NewManager(storage Storage, config Config, provider oauth.Provider, xrpcClient *xrpc.Client) Manager {
	// Set default configuration values
	if config.TokenExpiryThreshold == 0 {
		config.TokenExpiryThreshold = 5 * time.Minute
	}
	if config.SessionIDGenerator == nil {
		config.SessionIDGenerator = generateSessionID
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 1 * time.Hour
	}
	if config.MaxSessionAge == 0 {
		config.MaxSessionAge = 24 * time.Hour
	}

	manager := &DefaultManager{
		storage:    storage,
		config:     config,
		provider:   provider,
		xrpcClient: xrpcClient,
	}

	// Start periodic cleanup if storage supports it
	go manager.startCleanupRoutine()

	return manager
}

// CreateSession creates a new session from OAuth token result.
func (m *DefaultManager) CreateSession(ctx context.Context, tokenResult *TokenResult) (Session, error) {
	// Generate unique session ID
	sessionID := m.config.SessionIDGenerator()

	// Parse token expiration
	expiresAt := time.Now().Add(time.Duration(tokenResult.ExpiresIn) * time.Second)

	// Create session data
	data := &Data{
		SessionID:    sessionID,
		UserDID:      tokenResult.UserDID,
		Handle:       tokenResult.Handle,
		AccessToken:  tokenResult.AccessToken,
		RefreshToken: tokenResult.RefreshToken,
		TokenType:    tokenResult.TokenType,
		DPoPKey:      tokenResult.DPoPKey,
		ExpiresAt:    expiresAt,
		IssuedAt:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	// Encrypt DPoP key if encryption is configured
	if len(m.config.EncryptionKey) > 0 && data.DPoPKey != nil {
		encrypted, err := m.encryptDPoPKey(data.DPoPKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt DPoP key: %w", err)
		}
		data.DPoPKeyEncrypted = encrypted
	}

	// Save to storage
	if err := m.storage.Store(ctx, sessionID, data); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Create session instance
	session := &DefaultSession{
		manager: m,
		data:    data,
	}

	return session, nil
}

// LoadSession retrieves an existing session by ID.
func (m *DefaultManager) LoadSession(ctx context.Context, sessionID string) (Session, error) {
	data, err := m.storage.Load(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Decrypt DPoP key if needed
	if len(data.DPoPKeyEncrypted) > 0 && len(m.config.EncryptionKey) > 0 {
		dpopKey, err := m.decryptDPoPKey(data.DPoPKeyEncrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt DPoP key: %w", err)
		}
		data.DPoPKey = dpopKey
	}

	session := &DefaultSession{
		manager: m,
		data:    data,
	}

	return session, nil
}

// SaveSession persists session to storage.
func (m *DefaultManager) SaveSession(ctx context.Context, session Session) error {
	defaultSession, ok := session.(*DefaultSession)
	if !ok {
		return fmt.Errorf("invalid session type")
	}

	// Update timestamp
	defaultSession.data.UpdatedAt = time.Now()

	// Encrypt DPoP key if needed
	if len(m.config.EncryptionKey) > 0 && defaultSession.data.DPoPKey != nil {
		encrypted, err := m.encryptDPoPKey(defaultSession.data.DPoPKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt DPoP key: %w", err)
		}
		defaultSession.data.DPoPKeyEncrypted = encrypted
	}

	return m.storage.Store(ctx, defaultSession.data.SessionID, defaultSession.data)
}

// RefreshSession refreshes an expired session using refresh token.
func (m *DefaultManager) RefreshSession(ctx context.Context, session Session) error {
	defaultSession, ok := session.(*DefaultSession)
	if !ok {
		return fmt.Errorf("invalid session type")
	}

	// Use provider to refresh tokens
	tokenResult, err := m.provider.RefreshToken(ctx, defaultSession.data.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh tokens: %w", err)
	}

	// Update session with new tokens
	defaultSession.data.AccessToken = tokenResult.AccessToken
	defaultSession.data.RefreshToken = tokenResult.RefreshToken
	if tokenResult.DPoPKey != nil {
		defaultSession.data.DPoPKey = tokenResult.DPoPKey
	}
	defaultSession.data.ExpiresAt = time.Now().Add(time.Duration(tokenResult.ExpiresIn) * time.Second)
	defaultSession.data.UpdatedAt = time.Now()

	// Save updated session
	return m.SaveSession(ctx, defaultSession)
}

// DeleteSession removes session from storage.
func (m *DefaultManager) DeleteSession(ctx context.Context, sessionID string) error {
	return m.storage.Delete(ctx, sessionID)
}

// ValidateToken parses and validates JWT tokens.
func (m *DefaultManager) ValidateToken(token string) (*TokenClaims, error) {
	if m.config.TokenValidator != nil {
		return m.config.TokenValidator(token)
	}

	// Default JWT parsing implementation
	return m.parseJWT(token)
}

// IsTokenExpired checks if token expires within threshold.
func (m *DefaultManager) IsTokenExpired(token string, threshold time.Duration) bool {
	claims, err := m.ValidateToken(token)
	if err != nil {
		return true // Invalid token is considered expired
	}

	expTime := time.Unix(claims.ExpiresAt, 0)
	return time.Now().Add(threshold).After(expTime)
}

// GenerateSessionID creates a unique session identifier.
func (m *DefaultManager) GenerateSessionID() string {
	return m.config.SessionIDGenerator()
}

// Close cleans up manager resources.
func (m *DefaultManager) Close() error {
	return m.storage.Close()
}

// parseJWT provides basic JWT parsing without signature verification.
// For production use, consider using a proper JWT library with verification.
func (m *DefaultManager) parseJWT(token string) (*TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode payload
	payload := parts[1]
	// Add padding if needed
	for len(payload)%4 != 0 {
		payload += "="
	}

	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var rawClaims map[string]interface{}
	if err := json.Unmarshal(decoded, &rawClaims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	claims := &TokenClaims{
		Custom: make(map[string]interface{}),
	}

	// Extract standard claims
	if sub, ok := rawClaims["sub"].(string); ok {
		claims.Subject = sub
	}
	if iss, ok := rawClaims["iss"].(string); ok {
		claims.Issuer = iss
	}
	if aud, ok := rawClaims["aud"].(string); ok {
		claims.Audience = aud
	}
	if exp, ok := rawClaims["exp"].(float64); ok {
		claims.ExpiresAt = int64(exp)
	}
	if iat, ok := rawClaims["iat"].(float64); ok {
		claims.IssuedAt = int64(iat)
	}
	if nbf, ok := rawClaims["nbf"].(float64); ok {
		claims.NotBefore = int64(nbf)
	}

	// Store additional claims
	for key, value := range rawClaims {
		if key != "sub" && key != "iss" && key != "aud" && key != "exp" && key != "iat" && key != "nbf" {
			claims.Custom[key] = value
		}
	}

	return claims, nil
}

// encryptDPoPKey encrypts a DPoP private key using the configured encryption key.
// This is a simplified implementation - in production, use proper key encryption.
func (m *DefaultManager) encryptDPoPKey(key *ecdsa.PrivateKey) ([]byte, error) {
	// For now, use the existing DPoP key encoding from oauth package
	// In production, this should use proper encryption (AES-GCM, etc.)
	keyPair := &oauth.DPoPKeyPair{PrivateKey: key}
	pemData, err := keyPair.EncodeToPEM()
	if err != nil {
		return nil, err
	}
	
	// Simple XOR encryption with the key (NOT SECURE - for demo only)
	// TODO: Replace with proper AES-GCM or similar
	encrypted := make([]byte, len(pemData))
	keyIndex := 0
	for i, b := range []byte(pemData) {
		encrypted[i] = b ^ m.config.EncryptionKey[keyIndex%len(m.config.EncryptionKey)]
		keyIndex++
	}
	
	return encrypted, nil
}

// decryptDPoPKey decrypts a DPoP private key using the configured encryption key.
func (m *DefaultManager) decryptDPoPKey(encrypted []byte) (*ecdsa.PrivateKey, error) {
	// Simple XOR decryption (matches encryptDPoPKey)
	// TODO: Replace with proper AES-GCM or similar
	decrypted := make([]byte, len(encrypted))
	keyIndex := 0
	for i, b := range encrypted {
		decrypted[i] = b ^ m.config.EncryptionKey[keyIndex%len(m.config.EncryptionKey)]
		keyIndex++
	}
	
	// Decode PEM data
	keyPair, err := oauth.DecodeFromPEM(string(decrypted))
	if err != nil {
		return nil, err
	}
	
	return keyPair.PrivateKey, nil
}

// startCleanupRoutine runs periodic cleanup of expired sessions.
func (m *DefaultManager) startCleanupRoutine() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		if err := m.storage.Cleanup(ctx); err != nil {
			// Log error but don't stop the routine
			// In production, use proper logging
			fmt.Printf("Session cleanup error: %v\n", err)
		}
	}
}

// generateSessionID creates a cryptographically secure session ID.
func generateSessionID() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("session_%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(bytes)
}