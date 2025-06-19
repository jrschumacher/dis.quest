package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jrschumacher/dis.quest/internal/testutil"
)

func TestTopicRepository_GetTopic_NotFound(t *testing.T) {
	// Create test database
	dbService := testutil.TestDatabase(t)
	repo := NewRepository(dbService)

	// Try to get a non-existent topic
	_, err := repo.Topics().GetTopic(context.Background(), "did:plc:nonexistent", "nonexistent-rkey")
	
	if err == nil {
		t.Fatal("expected error when topic not found")
	}
	
	if !errors.Is(err, ErrTopicNotFound) {
		t.Errorf("expected ErrTopicNotFound, got %v", err)
	}
}

func TestTopicRepository_UpdateSelectedAnswer_NotFound(t *testing.T) {
	// Create test database
	dbService := testutil.TestDatabase(t)
	repo := NewRepository(dbService)

	// Try to update selected answer for non-existent topic
	err := repo.Topics().UpdateSelectedAnswer(context.Background(), "did:plc:nonexistent", "nonexistent-rkey", "message-rkey", "did:plc:user")
	
	if err == nil {
		t.Fatal("expected error when topic not found")
	}
	
	if !errors.Is(err, ErrTopicNotFound) {
		t.Errorf("expected ErrTopicNotFound, got %v", err)
	}
}

func TestTopicRepository_UpdateSelectedAnswer_Unauthorized(t *testing.T) {
	// Create test database
	dbService := testutil.TestDatabase(t)
	repo := NewRepository(dbService)

	// Create a topic with one user
	ctx := context.Background()
	ownerDID := "did:plc:owner"
	otherDID := "did:plc:other"
	
	topic, err := repo.Topics().CreateTopic(ctx, CreateTopicParams{
		Did:            ownerDID,
		Rkey:           "test-topic",
		Subject:        "Test Topic",
		InitialMessage: "Test message",
		Category:       "test",
	})
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	// Try to update selected answer with different user
	err = repo.Topics().UpdateSelectedAnswer(ctx, topic.DID, topic.Rkey, "message-rkey", otherDID)
	
	if err == nil {
		t.Fatal("expected error when unauthorized user tries to update")
	}
	
	if !errors.Is(err, ErrTopicOwnershipRequired) {
		t.Errorf("expected ErrTopicOwnershipRequired, got %v", err)
	}
}