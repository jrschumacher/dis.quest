package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jrschumacher/dis.quest/internal/config"
)

// TestAuthHandlers_Unit tests individual HTTP authentication handlers (Phase 0 scope)
func TestAuthHandlers_Unit(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		AppEnv:           config.EnvTest,
		OAuthClientID:    "test-client-id",
		OAuthRedirectURL: "http://localhost:3000/auth/callback",
		PublicDomain:     "http://localhost:3000",
		AppName:          "dis.quest Test",
	}

	// Create test server
	mux := http.NewServeMux()
	RegisterRoutes(mux, "/auth", cfg)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test client metadata endpoint
	t.Run("ClientMetadata", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/auth/client-metadata.json")
		if err != nil {
			t.Fatalf("Failed to get client metadata: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		contentType := resp.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected content type application/json, got %s", contentType)
		}

		// Verify response contains expected JSON structure
		// (In a real test, you might parse and validate the JSON structure)
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		body := string(buf[:n])

		expectedFields := []string{
			"client_id",
			"client_name", 
			"client_uri",
			"redirect_uris",
			"grant_types",
			"response_types",
			"scope",
			"token_endpoint_auth_method",
			"application_type",
			"dpop_bound_access_tokens",
		}

		for _, field := range expectedFields {
			if !strings.Contains(body, field) {
				t.Errorf("Client metadata should contain field %s", field)
			}
		}

		t.Logf("Client metadata response: %s", body)
	})

	// Test login handler (GET request) - should return Method Not Allowed
	t.Run("LoginHandler_GET", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/auth/login")
		if err != nil {
			t.Fatalf("Failed to access login endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Login handler only accepts POST, so GET should return 405
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 (Method Not Allowed), got %d", resp.StatusCode)
		}
	})

	// Test login handler with incomplete form data (POST request) - should return Bad Request
	t.Run("LoginHandler_POST_Incomplete", func(t *testing.T) {
		// Prepare form data with missing password
		formData := url.Values{}
		formData.Set("handle", "test.bsky.social")
		
		resp, err := http.PostForm(server.URL + "/auth/login", formData)
		if err != nil {
			t.Fatalf("Failed to post to login endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Login handler should return 400 Bad Request when missing password
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 (Bad Request), got %d", resp.StatusCode)
		}
	})

	// Test login handler with complete form data (POST request) - will fail PDS discovery
	t.Run("LoginHandler_POST_Complete", func(t *testing.T) {
		// Prepare form data with both handle and password
		formData := url.Values{}
		formData.Set("handle", "test.bsky.social")
		formData.Set("password", "test-password")
		
		resp, err := http.PostForm(server.URL + "/auth/login", formData)
		if err != nil {
			t.Fatalf("Failed to post to login endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Login handler will try to authenticate and fail, should return 401 Unauthorized
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 (Unauthorized), got %d", resp.StatusCode)
		}
	})

	// Test logout handler
	t.Run("LogoutHandler", func(t *testing.T) {
		client := &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				// Don't follow redirects, we want to inspect them
				return http.ErrUseLastResponse
			},
		}
		
		resp, err := client.Get(server.URL + "/auth/logout")
		if err != nil {
			t.Fatalf("Failed to access logout endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Logout should redirect with 303 status
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("Expected status 303 (See Other), got %d", resp.StatusCode)
		}

		// Check redirect location
		location := resp.Header.Get("Location")
		if location != "/" {
			t.Errorf("Expected redirect to '/', got '%s'", location)
		}

		// Check that logout clears session cookies
		cookies := resp.Cookies()
		hasSessionClear := false
		hasRefreshClear := false
		for _, cookie := range cookies {
			if cookie.Name == "dsq_session" && cookie.MaxAge == -1 {
				hasSessionClear = true
			}
			if cookie.Name == "dsq_refresh" && cookie.MaxAge == -1 {
				hasRefreshClear = true
			}
		}
		
		if !hasSessionClear {
			t.Error("Logout should clear session cookie with MaxAge -1")
		}
		if !hasRefreshClear {
			t.Error("Logout should clear refresh cookie with MaxAge -1")
		}
	})

	// Test redirect handler
	t.Run("RedirectHandler", func(t *testing.T) {
		client := &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				// Don't follow redirects, we want to inspect them
				return http.ErrUseLastResponse
			},
		}
		
		// Test redirect handler with handle parameter
		resp, err := client.Get(server.URL + "/auth/redirect?handle=test.bsky.social")
		if err != nil {
			t.Fatalf("Failed to access redirect endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Redirect handler should redirect to OAuth provider
		if resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusFound {
			t.Errorf("Expected redirect status (302 or 303), got %d", resp.StatusCode)
		}

		location := resp.Header.Get("Location")
		if location == "" {
			t.Error("Redirect response should have Location header")
		} else {
			// Parse the redirect URL to verify OAuth parameters
			parsedURL, err := url.Parse(location)
			if err != nil {
				t.Fatalf("Failed to parse redirect URL: %v", err)
			}

			query := parsedURL.Query()
			
			// Check for required OAuth parameters
			requiredParams := []string{"client_id", "redirect_uri", "response_type", "state", "code_challenge", "code_challenge_method"}
			for _, param := range requiredParams {
				if query.Get(param) == "" {
					t.Errorf("OAuth redirect URL should contain parameter %s", param)
				}
			}

			t.Logf("OAuth redirect URL: %s", location)
		}
	})

	// Test redirect handler without handle parameter
	t.Run("RedirectHandler_NoHandle", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/auth/redirect")
		if err != nil {
			t.Fatalf("Failed to access redirect endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Should return an error when handle is missing
		if resp.StatusCode == http.StatusOK {
			t.Error("Redirect handler should return error when handle parameter is missing")
		}
	})

	// Test callback handler (this would normally be called by the OAuth provider)
	t.Run("CallbackHandler", func(t *testing.T) {
		// Test callback with missing parameters (should return error)
		resp, err := http.Get(server.URL + "/auth/callback")
		if err != nil {
			t.Fatalf("Failed to access callback endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Callback without proper parameters should return an error
		if resp.StatusCode == http.StatusOK {
			t.Error("Callback handler should return error when required parameters are missing")
		}

		// Test callback with some parameters (still missing valid code exchange)
		resp2, err := http.Get(server.URL + "/auth/callback?code=test-code&state=test-state")
		if err != nil {
			t.Fatalf("Failed to access callback endpoint with parameters: %v", err)
		}
		defer func() { _ = resp2.Body.Close() }()

		// This will likely fail due to invalid code, but should at least attempt to process
		// In a real integration test, you might mock the token exchange
		t.Logf("Callback with parameters returned status: %d", resp2.StatusCode)
	})
}

// TestAuthFlow_CookieIntegration tests cookie handling in the authentication flow
func TestAuthFlow_CookieIntegration(t *testing.T) {
	cfg := &config.Config{
		AppEnv:           config.EnvTest,
		OAuthClientID:    "test-client-id", 
		OAuthRedirectURL: "http://localhost:3000/auth/callback",
		PublicDomain:     "http://localhost:3000",
		AppName:          "dis.quest Test",
	}

	// Create test server
	mux := http.NewServeMux()
	RegisterRoutes(mux, "/auth", cfg)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test that login/logout flow properly manages cookies
	t.Run("Cookie Lifecycle", func(t *testing.T) {
		client := &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				// Don't follow redirects, we want to inspect them
				return http.ErrUseLastResponse
			},
		}

		// 1. Access logout to clear any existing cookies
		req, _ := http.NewRequest("GET", server.URL + "/auth/logout", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to access logout: %v", err)
		}
		_ = resp.Body.Close()

		// 2. Collect any cookies set by logout (should be clearing cookies)
		var cookieJar []*http.Cookie
		cookieJar = append(cookieJar, resp.Cookies()...)

		// 3. Verify that logout sets clearing cookies
		hasSessionClear := false
		hasRefreshClear := false
		for _, cookie := range cookieJar {
			if cookie.Name == "dsq_session" && cookie.MaxAge == -1 {
				hasSessionClear = true
			}
			if cookie.Name == "dsq_refresh" && cookie.MaxAge == -1 {
				hasRefreshClear = true
			}
		}

		if !hasSessionClear {
			t.Error("Logout should set session cookie with MaxAge -1")
		}
		if !hasRefreshClear {
			t.Error("Logout should set refresh cookie with MaxAge -1")
		}

		t.Logf("Collected %d cookies from logout", len(cookieJar))
	})
}

// TestHandlerRegistration_Integration tests that all handlers are properly registered
func TestHandlerRegistration_Integration(t *testing.T) {
	cfg := &config.Config{
		AppEnv:           config.EnvTest,
		OAuthClientID:    "test-client-id",
		OAuthRedirectURL: "http://localhost:3000/auth/callback", 
		PublicDomain:     "http://localhost:3000",
		AppName:          "dis.quest Test",
	}

	// Create test server
	mux := http.NewServeMux()
	RegisterRoutes(mux, "/auth", cfg)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test that all expected endpoints are registered and responding
	endpoints := map[string]int{
		"/auth/login":                 http.StatusMethodNotAllowed, // GET not allowed, needs POST
		"/auth/logout":                http.StatusSeeOther,         // Should redirect after clearing cookies
		"/auth/redirect":              http.StatusBadRequest,       // Should fail without handle param
		"/auth/callback":              http.StatusBadRequest,       // Should fail without proper OAuth params
		"/auth/client-metadata.json": http.StatusOK,               // Should serve JSON metadata
	}

	for endpoint, expectedStatus := range endpoints {
		t.Run("Endpoint_"+endpoint, func(t *testing.T) {
			client := &http.Client{
				CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
					// Don't follow redirects, we want to inspect them
					return http.ErrUseLastResponse
				},
			}
			
			resp, err := client.Get(server.URL + endpoint)
			if err != nil {
				t.Fatalf("Failed to access %s: %v", endpoint, err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != expectedStatus {
				t.Errorf("Endpoint %s returned status %d, expected %d", endpoint, resp.StatusCode, expectedStatus)
			}
		})
	}
}