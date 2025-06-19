package repository

import "errors"

// Common repository errors that can be tested for
var (
	ErrTopicNotFound       = errors.New("topic not found")
	ErrMessageNotFound     = errors.New("message not found")
	ErrUnauthorizedAccess  = errors.New("unauthorized access")
	ErrInvalidInput        = errors.New("invalid input")
	ErrTopicOwnershipRequired = errors.New("only topic creator can perform this action")
)