package auth

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jrschumacher/dis.quest/pkg/atproto"
	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
)

// CreateSessionRequest represents a session creation request
type CreateSessionRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// CreateSessionResponse represents a session creation response
type CreateSessionResponse struct {
	AccessJwt  string `json:"accessJwt"`
	RefreshJwt string `json:"refreshJwt"`
	Did        string `json:"did"`
	Handle     string `json:"handle"`
}

// CreateSession calls the ATProto createSession endpoint with handle and app password
func CreateSession(pds, handle, password string) (*CreateSessionResponse, error) {
	url := fmt.Sprintf("%s/xrpc/com.atproto.server.createSession", pds)
	body, _ := json.Marshal(CreateSessionRequest{
		Identifier: handle,
		Password:   password,
	})
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// TODO: refactor to allow injecting an HTTP client so this can be tested without network access
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return nil, ErrInvalidCredentials
	}
	var out CreateSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DPoPKeyPair is a type alias for the consolidated implementation
type DPoPKeyPair = oauth.DPoPKeyPair

// GenerateDPoPKeyPair generates a new ECDSA P-256 keypair for DPoP
func GenerateDPoPKeyPair() (*DPoPKeyPair, error) {
	return oauth.GenerateDPoPKeyPair()
}

// EncodeDPoPPrivateKeyToPEM encodes the private key as PEM for storage (optional)
func EncodeDPoPPrivateKeyToPEM(key *ecdsa.PrivateKey) (string, error) {
	keyPair := &oauth.DPoPKeyPair{PrivateKey: key}
	return keyPair.EncodeToPEM()
}

// DecodeDPoPPrivateKeyFromPEM decodes a PEM-encoded private key
func DecodeDPoPPrivateKeyFromPEM(pemStr string) (*ecdsa.PrivateKey, error) {
	keyPair, err := oauth.DecodeFromPEM(pemStr)
	if err != nil {
		// Convert the generic error to our specific error for backward compatibility
		if strings.Contains(err.Error(), "invalid PEM block") {
			return nil, ErrInvalidPEMBlock
		}
		return nil, err
	}
	return keyPair.PrivateKey, nil
}

// SetDPoPKeyCookie stores the DPoP private key in a secure, HttpOnly cookie
func SetDPoPKeyCookie(w http.ResponseWriter, key *ecdsa.PrivateKey, isDev bool) error {
	return oauth.SetDPoPKeyCookie(w, key, isDev)
}

// GetDPoPKeyFromCookie retrieves and decodes the DPoP private key from the cookie
func GetDPoPKeyFromCookie(r *http.Request) (*ecdsa.PrivateKey, error) {
	return oauth.GetDPoPKeyFromCookie(r)
}

// ClearDPoPKeyCookie clears the DPoP key cookie
func ClearDPoPKeyCookie(w http.ResponseWriter, isDev bool) {
	oauth.ClearDPoPKeyCookie(w, isDev)
}

// SetDPoPNonceCookie stores the DPoP nonce in a secure, HttpOnly cookie
func SetDPoPNonceCookie(w http.ResponseWriter, nonce string, isDev bool) error {
	return oauth.SetDPoPNonceCookie(w, nonce, isDev)
}

// GetDPoPNonceFromCookie retrieves the DPoP nonce from the cookie
func GetDPoPNonceFromCookie(r *http.Request) (string, error) {
	return oauth.GetDPoPNonceFromCookie(r)
}

// ClearDPoPNonceCookie clears the DPoP nonce cookie
func ClearDPoPNonceCookie(w http.ResponseWriter, isDev bool) {
	oauth.ClearDPoPNonceCookie(w, isDev)
}

// SetAuthServerIssuerCookie stores the auth server issuer in a secure, HttpOnly cookie
func SetAuthServerIssuerCookie(w http.ResponseWriter, issuer string, isDev bool) error {
	return oauth.SetAuthServerIssuerCookie(w, issuer, isDev)
}

// GetAuthServerIssuerFromCookie retrieves the auth server issuer from the cookie
func GetAuthServerIssuerFromCookie(r *http.Request) (string, error) {
	return oauth.GetAuthServerIssuerFromCookie(r)
}

