package auth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"time"
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

// DPoPKeyPair holds an ECDSA P-256 keypair for DPoP
// Only the private key is needed to sign DPoP JWTs; public key is used for JWK
type DPoPKeyPair struct {
	PrivateKey *ecdsa.PrivateKey
}

// GenerateDPoPKeyPair generates a new ECDSA P-256 keypair for DPoP
func GenerateDPoPKeyPair() (*DPoPKeyPair, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return &DPoPKeyPair{PrivateKey: priv}, nil
}

// EncodeDPoPPrivateKeyToPEM encodes the private key as PEM for storage (optional)
func EncodeDPoPPrivateKeyToPEM(key *ecdsa.PrivateKey) (string, error) {
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", err
	}
	pemBlock := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	return base64.RawURLEncoding.EncodeToString(pemBlock), nil
}

// DecodeDPoPPrivateKeyFromPEM decodes a PEM-encoded private key
func DecodeDPoPPrivateKeyFromPEM(pemStr string) (*ecdsa.PrivateKey, error) {
	pemBytes, err := base64.RawURLEncoding.DecodeString(pemStr)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

// DPoPPublicJWK returns the public key as a JWK map (for DPoP JWT header)
func (k *DPoPKeyPair) DPoPPublicJWK() map[string]interface{} {
	pub := k.PrivateKey.PublicKey
	return map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   base64.RawURLEncoding.EncodeToString(pub.X.Bytes()),
		"y":   base64.RawURLEncoding.EncodeToString(pub.Y.Bytes()),
		"alg": "ES256",
		"use": "sig",
	}
}

const dpopKeyCookieName = "dpop_key"

// SetDPoPKeyCookie stores the DPoP private key in a secure, HttpOnly cookie
func SetDPoPKeyCookie(w http.ResponseWriter, key *ecdsa.PrivateKey, isDev bool) error {
	pemStr, err := EncodeDPoPPrivateKeyToPEM(key)
	if err != nil {
		return err
	}
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     dpopKeyCookieName,
		Value:    pemStr,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes for OAuth flow
	})
	return nil
}

// GetDPoPKeyFromCookie retrieves and decodes the DPoP private key from the cookie
func GetDPoPKeyFromCookie(r *http.Request) (*ecdsa.PrivateKey, error) {
	cookie, err := r.Cookie(dpopKeyCookieName)
	if err != nil {
		return nil, err
	}
	return DecodeDPoPPrivateKeyFromPEM(cookie.Value)
}

// ClearDPoPKeyCookie clears the DPoP key cookie
func ClearDPoPKeyCookie(w http.ResponseWriter, isDev bool) {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     dpopKeyCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// DPoPJWTHeader represents the header of a DPoP JWT
type DPoPJWTHeader struct {
	Typ string                 `json:"typ"`
	Alg string                 `json:"alg"`
	JWK map[string]interface{} `json:"jwk"`
}

// DPoPJWTPayload represents the payload of a DPoP JWT
type DPoPJWTPayload struct {
	JTI   string `json:"jti"`
	HTM   string `json:"htm"`
	HTU   string `json:"htu"`
	IAT   int64  `json:"iat"`
	Nonce string `json:"nonce,omitempty"`
}

// CreateDPoPJWT creates a DPoP JWT for the given HTTP method and URL
func CreateDPoPJWT(key *ecdsa.PrivateKey, method, targetURL string) (string, error) {
	return CreateDPoPJWTWithNonce(key, method, targetURL, "")
}

// CreateDPoPJWTWithNonce creates a DPoP JWT for the given HTTP method and URL with optional nonce
func CreateDPoPJWTWithNonce(key *ecdsa.PrivateKey, method, targetURL, nonce string) (string, error) {
	// Parse the URL to get the scheme, host, and path (no query or fragment)
	u, err := url.Parse(targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid target URL: %w", err)
	}
	
	// HTU should be scheme + host + path (no query or fragment)
	htu := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
	
	// Create the key pair wrapper to get JWK
	keyPair := &DPoPKeyPair{PrivateKey: key}
	
	// Create header
	header := DPoPJWTHeader{
		Typ: "dpop+jwt",
		Alg: "ES256",
		JWK: keyPair.DPoPPublicJWK(),
	}
	
	// Generate random JTI (nonce)
	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		return "", fmt.Errorf("failed to generate JTI: %w", err)
	}
	jti := base64.RawURLEncoding.EncodeToString(jtiBytes)
	
	// Create payload
	payload := DPoPJWTPayload{
		JTI:   jti,
		HTM:   method,
		HTU:   htu,
		IAT:   time.Now().Unix(),
		Nonce: nonce,
	}
	
	// Encode header and payload
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}
	
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)
	
	// Create signing input
	signingInput := headerEncoded + "." + payloadEncoded
	
	// Sign
	hash := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign DPoP JWT: %w", err)
	}
	
	// Encode signature in IEEE P1363 format (fixed 32+32 bytes for P-256)
	signature := make([]byte, 64) // 32 bytes for r + 32 bytes for s
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	
	// Pad r to 32 bytes (copy to end to preserve leading zeros)
	copy(signature[32-len(rBytes):32], rBytes)
	// Pad s to 32 bytes
	copy(signature[64-len(sBytes):64], sBytes)
	
	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)
	
	return signingInput + "." + signatureEncoded, nil
}

// CalculateJWKThumbprint calculates the SHA-256 thumbprint of a JWK
// This should match the 'jkt' claim in the access token
func CalculateJWKThumbprint(jwk map[string]interface{}) (string, error) {
	// Create a canonical JSON representation of the JWK
	// Per RFC 7638, only include the required fields in alphabetical order
	canonical := map[string]interface{}{
		"crv": jwk["crv"],
		"kty": jwk["kty"],
		"x":   jwk["x"],
		"y":   jwk["y"],
	}
	
	// Marshal to JSON (Go's json.Marshal produces consistent output)
	jsonBytes, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("failed to marshal canonical JWK: %w", err)
	}
	
	// Calculate SHA-256 hash
	hash := sha256.Sum256(jsonBytes)
	
	// Return base64url-encoded thumbprint
	return base64.RawURLEncoding.EncodeToString(hash[:]), nil
}

// GetJWKThumbprint returns the JWK thumbprint for this DPoP key
func (k *DPoPKeyPair) GetJWKThumbprint() (string, error) {
	jwk := k.DPoPPublicJWK()
	return CalculateJWKThumbprint(jwk)
}
