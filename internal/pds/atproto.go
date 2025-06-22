// Package pds provides real ATProtocol PDS service implementation
package pds

import (
	"context"
	"fmt"
	"time"

	"github.com/jrschumacher/dis.quest/internal/logger"
)

// ATProtoService implements Service interface with real ATProtocol PDS calls
type ATProtoService struct {
	xrpc *XRPCClient
}

// NewATProtoService creates a new ATProtocol PDS service
func NewATProtoService() *ATProtoService {
	return &ATProtoService{
		xrpc: NewXRPCClient(),
	}
}

// CreateTopic creates a topic in the user's real PDS using lexicon abstractions
func (s *ATProtoService) CreateTopic(userDID string, params CreateTopicParams) (*Topic, error) {
	return s.CreateTopicWithToken(userDID, params, "")
}

// CreateTopicWithToken creates a topic in the user's real PDS with an access token
func (s *ATProtoService) CreateTopicWithToken(userDID string, params CreateTopicParams, accessToken string) (*Topic, error) {
	return s.CreateTopicWithDPoP(userDID, params, accessToken, nil)
}

// CreateTopicWithDPoP creates a topic in the user's real PDS with access token and DPoP authentication
func (s *ATProtoService) CreateTopicWithDPoP(userDID string, params CreateTopicParams, accessToken string, dpopKey interface{}) (*Topic, error) {
	logger.Info("Starting CreateTopicWithDPoP", 
		"userDID", userDID, 
		"title", params.Title,
		"hasAccessToken", accessToken != "",
		"hasDPoPKey", dpopKey != nil)
	// Create lexicon record
	topicRecord := &TopicRecord{
		Type:      TopicLexicon,
		Title:     params.Title,
		Summary:   params.Summary,
		Tags:      params.Tags,
		CreatedBy: userDID,
		CreatedAt: time.Now(),
	}

	// Validate the record
	recordData := topicRecord.ToMap()
	if err := ValidateLexicon(TopicLexicon, recordData); err != nil {
		return nil, fmt.Errorf("invalid topic record: %w", err)
	}

	// Generate unique rkey
	rkey := GenerateRKey("topic")

	// Create XRPC request
	req := CreateRecordRequest{
		Repo:       userDID,
		Collection: TopicLexicon,
		RKey:       rkey,
		Validate:   false, // Custom lexicons require validate: false
		Record:     recordData,
	}

	// Access token provided as parameter

	// Make the XRPC call with DPoP support
	logger.Info("About to call XRPC CreateRecordWithDPoP", 
		"repo", req.Repo,
		"collection", req.Collection,
		"rkey", req.RKey,
		"hasAccessToken", accessToken != "",
		"dpopKeyType", fmt.Sprintf("%T", dpopKey))
	
	ctx := context.Background()
	resp, err := s.xrpc.CreateRecordWithDPoP(ctx, req, accessToken, dpopKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create topic in PDS: %w", err)
	}

	// Convert to Topic struct
	topic := topicRecord.ToTopic(resp.URI, resp.CID)
	
	logger.Info("Successfully created topic in PDS", "uri", topic.URI, "cid", topic.CID, "lexicon", TopicLexicon)
	return topic, nil
}

// GetTopic retrieves a topic from the user's PDS using lexicon abstractions
func (s *ATProtoService) GetTopic(uri string) (*Topic, error) {
	// Parse AT URI
	components, err := ParseATUri(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid AT URI: %w", err)
	}

	// Verify this is a topic collection
	if components.Collection != TopicLexicon {
		return nil, fmt.Errorf("URI is not a topic record: %s", uri)
	}

	// TODO: Get access token from context/user session
	accessToken := ""

	// Make the XRPC call
	ctx := context.Background()
	resp, err := s.xrpc.GetRecord(ctx, components.DID, components.Collection, components.RKey, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get topic from PDS: %w", err)
	}

	// Parse into lexicon record
	topicRecord := &TopicRecord{}
	if err := topicRecord.FromMap(resp.Value); err != nil {
		return nil, fmt.Errorf("failed to parse topic record: %w", err)
	}

	// Convert to Topic struct
	topic := topicRecord.ToTopic(resp.URI, resp.CID)
	
	return topic, nil
}

