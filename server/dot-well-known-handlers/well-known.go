package dotwellknown

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/config"
)

const blueskyClientMetadataFilename = "bluesky-client-metadata.json"
const jwksFilename = "jwks.json"
const redirectURIPath = "/auth/callback"

type WellKnownRouter struct {
	baseRoute string
	cfg       *config.Config
}

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
	router := &WellKnownRouter{
		baseRoute: baseRoute,
		cfg:       cfg,
	}

	mux.HandleFunc(baseRoute, router.WellKnownHandler)
	mux.HandleFunc(baseRoute+"/"+blueskyClientMetadataFilename, router.BlueskyClientMetadataHandler)
	mux.HandleFunc(baseRoute+"/"+jwksFilename, router.JWKSHandler)
}

// WellKnownHandler handles requests to the /.well-known endpoint.
func (router *WellKnownRouter) WellKnownHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","message":"Well-known endpoint"}`))
}

func (router *WellKnownRouter) BlueskyClientMetadataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	publicDomain := router.cfg.PublicDomain
	appName := router.cfg.AppName
	metadata := BlueskyClientMetadata{
		ClientID:                publicDomain + "/.well-known/bluesky-client-metadata.json",
		ClientName:              appName,
		ClientURI:               publicDomain,
		RedirectURIs:            []string{"http://localhost:3000/oauth/callback", publicDomain + "/oauth/callback"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		Scope:                   "atproto",
		TokenEndpointAuthMethod: "none",
		ApplicationType:         "web",
		DpopBoundAccessTokens:   true,
		JWKSURI:                 publicDomain + "/.well-known/jwks.json",
	}
	json.NewEncoder(w).Encode(metadata)
}

// JWKSHandler serves the public JWKS from keys/jwks.public.json.
func (router *WellKnownRouter) JWKSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, router.cfg.JWKS)
}
