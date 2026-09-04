//go:build integration

// Package apptest builds the real Artmission backend (real Postgres, real
// HTTP handler stack, zero mocks) for BDD tests to drive over HTTP. It is
// the one shared seam every domain's godog suite (auth today, others
// later) imports — container lifecycle, app wiring, and an HTTP client
// live here so a domain test package only holds features/steps/fixtures.
package apptest

import (
	"context"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	appmigrations "github.com/AiSiriRak/Artmission/backend/internal/pkg/migrations"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// postgresImage matches the version pinned in backend/docker-compose.yaml
// so the BDD suite exercises the same engine as local dev/production.
const postgresImage = "postgres:16-alpine"

// Postgres is a disposable, migrated Postgres instance backing one test
// binary run. It never touches the developer's docker-compose database.
type Postgres struct {
	DSN string
}

// StartPostgres launches a fresh container, applies every goose migration
// from internal/pkg/migrations, and registers container teardown on tb.
// Call it once per test package (from TestXxx), not once per scenario —
// container startup dominates otherwise.
func StartPostgres(ctx context.Context, tb testing.TB) *Postgres {
	tb.Helper()

	container, err := postgres.Run(ctx, postgresImage,
		postgres.WithDatabase("artmission_test"),
		postgres.WithUsername("artmission_test"),
		postgres.WithPassword("artmission_test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		tb.Fatalf("apptest: start postgres container: %v", err)
	}
	tb.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			tb.Logf("apptest: terminate postgres container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		tb.Fatalf("apptest: read postgres connection string: %v", err)
	}

	if err := migrateUp(dsn); err != nil {
		tb.Fatalf("apptest: migrate up: %v", err)
	}

	return &Postgres{DSN: dsn}
}

// migrateUp applies every migration in internal/pkg/migrations, the exact
// same embed.FS cmd/cmd_serve.go applies on real startup.
func migrateUp(dsn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	db, err := database.NewPostgresDB(config.Database{DSN: dsn})
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(appmigrations.Migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, db.DB, ".")
}
