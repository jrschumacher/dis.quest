package session

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/jrschumacher/dis.quest/pkg/atproto/xrpc"
)

// DefaultSession implements the Session interface.
type DefaultSession struct {
	manager *DefaultManager
	data    *Data
}

// GetSessionID returns the session identifier.
func (s *DefaultSession) GetSessionID() string {
	return s.data.SessionID
}

// GetUserDID returns the authenticated user's DID.
func (s *DefaultSession) GetUserDID() string {
	return s.data.UserDID
}

// GetHandle returns the authenticated user's handle.
func (s *DefaultSession) GetHandle() string {
	return s.data.Handle
}

// GetAccessToken returns the current access token.
func (s *DefaultSession) GetAccessToken() string {
	return s.data.AccessToken
}

// GetRefreshToken returns the current refresh token.
func (s *DefaultSession) GetRefreshToken() string {
	return s.data.RefreshToken
}

// GetDPoPKey returns the DPoP private key.
func (s *DefaultSession) GetDPoPKey() *ecdsa.PrivateKey {
	return s.data.DPoPKey
}

// IsExpired checks if the session is expired.
func (s *DefaultSession) IsExpired() bool {
	return s.manager.IsTokenExpired(s.data.AccessToken, s.manager.config.TokenExpiryThreshold)
}

// Refresh refreshes the session using the refresh token.
func (s *DefaultSession) Refresh(ctx context.Context) error {
	return s.manager.RefreshSession(ctx, s)
}

// UpdateTokens updates the session with new tokens.
func (s *DefaultSession) UpdateTokens(accessToken, refreshToken string, expiresIn int64) error {
	s.data.AccessToken = accessToken
	s.data.RefreshToken = refreshToken
	s.data.ExpiresAt = s.data.UpdatedAt.Add(time.Duration(expiresIn) * time.Second)
	return nil
}

// Save persists the session to storage.
func (s *DefaultSession) Save(ctx context.Context) error {
	return s.manager.SaveSession(ctx, s)
}

// Delete removes the session from storage.
func (s *DefaultSession) Delete(ctx context.Context) error {
	return s.manager.DeleteSession(ctx, s.data.SessionID)
}

// CreateRecord creates a new record in the user's PDS.
func (s *DefaultSession) CreateRecord(ctx context.Context, collection, rkey string, record interface{}) (*RecordResult, error) {
	client := s.getXRPCClient()
	if client == nil {
		return nil, fmt.Errorf("XRPC client not available")
	}

	resp, err := client.CreateRecord(ctx, s.data.UserDID, collection, rkey, record, s.data.AccessToken, s.data.DPoPKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	return &RecordResult{
		URI: resp.URI,
		CID: resp.CID,
	}, nil
}

// GetRecord retrieves a record from the user's PDS.
func (s *DefaultSession) GetRecord(ctx context.Context, collection, rkey string, result interface{}) error {
	client := s.getXRPCClient()
	if client == nil {
		return fmt.Errorf("XRPC client not available")
	}

	return client.GetRecord(ctx, s.data.UserDID, collection, rkey, result, s.data.AccessToken, s.data.DPoPKey)
}

// ListRecords lists records from the user's PDS.
func (s *DefaultSession) ListRecords(ctx context.Context, collection string, limit int, cursor string) (*ListRecordsResult, error) {
	client := s.getXRPCClient()
	if client == nil {
		return nil, fmt.Errorf("XRPC client not available")
	}

	resp, err := client.ListRecords(ctx, s.data.UserDID, collection, limit, cursor, s.data.AccessToken, s.data.DPoPKey)
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	// Convert xrpc response to session response format
	records := make([]Record, len(resp.Records))
	for i, record := range resp.Records {
		records[i] = Record{
			URI:   record.URI,
			CID:   record.CID,
			Value: record.Value,
		}
	}

	return &ListRecordsResult{
		Records: records,
		Cursor:  resp.Cursor,
	}, nil
}

// UpdateRecord updates an existing record in the user's PDS.
func (s *DefaultSession) UpdateRecord(ctx context.Context, collection, rkey string, record interface{}) (*RecordResult, error) {
	client := s.getXRPCClient()
	if client == nil {
		return nil, fmt.Errorf("XRPC client not available")
	}

	resp, err := client.UpdateRecord(ctx, s.data.UserDID, collection, rkey, record, s.data.AccessToken, s.data.DPoPKey)
	if err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	return &RecordResult{
		URI: resp.URI,
		CID: resp.CID,
	}, nil
}

// DeleteRecord deletes a record from the user's PDS.
func (s *DefaultSession) DeleteRecord(ctx context.Context, collection, rkey string) error {
	client := s.getXRPCClient()
	if client == nil {
		return fmt.Errorf("XRPC client not available")
	}

	return client.DeleteRecord(ctx, s.data.UserDID, collection, rkey, s.data.AccessToken, s.data.DPoPKey)
}

// GetData returns the session data.
func (s *DefaultSession) GetData() *Data {
	return s.data
}

// GetMetadata returns a metadata value by key.
func (s *DefaultSession) GetMetadata(key string) interface{} {
	if s.data.Metadata == nil {
		return nil
	}
	return s.data.Metadata[key]
}

// SetMetadata sets a metadata value.
func (s *DefaultSession) SetMetadata(key string, value interface{}) {
	if s.data.Metadata == nil {
		s.data.Metadata = make(map[string]interface{})
	}
	s.data.Metadata[key] = value
}

// getXRPCClient returns the XRPC client for ATProtocol operations.
func (s *DefaultSession) getXRPCClient() *xrpc.Client {
	return s.manager.xrpcClient
}