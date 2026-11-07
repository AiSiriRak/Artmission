package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
)

type PaymentQRWorker struct {
	usecase  payment.Usecase
	interval time.Duration
	log      *slog.Logger
}

func NewPaymentQRWorker(uc payment.Usecase, interval time.Duration, log *slog.Logger) *PaymentQRWorker {
	return &PaymentQRWorker{usecase: uc, interval: interval, log: log}
}

func (w *PaymentQRWorker) Run(ctx context.Context) {
	w.sweep(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.sweep(ctx)
		}
	}
}

func (w *PaymentQRWorker) sweep(ctx context.Context) {
	cancelledOrders, err := w.usecase.ExpirePayments(ctx)
	if err != nil {
		w.log.Error("payment QR expiry sweep failed", "err", err)
	} else if len(cancelledOrders) > 0 {
		w.log.Info("orders cancelled after payment QR expiry", "count", len(cancelledOrders))
	}
}

var _ Worker = (*PaymentQRWorker)(nil)
