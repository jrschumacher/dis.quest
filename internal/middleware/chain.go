// Package middleware provides HTTP middleware chain functionality
package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	"github.com/jrschumacher/dis.quest/components"
)

// Chain represents a middleware chain that can be applied to handlers
type Chain struct {
	middlewares []func(http.Handler) http.Handler
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...func(http.Handler) http.Handler) *Chain {
	return &Chain{
		middlewares: append([]func(http.Handler) http.Handler(nil), middlewares...),
	}
}

// Then applies the middleware chain to a handler
func (c *Chain) Then(handler http.Handler) http.Handler {
	if handler == nil {
		handler = http.DefaultServeMux
	}

	// Apply middlewares in reverse order so they execute in the order specified
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}

	return handler
}

// ThenFunc applies the middleware chain to a handler function
func (c *Chain) ThenFunc(handlerFunc http.HandlerFunc) http.Handler {
	return c.Then(handlerFunc)
}

// Append adds middlewares to the end of the chain
func (c *Chain) Append(middlewares ...func(http.Handler) http.Handler) *Chain {
	newMiddlewares := make([]func(http.Handler) http.Handler, len(c.middlewares)+len(middlewares))
	copy(newMiddlewares, c.middlewares)
	copy(newMiddlewares[len(c.middlewares):], middlewares)

	return &Chain{middlewares: newMiddlewares}
}

// Prepend adds middlewares to the beginning of the chain
func (c *Chain) Prepend(middlewares ...func(http.Handler) http.Handler) *Chain {
	newMiddlewares := make([]func(http.Handler) http.Handler, len(middlewares)+len(c.middlewares))
	copy(newMiddlewares, middlewares)
	copy(newMiddlewares[len(middlewares):], c.middlewares)

	return &Chain{middlewares: newMiddlewares}
}

// Common middleware chains for reuse
var (
	// PublicChain is for public routes that don't require authentication
	PublicChain = NewChain()

	// AuthenticatedChain is for routes that require authentication but not user context
	AuthenticatedChain = NewChain(AuthMiddleware)

	// UserContextChain is for routes that need user context but authentication is optional
	UserContextChain = NewChain(UserContextMiddleware)

	// ProtectedChain is for routes that require both authentication and user context
	ProtectedChain = NewChain(AuthMiddleware, UserContextMiddleware, RequireUserContext)
)

// Helper functions for common middleware combinations

// WithAuth wraps a handler with authentication middleware
func WithAuth(handler http.Handler) http.Handler {
	return AuthenticatedChain.Then(handler)
}

// WithAuthFunc wraps a handler function with authentication middleware
func WithAuthFunc(handlerFunc http.HandlerFunc) http.Handler {
	return AuthenticatedChain.ThenFunc(handlerFunc)
}

// WithUserContext wraps a handler with user context middleware
func WithUserContext(handler http.Handler) http.Handler {
	return UserContextChain.Then(handler)
}

// WithUserContextFunc wraps a handler function with user context middleware
func WithUserContextFunc(handlerFunc http.HandlerFunc) http.Handler {
	return UserContextChain.ThenFunc(handlerFunc)
}

// WithProtection wraps a handler with full authentication and user context
func WithProtection(handler http.Handler) http.Handler {
	return ProtectedChain.Then(handler)
}

// WithProtectionFunc wraps a handler function with full authentication and user context
func WithProtectionFunc(handlerFunc http.HandlerFunc) http.Handler {
	return ProtectedChain.ThenFunc(handlerFunc)
}

// Custom chains can be built for specific needs

// WithMiddleware creates a new chain with the specified middlewares
func WithMiddleware(middlewares ...func(http.Handler) http.Handler) *Chain {
	return NewChain(middlewares...)
}

// Apply is a shorthand for creating a chain and applying it to a handler
func Apply(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return NewChain(middlewares...).Then(handler)
}

// ApplyFunc is a shorthand for creating a chain and applying it to a handler function
func ApplyFunc(handlerFunc http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return NewChain(middlewares...).ThenFunc(handlerFunc)
}

// PageWrapper returns a middleware that wraps the handler's HTML output in components.Page.
func PageWrapper(appEnv string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := httptest.NewRecorder()
			next.ServeHTTP(rw, r)

			content := templ.ComponentFunc(func(ctx context.Context, wtr io.Writer) error {
				_, err := wtr.Write(rw.Body.Bytes())
				return err
			})

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := components.Page(appEnv, content).Render(r.Context(), w); err != nil {
				http.Error(w, "Failed to render page", http.StatusInternalServerError)
			}
		})
	}
}
