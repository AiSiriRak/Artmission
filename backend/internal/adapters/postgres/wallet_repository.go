package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type walletRepository struct{ exec baserepo.Executor }

var _ payment.WalletRepository = (*walletRepository)(nil)

func NewWalletRepository(db *bun.DB) payment.WalletRepository {
	return &walletRepository{exec: baserepo.NewExecutor(db)}
}

func (r *walletRepository) GetWallet(ctx context.Context, userID uuid.UUID) (*payment.Wallet, error) {
	wallet := &payment.Wallet{UserID: userID}
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewRaw(`SELECT user_id, balance_satang, updated_at FROM wallets WHERE user_id = ?`, userID).Scan(ctx, wallet)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return wallet, nil
	}
	if err != nil {
		return nil, apperror.Internal("failed to load wallet", err)
	}
	return wallet, nil
}

func (r *walletRepository) ListWalletTransactions(ctx context.Context, userID uuid.UUID, query payment.TransactionQuery) (payment.TransactionPage, error) {
	page := payment.TransactionPage{Transactions: []payment.WalletTransaction{}}
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		if err := idb.NewRaw(`SELECT COUNT(*) FROM wallet_entries w
			JOIN payment_transactions t ON t.id = w.payment_transaction_id
			WHERE w.user_id = ? AND t.recipient_user_id = ? AND t.transfer_status = 'SUCCESS'
			AND t.type IN ('RELEASE_TO_ARTIST','REFUND_TO_CUSTOMER') AND w.amount_satang > 0`, userID, userID).Scan(ctx, &page.Total); err != nil {
			return err
		}
		return idb.NewRaw(`SELECT t.id, p.order_id, o.name AS order_name, t.type, w.amount_satang,
			p.settlement_outcome, w.created_at FROM wallet_entries w
			JOIN payment_transactions t ON t.id = w.payment_transaction_id
			JOIN payments p ON p.id = t.payment_id JOIN orders o ON o.id = p.order_id
			WHERE w.user_id = ? AND t.recipient_user_id = ? AND t.transfer_status = 'SUCCESS'
			AND t.type IN ('RELEASE_TO_ARTIST','REFUND_TO_CUSTOMER') AND w.amount_satang > 0
			ORDER BY w.created_at DESC, t.id DESC LIMIT ? OFFSET ?`, userID, userID, query.Limit, query.Offset).Scan(ctx, &page.Transactions)
	})
	if err != nil {
		return payment.TransactionPage{}, apperror.Internal("failed to load wallet transactions", err)
	}
	return page, nil
}
