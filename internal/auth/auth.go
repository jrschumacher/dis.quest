package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"os"

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

// OAuth2 config for Bluesky/ATProto
func OAuth2Config(provider string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URI"),
		Scopes:       []string{"openid", "offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  provider + "/xrpc/com.atproto.server.requestOAuth2Token",  // Example, update as needed
			TokenURL: provider + "/xrpc/com.atproto.server.exchangeOAuth2Token", // Example, update as needed
		},
	}
}

// Session utilities
const (
	sessionCookieName      = "dsq_session"
	refreshTokenCookieName = "dsq_refresh"
)

func SetSessionCookie(w http.ResponseWriter, accessToken string, refreshToken ...string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	})
	if len(refreshToken) > 0 && refreshToken[0] != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     refreshTokenCookieName,
			Value:    refreshToken[0],
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // Set to true in production with HTTPS
		})
	}
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
	})
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
func ExchangeCodeForToken(ctx context.Context, provider, code, codeVerifier string) (*oauth2.Token, error) {
	conf := OAuth2Config(provider)
	return conf.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
}
