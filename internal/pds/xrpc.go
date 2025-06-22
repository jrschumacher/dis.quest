// Package pds provides reusable ATProtocol XRPC client abstractions
package pds

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/pkg/atproto/xrpc"
)

// XRPCClient provides reusable ATProtocol XRPC operations
// This is a wrapper around pkg/atproto/xrpc.Client to maintain compatibility
type XRPCClient struct {
	client *xrpc.Client
}

// NewXRPCClient creates a new XRPC client
func NewXRPCClient() *XRPCClient {
	return &XRPCClient{
		client: xrpc.NewClient(),
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
	components, err := xrpc.ParseATUri(uri)
	if err != nil {
		return nil, err
	}
	return &ATUriComponents{
		DID:        components.DID,
		Collection: components.Collection,
		RKey:       components.RKey,
	}, nil
}

// ResolvePDS resolves the PDS endpoint for a given DID
func (c *XRPCClient) ResolvePDS(did string) (string, error) {
	// Create a resolver since the client doesn't expose ResolvePDS directly
	resolver := xrpc.NewDIDResolver()
	return resolver.ResolvePDS(did)
}

// CreateRecordRequest represents the request body for com.atproto.repo.createRecord
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
	// Convert interface{} dpopKey to *ecdsa.PrivateKey to maintain compatibility
	var ecdsaKey *ecdsa.PrivateKey
	if dpopKey != nil {
		var ok bool
		ecdsaKey, ok = dpopKey.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("DPoP key is not the correct type: %T", dpopKey)
		}
	}

	// Log the request details to maintain debug compatibility
	logger.Info("Making XRPC createRecord request", 
		"repo", req.Repo,
		"collection", req.Collection,
		"rkey", req.RKey,
		"validate", req.Validate,
		"hasAuth", accessToken != "",
		"hasDPoP", ecdsaKey != nil)

	// Use the underlying pkg/atproto/xrpc client
	response, err := c.client.CreateRecord(ctx, req.Repo, req.Collection, req.RKey, req.Record, accessToken, ecdsaKey)
	if err != nil {
		return nil, err
	}

	// Convert response to maintain compatibility
	result := &CreateRecordResponse{
		URI: response.URI,
		CID: response.CID,
	}

	logger.Info("Successfully created record in PDS", "uri", result.URI, "cid", result.CID)
	return result, nil
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
	// Convert interface{} dpopKey to *ecdsa.PrivateKey to maintain compatibility
	var ecdsaKey *ecdsa.PrivateKey
	if dpopKey != nil {
		var ok bool
		ecdsaKey, ok = dpopKey.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("DPoP key is not the correct type: %T", dpopKey)
		}
	}

	// Log the request details to maintain debug compatibility
	logger.Info("Making XRPC getRecord request", 
		"repo", repo,
		"collection", collection,
		"rkey", rkey,
		"hasAuth", accessToken != "",
		"hasDPoP", ecdsaKey != nil)

	// The pkg/atproto client GetRecord method expects a result parameter to unmarshal into,
	// but our interface returns the raw response. We'll get it as a map and convert.
	var rawValue interface{}
	err := c.client.GetRecord(ctx, repo, collection, rkey, &rawValue, accessToken, ecdsaKey)
	if err != nil {
		// Check if it's a "not found" error to maintain compatibility
		if strings.Contains(err.Error(), "record not found") {
			return nil, fmt.Errorf("record not found: %s/%s/%s", repo, collection, rkey)
		}
		return nil, err
	}

	// For compatibility, we need to reconstruct the URI and CID
	// The underlying client doesn't return these, so we construct them
	uri := fmt.Sprintf("at://%s/%s/%s", repo, collection, rkey)
	
	// Convert response to maintain compatibility
	result := &GetRecordResponse{
		URI: uri,
		CID: "", // CID is not returned by the simplified interface
	}
	
	// Convert rawValue to map[string]interface{} for compatibility
	if rawValue != nil {
		if vm, ok := rawValue.(map[string]interface{}); ok {
			result.Value = vm
		} else {
			// Try to marshal and unmarshal to convert
			if jsonBytes, marshalErr := json.Marshal(rawValue); marshalErr == nil {
				var valueMap map[string]interface{}
				if unmarshalErr := json.Unmarshal(jsonBytes, &valueMap); unmarshalErr == nil {
					result.Value = valueMap
				}
			}
		}
	}

	return result, nil
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
	// For now, return an error since putRecord is not implemented in pkg/atproto client
	return nil, fmt.Errorf("PutRecord not yet implemented in pkg/atproto client")
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
	return c.ListRecordsWithDPoP(ctx, repo, collection, limit, cursor, "", accessToken, nil)
}

// ListRecordsWithDPoP lists records from a repository with DPoP authentication
func (c *XRPCClient) ListRecordsWithDPoP(ctx context.Context, repo, collection string, limit int, cursor, rvKey, accessToken string, dpopKey interface{}) (*ListRecordsResponse, error) {
	// Convert interface{} dpopKey to *ecdsa.PrivateKey to maintain compatibility
	var ecdsaKey *ecdsa.PrivateKey
	if dpopKey != nil {
		var ok bool
		ecdsaKey, ok = dpopKey.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("DPoP key is not the correct type: %T", dpopKey)
		}
	}

	// Log the request details to maintain debug compatibility
	logger.Info("Making XRPC listRecords request", 
		"repo", repo,
		"collection", collection,
		"limit", limit,
		"cursor", cursor,
		"rvKey", rvKey,
		"hasAuth", accessToken != "",
		"hasDPoP", ecdsaKey != nil)

	// Note: The pkg/atproto client doesn't support rvKey (rkeyStart) parameter,
	// so we ignore it for now. This is a limitation but maintains basic functionality.
	if rvKey != "" {
		logger.Warn("rvKey parameter not supported by pkg/atproto client, ignoring", "rvKey", rvKey)
	}

	// Use the underlying pkg/atproto/xrpc client
	response, err := c.client.ListRecords(ctx, repo, collection, limit, cursor, accessToken, ecdsaKey)
	if err != nil {
		return nil, err
	}

	// Convert response to maintain compatibility
	// The pkg/atproto client returns xrpc.ListRecordsResponse, we need pds.ListRecordsResponse
	result := &ListRecordsResponse{
		Records: make([]struct {
			URI   string                 `json:"uri"`
			CID   string                 `json:"cid"`
			Value map[string]interface{} `json:"value"`
		}, len(response.Records)),
		Cursor: response.Cursor,
	}

	// Convert each record
	for i, record := range response.Records {
		// Convert interface{} to map[string]interface{} for compatibility
		var valueMap map[string]interface{}
		if record.Value != nil {
			if vm, ok := record.Value.(map[string]interface{}); ok {
				valueMap = vm
			} else {
				// Try to marshal and unmarshal to convert
				if jsonBytes, marshalErr := json.Marshal(record.Value); marshalErr == nil {
					if unmarshalErr := json.Unmarshal(jsonBytes, &valueMap); unmarshalErr != nil {
						logger.Debug("Failed to unmarshal record value", "error", unmarshalErr)
					}
				}
			}
		}

		result.Records[i] = struct {
			URI   string                 `json:"uri"`
			CID   string                 `json:"cid"`
			Value map[string]interface{} `json:"value"`
		}{
			URI:   record.URI,
			CID:   record.CID,
			Value: valueMap,
		}
	}

	return result, nil
}