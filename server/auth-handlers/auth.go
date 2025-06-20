// Package auth handles HTTP routes for authentication
package auth

import (
	"fmt"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/svrlib"
	"golang.org/x/oauth2"
)

// Router handles authentication-related HTTP routes
type Router struct {
	*svrlib.Router
}

// RegisterRoutes registers all /auth/* routes on the given mux, with the prefix handled by the caller.
func RegisterRoutes(mux *http.ServeMux, prefix string, cfg *config.Config) {
	router := &Router{svrlib.NewRouter(mux, prefix, cfg)}
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
	http.Redirect(w, r, "/discussion", http.StatusSeeOther)
}

// LogoutHandler handles /auth/logout requests
func (rt *Router) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LogoutHandlerWithConfig handles /auth/logout requests with config for cookie security
func (rt *Router) LogoutHandlerWithConfig(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	auth.ClearSessionCookieWithEnv(w, cfg.AppEnv == "development")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// RedirectHandler handles /auth/redirect requests
func (rt *Router) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("handle")
	if handle == "" {
		writeError(w, http.StatusBadRequest, "Missing handle", "param", "handle")
		return
	}
	metadata, err := auth.DiscoverAuthorizationServer(handle)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to discover authorization server", "handle", handle, "error", err)
		return
	}
	codeVerifier, codeChallenge, err := auth.GeneratePKCE()
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
	http.SetCookie(w, &http.Cookie{
		Name:     "pkce_verifier",
		Value:    codeVerifier,
		Path:     "/",
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_handle",
		Value:    handle,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})
	state := auth.GenerateStateToken()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})
	conf := auth.OAuth2Config(metadata, cfg)
	url := conf.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("login_hint", handle),
	)
	http.Redirect(w, r, url, http.StatusFound)
}

// CallbackHandler handles /auth/callback requests
func (rt *Router) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	handleCookie, err := r.Cookie("oauth_handle")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Missing handle context")
		return
	}
	handle := handleCookie.Value
	metadata, err := auth.DiscoverAuthorizationServer(handle)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to rediscover authorization server", "handle", handle, "error", err)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
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
	// Retrieve DPoP private key from secure cookie
	dpopKey, err := auth.GetDPoPKeyFromCookie(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Missing DPoP key", "handle", handle)
		return
	}
	cfg := rt.Config
	logger.Info("Starting token exchange with DPoP", "handle", handle, "code", code[:10]+"...", "tokenEndpoint", metadata.TokenEndpoint)
	token, err := auth.ExchangeCodeForTokenWithDPoP(ctx, metadata, code, verCookie.Value, dpopKey, cfg)
	if err != nil {
		logger.Error("Token exchange failed", "handle", handle, "error", err, "tokenEndpoint", metadata.TokenEndpoint)
		writeError(w, http.StatusUnauthorized, "Token exchange failed", "handle", handle, "error", err)
		return
	}
	logger.Info("Token exchange successful", "handle", handle)
	refreshToken := ""
	if token.RefreshToken != "" {
		refreshToken = token.RefreshToken
	}
	// Use config for secure flag
	auth.SetSessionCookieWithEnv(w, token.AccessToken, []string{refreshToken}, cfg.AppEnv == "development")
	http.Redirect(w, r, "/discussion", http.StatusSeeOther)
}

// ClientMetadataHandler serves the OAuth client metadata JSON for Bluesky
func (rt *Router) ClientMetadataHandler(w http.ResponseWriter, _ *http.Request) {
	cfg := rt.Config
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
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
	  "token_endpoint_auth_method": "none"
	}`, cfg.OAuthClientID, cfg.AppName, cfg.PublicDomain, cfg.OAuthRedirectURL)
	
	_, _ = w.Write([]byte(metadata))
}

// writeError is a helper to write an error response and log it
func writeError(w http.ResponseWriter, status int, reason string, logFields ...any) {
	http.Error(w, reason, status)
	logger.Error(reason, logFields...)
}
