package svrlib

import (
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/config"
)

type Router struct {
	Config    *config.Config
	Mux       *http.ServeMux
	BaseRoute string
}

func NewRouter(mux *http.ServeMux, baseRoute string, cfg *config.Config) *Router {
	return &Router{cfg, mux, baseRoute}
}
