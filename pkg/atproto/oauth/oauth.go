// Package oauth provides OAuth2 authentication with DPoP support for ATProtocol
package oauth

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	oauth "tangled.sh/icyphox.sh/atproto-oauth"
	"tangled.sh/icyphox.sh/atproto-oauth/helpers"
)

// OAuthClient handles OAuth2 authentication with ATProtocol servers
type OAuthClient struct {
	config      *Config
	oauthClient *oauth.Client
}

// Config contains OAuth configuration
type Config struct {
	ClientID       string
	ClientURI      string
	RedirectURI    string
	JWKSPrivateKey string
	JWKSPublicKey  string
	Scope          string
}

// TokenResult contains the result of a token exchange
type TokenResult struct {
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	UserDID      string            `json:"sub"`
	ExpiresIn    int64             `json:"expires_in"`
	DPoPKey      *ecdsa.PrivateKey `json:"-"`
}

// NewOAuthClient creates a new OAuth client
func NewOAuthClient(config *Config) (*OAuthClient, error) {
	// Extract first JWK from JWKS for client authentication
	clientJwkBytes, err := extractJWKFromJWKS(config.JWKSPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JWK from JWKS: %w", err)
	}

	// Parse JWK using tangled library helpers
	clientJwk, err := helpers.ParseJWKFromBytes(clientJwkBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client JWK: %w", err)
	}

	// Create tangled OAuth client with proper configuration
	oauthClient, err := oauth.NewClient(oauth.ClientArgs{
		ClientId:    config.ClientID,
		ClientJwk:   clientJwk,
		RedirectUri: config.RedirectURI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth client: %w", err)
	}

	return &OAuthClient{
		config:      config,
		oauthClient: oauthClient,
	}, nil
}

// GetAuthURL generates the OAuth authorization URL with PKCE
func (c *OAuthClient) GetAuthURL(state, codeChallenge string) string {
	// For ATProtocol, we typically use bsky.social as the default authorization server
	return fmt.Sprintf("https://bsky.social/oauth/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s&code_challenge=%s&code_challenge_method=S256",
		c.config.ClientID, c.config.RedirectURI, c.config.Scope, state, codeChallenge)
}

// ExchangeCode exchanges authorization code for access token
func (c *OAuthClient) ExchangeCode(ctx context.Context, code, codeVerifier string, dpopKey *ecdsa.PrivateKey, dpopNonce, authServerIssuer string) (*TokenResult, error) {
	log.Printf("[OAuth] Starting token exchange with code: %s", code[:8]+"...")

	// Convert DPoP key to JWK format for tangled library
	dpopKeyStr, err := convertECDSAToJWK(dpopKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DPoP key: %w", err)
	}

	// Parse DPoP key using tangled library helpers
	jwkKey, err := helpers.ParseJWKFromBytes([]byte(dpopKeyStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DPoP JWK: %w", err)
	}

	// Default auth server issuer if not provided
	if authServerIssuer == "" {
		authServerIssuer = "https://bsky.social"
	}

	log.Printf("[OAuth] Using tangled InitialTokenRequest with issuer: %s, nonce: %s", authServerIssuer, dpopNonce)

	// Use tangled library's InitialTokenRequest
	tokenResp, err := c.oauthClient.InitialTokenRequest(
		ctx,
		code,
		authServerIssuer,
		codeVerifier,
		dpopNonce,
		jwkKey,
	)
	if err != nil {
		log.Printf("[OAuth] Token exchange failed: %v", err)
		return nil, fmt.Errorf("tangled token exchange failed: %w", err)
	}

	log.Printf("[OAuth] Token exchange successful, got access token")

	return &TokenResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		DPoPKey:      dpopKey,
		UserDID:      tokenResp.Sub,
		ExpiresIn:    int64(tokenResp.ExpiresIn),
	}, nil
}

// RefreshToken refreshes an expired access token
func (c *OAuthClient) RefreshToken(ctx context.Context, refreshToken string, dpopKey *ecdsa.PrivateKey, dpopNonce, authServerIssuer string) (*TokenResult, error) {
	log.Printf("[OAuth] Starting token refresh")

	// Convert DPoP key to JWK format for tangled library
	dpopKeyStr, err := convertECDSAToJWK(dpopKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DPoP key: %w", err)
	}

	// Parse DPoP key using tangled library helpers
	jwkKey, err := helpers.ParseJWKFromBytes([]byte(dpopKeyStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DPoP JWK: %w", err)
	}

	// Use tangled library's RefreshTokenRequest
	tokenResp, err := c.oauthClient.RefreshTokenRequest(
		ctx,
		refreshToken,
		authServerIssuer,
		dpopNonce,
		jwkKey,
	)
	if err != nil {
		log.Printf("[OAuth] Token refresh failed: %v", err)
		return nil, fmt.Errorf("tangled token refresh failed: %w", err)
	}

	log.Printf("[OAuth] Token refresh successful")

	return &TokenResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		DPoPKey:      dpopKey,
		UserDID:      tokenResp.Sub,
		ExpiresIn:    int64(tokenResp.ExpiresIn),
	}, nil
}

// extractJWKFromJWKS extracts the first JWK from a JWKS format
func extractJWKFromJWKS(jwksStr string) ([]byte, error) {
	var jwks struct {
		Keys []json.RawMessage `json:"keys"`
	}

	if err := json.Unmarshal([]byte(jwksStr), &jwks); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	if len(jwks.Keys) == 0 {
		return nil, fmt.Errorf("no keys found in JWKS")
	}

	// Return the first key as raw bytes
	return jwks.Keys[0], nil
}

// convertECDSAToJWK converts an ECDSA private key to JWK format for tangled library
func convertECDSAToJWK(key *ecdsa.PrivateKey) (string, error) {
	jwk := map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   encodeCoordinate(key.PublicKey.X.Bytes()),
		"y":   encodeCoordinate(key.PublicKey.Y.Bytes()),
		"d":   encodeCoordinate(key.D.Bytes()),
		"alg": "ES256",
		"use": "sig",
	}

	jwkBytes, err := json.Marshal(jwk)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWK: %w", err)
	}

	return string(jwkBytes), nil
}

// encodeCoordinate encodes a coordinate value for JWK using base64url encoding
func encodeCoordinate(bytes []byte) string {
	// Ensure 32 bytes for P-256
	if len(bytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(bytes):], bytes)
		bytes = padded
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}