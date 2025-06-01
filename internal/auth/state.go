package auth

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateStateToken generates a random string suitable for use as an OAuth2 state parameter.
func GenerateStateToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		// fallback: not cryptographically secure, but avoids panic
		return base64.RawURLEncoding.EncodeToString([]byte("fallback_state_token"))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
