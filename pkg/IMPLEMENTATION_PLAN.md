# ATProtocol Package Consolidation Plan

## Current State Assessment (December 2024)

After implementing the initial migration, we have identified significant **code duplication** and **scattered ATProtocol functionality** across internal packages. This assessment outlines the proper consolidation strategy to create a complete, reusable ATProtocol SDK.

### ✅ **What's Working**
- Complete OAuth + DPoP authentication flow 
- Custom lexicon creation (`quest.dis.topic` records in production PDS)
- End-to-end PDS operations with nonce retry logic
- All tests passing, production-ready implementations

### ❌ **What Needs Consolidation**

#### **Code Duplication Analysis**
| Functionality | Current Locations | Issue |
|---------------|-------------------|-------|
| **DPoP Implementation** | `internal/auth/session.go`, `pkg/atproto/oauth/dpop.go`, `pkg/atproto/xrpc/dpop_utils.go` | 3 different implementations |
| **OAuth Clients** | `internal/auth/auth.go`, `internal/oauth/manual.go`, `internal/oauth/tangled.go`, `pkg/atproto/oauth/oauth.go` | Multiple OAuth approaches |
| **XRPC Operations** | `internal/pds/xrpc.go`, `pkg/atproto/xrpc/client.go` | Wrapper layer for compatibility |
| **Session Management** | `internal/auth/session.go`, `pkg/atproto/session.go` | Two different session models |

#### **Misplaced ATProtocol Code**
- **`internal/auth/`** contains production DPoP, PKCE, PAR implementations that should be in `pkg/atproto`
- **`internal/oauth/`** contains OAuth providers that should be part of the SDK  
- **`internal/pds/`** mostly wraps `pkg/atproto` for backward compatibility

## Consolidation Strategy

### **Target Architecture**

```
pkg/atproto/          # Complete ATProtocol SDK
├── client.go         # Main client with provider selection
├── session.go        # Unified session management  
├── config.go         # Configuration structures
├── errors.go         # ATProtocol error definitions
├── oauth/
│   ├── oauth.go      # OAuth interface and implementations
│   ├── dpop.go       # Single DPoP implementation
│   ├── pkce.go       # PKCE implementation  
│   ├── par.go        # PAR implementation
│   └── providers.go  # Manual, Tangled providers
├── xrpc/
│   ├── client.go     # Enhanced XRPC with nonce retry
│   ├── records.go    # Record CRUD operations
│   └── resolver.go   # DID resolution
└── lexicon/
    ├── lexicon.go    # Custom lexicon support
    └── validation.go # Lexicon validation

internal/             # Web application concerns only
├── auth/
│   ├── cookies.go    # HTTP cookie management
│   ├── middleware.go # HTTP middleware integration
│   └── config.go     # Web app auth configuration  
├── middleware/       # HTTP middleware (using pkg/atproto)
└── handlers/         # HTTP handlers (using pkg/atproto)
```

### **Phase 1: Core ATProtocol Consolidation**

#### **1.1 Consolidate DPoP Implementation** ✅ **COMPLETED**
**Moved to `pkg/atproto/oauth/dpop.go`:**
- ✅ `internal/auth/session.go` lines 241-332: `CreateDPoPJWT*` functions 
- ✅ `internal/auth/session.go` lines 64-115: `DPoPKeyPair` struct and methods
- ✅ `internal/auth/session.go` lines 334-357: `CalculateJWKThumbprint`
- ✅ Removed duplicate implementations in `pkg/atproto/xrpc/dpop_utils.go`
- ✅ Updated `pkg/atproto/xrpc/client.go` to use consolidated API
- ✅ All existing APIs preserved with delegation pattern

**Result**: ✅ Single, production-tested DPoP implementation with zero breaking changes

#### **1.2 Consolidate OAuth Providers**
**Move to `pkg/atproto/oauth/`:**
- `internal/auth/auth.go`: PKCE implementation and DPoP transport ✅
- `internal/oauth/manual.go`: Manual OAuth provider implementation ✅
- `internal/oauth/tangled.go`: Tangled OAuth provider implementation ✅
- `internal/auth/discover.go`: Authorization server discovery ✅
- `internal/auth/par.go`: PAR implementation ✅

**Enhanced Client Interface:**
```go
type Client struct {
    config   *Config
    provider OAuthProvider
}

// Support multiple OAuth implementations
func New(config Config, providerType ProviderType) (*Client, error) {
    switch providerType {
    case ProviderTypeManual:
        return newManualClient(config)
    case ProviderTypeTangled:  
        return newTangledClient(config)
    default:
        return newManualClient(config) // Fallback to proven implementation
    }
}
```

#### **1.3 Enhanced XRPC Client**
**Move to `pkg/atproto/xrpc/`:**
- `internal/pds/xrpc.go`: Nonce retry logic (lines 151-184) ✅
- `internal/pds/lexicons.go`: Custom lexicon support ✅
- Enhanced error handling and logging compatibility ✅

#### **1.4 Unified Session Management**
**Enhanced `pkg/atproto/session.go`:**
```go
type Session struct {
    client       *Client
    accessToken  string
    refreshToken string  
    userDID      string
    dpopKey      *ecdsa.PrivateKey
    expiresIn    int64
    
    // Core ATProtocol operations
    CreateRecord(collection, rkey string, record interface{}) error
    GetRecord(collection, rkey string, result interface{}) error
    ListRecords(collection string, limit int, cursor string) (*ListResponse, error)
    
    // Session management
    IsExpired() bool
    Refresh(ctx context.Context) error
    
    // Optional web integration (for internal/ packages)
    SaveToCookies(w http.ResponseWriter, isDev bool) error // Optional
}
```

### **Phase 2: Internal Package Cleanup**

