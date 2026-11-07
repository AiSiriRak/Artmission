package postgres

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// The caller must hold the terminal order transition in the same transaction.
func settlePaymentTx(ctx context.Context, idb bun.IDB, orderIDs []uuid.UUID, outcome payment.SettlementOutcome) error {
	if len(orderIDs) == 0 {
		return nil
	}
	terminal := "CANCEL"
	if outcome == payment.Completed {
		terminal = "SUCCESS"
	}
	if _, _, err := payment.CalculateSettlement(outcome, 0); err != nil {
		return err
	}
	var held []struct {
		ID           uuid.UUID
		AmountSatang int64
		CustomerID   uuid.UUID
		ArtistID     uuid.UUID
	}
	err := idb.NewRaw(`UPDATE payments AS p SET status = 'SETTLED', settlement_outcome = ?, settled_at = NOW(), updated_at = NOW()
		FROM orders AS o WHERE o.id = p.order_id AND o.id IN (?) AND o.status = ? AND p.status = 'HELD'
		RETURNING p.id, p.amount_satang, o.customer_id, o.artist_id`, outcome, bun.List(orderIDs), terminal).Scan(ctx, &held)
	if err != nil {
		return err
	}
	type credit struct {
		paymentID, recipientID uuid.UUID
		kind                   payment.TransactionType
		amount                 int64
	}
	var credits []credit
	for _, p := range held {
		refund, artist, err := payment.CalculateSettlement(outcome, p.AmountSatang)
		if err != nil {
			return err
		}
		for _, allocation := range []credit{
			{p.ID, p.ArtistID, payment.ReleaseToArtist, artist},
			{p.ID, p.CustomerID, payment.RefundToCustomer, refund},
		} {
			if allocation.amount == 0 {
				continue
			}
			credits = append(credits, allocation)
		}
	}
	// Batch settlements acquire shared recipient wallets in the same order to avoid deadlocks.
	sort.Slice(credits, func(i, j int) bool { return credits[i].recipientID.String() < credits[j].recipientID.String() })
	for _, credit := range credits {
		transactionID := uuid.New()
		if _, err := idb.NewRaw(`INSERT INTO payment_transactions
			(id, payment_id, type, amount_satang, recipient_user_id, transfer_status, transfer_ref, transferred_at)
			VALUES (?, ?, ?, ?, ?, 'SUCCESS', ?, NOW())`, transactionID, credit.paymentID, credit.kind,
			credit.amount, credit.recipientID, "wallet-credit-"+transactionID.String()).Exec(ctx); err != nil {
			return err
		}
		if _, err := idb.NewRaw(`INSERT INTO wallet_entries (id, user_id, payment_transaction_id, kind, amount_satang)
			VALUES (?, ?, ?, ?, ?)`, uuid.New(), credit.recipientID, transactionID, credit.kind, credit.amount).Exec(ctx); err != nil {
			return err
		}
		if _, err := idb.NewRaw(`INSERT INTO wallets (user_id, balance_satang, updated_at) VALUES (?, ?, NOW())
			ON CONFLICT (user_id) DO UPDATE SET balance_satang = wallets.balance_satang + EXCLUDED.balance_satang,
			updated_at = EXCLUDED.updated_at`, credit.recipientID, credit.amount).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *orderRepository) attachPaymentSummaries(ctx context.Context, orders []order.Order) error {
	if len(orders) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(orders))
	byID := make(map[uuid.UUID]*order.Order, len(orders))
	for i := range orders {
		ids[i] = orders[i].ID
		byID[ids[i]] = &orders[i]
	}
	return r.exec.Run(ctx, func(idb bun.IDB) error {
		var rows []struct {
			RefundTransactionID        *uuid.UUID
			ArtistPayoutTransactionID  *uuid.UUID
			OrderID                    uuid.UUID
			Status                     string
			Source                     string
			AmountSatang               int64
			SettlementOutcome          *payment.SettlementOutcome
			SettledAt                  *time.Time
			RefundSatang               int64
			ArtistPayoutSatang         int64
			RefundTransferStatus       *payment.TransferStatus
			ArtistPayoutTransferStatus *payment.TransferStatus
		}
		if err := idb.NewRaw(`SELECT p.order_id, p.status, p.source, p.amount_satang, p.settlement_outcome, p.settled_at,
			COALESCE(r.amount_satang, 0) AS refund_satang, COALESCE(a.amount_satang, 0) AS artist_payout_satang,
			r.transfer_status AS refund_transfer_status, a.transfer_status AS artist_payout_transfer_status,
			r.id AS refund_transaction_id, a.id AS artist_payout_transaction_id
			FROM payments p
			LEFT JOIN payment_transactions r ON r.payment_id = p.id AND r.type = 'REFUND_TO_CUSTOMER'
			LEFT JOIN payment_transactions a ON a.payment_id = p.id AND a.type = 'RELEASE_TO_ARTIST'
			WHERE p.order_id IN (?)`, bun.List(ids)).Scan(ctx, &rows); err != nil {
			return fmt.Errorf("load payment summaries: %w", err)
		}
		for _, row := range rows {
			byID[row.OrderID].Payment = &payment.Summary{
				RefundTransactionID: row.RefundTransactionID, ArtistPayoutTransactionID: row.ArtistPayoutTransactionID,
				Status: row.Status, Source: row.Source, AmountSatang: row.AmountSatang,
				SettlementOutcome: row.SettlementOutcome, SettledAt: row.SettledAt,
				RefundSatang: row.RefundSatang, ArtistPayoutSatang: row.ArtistPayoutSatang,
				RefundTransferStatus: row.RefundTransferStatus, ArtistPayoutTransferStatus: row.ArtistPayoutTransferStatus,
			}
		}
		return nil
	})
}
