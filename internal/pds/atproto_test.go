package pds

import (
	"context"
	"testing"
	"time"
)

func TestMockATProtoService_CreateAndGetTopic(t *testing.T) {
	service := NewMockATProtoService()
	ctx := context.Background()

	// Test topic creation
	subject := "Test Discussion Topic"
	initialMessage := "This is a test topic for unit testing"
	category := "testing"

	topic, err := service.CreateTopic(ctx, subject, initialMessage, &category)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Verify topic properties
	if topic.Type != "quest.dis.topic" {
		t.Errorf("Expected type 'quest.dis.topic', got '%s'", topic.Type)
	}
	if topic.Subject != subject {
		t.Errorf("Expected subject '%s', got '%s'", subject, topic.Subject)
	}
	if topic.InitialMessage != initialMessage {
		t.Errorf("Expected initial message '%s', got '%s'", initialMessage, topic.InitialMessage)
	}
	if topic.Category == nil || *topic.Category != category {
		t.Errorf("Expected category '%s', got '%v'", category, topic.Category)
	}
	if topic.Rkey == "" {
		t.Error("Expected rkey to be generated")
	}
	if topic.CID == "" {
		t.Error("Expected CID to be generated")
	}
	if topic.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	// Test topic retrieval
	retrievedTopic, err := service.GetTopic(ctx, topic.Rkey)
	if err != nil {
		t.Fatalf("Failed to get topic: %v", err)
	}

	if retrievedTopic.Subject != topic.Subject {
		t.Errorf("Retrieved topic subject mismatch: expected '%s', got '%s'", topic.Subject, retrievedTopic.Subject)
	}
	if retrievedTopic.Rkey != topic.Rkey {
		t.Errorf("Retrieved topic rkey mismatch: expected '%s', got '%s'", topic.Rkey, retrievedTopic.Rkey)
	}
}

func TestMockATProtoService_CreateAndGetMessage(t *testing.T) {
	service := NewMockATProtoService()
	ctx := context.Background()

	// First create a topic
	topic, err := service.CreateTopic(ctx, "Test Topic", "Initial message", nil)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Test message creation
	content := "This is a test message"
	message, err := service.CreateMessage(ctx, content, topic.Rkey, nil)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Verify message properties
	if message.Type != "quest.dis.message" {
		t.Errorf("Expected type 'quest.dis.message', got '%s'", message.Type)
	}
	if message.Content != content {
		t.Errorf("Expected content '%s', got '%s'", content, message.Content)
	}
	if message.TopicRkey != topic.Rkey {
		t.Errorf("Expected topic rkey '%s', got '%s'", topic.Rkey, message.TopicRkey)
	}
	if message.ParentRkey != nil {
		t.Errorf("Expected parent rkey to be nil, got '%v'", message.ParentRkey)
	}
	if message.Rkey == "" {
		t.Error("Expected rkey to be generated")
	}

	// Test reply message creation
	replyContent := "This is a reply message"
	replyMessage, err := service.CreateMessage(ctx, replyContent, topic.Rkey, &message.Rkey)
	if err != nil {
		t.Fatalf("Failed to create reply message: %v", err)
	}

	if replyMessage.ParentRkey == nil || *replyMessage.ParentRkey != message.Rkey {
		t.Errorf("Expected parent rkey '%s', got '%v'", message.Rkey, replyMessage.ParentRkey)
	}

	// Test message retrieval
	retrievedMessage, err := service.GetMessage(ctx, message.Rkey)
	if err != nil {
		t.Fatalf("Failed to get message: %v", err)
	}

	if retrievedMessage.Content != message.Content {
		t.Errorf("Retrieved message content mismatch: expected '%s', got '%s'", message.Content, retrievedMessage.Content)
	}
}

