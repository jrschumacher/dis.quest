// Package app provides the main application HTTP handlers
package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jrschumacher/dis.quest/components"
	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/db"
	"github.com/jrschumacher/dis.quest/internal/httputil"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/middleware"
	"github.com/jrschumacher/dis.quest/internal/svrlib"
	"github.com/jrschumacher/dis.quest/internal/validation"
	datastar "github.com/starfederation/datastar/sdk/go"
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

	renderMiddleware := middleware.PageWrapper(cfg.AppEnv)

	// Public routes
	mux.Handle("/",
		middleware.ApplyFunc(router.LandingHandler,
			renderMiddleware,
		),
	)
	mux.Handle("/login",
		middleware.ApplyFunc(router.LoginHandler,
			renderMiddleware,
		),
	)

	// Protected routes with explicit middleware options
	mux.Handle("/discussion",
		middleware.ApplyFunc(router.DiscussionHandler,
			middleware.AuthMiddleware,
			middleware.UserContextMiddleware,
			middleware.RequireUserContext,
			renderMiddleware,
		),
	)

	mux.Handle("/topics",
		middleware.ApplyFunc(router.TopicsHandler,
			middleware.UserContextMiddleware,
			renderMiddleware,
		),
	)

	// API routes with custom middleware chains
	mux.Handle("/api/topics",
		middleware.ApplyFunc(router.TopicsAPIHandler,
			middleware.UserContextMiddleware,
		),
	)

	mux.Handle("/api/topics/{id}/messages",
		middleware.ApplyFunc(router.MessagesAPIHandler,
			middleware.UserContextMiddleware,
		),
	)

	// Datastar API endpoints
	mux.Handle("/api/messages/{id}/like",
		middleware.ApplyFunc(router.LikeMessageHandler,
			middleware.UserContextMiddleware,
		),
	)

	mux.Handle("/api/messages/reply",
		middleware.ApplyFunc(router.ReplyMessageHandler,
			middleware.UserContextMiddleware,
		),
	)

	// SSE endpoint for datastar real-time updates
	mux.HandleFunc("/stream/topics", router.TopicsStreamHandler)

	return router
}

// LandingHandler renders the landing page
func (r *Router) LandingHandler(w http.ResponseWriter, _ *http.Request) {
	// You can replace this with a custom landing component if desired
	w.WriteHeader(http.StatusOK)
}

// LoginHandler renders the login page
func (r *Router) LoginHandler(w http.ResponseWriter, req *http.Request) {
	_ = components.Login().Render(req.Context(), w)
}

// DiscussionHandler shows the discussion page with mock data
func (r *Router) DiscussionHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// Create mock data matching the new component structure
	messages := []components.MessageSignals{
		{
			Id:      "msg-1",
			Author:  "@bob",
			Date:    "2025-05-26",
			Content: "I agree with the topic and would like to add...",
			Liked:   false,
			ThreadedReplies: []components.MessageSignals{
				{
					Id:      "msg-1-reply-1",
					Author:  "@eve",
					Date:    "2025-05-26",
					Content: "Replying to @bob: Good point!",
					Liked:   false,
				},
			},
		},
		{
			Id:      "msg-2",
			Author:  "@carol",
			Date:    "2025-05-26",
			Content: "Here's another perspective on this topic.",
			Liked:   false,
			ThreadedReplies: []components.MessageSignals{
				{
					Id:      "msg-2-reply-1",
					Author:  "@frank",
					Date:    "2025-05-26",
					Content: "Replying to @carol: I disagree.",
					Liked:   false,
				},
			},
		},
		{
			Id:      "msg-3",
			Author:  "@dave",
			Date:    "2025-05-26",
			Content: "What about edge cases?",
			Liked:   false,
		},
	}

	// Build initial signal state for datastar
	initialSignals := make(map[string]any)
	for _, msg := range messages {
		initialSignals["liked_"+msg.Id] = msg.Liked
		for _, reply := range msg.ThreadedReplies {
			initialSignals["liked_"+reply.Id] = reply.Liked
		}
	}

	signals := components.DiscussionSignals{
		TopicId:        "topic-1",
		Messages:       messages,
		InitialSignals: initialSignals,
	}

	// Render discussion component with mock data
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := components.Discussion(signals)
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
		Limit: func() int32 {
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

// TopicsStreamHandler streams the latest topics as HTML fragments for datastar
func (r *Router) TopicsStreamHandler(w http.ResponseWriter, req *http.Request) {
	sse := datastar.NewSSE(w, req)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	ctx := req.Context()

	// Send initial data
	r.sendTopicsUpdate(ctx, sse)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.sendTopicsUpdate(ctx, sse)
		}
	}
}

// sendTopicsUpdate sends the latest topics via SSE
func (r *Router) sendTopicsUpdate(ctx context.Context, sse *datastar.ServerSentEventGenerator) {
	topics, err := r.dbService.Queries().ListTopics(ctx, db.ListTopicsParams{
		Limit:  5,
		Offset: 0,
	})
	if err != nil {
		logger.Error("Failed to fetch topics for SSE", "error", err)
		return
	}

	// Build HTML fragment for topics
	html := `<div id="live-topics">`
	html += `<h3>🔴 Live Topics (updates every 3s)</h3>`
	if len(topics) == 0 {
		html += `<p><em>No topics yet. Create one to see it appear here!</em></p>`
	} else {
		html += `<ul>`
		for _, topic := range topics {
			html += fmt.Sprintf(`<li><strong>%s</strong> - %s</li>`,
				topic.Subject,
				topic.CreatedAt.Format("15:04:05"))
		}
		html += `</ul>`
	}
	html += fmt.Sprintf(`<small>Last updated: %s | Total: %d topics</small>`,
		time.Now().Format("15:04:05"), len(topics))
	html += `</div>`

	// Send fragment merge
	sse.MergeFragments(html)

	// Send signal update
	signals := map[string]any{
		"last_update":   time.Now().Format("15:04:05"),
		"topic_count":   len(topics),
		"stream_active": true,
	}
	signalsJSON, err := json.Marshal(signals)
	if err != nil {
		logger.Error("Failed to marshal signals", "error", err)
		return
	}
	sse.MergeSignals(signalsJSON)
}
