package auth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
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
		return nil, errors.New("invalid credentials or failed to create session")
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
		return nil, errors.New("invalid PEM block")
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
		// Optionally: Short expiry, e.g. 10 min
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
