package middleware

import (
	"context"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/oauth"
	"github.com/jrschumacher/dis.quest/pkg/atproto"
)

// TokenRefreshMiddleware automatically refreshes access tokens when they're close to expiring
func TokenRefreshMiddleware(oauthService *oauth.Service) func(http.Handler) http.Handler {
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

			// Get refresh token
			refreshToken, err := auth.GetRefreshTokenCookie(r)
			if err != nil {
				logger.Error("No refresh token available", "error", err)
				// Clear session and redirect to login
				auth.ClearSessionCookie(w)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Perform token refresh
			ctx := context.WithValue(r.Context(), "http_request", r)
			tokenResult, err := oauthService.RefreshToken(ctx, refreshToken)
			if err != nil {
				logger.Error("Token refresh failed", "error", err)
				// Clear session and redirect to login
				auth.ClearSessionCookie(w)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Create enhanced session wrapper with refreshed tokens
			sessionWrapper, err := auth.NewSessionWrapper(
				tokenResult.AccessToken,
				tokenResult.RefreshToken,
				tokenResult.UserDID,
				tokenResult.DPoPKey,
				nil, // atproto client
			)
			if err != nil {
				logger.Error("Failed to create session wrapper after refresh", "error", err)
				// Fall back to clearing session
				auth.ClearSessionCookie(w)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// Set the atproto session if available from token result
			if tokenResult.AtprotoSession != nil {
				if atprotoSession, ok := tokenResult.AtprotoSession.(*atproto.Session); ok {
					sessionWrapper.SetAtprotoSession(atprotoSession)
				}
				logger.Info("Enhanced session refreshed with atproto.Session", "userDID", tokenResult.UserDID)
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
func AutoRefreshGroup(mux *http.ServeMux, oauthService *oauth.Service, middlewares ...func(http.Handler) http.Handler) *RouteGroup {
	// Prepend token refresh middleware to the list
	allMiddlewares := append([]func(http.Handler) http.Handler{TokenRefreshMiddleware(oauthService)}, middlewares...)
	return NewRouteGroup(mux, allMiddlewares...)
}