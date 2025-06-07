# Middleware Patterns Guide

This document shows the clean middleware patterns available in dis.quest.

## Quick Reference

### Before (Nested/Messy)
```go
// This gets messy quickly
mux.Handle("/api/data", middleware1(middleware2(middleware3(handler))))
```

### After (Clean Chains)
```go
// Much cleaner and readable
mux.Handle("/api/data", 
    middleware.WithMiddleware(
        middleware1,
        middleware2, 
        middleware3,
    ).ThenFunc(handler))
```

## Common Patterns

### 1. Predefined Chains (Recommended)

```go
// Public routes (no authentication)
mux.Handle("/", middleware.PublicChain.ThenFunc(homeHandler))

// Authentication required
mux.Handle("/dashboard", middleware.AuthenticatedChain.ThenFunc(dashboardHandler))

// Full protection (auth + user context + validation)
mux.Handle("/admin", middleware.ProtectedChain.ThenFunc(adminHandler))
```

### 2. Helper Functions (Quick & Simple)

```go
// Quick authentication wrapper
mux.Handle("/profile", middleware.WithAuthFunc(profileHandler))

// Full protection wrapper
mux.Handle("/settings", middleware.WithProtectionFunc(settingsHandler))

// Just user context (optional auth)
mux.Handle("/public-profile", middleware.WithUserContextFunc(publicProfileHandler))
```

### 3. Custom Chains (Reusable)

```go
// Define once, use many times
apiChain := middleware.NewChain(
    loggingMiddleware,
    corsMiddleware,
    middleware.UserContextMiddleware,
)

// Apply to multiple routes
mux.Handle("/api/users", apiChain.ThenFunc(usersHandler))
mux.Handle("/api/posts", apiChain.ThenFunc(postsHandler))
mux.Handle("/api/comments", apiChain.ThenFunc(commentsHandler))
```

### 4. Composable Chains (Advanced)

```go
// Build base chain
baseAPI := middleware.NewChain(loggingMiddleware, corsMiddleware)

// Extend for different security levels
publicAPI := baseAPI.Append(middleware.UserContextMiddleware)
protectedAPI := baseAPI.Append(middleware.AuthMiddleware, middleware.UserContextMiddleware)
adminAPI := protectedAPI.Append(adminMiddleware)

// Use throughout the application
mux.Handle("/api/public/stats", publicAPI.ThenFunc(statsHandler))
mux.Handle("/api/user/profile", protectedAPI.ThenFunc(profileHandler))
mux.Handle("/api/admin/users", adminAPI.ThenFunc(adminUsersHandler))
```

### 5. Inline Chains (Flexible)

```go
// For one-off routes with specific needs
mux.Handle("/api/special", 
    middleware.WithMiddleware(
        rateLimitMiddleware,
        middleware.AuthMiddleware,
        specialValidationMiddleware,
    ).ThenFunc(specialHandler))
```

## Route Organization Examples

### App Routes
```go
// server/app/app.go
func RegisterRoutes(mux *http.ServeMux, prefix string, cfg *config.Config, dbService *db.Service) {
    // Public pages
    mux.Handle("/", middleware.PublicChain.ThenFunc(homeHandler))
    mux.Handle("/about", middleware.PublicChain.ThenFunc(aboutHandler))
    
    // User pages (require authentication)
    mux.Handle("/dashboard", middleware.WithProtectionFunc(dashboardHandler))
    mux.Handle("/settings", middleware.WithProtectionFunc(settingsHandler))
    
    // API routes with consistent middleware
    apiChain := middleware.NewChain(
        corsMiddleware,
        jsonMiddleware,
        middleware.UserContextMiddleware,
    )
    
    mux.Handle("/api/topics", apiChain.ThenFunc(topicsHandler))
    mux.Handle("/api/messages", apiChain.ThenFunc(messagesHandler))
}
```

### Auth Routes
```go
// server/auth-handlers/auth.go
func RegisterRoutes(mux *http.ServeMux, prefix string, cfg *config.Config) {
    // Auth endpoints are typically public (pre-authentication)
    mux.Handle(prefix+"/login", middleware.PublicChain.ThenFunc(loginHandler))
    mux.Handle(prefix+"/callback", middleware.PublicChain.ThenFunc(callbackHandler))
    mux.Handle(prefix+"/logout", middleware.WithAuthFunc(logoutHandler)) // Requires existing session
}
```

## Available Middleware

### Core Middleware
- `AuthMiddleware` - Validates session cookie, redirects if missing
- `UserContextMiddleware` - Extracts user info from JWT (optional auth)
- `RequireUserContext` - Ensures user context exists, redirects if not

### Predefined Chains
- `PublicChain` - No middleware (public access)
- `AuthenticatedChain` - Just authentication required
- `UserContextChain` - User context extraction (auth optional)
- `ProtectedChain` - Full protection (auth + user context + validation)

### Helper Functions
- `WithAuth(handler)` - Wrap with authentication
- `WithUserContext(handler)` - Wrap with user context
- `WithProtection(handler)` - Wrap with full protection
- `WithMiddleware(...).Then(handler)` - Custom middleware chain

## Benefits

1. **Readable**: Clear intent, easy to understand
2. **Composable**: Build complex chains from simple parts
3. **Reusable**: Define once, use everywhere
4. **Maintainable**: Easy to add/remove middleware
5. **Type-Safe**: Compile-time checking
6. **Performance**: No runtime overhead vs manual nesting

## Migration Guide

### Old Pattern
```go
mux.Handle("/protected", 
    middleware.AuthMiddleware(
        middleware.UserContextMiddleware(
            middleware.RequireUserContext(
                http.HandlerFunc(handler)))))
```

### New Pattern
```go
mux.Handle("/protected", middleware.WithProtectionFunc(handler))
```

The new patterns are not only cleaner but also more maintainable and easier to test.