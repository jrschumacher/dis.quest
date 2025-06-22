// Package jwt provides utilities for working with ATProtocol JWT tokens.
// This includes parsing, validation, and claim extraction for access tokens
// issued by ATProtocol Personal Data Servers (PDS).
package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// Claims represents the standard claims in an ATProtocol JWT token
type Claims struct {
	// Standard JWT claims
	Issuer   string    `json:"iss"`   // PDS endpoint that issued the token
	Subject  string    `json:"sub"`   // User's DID
	Audience []string  `json:"aud"`   // Token audience
	Expiry   time.Time `json:"exp"`   // Token expiration
	IssuedAt time.Time `json:"iat"`   // Token issue time
	
	// ATProtocol specific claims
	Scope string `json:"scope"` // OAuth scope (e.g., "atproto")
}

// ExtractDID quickly extracts the user's DID from a token without verification.
// This is useful for middleware that needs the DID before full validation.
func ExtractDID(tokenString string) (string, error) {
	claims, err := ParseClaims(tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}
	
	if claims.Subject == "" {
		return "", fmt.Errorf("token missing subject (DID)")
	}
	
	return claims.Subject, nil
}

// ParseClaims extracts claims from a token without verification.
// WARNING: Only use this when you don't need cryptographic verification,
// such as for extracting the issuer to fetch JWKS, or in development.
func ParseClaims(tokenString string) (*Claims, error) {
	// Parse without verification
	token, err := jwt.Parse([]byte(tokenString), jwt.WithVerify(false), jwt.WithValidate(false))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}
	
	claims := &Claims{
		Issuer:  token.Issuer(),
		Subject: token.Subject(),
	}
	
	// Handle times
	if !token.Expiration().IsZero() {
		claims.Expiry = token.Expiration()
	}
	if !token.IssuedAt().IsZero() {
		claims.IssuedAt = token.IssuedAt()
	}
	
	// Handle audience (can be string or []string in JWT)
	if aud := token.Audience(); len(aud) > 0 {
		claims.Audience = aud
	}
	
	// Extract ATProtocol scope
	if scopeClaim, ok := token.Get("scope"); ok {
		if scope, ok := scopeClaim.(string); ok {
			claims.Scope = scope
		}
	}
	
	return claims, nil
}

// Validate performs full JWT validation including signature verification.
// It automatically fetches the JWKS from the token issuer's well-known endpoint.
func Validate(ctx context.Context, tokenString string) (*Claims, error) {
	// First parse without verification to get issuer
	unverifiedClaims, err := ParseClaims(tokenString)
	if err != nil {
		return nil, err
	}
	
	if unverifiedClaims.Issuer == "" {
		return nil, fmt.Errorf("token missing issuer")
	}
	
	// Fetch JWKS from issuer
	keySet, err := FetchJWKS(ctx, unverifiedClaims.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	
	// Validate with fetched keys
	return ValidateWithKeySet(tokenString, keySet)
}

// ValidateWithKeySet validates a JWT using a provided JWK Set.
// This is useful when you want to cache JWKS or provide your own keys.
func ValidateWithKeySet(tokenString string, keySet jwk.Set) (*Claims, error) {
	// Parse and verify the JWT
	token, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(keySet))
	if err != nil {
		return nil, fmt.Errorf("failed to verify JWT: %w", err)
	}
	
	// Token is valid, extract claims
	claims := &Claims{
		Issuer:   token.Issuer(),
		Subject:  token.Subject(),
		Expiry:   token.Expiration(),
		IssuedAt: token.IssuedAt(),
	}
	
	// Handle audience
	if aud := token.Audience(); len(aud) > 0 {
		claims.Audience = aud
	}
	
	// Extract scope
	if scopeClaim, ok := token.Get("scope"); ok {
		if scope, ok := scopeClaim.(string); ok {
			claims.Scope = scope
		}
	}
	
	return claims, nil
}

// FetchJWKS fetches the JSON Web Key Set from an issuer's well-known endpoint.
// ATProtocol PDS instances expose their public keys at /.well-known/jwks.json
func FetchJWKS(ctx context.Context, issuer string) (jwk.Set, error) {
	jwksURL := fmt.Sprintf("%s/.well-known/jwks.json", issuer)
	
	// Use jwx library's built-in fetcher with context
	set, err := jwk.Fetch(ctx, jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS from %s: %w", jwksURL, err)
	}
	
	return set, nil
}

// IsExpired checks if the token is expired based on its claims.
// This does not perform signature verification.
func IsExpired(tokenString string) (bool, error) {
	claims, err := ParseClaims(tokenString)
	if err != nil {
		return false, err
	}
	
	if claims.Expiry.IsZero() {
		return false, fmt.Errorf("token has no expiry claim")
	}
	
	return time.Now().After(claims.Expiry), nil
}

// TimeUntilExpiry returns the duration until the token expires.
// Returns a negative duration if already expired.
func TimeUntilExpiry(tokenString string) (time.Duration, error) {
	claims, err := ParseClaims(tokenString)
	if err != nil {
		return 0, err
	}
	
	if claims.Expiry.IsZero() {
		return 0, fmt.Errorf("token has no expiry claim")
	}
	
	return time.Until(claims.Expiry), nil
}