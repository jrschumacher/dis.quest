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

#### **1.2 Consolidate OAuth Providers** ✅ **COMPLETED**
**Moved to `pkg/atproto/oauth/`:**
- ✅ `internal/auth/auth.go`: PKCE implementation and DPoP transport → `pkce.go`
- ✅ `internal/oauth/manual.go`: Manual OAuth provider implementation → `providers.go`
- ✅ `internal/oauth/tangled.go`: Tangled OAuth provider implementation → `providers.go`
- ✅ `internal/auth/discover.go`: Authorization server discovery → `discovery.go`
- ✅ `internal/auth/par.go`: PAR implementation → `par.go`

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

#### **1.3 Enhanced XRPC Client** ✅ **COMPLETED**
**Already in place:**
- ✅ `pkg/atproto/xrpc/client.go`: Nonce retry logic already implemented (lines 163-194)
- ✅ `internal/pds/xrpc.go`: Already acts as wrapper around pkg/atproto/xrpc
- ✅ Enhanced error handling and logging compatibility working

#### **1.4 Unified Session Management** ✅ **COMPLETED**
**Enhanced `pkg/atproto/client.go`:**
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

### **Phase 2: Internal Package Cleanup** ✅ **MOSTLY COMPLETED**

#### **2.1 Minimal `internal/auth/`** ✅ **COMPLETED**
**Minimized to web application concerns only:**
- ✅ HTTP cookie management (`SetSessionCookie*`, `GetSessionCookie`) → `auth_web.go`
- ✅ SessionWrapper for backward compatibility → `session_wrapper.go`
- ✅ Compatibility layer delegating to pkg/atproto → `compat.go`
- ✅ Error definitions preserved → `errors.go`
- ✅ All OAuth/DPoP/PKCE implementation moved to pkg/atproto

#### **2.2 Remove `internal/oauth/`** ✅ **COMPLETED**
**OAuth providers successfully migrated and flattened:**
- ✅ OAuth providers moved to `pkg/atproto/oauth/providers.go`
- ✅ ~~Minimal compatibility service created → `service.go`~~ **FLATTENED**
- ✅ Direct `pkg/atproto.Client` usage replaces `internal/oauth.Service`
- ✅ Provider selection simplified to manual (tangled + manual flattened)
- ✅ All application code updated to use `*atproto.Client` directly
- ✅ Middleware updated to work with `*atproto.Client`
- ✅ Authentication handlers updated to use direct client
- ✅ `/internal/oauth/` directory completely removed

#### **2.3 Minimal `internal/pds/`** ⏳ **PENDING**
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

### **✅ CURRENT STATUS (2025-06-22)**
**Phases 1 & 2 are COMPLETED**: Core ATProtocol consolidation and internal package cleanup are done. The only remaining task in Phase 2 is simplifying `internal/pds/`. We have successfully:

- ✅ **Consolidated all ATProtocol functionality** into `pkg/atproto/`
- ✅ **Flattened OAuth architecture** - removed wrapper services, direct client usage
- ✅ **Minimized internal packages** to web-only concerns (cookies, middleware)
- ✅ **Unified session management** with backward-compatible SessionWrapper
- ✅ **Updated all application code** to use `pkg/atproto` directly

**Next**: Implement Phase 3 (Universal Session Management Abstraction) which will make Phase 2.3 (`internal/pds/` simplification) much cleaner and provide a complete reusable ATProtocol SDK

### **Week 1: Core Consolidation** ✅ **COMPLETED**
- [x] **COMPLETED**: Move DPoP implementation to `pkg/atproto/oauth/dpop.go` ✅
- [x] **COMPLETED**: Move OAuth providers to `pkg/atproto/oauth/` ✅
- [x] **COMPLETED**: Enhanced XRPC client with nonce retry logic ✅
- [x] **COMPLETED**: Create unified client interface ✅

### **Week 2: Internal Package Cleanup** ✅ **COMPLETED**
- [x] **COMPLETED**: Minimize `internal/auth/` to web concerns only ✅
- [x] **COMPLETED**: Remove `internal/oauth/` completely (flattened) ✅  
- [x] **COMPLETED**: Update all application code to use `pkg/atproto` directly ✅
- [x] **COMPLETED**: Flatten OAuth service layer for direct client usage ✅
- [ ] **PENDING**: Simplify `internal/pds/` service interfaces

### **Week 3: Universal Session Management Abstraction** ✅ **MOSTLY COMPLETED**
- [x] **COMPLETED**: Deep analysis of internal/auth abstraction opportunities ✅
- [x] **COMPLETED**: Create pkg/atproto/session/ package with generic interfaces ✅
- [x] **COMPLETED**: Implement storage backends (Memory, Cookie, File) ✅
- [x] **COMPLETED**: Update atproto.Client to use session manager ✅
- [x] **COMPLETED**: Enhanced examples showing new session management ✅
- [ ] **IN PROGRESS**: Refactor internal/auth to use new session system
- [ ] **PENDING**: Move lexicon support to `pkg/atproto/lexicon/`

