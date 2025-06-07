package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jrschumacher/dis.quest/internal/db"
)

// topicRepository implements TopicRepository
type topicRepository struct {
	dbService *db.Service
}

// CreateTopic creates a new topic with automatic participation
func (r *topicRepository) CreateTopic(ctx context.Context, params CreateTopicParams) (*TopicDetail, error) {
	now := time.Now()
	
	// Use the service's transaction-based method
	result, err := r.dbService.CreateTopicWithParticipation(ctx, db.CreateTopicWithParticipationParams{
		Did:            params.Did,
		Rkey:           params.Rkey,
		Subject:        params.Subject,
		InitialMessage: params.InitialMessage,
		Category:       sql.NullString{String: params.Category, Valid: params.Category != ""},
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create topic: %w", err)
	}
	
	// Convert to repository model
	return &TopicDetail{
		DID:            result.Topic.Did,
		Rkey:           result.Topic.Rkey,
		Subject:        result.Topic.Subject,
		InitialMessage: result.Topic.InitialMessage,
		Category:       result.Topic.Category.String,
		SelectedAnswer: result.Topic.SelectedAnswer.String,
		CreatedAt:      result.Topic.CreatedAt,
		UpdatedAt:      result.Topic.UpdatedAt,
		MessageCount:   0, // New topic has no messages yet
	}, nil
}

// GetTopic retrieves a topic by DID and rkey
func (r *topicRepository) GetTopic(ctx context.Context, did, rkey string) (*TopicDetail, error) {
	topic, err := r.dbService.Queries().GetTopic(ctx, db.GetTopicParams{
		Did:  did,
		Rkey: rkey,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("topic not found")
		}
		return nil, fmt.Errorf("failed to get topic: %w", err)
	}
	
	// Get message count for this topic
	messages, err := r.dbService.Queries().GetMessagesByTopic(ctx, db.GetMessagesByTopicParams{
		TopicDid:  did,
		TopicRkey: rkey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get message count: %w", err)
	}
	
	// Get participants
	participations, err := r.dbService.Queries().GetParticipationsByTopic(ctx, db.GetParticipationsByTopicParams{
		TopicDid:  did,
		TopicRkey: rkey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	
	participants := make([]ParticipantInfo, len(participations))
	for i, p := range participations {
		participants[i] = ParticipantInfo{
			DID:    p.Did,
			Status: p.Status,
		}
	}
	
	return &TopicDetail{
		DID:            topic.Did,
		Rkey:           topic.Rkey,
		Subject:        topic.Subject,
		InitialMessage: topic.InitialMessage,
		Category:       topic.Category.String,
		SelectedAnswer: topic.SelectedAnswer.String,
		CreatedAt:      topic.CreatedAt,
		UpdatedAt:      topic.UpdatedAt,
		MessageCount:   len(messages),
		Participants:   participants,
	}, nil
}

// ListTopics retrieves a paginated list of topics
func (r *topicRepository) ListTopics(ctx context.Context, params ListTopicsParams) ([]*TopicSummary, error) {
	topics, err := r.dbService.Queries().ListTopics(ctx, db.ListTopicsParams{
		Limit:  int64(params.Limit),
		Offset: int64(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}
	
	summaries := make([]*TopicSummary, len(topics))
	for i, topic := range topics {
		// Get message count for each topic
		messages, err := r.dbService.Queries().GetMessagesByTopic(ctx, db.GetMessagesByTopicParams{
			TopicDid:  topic.Did,
			TopicRkey: topic.Rkey,
		})
		messageCount := 0
		lastActivity := topic.CreatedAt
		if err == nil {
			messageCount = len(messages)
			// Find the most recent message timestamp
			for _, msg := range messages {
				if msg.CreatedAt.After(lastActivity) {
					lastActivity = msg.CreatedAt
				}
			}
		}
		
		summaries[i] = &TopicSummary{
			DID:          topic.Did,
			Rkey:         topic.Rkey,
			Subject:      topic.Subject,
			Category:     topic.Category.String,
			MessageCount: messageCount,
			LastActivity: lastActivity,
			CreatedAt:    topic.CreatedAt,
			HasAnswer:    topic.SelectedAnswer.Valid && topic.SelectedAnswer.String != "",
		}
	}
	
	return summaries, nil
}

// GetTopicsByCategory retrieves topics by category
func (r *topicRepository) GetTopicsByCategory(ctx context.Context, category string, limit int) ([]*TopicSummary, error) {
	topics, err := r.dbService.Queries().GetTopicsByCategory(ctx, db.GetTopicsByCategoryParams{
		Category: sql.NullString{String: category, Valid: category != ""},
		Limit:    int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get topics by category: %w", err)
	}
	
	summaries := make([]*TopicSummary, len(topics))
	for i, topic := range topics {
		// Get message count for each topic (simplified for category view)
		messages, err := r.dbService.Queries().GetMessagesByTopic(ctx, db.GetMessagesByTopicParams{
			TopicDid:  topic.Did,
			TopicRkey: topic.Rkey,
		})
		messageCount := 0
		if err == nil {
			messageCount = len(messages)
		}
		
		summaries[i] = &TopicSummary{
			DID:          topic.Did,
			Rkey:         topic.Rkey,
			Subject:      topic.Subject,
			Category:     topic.Category.String,
			MessageCount: messageCount,
			LastActivity: topic.UpdatedAt,
			CreatedAt:    topic.CreatedAt,
			HasAnswer:    topic.SelectedAnswer.Valid && topic.SelectedAnswer.String != "",
		}
	}
	
	return summaries, nil
}

// UpdateSelectedAnswer updates the selected answer for a Q&A topic
func (r *topicRepository) UpdateSelectedAnswer(ctx context.Context, topicDID, topicRkey, messageRkey string, userDID string) error {
	// First verify the user owns the topic
	topic, err := r.dbService.Queries().GetTopic(ctx, db.GetTopicParams{
		Did:  topicDID,
		Rkey: topicRkey,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("topic not found")
		}
		return fmt.Errorf("failed to get topic: %w", err)
	}
	
	if topic.Did != userDID {
		return fmt.Errorf("unauthorized: only topic creator can select answer")
	}
	
	// Update the selected answer
	err = r.dbService.Queries().UpdateTopicSelectedAnswer(ctx, db.UpdateTopicSelectedAnswerParams{
		SelectedAnswer: sql.NullString{String: messageRkey, Valid: messageRkey != ""},
		UpdatedAt:      time.Now(),
		Did:            topicDID,
		Rkey:           topicRkey,
	})
	if err != nil {
		return fmt.Errorf("failed to update selected answer: %w", err)
	}
	
	return nil
}