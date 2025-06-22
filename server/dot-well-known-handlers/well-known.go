// Package dotwellknown handles .well-known endpoints for OAuth2 and JWKS
package dotwellknown

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/svrlib"
	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
	"golang.org/x/oauth2"
)

const blueskyClientMetadataFilename = "bluesky-client-metadata.json"
const jwksFilename = "jwks.json"
const redirectURIPath = "/auth/callback"

// WellKnownRouter handles .well-known HTTP routes
type WellKnownRouter struct {
	*svrlib.Router
}

// BlueskyClientMetadata represents OAuth2 client metadata for Bluesky
type BlueskyClientMetadata struct {
	ClientID                string   `json:"client_id"`
	ClientName              string   `json:"client_name"`
	ClientURI               string   `json:"client_uri"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	Scope                   string   `json:"scope"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	ApplicationType         string   `json:"application_type"`
	DpopBoundAccessTokens   bool     `json:"dpop_bound_access_tokens"`
	JWKSURI                 string   `json:"jwks_uri"`
}

// RegisterRoutes registers the /.well-known route on the given mux.
func RegisterRoutes(mux *http.ServeMux, baseRoute string, cfg *config.Config) {
	router := &WellKnownRouter{svrlib.NewRouter(mux, baseRoute, cfg)}

	mux.HandleFunc(baseRoute, router.WellKnownHandler)
	mux.HandleFunc(baseRoute+"/"+blueskyClientMetadataFilename, router.BlueskyClientMetadataHandler)
	mux.HandleFunc(baseRoute+"/"+jwksFilename, router.JWKSHandler)
}

// WellKnownHandler serves the base .well-known endpoint
func (rt *WellKnownRouter) WellKnownHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","message":"Well-known endpoint"}`))
}

// BlueskyClientMetadataHandler serves the Bluesky OAuth2 client metadata
func (rt *WellKnownRouter) BlueskyClientMetadataHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	publicDomain := rt.Config.PublicDomain
	appName := rt.Config.AppName
	metadata := BlueskyClientMetadata{
		ClientID:                publicDomain + "/.well-known/bluesky-client-metadata.json",
		ClientName:              appName,
		ClientURI:               publicDomain,
		RedirectURIs:            []string{"http://localhost:3000" + redirectURIPath, publicDomain + redirectURIPath},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		Scope:                   "atproto",
		TokenEndpointAuthMethod: "none",
		ApplicationType:         "web",
		DpopBoundAccessTokens:   true,
		JWKSURI:                 publicDomain + "/.well-known/jwks.json",
	}
	_ = json.NewEncoder(w).Encode(metadata)
}

// JWKSHandler serves the public JWKS from keys/jwks.public.json.
func (rt *WellKnownRouter) JWKSHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, rt.Config.JWKSPublic)
}

// RedirectHandler handles OAuth2 redirect with dynamically generated redirect URI.
func (rt *WellKnownRouter) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("handle")
	if handle == "" {
		http.Error(w, "missing handle parameter", http.StatusBadRequest)
		return
	}

	// Use PKCE and state helpers from internal/auth
	state := generateStateToken()
	codeVerifier, codeChallenge, err := oauth.GeneratePKCE()
	if err != nil {
		http.Error(w, "failed to generate PKCE", http.StatusInternalServerError)
		return
	}

	// Set cookies for state, codeVerifier, and handle (for callback validation)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
	})
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
	})

	// Use PublicDomain from config for redirect URI
	publicDomain := rt.Config.PublicDomain
	redirectURI := publicDomain + "/auth/callback"

	// Get OAuth2 config with correct redirect URI
	metadata, err := oauth.DiscoverAuthorizationServer(handle)
	if err != nil {
		http.Error(w, "failed to discover authorization server", http.StatusInternalServerError)
		return
	}
	cfg := rt.Config
	providerConfig := &oauth.ProviderConfig{
		ClientID:       cfg.OAuthClientID,
		ClientURI:      cfg.PublicDomain,
		RedirectURI:    cfg.OAuthRedirectURL,
		PDSEndpoint:    cfg.PDSEndpoint,
		JWKSPrivateKey: cfg.JWKSPrivate,
		JWKSPublicKey:  cfg.JWKSPublic,
		Scope:          "atproto transition:generic",
	}
	conf := oauth.OAuth2Config(metadata, providerConfig)
	conf.RedirectURL = redirectURI

	// Generate auth URL with required parameters
	authURL := conf.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("handle", handle),
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

// generateStateToken generates a random state token for OAuth flows
func generateStateToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		// fallback: not cryptographically secure, but avoids panic
		return base64.RawURLEncoding.EncodeToString([]byte("fallback_state_token"))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
