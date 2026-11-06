package rest

import (
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/google/uuid"
)

type paymentSummaryView struct {
	PaymentStatus              string                     `json:"payment_status"`
	Source                     string                     `json:"source" doc:"MOCK_WALLET means simulated credits only; EXTERNAL is historical/external accounting."`
	AmountSatang               int64                      `json:"amount_satang"`
	SettlementOutcome          *payment.SettlementOutcome `json:"settlement_outcome"`
	SettledAt                  *time.Time                 `json:"settled_at"`
	RefundSatang               int64                      `json:"refund_satang"`
	ArtistPayoutSatang         int64                      `json:"artist_payout_satang"`
	RefundTransferStatus       *payment.TransferStatus    `json:"refund_transfer_status"`
	ArtistPayoutTransferStatus *payment.TransferStatus    `json:"artist_payout_transfer_status"`
	RefundTransactionID        *uuid.UUID                 `json:"refund_transaction_id"`
	ArtistPayoutTransactionID  *uuid.UUID                 `json:"artist_payout_transaction_id"`
}

func toPaymentSummaryView(p *payment.Summary) *paymentSummaryView {
	if p == nil {
		return nil
	}
	return &paymentSummaryView{
		PaymentStatus: p.Status, Source: p.Source, AmountSatang: p.AmountSatang,
		SettlementOutcome: p.SettlementOutcome, SettledAt: p.SettledAt,
		RefundSatang: p.RefundSatang, ArtistPayoutSatang: p.ArtistPayoutSatang,
		RefundTransferStatus: p.RefundTransferStatus, ArtistPayoutTransferStatus: p.ArtistPayoutTransferStatus,
		RefundTransactionID: p.RefundTransactionID, ArtistPayoutTransactionID: p.ArtistPayoutTransactionID,
	}
}
