// Package pds provides lexicon definitions and utilities for quest.dis.* schemas
package pds

import (
	"fmt"
	"time"
)

// Lexicon collection names
const (
	TopicLexicon         = "quest.dis.topic"
	MessageLexicon       = "quest.dis.message"
	ParticipationLexicon = "quest.dis.participation"
)

// LexiconRecord represents a generic lexicon record with common fields
type LexiconRecord struct {
	Type      string                 `json:"$type"`
	CreatedAt time.Time              `json:"createdAt"`
	Data      map[string]interface{} `json:"-"` // Additional fields
}

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

// ToMap converts TopicRecord to map for XRPC calls
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

// FromMap creates a TopicRecord from XRPC response data
func (t *TopicRecord) FromMap(data map[string]interface{}) error {
	if typeVal, ok := data["$type"].(string); ok {
		t.Type = typeVal
	}
	if title, ok := data["title"].(string); ok {
		t.Title = title
	}
	if summary, ok := data["summary"].(string); ok {
		t.Summary = summary
	}
	if createdBy, ok := data["createdBy"].(string); ok {
		t.CreatedBy = createdBy
	}
	if selectedAnswer, ok := data["selectedAnswer"].(string); ok {
		t.SelectedAnswer = selectedAnswer
	}

	// Parse createdAt
	if createdAtStr, ok := data["createdAt"].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			t.CreatedAt = parsedTime
		}
	}

	// Parse tags array
	if tagsIface, ok := data["tags"].([]interface{}); ok {
		t.Tags = make([]string, 0, len(tagsIface))
		for _, tag := range tagsIface {
			if tagStr, ok := tag.(string); ok {
				t.Tags = append(t.Tags, tagStr)
			}
		}
	}

	return nil
}

// ToTopic converts TopicRecord to Topic struct
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

// MessageRecord represents a quest.dis.message lexicon record
type MessageRecord struct {
	Type      string    `json:"$type"`
	Topic     string    `json:"topic"`
	Content   string    `json:"content"`
	ReplyTo   string    `json:"replyTo,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// ToMap converts MessageRecord to map for XRPC calls
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

// FromMap creates a MessageRecord from XRPC response data
func (m *MessageRecord) FromMap(data map[string]interface{}) error {
	if typeVal, ok := data["$type"].(string); ok {
		m.Type = typeVal
	}
	if topic, ok := data["topic"].(string); ok {
		m.Topic = topic
	}
	if content, ok := data["content"].(string); ok {
		m.Content = content
	}
	if replyTo, ok := data["replyTo"].(string); ok {
		m.ReplyTo = replyTo
	}

	// Parse createdAt
	if createdAtStr, ok := data["createdAt"].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			m.CreatedAt = parsedTime
		}
	}

	return nil
}

// ToMessage converts MessageRecord to Message struct
func (m *MessageRecord) ToMessage(uri, cid string) *Message {
	return &Message{
		URI:       uri,
		CID:       cid,
		Topic:     m.Topic,
		Content:   m.Content,
		ReplyTo:   m.ReplyTo,
		CreatedAt: m.CreatedAt,
	}
}

// ParticipationRecord represents a quest.dis.participation lexicon record
type ParticipationRecord struct {
	Type        string    `json:"$type"`
	Topic       string    `json:"topic"`
	Participant string    `json:"participant"`
	Role        string    `json:"role,omitempty"`
	JoinedAt    time.Time `json:"joinedAt"`
}

// ToMap converts ParticipationRecord to map for XRPC calls
func (p *ParticipationRecord) ToMap() map[string]interface{} {
	record := map[string]interface{}{
		"$type":       p.Type,
		"topic":       p.Topic,
		"participant": p.Participant,
		"joinedAt":    p.JoinedAt.Format(time.RFC3339),
	}

	if p.Role != "" {
		record["role"] = p.Role
	}

	return record
}

// FromMap creates a ParticipationRecord from XRPC response data
func (p *ParticipationRecord) FromMap(data map[string]interface{}) error {
	if typeVal, ok := data["$type"].(string); ok {
		p.Type = typeVal
	}
	if topic, ok := data["topic"].(string); ok {
		p.Topic = topic
	}
	if participant, ok := data["participant"].(string); ok {
		p.Participant = participant
	}
	if role, ok := data["role"].(string); ok {
		p.Role = role
	}

	// Parse joinedAt
	if joinedAtStr, ok := data["joinedAt"].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, joinedAtStr); err == nil {
			p.JoinedAt = parsedTime
		}
	}

	return nil
}

// ToParticipation converts ParticipationRecord to Participation struct
func (p *ParticipationRecord) ToParticipation(uri, cid string) *Participation {
	return &Participation{
		URI:         uri,
		CID:         cid,
		Topic:       p.Topic,
		Participant: p.Participant,
		Role:        p.Role,
		JoinedAt:    p.JoinedAt,
	}
}

// GenerateRKey generates a unique record key for lexicon records
func GenerateRKey(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// ValidateLexicon validates that a record conforms to expected lexicon schema
func ValidateLexicon(lexicon string, record map[string]interface{}) error {
	// Check required $type field
	if typeVal, ok := record["$type"].(string); !ok || typeVal != lexicon {
		return fmt.Errorf("invalid or missing $type field, expected %s", lexicon)
	}

	// Lexicon-specific validation
	switch lexicon {
	case TopicLexicon:
		return validateTopicRecord(record)
	case MessageLexicon:
		return validateMessageRecord(record)
	case ParticipationLexicon:
		return validateParticipationRecord(record)
	default:
		return fmt.Errorf("unknown lexicon: %s", lexicon)
	}
}

func validateTopicRecord(record map[string]interface{}) error {
	if _, ok := record["title"].(string); !ok {
		return fmt.Errorf("missing required field: title")
	}
	if _, ok := record["createdBy"].(string); !ok {
		return fmt.Errorf("missing required field: createdBy")
	}
	if _, ok := record["createdAt"].(string); !ok {
		return fmt.Errorf("missing required field: createdAt")
	}
	return nil
}

func validateMessageRecord(record map[string]interface{}) error {
	if _, ok := record["topic"].(string); !ok {
		return fmt.Errorf("missing required field: topic")
	}
	if _, ok := record["content"].(string); !ok {
		return fmt.Errorf("missing required field: content")
	}
	if _, ok := record["createdAt"].(string); !ok {
		return fmt.Errorf("missing required field: createdAt")
	}
	return nil
}

func validateParticipationRecord(record map[string]interface{}) error {
	if _, ok := record["topic"].(string); !ok {
		return fmt.Errorf("missing required field: topic")
	}
	if _, ok := record["participant"].(string); !ok {
		return fmt.Errorf("missing required field: participant")
	}
	if _, ok := record["joinedAt"].(string); !ok {
		return fmt.Errorf("missing required field: joinedAt")
	}
	return nil
}