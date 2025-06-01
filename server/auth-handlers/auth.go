package auth

import (
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/svrlib"
	"golang.org/x/oauth2"
)

type AuthRouter struct {
	*svrlib.Router
}

// RegisterRoutes registers all /auth/* routes on the given mux, with the prefix handled by the caller.
func RegisterRoutes(mux *http.ServeMux, prefix string, cfg *config.Config) {
	router := &AuthRouter{svrlib.NewRouter(mux, prefix, cfg)}
	mux.HandleFunc(prefix+"/login", router.LoginHandler)
	mux.HandleFunc(prefix+"/logout", router.LogoutHandler)
	mux.HandleFunc(prefix+"/redirect", router.RedirectHandler)
	mux.HandleFunc(prefix+"/callback", router.CallbackHandler)
}

// LoginHandler handles POST /login requests
func (rt *AuthRouter) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Info("Method not allowed", "path", r.URL.Path)
		return
	}
	// TODO: Only needed if supporting app password login. Remove if not supporting direct app password login.
	// TODO: Parse handle and app password from form
	// TODO: Call ATProto session create endpoint
	// TODO: On success, set session cookie
	logger.Info("Stub: Handle ATProto login")
	http.Error(w, "[Stub] Handle ATProto login (handle + app password)", http.StatusNotImplemented)
}

// LogoutHandler handles /auth/logout requests
func (rt *AuthRouter) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// RedirectHandler handles /auth/redirect requests
func (rt *AuthRouter) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("handle")
	if handle == "" {
		http.Error(w, "Missing handle", http.StatusBadRequest)
		logger.Error("Missing handle param")
		return
	}
	provider, err := auth.DiscoverPDS(handle)
	if err != nil {
		http.Error(w, "Failed to discover PDS", http.StatusInternalServerError)
		logger.Error("Failed to discover PDS", "error", err)
		return
	}
	codeVerifier, codeChallenge, err := auth.GeneratePKCE()
	if err != nil {
		http.Error(w, "Failed to generate PKCE challenge", http.StatusInternalServerError)
		logger.Error("Failed to generate PKCE challenge", "error", err)
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
	conf := auth.OAuth2Config(provider)
	url := conf.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("handle", handle),
	)
	http.Redirect(w, r, url, http.StatusFound)
}

// CallbackHandler handles /auth/callback requests
func (rt *AuthRouter) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	handleCookie, err := r.Cookie("oauth_handle")
	if err != nil {
		http.Error(w, "Missing handle context", http.StatusBadRequest)
		logger.Error("Missing handle cookie")
		return
	}
	provider, err := auth.DiscoverPDS(handleCookie.Value)
	if err != nil {
		http.Error(w, "Failed to rediscover PDS", http.StatusInternalServerError)
		logger.Error("Failed to rediscover PDS", "error", err)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		logger.Error("Missing code param")
		return
	}
	// State validation
	state := r.URL.Query().Get("state")
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || state != stateCookie.Value {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		logger.Error("Invalid state", "expected", stateCookie.Value, "got", state)
		return
	}
	verCookie, err := r.Cookie("pkce_verifier")
	if err != nil {
		http.Error(w, "Missing PKCE verifier", http.StatusBadRequest)
		logger.Error("Missing PKCE verifier cookie")
		return
	}
	token, err := auth.ExchangeCodeForToken(ctx, provider, code, verCookie.Value)
	if err != nil {
		http.Error(w, "Token exchange failed", http.StatusUnauthorized)
		logger.Error("Token exchange failed", "error", err)
		return
	}
	// Store both access and refresh tokens in cookies for long-lived sessions
	refreshToken := ""
	if token.RefreshToken != "" {
		refreshToken = token.RefreshToken
	}
	auth.SetSessionCookie(w, token.AccessToken, refreshToken)
	http.Redirect(w, r, "/discussion", http.StatusSeeOther)
}
