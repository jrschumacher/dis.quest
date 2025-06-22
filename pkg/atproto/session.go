package atproto

import "crypto/ecdsa"

// Additional session methods that may be needed for advanced usage

// GetAccessToken returns the current access token (use carefully)
func (s *Session) GetAccessToken() string {
	return s.accessToken
}

// GetDPoPKey returns the DPoP private key (use carefully)
func (s *Session) GetDPoPKey() *ecdsa.PrivateKey {
	return s.dpopKey
}

// GetRefreshToken returns the refresh token (use carefully)
func (s *Session) GetRefreshToken() string {
	return s.refreshToken
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	// Simple check - in production you'd want to check actual expiration time
	return s.expiresIn <= 0
}