// Package lexicons provides backwards compatibility with the old internal/pds interface
package lexicons

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"
	"time"
)

// LegacyPDSService provides backwards compatibility with the old internal/pds interface
type LegacyPDSService struct {
	service *Service
	isTest  bool
}

// NewLegacyPDSService creates a backwards-compatible PDS service
func NewLegacyPDSService() *LegacyPDSService {
	return &LegacyPDSService{
		service: NewService(),
	}
}

// Post represents a legacy post structure for backwards compatibility
type Post struct {
	ID             string
	Content        string
	SelectedAnswer string
}

// Legacy interface methods - these maintain the old signature but delegate to new implementation

// CreateTopic creates a topic (old interface)
func (l *LegacyPDSService) CreateTopic(userDID string, params CreateTopicParams) (*Topic, error) {
	if l.isTest {
		// Return a mock topic for testing
		return &Topic{
			URI:       fmt.Sprintf("at://%s/quest.dis.topic/test-topic-123", userDID),
			CID:       "bafyreigbtj4x7ip5legnfznufuopl4sg4knzc2cof6duas4b3q2fy6swua",
			Title:     params.Title,
			Summary:   params.Summary,
			Tags:      params.Tags,
			CreatedBy: userDID,
			CreatedAt: time.Now(),
		}, nil
	}
	// For backward compatibility, delegate to auth-aware method with dummy auth
	return l.service.CreateTopic(context.Background(), userDID, "", nil, params)
}

// CreateTopicWithAuth creates a topic with authentication (transition method)
func (l *LegacyPDSService) CreateTopicWithAuth(ctx context.Context, _ /* pdsEndpoint */, accessToken string, dpopKey *ecdsa.PrivateKey, userDID string, params CreateTopicParams) (*Topic, error) {
	return l.service.CreateTopic(ctx, userDID, accessToken, dpopKey, params)
}

// GetTopic retrieves a topic (old interface)
func (l *LegacyPDSService) GetTopic(uri string) (*Topic, error) {
	// For backward compatibility and testing, delegate to auth-aware method with dummy auth
	userDID := extractUserDIDFromURI(uri)
	return l.service.GetTopic(context.Background(), userDID, "", nil, uri)
}

// GetTopicWithAuth retrieves a topic with authentication (transition method)
func (l *LegacyPDSService) GetTopicWithAuth(ctx context.Context, _ /* pdsEndpoint */, accessToken string, dpopKey *ecdsa.PrivateKey, uri string) (*Topic, error) {
	// Note: pdsEndpoint parameter is ignored in new implementation
	// Extract userDID from URI or use a placeholder for now
	userDID := extractUserDIDFromURI(uri)
	return l.service.GetTopic(ctx, userDID, accessToken, dpopKey, uri)
}

// UpdateTopicSelectedAnswer updates topic's selected answer (old interface)
func (l *LegacyPDSService) UpdateTopicSelectedAnswer(topicURI, answerURI string) error {
	// For backward compatibility and testing, delegate to auth-aware method with dummy auth
	userDID := extractUserDIDFromURI(topicURI)
	return l.service.UpdateTopicSelectedAnswer(context.Background(), userDID, "", nil, topicURI, answerURI)
}

// UpdateTopicSelectedAnswerWithAuth updates topic's selected answer with authentication (transition method)
func (l *LegacyPDSService) UpdateTopicSelectedAnswerWithAuth(ctx context.Context, _ /* pdsEndpoint */, accessToken string, dpopKey *ecdsa.PrivateKey, topicURI, answerURI string) error {
	// Note: pdsEndpoint parameter is ignored in new implementation
	// Extract userDID from URI or use a placeholder for now
	userDID := extractUserDIDFromURI(topicURI)
	return l.service.UpdateTopicSelectedAnswer(ctx, userDID, accessToken, dpopKey, topicURI, answerURI)
}

// CreateMessage creates a message (old interface)
func (l *LegacyPDSService) CreateMessage(userDID string, params CreateMessageParams) (*Message, error) {
	// For backward compatibility and testing, delegate to auth-aware method with dummy auth
	return l.service.CreateMessage(context.Background(), userDID, "", nil, params)
}

