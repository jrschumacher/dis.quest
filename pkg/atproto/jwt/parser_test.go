package jwt

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func TestExtractDID(t *testing.T) {
	// Create a test token
	token := jwt.New()
	token.Set(jwt.SubjectKey, "did:plc:testuser123")
	token.Set(jwt.IssuerKey, "https://test.pds.example.com")
	token.Set(jwt.IssuedAtKey, time.Now())
	token.Set(jwt.ExpirationKey, time.Now().Add(time.Hour))
	token.Set("scope", "atproto")
	
	// Sign with test key
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, key))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	
	// Test ExtractDID
	did, err := ExtractDID(string(signed))
	if err != nil {
		t.Fatalf("Failed to extract DID: %v", err)
	}
	
	if did != "did:plc:testuser123" {
		t.Errorf("Expected DID 'did:plc:testuser123', got '%s'", did)
	}
}

func TestParseClaims(t *testing.T) {
	// Create a test token with all claims
	token := jwt.New()
	token.Set(jwt.SubjectKey, "did:plc:testuser123")
	token.Set(jwt.IssuerKey, "https://test.pds.example.com")
	token.Set(jwt.AudienceKey, []string{"https://app.example.com"})
	token.Set(jwt.IssuedAtKey, time.Now())
	token.Set(jwt.ExpirationKey, time.Now().Add(time.Hour))
	token.Set("scope", "atproto transition:generic")
	
	// Sign with test key
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, key))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	
	// Test ParseClaims
	claims, err := ParseClaims(string(signed))
	if err != nil {
		t.Fatalf("Failed to parse claims: %v", err)
	}
	
	// Verify claims
	if claims.Subject != "did:plc:testuser123" {
		t.Errorf("Expected subject 'did:plc:testuser123', got '%s'", claims.Subject)
	}
	
	if claims.Issuer != "https://test.pds.example.com" {
		t.Errorf("Expected issuer 'https://test.pds.example.com', got '%s'", claims.Issuer)
	}
	
	if len(claims.Audience) != 1 || claims.Audience[0] != "https://app.example.com" {
		t.Errorf("Expected audience ['https://app.example.com'], got %v", claims.Audience)
	}
	
	if claims.Scope != "atproto transition:generic" {
		t.Errorf("Expected scope 'atproto transition:generic', got '%s'", claims.Scope)
	}
}

func TestValidateWithKeySet(t *testing.T) {
	// Generate test key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Create JWK from public key
	jwkKey, err := jwk.FromRaw(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to create JWK: %v", err)
	}
	jwkKey.Set(jwk.AlgorithmKey, jwa.ES256)
	jwkKey.Set(jwk.KeyIDKey, "test-key-1")
	
	// Create key set
	keySet := jwk.NewSet()
	keySet.AddKey(jwkKey)
	
	// Create and sign token
	token := jwt.New()
	token.Set(jwt.SubjectKey, "did:plc:testuser123")
	token.Set(jwt.IssuerKey, "https://test.pds.example.com")
	token.Set(jwt.IssuedAtKey, time.Now())
	token.Set(jwt.ExpirationKey, time.Now().Add(time.Hour))
	token.Set("scope", "atproto")
	
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, privateKey))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	
	// Test validation
	claims, err := ValidateWithKeySet(string(signed), keySet)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	
	if claims.Subject != "did:plc:testuser123" {
		t.Errorf("Expected subject 'did:plc:testuser123', got '%s'", claims.Subject)
	}
}

func TestIsExpired(t *testing.T) {
	// Create expired token
	token := jwt.New()
	token.Set(jwt.SubjectKey, "did:plc:testuser123")
	token.Set(jwt.ExpirationKey, time.Now().Add(-time.Hour)) // Expired 1 hour ago
	
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signed, _ := jwt.Sign(token, jwt.WithKey(jwa.ES256, key))
	
	expired, err := IsExpired(string(signed))
	if err != nil {
		t.Fatalf("Failed to check expiry: %v", err)
	}
	
	if !expired {
		t.Error("Expected token to be expired")
	}
	
	// Create valid token
	token2 := jwt.New()
	token2.Set(jwt.SubjectKey, "did:plc:testuser123")
	token2.Set(jwt.ExpirationKey, time.Now().Add(time.Hour)) // Expires in 1 hour
	
	signed2, _ := jwt.Sign(token2, jwt.WithKey(jwa.ES256, key))
	
	expired2, err := IsExpired(string(signed2))
	if err != nil {
		t.Fatalf("Failed to check expiry: %v", err)
	}
	
	if expired2 {
		t.Error("Expected token to be valid")
	}
}

func TestTimeUntilExpiry(t *testing.T) {
	// Create token expiring in 1 hour
	token := jwt.New()
	token.Set(jwt.SubjectKey, "did:plc:testuser123")
	expiry := time.Now().Add(time.Hour)
	token.Set(jwt.ExpirationKey, expiry)
	
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signed, _ := jwt.Sign(token, jwt.WithKey(jwa.ES256, key))
	
	duration, err := TimeUntilExpiry(string(signed))
	if err != nil {
		t.Fatalf("Failed to get time until expiry: %v", err)
	}
	
	// Should be approximately 1 hour (allow 1 second tolerance for test execution)
	if duration < 59*time.Minute || duration > 61*time.Minute {
		t.Errorf("Expected duration around 1 hour, got %v", duration)
	}
}

// TestValidate would require a mock HTTP server or integration test
// since it fetches JWKS from a remote endpoint
func TestValidate(t *testing.T) {
	t.Skip("Skipping integration test that requires remote JWKS endpoint")
	
	// In a real test, you would:
	// 1. Set up a mock HTTP server serving JWKS
	// 2. Create a token signed with the corresponding private key
	// 3. Call Validate() which would fetch JWKS from the mock server
	// 4. Verify the claims
}