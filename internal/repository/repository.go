package repository

import (
	"context"
	"time"

	"github.com/jrschumacher/dis.quest/internal/db"
)

// TopicRepository provides high-level operations for topics
type TopicRepository interface {
	CreateTopic(ctx context.Context, params CreateTopicParams) (*TopicDetail, error)
	GetTopic(ctx context.Context, did, rkey string) (*TopicDetail, error)
	ListTopics(ctx context.Context, params ListTopicsParams) ([]*TopicSummary, error)
	GetTopicsByCategory(ctx context.Context, category string, limit int) ([]*TopicSummary, error)
	UpdateSelectedAnswer(ctx context.Context, topicDID, topicRkey, messageRkey string, userDID string) error
}

// MessageRepository provides high-level operations for messages
type MessageRepository interface {
	CreateMessage(ctx context.Context, params CreateMessageParams) (*MessageDetail, error)
	GetMessage(ctx context.Context, did, rkey string) (*MessageDetail, error)
	GetMessagesByTopic(ctx context.Context, topicDID, topicRkey string) ([]*MessageDetail, error)
	GetRepliesByMessage(ctx context.Context, topicDID, topicRkey, parentRkey string) ([]*MessageDetail, error)
	DeleteMessage(ctx context.Context, did, rkey string, userDID string) error
}

// ParticipationRepository provides high-level operations for participation
type ParticipationRepository interface {
	CreateParticipation(ctx context.Context, params CreateParticipationParams) (*ParticipationDetail, error)
	GetParticipation(ctx context.Context, userDID, topicDID, topicRkey string) (*ParticipationDetail, error)
	GetParticipationsByUser(ctx context.Context, userDID string) ([]*ParticipationDetail, error)
	GetParticipationsByTopic(ctx context.Context, topicDID, topicRkey string) ([]*ParticipationDetail, error)
	UpdateParticipationStatus(ctx context.Context, userDID, topicDID, topicRkey, status string) error
	DeleteParticipation(ctx context.Context, userDID, topicDID, topicRkey string) error
}

// Repository aggregates all repository interfaces
type Repository interface {
	Topics() TopicRepository
	Messages() MessageRepository
	Participation() ParticipationRepository
}

// CreateTopicParams represents parameters for creating a topic
type CreateTopicParams struct {
	Did            string
	Rkey           string
	Subject        string
	InitialMessage string
	Category       string
}

// CreateMessageParams represents parameters for creating a message
type CreateMessageParams struct {
	Did               string
	Rkey              string
	TopicDID          string
	TopicRkey         string
	ParentMessageRkey string
	Content           string
}

// CreateParticipationParams represents parameters for creating participation
type CreateParticipationParams struct {
	Did       string
	TopicDID  string
	TopicRkey string
	Status    string
}

// ListTopicsParams represents parameters for listing topics
type ListTopicsParams struct {
	Limit  int
	Offset int
}

// TopicDetail represents a topic with full details
type TopicDetail struct {
	DID            string            `json:"did"`
	Rkey           string            `json:"rkey"`
	Subject        string            `json:"subject"`
	InitialMessage string            `json:"initial_message"`
	Category       string            `json:"category,omitempty"`
	SelectedAnswer string            `json:"selected_answer,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	MessageCount   int               `json:"message_count,omitempty"`
	Participants   []ParticipantInfo `json:"participants,omitempty"`
}

// TopicSummary represents a topic summary for listings
type TopicSummary struct {
	DID            string    `json:"did"`
	Rkey           string    `json:"rkey"`
	Subject        string    `json:"subject"`
	Category       string    `json:"category,omitempty"`
	MessageCount   int       `json:"message_count"`
	LastActivity   time.Time `json:"last_activity"`
	CreatedAt      time.Time `json:"created_at"`
	HasAnswer      bool      `json:"has_answer"`
}

// MessageDetail represents a message with full details
type MessageDetail struct {
	DID               string    `json:"did"`
	Rkey              string    `json:"rkey"`
	TopicDID          string    `json:"topic_did"`
	TopicRkey         string    `json:"topic_rkey"`
	ParentMessageRkey string    `json:"parent_message_rkey,omitempty"`
	Content           string    `json:"content"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	IsAnswer          bool      `json:"is_answer,omitempty"`
	ReplyCount        int       `json:"reply_count,omitempty"`
}

// ParticipationDetail represents participation with full details
type ParticipationDetail struct {
	Did       string    `json:"did"`
	TopicDID  string    `json:"topic_did"`
	TopicRkey string    `json:"topic_rkey"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ParticipantInfo represents basic participant information
type ParticipantInfo struct {
	DID    string `json:"did"`
	Status string `json:"status"`
}

// repositoryImpl implements the Repository interface using the database service
type repositoryImpl struct {
	dbService *db.Service
	topics    TopicRepository
	messages  MessageRepository
	participation ParticipationRepository
}

// NewRepository creates a new repository instance
func NewRepository(dbService *db.Service) Repository {
	repo := &repositoryImpl{
		dbService: dbService,
	}
	
	repo.topics = &topicRepository{dbService: dbService}
	repo.messages = &messageRepository{dbService: dbService}
	repo.participation = &participationRepository{dbService: dbService}
	
	return repo
}

// Topics returns the topic repository
func (r *repositoryImpl) Topics() TopicRepository {
	return r.topics
}

// Messages returns the message repository
func (r *repositoryImpl) Messages() MessageRepository {
	return r.messages
}

// Participation returns the participation repository
func (r *repositoryImpl) Participation() ParticipationRepository {
	return r.participation
}