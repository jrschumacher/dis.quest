package middleware

import (
	"context"
	"net/http"
)

// TestUserContextMiddleware creates a fake user context for testing
func TestUserContextMiddleware(userDID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a test user context
			userCtx := &UserContext{
				DID:   userDID,
				PDS:   "test-pds",
				Scope: "test-scope",
			}

			// Add user context to request context
			ctx := context.WithValue(r.Context(), userContextKey, userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TestAuthMiddleware bypasses authentication for testing
func TestAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication in tests
		next.ServeHTTP(w, r)
	})
}

// TestProtectedChain is like ProtectedChain but uses test middleware
func TestProtectedChain(userDID string) *Chain {
	return NewChain(
		TestAuthMiddleware,
		TestUserContextMiddleware(userDID),
	)
}