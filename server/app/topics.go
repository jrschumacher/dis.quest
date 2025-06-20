package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jrschumacher/dis.quest/internal/db"
	"github.com/jrschumacher/dis.quest/internal/httputil"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/middleware"
	"github.com/jrschumacher/dis.quest/internal/validation"
	datastar "github.com/starfederation/datastar/sdk/go"
)

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