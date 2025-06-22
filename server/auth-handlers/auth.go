// Package auth handles HTTP routes for authentication
package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/oauth"
	"github.com/jrschumacher/dis.quest/internal/svrlib"
	"github.com/jrschumacher/dis.quest/pkg/atproto"
)

// Router handles authentication-related HTTP routes
type Router struct {
	*svrlib.Router
	oauthService *oauth.Service
}

// RegisterRoutes registers all /auth/* routes on the given mux, with the prefix handled by the caller.
func RegisterRoutes(mux *http.ServeMux, prefix string, cfg *config.Config, oauthService *oauth.Service) {
	router := &Router{
		Router:       svrlib.NewRouter(mux, prefix, cfg),
		oauthService: oauthService,
	}
	// Pass config to handlers for env-aware cookie security
	routerConfig := cfg

	// Wrap handlers to inject config for cookie security
	mux.HandleFunc(prefix+"/login", func(w http.ResponseWriter, r *http.Request) { router.LoginHandlerWithConfig(w, r, routerConfig) })
	mux.HandleFunc(prefix+"/logout", func(w http.ResponseWriter, r *http.Request) { router.LogoutHandlerWithConfig(w, r, routerConfig) })
	mux.HandleFunc(prefix+"/redirect", router.RedirectHandler)
	mux.HandleFunc(prefix+"/callback", router.CallbackHandler)
	mux.HandleFunc(prefix+"/client-metadata.json", router.ClientMetadataHandler)
}

// LoginHandler handles POST /login requests
func (rt *Router) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "path", r.URL.Path)
		return
	}
	// TODO: Only needed if supporting app password login. Remove if not supporting direct app password login.
	// TODO: Parse handle and app password from form
	// TODO: Call ATProto session create endpoint
	// TODO: On success, set session cookie
	logger.Info("Stub: Handle ATProto login")
	http.Error(w, "[Stub] Handle ATProto login (handle + app password)", http.StatusNotImplemented)
}

// LoginHandlerWithConfig handles POST /login requests with config for cookie security
func (rt *Router) LoginHandlerWithConfig(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "path", r.URL.Path)
		return
	}
	handle := r.FormValue("handle")
	password := r.FormValue("password")
	if handle == "" || password == "" {
		writeError(w, http.StatusBadRequest, "Missing handle or password")
		return
	}
	provider, err := auth.DiscoverPDS(handle)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to discover PDS", "handle", handle, "error", err)
		return
	}
	session, err := auth.CreateSession(provider, handle, password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid credentials", "handle", handle, "error", err)
		return
	}
	auth.SetSessionCookieWithEnv(w, session.AccessJwt, []string{session.RefreshJwt}, cfg.AppEnv == "development")
	
	// Check for redirect URL and use it, otherwise default to /discussion
	redirectURL := "/discussion"
	if cookie, err := r.Cookie("redirect_after_login"); err == nil && cookie.Value != "" {
		if isValidRedirectURL(cookie.Value) {
			redirectURL = cookie.Value
		}
		// Clear the redirect cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "redirect_after_login",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   cfg.AppEnv != "development",
			SameSite: http.SameSiteLaxMode,
		})
	}
	
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// LogoutHandler handles /auth/logout requests
func (rt *Router) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.ClearSessionCookie(w)
	
	// Check for redirect parameter to redirect to login with return URL
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL != "" && isValidRedirectURL(redirectURL) {
		// Redirect to login page with the redirect parameter
		loginURL := fmt.Sprintf("/login?redirect=%s", redirectURL)
		http.Redirect(w, r, loginURL, http.StatusSeeOther)
		return
	}
	
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LogoutHandlerWithConfig handles /auth/logout requests with config for cookie security
func (rt *Router) LogoutHandlerWithConfig(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	auth.ClearSessionCookieWithEnv(w, cfg.AppEnv == "development")
	
	// Check for redirect parameter to redirect to login with return URL
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL != "" && isValidRedirectURL(redirectURL) {
		// Redirect to login page with the redirect parameter
		loginURL := fmt.Sprintf("/login?redirect=%s", redirectURL)
		http.Redirect(w, r, loginURL, http.StatusSeeOther)
		return
	}
	
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// RedirectHandler handles /auth/redirect requests
func (rt *Router) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("handle")
	if handle == "" {
		writeError(w, http.StatusBadRequest, "Missing handle", "param", "handle")
		return
	}
	
	// Store redirect URL in cookie if provided
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL != "" {
		// Validate redirect URL to prevent open redirects
		if isValidRedirectURL(redirectURL) {
			http.SetCookie(w, &http.Cookie{
				Name:     "redirect_after_login",
				Value:    redirectURL,
				Path:     "/",
				MaxAge:   600, // 10 minutes
				HttpOnly: true,
				Secure:   rt.Config.AppEnv != "development",
				SameSite: http.SameSiteLaxMode,
			})
		}
	}
	metadata, err := auth.DiscoverAuthorizationServer(handle)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to discover authorization server", "handle", handle, "error", err)
		return
	}
	codeVerifier, _, err := auth.GeneratePKCE()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate PKCE challenge", "handle", handle, "error", err)
		return
	}
	// Generate and store DPoP keypair in secure cookie
	dpopKey, err := auth.GenerateDPoPKeyPair()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate DPoP keypair", "handle", handle, "error", err)
		return
	}
	cfg := rt.Config
	if err := auth.SetDPoPKeyCookie(w, dpopKey.PrivateKey, cfg.AppEnv == "development"); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to set DPoP key cookie", "handle", handle, "error", err)
		return
	}
	secure := cfg.AppEnv != "development"
	http.SetCookie(w, &http.Cookie{
		Name:     "pkce_verifier",
		Value:    codeVerifier,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_handle",
		Value:    handle,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})
	state := auth.GenerateStateToken()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})
	
	// Use PAR (Pushed Authorization Request) instead of direct OAuth redirect
	parClient := auth.NewPARClient()
	
	// Get PAR endpoint from authorization server metadata
	parEndpoint := metadata.PushedAuthorizationRequestEndpoint
	if parEndpoint == "" {
		// Fallback: construct PAR endpoint from issuer
		parEndpoint = strings.TrimSuffix(metadata.Issuer, "/") + "/oauth/par"
	}
	
	// Perform PAR request
	parResp, err := parClient.PerformPAR(r.Context(), parEndpoint, metadata, codeVerifier, state, dpopKey.PrivateKey, cfg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to perform PAR request", "handle", handle, "error", err)
		return
	}
	
	// Store DPoP nonce if present (needed for token exchange)
	if parResp.DPoPNonce != "" {
		if err := auth.SetDPoPNonceCookie(w, parResp.DPoPNonce, cfg.AppEnv == "development"); err != nil {
			logger.Error("failed to set DPoP nonce cookie", "error", err)
		}
	}
	
	// Store auth server issuer for token exchange
	if parResp.AuthServerIssuer != "" {
		if err := auth.SetAuthServerIssuerCookie(w, parResp.AuthServerIssuer, cfg.AppEnv == "development"); err != nil {
			logger.Error("failed to set auth server issuer cookie", "error", err)
		}
	}
	
	// Redirect using PAR request_uri instead of direct parameters
	authURL := fmt.Sprintf("%s?client_id=%s&request_uri=%s", 
		metadata.AuthorizationEndpoint,
		url.QueryEscape(cfg.OAuthClientID), 
		url.QueryEscape(parResp.RequestURI))
	
	logger.Info("Redirecting to authorization server with PAR", "authURL", authURL, "requestURI", parResp.RequestURI)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// CallbackHandler handles /auth/callback requests
