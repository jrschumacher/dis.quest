package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/logger"
)

// Service wraps the database connection and provides methods for database operations
type Service struct {
	db      *sql.DB
	queries *Queries
	driver  DatabaseDriver
}

// NewService creates a new database service instance
func NewService(cfg *config.Config) (*Service, error) {
	db, driver, err := OpenDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	queries := New(db)
	
	logger.Info("Database service initialized", 
		"driver", string(driver),
		"url", cfg.DatabaseURL)

	return &Service{
		db:      db,
		queries: queries,
		driver:  driver,
	}, nil
}

// Close closes the database connection
func (s *Service) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Queries returns the SQLC queries instance
func (s *Service) Queries() *Queries {
	return s.queries
}

// DB returns the underlying database connection
func (s *Service) DB() *sql.DB {
	return s.db
}

// Driver returns the database driver type
func (s *Service) Driver() DatabaseDriver {
	return s.driver
}

// IsPostgreSQL returns true if using PostgreSQL
func (s *Service) IsPostgreSQL() bool {
	return s.driver == PostgreSQL
}

// IsSQLite returns true if using SQLite
func (s *Service) IsSQLite() bool {
	return s.driver == SQLite
}

// WithTx executes a function within a database transaction
func (s *Service) WithTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	queries := s.queries.WithTx(tx)
	if err := fn(queries); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// CreateTopicWithParticipation creates a topic and automatically adds the creator as a participant
// This is an example of a complex operation that requires a transaction
func (s *Service) CreateTopicWithParticipation(ctx context.Context, params CreateTopicWithParticipationParams) (*TopicWithParticipation, error) {
	var result TopicWithParticipation
	
	err := s.WithTx(ctx, func(q *Queries) error {
		// Create the topic
		topic, err := q.CreateTopic(ctx, CreateTopicParams{
			Did:            params.Did,
			Rkey:           params.Rkey,
			Subject:        params.Subject,
			InitialMessage: params.InitialMessage,
			Category:       params.Category,
			CreatedAt:      params.CreatedAt,
			UpdatedAt:      params.UpdatedAt,
			SelectedAnswer: sql.NullString{}, // No selected answer initially
		})
		if err != nil {
			return fmt.Errorf("failed to create topic: %w", err)
		}
		
		// Create participation record for the topic creator
		participation, err := q.CreateParticipation(ctx, CreateParticipationParams{
			Did:       params.Did,
			TopicDid:  params.Did,
			TopicRkey: params.Rkey,
			Status:    "active", // Creator is automatically active
			CreatedAt: params.CreatedAt,
			UpdatedAt: params.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create participation: %w", err)
		}
		
		result.Topic = topic
		result.Participation = participation
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return &result, nil
}

// CreateTopicWithParticipationParams represents the parameters for creating a topic with participation
type CreateTopicWithParticipationParams struct {
	Did            string
	Rkey           string
	Subject        string
	InitialMessage string
	Category       sql.NullString
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TopicWithParticipation represents a topic along with the creator's participation
type TopicWithParticipation struct {
	Topic         Topic
	Participation Participation
}