// UpdateTopicSelectedAnswer updates the selected answer for a topic using lexicon abstractions
func (s *ATProtoService) UpdateTopicSelectedAnswer(topicURI, answerURI string) error {
	// Parse AT URI
	components, err := ParseATUri(topicURI)
	if err != nil {
		return fmt.Errorf("invalid AT URI: %w", err)
	}

	// First get the current topic to preserve other fields
	topic, err := s.GetTopic(topicURI)
	if err != nil {
		return fmt.Errorf("failed to get current topic: %w", err)
	}

	// Create updated lexicon record
	topicRecord := &TopicRecord{
		Type:           TopicLexicon,
		Title:          topic.Title,
		Summary:        topic.Summary,
		Tags:           topic.Tags,
		CreatedBy:      topic.CreatedBy,
		CreatedAt:      topic.CreatedAt,
		SelectedAnswer: answerURI,
	}

	// Validate the updated record
	recordData := topicRecord.ToMap()
	if err := ValidateLexicon(TopicLexicon, recordData); err != nil {
		return fmt.Errorf("invalid updated topic record: %w", err)
	}

	// Create XRPC request
	req := PutRecordRequest{
		Repo:       components.DID,
		Collection: components.Collection,
		RKey:       components.RKey,
		Validate:   true,
		Record:     recordData,
	}

	// TODO: Get access token from context/user session
	accessToken := ""

	// Make the XRPC call
	ctx := context.Background()
	_, err = s.xrpc.PutRecord(ctx, req, accessToken)
	if err != nil {
		return fmt.Errorf("failed to update topic in PDS: %w", err)
	}

	logger.Info("Successfully updated topic selected answer", "uri", topicURI, "answer", answerURI)
	return nil
}

// Implement the rest of the Service interface methods...

// CreateMessage creates a message in the user's PDS using lexicon abstractions
func (s *ATProtoService) CreateMessage(userDID string, params CreateMessageParams) (*Message, error) {
	// Create lexicon record
	messageRecord := &MessageRecord{
		Type:      MessageLexicon,
		Topic:     params.Topic,
		Content:   params.Content,
		ReplyTo:   params.ReplyTo,
		CreatedAt: time.Now(),
	}

	// Validate the record
	recordData := messageRecord.ToMap()
	if err := ValidateLexicon(MessageLexicon, recordData); err != nil {
		return nil, fmt.Errorf("invalid message record: %w", err)
	}

	// Generate unique rkey
	rkey := GenerateRKey("msg")

	// Create XRPC request
	req := CreateRecordRequest{
		Repo:       userDID,
		Collection: MessageLexicon,
		RKey:       rkey,
		Validate:   true,
		Record:     recordData,
	}

	// TODO: Get access token from context/user session
	accessToken := ""

	// Make the XRPC call
	ctx := context.Background()
	resp, err := s.xrpc.CreateRecord(ctx, req, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create message in PDS: %w", err)
	}

	// Convert to Message struct
	message := messageRecord.ToMessage(resp.URI, resp.CID)
	
	logger.Info("Successfully created message in PDS", "uri", message.URI, "cid", message.CID, "lexicon", MessageLexicon)
	return message, nil
}

// GetMessage retrieves a message by URI
func (s *ATProtoService) GetMessage(_ string) (*Message, error) {
	return nil, fmt.Errorf("GetMessage not yet implemented for ATProto service")
}

// GetMessagesByTopic retrieves all messages for a topic
func (s *ATProtoService) GetMessagesByTopic(_ string) ([]*Message, error) {
	return nil, fmt.Errorf("GetMessagesByTopic not yet implemented for ATProto service")
}

// CreateParticipation creates a participation record in the user's PDS
func (s *ATProtoService) CreateParticipation(_ string, _ CreateParticipationParams) (*Participation, error) {
	return nil, fmt.Errorf("CreateParticipation not yet implemented for ATProto service")
}

// GetParticipationsByTopic retrieves all participations for a topic
func (s *ATProtoService) GetParticipationsByTopic(_ string) ([]*Participation, error) {
	return nil, fmt.Errorf("GetParticipationsByTopic not yet implemented for ATProto service")
}

// CreatePost creates a legacy post (deprecated)
func (s *ATProtoService) CreatePost(_ string) (*Post, error) {
	return nil, fmt.Errorf("CreatePost not implemented for ATProto service")
}

// GetPost retrieves a legacy post (deprecated)
func (s *ATProtoService) GetPost(_ string) (*Post, error) {
	return nil, fmt.Errorf("GetPost not implemented for ATProto service")
}

// SetSelectedAnswer sets selected answer for legacy post (deprecated)
func (s *ATProtoService) SetSelectedAnswer(_, _ string) error {
	return fmt.Errorf("SetSelectedAnswer not implemented for ATProto service")
}