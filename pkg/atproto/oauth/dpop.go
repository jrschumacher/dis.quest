// Package oauth provides DPoP (Demonstration of Proof-of-Possession) utilities for ATProtocol
package oauth

import (
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

// DPoPKeyPair holds an ECDSA P-256 keypair for DPoP
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

// EncodeToPEM encodes the private key as PEM for storage
func (k *DPoPKeyPair) EncodeToPEM() (string, error) {
	b, err := x509.MarshalECPrivateKey(k.PrivateKey)
	if err != nil {
		return "", err
	}
	pemBlock := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	return base64.RawURLEncoding.EncodeToString(pemBlock), nil
}

// DecodeFromPEM decodes a PEM-encoded private key
func DecodeFromPEM(pemStr string) (*DPoPKeyPair, error) {
	pemBytes, err := base64.RawURLEncoding.DecodeString(pemStr)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM block")
	}
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return &DPoPKeyPair{PrivateKey: key}, nil
}

// PublicJWK returns the public key as a JWK map (for DPoP JWT header)
func (k *DPoPKeyPair) PublicJWK() map[string]interface{} {
	pub := k.PrivateKey.PublicKey
	
	// Ensure 32 bytes for P-256 coordinates
	xBytes := pub.X.Bytes()
	if len(xBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(xBytes):], xBytes)
		xBytes = padded
	}
	
	yBytes := pub.Y.Bytes()
	if len(yBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(yBytes):], yBytes)
		yBytes = padded
	}
	
	return map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   base64.RawURLEncoding.EncodeToString(xBytes),
		"y":   base64.RawURLEncoding.EncodeToString(yBytes),
		"alg": "ES256",
		"use": "sig",
	}
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
	Ath   string `json:"ath,omitempty"` // Access token hash (base64url(SHA256(access_token)))
}

// CreateDPoPJWT creates a DPoP JWT for the given HTTP method and URL
func (k *DPoPKeyPair) CreateDPoPJWT(method, targetURL string) (string, error) {
	return k.CreateDPoPJWTWithNonce(method, targetURL, "")
}

// CreateDPoPJWTWithNonce creates a DPoP JWT for the given HTTP method and URL with optional nonce
func (k *DPoPKeyPair) CreateDPoPJWTWithNonce(method, targetURL, nonce string) (string, error) {
	return k.CreateDPoPJWTWithAccessToken(method, targetURL, nonce, "")
}

// CreateDPoPJWTWithAccessToken creates a DPoP JWT with access token hash (ath claim)
func (k *DPoPKeyPair) CreateDPoPJWTWithAccessToken(method, targetURL, nonce, accessToken string) (string, error) {
	// Parse the URL to get the scheme, host, and path (no query or fragment)
	u, err := url.Parse(targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid target URL: %w", err)
	}

	// HTU should be scheme + host + path (no query or fragment)
	htu := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)

	// Create header
	header := DPoPJWTHeader{
		Typ: "dpop+jwt",
		Alg: "ES256",
		JWK: k.PublicJWK(),
	}

	// Generate random JTI (nonce)
	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		return "", fmt.Errorf("failed to generate JTI: %w", err)
	}
	jti := base64.RawURLEncoding.EncodeToString(jtiBytes)

	// Calculate access token hash if provided
	var ath string
	if accessToken != "" {
		hash := sha256.Sum256([]byte(accessToken))
		ath = base64.RawURLEncoding.EncodeToString(hash[:])
	}

	// Create payload
	payload := DPoPJWTPayload{
		JTI:   jti,
		HTM:   method,
		HTU:   htu,
		IAT:   time.Now().Unix(),
		Nonce: nonce,
		Ath:   ath,
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
	r, s, err := ecdsa.Sign(rand.Reader, k.PrivateKey, hash[:])
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
func (k *DPoPKeyPair) CalculateJWKThumbprint() (string, error) {
	jwk := k.PublicJWK()
	
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

// Cookie names for DPoP key and nonce storage
const (
	DPoPKeyCookieName           = "dpop_key"
	DPoPNonceCookieName         = "dpop_nonce"
	AuthServerIssuerCookieName  = "auth_server_issuer"
)

// SetDPoPKeyCookie stores the DPoP private key in a secure, HttpOnly cookie
func SetDPoPKeyCookie(w http.ResponseWriter, key *ecdsa.PrivateKey, isDev bool) error {
	keyPair := &DPoPKeyPair{PrivateKey: key}
	pemStr, err := keyPair.EncodeToPEM()
	if err != nil {
		return err
	}
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     DPoPKeyCookieName,
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
	cookie, err := r.Cookie(DPoPKeyCookieName)
	if err != nil {
		return nil, err
	}
	keyPair, err := DecodeFromPEM(cookie.Value)
	if err != nil {
		return nil, err
	}
	return keyPair.PrivateKey, nil
}

// ClearDPoPKeyCookie clears the DPoP key cookie
func ClearDPoPKeyCookie(w http.ResponseWriter, isDev bool) {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     DPoPKeyCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// SetDPoPNonceCookie stores the DPoP nonce in a secure, HttpOnly cookie
func SetDPoPNonceCookie(w http.ResponseWriter, nonce string, isDev bool) error {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     DPoPNonceCookieName,
		Value:    nonce,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes for OAuth flow
	})
	return nil
}

// GetDPoPNonceFromCookie retrieves the DPoP nonce from the cookie
func GetDPoPNonceFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(DPoPNonceCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ClearDPoPNonceCookie clears the DPoP nonce cookie
func ClearDPoPNonceCookie(w http.ResponseWriter, isDev bool) {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     DPoPNonceCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// SetAuthServerIssuerCookie stores the authorization server issuer in a cookie
func SetAuthServerIssuerCookie(w http.ResponseWriter, issuer string, isDev bool) error {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     AuthServerIssuerCookieName,
		Value:    issuer,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes for OAuth flow
	})
	return nil
}

// GetAuthServerIssuerFromCookie retrieves the authorization server issuer from the cookie
func GetAuthServerIssuerFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(AuthServerIssuerCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ClearAuthServerIssuerCookie clears the authorization server issuer cookie
func ClearAuthServerIssuerCookie(w http.ResponseWriter, isDev bool) {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     AuthServerIssuerCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}