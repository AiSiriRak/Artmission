package rest

import (
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
)

func TestOrderPaymentViewPreservesPersistedAllocation(t *testing.T) {
	status := payment.TransferPending
	outcome := payment.CustomerCancelled
	o := &order.Order{Name: "retained", PriceSatangOrder: 10001, Payment: &payment.Summary{
		Status: "SETTLED", Source: "MOCK_WALLET", AmountSatang: 10001, SettlementOutcome: &outcome,
		RefundSatang: 4999, ArtistPayoutSatang: 5002, RefundTransferStatus: &status, ArtistPayoutTransferStatus: &status,
	}}
	summary := toOrderSummaryView(o)
	detail := toOrderDetailView(&order.OrderDetail{Order: *o}, "", nil)
	for _, p := range []*paymentSummaryView{summary.Payment, detail.Payment} {
		if p == nil || p.RefundSatang != 4999 || p.ArtistPayoutSatang != 5002 || *p.RefundTransferStatus != payment.TransferPending {
			t.Fatalf("recalculated historical data or completed pending transfer: %+v", p)
		}
	}
	if summary.Name != "retained" || summary.PriceSatang != 10001 {
		t.Fatal("existing fields lost")
	}
	if toOrderSummaryView(&order.Order{}).Payment != nil {
		t.Fatal("invented payment")
	}
}
