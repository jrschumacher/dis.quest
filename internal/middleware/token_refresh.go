package middleware

import (
	"context"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/pkg/atproto"
)

// TokenRefreshMiddleware automatically refreshes access tokens when they're close to expiring
func TokenRefreshMiddleware(atprotoClient *atproto.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only refresh tokens for authenticated requests
			accessToken, err := auth.GetSessionCookie(r)
			if err != nil {
				// No access token, skip refresh
				next.ServeHTTP(w, r)
				return
			}

			// Check if token is expiring within 5 minutes
			if !auth.IsTokenExpiringSoon(accessToken, 5) {
				// Token is still valid, continue
				next.ServeHTTP(w, r)
				return
			}

			logger.Info("Access token expiring soon, attempting refresh")

			// Check if refresh token is available
			_, err = auth.GetRefreshTokenCookie(r)
			if err != nil {
				logger.Error("No refresh token available", "error", err)
				// Clear session and redirect to login
				auth.ClearSessionCookie(w)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Load the current session and attempt to refresh it directly
			sessionWrapper, err := auth.LoadSessionFromCookies(r)
			if err != nil {
				logger.Error("Failed to load session for refresh", "error", err)
				auth.ClearSessionCookie(w)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// If we have an atproto session, try to refresh it
			if atprotoSession := sessionWrapper.GetAtprotoSession(); atprotoSession != nil {
				ctx := context.WithValue(r.Context(), "http_request", r)
				if err := atprotoSession.Refresh(ctx); err != nil {
					logger.Error("Failed to refresh atproto session", "error", err)
					auth.ClearSessionCookie(w)
					http.Redirect(w, r, "/login", http.StatusFound)
					return
				}
				logger.Info("ATProtocol session refreshed", "userDID", atprotoSession.GetUserDID())
			} else {
				// Legacy refresh not supported - redirect to login
				logger.Warn("Legacy session refresh not supported, redirecting to login")
				auth.ClearSessionCookie(w)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Save refreshed session to cookies using the wrapper
			if err := sessionWrapper.SaveToCookies(w, false); err != nil {
				logger.Error("Failed to save refreshed session to cookies", "error", err)
				// Fall back to clearing session
				auth.ClearSessionCookie(w)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			logger.Info("Token refresh successful with enhanced session wrapper")

			// Continue with the refreshed token
			next.ServeHTTP(w, r)
		})
	}
}

// AutoRefreshGroup creates a middleware group that includes token refresh
func AutoRefreshGroup(mux *http.ServeMux, atprotoClient *atproto.Client, middlewares ...func(http.Handler) http.Handler) *RouteGroup {
	// Prepend token refresh middleware to the list
	allMiddlewares := append([]func(http.Handler) http.Handler{TokenRefreshMiddleware(atprotoClient)}, middlewares...)
	return NewRouteGroup(mux, allMiddlewares...)
}