func (rt *Router) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Log all query parameters for debugging
	logger.Info("OAuth callback received", "url", r.URL.String(), "params", r.URL.Query())
	
	handleCookie, err := r.Cookie("oauth_handle")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Missing handle context")
		return
	}
	handle := handleCookie.Value
	
	// Check for error parameter first
	if errorParam := r.URL.Query().Get("error"); errorParam != "" {
		errorDesc := r.URL.Query().Get("error_description")
		logger.Error("OAuth authorization failed", "handle", handle, "error", errorParam, "description", errorDesc)
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Authorization failed: %s - %s", errorParam, errorDesc), "handle", handle)
		return
	}
	
	// Note: Authorization server discovery is now handled by the OAuth provider
	code := r.URL.Query().Get("code")
	if code == "" {
		logger.Error("No authorization code received", "handle", handle, "allParams", r.URL.Query())
		writeError(w, http.StatusBadRequest, "Missing code", "handle", handle)
		return
	}
	// State validation
	state := r.URL.Query().Get("state")
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || state != stateCookie.Value {
		writeError(w, http.StatusBadRequest, "Invalid state", "handle", handle, "expected", stateCookie.Value, "got", state)
		return
	}
	verCookie, err := r.Cookie("pkce_verifier")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Missing PKCE verifier", "handle", handle)
		return
	}
	// Note: DPoP key management is now handled by the OAuth provider
	cfg := rt.Config
	
	// Create OAuth service with configured provider
	oauthService, err := oauth.NewService(cfg)
	if err != nil {
		logger.Error("Failed to create OAuth service", "handle", handle, "error", err)
		writeError(w, http.StatusInternalServerError, "OAuth service initialization failed", "handle", handle, "error", err)
		return
	}
	
	logger.Info("Starting token exchange", "handle", handle, "code", code[:10]+"...", "provider", oauthService.GetProviderName())
	// Inject HTTP request into context for provider access to cookies/session
	ctxWithRequest := context.WithValue(ctx, "http_request", r)
	tokenResult, err := oauthService.ExchangeToken(ctxWithRequest, code, verCookie.Value)
	if err != nil {
		logger.Error("Token exchange failed", "handle", handle, "error", err, "provider", oauthService.GetProviderName())
		writeError(w, http.StatusUnauthorized, "Token exchange failed", "handle", handle, "error", err)
		return
	}
	logger.Info("Token exchange successful", "handle", handle, "provider", oauthService.GetProviderName())
	
	// Create enhanced session wrapper with pkg/atproto.Session integration
	sessionWrapper, err := auth.NewSessionWrapper(
		tokenResult.AccessToken, 
		tokenResult.RefreshToken, 
		tokenResult.UserDID, 
		tokenResult.DPoPKey, 
		nil, // atproto client will be set if available in tokenResult
	)
	if err != nil {
		logger.Error("Failed to create session wrapper", "handle", handle, "error", err)
		writeError(w, http.StatusInternalServerError, "Session creation failed", "handle", handle, "error", err)
		return
	}
	
	// Set the atproto session if available from token result
	if tokenResult.AtprotoSession != nil {
		if atprotoSession, ok := tokenResult.AtprotoSession.(*atproto.Session); ok {
			sessionWrapper.SetAtprotoSession(atprotoSession)
			logger.Info("Enhanced session created with atproto.Session", "handle", handle, "userDID", tokenResult.UserDID)
		}
	}
	
	// Save session to cookies using the wrapper
	if err := sessionWrapper.SaveToCookies(w, cfg.AppEnv == "development"); err != nil {
		logger.Error("Failed to save session to cookies", "handle", handle, "error", err)
		writeError(w, http.StatusInternalServerError, "Session storage failed", "handle", handle, "error", err)
		return
	}
	
	// Check for redirect URL and use it, otherwise default to /discussion
	redirectURL := "/discussion"
	if cookie, err := r.Cookie("redirect_after_login"); err == nil && cookie.Value != "" {
		if isValidRedirectURL(cookie.Value) {
			redirectURL = cookie.Value
		}
		// Clear the redirect cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "redirect_after_login",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   cfg.AppEnv != "development",
			SameSite: http.SameSiteLaxMode,
		})
	}
	
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// getClientAuthJWK creates a JWK for client authentication from the app's JWKS
func getClientAuthJWK(cfg *config.Config) map[string]interface{} {
	// For now, use the same key from JWKS for client authentication
	// In production, you might want a separate client auth key
	var jwks struct {
		Keys []map[string]interface{} `json:"keys"`
	}
	
	if err := json.Unmarshal([]byte(cfg.JWKSPublic), &jwks); err != nil || len(jwks.Keys) == 0 {
		// Fallback: generate a temporary key if JWKS parsing fails
		key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		return map[string]interface{}{
			"kty": "EC",
			"crv": "P-256", 
			"x":   base64.RawURLEncoding.EncodeToString(key.PublicKey.X.Bytes()),
			"y":   base64.RawURLEncoding.EncodeToString(key.PublicKey.Y.Bytes()),
			"alg": "ES256",
			"use": "sig",
		}
	}
	
	return jwks.Keys[0]
}

