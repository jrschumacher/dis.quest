// Package svrlib provides common server routing utilities
package svrlib

import (
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/config"
)

// Router wraps HTTP routing functionality with configuration
type Router struct {
	Config    *config.Config
	Mux       *http.ServeMux
	BaseRoute string
}

// NewRouter creates a new Router with the given mux, base route, and configuration
func NewRouter(mux *http.ServeMux, baseRoute string, cfg *config.Config) *Router {
	return &Router{cfg, mux, baseRoute}
}
