package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/config"
	"golang.org/x/oauth2"
)

// PKCE utilities
func GeneratePKCE() (codeVerifier, codeChallenge string, err error) {
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	codeVerifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	codeChallenge = base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return
}

// OAuth2 config for Bluesky/ATProto (new OAuth2 flow)
func OAuth2Config(provider string, cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.OAuthClientID,
		ClientSecret: "", // Not required for public clients
		RedirectURL:  cfg.OAuthRedirectURL,
		Scopes:       []string{"atproto", "transition:generic"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  provider + "/oauth/authorize",
			TokenURL: provider + "/oauth/token",
		},
	}
}

// Session utilities
const (
	sessionCookieName      = "dsq_session"
	refreshTokenCookieName = "dsq_refresh"
)

func SetSessionCookieWithEnv(w http.ResponseWriter, accessToken string, refreshToken []string, isDev bool) {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
	})
	if len(refreshToken) > 0 && refreshToken[0] != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     refreshTokenCookieName,
			Value:    refreshToken[0],
			Path:     "/",
			HttpOnly: true,
			Secure:   secure,
		})
	}
}

// Update SetSessionCookie to call the new function for backward compatibility
func SetSessionCookie(w http.ResponseWriter, accessToken string, refreshToken ...string) {
	// Default to production (secure) if not using the new function
	SetSessionCookieWithEnv(w, accessToken, refreshToken, false)
}

func ClearSessionCookieWithEnv(w http.ResponseWriter, isDev bool) {
	secure := !isDev
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	// Default to production (secure) if not using the new function
	ClearSessionCookieWithEnv(w, false)
}

func GetSessionCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func GetRefreshTokenCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// Exchange code for token
func ExchangeCodeForToken(ctx context.Context, provider, code, codeVerifier string, cfg *config.Config) (*oauth2.Token, error) {
	conf := OAuth2Config(provider, cfg)
	return conf.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
}
