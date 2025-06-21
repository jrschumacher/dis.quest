// Package pds provides interfaces and mocks for Personal Data Server (PDS) interactions.
package pds

import (
	"fmt"
	"time"
)

// Topic represents a quest.dis.topic record in the PDS
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

// Message represents a quest.dis.message record in the PDS
type Message struct {
	URI       string    `json:"uri"`
	CID       string    `json:"cid"`
	Topic     string    `json:"topic"`
	Content   string    `json:"content"`
	ReplyTo   string    `json:"replyTo,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// Participation represents a quest.dis.participation record in the PDS
type Participation struct {
	URI         string    `json:"uri"`
	CID         string    `json:"cid"`
	Topic       string    `json:"topic"`
	Participant string    `json:"participant"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joinedAt"`
}

// Post represents a minimal post structure for testing.
type Post struct {
	ID             string
	Content        string
	SelectedAnswer string
}

// Service defines the interface for PDS operations.
type Service interface {
	// Topic operations
	CreateTopic(userDID string, params CreateTopicParams) (*Topic, error)
	GetTopic(uri string) (*Topic, error)
	UpdateTopicSelectedAnswer(topicURI, answerURI string) error
	
	// Message operations
	CreateMessage(userDID string, params CreateMessageParams) (*Message, error)
	GetMessage(uri string) (*Message, error)
	GetMessagesByTopic(topicURI string) ([]*Message, error)
	
	// Participation operations
	CreateParticipation(userDID string, params CreateParticipationParams) (*Participation, error)
	GetParticipationsByTopic(topicURI string) ([]*Participation, error)
	
	// Legacy post operations (for backwards compatibility)
	CreatePost(content string) (*Post, error)
	GetPost(id string) (*Post, error)
	SetSelectedAnswer(postID, answerID string) error
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

// CreateParticipationParams contains parameters for creating participation
type CreateParticipationParams struct {
	Topic string `json:"topic"`
	Role  string `json:"role,omitempty"`
}

// MockService is an in-memory mock implementation of Service.
type MockService struct {
	topics        map[string]*Topic
	messages      map[string]*Message  
	participations map[string]*Participation
	posts         map[string]*Post // Legacy
	nextID        int
}

// NewMockService creates a new MockService.
func NewMockService() *MockService {
	return &MockService{
		topics:        make(map[string]*Topic),
		messages:      make(map[string]*Message),
		participations: make(map[string]*Participation),
		posts:         make(map[string]*Post),
		nextID:        1,
	}
}

// Topic operations

// CreateTopic creates a new topic in the user's PDS
func (m *MockService) CreateTopic(userDID string, params CreateTopicParams) (*Topic, error) {
	uri := fmt.Sprintf("at://%s/quest.dis.topic/topic-%d", userDID, m.nextID)
	cid := fmt.Sprintf("bafyrei%d", m.nextID)
	m.nextID++
	
	topic := &Topic{
		URI:       uri,
		CID:       cid,
		Title:     params.Title,
		Summary:   params.Summary,
		Tags:      params.Tags,
		CreatedBy: userDID,
		CreatedAt: time.Now(),
	}
	
	m.topics[uri] = topic
	return topic, nil
}

// GetTopic retrieves a topic by URI
func (m *MockService) GetTopic(uri string) (*Topic, error) {
	topic, ok := m.topics[uri]
	if !ok {
		return nil, fmt.Errorf("topic %s not found", uri)
	}
	return topic, nil
}

// UpdateTopicSelectedAnswer updates the selected answer for a topic
func (m *MockService) UpdateTopicSelectedAnswer(topicURI, answerURI string) error {
	topic, ok := m.topics[topicURI]
	if !ok {
		return fmt.Errorf("topic %s not found", topicURI)
	}
	topic.SelectedAnswer = answerURI
	return nil
}

// Message operations

// CreateMessage creates a new message in the user's PDS
func (m *MockService) CreateMessage(userDID string, params CreateMessageParams) (*Message, error) {
	uri := fmt.Sprintf("at://%s/quest.dis.message/msg-%d", userDID, m.nextID)
	cid := fmt.Sprintf("bafyrei%d", m.nextID)
	m.nextID++
	
	message := &Message{
		URI:       uri,
		CID:       cid,
		Topic:     params.Topic,
		Content:   params.Content,
		ReplyTo:   params.ReplyTo,
		CreatedAt: time.Now(),
	}
	
	m.messages[uri] = message
	return message, nil
}

// GetMessage retrieves a message by URI
func (m *MockService) GetMessage(uri string) (*Message, error) {
	message, ok := m.messages[uri]
	if !ok {
		return nil, fmt.Errorf("message %s not found", uri)
	}
	return message, nil
}

// GetMessagesByTopic retrieves all messages for a topic
func (m *MockService) GetMessagesByTopic(topicURI string) ([]*Message, error) {
	var messages []*Message
	for _, message := range m.messages {
		if message.Topic == topicURI {
			messages = append(messages, message)
		}
	}
	return messages, nil
}

// Participation operations

// CreateParticipation creates a new participation record in the user's PDS
func (m *MockService) CreateParticipation(userDID string, params CreateParticipationParams) (*Participation, error) {
	uri := fmt.Sprintf("at://%s/quest.dis.participation/participation-%d", userDID, m.nextID)
	cid := fmt.Sprintf("bafyrei%d", m.nextID)
	m.nextID++
	
	role := params.Role
	if role == "" {
		role = "contributor"
	}
	
	participation := &Participation{
		URI:         uri,
		CID:         cid,
		Topic:       params.Topic,
		Participant: userDID,
		Role:        role,
		JoinedAt:    time.Now(),
	}
	
	m.participations[uri] = participation
	return participation, nil
}

// GetParticipationsByTopic retrieves all participations for a topic
func (m *MockService) GetParticipationsByTopic(topicURI string) ([]*Participation, error) {
	var participations []*Participation
	for _, participation := range m.participations {
		if participation.Topic == topicURI {
			participations = append(participations, participation)
		}
	}
	return participations, nil
}

// Legacy post operations (for backwards compatibility)

// CreatePost creates a new post with the given content
func (m *MockService) CreatePost(content string) (*Post, error) {
	id := fmt.Sprintf("mock-%d", len(m.posts)+1)
	post := &Post{ID: id, Content: content}
	m.posts[id] = post
	return post, nil
}

// GetPost retrieves a post by its ID
func (m *MockService) GetPost(id string) (*Post, error) {
	post, ok := m.posts[id]
	if !ok {
		return nil, nil
	}
	return post, nil
}

// SetSelectedAnswer sets the selected answer for a post
func (m *MockService) SetSelectedAnswer(postID, answerID string) error {
	post, ok := m.posts[postID]
	if !ok {
		return fmt.Errorf("post %s not found", postID)
	}
	post.SelectedAnswer = answerID
	return nil
}
