package middleware

import (
	"context"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/jwtutil"
	"github.com/jrschumacher/dis.quest/internal/logger"
)

// UserContext holds user information extracted from JWT
type UserContext struct {
	DID    string
	Handle string
	PDS    string
	Scope  string
}

type contextKey string

const userContextKey contextKey = "user"

// UserContextMiddleware extracts user information from JWT and adds it to request context
func UserContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the session token
		token, err := auth.GetSessionCookie(r)
		if err != nil {
			// No token - continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Parse JWT to get claims (without verification for now in development)
		// TODO: In production, implement proper JWT verification with JWKS
		claims, err := jwtutil.ParseJWTWithoutVerification(token)
		if err != nil {
			logger.Warn("Failed to parse JWT claims", "error", err)
			// Continue without user context rather than failing
			next.ServeHTTP(w, r)
			return
		}

		// Validate that we have the minimum required claims
		if claims.Sub == "" {
			logger.Warn("JWT missing required subject (DID)")
			next.ServeHTTP(w, r)
			return
		}

		// Create user context with available information
		userCtx := &UserContext{
			DID:   claims.Sub,
			PDS:   claims.Iss,
			Scope: claims.Scope,
		}

		// Log user context creation for debugging
		logger.Debug("User context created", "did", userCtx.DID, "pds", userCtx.PDS)

		// Add user context to request context
		ctx := context.WithValue(r.Context(), userContextKey, userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserContext extracts user context from request context
func GetUserContext(r *http.Request) (*UserContext, bool) {
	userCtx, ok := r.Context().Value(userContextKey).(*UserContext)
	return userCtx, ok
}

// RequireUserContext middleware that ensures user context exists
func RequireUserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userCtx, ok := GetUserContext(r)
		if !ok || userCtx.DID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
