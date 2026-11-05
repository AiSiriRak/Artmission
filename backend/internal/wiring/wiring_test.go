package wiring

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/objectstorage"
)

func TestWalletReadsAreAuthenticatedAndManualWalletRoutesAreRemoved(t *testing.T) {
	for _, production := range []bool{true, false} {
		server := Wire(Config{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), ObjectStorage: &objectstorage.Client{}, App: config.App{BasePath: "/api/v1", IsProduction: production}})
		for _, endpoint := range []struct{ method, path string }{
			{"GET", "/mock-wallet/me"}, {"POST", "/mock-wallet/top-ups"},
			{"POST", "/orders/00000000-0000-0000-0000-000000000001/payment/mock-wallet"},
			{"POST", "/mock-wallet/transfers/00000000-0000-0000-0000-000000000001/execute"},
		} {
			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, httptest.NewRequest(endpoint.method, "/api/v1"+endpoint.path, nil))
			want := 404
			if rec.Code != want {
				t.Fatalf("production=%t %s: %d want %d", production, endpoint.path, rec.Code, want)
			}
		}
		for _, path := range []string{"/wallet/me", "/transaction/me"} {
			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1"+path, nil))
			if rec.Code != 401 {
				t.Fatalf("production=%t %s: %d want 401", production, path, rec.Code)
			}
		}
	}
}
