// Package validation provides structured validation error handling
package validation

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Error represents a validation error with field-specific details
type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Errors represents multiple validation errors
type Errors []Error

// Error implements the error interface
func (ve Errors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}

	var messages []string
	for _, err := range ve {
		if err.Field != "" {
			messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
		} else {
			messages = append(messages, err.Message)
		}
	}

	return strings.Join(messages, "; ")
}

// Add adds a validation error
func (ve *Errors) Add(field, message string) {
	*ve = append(*ve, Error{Field: field, Message: message})
}

// HasErrors returns true if there are validation errors
func (ve Errors) HasErrors() bool {
	return len(ve) > 0
}

// ValidateRequired checks if a value is not empty
func ValidateRequired(value string, fieldName string) *Error {
	if strings.TrimSpace(value) == "" {
		return &Error{
			Field:   fieldName,
			Message: "is required",
		}
	}
	return nil
}

// ValidateMaxLength checks if a string doesn't exceed the maximum length
func ValidateMaxLength(value string, maxLength int, fieldName string) *Error {
	if utf8.RuneCountInString(value) > maxLength {
		return &Error{
			Field:   fieldName,
			Message: fmt.Sprintf("must not exceed %d characters", maxLength),
		}
	}
	return nil
}

// ValidateMinLength checks if a string meets the minimum length
func ValidateMinLength(value string, minLength int, fieldName string) *Error {
	if utf8.RuneCountInString(strings.TrimSpace(value)) < minLength {
		return &Error{
			Field:   fieldName,
			Message: fmt.Sprintf("must be at least %d characters", minLength),
		}
	}
	return nil
}

// ValidateDID checks if a string looks like a valid DID
func ValidateDID(value string, fieldName string) *Error {
	if !strings.HasPrefix(value, "did:") {
		return &Error{
			Field:   fieldName,
			Message: "must be a valid DID",
		}
	}

	parts := strings.Split(value, ":")
	if len(parts) < 3 {
		return &Error{
			Field:   fieldName,
			Message: "must be a valid DID format",
		}
	}

	return nil
}

// ValidateRkey checks if a string is a valid record key
func ValidateRkey(value string, fieldName string) *Error {
	if strings.TrimSpace(value) == "" {
		return &Error{
			Field:   fieldName,
			Message: "is required",
		}
	}

	// Basic rkey validation - no spaces, reasonable length
	if strings.Contains(value, " ") {
		return &Error{
			Field:   fieldName,
			Message: "cannot contain spaces",
		}
	}

	if len(value) > 100 {
		return &Error{
			Field:   fieldName,
			Message: "must not exceed 100 characters",
		}
	}

	return nil
}

// TopicValidation validates topic creation parameters
type TopicValidation struct {
	Subject        string
	InitialMessage string
	Category       string
}

// Validate validates topic fields
func (tv *TopicValidation) Validate() error {
	var errors Errors

	// Validate subject
	if err := ValidateRequired(tv.Subject, "subject"); err != nil {
		errors.Add(err.Field, err.Message)
	} else {
		if err := ValidateMinLength(tv.Subject, 3, "subject"); err != nil {
			errors.Add(err.Field, err.Message)
		}
		if err := ValidateMaxLength(tv.Subject, 200, "subject"); err != nil {
			errors.Add(err.Field, err.Message)
		}
	}

	// Validate initial message
	if err := ValidateRequired(tv.InitialMessage, "initial_message"); err != nil {
		errors.Add(err.Field, err.Message)
	} else {
		if err := ValidateMinLength(tv.InitialMessage, 10, "initial_message"); err != nil {
			errors.Add(err.Field, err.Message)
		}
		if err := ValidateMaxLength(tv.InitialMessage, 5000, "initial_message"); err != nil {
			errors.Add(err.Field, err.Message)
		}
	}

	// Validate category (optional)
	if tv.Category != "" {
		if err := ValidateMaxLength(tv.Category, 50, "category"); err != nil {
			errors.Add(err.Field, err.Message)
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// MessageValidation validates message creation parameters
type MessageValidation struct {
	Content           string
	ParentMessageRkey string
}

// Validate validates message fields
func (mv *MessageValidation) Validate() error {
	var errors Errors

	// Validate content
	if err := ValidateRequired(mv.Content, "content"); err != nil {
		errors.Add(err.Field, err.Message)
	} else {
		if err := ValidateMinLength(mv.Content, 1, "content"); err != nil {
			errors.Add(err.Field, err.Message)
		}
		if err := ValidateMaxLength(mv.Content, 2000, "content"); err != nil {
			errors.Add(err.Field, err.Message)
		}
	}

	// Validate parent message rkey (optional)
	if mv.ParentMessageRkey != "" {
		if err := ValidateRkey(mv.ParentMessageRkey, "parent_message_rkey"); err != nil {
			errors.Add(err.Field, err.Message)
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}
