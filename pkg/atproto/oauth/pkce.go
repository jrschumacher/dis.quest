// Package oauth provides PKCE support for OAuth2 flows
package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
)

// PKCETransport adds PKCE code_verifier to token exchange requests
type PKCETransport struct {
	Base         http.RoundTripper
	CodeVerifier string
}

// RoundTrip implements http.RoundTripper
func (t *PKCETransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only modify token endpoint requests (POST with grant_type=authorization_code)
	if req.Method == "POST" && req.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if req.Body != nil {
			// Read the existing body
			body, err := io.ReadAll(req.Body)
			if err == nil {
				_ = req.Body.Close()
				
				// Parse form values
				values, err := url.ParseQuery(string(body))
				if err == nil && values.Get("grant_type") == "authorization_code" {
					// Add code_verifier
					values.Set("code_verifier", t.CodeVerifier)
					newBody := values.Encode()
					req.Body = io.NopCloser(strings.NewReader(newBody))
					req.ContentLength = int64(len(newBody))
				}
			}
		}
	}
	
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// DPoPPKCETransport adds both PKCE code_verifier and DPoP header to token exchange requests
type DPoPPKCETransport struct {
	Base         http.RoundTripper
	CodeVerifier string
	DPoPKey      DPoPKeyPair
	TargetURL    string
	Config       *ProviderConfig
	Metadata     *AuthorizationServerMetadata
}

// RoundTrip implements http.RoundTripper for DPoP + PKCE with nonce retry
func (t *DPoPPKCETransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request body since we might need to retry
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		if err := req.Body.Close(); err != nil {
			// Ignore close error, common in HTTP clients
			_ = err
		}
	}
	
	// Helper to create request with DPoP and PKCE
	makeRequest := func(nonce string) (*http.Request, error) {
		// Clone the original request
		newReq := req.Clone(req.Context())
		
		// Create DPoP JWT with optional nonce
		dpopJWT, err := t.DPoPKey.CreateDPoPJWTWithNonce(req.Method, t.TargetURL, nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to create DPoP JWT: %w", err)
		}
		
		// Add DPoP header
		newReq.Header.Set("DPoP", dpopJWT)
		
		// Restore and modify body for PKCE and client_assertion
		if bodyBytes != nil {
			// Parse form values
			values, err := url.ParseQuery(string(bodyBytes))
			if err == nil && values.Get("grant_type") == "authorization_code" {
				// Add code_verifier for PKCE
				values.Set("code_verifier", t.CodeVerifier)
				
				// Add client_assertion for private_key_jwt authentication
				parClient := NewPARClient()
				clientAssertion, err := parClient.CreateClientAssertion(t.Config.ClientID, t.Metadata.Issuer, t.Config)
				if err == nil {
					values.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
					values.Set("client_assertion", clientAssertion)
				}
				
				newBody := values.Encode()
				newReq.Body = io.NopCloser(strings.NewReader(newBody))
				newReq.ContentLength = int64(len(newBody))
			} else {
				newReq.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
				newReq.ContentLength = int64(len(bodyBytes))
			}
		}
		
		return newReq, nil
	}
	
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	
	// First attempt without nonce
	firstReq, err := makeRequest("")
	if err != nil {
		return nil, err
	}
	
	resp, err := base.RoundTrip(firstReq)
	if err != nil {
		return resp, err
	}
	
	// Check if we got a use_dpop_nonce error
	if resp.StatusCode == 400 {
		// Read response to check for nonce requirement
		respBody, err := io.ReadAll(resp.Body)
		if err := resp.Body.Close(); err != nil {
			// Ignore close error, common in HTTP clients
			_ = err
		}
		if err == nil && strings.Contains(string(respBody), "use_dpop_nonce") {
			// Extract nonce from DPoP-Nonce header
			if nonce := resp.Header.Get("DPoP-Nonce"); nonce != "" {
				// Retry with nonce
				retryReq, err := makeRequest(nonce)
				if err != nil {
					return nil, err
				}
				return base.RoundTrip(retryReq)
			}
		}
		// Restore the response body for the original error
		resp.Body = io.NopCloser(strings.NewReader(string(respBody)))
	}
	
	return resp, err
}

// GeneratePKCE generates a PKCE code verifier and challenge for OAuth2 flows
func GeneratePKCE() (codeVerifier, codeChallenge string, err error) {
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	codeVerifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	codeChallenge = base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return
}

// OAuth2Config creates an OAuth2 configuration for Bluesky/ATProto using authorization server metadata
func OAuth2Config(metadata *AuthorizationServerMetadata, cfg *ProviderConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: "", // Not required for public clients
		RedirectURL:  cfg.RedirectURI,
		// CRITICAL: Request scopes as separate items based on server metadata
		Scopes:       []string{"atproto", "transition:generic", "transition:chat.bsky"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  metadata.AuthorizationEndpoint,
			TokenURL: metadata.TokenEndpoint,
		},
	}
}

// ExchangeCodeForToken exchanges an authorization code for an access token
func ExchangeCodeForToken(ctx context.Context, metadata *AuthorizationServerMetadata, code, codeVerifier string, cfg *ProviderConfig) (*oauth2.Token, error) {
	conf := OAuth2Config(metadata, cfg)
	// For PKCE, we need to include the code_verifier in the token exchange request
	// The oauth2 library doesn't directly support this, so we need to add it via context
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: &PKCETransport{
			Base: http.DefaultTransport,
			CodeVerifier: codeVerifier,
		},
	})
	return conf.Exchange(ctx, code)
}

// ExchangeCodeForTokenWithDPoP exchanges an authorization code for an access token using DPoP
func ExchangeCodeForTokenWithDPoP(ctx context.Context, metadata *AuthorizationServerMetadata, code, codeVerifier string, dpopKey DPoPKeyPair, cfg *ProviderConfig) (*oauth2.Token, error) {
	conf := OAuth2Config(metadata, cfg)
	
	// Use a custom transport that adds both PKCE and DPoP with nonce retry
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: &DPoPPKCETransport{
			Base:         http.DefaultTransport,
			CodeVerifier: codeVerifier,
			DPoPKey:      dpopKey,
			TargetURL:    metadata.TokenEndpoint,
			Config:       cfg,
			Metadata:     metadata,
		},
	})
	return conf.Exchange(ctx, code)
}