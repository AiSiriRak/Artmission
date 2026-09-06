package postgres

import (
	"context"
	"database/sql"
	"errors"

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

func NewAccountDeletionRepository(db *bun.DB) user.AccountDeletionRepository {
	return &accountDeletionRepository{exec: baserepo.NewExecutor(db)}
}

func (r *accountDeletionRepository) HasOrdersInStatuses(ctx context.Context, userID uuid.UUID, statuses []order.Status) (exists bool, err error) {
	err = r.exec.Run(ctx, func(idb bun.IDB) error {
		exists, err = idb.NewSelect().
			TableExpr("orders AS o").
			Where("(o.customer_id = ? OR o.artist_id = ?)", userID, userID).
			Where("o.status IN (?)", bun.In(statuses)).
			Exists(ctx)
		return err
	})
	if err != nil {
		return false, apperror.Internal("failed to check active orders", err)
	}
	return exists, nil
}

func (r *accountDeletionRepository) DeleteBankAccountByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.deleteByUserID(ctx, "bank_accounts", userID, "failed to delete bank account")
}

func (r *accountDeletionRepository) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.deleteByUserID(ctx, "sessions", userID, "failed to invalidate sessions")
}

func (r *accountDeletionRepository) SoftDeleteUserByID(ctx context.Context, userID uuid.UUID) error {
	model := &userModel{ID: userID}
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
