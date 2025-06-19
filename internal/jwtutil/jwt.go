// Package jwtutil provides utilities for working with JWT tokens in dis.quest
package jwtutil

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	// ErrMissingSubject is returned when the JWT is missing a subject (DID)
	ErrMissingSubject = fmt.Errorf("missing subject (DID) in token")
	// ErrMissingIssuer is returned when the JWT is missing an issuer
	ErrMissingIssuer = fmt.Errorf("missing issuer in token")
	// ErrInvalidToken is returned when the JWT is in an invalid format
	ErrInvalidToken = fmt.Errorf("invalid token format")
)

// JWTClaims represents the claims we care about from a JWT token
// (adapted from ATProto, but not limited to it)
type JWTClaims struct {
	Iss   string `json:"iss"`   // Issuer (PDS)
	Sub   string `json:"sub"`   // Subject (DID)
	Aud   string `json:"aud"`   // Audience
	Exp   int64  `json:"exp"`   // Expiry time
	Iat   int64  `json:"iat"`   // Issued at
	Scope string `json:"scope"` // Token scope
}

// ParseAndValidateJWT parses and validates a JWT token using the jwx library
func ParseAndValidateJWT(_ context.Context, tokenString string, keySet jwk.Set) (*JWTClaims, error) {
	// Parse and verify the JWT with the provided key set
	token, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(keySet))
	if err != nil {
		return nil, fmt.Errorf("failed to parse and verify JWT: %w", err)
	}

	// Extract claims into our struct
	claims := &JWTClaims{
		Iss: token.Issuer(),
		Sub: token.Subject(),
		Exp: token.Expiration().Unix(),
		Iat: token.IssuedAt().Unix(),
	}

	// Get audience (may be a string or []string)
	if aud := token.Audience(); len(aud) > 0 {
		claims.Aud = aud[0]
	}

	// Extract scope from private claims
	if scopeClaim, ok := token.Get("scope"); ok {
		if scope, ok := scopeClaim.(string); ok {
			claims.Scope = scope
		}
	}

	return claims, nil
}

// ParseJWTWithoutVerification extracts claims from a JWT without verification
// Note: This should only be used in development or for extracting issuer info to fetch keys
func ParseJWTWithoutVerification(tokenString string) (*JWTClaims, error) {
	// Parse JWT without verification
	token, err := jwt.Parse([]byte(tokenString), jwt.WithVerify(false), jwt.WithValidate(false))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	claims := &JWTClaims{
		Iss: token.Issuer(),
		Sub: token.Subject(),
	}

	if !token.Expiration().IsZero() {
		claims.Exp = token.Expiration().Unix()
	}
	if !token.IssuedAt().IsZero() {
		claims.Iat = token.IssuedAt().Unix()
	}

	// Get audience
	if aud := token.Audience(); len(aud) > 0 {
		claims.Aud = aud[0]
	}

	// Extract scope
	if scopeClaim, ok := token.Get("scope"); ok {
		if scope, ok := scopeClaim.(string); ok {
			claims.Scope = scope
		}
	}

	return claims, nil
}

// ExtractDIDFromJWT extracts the DID from a JWT token without full verification
// This is useful for getting the user DID before doing full validation
func ExtractDIDFromJWT(tokenString string) (string, error) {
	claims, err := ParseJWTWithoutVerification(tokenString)
	if err != nil {
		return "", err
	}

	if claims.Sub == "" {
		return "", fmt.Errorf("missing subject (DID)")
	}

	return claims.Sub, nil
}

// GetJWKSFromIssuer fetches the JWKS from an issuer's well-known endpoint
func GetJWKSFromIssuer(ctx context.Context, issuer string) (jwk.Set, error) {
	jwksURL := fmt.Sprintf("%s/.well-known/jwks.json", issuer)

	// Fetch the JWKS
	set, err := jwk.Fetch(ctx, jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS from %s: %w", jwksURL, err)
	}

	return set, nil
}

// VerifyJWT verifies a JWT token by fetching JWKS from the issuer
func VerifyJWT(ctx context.Context, tokenString string) (*JWTClaims, error) {
	// First, parse without verification to get the issuer
	unverifiedClaims, err := ParseJWTWithoutVerification(tokenString)
	if err != nil {
		return nil, err
	}

	if unverifiedClaims.Iss == "" {
		return nil, fmt.Errorf("missing issuer")
	}

	// Fetch JWKS from the issuer
	keySet, err := GetJWKSFromIssuer(ctx, unverifiedClaims.Iss)
	if err != nil {
		return nil, err
	}

	// Now parse and verify with the proper keys
	return ParseAndValidateJWT(ctx, tokenString, keySet)
}
