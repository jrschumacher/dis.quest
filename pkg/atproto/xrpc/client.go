// Package xrpc provides XRPC client functionality for ATProtocol Personal Data Servers
package xrpc

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

	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
)

// Client provides ATProtocol XRPC operations with DPoP support
type Client struct {
	httpClient *http.Client
	resolver   *DIDResolver
}

// NewClient creates a new XRPC client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		resolver: NewDIDResolver(),
	}
}

// CreateRecordRequest represents a request to create a record
type CreateRecordRequest struct {
	Repo       string      `json:"repo"`
	Collection string      `json:"collection"`
	RKey       string      `json:"rkey,omitempty"`
	Validate   bool        `json:"validate,omitempty"`
	Record     interface{} `json:"record"`
}

// RecordResponse represents the response from record operations
type RecordResponse struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

// GetRecordResponse represents the response from getting a record
type GetRecordResponse struct {
	URI   string      `json:"uri"`
	CID   string      `json:"cid"`
	Value interface{} `json:"value"`
}

// ListRecordsResponse represents the response from listing records
type ListRecordsResponse struct {
	Records []struct {
		URI   string      `json:"uri"`
		CID   string      `json:"cid"`
		Value interface{} `json:"value"`
	} `json:"records"`
	Cursor string `json:"cursor,omitempty"`
}

// CreateRecord creates a new record in the specified repository
func (c *Client) CreateRecord(ctx context.Context, repo, collection, rkey string, record interface{}, accessToken string, dpopKey *ecdsa.PrivateKey) (*RecordResponse, error) {
	req := CreateRecordRequest{
		Repo:       repo,
		Collection: collection,
		RKey:       rkey,
		Validate:   false, // Set to false for custom lexicons
		Record:     record,
	}

	return c.createRecordWithDPoP(ctx, req, accessToken, dpopKey)
}

// GetRecord retrieves a record from the specified repository
func (c *Client) GetRecord(ctx context.Context, repo, collection, rkey string, result interface{}, accessToken string, dpopKey *ecdsa.PrivateKey) error {
	response, err := c.getRecordWithDPoP(ctx, repo, collection, rkey, accessToken, dpopKey)
	if err != nil {
		return err
	}

	// Convert the response value to the desired type
	if result != nil {
		valueBytes, err := json.Marshal(response.Value)
		if err != nil {
			return fmt.Errorf("failed to marshal record value: %w", err)
		}
		if err := json.Unmarshal(valueBytes, result); err != nil {
			return fmt.Errorf("failed to unmarshal record value: %w", err)
		}
	}

	return nil
}

// ListRecords lists records from a collection
func (c *Client) ListRecords(ctx context.Context, repo, collection string, limit int, cursor, accessToken string, dpopKey *ecdsa.PrivateKey) (*ListRecordsResponse, error) {
	return c.listRecordsWithDPoP(ctx, repo, collection, limit, cursor, accessToken, dpopKey)
}

// UpdateRecord updates an existing record
func (c *Client) UpdateRecord(ctx context.Context, repo, collection, rkey string, record interface{}, accessToken string, dpopKey *ecdsa.PrivateKey) (*RecordResponse, error) {
	// ATProtocol uses putRecord for updates
	return c.putRecordWithDPoP(ctx, repo, collection, rkey, record, accessToken, dpopKey)
}

// DeleteRecord deletes a record
func (c *Client) DeleteRecord(ctx context.Context, repo, collection, rkey string, accessToken string, dpopKey *ecdsa.PrivateKey) error {
	return c.deleteRecordWithDPoP(ctx, repo, collection, rkey, accessToken, dpopKey)
}

