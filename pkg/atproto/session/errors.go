package session

import "errors"

// Session management errors
var (
	// Storage errors
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrInvalidSessionKey = errors.New("invalid session key")
	ErrStorageFailed   = errors.New("storage operation failed")
	
	// Session errors
	ErrInvalidSession  = errors.New("invalid session")
	ErrTokenRefreshFailed = errors.New("token refresh failed")
	ErrTokenInvalid    = errors.New("invalid token")
	ErrTokenExpired    = errors.New("token expired")
	
	// Encryption errors
	ErrEncryptionFailed = errors.New("encryption failed")
	ErrDecryptionFailed = errors.New("decryption failed")
	ErrInvalidEncryptionKey = errors.New("invalid encryption key")
	
	// Configuration errors
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrMissingProvider  = errors.New("missing OAuth provider")
	ErrMissingStorage   = errors.New("missing session storage")
	
	// HTTP context errors
	ErrMissingHTTPContext = errors.New("HTTP context not found")
	ErrMissingRequest     = errors.New("HTTP request not found in context")
	ErrMissingResponse    = errors.New("HTTP response writer not found in context")
)