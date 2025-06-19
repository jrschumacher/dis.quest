// Package app provides the main application HTTP handlers
package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/jrschumacher/dis.quest/components"
	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/db"
	"github.com/jrschumacher/dis.quest/internal/httputil"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/middleware"
	"github.com/jrschumacher/dis.quest/internal/svrlib"
	"github.com/jrschumacher/dis.quest/internal/validation"
)

// Router handles application-specific HTTP routes
type Router struct {
	*svrlib.Router
	dbService *db.Service
}

// RegisterRoutes registers all application routes and returns a Router
func RegisterRoutes(mux *http.ServeMux, _ string, cfg *config.Config, dbService *db.Service) *Router {
	router := &Router{
		Router:    svrlib.NewRouter(mux, "/", cfg),
		dbService: dbService,
	}

	// Public routes
	mux.Handle("/", templ.Handler(components.Page(cfg.AppEnv)))
	mux.Handle("/login", templ.Handler(components.Login()))
	
	// Protected routes with clean middleware chains
	mux.Handle("/discussion", 
		middleware.WithProtectionFunc(router.DiscussionHandler))
	
	mux.Handle("/topics", 
		middleware.WithUserContextFunc(router.TopicsHandler))
	
	// API routes with custom middleware chains
	mux.Handle("/api/topics", 
		middleware.WithMiddleware(
			middleware.UserContextMiddleware,
		).ThenFunc(router.TopicsAPIHandler))
	
	mux.Handle("/api/topics/{id}/messages", 
		middleware.WithMiddleware(
			middleware.UserContextMiddleware,
		).ThenFunc(router.MessagesAPIHandler))

	return router
}

// DiscussionHandler shows the discussion page with real data
func (r *Router) DiscussionHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	
	// Get topics from database
	_, err := r.dbService.Queries().ListTopics(ctx, db.ListTopicsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		logger.Error("Failed to fetch topics", "error", err)
		http.Error(w, "Failed to load discussions", http.StatusInternalServerError)
		return
	}
	
	// Render discussion component with real data
	// TODO: Pass topics data to component once we update the component interface
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := components.Discussion()
	if err := component.Render(ctx, w); err != nil {
		logger.Error("Failed to render discussion page", "error", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// TopicsHandler shows the topics listing page
func (r *Router) TopicsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	
	// Get topics from database
	topics, err := r.dbService.Queries().ListTopics(ctx, db.ListTopicsParams{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		logger.Error("Failed to fetch topics", "error", err)
		http.Error(w, "Failed to load topics", http.StatusInternalServerError)
		return
	}
	
	// For now, return JSON (later we'll create a proper template)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(topics); err != nil {
		logger.Error("Failed to encode topics", "error", err)
	}
}

// TopicsAPIHandler handles REST API operations for topics
func (r *Router) TopicsAPIHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.listTopicsAPI(w, req)
	case http.MethodPost:
		r.createTopicAPI(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) listTopicsAPI(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	
	// Parse pagination parameters
	limitStr := req.URL.Query().Get("limit")
	offsetStr := req.URL.Query().Get("offset")
	
	limit := int64(20) // default
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	offset := int64(0) // default
	if offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 64); err == nil && o >= 0 {
			offset = o
		}
	}
	
	topics, err := r.dbService.Queries().ListTopics(ctx, db.ListTopicsParams{
		Limit:  func() int32 {
			if limit < 0 || limit > 2147483647 {
				return 2147483647
			}
			return int32(limit) // #nosec G115
		}(),
		Offset: func() int32 {
			if offset < 0 || offset > 2147483647 {
				return 0
			}
			return int32(offset) // #nosec G115
		}(),
	})
	if err != nil {
		logger.Error("Failed to fetch topics", "error", err)
		http.Error(w, "Failed to fetch topics", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(topics); err != nil {
		logger.Error("Failed to encode topics", "error", err)
	}
}

func (r *Router) createTopicAPI(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	
	// Get user context
	userCtx, ok := middleware.GetUserContext(req)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	// Parse request body
	var createReq struct {
		Subject        string `json:"subject"`
		InitialMessage string `json:"initial_message"`
		Category       string `json:"category,omitempty"`
	}
	
	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid JSON in request body")
		return
	}
	
	// Validate input
	validator := validation.TopicValidation{
		Subject:        createReq.Subject,
		InitialMessage: createReq.InitialMessage,
		Category:       createReq.Category,
	}
	
	if err := validator.Validate(); err != nil {
		if validationErrors, ok := err.(validation.Errors); ok {
			httputil.WriteValidationError(w, validationErrors)
		} else {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	
	// Generate a simple rkey (timestamp-based for now)
	rkey := fmt.Sprintf("topic-%d", time.Now().UnixNano())
	
	// Create topic with automatic participation using transaction
	now := time.Now()
	result, err := r.dbService.CreateTopicWithParticipation(ctx, db.CreateTopicWithParticipationParams{
		Did:            userCtx.DID,
		Rkey:           rkey,
		Subject:        createReq.Subject,
		InitialMessage: createReq.InitialMessage,
		Category:       sql.NullString{String: createReq.Category, Valid: createReq.Category != ""},
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		httputil.WriteInternalError(w, err, "Failed to create topic", "did", userCtx.DID)
		return
	}
	
	httputil.WriteCreated(w, result.Topic)
}

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
