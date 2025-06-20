package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jrschumacher/dis.quest/internal/db"
	"github.com/jrschumacher/dis.quest/internal/httputil"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/middleware"
	"github.com/jrschumacher/dis.quest/internal/validation"
)

// MessagesAPIHandler handles REST API operations for messages within a topic
func (r *Router) MessagesAPIHandler(w http.ResponseWriter, req *http.Request) {
	// Extract topic ID from URL path
	// Note: In Go 1.22+, we can use path parameters directly
	topicID := req.URL.Path[len("/api/topics/"):]
	if idx := len(topicID) - len("/messages"); idx > 0 && topicID[idx:] == "/messages" {
		topicID = topicID[:idx]
	}

	switch req.Method {
	case http.MethodGet:
		r.listMessagesAPI(w, req, topicID)
	case http.MethodPost:
		r.createMessageAPI(w, req, topicID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) listMessagesAPI(w http.ResponseWriter, req *http.Request, topicID string) {
	ctx := req.Context()

	// For now, assume topicID format is "did:rkey"
	// TODO: Implement proper topic ID parsing
	parts := []string{topicID, topicID} // placeholder
	if len(parts) != 2 {
		http.Error(w, "Invalid topic ID format", http.StatusBadRequest)
		return
	}

	messages, err := r.dbService.Queries().GetMessagesByTopic(ctx, db.GetMessagesByTopicParams{
		TopicDid:  parts[0],
		TopicRkey: parts[1],
	})
	if err != nil {
		logger.Error("Failed to fetch messages", "error", err, "topicID", topicID)
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		logger.Error("Failed to encode messages", "error", err)
	}
}

func (r *Router) createMessageAPI(w http.ResponseWriter, req *http.Request, topicID string) {
	ctx := req.Context()

	// Get user context
	userCtx, ok := middleware.GetUserContext(req)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse request body
	var createReq struct {
		Content           string `json:"content"`
		ParentMessageRkey string `json:"parent_message_rkey,omitempty"`
	}

	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid JSON in request body")
		return
	}

	// Validate input
	validator := validation.MessageValidation{
		Content:           createReq.Content,
		ParentMessageRkey: createReq.ParentMessageRkey,
	}

	if err := validator.Validate(); err != nil {
		if validationErrors, ok := err.(validation.Errors); ok {
			httputil.WriteValidationError(w, validationErrors)
		} else {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	// For now, assume topicID format is "did:rkey"
	// TODO: Implement proper topic ID parsing
	parts := []string{topicID, topicID} // placeholder
	if len(parts) != 2 {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid topic ID format")
		return
	}

	// Generate a simple rkey for the message
	rkey := fmt.Sprintf("msg-%d", time.Now().UnixNano())

	// Create message
	now := time.Now()
	message, err := r.dbService.Queries().CreateMessage(ctx, db.CreateMessageParams{
		Did:               userCtx.DID,
		Rkey:              rkey,
		TopicDid:          parts[0],
		TopicRkey:         parts[1],
		ParentMessageRkey: sql.NullString{String: createReq.ParentMessageRkey, Valid: createReq.ParentMessageRkey != ""},
		Content:           createReq.Content,
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		httputil.WriteInternalError(w, err, "Failed to create message", "did", userCtx.DID, "topicID", topicID)
		return
	}

	httputil.WriteCreated(w, message)
}

// LikeMessageHandler handles liking/unliking messages
func (r *Router) LikeMessageHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract message ID from URL path
	messageID := req.PathValue("id")
	if messageID == "" {
		http.Error(w, "Message ID required", http.StatusBadRequest)
		return
	}

	logger.Info("Like toggled", "messageID", messageID)

	// For datastar, we just need to return success - the client handles the state
	w.WriteHeader(http.StatusOK)
}

// ReplyMessageHandler handles posting replies to messages
func (r *Router) ReplyMessageHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user context
	userCtx, ok := middleware.GetUserContext(req)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse form data from Datastar
	if err := req.ParseForm(); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	content := req.FormValue("reply_content")
	if content == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Reply content is required")
		return
	}

	// For demonstration, just return success
	// In real implementation, you'd save to database
	logger.Info("Reply posted", "user", userCtx.DID, "content", content)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]any{
		"success":    true,
		"message":    "Reply posted successfully",
		"show_reply": false, // Hide the reply form after posting
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode reply response", "error", err)
	}
}