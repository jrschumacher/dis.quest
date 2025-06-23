package app

import (
	"encoding/json"
	"net/http"

	"github.com/jrschumacher/dis.quest/components"
	"github.com/jrschumacher/dis.quest/internal/db"
	"github.com/jrschumacher/dis.quest/internal/logger"
)

// LandingHandler renders the landing page
func (r *Router) LandingHandler(w http.ResponseWriter, _ *http.Request) {
	// You can replace this with a custom landing component if desired
	w.WriteHeader(http.StatusOK)
}

// LoginHandler renders the login page
func (r *Router) LoginHandler(w http.ResponseWriter, req *http.Request) {
	redirectURL := req.URL.Query().Get("redirect")
	if redirectURL != "" {
		_ = components.LoginWithRedirect(redirectURL).Render(req.Context(), w)
	} else {
		_ = components.Login().Render(req.Context(), w)
	}
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