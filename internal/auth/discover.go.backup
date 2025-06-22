package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AuthorizationServerMetadata represents OAuth Authorization Server metadata
type AuthorizationServerMetadata struct {
	Issuer                               string   `json:"issuer"`
	AuthorizationEndpoint                string   `json:"authorization_endpoint"`
	TokenEndpoint                        string   `json:"token_endpoint"`
	PushedAuthorizationRequestEndpoint   string   `json:"pushed_authorization_request_endpoint"`
	ScopesSupported                      []string `json:"scopes_supported"`
	DPoPSigningAlgValuesSupported        []string `json:"dpop_signing_alg_values_supported"`
}

// DiscoverPDS returns the PDS base URL for a given handle (Bluesky username).
// For Bluesky, this is always https://bsky.social. In the future, this could look up a handle in DNS or other registry.
func DiscoverPDS(_ string) (string, error) {
	// For now, always return Bluesky's PDS endpoint
	return "https://bsky.social", nil
}

// DiscoverAuthorizationServer discovers the OAuth authorization server metadata for a given handle
func DiscoverAuthorizationServer(handle string) (*AuthorizationServerMetadata, error) {
	// For Bluesky handles, we need to resolve to the authorization server
	// First discover the PDS
	pds, err := DiscoverPDS(handle)
	if err != nil {
		return nil, fmt.Errorf("failed to discover PDS for handle %s: %w", handle, err)
	}
	
	// For Bluesky, the authorization server is typically the same as the PDS
	// but we should fetch the metadata to be sure
	metadataURL := strings.TrimSuffix(pds, "/") + "/.well-known/oauth-authorization-server"
	
	// #nosec G107 -- URL is constructed from trusted PDS discovery, not user input
	resp, err := http.Get(metadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch authorization server metadata from %s: %w", metadataURL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authorization server metadata endpoint returned status %d", resp.StatusCode)
	}
	
	var metadata AuthorizationServerMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode authorization server metadata: %w", err)
	}
	
	return &metadata, nil
}
