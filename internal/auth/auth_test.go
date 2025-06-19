package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGeneratePKCE(t *testing.T) {
	verifier, challenge, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE error: %v", err)
	}
	if verifier == "" || challenge == "" {
		t.Fatalf("expected non-empty values")
	}
	if verifier == challenge {
		t.Fatalf("verifier and challenge should differ")
	}
}

func TestGenerateStateToken(t *testing.T) {
	token := GenerateStateToken()
	if token == "" {
		t.Fatalf("expected non-empty token")
	}
}

func TestDPoPKeyEncodeDecode(t *testing.T) {
	keypair, err := GenerateDPoPKeyPair()
	if err != nil {
		t.Fatalf("GenerateDPoPKeyPair error: %v", err)
	}
	pemStr, err := EncodeDPoPPrivateKeyToPEM(keypair.PrivateKey)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	decoded, err := DecodeDPoPPrivateKeyFromPEM(pemStr)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if decoded == nil || decoded.X.Cmp(keypair.PrivateKey.X) != 0 {
		t.Fatalf("decoded key mismatch")
	}
}

func TestDPoPKeyCookieRoundTrip(t *testing.T) {
	keypair, err := GenerateDPoPKeyPair()
	if err != nil {
		t.Fatalf("GenerateDPoPKeyPair error: %v", err)
	}
	rr := httptest.NewRecorder()
	if err := SetDPoPKeyCookie(rr, keypair.PrivateKey, true); err != nil {
		t.Fatalf("SetDPoPKeyCookie error: %v", err)
	}
	req := httptest.NewRequest("GET", "/", nil)
	for _, c := range rr.Result().Cookies() {
		req.AddCookie(c)
	}
	got, err := GetDPoPKeyFromCookie(req)
	if err != nil {
		t.Fatalf("GetDPoPKeyFromCookie error: %v", err)
	}
	if got == nil || got.X.Cmp(keypair.PrivateKey.X) != 0 {
		t.Fatalf("cookie roundtrip mismatch")
	}
}

func TestSessionCookieRoundTrip(t *testing.T) {
	rr := httptest.NewRecorder()
	SetSessionCookieWithEnv(rr, "access", []string{"refresh"}, true)
	req := httptest.NewRequest("GET", "/", nil)
	for _, c := range rr.Result().Cookies() {
		req.AddCookie(c)
	}
	access, err := GetSessionCookie(req)
	if err != nil {
		t.Fatalf("GetSessionCookie error: %v", err)
	}
	if access != "access" {
		t.Fatalf("expected access token, got %s", access)
	}
	refresh, err := GetRefreshTokenCookie(req)
	if err != nil {
		t.Fatalf("GetRefreshTokenCookie error: %v", err)
	}
	if refresh != "refresh" {
		t.Fatalf("expected refresh token, got %s", refresh)
	}
}

func TestGetSessionCookie_NotFound(t *testing.T) {
	// Test when no session cookie is present
	req := httptest.NewRequest("GET", "/", nil)
	_, err := GetSessionCookie(req)
	if err == nil {
		t.Fatal("expected error when no session cookie present")
	}
	if !errors.Is(err, http.ErrNoCookie) {
		t.Errorf("expected http.ErrNoCookie, got %v", err)
	}
}

func TestGetRefreshTokenCookie_NotFound(t *testing.T) {
	// Test when no refresh token cookie is present
	req := httptest.NewRequest("GET", "/", nil)
	_, err := GetRefreshTokenCookie(req)
	if err == nil {
		t.Fatal("expected error when no refresh token cookie present")
	}
	if !errors.Is(err, http.ErrNoCookie) {
		t.Errorf("expected http.ErrNoCookie, got %v", err)
	}
}

func TestDecodeDPoPPrivateKeyFromPEM_InvalidPEM(t *testing.T) {
	// Test with invalid PEM data
	_, err := DecodeDPoPPrivateKeyFromPEM("not-a-pem-block")
	if err == nil {
		t.Fatal("expected error with invalid PEM data")
	}
	if !errors.Is(err, ErrInvalidPEMBlock) {
		t.Errorf("expected ErrInvalidPEMBlock, got %v", err)
	}
}