### **Week 4: Testing & Documentation**
- [ ] Comprehensive testing of consolidated package
- [ ] Update documentation and examples
- [ ] Performance testing and optimization
- [ ] Migration guide for other projects

### **Phase 3: Universal Session Management Abstraction** 🚀 **NEW - IN PROGRESS**

Based on deep analysis of `/internal/auth`, we identified a powerful abstraction opportunity. The current architecture mixes ATProtocol session semantics with web-specific transport concerns. By separating these, we can create a universal session management system.

#### **3.1 Session Abstraction Architecture** ⏳ **IN PROGRESS**
**Create generic session management layer in `pkg/atproto/session/`:**

```go
// Generic session operations (storage-agnostic)
type Manager struct {
    client  *atproto.Client
    storage SessionStorage
    config  Config
}

type SessionStorage interface {
    Store(ctx context.Context, key string, data *SessionData) error
    Load(ctx context.Context, key string) (*SessionData, error)
    Delete(ctx context.Context, key string) error
    Cleanup(ctx context.Context) error
}

// Built-in storage implementations
type MemoryStorage struct{}      // Development/testing
type CookieStorage struct{}      // Simple web apps
type FileStorage struct{}        // CLI applications
type RedisStorage struct{}       // Production web apps (extensible)
```

#### **3.2 Web Transport Layer Separation**
**Keep HTTP-specific concerns in `internal/auth` as thin wrapper:**

```go
// internal/auth/web_session.go - delegates to pkg/atproto/session
type WebSessionManager struct {
    sessionManager *session.Manager
    cookieConfig   CookieConfig
}

// HTTP-specific methods
func (w *WebSessionManager) SaveToCookies(ctx context.Context, w http.ResponseWriter, session *atproto.Session) error
func (w *WebSessionManager) LoadFromCookies(ctx context.Context, r *http.Request) (*atproto.Session, error)
```

### **Phase 4: Extensible Session Storage Interface** 🚀 **ENHANCED**

The session abstraction enables pluggable storage backends for different application types:

#### **4.1 Session Storage Interface Design**
**Create pluggable session storage interface:**
```go
// SessionStorage defines the interface for session persistence
type SessionStorage interface {
    // Store saves a session with the given key
    Store(ctx context.Context, key string, session *SessionData) error
    
    // Load retrieves a session by key
    Load(ctx context.Context, key string) (*SessionData, error)
    
    // Delete removes a session
    Delete(ctx context.Context, key string) error
    
    // Cleanup removes expired sessions (called periodically)
    Cleanup(ctx context.Context) error
}

// SessionData contains the session information to be stored
type SessionData struct {
    AccessToken  string
    RefreshToken string
    UserDID      string
    DPoPKey      *ecdsa.PrivateKey // Encrypted in storage
    ExpiresAt    time.Time
    CreatedAt    time.Time
}
```

#### **4.2 Built-in Storage Implementations**
**Default implementations provided:**
- **CookieSessionStorage** (current implementation) - Default for simple apps
- **MemorySessionStorage** - In-memory storage for development/testing
- **EncryptedCookieStorage** - Cookie storage with AES encryption for DPoP keys

#### **4.3 Enhanced Client Configuration**
**Extended client to support custom session storage:**
```go
type Config struct {
    ClientID       string
    RedirectURI    string
    PDSEndpoint    string
    JWKSPrivateKey string
    Scope          string
    
    // Optional: Custom session storage (defaults to cookies)
    SessionStorage SessionStorage
    
    // Optional: Session encryption key for sensitive data
    SessionEncryptionKey []byte
}

// New creates client with optional custom session storage
func New(config Config, providerType ProviderType) (*Client, error) {
    if config.SessionStorage == nil {
        // Default to cookie-based storage
        config.SessionStorage = NewCookieSessionStorage(config.SessionEncryptionKey)
    }
    // ... rest of implementation
}
```

#### **4.4 Session Interface Usage**
**Developers can implement custom storage:**
```go
// Example: Redis session storage
type RedisSessionStorage struct {
    client *redis.Client
    prefix string
}

func (r *RedisSessionStorage) Store(ctx context.Context, key string, session *SessionData) error {
    // Encrypt DPoP key, serialize session, store in Redis
}

// Use custom storage
client, err := atproto.New(atproto.Config{
    ClientID:       "...",
    SessionStorage: &RedisSessionStorage{client: redisClient},
}, atproto.ProviderTypeManual)
```

#### **4.5 Migration Strategy**
- **Backward Compatible**: Existing cookie-based API unchanged
- **Opt-in**: Developers can choose to use custom storage
- **Secure Defaults**: Cookie storage remains secure default
- **Easy Migration**: Simple interface to implement custom storage

#### **Benefits of Session Storage Interface**
✅ **Production Ready**: Support for Redis, database, distributed sessions  
✅ **Security**: Encrypted storage for sensitive DPoP keys  
✅ **Scalability**: Support for multi-instance applications  
✅ **Flexibility**: Easy to implement custom storage backends  
✅ **Backward Compatible**: No breaking changes to existing API  

---

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