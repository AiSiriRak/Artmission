package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger is a dependency readyz checks before reporting the instance ready
// for traffic. *bun.DB satisfies this directly via its embedded *sql.DB
// (no adapter needed). Add further Pingers (object storage, etc.) as those
// dependencies land.
type Pinger interface {
	PingContext(ctx context.Context) error
}

const readinessCheckTimeout = 3 * time.Second

// registerHealthRoutes attaches /livez and /readyz directly to mux,
// unprefixed and outside huma/OpenAPI: they're infra plumbing (Cloud Run
// startup/liveness probes, uptime checks), not part of the versioned API.
//
// livez only confirms the HTTP server itself is responsive — it never
// checks dependencies, so a transient DB outage doesn't make an
// orchestrator kill and restart a perfectly healthy process. readyz checks
// every dependency in readiness and reports whether this instance can
// actually serve real traffic.
func registerHealthRoutes(mux *http.ServeMux, readiness []Pinger) {
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) {
		writeHealthStatus(w, http.StatusOK, "ok", "")
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readinessCheckTimeout)
		defer cancel()

		for _, p := range readiness {
			if err := p.PingContext(ctx); err != nil {
				writeHealthStatus(w, http.StatusServiceUnavailable, "unavailable", err.Error())
				return
			}
		}
		writeHealthStatus(w, http.StatusOK, "ok", "")
	})
}

func writeHealthStatus(w http.ResponseWriter, status int, statusText, errText string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	}{Status: statusText, Error: errText})
}
