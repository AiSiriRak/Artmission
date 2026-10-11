package postgres

import (
	"context"
	"database/sql"
	"errors"

	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type accountDeletionRepository struct {
	exec baserepo.Executor
}

var _ user.AccountDeletionRepository = (*accountDeletionRepository)(nil)

// NewAccountDeletionRepository returns a repository that deletes account-related rows in a guarded workflow.
func NewAccountDeletionRepository(db *bun.DB) user.AccountDeletionRepository {
	return &accountDeletionRepository{exec: baserepo.NewExecutor(db)}
}

// LockUserByIDForDeletion acquires a row lock so account deletion cannot race session or order updates.
func (r *accountDeletionRepository) LockUserByIDForDeletion(ctx context.Context, userID uuid.UUID) error {
	model := &pgmodel.User{ID: userID}
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		// FOR UPDATE conflicts with session and order key-share locks; a plain soft-delete UPDATE does not.
		return idb.NewSelect().Model(model).Column("id").WherePK().For("UPDATE").Scan(ctx)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.ErrUserNotFound
		}
		return apperror.Internal("failed to lock account for deletion", err)
	}
	return nil
}

// HasOrdersInStatuses reports whether a user has any orders in one of the supplied statuses.
func (r *accountDeletionRepository) HasOrdersInStatuses(ctx context.Context, userID uuid.UUID, statuses []order.Status) (exists bool, err error) {
	err = r.exec.Run(ctx, func(idb bun.IDB) error {
		exists, err = idb.NewSelect().
			TableExpr("orders AS o").
			Where("(o.customer_id = ? OR o.artist_id = ?)", userID, userID).
			Where("o.status IN (?)", bun.List(statuses)).
			Exists(ctx)
		return err
	})
	if err != nil {
		return false, apperror.Internal("failed to check active orders", err)
	}
	return exists, nil
}

// DeleteBankAccountByUserID removes the bank account record linked to the user.
func (r *accountDeletionRepository) DeleteBankAccountByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.deleteByUserID(ctx, "bank_accounts", userID, "failed to delete bank account")
}

// DeleteSessionsByUserID removes all sessions belonging to the user.
func (r *accountDeletionRepository) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.deleteByUserID(ctx, "sessions", userID, "failed to invalidate sessions")
}

// SoftDeleteUserByID removes the user record by ID and maps a missing row to a domain error.
func (r *accountDeletionRepository) SoftDeleteUserByID(ctx context.Context, userID uuid.UUID) error {
	model := &pgmodel.User{ID: userID}
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewDelete().Model(model).WherePK().Returning("id").Scan(ctx)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.ErrUserNotFound
		}
		return apperror.Internal("failed to delete account", err)
	}
	return nil
}

// deleteByUserID deletes all rows from a table whose user_id matches the account owner.
func (r *accountDeletionRepository) deleteByUserID(ctx context.Context, table string, userID uuid.UUID, message string) error {
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		_, err := idb.NewDelete().Table(table).Where("user_id = ?", userID).Exec(ctx)
		return err
	})
	if err != nil {
		return apperror.Internal(message, err)
	}
	return nil
}
