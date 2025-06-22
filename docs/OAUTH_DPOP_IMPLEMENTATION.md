# OAuth + DPoP Implementation for ATProtocol

**Status: COMPLETED ✅**  
**Date: 2025-06-22**  
**Result: Successfully implemented working OAuth with DPoP for ATProtocol/Bluesky integration**

## Problem Statement

Initial implementation was failing with "Bad token scope" errors when attempting to write records to Personal Data Servers (PDS) via ATProtocol OAuth. The core issue was missing proper client authentication for DPoP (Demonstration of Proof-of-Possession) flows.

## Root Cause Analysis

### Key Discovery: Two-Layer Authentication Required

ATProtocol OAuth with DPoP requires **two separate authentication mechanisms**:

1. **Client Authentication**: `private_key_jwt` to authenticate the OAuth client during token exchange
2. **Proof of Possession**: DPoP headers to prove possession of the bound key

### Critical Technical Issues Identified

1. **Missing `private_key_jwt`**: Our manual implementation lacked proper client assertion JWTs
2. **Authorization Header Format**: Must use `"DPoP"` not `"Bearer"` per RFC
3. **PDS Resolution**: Must resolve DID to actual PDS endpoint, not use generic endpoints
4. **DPoP Nonce Handling**: Server requires nonce retry pattern for some operations
5. **Access Token Hash**: DPoP JWT must include `ath` claim binding to access token

## Solution: OAuth Provider Abstraction + Tangled Integration

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                OAuth Provider Interface                     │
├─────────────────────────────────────────────────────────────┤
│  GetAuthURL() | ExchangeToken() | CreateAuthorizedClient()  │
└─────────────────────────────────────────────────────────────┘
                               │
                ┌──────────────┴──────────────┐
                │                             │
    ┌───────────▼──────────┐      ┌──────────▼─────────────┐
    │   Manual Provider    │      │   Tangled Provider     │
    │   (Original impl)    │      │   (Working solution)   │
    │   - OAuth2 + PKCE    │      │   - tangled.sh library │
    │   - Manual DPoP      │      │   - Automatic client   │
    │   - Missing client   │      │     authentication     │
    │     authentication  │      │   - Full DPoP support   │
    │   - No PAR support   │      │   - PAR integration    │
    └──────────────────────┘      └────────────────────────┘
```

### Implementation Files

```
internal/oauth/
├── provider.go     # Interface definition
├── factory.go      # Provider factory  
├── manual.go       # Original implementation (preserved)
├── tangled.go      # Working tangled-sh implementation
└── config.go       # Shared configuration

internal/auth/session.go    # DPoP key & nonce management
internal/pds/xrpc.go       # DPoP-enabled XRPC client
server/app/dev.go          # Testing interface
config.yaml                # Provider selection
```

## Technical Implementation Details

### 1. DPoP JWT Structure (RFC Compliant)

**Header:**
```json
{
  "typ": "dpop+jwt",
  "alg": "ES256", 
  "jwk": {
    "kty": "EC",
    "crv": "P-256",
    "x": "...", 
    "y": "...",
    "alg": "ES256",
    "use": "sig"
  }
}
```

**Payload:**
```json
{
  "jti": "random-nonce",
  "htm": "POST",
  "htu": "https://pds-endpoint/xrpc/com.atproto.repo.createRecord",
  "iat": 1640995200,
  "nonce": "server-provided-nonce",  // From DPoP-Nonce header
  "ath": "base64url(SHA256(access_token))"  // Token binding
}
```

### 2. Authorization Header Format

**CRITICAL**: Must use `DPoP` not `Bearer`

```http
Authorization: DPoP eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiJ9...
DPoP: eyJ0eXAiOiJkcG9wK2p3dCIsImFsZyI6IkVTMjU2In0...
```

### 3. Nonce Retry Pattern

```go
// First request (no nonce)
resp, err := makeRequest("")
if resp.StatusCode == 401 && resp.Header.Get("DPoP-Nonce") != "" {
    // Retry with server-provided nonce
    nonce := resp.Header.Get("DPoP-Nonce")
    resp, err = makeRequest(nonce)
}
```

### 4. PDS Resolution

**Wrong:** Using generic endpoints like `bsky.social`  
**Right:** Resolve DID to actual PDS via PLC directory

```go
// Resolve did:plc:abc123 → https://meadow.us-east.host.bsky.network
pdsEndpoint, err := xrpcClient.ResolvePDS(userDID)
```

### 5. PAR (Pushed Authorization Request) Integration

**Critical Component**: ATProtocol OAuth servers require PAR for DPoP flows to obtain server-generated nonces.

**PAR Flow in ATProtocol**:
```
1. Client → PAR Endpoint: POST authorization parameters
   {
     "client_id": "...",
     "response_type": "code", 
     "scope": "atproto transition:generic",
     "code_challenge": "...",
     "code_challenge_method": "S256"
   }