#### **2.1 Minimal `internal/auth/`**
**Keep only web application concerns:**
- HTTP cookie management (`SetSessionCookie*`, `GetSessionCookie`)
- Web-specific secure cookie settings
- Environment-specific configuration (dev vs prod)

#### **2.2 Remove `internal/oauth/`**
- OAuth providers moved to `pkg/atproto/oauth/`
- Keep only minimal factory if needed for config-based selection

#### **2.3 Minimal `internal/pds/`**
- Keep application service interfaces
- Remove XRPC wrapper (use `pkg/atproto` directly)
- Keep mock implementations for testing

### **Phase 3: Update Application Code**

#### **3.1 Direct pkg/atproto Usage**
```go
// Before: Multiple internal imports
import (
    "github.com/jrschumacher/dis.quest/internal/auth"
    "github.com/jrschumacher/dis.quest/internal/oauth" 
    "github.com/jrschumacher/dis.quest/internal/pds"
)

// After: Single import
import "github.com/jrschumacher/dis.quest/pkg/atproto"

// Simple client usage
client, err := atproto.New(atproto.Config{
    ClientID:    cfg.OAuthClientID,
    RedirectURI: cfg.OAuthRedirectURL,
    JWKSPrivate: cfg.JWKSPrivate,
}, atproto.ProviderTypeManual)

// OAuth flow  
authURL := client.GetAuthURL(state, codeChallenge)
session, err := client.ExchangeCode(code, codeVerifier)

// PDS operations
err = session.CreateRecord("quest.dis.topic", "my-topic", topicData)
```

#### **3.2 Middleware Updates**
```go
// Simplified middleware using pkg/atproto
func AuthMiddleware(client *atproto.Client) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Load session from cookies
            session, err := atproto.LoadSessionFromCookies(r, client)
            if err != nil {
                http.Redirect(w, r, "/login", http.StatusFound)
                return
            }
            
            // Auto-refresh if needed
            if session.IsExpired() {
                if err := session.Refresh(r.Context()); err != nil {
                    http.Redirect(w, r, "/login", http.StatusFound) 
                    return
                }
                session.SaveToCookies(w, isDev)
            }
            
            ctx := context.WithValue(r.Context(), "session", session)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Implementation Timeline

### **Week 1: Core Consolidation**
- [x] **COMPLETED**: Move DPoP implementation to `pkg/atproto/oauth/dpop.go` ✅
- [ ] Move OAuth providers to `pkg/atproto/oauth/`
- [ ] Enhance XRPC client with nonce retry logic
- [ ] Create unified client interface

### **Week 2: Session & XRPC Enhancement** 
- [ ] Unified session management with web integration
- [ ] Enhanced XRPC client with all features
- [ ] Move lexicon support to `pkg/atproto/lexicon/`
- [ ] Update examples to show simplified usage

### **Week 3: Internal Package Cleanup**
- [ ] Minimize `internal/auth/` to web concerns only
- [ ] Remove `internal/oauth/` (move to pkg/atproto)
- [ ] Simplify `internal/pds/` service interfaces
- [ ] Update all application code to use `pkg/atproto` directly

### **Week 4: Testing & Documentation**
- [ ] Comprehensive testing of consolidated package
- [ ] Update documentation and examples
- [ ] Performance testing and optimization
- [ ] Migration guide for other projects

## Benefits

### **For pkg/atproto (Reusable SDK)**
✅ **Complete SDK**: Single import for all ATProtocol operations  
✅ **Production Tested**: Built from proven, working implementations  
✅ **Standards Compliant**: RFC-compliant DPoP, OAuth2, PKCE, PAR  
✅ **Minimal Dependencies**: Only standard library + tangled.sh OAuth  
✅ **Simple API**: Dead-simple interface for Go developers  

### **For dis.quest Application**
✅ **Reduced Complexity**: Single ATProtocol import instead of 3-4 internal packages  
✅ **Better Maintainability**: Core ATProtocol logic centralized  
✅ **Enhanced Features**: Access to complete SDK capabilities  
✅ **Future Proof**: Automatic benefits from SDK improvements  

### **For Other Projects**
✅ **Drop-in ATProtocol Support**: Complete client library ready to use  
✅ **Battle Tested**: Proven in production with real PDS operations  
✅ **Well Documented**: Comprehensive examples and documentation  
✅ **Community Friendly**: Open source, reusable ATProtocol client  

## Risk Mitigation

### **Technical Risks**
- **Breaking Changes**: Maintain backward compatibility during migration
- **Test Coverage**: Comprehensive testing at each consolidation step  
- **Rollback Plan**: Keep internal packages until pkg/atproto is fully validated

### **Implementation Risks**
- **Incremental Migration**: Move one component at a time
- **Parallel Development**: Keep both approaches working during transition
- **Validation**: Test each phase with existing application features

## Success Criteria

### **Functional Requirements**
- [ ] Complete OAuth flow in <10 lines of code using pkg/atproto
- [ ] All existing dis.quest functionality preserved
- [ ] Custom lexicon operations working identically  
- [ ] Token refresh and DPoP nonce handling automatic
- [ ] Production PDS compatibility maintained

### **Code Quality Requirements**
- [ ] Single source of truth for all ATProtocol operations
- [ ] No code duplication between pkg/atproto and internal packages
- [ ] Clean separation: ATProtocol logic in pkg/, web logic in internal/
- [ ] Comprehensive test coverage for all consolidated components

### **Developer Experience Requirements**
- [ ] Simpler imports and initialization for application code
- [ ] Clear, documented API for pkg/atproto
- [ ] Easy migration path for other applications
- [ ] Comprehensive examples showing best practices

---

**Next Action**: Begin Phase 1.1 - Consolidate DPoP implementation into `pkg/atproto/oauth/dpop.go`