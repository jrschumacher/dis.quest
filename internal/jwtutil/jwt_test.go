package jwtutil

import (
	"context"
	"testing"
)

func TestExtractDIDFromJWT_InvalidToken(t *testing.T) {
	// Test with an invalid token format
	_, err := ExtractDIDFromJWT("invalid.jwt.token")
	if err == nil {
		t.Fatal("expected error with invalid token")
	}
	// The function should fail during parsing
}

func TestVerifyJWT_InvalidToken(t *testing.T) {
	// Test with an invalid token format
	_, err := VerifyJWT(context.Background(), "invalid.jwt.token")
	if err == nil {
		t.Fatal("expected error with invalid token")
	}
}
