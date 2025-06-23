package middleware

import (
	"net/http"
)

// RouteGroup represents a group of routes with common middleware
type RouteGroup struct {
	mux         *http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

// NewRouteGroup creates a new route group with optional middleware
func NewRouteGroup(mux *http.ServeMux, middlewares ...func(http.Handler) http.Handler) *RouteGroup {
	return &RouteGroup{
		mux:         mux,
		middlewares: middlewares,
	}
}

// Handle registers a handler with the group's middleware stack
func (rg *RouteGroup) Handle(pattern string, handler http.Handler) {
	rg.mux.Handle(pattern, ApplyFunc(handler.ServeHTTP, rg.middlewares...))
}

// HandleFunc registers a handler function with the group's middleware stack
func (rg *RouteGroup) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	rg.mux.Handle(pattern, ApplyFunc(handlerFunc, rg.middlewares...))
}

// Group creates a sub-group with additional middleware
func (rg *RouteGroup) Group(middlewares ...func(http.Handler) http.Handler) *RouteGroup {
	allMiddlewares := make([]func(http.Handler) http.Handler, len(rg.middlewares)+len(middlewares))
	copy(allMiddlewares, rg.middlewares)
	copy(allMiddlewares[len(rg.middlewares):], middlewares)
	
	return &RouteGroup{
		mux:         rg.mux,
		middlewares: allMiddlewares,
	}
}

// Common middleware stacks for convenience

// PublicGroup creates a route group for public routes with layout
func PublicGroup(mux *http.ServeMux, appEnv string) *RouteGroup {
	return NewRouteGroup(mux, LayoutMiddleware(appEnv))
}

// ProtectedPageGroup creates a route group for protected pages with auth + layout
func ProtectedPageGroup(mux *http.ServeMux, appEnv string) *RouteGroup {
	return NewRouteGroup(mux,
		AuthMiddleware,
		UserContextMiddleware,
		RequireUserContext,
		LayoutMiddleware(appEnv),
	)
}

// ProtectedAPIGroup creates a route group for protected API routes with auth only
func ProtectedAPIGroup(mux *http.ServeMux) *RouteGroup {
	return NewRouteGroup(mux,
		AuthMiddleware,
		UserContextMiddleware,
		RequireUserContext,
	)
}

// APIGroup creates a route group for API routes with user context (no auth required)
func APIGroup(mux *http.ServeMux) *RouteGroup {
	return NewRouteGroup(mux, UserContextMiddleware)
}

// RawGroup creates a route group with no middleware (for SSE, etc.)
func RawGroup(mux *http.ServeMux) *RouteGroup {
	return NewRouteGroup(mux)
}