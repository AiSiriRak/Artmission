package worker

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/google/uuid"
)

type expiryUsecase struct {
	order.OrderUsecase
	cancelExpired func(context.Context) ([]uuid.UUID, error)
}

func (u expiryUsecase) CancelExpiredOrders(ctx context.Context) ([]uuid.UUID, error) {
	return u.cancelExpired(ctx)
}

var _ order.OrderUsecase = expiryUsecase{}

func TestOrderExpiryWorker_RunsSweepImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calls := 0
	uc := expiryUsecase{
		cancelExpired: func(context.Context) ([]uuid.UUID, error) {
			calls++
			cancel()
			return nil, nil
		},
	}
	worker := NewOrderExpiryWorker(uc, time.Hour, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	worker.Run(ctx)

	if calls != 1 {
		t.Fatalf("sweep calls = %d, want 1 immediate sweep", calls)
	}
}

func TestOrderExpiryWorker_LogsFailureAndRetriesOnNextTick(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var logs bytes.Buffer
	calls := 0
	uc := expiryUsecase{
		cancelExpired: func(context.Context) ([]uuid.UUID, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("database unavailable")
			}
			cancel()
			return nil, nil
		},
	}
	worker := NewOrderExpiryWorker(uc, time.Millisecond, slog.New(slog.NewTextHandler(&logs, nil)))

	worker.Run(ctx)

	if calls < 2 {
		t.Fatalf("sweep calls = %d, want retry after failure", calls)
	}
	if !strings.Contains(logs.String(), "order expiry sweep failed") {
		t.Fatalf("expected failed sweep to be logged, got %q", logs.String())
	}
}
