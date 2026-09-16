// Package seed loads deterministic local/dev fixture data.
package seed

import (
	"context"
	"fmt"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/objectstorage"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// seedNamespace is an arbitrary, fixed UUID used only to derive
// deterministic IDs (see id).
var seedNamespace = uuid.MustParse("8f4a1e2c-6b3d-4f7a-9c1e-2d5b8a3f6e91")

// id creates deterministic UUID from namespace + custom data.
// See how UUID V5 works for more details.
func id(kind, key string) uuid.UUID {
	return uuid.NewSHA1(seedNamespace, []byte(kind+":"+key))
}

type deps struct {
	db      bun.IDB
	storage *objectstorage.Client
}

type step func(ctx context.Context, d deps) error

var upSteps = []step{
	seedCategoriesUp,
	seedStylesUp,
	seedUsersUp,
	seedBankAccountsUp,
	seedArtistProfilesUp,
	seedArtworksUp,
	seedArtworkImagesUp,
	seedArtworkStylesUp,
	seedOrdersUp,
	seedOrderDeliverablesUp,
}

// Declare steps only the tables that actually need an explicit delete.
// The rest shall be deleted via ON DELETE CASCADE.
var downSteps = []step{
	seedOrderDeliverablesDown,
	seedOrdersDown,
	seedUsersDown,
	seedArtworkImagesDown,
}

// Up seeds every table. Safe to re-run: every insert is ON CONFLICT DO
// NOTHING against a deterministic primary key, and image uploads
// overwrite the same deterministic key with identical bytes.
func Up(ctx context.Context, db *bun.DB, storage *objectstorage.Client) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		d := deps{db: tx, storage: storage}
		for _, s := range upSteps {
			if err := s(ctx, d); err != nil {
				return err
			}
		}
		return nil
	})
}

func Down(ctx context.Context, db *bun.DB, storage *objectstorage.Client) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		d := deps{db: tx, storage: storage}
		for _, s := range downSteps {
			if err := s(ctx, d); err != nil {
				return err
			}
		}
		return nil
	})
}

func Reseed(ctx context.Context, db *bun.DB, storage *objectstorage.Client) error {
	if err := Down(ctx, db, storage); err != nil {
		return err
	}
	return Up(ctx, db, storage)
}

// seedTable inserts rows into d.db, doing nothing for any row whose
// conflict target already exists. conflictTarget is the column (or
// "a, b" column list) enforcing each row's identity, e.g. "id" or
// "user_id".
func seedTable[T any](ctx context.Context, d deps, rows []T, conflictTarget string) error {
	if len(rows) == 0 {
		return nil
	}
	if _, err := d.db.NewInsert().Model(&rows).On("CONFLICT (" + conflictTarget + ") DO NOTHING").Exec(ctx); err != nil {
		return fmt.Errorf("seed table: %w", err)
	}
	return nil
}

// deleteByIDs hard-deletes rows. ForceDelete is required for User: bun would
// otherwise stamp deleted_at and Up's ON CONFLICT (id) would skip them.
func deleteByIDs[T any](ctx context.Context, d deps, idColumn string, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	if _, err := d.db.NewDelete().Model((*T)(nil)).Where(idColumn+" IN (?)", bun.List(ids)).ForceDelete().Exec(ctx); err != nil {
		return fmt.Errorf("delete seeded rows: %w", err)
	}
	return nil
}
