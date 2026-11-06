package payment

import (
	"context"

	"github.com/google/uuid"
)

type WalletRepository interface {
	GetWallet(context.Context, uuid.UUID) (*Wallet, error)
	ListWalletTransactions(context.Context, uuid.UUID, TransactionQuery) (TransactionPage, error)
}

type WalletUsecase interface {
	GetWallet(context.Context, uuid.UUID) (*Wallet, error)
	ListWalletTransactions(context.Context, uuid.UUID, TransactionQuery) (TransactionPage, error)
}
