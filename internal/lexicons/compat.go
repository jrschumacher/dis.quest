// Package lexicons provides backwards compatibility with the old internal/pds interface
package lexicons

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"
)

// Legacy interface compatibility - delegates to the new Service
type LegacyPDSService struct {
	service *Service
}

// NewLegacyPDSService creates a backwards-compatible PDS service
func NewLegacyPDSService() *LegacyPDSService {
	return &LegacyPDSService{
		service: NewService(),
	}
}

// Legacy Post structure for backwards compatibility
type Post struct {
	ID             string
	Content        string
	SelectedAnswer string
}

// Legacy interface methods - these maintain the old signature but delegate to new implementation

// CreateTopic creates a topic (old interface)
func (l *LegacyPDSService) CreateTopic(userDID string, params CreateTopicParams) (*Topic, error) {
	// This method lacks context and auth info, so it's a placeholder
	// In practice, callers should migrate to the new context-aware methods
	return nil, fmt.Errorf("legacy CreateTopic method requires migration to CreateTopicWithAuth")
}

// CreateTopicWithAuth creates a topic with authentication (transition method)
func (l *LegacyPDSService) CreateTopicWithAuth(ctx context.Context, pdsEndpoint, accessToken string, dpopKey *ecdsa.PrivateKey, userDID string, params CreateTopicParams) (*Topic, error) {
	return l.service.CreateTopic(ctx, userDID, accessToken, dpopKey, params)
}

// GetTopic retrieves a topic (old interface)
func (l *LegacyPDSService) GetTopic(uri string) (*Topic, error) {
	// This method lacks context and auth info, so it's a placeholder
	return nil, fmt.Errorf("legacy GetTopic method requires migration to GetTopicWithAuth")
}

// GetTopicWithAuth retrieves a topic with authentication (transition method)
func (l *LegacyPDSService) GetTopicWithAuth(ctx context.Context, pdsEndpoint, accessToken string, dpopKey *ecdsa.PrivateKey, uri string) (*Topic, error) {
	// Note: pdsEndpoint parameter is ignored in new implementation
	// Extract userDID from URI or use a placeholder for now
	userDID := extractUserDIDFromURI(uri)
	return l.service.GetTopic(ctx, userDID, accessToken, dpopKey, uri)
}

// UpdateTopicSelectedAnswer updates topic's selected answer (old interface)
func (l *LegacyPDSService) UpdateTopicSelectedAnswer(topicURI, answerURI string) error {
	// This method lacks context and auth info, so it's a placeholder
	return fmt.Errorf("legacy UpdateTopicSelectedAnswer method requires migration to UpdateTopicSelectedAnswerWithAuth")
}

// UpdateTopicSelectedAnswerWithAuth updates topic's selected answer with authentication (transition method)
func (l *LegacyPDSService) UpdateTopicSelectedAnswerWithAuth(ctx context.Context, pdsEndpoint, accessToken string, dpopKey *ecdsa.PrivateKey, topicURI, answerURI string) error {
	// Note: pdsEndpoint parameter is ignored in new implementation
	// Extract userDID from URI or use a placeholder for now
	userDID := extractUserDIDFromURI(topicURI)
	return l.service.UpdateTopicSelectedAnswer(ctx, userDID, accessToken, dpopKey, topicURI, answerURI)
}

// CreateMessage creates a message (old interface)
func (l *LegacyPDSService) CreateMessage(userDID string, params CreateMessageParams) (*Message, error) {
	// This method lacks context and auth info, so it's a placeholder
	return nil, fmt.Errorf("legacy CreateMessage method requires migration to CreateMessageWithAuth")
}

// CreateMessageWithAuth creates a message with authentication (transition method)
func (l *LegacyPDSService) CreateMessageWithAuth(ctx context.Context, pdsEndpoint, accessToken string, dpopKey *ecdsa.PrivateKey, userDID string, params CreateMessageParams) (*Message, error) {
	return l.service.CreateMessage(ctx, userDID, accessToken, dpopKey, params)
}

// GetMessage retrieves a message (old interface)
func (l *LegacyPDSService) GetMessage(uri string) (*Message, error) {
	// This method lacks context and auth info, so it's a placeholder
	return nil, fmt.Errorf("legacy GetMessage method requires migration to GetMessageWithAuth")
}

// GetMessageWithAuth retrieves a message with authentication (transition method)
func (l *LegacyPDSService) GetMessageWithAuth(ctx context.Context, pdsEndpoint, accessToken string, dpopKey *ecdsa.PrivateKey, uri string) (*Message, error) {
	// Note: pdsEndpoint parameter is ignored in new implementation
	// Extract userDID from URI or use a placeholder for now
	userDID := extractUserDIDFromURI(uri)
	return l.service.GetMessage(ctx, userDID, accessToken, dpopKey, uri)
}

// GetMessagesByTopic retrieves messages by topic (old interface)
func (l *LegacyPDSService) GetMessagesByTopic(topicURI string) ([]*Message, error) {
	// This method lacks context and auth info, so it's a placeholder
	return nil, fmt.Errorf("legacy GetMessagesByTopic method requires migration to ListMessages")
}

// CreateParticipation creates a participation (old interface)
func (l *LegacyPDSService) CreateParticipation(userDID string, params CreateParticipationParams) (*Participation, error) {
	// This method lacks context and auth info, so it's a placeholder
	return nil, fmt.Errorf("legacy CreateParticipation method requires migration to CreateParticipationWithAuth")
}

// CreateParticipationWithAuth creates a participation with authentication (transition method)
func (l *LegacyPDSService) CreateParticipationWithAuth(ctx context.Context, pdsEndpoint, accessToken string, dpopKey *ecdsa.PrivateKey, userDID string, params CreateParticipationParams) (*Participation, error) {
	return l.service.CreateParticipation(ctx, userDID, accessToken, dpopKey, params)
}

// GetParticipationsByTopic retrieves participations by topic (old interface)
func (l *LegacyPDSService) GetParticipationsByTopic(topicURI string) ([]*Participation, error) {
	// This method lacks context and auth info, so it's a placeholder
	return nil, fmt.Errorf("legacy GetParticipationsByTopic method requires migration to ListParticipations")
}

// Legacy post operations (for backwards compatibility)
func (l *LegacyPDSService) CreatePost(content string) (*Post, error) {
	return nil, fmt.Errorf("legacy CreatePost method is deprecated, use CreateTopic or CreateMessage")
}

func (l *LegacyPDSService) GetPost(id string) (*Post, error) {
	return nil, fmt.Errorf("legacy GetPost method is deprecated, use GetTopic or GetMessage")
}

func (l *LegacyPDSService) SetSelectedAnswer(postID, answerID string) error {
	return fmt.Errorf("legacy SetSelectedAnswer method is deprecated, use UpdateTopicSelectedAnswer")
}

// NewATProtoService creates a new legacy service (for dev.go compatibility)
func NewATProtoService() *LegacyPDSService {
	return NewLegacyPDSService()
}

// NewMockService creates a mock service for testing
func NewMockService() *LegacyPDSService {
	return NewLegacyPDSService()
}

// extractUserDIDFromURI extracts the DID from an AT Protocol URI
// URI format: at://did:plc:example/collection/rkey
func extractUserDIDFromURI(uri string) string {
	if !strings.HasPrefix(uri, "at://") {
		return "" // Invalid URI format
	}
	
	// Remove "at://" prefix
	remaining := strings.TrimPrefix(uri, "at://")
	
	// Split by "/" and take the first part (the DID)
	parts := strings.Split(remaining, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	
	return ""
}