// ClientMetadataHandler serves the OAuth client metadata JSON for Bluesky
func (rt *Router) ClientMetadataHandler(w http.ResponseWriter, _ *http.Request) {
	cfg := rt.Config
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// Get client authentication public key from JWKS
	clientPublicJWK := getClientAuthJWK(cfg)
	clientJWKJSON, _ := json.Marshal(clientPublicJWK)
	
	// Use config values for dynamic metadata
	metadata := fmt.Sprintf(`{
	  "client_id": "%s",
	  "client_name": "%s", 
	  "client_uri": "%s",
	  "application_type": "web",
	  "dpop_bound_access_tokens": true,
	  "grant_types": ["authorization_code", "refresh_token"],
	  "scope": "atproto transition:generic",
	  "response_types": ["code"],  
	  "redirect_uris": ["%s"],
	  "token_endpoint_auth_method": "private_key_jwt",
	  "token_endpoint_auth_signing_alg": "ES256",
	  "jwks": {
		"keys": [%s]
	  }
	}`, cfg.OAuthClientID, cfg.AppName, cfg.PublicDomain, cfg.OAuthRedirectURL, string(clientJWKJSON))
	
	_, _ = w.Write([]byte(metadata))
}

// isValidRedirectURL validates that the redirect URL is safe to prevent open redirects
func isValidRedirectURL(url string) bool {
	// Only allow relative URLs that start with /
	// This prevents open redirects to external sites
	return strings.HasPrefix(url, "/") && !strings.HasPrefix(url, "//")
}

// writeError is a helper to write an error response and log it
func writeError(w http.ResponseWriter, status int, reason string, logFields ...any) {
	http.Error(w, reason, status)
	logger.Error(reason, logFields...)
}