// createRecordWithDPoP creates a record with DPoP authentication and nonce retry support
func (c *Client) createRecordWithDPoP(ctx context.Context, req CreateRecordRequest, accessToken string, dpopKey *ecdsa.PrivateKey) (*RecordResponse, error) {
	pdsEndpoint, err := c.resolver.ResolvePDS(req.Repo)
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
			keyPair := &oauth.DPoPKeyPair{PrivateKey: dpopKey}
			dpopJWT, err := keyPair.CreateDPoPJWTWithAccessToken("POST", url, nonce, accessToken)
			if err != nil {
				return nil, fmt.Errorf("failed to create DPoP JWT: %w", err)
			}
			httpReq.Header.Set("DPoP", dpopJWT)
		}

		return c.httpClient.Do(httpReq)
	}

	// First attempt without nonce
	resp, err := makeRequest("")
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check for DPoP nonce requirement
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("PDS request failed with status: %d (unable to read error details: %v)", resp.StatusCode, readErr)
		}

		// Check if error indicates DPoP nonce is needed
		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			if errorResp["error"] == "use_dpop_nonce" ||
				strings.Contains(fmt.Sprintf("%v", errorResp["message"]), "nonce") {
				// Get nonce from DPoP-Nonce header
				if dpopNonce := resp.Header.Get("DPoP-Nonce"); dpopNonce != "" {
					// Close the first response and retry with nonce
					resp.Body.Close()
					retryResp, retryErr := makeRequest(dpopNonce)
					if retryErr != nil {
						return nil, fmt.Errorf("failed to retry request with nonce: %w", retryErr)
					}
					resp = retryResp
					defer resp.Body.Close()
				} else {
					return nil, fmt.Errorf("DPoP nonce required but not provided in response header")
				}
			} else {
				return nil, fmt.Errorf("PDS request failed with status: %d, response: %s", resp.StatusCode, string(body))
			}
		} else {
			return nil, fmt.Errorf("PDS request failed with status: %d, response: %s", resp.StatusCode, string(body))
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("PDS request failed with status: %d (unable to read error details: %v)", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("PDS request failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	var response RecordResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// getRecordWithDPoP retrieves a record with DPoP authentication
func (c *Client) getRecordWithDPoP(ctx context.Context, repo, collection, rkey string, accessToken string, dpopKey *ecdsa.PrivateKey) (*GetRecordResponse, error) {
	pdsEndpoint, err := c.resolver.ResolvePDS(repo)
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
		keyPair := &oauth.DPoPKeyPair{PrivateKey: dpopKey}
		dpopJWT, err := keyPair.CreateDPoPJWTWithAccessToken("GET", url, "", accessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create DPoP JWT: %w", err)
		}
		httpReq.Header.Set("DPoP", dpopJWT)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

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

// listRecordsWithDPoP lists records with DPoP authentication
func (c *Client) listRecordsWithDPoP(ctx context.Context, repo, collection string, limit int, cursor, accessToken string, dpopKey *ecdsa.PrivateKey) (*ListRecordsResponse, error) {
	pdsEndpoint, err := c.resolver.ResolvePDS(repo)
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

	reqURL := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?%s", pdsEndpoint, params.Encode())

	// Create request with DPoP nonce retry support
	makeRequest := func(nonce string) (*http.Response, error) {
		httpReq, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Authorization", "DPoP "+accessToken)

		// Create and add DPoP header
		keyPair := &oauth.DPoPKeyPair{PrivateKey: dpopKey}
		dpopHeader, err := keyPair.CreateDPoPJWTWithAccessToken("GET", reqURL, nonce, accessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create DPoP header: %w", err)
		}
		httpReq.Header.Set("DPoP", dpopHeader)

		return c.httpClient.Do(httpReq)
	}

	// First attempt without nonce
	resp, err := makeRequest("")
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check for DPoP nonce requirement
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("PDS request failed with status: %d (unable to read error details: %v)", resp.StatusCode, readErr)
		}

		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			if errorResp["error"] == "use_dpop_nonce" ||
				strings.Contains(fmt.Sprintf("%v", errorResp["message"]), "nonce") {
				if dpopNonce := resp.Header.Get("DPoP-Nonce"); dpopNonce != "" {
					resp.Body.Close()
					retryResp, retryErr := makeRequest(dpopNonce)
					if retryErr != nil {
						return nil, fmt.Errorf("failed to retry request with nonce: %w", retryErr)
					}
					resp = retryResp
					defer resp.Body.Close()
				}
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PDS request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var response ListRecordsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// Placeholder implementations for putRecord and deleteRecord
func (c *Client) putRecordWithDPoP(ctx context.Context, repo, collection, rkey string, record interface{}, accessToken string, dpopKey *ecdsa.PrivateKey) (*RecordResponse, error) {
	// Implementation similar to createRecord but for PUT
	return nil, fmt.Errorf("putRecord not yet implemented")
}

func (c *Client) deleteRecordWithDPoP(ctx context.Context, repo, collection, rkey string, accessToken string, dpopKey *ecdsa.PrivateKey) error {
	// Implementation for DELETE operation
	return fmt.Errorf("deleteRecord not yet implemented")
}