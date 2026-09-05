//go:build integration

package apptest

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/logger"
	"github.com/AiSiriRak/Artmission/backend/internal/wiring"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Fixed test-only settings — never sourced from .env/config.yaml, so this
// suite can never accidentally point at (or read secrets meant for) a real
// environment.
const (
	basePath        = "/api/v1"
	jwtSecret       = "bdd-test-secret-never-used-outside-tests"
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 24 * time.Hour
)

// App is the real Artmission HTTP application — every adapter, usecase and
// handler wired through internal/wiring.Wire, exactly as cmd/cmd_serve.go
// does it, against a real (containerized) Postgres — served in-process via
// httptest.Server. No mocks, no fakes, no cobra/signal-handling glue.
type App struct {
	Server *httptest.Server

	// DB is a direct database handle — an escape hatch for seeding
	// fixture data no HTTP endpoint can create yet (e.g. orders; see
	// internal/modules/order/port.go). Prefer the real HTTP API wherever
	// it can express the precondition; reach for DB only when it can't.
	DB *bun.DB
}

// NewApp builds the app against dsn and starts it. Call once per test
// package (from TestXxx) against a freshly migrated database; every
// scenario in that package shares this one running instance.
func NewApp(tb testing.TB, dsn string) *App {
	tb.Helper()

	db, err := database.NewPostgresDB(config.Database{DSN: dsn})
	if err != nil {
		tb.Fatalf("apptest: connect to postgres: %v", err)
	}
	tb.Cleanup(func() { _ = db.Close() })

	server := wiring.Wire(wiring.Config{
		DB:     db,
		Logger: logger.NewLogger(false),
		App: config.App{
			// Never dialed — Start() is never called — so it need not be
			// a real, free port.
			Address:  "127.0.0.1:0",
			BasePath: basePath,
			// AllowedOrigins left nil: CORS is untestable via httptest
			// (no browser origin), so no scenario can exercise it.
			AllowedOrigins: nil,
			// IsProduction false: httptest serves plain http://, so a
			// Secure cookie set by the auth handler would never round-trip
			// back to this client.
			IsProduction: false,
		},
		Auth: config.Auth{
			JWTSecret:       jwtSecret,
			AccessTokenTTL:  accessTokenTTL,
			RefreshTokenTTL: refreshTokenTTL,
			// RefreshCookieDomain empty: httptest's Server.URL is a
			// bare 127.0.0.1 host, which can't carry a cookie Domain.
			RefreshCookieDomain: "",
		},
	})

	ts := httptest.NewServer(server.Handler())
	tb.Cleanup(ts.Close)

	return &App{Server: ts, DB: db}
}

// UserIDByEmail looks up a user's ID by email directly against the
// database. Registration responses carry no body (see RegisterOutput), so
// this is how a fixture that just registered an account gets its ID back.
// Looks up by email rather than username since username is no longer
// unique — it can collide across accounts.
func (a *App) UserIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	var id uuid.UUID
	err := a.DB.NewSelect().Table("users").Column("id").Where("email = ?", email).Scan(ctx, &id)
	return id, err
}

// BaseURL is the versioned API root, e.g. http://127.0.0.1:54321/api/v1.
func (a *App) BaseURL() string {
	return a.Server.URL + basePath
}
