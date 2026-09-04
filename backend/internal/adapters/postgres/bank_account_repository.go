package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type bankAccountModel struct {
	bun.BaseModel `bun:"table:bank_accounts,alias:ba"`

	UserID        uuid.UUID `bun:"user_id,pk"`
	BankName      string    `bun:"bank_name"`
	AccountNumber string    `bun:"account_number"`
	CreatedAt     time.Time `bun:"created_at,nullzero"`
	UpdatedAt     time.Time `bun:"updated_at,nullzero"`
}

func newBankAccountModel(ba *user.BankAccount) *bankAccountModel {
	return &bankAccountModel{
		UserID:        ba.UserID,
		BankName:      ba.BankName,
		AccountNumber: ba.AccountNumber,
		CreatedAt:     ba.CreatedAt,
		UpdatedAt:     ba.UpdatedAt,
	}
}

type bankAccountRepository struct {
	exec baserepo.Executor
}

var _ user.BankAccountRepository = (*bankAccountRepository)(nil)

func NewBankAccountRepository(db *bun.DB) user.BankAccountRepository {
	return &bankAccountRepository{exec: baserepo.NewExecutor(db)}
}

func (r *bankAccountRepository) Create(ctx context.Context, ba *user.BankAccount) error {
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		_, err := idb.NewInsert().Model(newBankAccountModel(ba)).Exec(ctx)
		return err
	})
	if err != nil {
		return apperror.Internal("failed to create bank account", err)
	}
	return nil
}

func (r *bankAccountRepository) UpdateByUserID(ctx context.Context, ba *user.BankAccount) error {
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		result, err := idb.NewUpdate().
			Model(newBankAccountModel(ba)).
			Column("bank_name", "account_number", "updated_at").
			Where("user_id = ?", ba.UserID).
			Exec(ctx)
		if err != nil {
			return err
		}
		if n, err := result.RowsAffected(); err != nil {
			return fmt.Errorf("read updated bank account rows: %w", err)
		} else if n == 0 {
			return user.ErrBankAccountNotFound
		}
		return nil
	})
	if err != nil {
		if err == user.ErrBankAccountNotFound {
			return err
		}
		return apperror.Internal("failed to update bank account", err)
	}
	return nil
}
