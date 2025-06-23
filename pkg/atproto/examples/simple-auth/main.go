// Simple authentication example for ATProtocol Go client library
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jrschumacher/dis.quest/pkg/atproto"
)

func main() {
	// Replace these with your actual configuration
	config := atproto.Config{
		ClientID:       "https://myapp.example.com/client-metadata.json",
		RedirectURI:    "https://myapp.example.com/auth/callback",
		PDSEndpoint:    "https://bsky.social",
		JWKSPrivateKey: `{
			"keys": [{
				"kty": "EC",
				"crv": "P-256",
				"x": "your-x-coordinate",
				"y": "your-y-coordinate", 
				"d": "your-private-key",
				"alg": "ES256"
			}]
		}`,
		Scope: "atproto transition:generic",
	}

	// Create client
	client, err := atproto.New(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Note: DPoP keys are now managed automatically by the provider

	// In a real app, you would:
	// 1. Generate PKCE parameters
	// 2. Redirect user to GetAuthURL()
	// 3. Handle the callback to get the authorization code
	
	// For this example, we'll simulate having received the auth code
	fmt.Println("Step 1: Generate OAuth URL")
	authURL := client.GetAuthURL("example-state", "example-code-challenge")
	fmt.Printf("User would visit: %s\n\n", authURL)

	// Simulate having the authorization code from OAuth callback
	fmt.Println("Step 2: Exchange authorization code for session")
	fmt.Println("(In real usage, you'd get these from the OAuth callback)")
	
	// This would come from your OAuth callback handler
	authCode := "simulated-auth-code"
	codeVerifier := "example-code-verifier"
	// Note: DPoP nonce and auth server issuer are now handled automatically

	// Exchange code for authenticated session
	session, err := client.ExchangeCode(
		context.Background(),
		authCode,
		codeVerifier,
	)
	if err != nil {
		log.Fatalf("Failed to exchange code: %v", err)
	}

	fmt.Printf("✅ Successfully authenticated!\n")
	fmt.Printf("User DID: %s\n\n", session.GetUserDID())

	// Example: Create a simple record
	fmt.Println("Step 3: Create a record in the user's PDS")
	
	// Define a simple record type
	type ExampleRecord struct {
		Type      string    `json:"$type"`
		Message   string    `json:"message"`
		CreatedAt time.Time `json:"createdAt"`
	}

	// Create the record
	record := ExampleRecord{
		Type:      "com.example.message",
		Message:   "Hello from ATProtocol Go client!",
		CreatedAt: time.Now(),
	}

	result, err := session.CreateRecord("com.example.message", "hello-1", record)
	if err != nil {
		log.Fatalf("Failed to create record: %v", err)
	}

	fmt.Printf("✅ Record created successfully!\n")
	fmt.Printf("Record URI: %s\n", result.URI)
	fmt.Printf("Record CID: %s\n\n", result.CID)

	// Example: Retrieve the record
	fmt.Println("Step 4: Retrieve the record")
	
	var retrievedRecord ExampleRecord
	err = session.GetRecord("com.example.message", "hello-1", &retrievedRecord)
	if err != nil {
		log.Fatalf("Failed to get record: %v", err)
	}

	fmt.Printf("✅ Record retrieved successfully!\n")
	fmt.Printf("Message: %s\n", retrievedRecord.Message)
	fmt.Printf("Created: %s\n\n", retrievedRecord.CreatedAt.Format(time.RFC3339))

	// Example: List records in the collection
	fmt.Println("Step 5: List all records in collection")
	
	records, err := session.ListRecords("com.example.message", 10, "")
	if err != nil {
		log.Fatalf("Failed to list records: %v", err)
	}

	fmt.Printf("✅ Found %d records in collection\n", len(records.Records))
	for i, record := range records.Records {
		fmt.Printf("%d. %s\n", i+1, record.URI)
	}

	fmt.Println("\n🎉 Example completed successfully!")
	fmt.Println("You now have a working ATProtocol client that can:")
	fmt.Println("- Authenticate users with OAuth+DPoP")
	fmt.Println("- Create custom lexicon records") 
	fmt.Println("- Retrieve and list records from Personal Data Servers")
	fmt.Println("- Handle DPoP nonce requirements automatically")
}