package health

import (
	"fmt"
	"net/http"

	"github.com/jrschumacher/dis.quest/internal/config"
)

// RegisterRoutes registers all health check routes on the given mux
func RegisterRoutes(mux *http.ServeMux, prefix string, cfg *config.Config) {
	mux.HandleFunc(prefix+"/healthz", HealthzHandler)
}

// HealthzHandler responds to /healthz requests for health checks
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "ok")
}
