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
	DID            string
	Handle         string
	PDS            string
	Scope          string
	SessionWrapper *auth.SessionWrapper // Enhanced: Include session wrapper when available
}

type contextKey string

const userContextKey contextKey = "user"

// UserContextMiddleware extracts user information from JWT and adds it to request context
func UserContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to load session wrapper from cookies first (enhanced approach)
		sessionWrapper, err := auth.LoadSessionFromCookies(r)
		if err != nil {
			// Fall back to basic token approach for backwards compatibility
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

			// Create user context with basic JWT information
			userCtx := &UserContext{
				DID:   claims.Sub,
				PDS:   claims.Iss,
				Scope: claims.Scope,
			}

			logger.Debug("User context created from JWT", "did", userCtx.DID, "pds", userCtx.PDS)

			// Add user context to request context
			ctx := context.WithValue(r.Context(), userContextKey, userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Enhanced: Create user context with session wrapper
		claims, err := jwtutil.ParseJWTWithoutVerification(sessionWrapper.GetAccessToken())
		if err != nil {
			logger.Warn("Failed to parse JWT claims from session wrapper", "error", err)
			// Continue without user context rather than failing
			next.ServeHTTP(w, r)
			return
		}

		// Create enhanced user context
		userCtx := &UserContext{
			DID:            sessionWrapper.GetUserDID(),
			PDS:            claims.Iss,
			Scope:          claims.Scope,
			SessionWrapper: sessionWrapper, // Include session wrapper for enhanced functionality
		}

		// Log enhanced user context creation for debugging
		logger.Debug("Enhanced user context created with session wrapper", 
			"did", userCtx.DID, 
			"pds", userCtx.PDS, 
			"hasAtprotoSession", userCtx.SessionWrapper.GetAtprotoSession() != nil)

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

// GetSessionWrapper extracts session wrapper from user context (if available)
func GetSessionWrapper(r *http.Request) (*auth.SessionWrapper, bool) {
	userCtx, ok := GetUserContext(r)
	if !ok || userCtx.SessionWrapper == nil {
		return nil, false
	}
	return userCtx.SessionWrapper, true
}

// GetAtprotoSession extracts atproto.Session from user context (if available)
func GetAtprotoSession(r *http.Request) (*auth.SessionWrapper, bool) {
	sessionWrapper, ok := GetSessionWrapper(r)
	if !ok || sessionWrapper.GetAtprotoSession() == nil {
		return nil, false
	}
	return sessionWrapper, true
}