// CreateDPoPJWT creates a DPoP JWT for the given HTTP method and URL
func CreateDPoPJWT(key *ecdsa.PrivateKey, method, targetURL string) (string, error) {
	keyPair := &oauth.DPoPKeyPair{PrivateKey: key}
	return keyPair.CreateDPoPJWT(method, targetURL)
}

// CreateDPoPJWTWithNonce creates a DPoP JWT for the given HTTP method and URL with optional nonce
func CreateDPoPJWTWithNonce(key *ecdsa.PrivateKey, method, targetURL, nonce string) (string, error) {
	keyPair := &oauth.DPoPKeyPair{PrivateKey: key}
	return keyPair.CreateDPoPJWTWithNonce(method, targetURL, nonce)
}

// CreateDPoPJWTWithAccessToken creates a DPoP JWT with access token hash (ath claim)
func CreateDPoPJWTWithAccessToken(key *ecdsa.PrivateKey, method, targetURL, nonce, accessToken string) (string, error) {
	keyPair := &oauth.DPoPKeyPair{PrivateKey: key}
	return keyPair.CreateDPoPJWTWithAccessToken(method, targetURL, nonce, accessToken)
}

// CalculateJWKThumbprint calculates the SHA-256 thumbprint of a JWK
func CalculateJWKThumbprint(_ map[string]interface{}) (string, error) {
	// Legacy function - deprecated in favor of DPoPKeyPair.CalculateJWKThumbprint()
	return "", fmt.Errorf("deprecated: use DPoPKeyPair.CalculateJWKThumbprint() instead")
}

// GetDPoPKeyFromRequest retrieves the DPoP private key from the request context or cookies
func GetDPoPKeyFromRequest(r *http.Request) (string, error) {
	// Get DPoP key from cookie (PEM encoded)
	dpopKey, err := GetDPoPKeyFromCookie(r)
	if err != nil {
		return "", fmt.Errorf("failed to get DPoP key from cookie: %w", err)
	}
	
	// Convert to JWK JSON format for tangled library
	keyPair := &DPoPKeyPair{PrivateKey: dpopKey}
	jwk := keyPair.PublicJWK()
	
	// Add private key component
	jwk["d"] = base64.RawURLEncoding.EncodeToString(dpopKey.D.Bytes())
	
	// Marshal to JSON
	jwkBytes, err := json.Marshal(jwk)
	if err != nil {
		return "", fmt.Errorf("failed to marshal DPoP key to JWK: %w", err)
	}
	
	return string(jwkBytes), nil
}

// SessionWrapper wraps pkg/atproto.Session while maintaining cookie-based session management
// This provides enhanced functionality while preserving existing session interface
type SessionWrapper struct {
	atprotoSession *atproto.Session
	accessToken    string
	refreshToken   string
	userDID        string
	dpopKey        *ecdsa.PrivateKey
}

// NewSessionWrapper creates a new session wrapper from authentication tokens
func NewSessionWrapper(accessToken, refreshToken, userDID string, dpopKey *ecdsa.PrivateKey, atprotoClient *atproto.Client) (*SessionWrapper, error) {
	// Create the internal atproto.Session
	// Note: We'll need to create this through the client's exchange process or mock it for compatibility
	// For now, we create a wrapper that manages both interfaces
	wrapper := &SessionWrapper{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		userDID:      userDID,
		dpopKey:      dpopKey,
	}
	
	// If we have an atproto client, we can create a proper session
	// This will be populated during OAuth flow
	if atprotoClient != nil {
		// The atproto.Session would typically be created during ExchangeCode
		// For now, we store the client reference for later use
		wrapper.atprotoSession = nil // Will be set during proper OAuth flow
	}
	
	return wrapper, nil
}

// GetAccessToken returns the current access token
func (sw *SessionWrapper) GetAccessToken() string {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.GetAccessToken()
	}
	return sw.accessToken
}

// GetRefreshToken returns the current refresh token
func (sw *SessionWrapper) GetRefreshToken() string {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.GetRefreshToken()
	}
	return sw.refreshToken
}

