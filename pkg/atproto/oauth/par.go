// Package oauth provides Pushed Authorization Request (PAR) support for ATProtocol OAuth
package oauth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PARClient handles Pushed Authorization Request (PAR) operations for ATProtocol OAuth
type PARClient struct {
	httpClient *http.Client
}

// NewPARClient creates a new PAR client
func NewPARClient() *PARClient {
	return &PARClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PARRequest represents a Pushed Authorization Request
type PARRequest struct {
	ClientID            string `json:"client_id"`
	ResponseType        string `json:"response_type"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	ClientAssertionType string `json:"client_assertion_type"`
	ClientAssertion     string `json:"client_assertion"`
}

// PARResponse represents the response from a PAR request
type PARResponse struct {
	RequestURI       string `json:"request_uri"`
	ExpiresIn        int    `json:"expires_in"`
	DPoPNonce        string `json:"-"` // Not part of JSON response, but captured for session
	AuthServerIssuer string `json:"-"` // Authorization server issuer for token exchange
}

// ClientAssertionClaims represents JWT claims for client assertion
type ClientAssertionClaims struct {
	Iss string `json:"iss"` // Client ID
	Sub string `json:"sub"` // Client ID
	Aud string `json:"aud"` // Authorization server
	Exp int64  `json:"exp"` // Expiration time
	Iat int64  `json:"iat"` // Issued at
	Jti string `json:"jti"` // JWT ID
}

// CreateClientAssertion creates a JWT for client authentication using private_key_jwt
func (p *PARClient) CreateClientAssertion(clientID, authServer string, cfg *ProviderConfig) (string, error) {
	// Parse the private key from JWKS
	var jwks struct {
		Keys []map[string]interface{} `json:"keys"`
	}

	if err := json.Unmarshal([]byte(cfg.JWKSPrivateKey), &jwks); err != nil || len(jwks.Keys) == 0 {
		return "", fmt.Errorf("failed to parse JWKS private key: %w", err)
	}

	// Get the private key (assuming first key)
	privateKeyJWK := jwks.Keys[0]

	// Extract key components
	kty, _ := privateKeyJWK["kty"].(string)
	if kty != "EC" {
		return "", fmt.Errorf("unsupported key type: %s", kty)
	}

	crv, _ := privateKeyJWK["crv"].(string)
	if crv != "P-256" {
		return "", fmt.Errorf("unsupported curve: %s", crv)
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(privateKeyJWK["x"].(string))
	if err != nil {
		return "", fmt.Errorf("failed to decode x coordinate: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(privateKeyJWK["y"].(string))
	if err != nil {
		return "", fmt.Errorf("failed to decode y coordinate: %w", err)
	}

	dBytes, err := base64.RawURLEncoding.DecodeString(privateKeyJWK["d"].(string))
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	// Reconstruct the private key
	curve := elliptic.P256()
	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)
	d := new(big.Int).SetBytes(dBytes)

	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: d,
	}

	// Create JWT header
	header := map[string]interface{}{
		"typ": "JWT",
		"alg": "ES256",
		"kid": privateKeyJWK["kid"],
	}

	// Create JWT claims
	now := time.Now()
	claims := ClientAssertionClaims{
		Iss: clientID,
		Sub: clientID,
		Aud: authServer,
		Exp: now.Add(5 * time.Minute).Unix(),
		Iat: now.Unix(),
		Jti: generateRandomString(32),
	}

	// Encode header and payload
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsBytes)

	// Create signing input
	signingInput := headerEncoded + "." + claimsEncoded

	// Sign
	hash := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Encode signature in IEEE P1363 format
	signature := make([]byte, 64) // 32 bytes for r + 32 bytes for s
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	copy(signature[32-len(rBytes):32], rBytes)
	copy(signature[64-len(sBytes):64], sBytes)

	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureEncoded, nil
}

// PerformPAR performs a Pushed Authorization Request
func (p *PARClient) PerformPAR(ctx context.Context, parEndpoint string, metadata *AuthorizationServerMetadata, codeVerifier, state string, dpopKey DPoPKeyPair, cfg *ProviderConfig) (*PARResponse, error) {
	// Create client assertion
	clientAssertion, err := p.CreateClientAssertion(cfg.ClientID, metadata.Issuer, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client assertion: %w", err)
	}

	// Generate code challenge
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Create DPoP proof for PAR request
	dpopJWT, err := dpopKey.CreateDPoPJWT("POST", parEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create DPoP proof: %w", err)
	}

	// Prepare PAR request data
	data := url.Values{}
	data.Set("client_id", cfg.ClientID)
	data.Set("response_type", "code")
	data.Set("code_challenge", codeChallenge)
	data.Set("code_challenge_method", "S256")
	data.Set("redirect_uri", cfg.RedirectURI)
	data.Set("scope", "atproto transition:generic")
	data.Set("state", state)
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", clientAssertion)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", parEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create PAR request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("DPoP", dpopJWT)

	// Make the request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make PAR request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Ignore close error
			_ = err
		}
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PAR response: %w", err)
	}

	// Check for DPoP nonce requirement
	if resp.StatusCode == http.StatusBadRequest {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if errorResp["error"] == "use_dpop_nonce" {
				// Retry with DPoP nonce
				dpopNonce := resp.Header.Get("DPoP-Nonce")
				if dpopNonce != "" {
					parResp, err := p.performPARWithNonce(ctx, parEndpoint, data, dpopKey, dpopNonce)
					if err != nil {
						return nil, err
					}
					// Store the nonce and auth server issuer for later use in token exchange
					parResp.DPoPNonce = dpopNonce
					parResp.AuthServerIssuer = metadata.Issuer
					return parResp, nil
				}
			}
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("PAR request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var parResp PARResponse
	if err := json.Unmarshal(body, &parResp); err != nil {
		return nil, fmt.Errorf("failed to parse PAR response: %w", err)
	}

	// Store auth server issuer for token exchange
	parResp.AuthServerIssuer = metadata.Issuer

	return &parResp, nil
}

// performPARWithNonce performs PAR with DPoP nonce
func (p *PARClient) performPARWithNonce(ctx context.Context, parEndpoint string, data url.Values, dpopKey DPoPKeyPair, nonce string) (*PARResponse, error) {
	// Create DPoP proof with nonce
	dpopJWT, err := dpopKey.CreateDPoPJWTWithNonce("POST", parEndpoint, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to create DPoP proof with nonce: %w", err)
	}

	// Create new request
	req, err := http.NewRequestWithContext(ctx, "POST", parEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create PAR retry request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("DPoP", dpopJWT)

	// Make the retry request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make PAR retry request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Ignore close error
			_ = err
		}
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PAR retry response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("PAR retry request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var parResp PARResponse
	if err := json.Unmarshal(body, &parResp); err != nil {
		return nil, fmt.Errorf("failed to parse PAR retry response: %w", err)
	}

	return &parResp, nil
}

// generateCodeChallenge creates a PKCE code challenge from a verifier
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[mathrand.Intn(len(charset))]
	}
	return string(b)
}