func TestMockATProtoService_SetAndGetParticipation(t *testing.T) {
	service := NewMockATProtoService()
	ctx := context.Background()

	// First create a topic
	topic, err := service.CreateTopic(ctx, "Test Topic", "Initial message", nil)
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Test participation creation
	status := "following"
	participation, err := service.SetParticipation(ctx, topic.Rkey, status)
	if err != nil {
		t.Fatalf("Failed to set participation: %v", err)
	}

	// Verify participation properties
	if participation.Type != "quest.dis.participation" {
		t.Errorf("Expected type 'quest.dis.participation', got '%s'", participation.Type)
	}
	if participation.TopicRkey != topic.Rkey {
		t.Errorf("Expected topic rkey '%s', got '%s'", topic.Rkey, participation.TopicRkey)
	}
	if participation.Status != status {
		t.Errorf("Expected status '%s', got '%s'", status, participation.Status)
	}
	if participation.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if participation.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	// Test participation retrieval
	retrievedParticipation, err := service.GetParticipation(ctx, topic.Rkey)
	if err != nil {
		t.Fatalf("Failed to get participation: %v", err)
	}

	if retrievedParticipation.Status != participation.Status {
		t.Errorf("Retrieved participation status mismatch: expected '%s', got '%s'", participation.Status, retrievedParticipation.Status)
	}
}

func TestMockATProtoService_GenericRecordOperations(t *testing.T) {
	service := NewMockATProtoService()
	ctx := context.Background()

	// Test custom record creation
	collection := "custom.test.record"
	rkey := "test-record-1"
	recordData := map[string]interface{}{
		"$type":     "custom.test.record",
		"title":     "Test Record",
		"content":   "This is test content",
		"createdAt": time.Now(),
	}

	record, err := service.CreateRecord(ctx, collection, rkey, recordData)
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	if record.Collection != collection {
		t.Errorf("Expected collection '%s', got '%s'", collection, record.Collection)
	}
	if record.Rkey != rkey {
		t.Errorf("Expected rkey '%s', got '%s'", rkey, record.Rkey)
	}
	if record.CID == "" {
		t.Error("Expected CID to be generated")
	}

	// Test record retrieval
	retrievedRecord, err := service.GetRecord(ctx, collection, rkey)
	if err != nil {
		t.Fatalf("Failed to get record: %v", err)
	}

	if retrievedRecord.Value["title"] != recordData["title"] {
		t.Errorf("Retrieved record title mismatch")
	}

	// Test record update
	originalCID := record.CID
	updatedData := map[string]interface{}{
		"$type":     "custom.test.record",
		"title":     "Updated Test Record",
		"content":   "This is updated content",
		"createdAt": time.Now(),
	}

	updatedRecord, err := service.UpdateRecord(ctx, collection, rkey, updatedData)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	if updatedRecord.Value["title"] != updatedData["title"] {
		t.Errorf("Updated record title mismatch")
	}
	if updatedRecord.CID == originalCID {
		t.Errorf("Expected CID to change after update. Original: %s, Updated: %s", originalCID, updatedRecord.CID)
	}

	// Test record deletion
	err = service.DeleteRecord(ctx, collection, rkey)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// Verify record is deleted
	_, err = service.GetRecord(ctx, collection, rkey)
	if err == nil {
		t.Error("Expected error when getting deleted record")
	}
}

func TestMockATProtoService_ListRecords(t *testing.T) {
	service := NewMockATProtoService()
	ctx := context.Background()

	collection := "quest.dis.topic"

	// Create multiple topics
	topics := []string{"Topic 1", "Topic 2", "Topic 3"}
	for _, subject := range topics {
		_, err := service.CreateTopic(ctx, subject, "Initial message", nil)
		if err != nil {
			t.Fatalf("Failed to create topic: %v", err)
		}
	}

	// List records
	records, cursor, err := service.ListRecords(ctx, collection, 10, "")
	if err != nil {
		t.Fatalf("Failed to list records: %v", err)
	}

	if len(records) != len(topics) {
		t.Errorf("Expected %d records, got %d", len(topics), len(records))
	}

	for _, record := range records {
		if record.Collection != collection {
			t.Errorf("Expected collection '%s', got '%s'", collection, record.Collection)
		}
	}

	// Test with limit
	limitedRecords, limitedCursor, err := service.ListRecords(ctx, collection, 2, "")
	if err != nil {
		t.Fatalf("Failed to list limited records: %v", err)
	}

	if len(limitedRecords) != 2 {
		t.Errorf("Expected 2 limited records, got %d", len(limitedRecords))
	}
	if limitedCursor == "" {
		t.Error("Expected cursor to be set when limiting results")
	}

	// Test cursor handling
	if cursor != "" {
		t.Error("Expected no cursor when all records fit in limit")
	}
}