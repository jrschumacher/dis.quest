// Package health provides HTTP handlers for health check endpoints
package health

import (
	"fmt"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/svrlib"
)

// Router handles health check HTTP routes
type Router struct {
	*svrlib.Router
}

// RegisterRoutes registers all health check routes on the given mux
func RegisterRoutes(mux *http.ServeMux, baseRoute string, cfg *config.Config) {
	router := &Router{svrlib.NewRouter(mux, baseRoute, cfg)}
	mux.HandleFunc(baseRoute, router.HealthHandler)
}

// HealthHandler responds to /health requests for health checks
func (rt *Router) HealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintln(w, "ok")
}
