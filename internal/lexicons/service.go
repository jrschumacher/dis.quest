// Package lexicons provides the application service for quest.dis.* operations
package lexicons

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/jrschumacher/dis.quest/pkg/atproto/pds"
)

// Service provides quest.dis.* specific operations using the generic PDS client
type Service struct {
	pdsClient *pds.Client
}

// NewService creates a new lexicons service
func NewService() *Service {
	return &Service{
		pdsClient: pds.NewClient(),
	}
}

// Topic represents a quest.dis.topic with metadata
type Topic struct {
	URI            string    `json:"uri"`
	CID            string    `json:"cid"`
	Title          string    `json:"title"`
	Summary        string    `json:"summary,omitempty"`
	Tags           []string  `json:"tags,omitempty"`
	CreatedBy      string    `json:"createdBy"`
	CreatedAt      time.Time `json:"createdAt"`
	SelectedAnswer string    `json:"selectedAnswer,omitempty"`
}

// Message represents a quest.dis.message with metadata
type Message struct {
	URI       string    `json:"uri"`
	CID       string    `json:"cid"`
	Topic     string    `json:"topic"`
	Content   string    `json:"content"`
	ReplyTo   string    `json:"replyTo,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// Participation represents a quest.dis.participation with metadata
type Participation struct {
	URI         string    `json:"uri"`
	CID         string    `json:"cid"`
	Topic       string    `json:"topic"`
	Participant string    `json:"participant"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joinedAt"`
}

