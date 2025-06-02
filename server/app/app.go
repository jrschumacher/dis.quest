package app

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/jrschumacher/dis.quest/components"
	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/middleware"
	"github.com/jrschumacher/dis.quest/internal/svrlib"
)

type AppRouter struct {
	*svrlib.Router
}

func RegisterRoutes(mux *http.ServeMux, _ string, cfg *config.Config) *AppRouter {
	router := &AppRouter{svrlib.NewRouter(mux, "/", cfg)}

	mux.Handle("/", templ.Handler(components.Page(cfg.AppEnv)))
	mux.Handle("/login", templ.Handler(components.Login()))
	mux.Handle("/discussion", middleware.AuthMiddleware(templ.Handler(components.Discussion())))

	return router
}