// GetUserDID returns the user's DID
func (sw *SessionWrapper) GetUserDID() string {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.GetUserDID()
	}
	return sw.userDID
}

// GetDPoPKey returns the DPoP private key
func (sw *SessionWrapper) GetDPoPKey() *ecdsa.PrivateKey {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.GetDPoPKey()
	}
	return sw.dpopKey
}

// CreateRecord creates a record using the internal atproto.Session if available
func (sw *SessionWrapper) CreateRecord(collection, rkey string, record interface{}) (*atproto.RecordResult, error) {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.CreateRecord(collection, rkey, record)
	}
	return nil, fmt.Errorf("atproto session not available for CreateRecord")
}

// GetRecord retrieves a record using the internal atproto.Session if available
func (sw *SessionWrapper) GetRecord(collection, rkey string, result interface{}) error {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.GetRecord(collection, rkey, result)
	}
	return fmt.Errorf("atproto session not available for GetRecord")
}

// ListRecords lists records using the internal atproto.Session if available
func (sw *SessionWrapper) ListRecords(collection string, limit int, cursor string) (*atproto.ListRecordsResult, error) {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.ListRecords(collection, limit, cursor)
	}
	return nil, fmt.Errorf("atproto session not available for ListRecords")
}

// UpdateRecord updates a record using the internal atproto.Session if available
func (sw *SessionWrapper) UpdateRecord(collection, rkey string, record interface{}) (*atproto.RecordResult, error) {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.UpdateRecord(collection, rkey, record)
	}
	return nil, fmt.Errorf("atproto session not available for UpdateRecord")
}

// DeleteRecord deletes a record using the internal atproto.Session if available
func (sw *SessionWrapper) DeleteRecord(collection, rkey string) error {
	if sw.atprotoSession != nil {
		return sw.atprotoSession.DeleteRecord(collection, rkey)
	}
	return fmt.Errorf("atproto session not available for DeleteRecord")
}

// SetAtprotoSession sets the internal atproto.Session (used during OAuth flow)
func (sw *SessionWrapper) SetAtprotoSession(session *atproto.Session) {
	sw.atprotoSession = session
}

// GetAtprotoSession returns the internal atproto.Session (if available)
func (sw *SessionWrapper) GetAtprotoSession() *atproto.Session {
	return sw.atprotoSession
}

// SaveToCookies saves session data to HTTP cookies
func (sw *SessionWrapper) SaveToCookies(w http.ResponseWriter, isDev bool) error {
	// Use existing cookie management functions
	SetSessionCookieWithEnv(w, sw.GetAccessToken(), []string{sw.GetRefreshToken()}, isDev)
	
	// Save DPoP key if available
	if sw.GetDPoPKey() != nil {
		if err := SetDPoPKeyCookie(w, sw.GetDPoPKey(), isDev); err != nil {
			return fmt.Errorf("failed to save DPoP key to cookie: %w", err)
		}
	}
	
	return nil
}

// LoadSessionFromCookies creates session wrapper from HTTP cookies  
func LoadSessionFromCookies(r *http.Request) (*SessionWrapper, error) {
	// Get access token from cookie
	accessToken, err := GetSessionCookie(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token from cookie: %w", err)
	}
	
	// Get refresh token (optional)
	refreshToken, _ := GetRefreshTokenCookie(r)
	
	// Get DPoP key (optional)
	dpopKey, _ := GetDPoPKeyFromCookie(r)
	
	// Extract user DID from access token (JWT parsing)
	userDID := ""
	// Simple JWT parsing to get subject
	parts := strings.Split(accessToken, ".")
	if len(parts) >= 2 {
		// Decode payload
		payload := parts[1]
		// Add padding if needed for base64 decoding
		for len(payload)%4 != 0 {
			payload += "="
		}
		if decoded, err := base64.StdEncoding.DecodeString(payload); err == nil {
			var claims map[string]interface{}
			if err := json.Unmarshal(decoded, &claims); err == nil {
				if sub, ok := claims["sub"].(string); ok {
					userDID = sub
				}
			}
		}
	}
	
	return NewSessionWrapper(accessToken, refreshToken, userDID, dpopKey, nil)
}
