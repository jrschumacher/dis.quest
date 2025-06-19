package auth

import "errors"

// Authentication and authorization errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials or failed to create session")
	ErrInvalidPEMBlock    = errors.New("invalid PEM block")
	ErrSessionNotFound    = errors.New("session not found")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInvalidToken       = errors.New("invalid token")
)