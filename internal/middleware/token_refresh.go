package middleware

import (
	"net/http"
	"time"

	"github.com/jrschumacher/dis.quest/internal/web"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/pkg/atproto"
	"github.com/jrschumacher/dis.quest/pkg/atproto/jwt"
)

// TokenRefreshMiddleware automatically refreshes access tokens when they're close to expiring
func TokenRefreshMiddleware(atprotoClient *atproto.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only refresh tokens for authenticated requests
			accessToken, err := web.GetSessionCookie(r)
			if err != nil {
				// No access token, skip refresh
				next.ServeHTTP(w, r)
				return
			}

			// Check if token is expiring within 5 minutes
			timeUntilExpiry, err := jwt.TimeUntilExpiry(accessToken)
			if err != nil || timeUntilExpiry > 5*time.Minute {
				// Token is still valid or we can't parse it, continue
				next.ServeHTTP(w, r)
				return
			}

			logger.Info("Access token expiring soon, attempting refresh")

			// Check if refresh token is available
			_, err = web.GetRefreshTokenCookie(r)
			if err != nil {
				logger.Error("No refresh token available", "error", err)
				// Clear session and redirect to login
				web.ClearSessionCookie(w)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// For now, redirect to login when token refresh is needed
			// TODO: Implement proper token refresh with ATProtocol session manager
			logger.Info("Token refresh needed, redirecting to login")
			web.ClearSessionCookie(w)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		})
	}
}

// AutoRefreshGroup creates a middleware group that includes token refresh
func AutoRefreshGroup(mux *http.ServeMux, atprotoClient *atproto.Client, middlewares ...func(http.Handler) http.Handler) *RouteGroup {
	// Prepend token refresh middleware to the list
	allMiddlewares := append([]func(http.Handler) http.Handler{TokenRefreshMiddleware(atprotoClient)}, middlewares...)
	return NewRouteGroup(mux, allMiddlewares...)
}