2. PAR Response ← Server: Returns request_uri + DPoP nonce
   {
     "request_uri": "urn:ietf:params:oauth:request_uri:...",
     "expires_in": 600
   }
   Headers: { "DPoP-Nonce": "server-generated-nonce" }

3. Client → Authorization: Redirect with request_uri  
   https://auth-server/authorize?request_uri=urn:ietf:params:oauth:request_uri:...

4. Authorization Code ← Server: Standard OAuth callback
   https://client/callback?code=...&state=...

5. Client → Token Exchange: Include DPoP nonce from PAR
   POST /token with DPoP JWT containing nonce from step 2
```

**Key PAR Benefits**:
- **DPoP Nonce Acquisition**: Server provides nonce required for DPoP JWT
- **Auth Server Discovery**: Identifies actual authorization server endpoint  
- **Secure Parameter Transport**: Large OAuth params sent securely via POST
- **Request Integrity**: Prevents parameter tampering in authorization URLs

**PAR Implementation in Tangled Provider**:
```go
// Extract DPoP nonce and auth server issuer from PAR response
parResp := oauth.PushedAuthorizationRequest(...)
dpopNonce := parResp.DPoPNonce           // From DPoP-Nonce header
authServerIss := parResp.AuthServerIss   // Actual auth server
requestURI := parResp.RequestURI         // For authorization redirect
```

## Key Discoveries from Tangled.sh Analysis

### Working OAuth Library Integration

**Library**: `tangled.sh/icyphox.sh/atproto-oauth@v0.0.0-20250526154904-3906c5336421`

**Key advantages:**
- Automatic `private_key_jwt` client assertion creation
- Built-in DPoP nonce handling  
- Proper JWK management and thumbprint calculation
- Handles auth server issuer resolution
- Tested against real ATProtocol infrastructure

### Critical Code Patterns from Tangled.sh

```go
// Proper token exchange with all required parameters
tokenResp, err := oauthClient.InitialTokenRequest(
    ctx,
    code,                           // Authorization code
    oauthRequest.AuthserverIss,     // Auth server issuer  
    oauthRequest.PkceVerifier,      // PKCE verifier
    oauthRequest.DpopAuthserverNonce, // DPoP nonce from PAR
    jwk,                            // JWK directly (not string!)
)
```

## Configuration

### Provider Selection

```yaml
# config.yaml
oauth_provider: tangled  # Use working tangled implementation
# oauth_provider: manual   # Fallback to original (for debugging)
```

### Testing Interface

Development testing available at `/dev/pds` with operations:
- `create_random_topic` - Test custom lexicon creation
- `test_standard_post` - Validate with standard Bluesky records  
- `check_server_scopes` - Verify OAuth server capabilities
- Token expiration and scope inspection

## Error Messages Encountered & Solutions

| Error | Cause | Solution |
|-------|-------|----------|
| "Bad token scope" | Missing `private_key_jwt` | Use tangled provider |
| "OAuth tokens are meant for PDS access only" | Wrong endpoint | Implement DID resolution |
| "use_dpop_nonce" | Server requires nonce | Add retry with nonce |
| "DPoP ath mismatch" | Missing token hash | Add `ath` claim to DPoP JWT |
| JWK parsing errors | JWKS vs JWK format | Extract single JWK from JWKS |

## Performance & Security

### Security Features
- ✅ DPoP proof-of-possession prevents token replay
- ✅ PKCE prevents authorization code interception  
- ✅ Private keys stored securely in HttpOnly cookies
- ✅ Token binding via JWK thumbprint (`jkt` claim)
- ✅ Nonce prevents DPoP JWT replay

### Error Handling
- ✅ Automatic retry for nonce requirements
- ✅ Comprehensive logging for debugging
- ✅ Graceful fallback to manual provider
- ✅ Token expiration detection and user guidance

## Rollback Strategy

**Instant rollback available**: Change `oauth_provider: manual` in config.yaml

**Zero risk**: Original implementation preserved in `manual.go`  
**Zero data loss**: All session/cookie handling identical between providers

## Future Improvements

1. **Token Refresh**: Implement automatic refresh token handling
2. **Session Persistence**: Add database storage for long-term sessions  
3. **Multi-PDS**: Support users with multiple PDS instances
4. **Rate Limiting**: Add request rate limiting for PDS operations
5. **Caching**: Cache DID resolution results

## References

- **ATProtocol OAuth Spec**: https://atproto.com/specs/oauth
- **DPoP RFC**: https://datatracker.ietf.org/doc/html/rfc9449
- **GitHub Issue #3212**: DPoP authorization header format requirements
- **Tangled.sh Source**: Working reference implementation
- **WhiteWind Alternative**: Session-based auth bypass for comparison

---

**Implementation Status**: ✅ **COMPLETE AND WORKING**  
**Next Steps**: Monitor production usage and implement future improvements as needed