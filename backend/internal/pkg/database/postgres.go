// Package database wires up the bun.DB Postgres connection used by every
// repository adapter.
package database

import (
	"context"

	"github.com/DeepAung/artmission/backend/internal/pkg/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// NewPostgresDB opens a pgx connection pool wrapped as a bun.DB.
func NewPostgresDB(cfg config.Database) (*bun.DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}

	// Bun builds SQL dynamically per call, so it doesn't produce the
	// stable, repeated query text pgx's default implicit prepared
	// statements need to pay off. Bun's own docs recommend simple
	// protocol for exactly this reason when pairing bun with pgx:
	// https://bun.uptrace.dev/postgres/#pgx
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	sqldb := stdlib.OpenDBFromPool(pool)
	return bun.NewDB(sqldb, pgdialect.New()), nil
}
