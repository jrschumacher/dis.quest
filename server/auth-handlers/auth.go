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
		w.WriteHeader(http.StatusMethodNotAllowed)
		logger.Info("Method not allowed", "path", r.URL.Path)
		fmt.Fprintln(w, "Method not allowed")
		return
	}
	// TODO: Parse handle and app password from form
	// TODO: Call ATProto session create endpoint
	// TODO: On success, set session cookie
	logger.Info("Stub: Handle ATProto login")
	fmt.Fprintln(w, "[Stub] Handle ATProto login (handle + app password)")
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
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("Missing handle param")
		fmt.Fprintln(w, "Missing handle")
		return
	}
	provider := "https://bsky.social" // TODO: Discover from handle if federated
	codeVerifier, codeChallenge, err := auth.GeneratePKCE()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error("Failed to generate PKCE challenge", "error", err)
		fmt.Fprintln(w, "Failed to generate PKCE challenge")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "pkce_verifier",
		Value:    codeVerifier,
		Path:     "/",
		HttpOnly: true,
	})
	conf := auth.OAuth2Config(provider)
	url := conf.AuthCodeURL("state-xyz",
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("handle", handle),
	)
	http.Redirect(w, r, url, http.StatusFound)
}

// CallbackHandler handles /auth/callback requests
func (rt *AuthRouter) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	provider := "https://bsky.social" // TODO: Discover from state/session if federated
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("Missing code param")
		fmt.Fprintln(w, "Missing code")
		return
	}
	verCookie, err := r.Cookie("pkce_verifier")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("Missing PKCE verifier cookie")
		fmt.Fprintln(w, "Missing PKCE verifier")
		return
	}
	token, err := auth.ExchangeCodeForToken(ctx, provider, code, verCookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Error("Token exchange failed", "error", err)
		fmt.Fprintf(w, "Token exchange failed: %v", err)
		return
	}
	auth.SetSessionCookie(w, token.AccessToken)
	http.Redirect(w, r, "/discussion", http.StatusSeeOther)
}