// CreateMessageWithAuth creates a message with authentication (transition method)
func (l *LegacyPDSService) CreateMessageWithAuth(ctx context.Context, _ /* pdsEndpoint */, accessToken string, dpopKey *ecdsa.PrivateKey, userDID string, params CreateMessageParams) (*Message, error) {
	return l.service.CreateMessage(ctx, userDID, accessToken, dpopKey, params)
}

// GetMessage retrieves a message (old interface)
func (l *LegacyPDSService) GetMessage(uri string) (*Message, error) {
	// For backward compatibility and testing, delegate to auth-aware method with dummy auth
	userDID := extractUserDIDFromURI(uri)
	return l.service.GetMessage(context.Background(), userDID, "", nil, uri)
}

// GetMessageWithAuth retrieves a message with authentication (transition method)
func (l *LegacyPDSService) GetMessageWithAuth(ctx context.Context, _ /* pdsEndpoint */, accessToken string, dpopKey *ecdsa.PrivateKey, uri string) (*Message, error) {
	// Note: pdsEndpoint parameter is ignored in new implementation
	// Extract userDID from URI or use a placeholder for now
	userDID := extractUserDIDFromURI(uri)
	return l.service.GetMessage(ctx, userDID, accessToken, dpopKey, uri)
}

// GetMessagesByTopic retrieves messages by topic (old interface)
func (l *LegacyPDSService) GetMessagesByTopic(_ /* topicURI */ string) ([]*Message, error) {
	// This method is deprecated but for testing return empty slice
	return []*Message{}, nil
}

// CreateParticipation creates a participation (old interface)
func (l *LegacyPDSService) CreateParticipation(userDID string, params CreateParticipationParams) (*Participation, error) {
	if l.isTest {
		// Return a mock participation for testing
		return &Participation{
			URI:         fmt.Sprintf("at://%s/quest.dis.participation/test-participation-123", userDID),
			CID:         "bafyreigbtj4x7ip5legnfznufuopl4sg4knzc2cof6duas4b3q2fy6swua",
			Topic:       params.Topic,
			Participant: userDID,
			Role:        params.Role,
			JoinedAt:    time.Now(),
		}, nil
	}
	// For backward compatibility, delegate to auth-aware method with dummy auth
	return l.service.CreateParticipation(context.Background(), userDID, "", nil, params)
}

// CreateParticipationWithAuth creates a participation with authentication (transition method)
func (l *LegacyPDSService) CreateParticipationWithAuth(ctx context.Context, _ /* pdsEndpoint */, accessToken string, dpopKey *ecdsa.PrivateKey, userDID string, params CreateParticipationParams) (*Participation, error) {
	return l.service.CreateParticipation(ctx, userDID, accessToken, dpopKey, params)
}

// GetParticipationsByTopic retrieves participations by topic (old interface)
func (l *LegacyPDSService) GetParticipationsByTopic(_ /* topicURI */ string) ([]*Participation, error) {
	// This method is deprecated but for testing return empty slice
	return []*Participation{}, nil
}

// CreatePost creates a legacy post (deprecated - for backwards compatibility)
func (l *LegacyPDSService) CreatePost(_ /* content */ string) (*Post, error) {
	return nil, fmt.Errorf("legacy CreatePost method is deprecated, use CreateTopic or CreateMessage")
}

// GetPost retrieves a legacy post (deprecated - for backwards compatibility)
func (l *LegacyPDSService) GetPost(_ /* id */ string) (*Post, error) {
	return nil, fmt.Errorf("legacy GetPost method is deprecated, use GetTopic or GetMessage")
}

// SetSelectedAnswer sets the selected answer for a legacy post (deprecated - for backwards compatibility)
func (l *LegacyPDSService) SetSelectedAnswer(_, _ /* postID, answerID */ string) error {
	return fmt.Errorf("legacy SetSelectedAnswer method is deprecated, use UpdateTopicSelectedAnswer")
}

// NewATProtoService creates a new legacy service (for dev.go compatibility)
func NewATProtoService() *LegacyPDSService {
	return NewLegacyPDSService()
}

// NewMockService creates a mock service for testing
func NewMockService() *LegacyPDSService {
	return &LegacyPDSService{
		service: nil, // Don't need real service for tests
		isTest:  true,
	}
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