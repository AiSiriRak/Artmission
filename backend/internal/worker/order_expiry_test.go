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
	cancelExpired func(context.Context) ([]uuid.UUID, error)
}

func (u expiryUsecase) ViewOrders(context.Context, order.ListQuery) (order.Page, error) {
	return order.Page{}, nil
}

func (u expiryUsecase) GetOrder(context.Context, order.Participant, uuid.UUID, uuid.UUID) (*order.OrderDetail, error) {
	return nil, nil
}

func (u expiryUsecase) ConfirmOrder(context.Context, uuid.UUID, uuid.UUID, order.ConfirmOrderInput) (order.Status, error) {
	return "", nil
}

func (u expiryUsecase) CancelOrder(context.Context, order.Participant, uuid.UUID, uuid.UUID) error {
	return nil
}

func (u expiryUsecase) CancelExpiredOrders(ctx context.Context) ([]uuid.UUID, error) {
	return u.cancelExpired(ctx)
}

func (u expiryUsecase) CreateOrder(context.Context, uuid.UUID, order.CreateInput) (*order.Order, error) {
	return nil, nil
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
