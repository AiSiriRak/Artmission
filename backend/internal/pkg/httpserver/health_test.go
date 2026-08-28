package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakePinger struct{ err error }

func (p fakePinger) PingContext(context.Context) error { return p.err }

func TestLivez_AlwaysOKRegardlessOfDependencies(t *testing.T) {
	mux := http.NewServeMux()
	registerHealthRoutes(mux, []Pinger{fakePinger{err: errors.New("db is down")}})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/livez", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("livez status = %d, want %d (must never depend on downstream health)", rec.Code, http.StatusOK)
	}
}

func TestReadyz_OKWhenAllDependenciesHealthy(t *testing.T) {
	mux := http.NewServeMux()
	registerHealthRoutes(mux, []Pinger{fakePinger{}, fakePinger{}})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("readyz status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestReadyz_UnavailableWhenADependencyFails(t *testing.T) {
	mux := http.NewServeMux()
	registerHealthRoutes(mux, []Pinger{fakePinger{}, fakePinger{err: errors.New("db is down")}})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("readyz status = %d, want %d when a dependency is unhealthy", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestReadyz_OKWithNoDependenciesConfigured(t *testing.T) {
	mux := http.NewServeMux()
	registerHealthRoutes(mux, nil)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("readyz status = %d, want %d", rec.Code, http.StatusOK)
	}
}
