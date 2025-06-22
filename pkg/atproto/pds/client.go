// Package pds provides a generic Personal Data Server (PDS) client for ATProtocol.
// This client is lexicon-agnostic and can work with any ATProtocol records.
package pds

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/jrschumacher/dis.quest/pkg/atproto/xrpc"
)

// Client provides generic PDS operations for any ATProtocol application
type Client struct {
	xrpc *xrpc.Client
}

// NewClient creates a new generic PDS client
func NewClient() *Client {
	return &Client{
		xrpc: xrpc.NewClient(),
	}
}

// Record represents a generic ATProtocol record
type Record struct {
	URI       string                 `json:"uri"`
	CID       string                 `json:"cid"`
	Value     map[string]interface{} `json:"value"`
	CreatedAt time.Time              `json:"createdAt,omitempty"`
}

// CreateRecordParams contains parameters for creating a record
type CreateRecordParams struct {
	Collection string                 `json:"collection"`
	RKey       string                 `json:"rkey,omitempty"` // Optional record key
	Record     map[string]interface{} `json:"record"`
}

// CreateRecordResult represents the result of creating a record
type CreateRecordResult struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

// ListRecordsParams contains parameters for listing records
type ListRecordsParams struct {
	Collection string `json:"collection"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
	Reverse    bool   `json:"reverse,omitempty"`
}

// ListRecordsResult represents the result of listing records
type ListRecordsResult struct {
	Records []Record `json:"records"`
	Cursor  string   `json:"cursor,omitempty"`
}

// CreateRecord creates a new record in the user's PDS
func (c *Client) CreateRecord(ctx context.Context, userDID string, accessToken string, dpopKey *ecdsa.PrivateKey, params CreateRecordParams) (*CreateRecordResult, error) {
	// Use the XRPC client with correct parameter order
	response, err := c.xrpc.CreateRecord(ctx, userDID, params.Collection, params.RKey, params.Record, accessToken, dpopKey)
	if err != nil {
		return nil, err
	}
	
	return &CreateRecordResult{
		URI: response.URI,
		CID: response.CID,
	}, nil
}

// GetRecord retrieves a record by URI
func (c *Client) GetRecord(ctx context.Context, userDID string, accessToken string, dpopKey *ecdsa.PrivateKey, uri string) (*Record, error) {
	// Parse URI to get components
	components, err := xrpc.ParseATUri(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid AT URI: %w", err)
	}

	// Get the record - using the XRPC interface that returns into a result interface
	var recordValue map[string]interface{}
	err = c.xrpc.GetRecord(ctx, components.DID, components.Collection, components.RKey, &recordValue, accessToken, dpopKey)
	if err != nil {
		return nil, err
	}

	// Convert to our Record format
	result := &Record{
		URI:   uri,
		Value: recordValue,
	}

	// Extract createdAt if present
	if createdAtStr, ok := recordValue["createdAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			result.CreatedAt = t
		}
	}

	return result, nil
}

// ListRecords lists records from a collection
func (c *Client) ListRecords(ctx context.Context, userDID string, accessToken string, dpopKey *ecdsa.PrivateKey, params ListRecordsParams) (*ListRecordsResult, error) {
	// Set default limit if not specified
	if params.Limit == 0 {
		params.Limit = 50
	}

	response, err := c.xrpc.ListRecords(ctx, userDID, params.Collection, params.Limit, params.Cursor, accessToken, dpopKey)
	if err != nil {
		return nil, err
	}

	// Convert to our Record format
	var result []Record
	for _, record := range response.Records {
		// Type assert the record value to map
		recordValue, ok := record.Value.(map[string]interface{})
		if !ok {
			continue // Skip records that don't have the expected format
		}
		
		rec := Record{
			URI:   record.URI,
			CID:   record.CID,
			Value: recordValue,
		}

		// Extract createdAt if present
		if createdAtStr, ok := recordValue["createdAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
				rec.CreatedAt = t
			}
		}

		result = append(result, rec)
	}

	return &ListRecordsResult{
		Records: result,
		Cursor:  response.Cursor,
	}, nil
}

// UpdateRecord updates an existing record
func (c *Client) UpdateRecord(ctx context.Context, userDID string, accessToken string, dpopKey *ecdsa.PrivateKey, uri string, record map[string]interface{}) (*CreateRecordResult, error) {
	// Parse URI to get components
	components, err := xrpc.ParseATUri(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid AT URI: %w", err)
	}

	response, err := c.xrpc.UpdateRecord(ctx, components.DID, components.Collection, components.RKey, record, accessToken, dpopKey)
	if err != nil {
		return nil, err
	}
	
	return &CreateRecordResult{
		URI: response.URI,
		CID: response.CID,
	}, nil
}

// DeleteRecord deletes a record by URI
func (c *Client) DeleteRecord(ctx context.Context, userDID string, accessToken string, dpopKey *ecdsa.PrivateKey, uri string) error {
	// Parse URI to get components
	components, err := xrpc.ParseATUri(uri)
	if err != nil {
		return fmt.Errorf("invalid AT URI: %w", err)
	}

	return c.xrpc.DeleteRecord(ctx, components.DID, components.Collection, components.RKey, accessToken, dpopKey)
}

// ResolvePDS resolves the PDS endpoint for a given DID
func (c *Client) ResolvePDS(ctx context.Context, did string) (string, error) {
	resolver := xrpc.NewDIDResolver()
	return resolver.ResolvePDS(did)
}