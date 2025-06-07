// Package pds provides ATProtocol operations for Personal Data Server interactions
package pds

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ATProtoRecord represents a generic ATProtocol record
type ATProtoRecord struct {
	Collection string                 `json:"collection"`
	Rkey       string                 `json:"rkey"`
	Value      map[string]interface{} `json:"value"`
	CID        string                 `json:"cid,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

// Topic represents a quest.dis.topic record
type Topic struct {
	Type           string    `json:"$type"`
	Subject        string    `json:"subject"`
	InitialMessage string    `json:"initialMessage"`
	Category       *string   `json:"category,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	// ATProto metadata
	Rkey string `json:"rkey,omitempty"`
	CID  string `json:"cid,omitempty"`
}

// Message represents a quest.dis.message record
type Message struct {
	Type          string     `json:"$type"`
	Content       string     `json:"content"`
	TopicRkey     string     `json:"topicRkey"`
	ParentRkey    *string    `json:"parentRkey,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	// ATProto metadata
	Rkey string `json:"rkey,omitempty"`
	CID  string `json:"cid,omitempty"`
}

// Participation represents a quest.dis.participation record
type Participation struct {
	Type      string    `json:"$type"`
	TopicRkey string    `json:"topicRkey"`
	Status    string    `json:"status"` // "following", "muted", "blocked"
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	// ATProto metadata
	Rkey string `json:"rkey,omitempty"`
	CID  string `json:"cid,omitempty"`
}

// ATProtoService defines the interface for ATProtocol PDS operations
type ATProtoService interface {
	// Record operations
	CreateRecord(ctx context.Context, collection string, rkey string, record interface{}) (*ATProtoRecord, error)
	GetRecord(ctx context.Context, collection string, rkey string) (*ATProtoRecord, error)
	UpdateRecord(ctx context.Context, collection string, rkey string, record interface{}) (*ATProtoRecord, error)
	DeleteRecord(ctx context.Context, collection string, rkey string) error
	ListRecords(ctx context.Context, collection string, limit int, cursor string) ([]*ATProtoRecord, string, error)

	// Quest-specific operations
	CreateTopic(ctx context.Context, subject, initialMessage string, category *string) (*Topic, error)
	GetTopic(ctx context.Context, rkey string) (*Topic, error)
	CreateMessage(ctx context.Context, content, topicRkey string, parentRkey *string) (*Message, error)
	GetMessage(ctx context.Context, rkey string) (*Message, error)
	SetParticipation(ctx context.Context, topicRkey, status string) (*Participation, error)
	GetParticipation(ctx context.Context, topicRkey string) (*Participation, error)
}

// MockATProtoService provides a mock implementation for testing
type MockATProtoService struct {
	records map[string]*ATProtoRecord
	counter int
}

// NewMockATProtoService creates a new mock ATProtocol service
func NewMockATProtoService() *MockATProtoService {
	return &MockATProtoService{
		records: make(map[string]*ATProtoRecord),
		counter: 1,
	}
}

// CreateRecord creates a new ATProtocol record
func (m *MockATProtoService) CreateRecord(_ context.Context, collection string, rkey string, record interface{}) (*ATProtoRecord, error) {
	if rkey == "" {
		rkey = fmt.Sprintf("mock-rkey-%d", m.counter)
		m.counter++
	}

	// Convert record to map for storage
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %w", err)
	}
	
	var recordMap map[string]interface{}
	if err := json.Unmarshal(recordBytes, &recordMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	m.counter++
	atRecord := &ATProtoRecord{
		Collection: collection,
		Rkey:       rkey,
		Value:      recordMap,
		CID:        fmt.Sprintf("mock-cid-%d", m.counter),
		CreatedAt:  time.Now(),
	}

	recordKey := fmt.Sprintf("%s/%s", collection, rkey)
	m.records[recordKey] = atRecord
	return atRecord, nil
}

// GetRecord retrieves an ATProtocol record
func (m *MockATProtoService) GetRecord(_ context.Context, collection string, rkey string) (*ATProtoRecord, error) {
	key := fmt.Sprintf("%s/%s", collection, rkey)
	record, exists := m.records[key]
	if !exists {
		return nil, fmt.Errorf("record not found: %s", key)
	}
	return record, nil
}

// UpdateRecord updates an existing ATProtocol record
func (m *MockATProtoService) UpdateRecord(_ context.Context, collection string, rkey string, record interface{}) (*ATProtoRecord, error) {
	key := fmt.Sprintf("%s/%s", collection, rkey)
	existing, exists := m.records[key]
	if !exists {
		return nil, fmt.Errorf("record not found: %s", key)
	}

	// Convert record to map for storage
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %w", err)
	}
	
	var recordMap map[string]interface{}
	if err := json.Unmarshal(recordBytes, &recordMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	m.counter++
	existing.Value = recordMap
	existing.CID = fmt.Sprintf("mock-cid-%d", m.counter)

	return existing, nil
}

// DeleteRecord deletes an ATProtocol record
func (m *MockATProtoService) DeleteRecord(_ context.Context, collection string, rkey string) error {
	key := fmt.Sprintf("%s/%s", collection, rkey)
	if _, exists := m.records[key]; !exists {
		return fmt.Errorf("record not found: %s", key)
	}
	delete(m.records, key)
	return nil
}

// ListRecords lists ATProtocol records in a collection
func (m *MockATProtoService) ListRecords(_ context.Context, collection string, limit int, _ string) ([]*ATProtoRecord, string, error) {
	var records []*ATProtoRecord
	var nextCursor string

	for _, record := range m.records {
		if record.Collection == collection {
			records = append(records, record)
		}
	}

	// Simple pagination - just return all for now
	if len(records) > limit {
		records = records[:limit]
		nextCursor = "mock-cursor"
	}

	return records, nextCursor, nil
}

// CreateTopic creates a quest.dis.topic record
func (m *MockATProtoService) CreateTopic(ctx context.Context, subject, initialMessage string, category *string) (*Topic, error) {
	topic := &Topic{
		Type:           "quest.dis.topic",
		Subject:        subject,
		InitialMessage: initialMessage,
		Category:       category,
		CreatedAt:      time.Now(),
	}

	record, err := m.CreateRecord(ctx, "quest.dis.topic", "", topic)
	if err != nil {
		return nil, err
	}

	topic.Rkey = record.Rkey
	topic.CID = record.CID
	return topic, nil
}

// GetTopic retrieves a quest.dis.topic record
func (m *MockATProtoService) GetTopic(ctx context.Context, rkey string) (*Topic, error) {
	record, err := m.GetRecord(ctx, "quest.dis.topic", rkey)
	if err != nil {
		return nil, err
	}

	recordBytes, err := json.Marshal(record.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal topic: %w", err)
	}

	var topic Topic
	if err := json.Unmarshal(recordBytes, &topic); err != nil {
		return nil, fmt.Errorf("failed to unmarshal topic: %w", err)
	}

	topic.Rkey = record.Rkey
	topic.CID = record.CID
	return &topic, nil
}

// CreateMessage creates a quest.dis.message record
func (m *MockATProtoService) CreateMessage(ctx context.Context, content, topicRkey string, parentRkey *string) (*Message, error) {
	message := &Message{
		Type:       "quest.dis.message",
		Content:    content,
		TopicRkey:  topicRkey,
		ParentRkey: parentRkey,
		CreatedAt:  time.Now(),
	}

	record, err := m.CreateRecord(ctx, "quest.dis.message", "", message)
	if err != nil {
		return nil, err
	}

	message.Rkey = record.Rkey
	message.CID = record.CID
	return message, nil
}

// GetMessage retrieves a quest.dis.message record
func (m *MockATProtoService) GetMessage(ctx context.Context, rkey string) (*Message, error) {
	record, err := m.GetRecord(ctx, "quest.dis.message", rkey)
	if err != nil {
		return nil, err
	}

	recordBytes, err := json.Marshal(record.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	var message Message
	if err := json.Unmarshal(recordBytes, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	message.Rkey = record.Rkey
	message.CID = record.CID
	return &message, nil
}

// SetParticipation creates or updates a quest.dis.participation record
func (m *MockATProtoService) SetParticipation(ctx context.Context, topicRkey, status string) (*Participation, error) {
	participation := &Participation{
		Type:      "quest.dis.participation",
		TopicRkey: topicRkey,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Use topicRkey as the rkey for participation records (one per topic)
	record, err := m.CreateRecord(ctx, "quest.dis.participation", topicRkey, participation)
	if err != nil {
		return nil, err
	}

	participation.Rkey = record.Rkey
	participation.CID = record.CID
	return participation, nil
}

// GetParticipation retrieves a quest.dis.participation record
func (m *MockATProtoService) GetParticipation(ctx context.Context, topicRkey string) (*Participation, error) {
	record, err := m.GetRecord(ctx, "quest.dis.participation", topicRkey)
	if err != nil {
		return nil, err
	}

	recordBytes, err := json.Marshal(record.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal participation: %w", err)
	}

	var participation Participation
	if err := json.Unmarshal(recordBytes, &participation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal participation: %w", err)
	}

	participation.Rkey = record.Rkey
	participation.CID = record.CID
	return &participation, nil
}