package health

import (
	"fmt"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/svrlib"
)

type HealthRouter struct {
	*svrlib.Router
}

// RegisterRoutes registers all health check routes on the given mux
func RegisterRoutes(mux *http.ServeMux, baseRoute string, cfg *config.Config) {
	router := &HealthRouter{svrlib.NewRouter(mux, baseRoute, cfg)}
	mux.HandleFunc(baseRoute, router.HealthHandler)
}

// HealthHandler responds to /health requests for health checks
func (rt *HealthRouter) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "ok")
}
