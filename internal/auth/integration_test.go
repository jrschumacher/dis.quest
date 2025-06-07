package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jrschumacher/dis.quest/internal/config"
	"golang.org/x/oauth2"
)

// TestOAuth2Components_Unit tests individual OAuth2 components (Phase 0 scope)
func TestOAuth2Components_Unit(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		AppEnv:           config.EnvTest,
		OAuthClientID:    "test-client-id",
		OAuthRedirectURL: "http://localhost:3000/auth/callback",
		PublicDomain:     "http://localhost:3000",
		AppName:          "dis.quest Test",
	}

	// Test PKCE generation
	codeVerifier, codeChallenge, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("Failed to generate PKCE: %v", err)
	}

	if codeVerifier == "" {
		t.Error("Code verifier should not be empty")
	}
	if codeChallenge == "" {
		t.Error("Code challenge should not be empty")
	}
	if len(codeVerifier) < 43 || len(codeVerifier) > 128 {
		t.Errorf("Code verifier length should be between 43-128 characters, got %d", len(codeVerifier))
	}

	t.Logf("Generated PKCE - Verifier: %s, Challenge: %s", codeVerifier, codeChallenge)

	// Test OAuth2 config creation
	provider := "https://bsky.social"
	oauth2Config := OAuth2Config(provider, cfg)
	if oauth2Config == nil {
		t.Fatal("OAuth2 config should not be nil")
	}

	if oauth2Config.ClientID != cfg.OAuthClientID {
		t.Errorf("Expected client ID %s, got %s", cfg.OAuthClientID, oauth2Config.ClientID)
	}
	if oauth2Config.RedirectURL != cfg.OAuthRedirectURL {
		t.Errorf("Expected redirect URL %s, got %s", cfg.OAuthRedirectURL, oauth2Config.RedirectURL)
	}
	if !strings.Contains(oauth2Config.Endpoint.AuthURL, provider) {
		t.Errorf("Auth URL should contain provider %s, got %s", provider, oauth2Config.Endpoint.AuthURL)
	}
	if !strings.Contains(oauth2Config.Endpoint.TokenURL, provider) {
		t.Errorf("Token URL should contain provider %s, got %s", provider, oauth2Config.Endpoint.TokenURL)
	}

	// Test authorization URL generation
	state := "test-state-token"
	authURL := oauth2Config.AuthCodeURL(state, 
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	parsedURL, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("Failed to parse auth URL: %v", err)
	}

	// Verify auth URL parameters
	query := parsedURL.Query()
	if query.Get("client_id") != cfg.OAuthClientID {
		t.Errorf("Auth URL should contain client_id %s", cfg.OAuthClientID)
	}
	if query.Get("redirect_uri") != cfg.OAuthRedirectURL {
		t.Errorf("Auth URL should contain redirect_uri %s", cfg.OAuthRedirectURL)
	}
	if query.Get("state") != state {
		t.Errorf("Auth URL should contain state %s", state)
	}
	if query.Get("code_challenge") != codeChallenge {
		t.Errorf("Auth URL should contain code_challenge %s", codeChallenge)
	}
	if query.Get("code_challenge_method") != "S256" {
		t.Error("Auth URL should contain code_challenge_method S256")
	}
	if query.Get("response_type") != "code" {
		t.Error("Auth URL should contain response_type code")
	}

	t.Logf("Generated auth URL: %s", authURL)
}

