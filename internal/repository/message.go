package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jrschumacher/dis.quest/internal/db"
)

// messageRepository implements MessageRepository
type messageRepository struct {
	dbService *db.Service
}

// CreateMessage creates a new message
func (r *messageRepository) CreateMessage(ctx context.Context, params CreateMessageParams) (*MessageDetail, error) {
	now := time.Now()
	
	message, err := r.dbService.Queries().CreateMessage(ctx, db.CreateMessageParams{
		Did:               params.Did,
		Rkey:              params.Rkey,
		TopicDid:          params.TopicDID,
		TopicRkey:         params.TopicRkey,
		ParentMessageRkey: sql.NullString{String: params.ParentMessageRkey, Valid: params.ParentMessageRkey != ""},
		Content:           params.Content,
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}
	
	// Check if this message is the selected answer
	topic, err := r.dbService.Queries().GetTopic(ctx, db.GetTopicParams{
		Did:  params.TopicDID,
		Rkey: params.TopicRkey,
	})
	isAnswer := false
	if err == nil && topic.SelectedAnswer.Valid {
		isAnswer = topic.SelectedAnswer.String == params.Rkey
	}
	
	return &MessageDetail{
		DID:               message.Did,
		Rkey:              message.Rkey,
		TopicDID:          message.TopicDid,
		TopicRkey:         message.TopicRkey,
		ParentMessageRkey: message.ParentMessageRkey.String,
		Content:           message.Content,
		CreatedAt:         message.CreatedAt,
		UpdatedAt:         message.UpdatedAt,
		IsAnswer:          isAnswer,
		ReplyCount:        0, // New message has no replies
	}, nil
}

// GetMessage retrieves a message by DID and rkey
func (r *messageRepository) GetMessage(ctx context.Context, did, rkey string) (*MessageDetail, error) {
	message, err := r.dbService.Queries().GetMessage(ctx, db.GetMessageParams{
		Did:  did,
		Rkey: rkey,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	
	// Check if this message is the selected answer
	topic, err := r.dbService.Queries().GetTopic(ctx, db.GetTopicParams{
		Did:  message.TopicDid,
		Rkey: message.TopicRkey,
	})
	isAnswer := false
	if err == nil && topic.SelectedAnswer.Valid {
		isAnswer = topic.SelectedAnswer.String == rkey
	}
	
	// Get reply count
	replies, err := r.dbService.Queries().GetRepliesByMessage(ctx, db.GetRepliesByMessageParams{
		TopicDid:          message.TopicDid,
		TopicRkey:         message.TopicRkey,
		ParentMessageRkey: sql.NullString{String: rkey, Valid: true},
	})
	replyCount := 0
	if err == nil {
		replyCount = len(replies)
	}
	
	return &MessageDetail{
		DID:               message.Did,
		Rkey:              message.Rkey,
		TopicDID:          message.TopicDid,
		TopicRkey:         message.TopicRkey,
		ParentMessageRkey: message.ParentMessageRkey.String,
		Content:           message.Content,
		CreatedAt:         message.CreatedAt,
		UpdatedAt:         message.UpdatedAt,
		IsAnswer:          isAnswer,
		ReplyCount:        replyCount,
	}, nil
}

// GetMessagesByTopic retrieves all messages for a topic
func (r *messageRepository) GetMessagesByTopic(ctx context.Context, topicDID, topicRkey string) ([]*MessageDetail, error) {
	messages, err := r.dbService.Queries().GetMessagesByTopic(ctx, db.GetMessagesByTopicParams{
		TopicDid:  topicDID,
		TopicRkey: topicRkey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by topic: %w", err)
	}
	
	// Get topic to check for selected answer
	topic, err := r.dbService.Queries().GetTopic(ctx, db.GetTopicParams{
		Did:  topicDID,
		Rkey: topicRkey,
	})
	selectedAnswer := ""
	if err == nil && topic.SelectedAnswer.Valid {
		selectedAnswer = topic.SelectedAnswer.String
	}
	
	details := make([]*MessageDetail, len(messages))
	for i, message := range messages {
		// Get reply count for each message
		replies, err := r.dbService.Queries().GetRepliesByMessage(ctx, db.GetRepliesByMessageParams{
			TopicDid:          message.TopicDid,
			TopicRkey:         message.TopicRkey,
			ParentMessageRkey: sql.NullString{String: message.Rkey, Valid: true},
		})
		replyCount := 0
		if err == nil {
			replyCount = len(replies)
		}
		
		details[i] = &MessageDetail{
			DID:               message.Did,
			Rkey:              message.Rkey,
			TopicDID:          message.TopicDid,
			TopicRkey:         message.TopicRkey,
			ParentMessageRkey: message.ParentMessageRkey.String,
			Content:           message.Content,
			CreatedAt:         message.CreatedAt,
			UpdatedAt:         message.UpdatedAt,
			IsAnswer:          selectedAnswer == message.Rkey,
			ReplyCount:        replyCount,
		}
	}
	
	return details, nil
}

// GetRepliesByMessage retrieves replies to a specific message
func (r *messageRepository) GetRepliesByMessage(ctx context.Context, topicDID, topicRkey, parentRkey string) ([]*MessageDetail, error) {
	replies, err := r.dbService.Queries().GetRepliesByMessage(ctx, db.GetRepliesByMessageParams{
		TopicDid:          topicDID,
		TopicRkey:         topicRkey,
		ParentMessageRkey: sql.NullString{String: parentRkey, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get replies: %w", err)
	}
	
	details := make([]*MessageDetail, len(replies))
	for i, reply := range replies {
		details[i] = &MessageDetail{
			DID:               reply.Did,
			Rkey:              reply.Rkey,
			TopicDID:          reply.TopicDid,
			TopicRkey:         reply.TopicRkey,
			ParentMessageRkey: reply.ParentMessageRkey.String,
			Content:           reply.Content,
			CreatedAt:         reply.CreatedAt,
			UpdatedAt:         reply.UpdatedAt,
			IsAnswer:          false, // Replies can't be selected answers
			ReplyCount:        0,     // We don't support nested replies yet
		}
	}
	
	return details, nil
}

// DeleteMessage deletes a message if the user owns it
func (r *messageRepository) DeleteMessage(ctx context.Context, did, rkey string, userDID string) error {
	// First verify the user owns the message
	message, err := r.dbService.Queries().GetMessage(ctx, db.GetMessageParams{
		Did:  did,
		Rkey: rkey,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("message not found")
		}
		return fmt.Errorf("failed to get message: %w", err)
	}
	
	if message.Did != userDID {
		return fmt.Errorf("unauthorized: only message author can delete")
	}
	
	// Delete the message
	err = r.dbService.Queries().DeleteMessage(ctx, db.DeleteMessageParams{
		Did:  did,
		Rkey: rkey,
	})
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	
	return nil
}