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
	mux.HandleFunc(baseRoute+"/healthz", router.HealthzHandler)
}

// HealthzHandler responds to /healthz requests for health checks
func (rt *HealthRouter) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "ok")
}
