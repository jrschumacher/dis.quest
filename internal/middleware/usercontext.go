package middleware

import (
	"context"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/web"
	"github.com/jrschumacher/dis.quest/pkg/atproto/jwt"
	"github.com/jrschumacher/dis.quest/internal/logger"
)

// UserContext holds user information extracted from JWT
type UserContext struct {
	DID    string
	Handle string
	PDS    string
	Scope  string
	
	// Raw session data for web applications
	SessionData *web.SimpleSessionData
}

type contextKey string

const userContextKey contextKey = "user"

// UserContextMiddleware extracts user information from JWT and adds it to request context
func UserContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to load session data from cookies
		sessionData, err := web.LoadSimpleSessionFromCookies(r)
		if err != nil {
			// Fall back to basic token approach for backwards compatibility
			token, err := web.GetSessionCookie(r)
			if err != nil {
				// No token - continue without user context
				next.ServeHTTP(w, r)
				return
			}

			// Parse JWT to get claims (without verification for now in development)
			// TODO: In production, implement proper JWT verification with JWKS
			claims, err := jwt.ParseClaims(token)
			if err != nil {
				logger.Warn("Failed to parse JWT claims", "error", err)
				// Continue without user context rather than failing
				next.ServeHTTP(w, r)
				return
			}

			// Validate that we have the minimum required claims
			if claims.Subject == "" {
				logger.Warn("JWT missing required subject (DID)")
				next.ServeHTTP(w, r)
				return
			}

			// Create user context with basic JWT information
			userCtx := &UserContext{
				DID:   claims.Subject,
				PDS:   claims.Issuer,
				Scope: claims.Scope,
			}

			logger.Debug("User context created from JWT", "did", userCtx.DID, "pds", userCtx.PDS)

			// Add user context to request context
			ctx := context.WithValue(r.Context(), userContextKey, userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Create user context with session data
		claims, err := jwt.ParseClaims(sessionData.AccessToken)
		if err != nil {
			logger.Warn("Failed to parse JWT claims from session data", "error", err)
			// Continue without user context rather than failing
			next.ServeHTTP(w, r)
			return
		}

		// Create user context with session data
		userCtx := &UserContext{
			DID:         sessionData.UserDID,
			PDS:         claims.Issuer,
			Scope:       claims.Scope,
			SessionData: sessionData, // Include session data for web functionality
		}

		// Log user context creation for debugging
		logger.Debug("User context created with session data", 
			"did", userCtx.DID, 
			"pds", userCtx.PDS,
			"hasDPoPKey", userCtx.SessionData.DPoPKey != nil)

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

// GetSessionData extracts session data from user context (if available)
func GetSessionData(r *http.Request) (*web.SimpleSessionData, bool) {
	userCtx, ok := GetUserContext(r)
	if !ok || userCtx.SessionData == nil {
		return nil, false
	}
	return userCtx.SessionData, true
}
