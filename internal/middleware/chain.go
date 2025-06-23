// Package middleware provides HTTP middleware chain functionality
package middleware

import (
	"net/http"
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

// ApplyFunc is a shorthand for creating a chain and applying it to a handler function
func ApplyFunc(handlerFunc http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return NewChain(middlewares...).ThenFunc(handlerFunc)
}

