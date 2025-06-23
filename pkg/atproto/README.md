# ATProtocol Go Client Library

A production-ready Go client library for building ATProtocol applications. Extracted from the proven implementation in [dis.quest](https://github.com/jrschumacher/dis.quest) with complete OAuth+DPoP authentication and custom lexicon support.

## Features

✅ **Production-Ready**: Battle-tested with real Personal Data Servers  
✅ **Complete OAuth+DPoP**: RFC-compliant authentication with automatic nonce handling  
✅ **Custom Lexicons**: Support for creating and managing custom record types  
✅ **Simple Interface**: Dead-simple API for Go developers  
✅ **Standards Compliant**: Full ATProtocol specification compliance  

## Quick Start

### Installation

```bash
go get github.com/jrschumacher/dis.quest/pkg/atproto
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jrschumacher/dis.quest/pkg/atproto"
    "github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
)

func main() {
    // Create client with your app's configuration
    client, err := atproto.New(atproto.Config{
        ClientID:       "https://myapp.com/client-metadata.json",
        RedirectURI:    "https://myapp.com/auth/callback",
        JWKSPrivateKey: yourJWKSPrivateKey, // JSON Web Key Set
        Scope:          "atproto transition:generic",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Generate OAuth URL for user authentication
    authURL := client.GetAuthURL(state, codeChallenge)
    fmt.Printf("Visit: %s\n", authURL)

    // After user completes OAuth flow, exchange code for session
    dpopKey, _ := oauth.GenerateDPoPKeyPair()
    session, err := client.ExchangeCode(
        context.Background(), 
        authCode, 
        codeVerifier,
        dpopKey.PrivateKey,
        dpopNonce,
        "https://bsky.social",
    )
    if err != nil {
        log.Fatal(err)
    }

    // Now you can interact with the user's Personal Data Server
    fmt.Printf("Authenticated as: %s\n", session.GetUserDID())
}
```

## Working with Records

### Creating Custom Lexicon Records

```go
// Define your custom record type
type TopicRecord struct {
    Type        string    `json:"$type"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"createdAt"`
}

// Create a new topic record
topic := TopicRecord{
    Type:        "quest.dis.topic",
    Title:       "My Discussion Topic",
    Description: "A place to discuss interesting things",
    CreatedAt:   time.Now(),
}

// Store in user's PDS
result, err := session.CreateRecord("quest.dis.topic", "my-topic-1", topic)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created record: %s\n", result.URI)
// Output: at://did:plc:abc123.../quest.dis.topic/my-topic-1
```

### Retrieving Records

```go
// Get a specific record
var retrievedTopic TopicRecord
err := session.GetRecord("quest.dis.topic", "my-topic-1", &retrievedTopic)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Topic: %s\n", retrievedTopic.Title)
```

### Listing Records

```go
// List all records in a collection
records, err := session.ListRecords("quest.dis.topic", 50, "")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d topics\n", len(records.Records))
for _, record := range records.Records {
    fmt.Printf("- %s\n", record.URI)
}
```

### Updating Records

```go
// Update an existing record
topic.Description = "Updated description"
result, err := session.UpdateRecord("quest.dis.topic", "my-topic-1", topic)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Updated record: %s\n", result.URI)
```

### Deleting Records

```go
// Delete a record
err := session.DeleteRecord("quest.dis.topic", "my-topic-1")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Record deleted")
```

## Advanced Usage

### Custom XRPC Operations

```go
// Get access to the lower-level XRPC client for custom operations
xrpcClient := client.NewXRPCClient()

// Use custom XRPC methods directly
response, err := xrpcClient.CreateRecord(
    context.Background(),
    session.GetUserDID(),
    "my.custom.lexicon",
    "record-key",
    customRecord,
    session.GetAccessToken(),
    session.GetDPoPKey(),
)
```

### DPoP Key Management

```go
import "github.com/jrschumacher/dis.quest/pkg/atproto/oauth"

// Generate a new DPoP key pair
dpopKey, err := oauth.GenerateDPoPKeyPair()
if err != nil {
    log.Fatal(err)
}

// Encode for storage
pemEncoded, err := dpopKey.EncodeToPEM()
if err != nil {
    log.Fatal(err)
}

// Later, decode from storage
restoredKey, err := oauth.DecodeFromPEM(pemEncoded)
if err != nil {
    log.Fatal(err)
}
```

## Configuration

### Required Configuration

- **ClientID**: Your application's client identifier (URL to client metadata)
- **RedirectURI**: OAuth callback URL for your application
- **JWKSPrivateKey**: JSON Web Key Set containing your app's private keys

### Optional Configuration

- **ClientURI**: Your application's homepage URL
- **JWKSPublicKey**: Public keys (derived from private keys if not provided)
- **Scope**: OAuth scope (defaults to "atproto transition:generic")

### Example Client Metadata

Your `ClientID` should point to a JSON document like this:

```json
{
  "client_id": "https://myapp.com/client-metadata.json",
  "client_name": "My ATProtocol App",
  "client_uri": "https://myapp.com",
  "redirect_uris": ["https://myapp.com/auth/callback"],
  "scope": "atproto transition:generic",
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "token_endpoint_auth_method": "private_key_jwt",
  "token_endpoint_auth_signing_alg": "ES256",
  "jwks_uri": "https://myapp.com/.well-known/jwks.json"
}
```

## Authentication Flow

### 1. Generate DPoP Key Pair

```go
dpopKey, err := oauth.GenerateDPoPKeyPair()
// Store this securely - you'll need it for the entire session
```

### 2. Generate PKCE Parameters

```go
codeVerifier := generateCodeVerifier() // Your implementation
codeChallenge := generateCodeChallenge(codeVerifier) // Your implementation
state := generateRandomState() // Your implementation
```

### 3. Redirect User to Authorization URL

```go
authURL := client.GetAuthURL(state, codeChallenge)
// Redirect user to authURL
```

### 4. Handle OAuth Callback

```go
// Extract code from callback
authCode := extractCodeFromCallback(req) // Your implementation

// Exchange for session
session, err := client.ExchangeCode(
    context.Background(),
    authCode,
    codeVerifier,
    dpopKey.PrivateKey,
    "", // nonce (handled automatically)
    "https://bsky.social", // auth server
)
```

## Error Handling

The library provides detailed error messages for common scenarios:

```go
result, err := session.CreateRecord("invalid.lexicon", "key", record)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "DPoP nonce"):
        // Automatic retry with nonce (handled internally)
    case strings.Contains(err.Error(), "not found"):
        // Record doesn't exist
    case strings.Contains(err.Error(), "unauthorized"):
        // Token expired or invalid
    default:
        log.Printf("Unexpected error: %v", err)
    }
}
```

## Best Practices

### 1. DPoP Key Security
- Generate DPoP keys securely and store them safely
- Never log or expose DPoP private keys
- Rotate keys periodically for security

### 2. Custom Lexicons
- Always set `validate: false` when working with custom lexicons
- Use clear, descriptive lexicon namespaces (e.g., `myapp.records.type`)
- Include `$type` field in all record structures

### 3. Error Handling
- Always check for errors from PDS operations
- Implement retry logic for transient failures
- Handle token expiration gracefully

### 4. Performance
- Reuse XRPC clients when possible
- Implement pagination for large record collections
- Cache DID to PDS endpoint resolutions

## Dependencies

This library uses minimal, proven dependencies:

- `tangled.sh/icyphox.sh/atproto-oauth` - Production OAuth provider
- Standard library packages for crypto and HTTP operations

## Examples

See the `examples/` directory for complete working examples:

- `examples/simple-auth/` - Basic authentication flow
- `examples/custom-lexicon/` - Working with custom record types  
- `examples/full-app/` - Complete application template

## Contributing

This library is extracted from the working implementation in [dis.quest](https://github.com/jrschumacher/dis.quest). 

## License

MIT License - see LICENSE file for details.

## Status

This library is based on production-tested code from dis.quest with the following proven capabilities:

- ✅ Complete OAuth + DPoP authentication working end-to-end
- ✅ Custom lexicon creation and CRUD operations (`quest.dis.topic` records successfully stored)
- ✅ Production PDS integration with real Personal Data Servers
- ✅ DPoP nonce handling and token refresh with session preservation

Last updated: 2025-06-22