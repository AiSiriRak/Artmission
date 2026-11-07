package payment

import (
	"context"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

const DefaultTransactionLimit = 20
const MaxTransactionLimit = 100

type walletUsecase struct{ repo WalletRepository }

func NewWalletUsecase(repo WalletRepository) WalletUsecase { return &walletUsecase{repo: repo} }

func (u *walletUsecase) GetWallet(ctx context.Context, userID uuid.UUID) (*Wallet, error) {
	if userID == uuid.Nil {
		return nil, apperror.InvalidInput("user_id is required", nil)
	}
	return u.repo.GetWallet(ctx, userID)
}

func (u *walletUsecase) ListWalletTransactions(ctx context.Context, userID uuid.UUID, query TransactionQuery) (TransactionPage, error) {
	if userID == uuid.Nil {
		return TransactionPage{}, apperror.InvalidInput("user_id is required", nil)
	}
	if query.Limit == 0 {
		query.Limit = DefaultTransactionLimit
	}
	if query.Limit < 1 || query.Limit > MaxTransactionLimit || query.Offset < 0 {
		return TransactionPage{}, apperror.InvalidInput("limit must be between 1 and 100 and offset must not be negative", nil)
	}
	return u.repo.ListWalletTransactions(ctx, userID, query)
}
