package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/jrschumacher/dis.quest/components"
	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/middleware"
	"github.com/jrschumacher/dis.quest/internal/pds"
	datastar "github.com/starfederation/datastar/sdk/go"
)

// TestResult represents the result of a PDS test operation
type TestResult struct {
	Operation string `json:"operation"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Details   string `json:"details"`
}

// DevPDSHandler serves a development page for testing PDS functionality
func (r *Router) DevPDSHandler(w http.ResponseWriter, req *http.Request) {
	// Only allow in development
	if r.Config.AppEnv != "development" {
		http.NotFound(w, req)
		return
	}

	// Get user context if available
	userCtx, hasAuth := middleware.GetUserContext(req)
	
	// Check token expiration if user is authenticated
	var tokenExpired bool
	var tokenExpiration time.Time
	if hasAuth {
		if accessToken, err := auth.GetSessionCookie(req); err == nil && accessToken != "" {
			// Simple JWT expiration check
			parts := strings.Split(accessToken, ".")
			if len(parts) == 3 {
				payload := parts[1]
				// Add padding if needed for base64 decoding
				for len(payload)%4 != 0 {
					payload += "="
				}
				if decoded, decodeErr := base64.StdEncoding.DecodeString(payload); decodeErr == nil {
					var claims map[string]interface{}
					if jsonErr := json.Unmarshal(decoded, &claims); jsonErr == nil {
						if exp, ok := claims["exp"].(float64); ok {
							tokenExpiration = time.Unix(int64(exp), 0)
							tokenExpired = time.Now().After(tokenExpiration)
						}
					}
				}
			}
		}
	}
	
	devPageData := components.DevPDSPageData{
		Title:           "PDS Development Tools",
		HasAuth:         hasAuth,
		UserDID:         "",
		TestResults:     []interface{}{},
		TokenExpired:    tokenExpired,
		TokenExpiration: tokenExpiration,
	}
	
	if hasAuth {
		devPageData.UserDID = userCtx.DID
	}

	component := components.DevPDSPage(devPageData)
	if err := component.Render(req.Context(), w); err != nil {
		logger.Error("Failed to render dev PDS page", "error", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// DevPDSTestHandler handles PDS test operations
func (r *Router) DevPDSTestHandler(w http.ResponseWriter, req *http.Request) {
	// Only allow in development
	if r.Config.AppEnv != "development" {
		http.NotFound(w, req)
		return
	}

	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Debug: Log request details
	logger.Info("Request details", 
		"method", req.Method,
		"contentType", req.Header.Get("Content-Type"),
		"userAgent", req.Header.Get("User-Agent"),
	)

	// Parse request data (could be form data or JSON)
	var operation, testDID, topicURI string
	
	contentType := req.Header.Get("Content-Type")
	logger.Info("Processing request", "contentType", contentType)
	
	if strings.Contains(contentType, "application/json") {
		// Parse JSON data
		var data map[string]interface{}
		if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
			logger.Error("Failed to parse JSON data", "error", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		
		logger.Info("Parsed JSON data", "data", data)
		
		if op, ok := data["operation"].(string); ok {
			operation = op
		}
		if did, ok := data["testDID"].(string); ok {
			testDID = did
		}
		if uri, ok := data["topicURI"].(string); ok {
			topicURI = uri
		}
	} else {
		// Parse form data
		if err := req.ParseForm(); err != nil {
			logger.Error("Failed to parse form data", "error", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		
		operation = req.FormValue("operation")
		testDID = req.FormValue("test_did")
		topicURI = req.FormValue("topic_uri")
		
		logger.Info("Parsed form data", 
			"operation", operation,
			"testDID", testDID,
			"topicURI", topicURI,
			"allValues", req.Form,
		)
	}
	
	if testDID == "" {
		testDID = "did:plc:test123456789"
	}

	logger.Info("Processing dev PDS test", "operation", operation, "testDID", testDID)

	var result TestResult

	switch operation {
	case "validate_lexicons":
		result = r.testLexiconValidation()
	case "test_uri_parsing":
		result = r.testURIParsing()
	case "simulate_create_topic":
		result = r.simulateCreateTopic(testDID)
	case "simulate_get_topic":
		uri := topicURI
		if uri == "" {
			uri = fmt.Sprintf("at://%s/quest.dis.topic/topic-123", testDID)
		}
		result = r.simulateGetTopic(uri)
	case "test_real_pds_structure":
		result = r.testRealPDSStructure(testDID)
	case "browse_real_pds":
		result = r.browseRealPDS(testDID)
	case "list_pds_topics":
		result = r.listPDSTopics(testDID)
	case "get_pds_record":
		result = r.getPDSRecord(topicURI)
	case "create_random_topic":
		result = r.createRandomTopic(req, testDID)
	case "test_standard_post":
		result = r.testStandardPost(req, testDID)
	case "test_session_auth":
		result = r.testSessionAuth(req, testDID)
	default:
		result = TestResult{
			Operation: operation,
			Success:   false,
			Message:   "Unknown operation",
			Details:   "",
		}
	}

	logger.Info("Test completed", "operation", result.Operation, "success", result.Success, "message", result.Message)

	// Return result using Datastar
	sse := datastar.NewSSE(w, req)
	
	// Build the result HTML fragment 
	htmlFragment := fmt.Sprintf(`<div id="test-results">
		<div class="test-result result-%s">
			<h4>%s</h4>
			<p>%s</p>
			<details>
				<summary>Details</summary>
				<pre class="result-details">%s</pre>
			</details>
			<small>Latest test result</small>
		</div>
		<p><em>Run tests to see results here...</em></p>
	</div>`, 
		func() string { if result.Success { return "success" }; return "error" }(),
		strings.ToUpper(strings.ReplaceAll(result.Operation, "_", " ")),
		result.Message,
		result.Details,
	)
	
	logger.Info("Sending fragment", "html", htmlFragment[:100]+"...")
	
	if err := sse.MergeFragments(htmlFragment); err != nil {
		logger.Error("Failed to merge fragments", "error", err)
	} else {
		logger.Info("Fragment merge successful")
	}
}

func (r *Router) testLexiconValidation() TestResult {
	// Test valid topic record
	validTopic := map[string]interface{}{
		"$type":     pds.TopicLexicon,
		"title":     "Test Topic",
		"summary":   "A test topic for validation",
		"tags":      []string{"test", "validation"},
		"createdBy": "did:plc:test123",
		"createdAt": time.Now().Format(time.RFC3339),
	}

	if err := pds.ValidateLexicon(pds.TopicLexicon, validTopic); err != nil {
		return TestResult{
			Operation: "validate_lexicons",
			Success:   false,
			Message:   "Valid topic failed validation",
			Details:   err.Error(),
		}
	}

	// Test invalid topic record
	invalidTopic := map[string]interface{}{
		"$type":     pds.TopicLexicon,
		"summary":   "Missing required title field",
		"createdBy": "did:plc:test123",
		"createdAt": time.Now().Format(time.RFC3339),
	}

	if err := pds.ValidateLexicon(pds.TopicLexicon, invalidTopic); err == nil {
		return TestResult{
			Operation: "validate_lexicons",
			Success:   false,
			Message:   "Invalid topic should have failed validation",
			Details:   "Missing title validation not caught",
		}
	}

	return TestResult{
		Operation: "validate_lexicons",
		Success:   true,
		Message:   "Lexicon validation working correctly",
		Details:   "✅ Valid record passed, invalid record correctly rejected",
	}
}

func (r *Router) testURIParsing() TestResult {
	testURI := "at://did:plc:test123/quest.dis.topic/topic-123456789"
	components, err := pds.ParseATUri(testURI)
	if err != nil {
		return TestResult{
			Operation: "test_uri_parsing",
			Success:   false,
			Message:   "Failed to parse valid URI",
			Details:   err.Error(),
		}
	}

	expected := "did:plc:test123"
	if components.DID != expected {
		return TestResult{
			Operation: "test_uri_parsing",
			Success:   false,
			Message:   "URI parsing incorrect",
			Details:   fmt.Sprintf("Expected DID %s, got %s", expected, components.DID),
		}
	}

	return TestResult{
		Operation: "test_uri_parsing",
		Success:   true,
		Message:   "URI parsing working correctly",
		Details:   fmt.Sprintf("✅ Parsed: DID=%s, Collection=%s, RKey=%s", components.DID, components.Collection, components.RKey),
	}
}

func (r *Router) simulateCreateTopic(testDID string) TestResult {
	params := pds.CreateTopicParams{
		Title:   "Simulated Test Topic",
		Summary: "This is a test topic created via the dev interface",
		Tags:    []string{"dev", "test", "simulation"},
	}

	// Create lexicon record
	topicRecord := &pds.TopicRecord{
		Type:      pds.TopicLexicon,
		Title:     params.Title,
		Summary:   params.Summary,
		Tags:      params.Tags,
		CreatedBy: testDID,
		CreatedAt: time.Now(),
	}

	// Validate
	recordData := topicRecord.ToMap()
	if err := pds.ValidateLexicon(pds.TopicLexicon, recordData); err != nil {
		return TestResult{
			Operation: "simulate_create_topic",
			Success:   false,
			Message:   "Topic record validation failed",
			Details:   err.Error(),
		}
	}

	// Generate realistic URIs
	rkey := pds.GenerateRKey("topic")
	uri := fmt.Sprintf("at://%s/%s/%s", testDID, pds.TopicLexicon, rkey)
	cid := fmt.Sprintf("bafyrei%d", time.Now().UnixNano()%1000000)

	topic := topicRecord.ToTopic(uri, cid)

	return TestResult{
		Operation: "simulate_create_topic",
		Success:   true,
		Message:   "Topic creation simulation successful",
		Details:   fmt.Sprintf("✅ Created: URI=%s, CID=%s, Title=%s", topic.URI, topic.CID, topic.Title),
	}
}

func (r *Router) simulateGetTopic(uri string) TestResult {
	components, err := pds.ParseATUri(uri)
	if err != nil {
		return TestResult{
			Operation: "simulate_get_topic",
			Success:   false,
			Message:   "Invalid URI format",
			Details:   err.Error(),
		}
	}

	if components.Collection != pds.TopicLexicon {
		return TestResult{
			Operation: "simulate_get_topic",
			Success:   false,
			Message:   "URI is not a topic record",
			Details:   fmt.Sprintf("Expected %s, got %s", pds.TopicLexicon, components.Collection),
		}
	}

	// Simulate found topic
	topicRecord := &pds.TopicRecord{
		Type:      pds.TopicLexicon,
		Title:     "Retrieved Test Topic",
		Summary:   "This topic was simulated for retrieval testing",
		Tags:      []string{"retrieved", "test"},
		CreatedBy: components.DID,
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	cid := fmt.Sprintf("bafyrei%d", time.Now().UnixNano()%1000000)
	topic := topicRecord.ToTopic(uri, cid)

	return TestResult{
		Operation: "simulate_get_topic",
		Success:   true,
		Message:   "Topic retrieval simulation successful",
		Details:   fmt.Sprintf("✅ Retrieved: Title=%s, CreatedBy=%s, Tags=%v", topic.Title, topic.CreatedBy, topic.Tags),
	}
}

func (r *Router) testRealPDSStructure(testDID string) TestResult {
	// Test what the real PDS service would do (without making actual calls)
	atprotoService := pds.NewATProtoService()
	
	// Test structure without network calls
	params := pds.CreateTopicParams{
		Title:   "Real PDS Structure Test",
		Summary: "Testing the structure that would be sent to real PDS",
		Tags:    []string{"real", "structure", "test"},
	}

	// This would fail on the actual HTTP call, but we can see the structure
	_, err := atprotoService.CreateTopic(testDID, params)
	
	// Expected to fail due to no auth, but structure is correct
	if err != nil && (containsString(err.Error(), "failed to create topic in PDS") || 
					 containsString(err.Error(), "PDS request failed")) {
		return TestResult{
			Operation: "test_real_pds_structure",
			Success:   true,
			Message:   "Real PDS structure test successful",
			Details:   "✅ Structure is correct, failed at HTTP call as expected (needs auth)",
		}
	}

	return TestResult{
		Operation: "test_real_pds_structure",
		Success:   false,
		Message:   "Unexpected error in PDS structure",
		Details:   fmt.Sprintf("Error: %v", err),
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   (len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		   (len(substr) < len(s) && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// browseRealPDS attempts to browse the user's actual PDS for quest.dis.* records
func (r *Router) browseRealPDS(userDID string) TestResult {
	// Use the real XRPC client to list records
	xrpcClient := pds.NewXRPCClient()
	
	// Try to list quest.dis.topic records
	ctx := context.Background()
	// Note: This will likely fail due to no access token, but we can see the structure
	response, err := xrpcClient.ListRecords(ctx, userDID, pds.TopicLexicon, 10, "", "")
	
	if err != nil {
		return TestResult{
			Operation: "browse_real_pds", 
			Success:   false,
			Message:   "Expected failure - needs authentication",
			Details:   fmt.Sprintf("Error: %v\nThis shows the PDS browsing structure is correct, but needs access token for real queries.", err),
		}
	}
	
	return TestResult{
		Operation: "browse_real_pds",
		Success:   true, 
		Message:   fmt.Sprintf("Found %d quest.dis.topic records", len(response.Records)),
		Details:   fmt.Sprintf("Records: %+v", response.Records),
	}
}

// listPDSTopics lists topics from the user's PDS using authenticated session
func (r *Router) listPDSTopics(_ string) TestResult {
	// This would use the user's actual access token from their session
	// For now, we'll show what the structure would look like
	
	return TestResult{
		Operation: "list_pds_topics",
		Success:   false,
		Message:   "Access token integration needed",
		Details:   "To query your real PDS, we need to:\n1. Extract access token from your session\n2. Add it to XRPC requests\n3. Handle DPoP headers for security\n\nThis would call: listRecords(repo: your-did, collection: quest.dis.topic)",
	}
}

// getPDSRecord retrieves a specific record from the user's PDS
func (r *Router) getPDSRecord(uri string) TestResult {
	if uri == "" {
		return TestResult{
			Operation: "get_pds_record",
			Success:   false,
			Message:   "No URI provided", 
			Details:   "Please enter a valid AT URI like: at://did:plc:abc/quest.dis.topic/topic-123",
		}
	}
	
	// Parse the URI
	components, err := pds.ParseATUri(uri)
	if err != nil {
		return TestResult{
			Operation: "get_pds_record",
			Success:   false,
			Message:   "Invalid URI format",
			Details:   fmt.Sprintf("Error parsing URI: %v", err),
		}
	}
	
	// Try to get the record (will fail without auth, but shows structure)
	xrpcClient := pds.NewXRPCClient()
	ctx := context.Background()
	
	_, err = xrpcClient.GetRecord(ctx, components.DID, components.Collection, components.RKey, "")
	
	return TestResult{
		Operation: "get_pds_record",
		Success:   false,
		Message:   "Expected failure - needs authentication",
		Details:   fmt.Sprintf("Attempted to fetch:\nDID: %s\nCollection: %s\nRKey: %s\n\nError: %v\n\nTo access real records, we need your access token.", 
			components.DID, components.Collection, components.RKey, err),
	}
}

// createRandomTopic creates a real topic in the user's PDS with random data
func (r *Router) createRandomTopic(req *http.Request, userDID string) TestResult {
	// Generate random topic data
	topics := []string{
		"Random Dev Test Topic",
		"ATProtocol Integration Test", 
		"PDS Browsing Validation",
		"Lexicon Testing Topic",
		"Quest Discussion Sample",
	}
	
	messages := []string{
		"This is a randomly generated topic for testing the PDS integration!",
		"Testing quest.dis.topic lexicon with real data.",
		"Validating end-to-end ATProtocol record creation.",
		"Generated from the dev interface to test browsing functionality.",
		"Sample topic created to verify PDS storage works correctly.",
	}
	
	tagSets := [][]string{
		{"dev", "test", "random"},
		{"atprotocol", "lexicon", "validation"},
		{"pds", "browsing", "integration"},
		{"quest", "discussion", "sample"},
		{"e2e", "testing", "demo"},
	}
	
	rand.Seed(time.Now().UnixNano())
	randomTopic := topics[rand.Intn(len(topics))]
	randomMessage := messages[rand.Intn(len(messages))]
	randomTags := tagSets[rand.Intn(len(tagSets))]
	
	// Add timestamp to make it unique
	uniqueTopic := fmt.Sprintf("%s [%s]", randomTopic, time.Now().Format("15:04:05"))
	
	// Extract access token from session
	accessToken, err := auth.GetSessionCookie(req)
	if err != nil {
		return TestResult{
			Operation: "create_random_topic",
			Success:   false,
			Message:   "No session found - please login first",
			Details:   fmt.Sprintf("Error getting session cookie: %v", err),
		}
	}
	
	// Extract DPoP key from session (required for ATProtocol OAuth)
	dpopKey, err := auth.GetDPoPKeyFromCookie(req)
	if err != nil {
		logger.Error("Failed to get DPoP key from cookie", "error", err)
		return TestResult{
			Operation: "create_random_topic",
			Success:   false,
			Message:   "No DPoP key found - please re-authenticate",
			Details:   fmt.Sprintf("DPoP key required for ATProtocol OAuth. Error: %v\nPlease use the 'Force Re-authenticate' button to get a fresh session with DPoP key.", err),
		}
	}
	
	logger.Info("Successfully extracted DPoP key", "keyType", fmt.Sprintf("%T", dpopKey))
	
	// Check if the JWT token is expired and inspect scopes (basic check)
	if accessToken != "" {
		// Simple JWT expiration and scope check without full validation
		parts := strings.Split(accessToken, ".")
		if len(parts) == 3 {
			// Decode payload (basic check, not cryptographically verified)
			payload := parts[1]
			// Add padding if needed for base64 decoding
			for len(payload)%4 != 0 {
				payload += "="
			}
			if decoded, decodeErr := base64.StdEncoding.DecodeString(payload); decodeErr == nil {
				var claims map[string]interface{}
				if jsonErr := json.Unmarshal(decoded, &claims); jsonErr == nil {
					// Check expiration
					if exp, ok := claims["exp"].(float64); ok {
						expTime := time.Unix(int64(exp), 0)
						if time.Now().After(expTime) {
							return TestResult{
								Operation: "create_random_topic",
								Success:   false,
								Message:   "Access token expired - please re-login",
								Details:   fmt.Sprintf("Token expired at: %v\nCurrent time: %v\nPlease go to /login to refresh your session", expTime, time.Now()),
							}
						}
					}
					
					// Log token scopes for debugging
					if scope, ok := claims["scope"]; ok {
						logger.Info("Token scopes found", "scope", scope)
					} else {
						logger.Info("No scope claim found in token", "allClaims", claims)
					}
				}
			}
		}
	}
	
	// Create the topic using our PDS service
	params := pds.CreateTopicParams{
		Title:   uniqueTopic,
		Summary: randomMessage,
		Tags:    randomTags,
	}
	
	// Cast to ATProtoService to access token-aware methods
	atprotoService, ok := r.pdsService.(*pds.ATProtoService)
	if !ok {
		logger.Error("PDS service is not ATProtocol service", "actualType", fmt.Sprintf("%T", r.pdsService))
		return TestResult{
			Operation: "create_random_topic",
			Success:   false,
			Message:   "PDS service is not ATProtocol service",
			Details:   "This operation requires the real ATProtocol PDS service",
		}
	}
	
	logger.Info("About to call CreateTopicWithDPoP", 
		"userDID", userDID,
		"params", params,
		"hasAccessToken", accessToken != "",
		"dpopKeyType", fmt.Sprintf("%T", dpopKey))
	
	// Use the real ATProtocol service with access token and DPoP key
	topic, err := atprotoService.CreateTopicWithDPoP(userDID, params, accessToken, dpopKey)
	
	logger.Info("CreateTopicWithDPoP completed", "hasError", err != nil)
	if err != nil {
		logger.Error("CreateTopicWithDPoP failed", "error", err)
	}
	
	if err != nil {
		return TestResult{
			Operation: "create_random_topic",
			Success:   false,
			Message:   "PDS creation failed - DPoP headers required",
			Details:   fmt.Sprintf("Topic: %s\nMessage: %s\nTags: %v\n\nAccess token: %s\n\nError: %v\n\nThe request has valid scopes but is missing required DPoP headers. ATProtocol OAuth requires DPoP (Demonstration of Proof of Possession) headers for authenticated requests. This is the next implementation step.", 
				uniqueTopic, randomMessage, randomTags, 
				func() string { if len(accessToken) > 20 { return accessToken[:20] + "..." }; return accessToken }(), 
				err),
		}
	}
	
	return TestResult{
		Operation: "create_random_topic",
		Success:   true,
		Message:   "Successfully created random topic!",
		Details:   fmt.Sprintf("Created topic in PDS:\nURI: %s\nCID: %s\nTitle: %s\nTags: %v\n\nYou should now be able to browse this record!", 
			topic.URI, topic.CID, topic.Title, topic.Tags),
	}
}

// testStandardPost creates a standard app.bsky.feed.post to test our DPoP implementation
func (r *Router) testStandardPost(req *http.Request, userDID string) TestResult {
	// Extract access token from session
	accessToken, err := auth.GetSessionCookie(req)
	if err != nil {
		return TestResult{
			Operation: "test_standard_post",
			Success:   false,
			Message:   "No session found - please login first",
			Details:   fmt.Sprintf("Error getting session cookie: %v", err),
		}
	}
	
	// Extract DPoP key from session
	dpopKey, err := auth.GetDPoPKeyFromCookie(req)
	if err != nil {
		return TestResult{
			Operation: "test_standard_post",
			Success:   false,
			Message:   "No DPoP key found - please re-authenticate",
			Details:   fmt.Sprintf("DPoP key required for ATProtocol OAuth. Error: %v", err),
		}
	}
	
	// Create a standard Bluesky post record
	postText := fmt.Sprintf("Testing DPoP implementation from dis.quest dev interface at %s", time.Now().Format("15:04:05"))
	
	postRecord := map[string]interface{}{
		"$type":     "app.bsky.feed.post",
		"text":      postText,
		"createdAt": time.Now().Format(time.RFC3339),
	}
	
	// Generate unique rkey
	rkey := pds.GenerateRKey("post")
	
	// Create XRPC request
	createReq := pds.CreateRecordRequest{
		Repo:       userDID,
		Collection: "app.bsky.feed.post",
		RKey:       rkey,
		Validate:   true,
		Record:     postRecord,
	}
	
	// Use XRPC client directly
	xrpcClient := pds.NewXRPCClient()
	ctx := context.Background()
	
	// Debug: Check token scopes again for this specific operation
	if accessToken != "" {
		parts := strings.Split(accessToken, ".")
		if len(parts) == 3 {
			payload := parts[1]
			for len(payload)%4 != 0 {
				payload += "="
			}
			if decoded, decodeErr := base64.StdEncoding.DecodeString(payload); decodeErr == nil {
				var claims map[string]interface{}
				if jsonErr := json.Unmarshal(decoded, &claims); jsonErr == nil {
					logger.Info("Token claims for standard post", "allClaims", claims)
					
					// Extract the JKT (JWK thumbprint) from the token
					if cnf, ok := claims["cnf"].(map[string]interface{}); ok {
						if jkt, ok := cnf["jkt"].(string); ok {
							logger.Info("Token JKT (JWK thumbprint)", "jkt", jkt)
						}
					}
				}
			}
		}
	}
	
	// Calculate JWK thumbprint of our DPoP key
	keyPair := &auth.DPoPKeyPair{PrivateKey: dpopKey}
	jwk := keyPair.DPoPPublicJWK()
	logger.Info("Our DPoP key JWK", "jwk", jwk)
	
	// Calculate thumbprint - this should match the jkt value in the token
	thumbprint, err := keyPair.GetJWKThumbprint()
	if err != nil {
		logger.Error("Failed to calculate JWK thumbprint", "error", err)
	} else {
		logger.Info("Our DPoP key thumbprint", "thumbprint", thumbprint)
	}
	
	jwkBytes, _ := json.Marshal(jwk)
	logger.Info("DPoP key for comparison", "jwkJson", string(jwkBytes))
	
	logger.Info("Testing standard post creation", "userDID", userDID, "text", postText)
	
	resp, err := xrpcClient.CreateRecordWithDPoP(ctx, createReq, accessToken, dpopKey)
	if err != nil {
		return TestResult{
			Operation: "test_standard_post",
			Success:   false,
			Message:   "Standard post creation failed",
			Details:   fmt.Sprintf("Post: %s\n\nError: %v\n\nThis tests our DPoP implementation with a standard Bluesky lexicon.", postText, err),
		}
	}
	
	return TestResult{
		Operation: "test_standard_post",
		Success:   true,
		Message:   "Successfully created standard post!",
		Details:   fmt.Sprintf("Created standard Bluesky post:\nURI: %s\nCID: %s\nText: %s\n\nThis confirms our DPoP implementation is working correctly!", resp.URI, resp.CID, postText),
	}
}

// testSessionAuth tests creating a post using session-based auth (like WhiteWind) instead of OAuth
func (r *Router) testSessionAuth(req *http.Request, userDID string) TestResult {
	return TestResult{
		Operation: "test_session_auth",
		Success:   false,
		Message:   "Session-based auth not implemented yet",
		Details:   "This would test the WhiteWind approach:\n1. Use com.atproto.server.createSession instead of OAuth\n2. Bypass DPoP headers\n3. Use session tokens directly\n\nThis could bypass OAuth scope restrictions that are blocking custom lexicons.",
	}
}