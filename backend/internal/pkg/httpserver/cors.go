package httpserver

import (
	"net/http"
	"slices"
)

// preflightMaxAge is how long a browser may cache a preflight response
// before re-checking, so most cross-origin requests skip the OPTIONS
// round-trip entirely.
const preflightMaxAge = "600" // 10 minutes, in seconds (spec unit).

// withCORS allows the configured origins to make credentialed requests
// (fetch with credentials: "include"), required for the httpOnly refresh
// cookie to be sent/received cross-origin by a separately-hosted frontend.
// Credentialed CORS cannot use a wildcard origin, so the request's Origin is
// echoed back only when it is allow-listed.
func withCORS(allowedOrigins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always vary by Origin, even when it's rejected below: without
		// this, a cache in front of the API (CDN, reverse proxy) could
		// serve a response meant for one origin to a different one.
		w.Header().Add("Vary", "Origin")

		origin := r.Header.Get("Origin")
		allowed := origin != "" && slices.Contains(allowedOrigins, origin)
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Only treat this as a CORS preflight if the browser actually
		// marked it as one; a bare OPTIONS request (e.g. a load balancer
		// health probe) should reach the router like any other request.
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			if allowed {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", preflightMaxAge)
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
