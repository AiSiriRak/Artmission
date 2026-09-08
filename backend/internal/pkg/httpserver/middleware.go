package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

type requestIDKey struct{}

const requestIDHeader = "X-Request-Id"

// withRequestID reuses an inbound X-Request-Id (e.g. set by an upstream
// proxy/gateway) so logs correlate across hops, generating one only when
// absent. Echoes it back so client-side logs/support tickets can reference
// it.
func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(setRequestID(r.Context(), id)))
	})
}

// withLogging logs one line per request after it completes: method, path,
// status, duration, request id. Must wrap withRecover (not the reverse) so
// the status it reports reflects a panic-recovered 500 instead of whatever
// default was set before the panic — see withRecover.
func withLogging(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", requestID(r.Context()),
		)
	})
}

// withRecover converts a panic in a handler into a 500 instead of crashing
// the server, logging the panic and a stack trace for investigation. Must
// be wrapped by withLogging (not the reverse): only then does it receive
// withLogging's statusRecorder as its ResponseWriter, so WriteHeader(500)
// here is visible to the access log instead of being written to a
// different, unwrapped writer.
func withRecover(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if rec == http.ErrAbortHandler {
				// Sentinel for an intentional abort (e.g. client
				// disconnect); net/http expects this to keep propagating
				// unrecovered so it isn't logged as a real failure.
				panic(rec)
			}

			log.Error("panic recovered",
				"error", rec,
				"path", r.URL.Path,
				"request_id", requestID(r.Context()),
				"stack", string(debug.Stack()),
			)
			w.WriteHeader(http.StatusInternalServerError)
		}()
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func setRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func requestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