// CreateTopicParams contains parameters for creating a topic
type CreateTopicParams struct {
	Title   string   `json:"title"`
	Summary string   `json:"summary,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// CreateMessageParams contains parameters for creating a message
type CreateMessageParams struct {
	Topic   string `json:"topic"`
	Content string `json:"content"`
	ReplyTo string `json:"replyTo,omitempty"`
}

// CreateParticipationParams contains parameters for creating a participation
type CreateParticipationParams struct {
	Topic string `json:"topic"`
	Role  string `json:"role"`
}

// CreateTopic creates a topic in the user's PDS
func (s *Service) CreateTopic(ctx context.Context, userDID, accessToken string, dpopKey *ecdsa.PrivateKey, params CreateTopicParams) (*Topic, error) {
	// Create lexicon record
	topicRecord := NewTopicRecord(params.Title, userDID)
	topicRecord.Summary = params.Summary
	topicRecord.Tags = params.Tags

	// Create in PDS
	createParams := pds.CreateRecordParams{
		Collection: TopicLexicon,
		Record:     topicRecord.ToMap(),
	}

	result, err := s.pdsClient.CreateRecord(ctx, userDID, accessToken, dpopKey, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create topic: %w", err)
	}

	// Return Topic with metadata
	return &Topic{
		URI:       result.URI,
		CID:       result.CID,
		Title:     topicRecord.Title,
		Summary:   topicRecord.Summary,
		Tags:      topicRecord.Tags,
		CreatedBy: topicRecord.CreatedBy,
		CreatedAt: topicRecord.CreatedAt,
	}, nil
}

// GetTopic retrieves a topic by URI
func (s *Service) GetTopic(ctx context.Context, userDID, accessToken string, dpopKey *ecdsa.PrivateKey, uri string) (*Topic, error) {
	record, err := s.pdsClient.GetRecord(ctx, userDID, accessToken, dpopKey, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get topic: %w", err)
	}

	// Parse the record data
	topicRecord, err := TopicRecordFromMap(record.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse topic record: %w", err)
	}

	return &Topic{
		URI:            record.URI,
		CID:            record.CID,
		Title:          topicRecord.Title,
		Summary:        topicRecord.Summary,
		Tags:           topicRecord.Tags,
		CreatedBy:      topicRecord.CreatedBy,
		CreatedAt:      topicRecord.CreatedAt,
		SelectedAnswer: topicRecord.SelectedAnswer,
	}, nil
}

// CreateMessage creates a message in the user's PDS
func (s *Service) CreateMessage(ctx context.Context, userDID, accessToken string, dpopKey *ecdsa.PrivateKey, params CreateMessageParams) (*Message, error) {
	// Create lexicon record
	messageRecord := NewMessageRecord(params.Topic, params.Content)
	messageRecord.ReplyTo = params.ReplyTo

	// Create in PDS
	createParams := pds.CreateRecordParams{
		Collection: MessageLexicon,
		Record:     messageRecord.ToMap(),
	}

	result, err := s.pdsClient.CreateRecord(ctx, userDID, accessToken, dpopKey, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Return Message with metadata
	return &Message{
		URI:       result.URI,
		CID:       result.CID,
		Topic:     messageRecord.Topic,
		Content:   messageRecord.Content,
		ReplyTo:   messageRecord.ReplyTo,
		CreatedAt: messageRecord.CreatedAt,
	}, nil
}

// GetMessage retrieves a message by URI
func (s *Service) GetMessage(ctx context.Context, userDID, accessToken string, dpopKey *ecdsa.PrivateKey, uri string) (*Message, error) {
	record, err := s.pdsClient.GetRecord(ctx, userDID, accessToken, dpopKey, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Parse the record data
	messageRecord, err := MessageRecordFromMap(record.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message record: %w", err)
	}

	return &Message{
		URI:       record.URI,
		CID:       record.CID,
		Topic:     messageRecord.Topic,
		Content:   messageRecord.Content,
		ReplyTo:   messageRecord.ReplyTo,
		CreatedAt: messageRecord.CreatedAt,
	}, nil
}

// ListTopics lists topics from the user's PDS
func (s *Service) ListTopics(ctx context.Context, userDID, accessToken string, dpopKey *ecdsa.PrivateKey, limit int, cursor string) ([]*Topic, string, error) {
	listParams := pds.ListRecordsParams{
		Collection: TopicLexicon,
		Limit:      limit,
		Cursor:     cursor,
	}

	result, err := s.pdsClient.ListRecords(ctx, userDID, accessToken, dpopKey, listParams)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list topics: %w", err)
	}

	var topics []*Topic
	for _, record := range result.Records {
		topicRecord, err := TopicRecordFromMap(record.Value)
		if err != nil {
			// Log error but continue with other records
			continue
		}

		topic := &Topic{
			URI:            record.URI,
			CID:            record.CID,
			Title:          topicRecord.Title,
			Summary:        topicRecord.Summary,
			Tags:           topicRecord.Tags,
			CreatedBy:      topicRecord.CreatedBy,
			CreatedAt:      topicRecord.CreatedAt,
			SelectedAnswer: topicRecord.SelectedAnswer,
		}
		topics = append(topics, topic)
	}

	return topics, result.Cursor, nil
}

// UpdateTopicSelectedAnswer updates the selectedAnswer field of a topic
func (s *Service) UpdateTopicSelectedAnswer(ctx context.Context, userDID, accessToken string, dpopKey *ecdsa.PrivateKey, topicURI, answerURI string) error {
	// Get the current topic
	record, err := s.pdsClient.GetRecord(ctx, userDID, accessToken, dpopKey, topicURI)
	if err != nil {
		return fmt.Errorf("failed to get topic for update: %w", err)
	}

	// Parse and update the record
	topicRecord, err := TopicRecordFromMap(record.Value)
	if err != nil {
		return fmt.Errorf("failed to parse topic record: %w", err)
	}

	topicRecord.SelectedAnswer = answerURI

	// Update in PDS
	_, err = s.pdsClient.UpdateRecord(ctx, userDID, accessToken, dpopKey, topicURI, topicRecord.ToMap())
	if err != nil {
		return fmt.Errorf("failed to update topic: %w", err)
	}

	return nil
}

// CreateParticipation creates a participation record in the user's PDS
func (s *Service) CreateParticipation(ctx context.Context, userDID, accessToken string, dpopKey *ecdsa.PrivateKey, params CreateParticipationParams) (*Participation, error) {
	// Create lexicon record
	participationRecord := NewParticipationRecord(params.Topic, userDID, params.Role)

	// Create in PDS
	createParams := pds.CreateRecordParams{
		Collection: ParticipationLexicon,
		Record:     participationRecord.ToMap(),
	}

	result, err := s.pdsClient.CreateRecord(ctx, userDID, accessToken, dpopKey, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create participation: %w", err)
	}

	// Return Participation with metadata
	return &Participation{
		URI:         result.URI,
		CID:         result.CID,
		Topic:       participationRecord.Topic,
		Participant: participationRecord.Participant,
		Role:        participationRecord.Role,
		JoinedAt:    participationRecord.JoinedAt,
	}, nil
}

// ResolvePDS resolves the PDS endpoint for a given DID
func (s *Service) ResolvePDS(ctx context.Context, did string) (string, error) {
	return s.pdsClient.ResolvePDS(ctx, did)
}