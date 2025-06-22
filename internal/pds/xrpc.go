// Package pds provides reusable ATProtocol XRPC client abstractions
package pds

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/logger"
)

// XRPCClient provides reusable ATProtocol XRPC operations
type XRPCClient struct {
	client *http.Client
}

// NewXRPCClient creates a new XRPC client
func NewXRPCClient() *XRPCClient {
	return &XRPCClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ATUriComponents represents parsed components of an AT URI
type ATUriComponents struct {
	DID        string
	Collection string
	RKey       string
}

// ParseATUri parses an AT URI into its components
func ParseATUri(uri string) (*ATUriComponents, error) {
	if !strings.HasPrefix(uri, "at://") {
		return nil, fmt.Errorf("invalid AT URI format: %s", uri)
	}

	// Remove at:// prefix and split
	parts := strings.Split(strings.TrimPrefix(uri, "at://"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid AT URI format: %s", uri)
	}

	components := &ATUriComponents{
		DID:        parts[0],
		Collection: parts[1],
	}

	if len(parts) >= 3 {
		components.RKey = parts[2]
	}

	return components, nil
}

// ResolvePDS resolves the PDS endpoint for a given DID
func (c *XRPCClient) ResolvePDS(did string) (string, error) {
	// Resolve DID to get actual PDS endpoint
	if strings.HasPrefix(did, "did:plc:") {
		return c.resolvePlcDID(did)
	}
	if strings.HasPrefix(did, "did:web:") {
		// Extract domain from did:web
		domain := strings.TrimPrefix(did, "did:web:")
		return fmt.Sprintf("https://%s", domain), nil
	}
	return "", fmt.Errorf("unsupported DID method: %s", did)
}

// CreateRecordRequest represents the request body for com.atproto.repo.createRecord
// resolvePlcDID resolves a did:plc DID to get the PDS endpoint
func (c *XRPCClient) resolvePlcDID(did string) (string, error) {
	// Query the PLC directory
	plcURL := fmt.Sprintf("https://plc.directory/%s", did)
	
	resp, err := c.client.Get(plcURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve DID %s: %w", did, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to resolve DID %s: status %d", did, resp.StatusCode)
	}
	
	var didDoc struct {
		Service []struct {
			Type            string `json:"type"`
			ServiceEndpoint string `json:"serviceEndpoint"`
		} `json:"service"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&didDoc); err != nil {
		return "", fmt.Errorf("failed to decode DID document for %s: %w", did, err)
	}
	
	// Find the AtprotoPersonalDataServer service
	for _, service := range didDoc.Service {
		if service.Type == "AtprotoPersonalDataServer" {
			return service.ServiceEndpoint, nil
		}
	}
	
	return "", fmt.Errorf("no PDS endpoint found in DID document for %s", did)
}

type CreateRecordRequest struct {
	Repo       string                 `json:"repo"`
	Collection string                 `json:"collection"`
	RKey       string                 `json:"rkey,omitempty"`
	Validate   bool                   `json:"validate,omitempty"`
	Record     map[string]interface{} `json:"record"`
}

// CreateRecordResponse represents the response from com.atproto.repo.createRecord
type CreateRecordResponse struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

// CreateRecord creates a record in a repository using com.atproto.repo.createRecord
func (c *XRPCClient) CreateRecord(ctx context.Context, req CreateRecordRequest, accessToken string) (*CreateRecordResponse, error) {
	return c.CreateRecordWithDPoP(ctx, req, accessToken, nil)
}

// CreateRecordWithDPoP creates a record with DPoP authentication with nonce retry support
func (c *XRPCClient) CreateRecordWithDPoP(ctx context.Context, req CreateRecordRequest, accessToken string, dpopKey interface{}) (*CreateRecordResponse, error) {
	pdsEndpoint, err := c.ResolvePDS(req.Repo)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS: %w", err)
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.createRecord", pdsEndpoint)
	
	// Helper function to make request with optional nonce
	makeRequest := func(nonce string) (*http.Response, error) {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		if accessToken != "" {
			httpReq.Header.Set("Authorization", fmt.Sprintf("DPoP %s", accessToken))
		}
		
		// Add DPoP header if DPoP key is provided
		if dpopKey != nil {
			if ecdsaKey, ok := dpopKey.(*ecdsa.PrivateKey); ok {
				logger.Info("Creating DPoP JWT", "method", "POST", "url", url, "hasNonce", nonce != "")
				
				// Debug: Log the DPoP key details
				if ecdsaKey.PublicKey.X != nil && ecdsaKey.PublicKey.Y != nil {
					logger.Info("DPoP key details", 
						"keyX", ecdsaKey.PublicKey.X.String()[:10]+"...",
						"keyY", ecdsaKey.PublicKey.Y.String()[:10]+"...")
				}
				
				var dpopJWT string
				dpopJWT, err = auth.CreateDPoPJWTWithAccessToken(ecdsaKey, "POST", url, nonce, accessToken)
				
				if err != nil {
					logger.Error("Failed to create DPoP JWT", "error", err)
					return nil, fmt.Errorf("failed to create DPoP JWT: %w", err)
				}
				
				// Debug: Log the DPoP JWT parts
				jwtParts := strings.Split(dpopJWT, ".")
				if len(jwtParts) == 3 {
					logger.Info("DPoP JWT created", 
						"headerLength", len(jwtParts[0]),
						"payloadLength", len(jwtParts[1]), 
						"signatureLength", len(jwtParts[2]),
						"nonce", nonce)
				}
				
				httpReq.Header.Set("DPoP", dpopJWT)
				logger.Info("Added DPoP header to request", "jwtLength", len(dpopJWT), "nonce", nonce)
			} else {
				logger.Error("DPoP key is not the correct type", "type", fmt.Sprintf("%T", dpopKey))
			}
		} else {
			logger.Error("No DPoP key provided to XRPC request")
		}

		// Log request details
		logger.Info("Making XRPC createRecord request", 
			"method", httpReq.Method,
			"url", httpReq.URL.String(),
			"hasAuth", httpReq.Header.Get("Authorization") != "",
			"hasDPoP", httpReq.Header.Get("DPoP") != "",
			"contentType", httpReq.Header.Get("Content-Type"),
			"nonce", nonce)

		return c.client.Do(httpReq)
	}

	// First attempt without nonce
	resp, err := makeRequest("")
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("failed to close response body", "error", err)
		}
	}()

	// Check for DPoP nonce requirement (can be 400 Bad Request or 401 Unauthorized)
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		// Read response to check for nonce requirement
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("PDS request failed with status: %d (unable to read error details: %v)", resp.StatusCode, readErr)
		}
		
		// Check if error indicates DPoP nonce is needed
		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			logger.Info("Checking for DPoP nonce error", "error", errorResp["error"], "message", errorResp["message"])
			if errorResp["error"] == "use_dpop_nonce" || 
			   strings.Contains(fmt.Sprintf("%v", errorResp["message"]), "nonce") {
				// Get nonce from DPoP-Nonce header
				if dpopNonce := resp.Header.Get("DPoP-Nonce"); dpopNonce != "" {
					logger.Info("Retrying createRecord with DPoP nonce", "nonce", dpopNonce)
					
					// Close the first response and retry with nonce
					resp.Body.Close()
					retryResp, retryErr := makeRequest(dpopNonce)
					if retryErr != nil {
						return nil, fmt.Errorf("failed to retry request with nonce: %w", retryErr)
					}
					
					// Use the retry response for the rest of the function
					resp = retryResp
					defer func() {
						if err := resp.Body.Close(); err != nil {
							logger.Error("failed to close retry response body", "error", err)
						}
					}()
				} else {
					logger.Error("DPoP nonce required but not provided in response header")
					return nil, fmt.Errorf("PDS request failed with status: %d, response: %s", resp.StatusCode, string(body))
				}
			} else {
				// Not a nonce error, return original error
				return nil, fmt.Errorf("PDS request failed with status: %d, response: %s", resp.StatusCode, string(body))
			}
		} else {
			// Failed to parse error response
			return nil, fmt.Errorf("PDS request failed with status: %d, response: %s", resp.StatusCode, string(body))
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Read error response body for detailed error information
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("PDS request failed with status: %d (unable to read error details: %v)", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("PDS request failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	var response CreateRecordResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Info("Successfully created record in PDS", "uri", response.URI, "cid", response.CID)
	return &response, nil
}

// GetRecordResponse represents the response from com.atproto.repo.getRecord
type GetRecordResponse struct {
	URI   string                 `json:"uri"`
	CID   string                 `json:"cid"`
	Value map[string]interface{} `json:"value"`
}

// GetRecord retrieves a record using com.atproto.repo.getRecord
func (c *XRPCClient) GetRecord(ctx context.Context, repo, collection, rkey string, accessToken string) (*GetRecordResponse, error) {
	return c.GetRecordWithDPoP(ctx, repo, collection, rkey, accessToken, nil)
}

// GetRecordWithDPoP retrieves a record with DPoP authentication
func (c *XRPCClient) GetRecordWithDPoP(ctx context.Context, repo, collection, rkey string, accessToken string, dpopKey interface{}) (*GetRecordResponse, error) {
	pdsEndpoint, err := c.ResolvePDS(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS: %w", err)
	}

	// Build query parameters
	params := url.Values{}
	params.Set("repo", repo)
	params.Set("collection", collection)
	params.Set("rkey", rkey)

	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?%s", pdsEndpoint, params.Encode())
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if accessToken != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("DPoP %s", accessToken))
	}
	
	// Add DPoP header if DPoP key is provided
	if dpopKey != nil {
		if ecdsaKey, ok := dpopKey.(*ecdsa.PrivateKey); ok {
			dpopJWT, err := auth.CreateDPoPJWTWithAccessToken(ecdsaKey, "GET", url, "", accessToken)
			if err != nil {
				return nil, fmt.Errorf("failed to create DPoP JWT: %w", err)
			}
			httpReq.Header.Set("DPoP", dpopJWT)
		}
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("record not found: %s/%s/%s", repo, collection, rkey)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PDS request failed with status: %d", resp.StatusCode)
	}

	var response GetRecordResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// PutRecordRequest represents the request body for com.atproto.repo.putRecord
type PutRecordRequest struct {
	Repo       string                 `json:"repo"`
	Collection string                 `json:"collection"`
	RKey       string                 `json:"rkey"`
	Validate   bool                   `json:"validate,omitempty"`
	Record     map[string]interface{} `json:"record"`
	SwapRecord string                 `json:"swapRecord,omitempty"`
}

// PutRecordResponse represents the response from com.atproto.repo.putRecord
type PutRecordResponse struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

// PutRecord updates a record using com.atproto.repo.putRecord
func (c *XRPCClient) PutRecord(ctx context.Context, req PutRecordRequest, accessToken string) (*PutRecordResponse, error) {
	pdsEndpoint, err := c.ResolvePDS(req.Repo)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS: %w", err)
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.putRecord", pdsEndpoint)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("DPoP %s", accessToken))
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PDS request failed with status: %d", resp.StatusCode)
	}

	var response PutRecordResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// ListRecordsResponse represents the response from com.atproto.repo.listRecords
type ListRecordsResponse struct {
	Records []struct {
		URI   string                 `json:"uri"`
		CID   string                 `json:"cid"`
		Value map[string]interface{} `json:"value"`
	} `json:"records"`
	Cursor string `json:"cursor,omitempty"`
}

// ListRecords lists records in a collection using com.atproto.repo.listRecords
func (c *XRPCClient) ListRecords(ctx context.Context, repo, collection string, limit int, cursor string, accessToken string) (*ListRecordsResponse, error) {
	pdsEndpoint, err := c.ResolvePDS(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS: %w", err)
	}

	params := url.Values{}
	params.Set("repo", repo)
	params.Set("collection", collection)
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?%s", pdsEndpoint, params.Encode())
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if accessToken != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("DPoP %s", accessToken))
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PDS request failed with status: %d", resp.StatusCode)
	}

	var response ListRecordsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}