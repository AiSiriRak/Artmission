//go:build integration

package apptest

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres"
	"github.com/AiSiriRak/Artmission/backend/internal/adapters/token"
	"github.com/AiSiriRak/Artmission/backend/internal/handler/rest"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/artist"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/logger"
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
// handler wired exactly as cmd/cmd_serve.go does it, against a real
// (containerized) Postgres — served in-process via httptest.Server. No
// mocks, no fakes, no cobra/signal-handling glue.
type App struct {
	Server *httptest.Server

	// DB is a direct database handle — an escape hatch for seeding
	// fixture data no HTTP endpoint can create yet (e.g. orders; see
	// internal/modules/order/port.go). Prefer the real HTTP API wherever
	// it can express the precondition; reach for DB only when it can't.
	DB *bun.DB
}

// NewApp builds the app against dsn and starts it. Call once per test
// package (from TestMain) against a freshly migrated database; every
// scenario in that package shares this one running instance.
func NewApp(tb testing.TB, dsn string) *App {
	tb.Helper()

	db, err := database.NewPostgresDB(config.Database{DSN: dsn})
	if err != nil {
		tb.Fatalf("apptest: connect to postgres: %v", err)
	}
	tb.Cleanup(func() { _ = db.Close() })

	userRepo := postgres.NewUserRepository(db)
	bankRepo := postgres.NewBankAccountRepository(db)
	artistRepo := postgres.NewArtistRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	tokenIssuer := token.NewJWTIssuer(jwtSecret)
	tx := baserepo.NewTransactioner(db)

	artistUsecase := artist.NewProfileUsecase(artistRepo)
	userUsecase := user.NewUserUsecase(userRepo, bankRepo, artistUsecase, tx)
	authUsecase := auth.NewAuthUsecase(userUsecase, sessionRepo, tokenIssuer, accessTokenTTL, refreshTokenTTL)
	orderUsecase := order.NewOrderUsecase(orderRepo)

	authHandler := rest.NewAuthHandler(userUsecase, authUsecase, basePath, false, "")
	orderHandler := rest.NewOrderHandler(orderUsecase, authUsecase)

	// Address is never dialed — Start() is never called — so it need not
	// be a real, free port.
	api, srv := httpserver.New("127.0.0.1:0", basePath, nil, logger.NewLogger(false), []httpserver.Pinger{db})
	rest.RegisterRoutes(api, authHandler, orderHandler)

	ts := httptest.NewServer(srv.Handler())
	tb.Cleanup(ts.Close)

	return &App{Server: ts, DB: db}
}

// UserIDByUsername looks up a user's ID by username directly against the
// database. Registration responses carry no body (see RegisterOutput), so
// this is how a fixture that just registered an account gets its ID back.
func (a *App) UserIDByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	var id uuid.UUID
	err := a.DB.NewSelect().Table("users").Column("id").Where("username = ?", username).Scan(ctx, &id)
	return id, err
}

// BaseURL is the versioned API root, e.g. http://127.0.0.1:54321/api/v1.
func (a *App) BaseURL() string {
	return a.Server.URL + basePath
}