// TestSessionCookieFlow_Integration tests session cookie management
func TestSessionCookieFlow_Integration(t *testing.T) {
	// Test session cookie setting and retrieval
	accessToken := "test-access-token"
	refreshTokens := []string{"test-refresh-token"}

	// Test development mode (insecure cookies)
	t.Run("Development Mode Cookies", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		SetSessionCookieWithEnv(recorder, accessToken, refreshTokens, true) // isDev = true

		cookies := recorder.Result().Cookies()
		if len(cookies) < 2 {
			t.Fatalf("Expected at least 2 cookies, got %d", len(cookies))
		}

		var sessionCookie, refreshCookie *http.Cookie
		for _, cookie := range cookies {
			switch cookie.Name {
			case "dsq_session":
				sessionCookie = cookie
			case "dsq_refresh":
				refreshCookie = cookie
			}
		}

		if sessionCookie == nil {
			t.Fatal("Session cookie not found")
		}
		if refreshCookie == nil {
			t.Fatal("Refresh cookie not found")
		}

		// Verify session cookie properties
		if sessionCookie.Value != accessToken {
			t.Errorf("Expected session cookie value %s, got %s", accessToken, sessionCookie.Value)
		}
		if sessionCookie.Secure {
			t.Error("Session cookie should not be secure in development mode")
		}
		if !sessionCookie.HttpOnly {
			t.Error("Session cookie should be HttpOnly")
		}
		if sessionCookie.Path != "/" {
			t.Errorf("Session cookie path should be /, got %s", sessionCookie.Path)
		}

		// Verify refresh cookie properties
		if refreshCookie.Value != refreshTokens[0] {
			t.Errorf("Expected refresh cookie value %s, got %s", refreshTokens[0], refreshCookie.Value)
		}
		if refreshCookie.Secure {
			t.Error("Refresh cookie should not be secure in development mode")
		}
		if !refreshCookie.HttpOnly {
			t.Error("Refresh cookie should be HttpOnly")
		}

		// Test cookie retrieval
		request := &http.Request{Header: http.Header{}}
		for _, cookie := range cookies {
			request.AddCookie(cookie)
		}

		retrievedAccessToken, err := GetSessionCookie(request)
		if err != nil {
			t.Fatalf("Failed to get session cookie: %v", err)
		}
		if retrievedAccessToken != accessToken {
			t.Errorf("Expected retrieved access token %s, got %s", accessToken, retrievedAccessToken)
		}

		retrievedRefreshToken, err := GetRefreshTokenCookie(request)
		if err != nil {
			t.Fatalf("Failed to get refresh cookie: %v", err)
		}
		if retrievedRefreshToken != refreshTokens[0] {
			t.Errorf("Expected retrieved refresh token %s, got %s", refreshTokens[0], retrievedRefreshToken)
		}
	})

	// Test production mode (secure cookies)
	t.Run("Production Mode Cookies", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		SetSessionCookieWithEnv(recorder, accessToken, refreshTokens, false) // isDev = false

		cookies := recorder.Result().Cookies()
		if len(cookies) < 2 {
			t.Fatalf("Expected at least 2 cookies, got %d", len(cookies))
		}

		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "dsq_session" {
				sessionCookie = cookie
				break
			}
		}

		if sessionCookie == nil {
			t.Fatal("Session cookie not found")
		}

		// Verify secure flag is set in production
		if !sessionCookie.Secure {
			t.Error("Session cookie should be secure in production mode")
		}
	})

	// Test session clearing
	t.Run("Session Clearing", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ClearSessionCookieWithEnv(recorder, true) // isDev = true

		cookies := recorder.Result().Cookies()
		if len(cookies) < 2 {
			t.Fatalf("Expected at least 2 cookies for clearing, got %d", len(cookies))
		}

		for _, cookie := range cookies {
			if cookie.Name == "dsq_session" || cookie.Name == "dsq_refresh" {
				if cookie.Value != "" {
					t.Errorf("Cookie %s should have empty value when clearing, got %s", cookie.Name, cookie.Value)
				}
				if cookie.MaxAge != -1 {
					t.Errorf("Cookie %s should have MaxAge -1 when clearing, got %d", cookie.Name, cookie.MaxAge)
				}
			}
		}
	})
}

// TestBackwardCompatibility_Integration tests backward compatibility functions
func TestBackwardCompatibility_Integration(t *testing.T) {
	accessToken := "test-access-token"
	refreshTokens := []string{"test-refresh-token"}

	// Test backward compatible SetSessionCookie (should default to production mode)
	recorder := httptest.NewRecorder()
	SetSessionCookie(recorder, accessToken, refreshTokens...)

	cookies := recorder.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "dsq_session" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Session cookie not found")
	}

	// Should default to secure (production mode)
	if !sessionCookie.Secure {
		t.Error("Backward compatible SetSessionCookie should default to secure cookies")
	}

	// Test backward compatible ClearSessionCookie (should default to production mode)
	clearRecorder := httptest.NewRecorder()
	ClearSessionCookie(clearRecorder)

	clearCookies := clearRecorder.Result().Cookies()
	var clearSessionCookie *http.Cookie
	for _, cookie := range clearCookies {
		if cookie.Name == "dsq_session" {
			clearSessionCookie = cookie
			break
		}
	}

	if clearSessionCookie == nil {
		t.Fatal("Clear session cookie not found")
	}

	// Should default to secure (production mode)
	if !clearSessionCookie.Secure {
		t.Error("Backward compatible ClearSessionCookie should default to secure cookies")
	}
}

// TestCookieRetrievalErrors_Integration tests error conditions in cookie retrieval
func TestCookieRetrievalErrors_Integration(t *testing.T) {
	// Test getting session cookie when not present
	request := &http.Request{Header: http.Header{}}
	
	_, err := GetSessionCookie(request)
	if err == nil {
		t.Error("Expected error when session cookie not present")
	}
	if !strings.Contains(err.Error(), "cookie not present") {
		t.Errorf("Error should mention cookie not present, got: %v", err)
	}

	// Test getting refresh cookie when not present
	_, err = GetRefreshTokenCookie(request)
	if err == nil {
		t.Error("Expected error when refresh cookie not present")
	}
	if !strings.Contains(err.Error(), "cookie not present") {
		t.Errorf("Error should mention cookie not present, got: %v", err)
	}
}