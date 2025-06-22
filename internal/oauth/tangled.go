package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	oauth "tangled.sh/icyphox.sh/atproto-oauth"
	"tangled.sh/icyphox.sh/atproto-oauth/helpers"
	
	"github.com/jrschumacher/dis.quest/internal/auth"
)

// TangledOAuthProvider implements OAuth using the tangled-sh library
// This provider focuses on token exchange to fix the "Bad token scope" error
// by using proper private_key_jwt client authentication
type TangledOAuthProvider struct {
	config      *Config
	oauthClient *oauth.Client
}

// NewTangledOAuthProvider creates a new tangled OAuth provider
func NewTangledOAuthProvider(config *Config) *TangledOAuthProvider {
	// Extract first JWK from JWKS for client authentication
	clientJwkBytes, err := extractJWKFromJWKS(config.JWKSPrivateKey)
	if err != nil {
		log.Printf("[TangledOAuthProvider] Failed to extract JWK from JWKS: %v", err)
		return nil
	}

	// Parse JWK using tangled library helpers
	clientJwk, err := helpers.ParseJWKFromBytes(clientJwkBytes)
	if err != nil {
		log.Printf("[TangledOAuthProvider] Failed to parse client JWK: %v", err)
		return nil
	}

	// Create tangled OAuth client with proper configuration
	oauthClient, err := oauth.NewClient(oauth.ClientArgs{
		ClientId:    config.ClientID,
		ClientJwk:   clientJwk,
		RedirectUri: config.RedirectURI,
	})
	if err != nil {
		log.Printf("[TangledOAuthProvider] Failed to create OAuth client: %v", err)
		return nil
	}

	return &TangledOAuthProvider{
		config:      config,
		oauthClient: oauthClient,
	}
}

// GetAuthURL generates the OAuth authorization URL with PKCE
func (t *TangledOAuthProvider) GetAuthURL(state, codeChallenge string) string {
	return fmt.Sprintf("%s/oauth/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s&code_challenge=%s&code_challenge_method=S256",
		t.config.PDSEndpoint, t.config.ClientID, t.config.RedirectURI, t.config.Scope, state, codeChallenge)
}

// ExchangeToken exchanges authorization code for access token using tangled library
func (t *TangledOAuthProvider) ExchangeToken(ctx context.Context, code, codeVerifier string) (*TokenResult, error) {
	log.Printf("[TangledOAuthProvider] Starting token exchange with code: %s", code[:8]+"...")

	// Extract HTTP request from context to access cookies
	req, ok := ctx.Value("http_request").(*http.Request)
	if !ok {
		return nil, fmt.Errorf("http request not found in context")
	}

	// Get DPoP key from cookies (in JWK format for tangled library)
	dpopKeyStr, err := auth.GetDPoPKeyFromRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get DPoP key: %w", err)
	}

	// Parse DPoP key using tangled library helpers
	jwkKey, err := helpers.ParseJWKFromBytes([]byte(dpopKeyStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DPoP JWK: %w", err)
	}

	// Get DPoP nonce from cookies
	dpopNonce, err := auth.GetDPoPNonceFromCookie(req)
	if err != nil {
		log.Printf("[TangledOAuthProvider] No DPoP nonce found: %v", err)
		dpopNonce = "" // Continue without nonce
	}

	// Get authorization server issuer from cookies (set during PAR)
	authServerIss, err := auth.GetAuthServerIssuerFromCookie(req)
	if err != nil {
		log.Printf("[TangledOAuthProvider] No auth server issuer found: %v", err)
		authServerIss = "https://bsky.social" // Fallback to default
	}

	log.Printf("[TangledOAuthProvider] Using tangled InitialTokenRequest with issuer: %s, nonce: %s", authServerIss, dpopNonce)

	// Use tangled library's InitialTokenRequest - this is the key fix for "Bad token scope"
	tokenResp, err := t.oauthClient.InitialTokenRequest(
		ctx,
		code,
		authServerIss,
		codeVerifier,
		dpopNonce,
		jwkKey,
	)
	if err != nil {
		log.Printf("[TangledOAuthProvider] Token exchange failed: %v", err)
		return nil, fmt.Errorf("tangled token exchange failed: %w", err)
	}

	log.Printf("[TangledOAuthProvider] Token exchange successful, got access token")

	// Convert tangled response to our TokenResult format
	// Extract ECDSA key from the original key for compatibility
	dpopECDSAKey, err := auth.GetDPoPKeyFromCookie(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get ECDSA key for result: %w", err)
	}

	return &TokenResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		DPoPKey:      dpopECDSAKey,
		UserDID:      tokenResp.Sub, // Subject from token response
		ExpiresIn:    int64(tokenResp.ExpiresIn),
	}, nil
}

// CreateAuthorizedClient creates an XRPC client with the given token
func (t *TangledOAuthProvider) CreateAuthorizedClient(token *TokenResult) (XRPCClient, error) {
	return nil, fmt.Errorf("tangled provider not yet fully implemented")
}

// RefreshToken refreshes an expired access token
func (t *TangledOAuthProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error) {
	// Get HTTP request from context for session access
	req, ok := ctx.Value("http_request").(*http.Request)
	if !ok {
		return nil, fmt.Errorf("HTTP request not found in context")
	}
	
	log.Printf("[TangledOAuthProvider] Starting token refresh")
	
	// Get auth server issuer from session/cookie
	authServerIssuer, err := auth.GetAuthServerIssuerFromCookie(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth server issuer: %w", err)
	}
	
	// Get DPoP key from cookies (in JWK format for tangled library)
	dpopKeyStr, err := auth.GetDPoPKeyFromRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get DPoP key: %w", err)
	}

	// Parse DPoP key using tangled library helpers
	jwkKey, err := helpers.ParseJWKFromBytes([]byte(dpopKeyStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DPoP JWK: %w", err)
	}
	
	// Get DPoP nonce (may be empty for refresh)
	dpopNonce, _ := auth.GetDPoPNonceFromCookie(req)
	
	// Use tangled library's RefreshTokenRequest
	tokenResp, err := t.oauthClient.RefreshTokenRequest(
		ctx,
		refreshToken,
		authServerIssuer,
		dpopNonce,
		jwkKey,
	)
	if err != nil {
		log.Printf("[TangledOAuthProvider] Token refresh failed: %v", err)
		return nil, fmt.Errorf("tangled token refresh failed: %w", err)
	}
	
	log.Printf("[TangledOAuthProvider] Token refresh successful")
	
	// Extract ECDSA key from the original key for compatibility
	dpopECDSAKey, err := auth.GetDPoPKeyFromCookie(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get ECDSA key for result: %w", err)
	}
	
	return &TokenResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		DPoPKey:      dpopECDSAKey,
		UserDID:      tokenResp.Sub,
		ExpiresIn:    int64(tokenResp.ExpiresIn),
	}, nil
}

// GetProviderName returns the name of this provider implementation
func (t *TangledOAuthProvider) GetProviderName() string {
	return "tangled"
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