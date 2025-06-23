// Package lexicons provides temporary stubs for dev functionality
package lexicons

import (
	"fmt"
	
	"github.com/jrschumacher/dis.quest/pkg/atproto/xrpc"
)

// Temporary stubs for functions used in dev.go that don't exist in the new structure
// These should be migrated to proper implementations over time

// ValidateLexicon is a stub for lexicon validation
func ValidateLexicon(lexicon string, record map[string]interface{}) error {
	// TODO: Implement proper lexicon validation
	// For now, just check basic required fields
	if lexicon == TopicLexicon {
		required := []string{"$type", "title", "createdBy", "createdAt"}
		for _, field := range required {
			if _, ok := record[field]; !ok {
				return fmt.Errorf("missing required field: %s", field)
			}
		}
	}
	return nil
}

// GenerateRKey generates a record key (placeholder implementation)
func GenerateRKey(prefix string) string {
	// TODO: Implement proper record key generation
	// For now, return a simple placeholder
	return fmt.Sprintf("%s-%d", prefix, 123456789)
}

// NewXRPCClient creates a stub XRPC client (placeholder)
func NewXRPCClient() *xrpc.Client {
	// TODO: Replace with proper XRPC client from pkg/atproto/xrpc
	return xrpc.NewClient()
}

// CreateRecordRequest is a placeholder type
type CreateRecordRequest struct {
	Repo       string                 `json:"repo"`
	Collection string                 `json:"collection"`
	RKey       string                 `json:"rkey,omitempty"`
	Validate   bool                   `json:"validate"`
	Record     map[string]interface{} `json:"record"`
}