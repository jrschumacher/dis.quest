package server

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/jrschumacher/dis.quest/components"
	"github.com/jrschumacher/dis.quest/internal/auth"
	"golang.org/x/oauth2"
)

func Start() {

	http.Handle("/", templ.Handler(components.Page(false)))
	http.Handle("/login", templ.Handler(components.Login()))
	http.Handle("/discussion", templ.Handler(components.Discussion()))

	http.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, "Method not allowed")
			return
		}
		// TODO: Parse handle and app password from form
		// TODO: Call ATProto session create endpoint
		// TODO: On success, set session cookie
		fmt.Fprintln(w, "[Stub] Handle ATProto login (handle + app password)")
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		auth.ClearSessionCookie(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/auth/redirect", func(w http.ResponseWriter, r *http.Request) {
		handle := r.URL.Query().Get("handle")
		if handle == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Missing handle")
			return
		}
		// Discover provider (for Bluesky, use https://bsky.social)
		provider := "https://bsky.social" // TODO: Discover from handle if federated
		codeVerifier, codeChallenge, err := auth.GeneratePKCE()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Failed to generate PKCE challenge")
			return
		}
		// Store codeVerifier in a temporary cookie (for demo; use state or session in production)
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
	})

	http.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		provider := "https://bsky.social" // TODO: Discover from state/session if federated
		code := r.URL.Query().Get("code")
		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Missing code")
			return
		}
		verCookie, err := r.Cookie("pkce_verifier")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Missing PKCE verifier")
			return
		}
		token, err := auth.ExchangeCodeForToken(ctx, provider, code, verCookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Token exchange failed: %v", err)
			return
		}
		auth.SetSessionCookie(w, token.AccessToken)
		http.Redirect(w, r, "/discussion", http.StatusSeeOther)
	})

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
