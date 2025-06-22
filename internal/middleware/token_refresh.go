package middleware

import (
	"context"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/oauth"
	"golang.org/x/oauth2"
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

			// Convert TokenResult back to oauth2.Token for compatibility
			token := &oauth2.Token{
				AccessToken:  tokenResult.AccessToken,
				RefreshToken: tokenResult.RefreshToken,
			}

			// Update cookies with new tokens
			newRefreshTokens := []string{}
			if token.RefreshToken != "" {
				newRefreshTokens = append(newRefreshTokens, token.RefreshToken)
			}
			auth.SetSessionCookieWithEnv(w, token.AccessToken, newRefreshTokens, false)

			// Preserve the DPoP key from the refresh result
			if tokenResult.DPoPKey != nil {
				if err := auth.SetDPoPKeyCookie(w, tokenResult.DPoPKey, false); err != nil {
					logger.Error("Failed to set DPoP key cookie after refresh", "error", err)
				}
			}

			logger.Info("Token refresh successful, DPoP key preserved")

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