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
	"github.com/jrschumacher/dis.quest/internal/pds"
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
		Subject        string   `json:"subject"`
		InitialMessage string   `json:"initial_message"`
		Category       string   `json:"category,omitempty"`
		Tags           []string `json:"tags,omitempty"`
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

	// Create topic in user's PDS using ATProtocol lexicon
	pdsTopic, err := r.pdsService.CreateTopic(userCtx.DID, pds.CreateTopicParams{
		Title:   createReq.Subject,
		Summary: createReq.InitialMessage,
		Tags:    createReq.Tags,
	})
	if err != nil {
		httputil.WriteInternalError(w, err, "Failed to create topic in PDS", "did", userCtx.DID)
		return
	}

	// Create participation record in user's PDS
	_, err = r.pdsService.CreateParticipation(userCtx.DID, pds.CreateParticipationParams{
		Topic: pdsTopic.URI,
		Role:  "moderator", // Topic creator is moderator
	})
	if err != nil {
		logger.Error("Failed to create participation record", "error", err, "topic", pdsTopic.URI)
		// Don't fail the request - topic creation succeeded
	}

	// Store topic metadata in local database for discoverability
	// Extract rkey from PDS URI for local storage
	rkey := fmt.Sprintf("topic-%d", time.Now().UnixNano())
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
		logger.Error("Failed to store topic metadata locally", "error", err, "pds_uri", pdsTopic.URI)
		// Continue - PDS creation succeeded, local storage is for indexing
	}

	// Return the PDS topic data with local DB metadata
	response := map[string]any{
		"pds_uri":    pdsTopic.URI,
		"pds_cid":    pdsTopic.CID,
		"title":      pdsTopic.Title,
		"summary":    pdsTopic.Summary,
		"tags":       pdsTopic.Tags,
		"created_by": pdsTopic.CreatedBy,
		"created_at": pdsTopic.CreatedAt,
	}

	if result != nil {
		response["local_id"] = result.Topic.Did + ":" + result.Topic.Rkey
	}

	httputil.WriteCreated(w, response)
}

// PDSTopicsHandler returns all topics from the user's PDS
func (r *Router) PDSTopicsHandler(w http.ResponseWriter, req *http.Request) {
	// Get user context
	userCtx, ok := middleware.GetUserContext(req)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// For now, we'll return topics from local DB that correspond to PDS records
	// In a full implementation, this would query the user's PDS directly
	ctx := req.Context()
	topics, err := r.dbService.Queries().GetTopicsByDID(ctx, userCtx.DID)
	if err != nil {
		httputil.WriteInternalError(w, err, "Failed to fetch user's topics", "did", userCtx.DID)
		return
	}

	// Transform to include PDS information
	response := make([]map[string]any, len(topics))
	for i, topic := range topics {
		response[i] = map[string]any{
			"local_id":   topic.Did + ":" + topic.Rkey,
			"subject":    topic.Subject,
			"message":    topic.InitialMessage,
			"category":   topic.Category.String,
			"created_at": topic.CreatedAt,
			"updated_at": topic.UpdatedAt,
			"pds_uri":    fmt.Sprintf("at://%s/quest.dis.topic/%s", topic.Did, topic.Rkey),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode PDS topics", "error", err)
	}
}

// PDSRecordHandler retrieves a specific record from the user's PDS by URI
func (r *Router) PDSRecordHandler(w http.ResponseWriter, req *http.Request) {
	// Get user context
	_, ok := middleware.GetUserContext(req)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get URI from query parameter
	uri := req.URL.Query().Get("uri")
	if uri == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Missing 'uri' parameter")
		return
	}

	// For this demo, we'll try to retrieve using the PDS service
	// This would ideally query the actual PDS, but for now we'll simulate
	if record, err := r.pdsService.GetTopic(uri); err == nil {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(record); err != nil {
			logger.Error("Failed to encode PDS record", "error", err)
		}
		return
	}

	// If topic not found, try message
	if record, err := r.pdsService.GetMessage(uri); err == nil {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(record); err != nil {
			logger.Error("Failed to encode PDS record", "error", err)
		}
		return
	}

	httputil.WriteError(w, http.StatusNotFound, "Record not found in PDS")
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
	if err := sse.MergeFragments(html); err != nil {
		logger.Error("Failed to merge fragments", "error", err)
		return
	}

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
	if err := sse.MergeSignals(signalsJSON); err != nil {
		logger.Error("Failed to merge signals", "error", err)
		return
	}
}
