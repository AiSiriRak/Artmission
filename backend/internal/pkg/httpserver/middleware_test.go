package httpserver

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, nil))
}

// TestPanicIsLoggedWithRecoveredStatus is a regression test for the bug
// where withRecover sitting outside withLogging caused a panic-recovered
// 500 response to be access-logged as 200: withRecover would write the 500
// to the raw ResponseWriter instead of withLogging's statusRecorder. The
// chain must be withLogging(withRecover(...)), matching how it's built in
// New().
func TestPanicIsLoggedWithRecoveredStatus(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)

	panics := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	handler := withRequestID(withLogging(log, withRecover(log, panics)))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("response status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var sawAccessLog bool
	for _, line := range bytes.Split(buf.Bytes(), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal(line, &entry); err != nil {
			t.Fatalf("failed to parse log line %q: %v", line, err)
		}
		if entry["msg"] != "request" {
			continue
		}
		sawAccessLog = true
		if status, _ := entry["status"].(float64); int(status) != http.StatusInternalServerError {
			t.Errorf("access log status = %v, want %d (panic recovery must be visible to the access log)", entry["status"], http.StatusInternalServerError)
		}
	}
	if !sawAccessLog {
		t.Fatal("no \"request\" access log line was emitted")
	}
}

func TestRequestID_ReusesInboundHeader(t *testing.T) {
	var gotID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = requestID(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestIDHeader, "upstream-id-123")
	rec := httptest.NewRecorder()
	withRequestID(next).ServeHTTP(rec, req)

	if gotID != "upstream-id-123" {
		t.Errorf("request id = %q, want the inbound header value reused, not a freshly generated one", gotID)
	}
	if got := rec.Header().Get(requestIDHeader); got != "upstream-id-123" {
		t.Errorf("response header %s = %q, want it echoed back", requestIDHeader, got)
	}
}

func TestRequestID_GeneratesWhenAbsent(t *testing.T) {
	rec := httptest.NewRecorder()
	withRequestID(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Header().Get(requestIDHeader) == "" {
		t.Error("expected a generated request id when none was supplied")
	}
}

func TestCORS_VaryIsAlwaysSetEvenForDisallowedOrigin(t *testing.T) {
	handler := withCORS([]string{"https://allowed.example.com"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary header = %q, want \"Origin\" even for a disallowed origin", got)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("disallowed origin must not get Access-Control-Allow-Origin")
	}
}

func TestCORS_PreflightSetsMaxAgeAndAllowHeaders(t *testing.T) {
	handler := withCORS([]string{"https://allowed.example.com"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight must not reach the wrapped handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://allowed.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Header().Get("Access-Control-Max-Age") == "" {
		t.Error("expected Access-Control-Max-Age on a successful preflight")
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "https://allowed.example.com" {
		t.Error("expected the allowed origin echoed back")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, http.MethodPut) {
		t.Errorf("Access-Control-Allow-Methods = %q, want it to include %q", got, http.MethodPut)
	}
}

func TestCORS_BareOptionsWithoutPreflightHeaderReachesHandler(t *testing.T) {
	var reached bool
	handler := withCORS([]string{"https://allowed.example.com"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
	}))

	// No Access-Control-Request-Method: not a CORS preflight (e.g. a load
	// balancer health probe), must not be short-circuited.
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !reached {
		t.Error("a bare OPTIONS request without Access-Control-Request-Method must reach the wrapped handler")
	}
}
