package baserepo

import (
	"context"

	"github.com/uptrace/bun"
)

type dbContextKey struct{}

// Executor runs a query against whatever bun.IDB is active for ctx: a
// transaction if Transactioner.Transaction started one, otherwise the pool.
// Repositories depend on Executor, never on *bun.DB directly, so their
// queries automatically join an ambient transaction without threading one
// through every method signature.
type Executor interface {
	Run(ctx context.Context, fn func(idb bun.IDB) error) error
}

// Transactioner starts a transaction and exposes it to nested Executor.Run
// calls via the context.
type Transactioner interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type (
	executorImpl      struct{ db *bun.DB }
	transactionerImpl struct{ db *bun.DB }
)

func NewExecutor(db *bun.DB) Executor {
	return &executorImpl{db}
}

func (e *executorImpl) Run(ctx context.Context, fn func(idb bun.IDB) error) error {
	idb, ok := ctx.Value(dbContextKey{}).(bun.IDB)
	if !ok {
		idb = e.db
	}
	return fn(idb)
}

func NewTransactioner(db *bun.DB) Transactioner {
	return &transactionerImpl{db}
}

func (t *transactionerImpl) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(context.WithValue(ctx, dbContextKey{}, tx))
	})
}
