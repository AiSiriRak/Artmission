// Package httpserver bootstraps the HTTP transport: the huma API bound to
// the stdlib net/http mux via humago, wrapped in domain-agnostic
// cross-cutting middleware (request id, logging, panic recovery, CORS).
// It has no knowledge of any domain module; route registration and
// domain-aware middleware (auth, role guards) live in
// internal/handler/rest.
package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

type Server struct {
	http *http.Server
	log  *slog.Logger
}

// New builds the huma API (registration happens by the caller against the
// returned API) and the underlying *http.Server.
func New(address, basePath string, allowedOrigins []string, log *slog.Logger, readiness []Pinger) (huma.API, *Server) {
	mux := http.NewServeMux()

	registerHealthRoutes(mux, readiness)

	config := huma.DefaultConfig("Artmission API", "0.1.0")
	config.DocsRenderer = huma.DocsRendererSwaggerUI
	config.Servers = []*huma.Server{{URL: basePath}}
	api := humago.NewWithPrefix(mux, basePath, config)

	handler := withRequestID(withLogging(log, withRecover(log, withCORS(allowedOrigins, mux))))

	return api, &Server{
		http: &http.Server{
			Addr:              address,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
		log: log,
	}
}

// Start runs the server until ctx is cancelled, then shuts down gracefully.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("listening", "address", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.log.Info("shutting down")
	return s.http.Shutdown(shutdownCtx)
}
