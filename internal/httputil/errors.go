package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/validation"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string                      `json:"error"`
	Message string                      `json:"message,omitempty"`
	Details []validation.ValidationError `json:"details,omitempty"`
}

// WriteError writes a standardized error response
func WriteError(w http.ResponseWriter, status int, message string, logFields ...any) {
	response := ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode error response", "error", err)
	}
	
	// Log the error with additional context
	logFields = append([]any{"status", status, "message", message}, logFields...)
	logger.Error("HTTP error response", logFields...)
}

// WriteValidationError writes a validation error response
func WriteValidationError(w http.ResponseWriter, validationErr validation.ValidationErrors) {
	response := ErrorResponse{
		Error:   "Validation Failed",
		Message: validationErr.Error(),
		Details: validationErr,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode validation error response", "error", err)
	}
	
	logger.Warn("Validation error", "errors", validationErr.Error())
}

// WriteInternalError writes a generic internal server error
func WriteInternalError(w http.ResponseWriter, err error, message string, logFields ...any) {
	response := ErrorResponse{
		Error:   "Internal Server Error",
		Message: message,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		logger.Error("Failed to encode internal error response", "error", encodeErr)
	}
	
	// Log the actual error with context
	logFields = append([]any{"error", err, "message", message}, logFields...)
	logger.Error("Internal server error", logFields...)
}

// WriteJSON writes a JSON response with proper error handling
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", "error", err)
		// Can't write another response at this point, but log it
	}
}

// WriteCreated writes a 201 Created response with JSON data
func WriteCreated(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, data)
}

// WriteSuccess writes a 200 OK response with JSON data
func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, data)
}