package session

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
)

// CookieStorage implements Storage interface using HTTP cookies.
// This is suitable for web applications where session data travels with HTTP requests.
// Sessions are automatically scoped to the user's browser.
type CookieStorage struct {
	config CookieConfig
}

// CookieConfig contains configuration for cookie-based session storage.
type CookieConfig struct {
	// Cookie names
	SessionCookieName string
	DPoPCookieName    string
	
	// Security settings
	MaxAge       int
	SecureInProd bool
	SameSite     http.SameSite
	Domain       string
	Path         string
	
	// Encryption key for sensitive data
	EncryptionKey []byte
}

// NewCookieStorage creates a new cookie-based session storage.
func NewCookieStorage(encryptionKey []byte, config CookieConfig) Storage {
	// Set default values
	if config.SessionCookieName == "" {
		config.SessionCookieName = "dsq_session_data"
	}
	if config.DPoPCookieName == "" {
		config.DPoPCookieName = "dsq_dpop_key"
	}
	if config.MaxAge == 0 {
		config.MaxAge = 3600 // 1 hour
	}
	if config.Path == "" {
		config.Path = "/"
	}
	if config.SameSite == 0 {
		config.SameSite = http.SameSiteLaxMode
	}

	config.EncryptionKey = encryptionKey

	return &CookieStorage{
		config: config,
	}
}

// Store saves session data as HTTP cookies.
// The key parameter is ignored since cookies are automatically scoped to the browser.
func (c *CookieStorage) Store(ctx context.Context, key string, data *Data) error {
	if data == nil {
		return fmt.Errorf("session data cannot be nil")
	}

	// Get response writer from context
	w, ok := ctx.Value("http_response_writer").(http.ResponseWriter)
	if !ok {
		return fmt.Errorf("http.ResponseWriter not found in context")
	}

	// Create a sanitized copy of data for cookie storage (without sensitive fields)
	cookieData := &Data{
		SessionID:    data.SessionID,
		UserDID:      data.UserDID,
		Handle:       data.Handle,
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
		TokenType:    data.TokenType,
		ExpiresAt:    data.ExpiresAt,
		IssuedAt:     data.IssuedAt,
		CreatedAt:    data.CreatedAt,
		UpdatedAt:    data.UpdatedAt,
		Metadata:     data.Metadata,
		// DPoPKey is handled separately for security
	}

	// Serialize main session data
	sessionJSON, err := json.Marshal(cookieData)
	if err != nil {
		return fmt.Errorf("failed to serialize session data: %w", err)
	}

	// Encode session data
	sessionValue := base64.StdEncoding.EncodeToString(sessionJSON)

	// Set main session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     c.config.SessionCookieName,
		Value:    sessionValue,
		Path:     c.config.Path,
		Domain:   c.config.Domain,
		MaxAge:   c.config.MaxAge,
		HttpOnly: true,
		Secure:   c.config.SecureInProd,
		SameSite: c.config.SameSite,
	})

	// Store DPoP key separately using the oauth package's secure cookie handling
	if data.DPoPKey != nil {
		isDev := !c.config.SecureInProd
		if err := oauth.SetDPoPKeyCookie(w, data.DPoPKey, isDev); err != nil {
			return fmt.Errorf("failed to set DPoP key cookie: %w", err)
		}
	}

	return nil
}

// Load retrieves session data from HTTP cookies.
// The key parameter is ignored since cookies are automatically provided with requests.
func (c *CookieStorage) Load(ctx context.Context, key string) (*Data, error) {
	// Get request from context
	r, ok := ctx.Value("http_request").(*http.Request)
	if !ok {
		return nil, fmt.Errorf("*http.Request not found in context")
	}

	// Get main session cookie
	cookie, err := r.Cookie(c.config.SessionCookieName)
	if err != nil {
		return nil, fmt.Errorf("session cookie not found: %w", err)
	}

	// Decode session data
	sessionJSON, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session cookie: %w", err)
	}

	// Deserialize session data
	var data Data
	if err := json.Unmarshal(sessionJSON, &data); err != nil {
		return nil, fmt.Errorf("failed to deserialize session data: %w", err)
	}

	// Check if session has expired
	if time.Now().After(data.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	// Load DPoP key from separate cookie
	dpopKey, err := oauth.GetDPoPKeyFromCookie(r)
	if err == nil && dpopKey != nil {
		data.DPoPKey = dpopKey
	}
	// Note: DPoP key is optional, so we don't fail if it's missing

	return &data, nil
}

// Delete removes session cookies.
func (c *CookieStorage) Delete(ctx context.Context, key string) error {
	// Get response writer from context
	w, ok := ctx.Value("http_response_writer").(http.ResponseWriter)
	if !ok {
		return fmt.Errorf("http.ResponseWriter not found in context")
	}

	// Clear main session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     c.config.SessionCookieName,
		Value:    "",
		Path:     c.config.Path,
		Domain:   c.config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.config.SecureInProd,
		SameSite: c.config.SameSite,
	})

	// Clear DPoP key cookie using oauth package
	isDev := !c.config.SecureInProd
	oauth.ClearDPoPKeyCookie(w, isDev)

	return nil
}

// Cleanup is a no-op for cookie storage since cookies auto-expire.
func (c *CookieStorage) Cleanup(ctx context.Context) error {
	// Cookies automatically expire based on MaxAge, so no cleanup needed
	return nil
}

// Close is a no-op for cookie storage.
func (c *CookieStorage) Close() error {
	return nil
}

// WithHTTPContext creates a new context with HTTP request and response writer.
// This is a helper function for cookie storage operations.
func WithHTTPContext(ctx context.Context, r *http.Request, w http.ResponseWriter) context.Context {
	ctx = context.WithValue(ctx, "http_request", r)
	ctx = context.WithValue(ctx, "http_response_writer", w)
	return ctx
}