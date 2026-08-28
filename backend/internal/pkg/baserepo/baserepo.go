// Package baserepo deduplicates the CRUD boilerplate every adapter
// repository would otherwise repeat.
package baserepo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DeepAung/artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// BaseRepo provides generic CRUD for a single bun model M keyed by a `uuid`
// primary key column named "id". Adapters embed or wrap it and add
// domain-specific queries (lookups by other columns, listings, etc.)
// alongside it.
type BaseRepo[M any] interface {
	Create(ctx context.Context, model *M) error
	FindByID(ctx context.Context, id uuid.UUID) (*M, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateByID(ctx context.Context, id uuid.UUID, model *M) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type baseRepo[M any] struct {
	executor Executor
	name     string
}

// NewBaseRepo builds a BaseRepo. name is used in "<name> not found" errors.
func NewBaseRepo[M any](db *bun.DB, name string) BaseRepo[M] {
	return &baseRepo[M]{executor: NewExecutor(db), name: name}
}

func (b *baseRepo[M]) Create(ctx context.Context, model *M) error {
	return b.executor.Run(ctx, func(idb bun.IDB) error {
		_, err := idb.NewInsert().Model(model).Exec(ctx)
		return err
	})
}

func (b *baseRepo[M]) FindByID(ctx context.Context, id uuid.UUID) (*M, error) {
	model := new(M)
	err := b.executor.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().Model(model).Where("id = ?", id).Scan(ctx)
	})
	return model, transformErrNotFound(err, b.name)
}

func (b *baseRepo[M]) ExistsByID(ctx context.Context, id uuid.UUID) (exists bool, err error) {
	err = b.executor.Run(ctx, func(idb bun.IDB) error {
		exists, err = idb.NewSelect().Model(new(M)).Where("id = ?", id).Exists(ctx)
		return err
	})
	return exists, err
}

func (b *baseRepo[M]) UpdateByID(ctx context.Context, id uuid.UUID, model *M) error {
	return b.executor.Run(ctx, func(idb bun.IDB) error {
		result, err := idb.NewUpdate().Model(model).Where("id = ?", id).Exec(ctx)
		if err != nil {
			return err
		}
		if n, err := result.RowsAffected(); err == nil && n == 0 {
			return apperror.NotFound(b.name + " not found")
		}
		return nil
	})
}

func (b *baseRepo[M]) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return b.executor.Run(ctx, func(idb bun.IDB) error {
		result, err := idb.NewDelete().Model(new(M)).Where("id = ?", id).Exec(ctx)
		if err != nil {
			return err
		}
		if n, err := result.RowsAffected(); err == nil && n == 0 {
			return apperror.NotFound(b.name + " not found")
		}
		return nil
	})
}

func transformErrNotFound(err error, name string) error {
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return apperror.NotFound(name + " not found")
	}
	return err
}
