// Package lexicons provides quest.dis.* lexicon definitions and utilities for this application
package lexicons

import (
	"fmt"
	"time"
)

// Lexicon collection names for quest.dis.* schemas
const (
	TopicLexicon         = "quest.dis.topic"
	MessageLexicon       = "quest.dis.message"
	ParticipationLexicon = "quest.dis.participation"
)

// TopicRecord represents a quest.dis.topic lexicon record
type TopicRecord struct {
	Type           string    `json:"$type"`
	Title          string    `json:"title"`
	Summary        string    `json:"summary,omitempty"`
	Tags           []string  `json:"tags,omitempty"`
	CreatedBy      string    `json:"createdBy"`
	CreatedAt      time.Time `json:"createdAt"`
	SelectedAnswer string    `json:"selectedAnswer,omitempty"`
}

// NewTopicRecord creates a new topic record with required fields
func NewTopicRecord(title, createdBy string) *TopicRecord {
	return &TopicRecord{
		Type:      TopicLexicon,
		Title:     title,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}
}

// ToMap converts TopicRecord to map for PDS operations
func (t *TopicRecord) ToMap() map[string]interface{} {
	record := map[string]interface{}{
		"$type":     t.Type,
		"title":     t.Title,
		"createdBy": t.CreatedBy,
		"createdAt": t.CreatedAt.Format(time.RFC3339),
	}

	if t.Summary != "" {
		record["summary"] = t.Summary
	}
	if len(t.Tags) > 0 {
		record["tags"] = t.Tags
	}
	if t.SelectedAnswer != "" {
		record["selectedAnswer"] = t.SelectedAnswer
	}

	return record
}

// ToTopic converts TopicRecord to service.Topic with metadata
func (t *TopicRecord) ToTopic(uri, cid string) *Topic {
	return &Topic{
		URI:            uri,
		CID:            cid,
		Title:          t.Title,
		Summary:        t.Summary,
		Tags:           t.Tags,
		CreatedBy:      t.CreatedBy,
		CreatedAt:      t.CreatedAt,
		SelectedAnswer: t.SelectedAnswer,
	}
}

// TopicRecordFromMap creates TopicRecord from map (for parsing PDS responses)
func TopicRecordFromMap(data map[string]interface{}) (*TopicRecord, error) {
	record := &TopicRecord{}

	// Required fields
	if typeVal, ok := data["$type"].(string); ok {
		record.Type = typeVal
	} else {
		return nil, fmt.Errorf("missing or invalid $type field")
	}

	if title, ok := data["title"].(string); ok {
		record.Title = title
	} else {
		return nil, fmt.Errorf("missing or invalid title field")
	}

	if createdBy, ok := data["createdBy"].(string); ok {
		record.CreatedBy = createdBy
	} else {
		return nil, fmt.Errorf("missing or invalid createdBy field")
	}

	if createdAtStr, ok := data["createdAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			record.CreatedAt = t
		} else {
			return nil, fmt.Errorf("invalid createdAt format: %w", err)
		}
	} else {
		return nil, fmt.Errorf("missing or invalid createdAt field")
	}

	// Optional fields
	if summary, ok := data["summary"].(string); ok {
		record.Summary = summary
	}

	if tagsInterface, ok := data["tags"].([]interface{}); ok {
		var tags []string
		for _, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
		record.Tags = tags
	}

	if selectedAnswer, ok := data["selectedAnswer"].(string); ok {
		record.SelectedAnswer = selectedAnswer
	}

	return record, nil
}

// MessageRecord represents a quest.dis.message lexicon record
type MessageRecord struct {
	Type      string    `json:"$type"`
	Topic     string    `json:"topic"`
	Content   string    `json:"content"`
	ReplyTo   string    `json:"replyTo,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// NewMessageRecord creates a new message record with required fields
func NewMessageRecord(topic, content string) *MessageRecord {
	return &MessageRecord{
		Type:      MessageLexicon,
		Topic:     topic,
		Content:   content,
		CreatedAt: time.Now(),
	}
}

// ToMap converts MessageRecord to map for PDS operations
func (m *MessageRecord) ToMap() map[string]interface{} {
	record := map[string]interface{}{
		"$type":     m.Type,
		"topic":     m.Topic,
		"content":   m.Content,
		"createdAt": m.CreatedAt.Format(time.RFC3339),
	}

	if m.ReplyTo != "" {
		record["replyTo"] = m.ReplyTo
	}

	return record
}

// MessageRecordFromMap creates MessageRecord from map (for parsing PDS responses)
func MessageRecordFromMap(data map[string]interface{}) (*MessageRecord, error) {
	record := &MessageRecord{}

	// Required fields
	if typeVal, ok := data["$type"].(string); ok {
		record.Type = typeVal
	} else {
		return nil, fmt.Errorf("missing or invalid $type field")
	}

	if topic, ok := data["topic"].(string); ok {
		record.Topic = topic
	} else {
		return nil, fmt.Errorf("missing or invalid topic field")
	}

	if content, ok := data["content"].(string); ok {
		record.Content = content
	} else {
		return nil, fmt.Errorf("missing or invalid content field")
	}

	if createdAtStr, ok := data["createdAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			record.CreatedAt = t
		} else {
			return nil, fmt.Errorf("invalid createdAt format: %w", err)
		}
	} else {
		return nil, fmt.Errorf("missing or invalid createdAt field")
	}

	// Optional fields
	if replyTo, ok := data["replyTo"].(string); ok {
		record.ReplyTo = replyTo
	}

	return record, nil
}

// ParticipationRecord represents a quest.dis.participation lexicon record
type ParticipationRecord struct {
	Type        string    `json:"$type"`
	Topic       string    `json:"topic"`
	Participant string    `json:"participant"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joinedAt"`
}

// NewParticipationRecord creates a new participation record with required fields
func NewParticipationRecord(topic, participant, role string) *ParticipationRecord {
	return &ParticipationRecord{
		Type:        ParticipationLexicon,
		Topic:       topic,
		Participant: participant,
		Role:        role,
		JoinedAt:    time.Now(),
	}
}

// ToMap converts ParticipationRecord to map for PDS operations
func (p *ParticipationRecord) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"$type":       p.Type,
		"topic":       p.Topic,
		"participant": p.Participant,
		"role":        p.Role,
		"joinedAt":    p.JoinedAt.Format(time.RFC3339),
	}
}

// ParticipationRecordFromMap creates ParticipationRecord from map (for parsing PDS responses)
func ParticipationRecordFromMap(data map[string]interface{}) (*ParticipationRecord, error) {
	record := &ParticipationRecord{}

	// Required fields
	if typeVal, ok := data["$type"].(string); ok {
		record.Type = typeVal
	} else {
		return nil, fmt.Errorf("missing or invalid $type field")
	}

	if topic, ok := data["topic"].(string); ok {
		record.Topic = topic
	} else {
		return nil, fmt.Errorf("missing or invalid topic field")
	}

	if participant, ok := data["participant"].(string); ok {
		record.Participant = participant
	} else {
		return nil, fmt.Errorf("missing or invalid participant field")
	}

	if role, ok := data["role"].(string); ok {
		record.Role = role
	} else {
		return nil, fmt.Errorf("missing or invalid role field")
	}

	if joinedAtStr, ok := data["joinedAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, joinedAtStr); err == nil {
			record.JoinedAt = t
		} else {
			return nil, fmt.Errorf("invalid joinedAt format: %w", err)
		}
	} else {
		return nil, fmt.Errorf("missing or invalid joinedAt field")
	}

	return record, nil
}