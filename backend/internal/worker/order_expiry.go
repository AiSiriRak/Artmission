package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
)

type OrderExpiryWorker struct {
	usecase  order.OrderUsecase
	interval time.Duration
	log      *slog.Logger
}

func NewOrderExpiryWorker(uc order.OrderUsecase, interval time.Duration, log *slog.Logger) *OrderExpiryWorker {
	return &OrderExpiryWorker{usecase: uc, interval: interval, log: log}
}

// Run sweeps once immediately (catching up after downtime), then on every tick,
// until ctx is cancelled.
func (w *OrderExpiryWorker) Run(ctx context.Context) {
	w.sweep(ctx)

	t := time.NewTicker(w.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.sweep(ctx)
		}
	}
}

func (w *OrderExpiryWorker) sweep(ctx context.Context) {
	ids, err := w.usecase.CancelExpiredOrders(ctx)
	if err != nil {
		w.log.Error("order expiry sweep failed", "err", err)
		return
	}
	if len(ids) > 0 {
		w.log.Info("orders auto-cancelled", "count", len(ids))
	}
}

type Worker interface {
	Run(ctx context.Context)
}

var _ Worker = (*OrderExpiryWorker)(nil)
