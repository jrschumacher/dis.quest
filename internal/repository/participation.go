package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jrschumacher/dis.quest/internal/db"
)

// participationRepository implements ParticipationRepository
type participationRepository struct {
	dbService *db.Service
}

// CreateParticipation creates a new participation record
func (r *participationRepository) CreateParticipation(ctx context.Context, params CreateParticipationParams) (*ParticipationDetail, error) {
	now := time.Now()
	
	participation, err := r.dbService.Queries().CreateParticipation(ctx, db.CreateParticipationParams{
		Did:       params.Did,
		TopicDid:  params.TopicDID,
		TopicRkey: params.TopicRkey,
		Status:    params.Status,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create participation: %w", err)
	}
	
	return &ParticipationDetail{
		Did:       participation.Did,
		TopicDID:  participation.TopicDid,
		TopicRkey: participation.TopicRkey,
		Status:    participation.Status,
		CreatedAt: participation.CreatedAt,
		UpdatedAt: participation.UpdatedAt,
	}, nil
}

// GetParticipation retrieves a participation record
func (r *participationRepository) GetParticipation(ctx context.Context, userDID, topicDID, topicRkey string) (*ParticipationDetail, error) {
	participation, err := r.dbService.Queries().GetParticipation(ctx, db.GetParticipationParams{
		Did:       userDID,
		TopicDid:  topicDID,
		TopicRkey: topicRkey,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("participation not found")
		}
		return nil, fmt.Errorf("failed to get participation: %w", err)
	}
	
	return &ParticipationDetail{
		Did:       participation.Did,
		TopicDID:  participation.TopicDid,
		TopicRkey: participation.TopicRkey,
		Status:    participation.Status,
		CreatedAt: participation.CreatedAt,
		UpdatedAt: participation.UpdatedAt,
	}, nil
}

// GetParticipationsByUser retrieves all participations for a user
func (r *participationRepository) GetParticipationsByUser(ctx context.Context, userDID string) ([]*ParticipationDetail, error) {
	participations, err := r.dbService.Queries().GetParticipationsByUser(ctx, userDID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participations by user: %w", err)
	}
	
	details := make([]*ParticipationDetail, len(participations))
	for i, participation := range participations {
		details[i] = &ParticipationDetail{
			Did:       participation.Did,
			TopicDID:  participation.TopicDid,
			TopicRkey: participation.TopicRkey,
			Status:    participation.Status,
			CreatedAt: participation.CreatedAt,
			UpdatedAt: participation.UpdatedAt,
		}
	}
	
	return details, nil
}

// GetParticipationsByTopic retrieves all participations for a topic
func (r *participationRepository) GetParticipationsByTopic(ctx context.Context, topicDID, topicRkey string) ([]*ParticipationDetail, error) {
	participations, err := r.dbService.Queries().GetParticipationsByTopic(ctx, db.GetParticipationsByTopicParams{
		TopicDid:  topicDID,
		TopicRkey: topicRkey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get participations by topic: %w", err)
	}
	
	details := make([]*ParticipationDetail, len(participations))
	for i, participation := range participations {
		details[i] = &ParticipationDetail{
			Did:       participation.Did,
			TopicDID:  participation.TopicDid,
			TopicRkey: participation.TopicRkey,
			Status:    participation.Status,
			CreatedAt: participation.CreatedAt,
			UpdatedAt: participation.UpdatedAt,
		}
	}
	
	return details, nil
}

// UpdateParticipationStatus updates the status of a participation
func (r *participationRepository) UpdateParticipationStatus(ctx context.Context, userDID, topicDID, topicRkey, status string) error {
	err := r.dbService.Queries().UpdateParticipationStatus(ctx, db.UpdateParticipationStatusParams{
		Status:    status,
		UpdatedAt: time.Now(),
		Did:       userDID,
		TopicDid:  topicDID,
		TopicRkey: topicRkey,
	})
	if err != nil {
		return fmt.Errorf("failed to update participation status: %w", err)
	}
	
	return nil
}

// DeleteParticipation removes a participation record
func (r *participationRepository) DeleteParticipation(ctx context.Context, userDID, topicDID, topicRkey string) error {
	err := r.dbService.Queries().DeleteParticipation(ctx, db.DeleteParticipationParams{
		Did:       userDID,
		TopicDid:  topicDID,
		TopicRkey: topicRkey,
	})
	if err != nil {
		return fmt.Errorf("failed to delete participation: %w", err)
	}
	
	return nil
}