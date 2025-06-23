// Package jwt provides utilities for working with JSON Web Tokens in ATProtocol applications.
//
// This package offers both quick extraction methods for middleware use cases and
// full validation with JWKS verification for secure token validation.
//
// # Quick Start
//
// Extract a user's DID without verification (useful for logging/middleware):
//
//	did, err := jwt.ExtractDID(accessToken)
//	if err != nil {
//	    log.Printf("Invalid token: %v", err)
//	}
//
// Parse all claims without verification:
//
//	claims, err := jwt.ParseClaims(accessToken)
//	if err != nil {
//	    log.Printf("Failed to parse token: %v", err)
//	}
//	fmt.Printf("User: %s from PDS: %s\n", claims.Subject, claims.Issuer)
//
// # Full Validation
//
// Validate a token with automatic JWKS fetching:
//
//	claims, err := jwt.Validate(ctx, accessToken)
//	if err != nil {
//	    // Token is invalid or verification failed
//	    return fmt.Errorf("unauthorized: %w", err)
//	}
//	// Token is valid, claims.Subject contains the user's DID
//
// # Token Expiry
//
// Check if a token is expired:
//
//	if expired, _ := jwt.IsExpired(token); expired {
//	    // Refresh the token
//	}
//
// Get time until expiry:
//
//	duration, _ := jwt.TimeUntilExpiry(token)
//	if duration < 5*time.Minute {
//	    // Refresh soon
//	}
//
// # Security Considerations
//
// - ParseClaims and ExtractDID do NOT verify signatures. Only use these for
//   non-security-critical operations like logging or extracting the issuer.
// - Always use Validate() or ValidateWithKeySet() when making authorization decisions.
// - JWKS are fetched over HTTPS from the issuer's well-known endpoint.
// - Consider caching JWKS to reduce network calls (not implemented in this package).
package jwt