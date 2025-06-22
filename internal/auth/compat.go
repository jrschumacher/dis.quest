// Package auth provides backward compatibility wrappers for web application code
// Core ATProtocol functionality has been moved to pkg/atproto
package auth

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
	"golang.org/x/oauth2"
)

// Type aliases for backward compatibility
type DPoPKeyPair = oauth.DPoPKeyPair
type AuthorizationServerMetadata = oauth.AuthorizationServerMetadata
type DPoPPKCETransport = oauth.DPoPPKCETransport
type ProviderConfig = oauth.ProviderConfig

// DPoP key management - delegate to pkg/atproto/oauth
func GenerateDPoPKeyPair() (*DPoPKeyPair, error) {
	return oauth.GenerateDPoPKeyPair()
}

func GetDPoPKeyFromCookie(r *http.Request) (*ecdsa.PrivateKey, error) {
	return oauth.GetDPoPKeyFromCookie(r)
}

func SetDPoPKeyCookie(w http.ResponseWriter, key *ecdsa.PrivateKey, isDev bool) error {
	return oauth.SetDPoPKeyCookie(w, key, isDev)
}

func SetDPoPNonceCookie(w http.ResponseWriter, nonce string, isDev bool) error {
	return oauth.SetDPoPNonceCookie(w, nonce, isDev)
}

func SetAuthServerIssuerCookie(w http.ResponseWriter, issuer string, isDev bool) error {
	return oauth.SetAuthServerIssuerCookie(w, issuer, isDev)
}

// PKCE functions - delegate to pkg/atproto/oauth  
func GeneratePKCE() (codeVerifier, codeChallenge string, err error) {
	return oauth.GeneratePKCE()
}

// OAuth configuration - delegate to pkg/atproto/oauth
func OAuth2Config(metadata *AuthorizationServerMetadata, cfg *config.Config) *oauth2.Config {
	providerConfig := &oauth.ProviderConfig{
		ClientID:       cfg.OAuthClientID,
		ClientURI:      cfg.PublicDomain,
		RedirectURI:    cfg.OAuthRedirectURL,
		PDSEndpoint:    cfg.PDSEndpoint,
		JWKSPrivateKey: cfg.JWKSPrivate,
		JWKSPublicKey:  cfg.JWKSPublic,
		Scope:          "atproto transition:generic",
	}
	return oauth.OAuth2Config(metadata, providerConfig)
}

// Token exchange - delegate to pkg/atproto/oauth
func ExchangeCodeForTokenWithDPoP(ctx context.Context, metadata *AuthorizationServerMetadata, code, codeVerifier string, dpopKey *ecdsa.PrivateKey, cfg *config.Config) (*oauth2.Token, error) {
	providerConfig := &oauth.ProviderConfig{
		ClientID:       cfg.OAuthClientID,
		ClientURI:      cfg.PublicDomain,
		RedirectURI:    cfg.OAuthRedirectURL,
		PDSEndpoint:    cfg.PDSEndpoint,
		JWKSPrivateKey: cfg.JWKSPrivate,
		JWKSPublicKey:  cfg.JWKSPublic,
		Scope:          "atproto transition:generic",
	}
	dpopKeyPair := &oauth.DPoPKeyPair{PrivateKey: dpopKey}
	return oauth.ExchangeCodeForTokenWithDPoP(ctx, metadata, code, codeVerifier, *dpopKeyPair, providerConfig)
}

// Discovery functions - delegate to pkg/atproto/oauth
func DiscoverPDS(handle string) (string, error) {
	return oauth.DiscoverPDS(handle)
}

func DiscoverAuthorizationServer(handle string) (*AuthorizationServerMetadata, error) {
	return oauth.DiscoverAuthorizationServer(handle)
}

// PAR functions - delegate to pkg/atproto/oauth
func NewPARClient() *oauth.PARClient {
	return oauth.NewPARClient()
}

// State token generation - used in OAuth flow
func GenerateStateToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		// fallback: not cryptographically secure, but avoids panic
		return base64.RawURLEncoding.EncodeToString([]byte("fallback_state_token"))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// DPoP key encoding/decoding - delegate to pkg/atproto/oauth
func EncodeDPoPPrivateKeyToPEM(key *ecdsa.PrivateKey) (string, error) {
	keyPair := &oauth.DPoPKeyPair{PrivateKey: key}
	return keyPair.EncodeToPEM()
}

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

// Token expiration checking
func IsTokenExpiringSoon(accessToken string, thresholdMinutes int) bool {
	// This function was originally in auth.go - delegate to a proper JWT parser
	// For now, provide a simple implementation that parses the JWT exp claim
	if accessToken == "" {
		return true
	}
	
	// Parse JWT to get expiration
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return true // Invalid token format
	}
	
	// Decode payload
	payload := parts[1]
	// Add padding if needed for base64 decoding
	for len(payload)%4 != 0 {
		payload += "="
	}
	
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return true // Can't decode, assume expired
	}
	
	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return true // Can't parse claims, assume expired
	}
	
	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		// Check if token expires within threshold
		return time.Now().Add(time.Duration(thresholdMinutes) * time.Minute).After(expTime)
	}
	
	return true // No exp claim, assume expired
}

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