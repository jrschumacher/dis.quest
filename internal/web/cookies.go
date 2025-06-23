// Package web provides HTTP-specific utilities for ATProtocol web applications
package web

import (
	"net/http"

	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
)

// Session cookie management
const (
	sessionCookieName      = "dsq_session"
	refreshTokenCookieName = "dsq_refresh"
)

// SetSessionCookieWithEnv sets session cookies with environment-specific security settings
func SetSessionCookieWithEnv(w http.ResponseWriter, accessToken string, refreshToken []string, isDev bool) {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600, // 1 hour
	})
	if len(refreshToken) > 0 && refreshToken[0] != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     refreshTokenCookieName,
			Value:    refreshToken[0],
			Path:     "/",
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400, // 24 hours
		})
	}
}

// SetSessionCookie sets session cookies with default production security settings
func SetSessionCookie(w http.ResponseWriter, accessToken string, refreshToken ...string) {
	// Default to production (secure) if not using the new function
	SetSessionCookieWithEnv(w, accessToken, refreshToken, false)
}

// ClearSessionCookieWithEnv clears session cookies with environment-specific settings
func ClearSessionCookieWithEnv(w http.ResponseWriter, isDev bool) {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	// Also clear the DPoP key cookie to prevent key reuse with new tokens
	oauth.ClearDPoPKeyCookie(w, isDev)
}

// ClearSessionCookie clears session cookies with default production settings
func ClearSessionCookie(w http.ResponseWriter) {
	// Default to production (secure) if not using the new function
	ClearSessionCookieWithEnv(w, false)
}

// GetSessionCookie retrieves the session cookie value from the request
func GetSessionCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetRefreshTokenCookie retrieves the refresh token cookie value from the request
func GetRefreshTokenCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}