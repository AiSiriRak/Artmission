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
	cancelExpired                 func(context.Context) ([]uuid.UUID, error)
	autoApproveExpiredDeliverable func(context.Context) ([]uuid.UUID, error)
	cancelExpiredDelayRequest     func(context.Context) ([]uuid.UUID, error)
}

func (u expiryUsecase) CancelExpiredOrders(ctx context.Context) ([]uuid.UUID, error) {
	return u.cancelExpired(ctx)
}

func (u expiryUsecase) CancelExpiredDelayRequest(ctx context.Context) ([]uuid.UUID, error) {
	return u.cancelExpiredDelayRequest(ctx)
}

func (u expiryUsecase) AutoApproveExpiredDeliverables(ctx context.Context) ([]uuid.UUID, error) {
	return u.autoApproveExpiredDeliverable(ctx)
}

var _ order.OrderUsecase = expiryUsecase{}

func TestOrderExpiryWorker_RunsSweepImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calls := 0
	deliverableCalls := 0
	delayRequestCalls := 0
	uc := expiryUsecase{
		cancelExpired: func(context.Context) ([]uuid.UUID, error) {
			calls++
			return nil, nil
		},
		autoApproveExpiredDeliverable: func(context.Context) ([]uuid.UUID, error) {
			deliverableCalls++
			return nil, nil
		},
		cancelExpiredDelayRequest: func(context.Context) ([]uuid.UUID, error) {
			delayRequestCalls++
			cancel()
			return nil, nil
		},
	}
	worker := NewOrderExpiryWorker(uc, time.Hour, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	worker.Run(ctx)

	if calls != 1 {
		t.Fatalf("sweep calls = %d, want 1 immediate sweep", calls)
	}
	if deliverableCalls != 1 {
		t.Fatalf("deliverable sweep calls = %d, want 1 immediate sweep", deliverableCalls)
	}
	if delayRequestCalls != 1 {
		t.Fatalf("delay request sweep calls = %d, want 1 immediate sweep", delayRequestCalls)
	}
}

func TestOrderExpiryWorker_LogsFailureAndRetriesOnNextTick(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var logs bytes.Buffer
	calls := 0
	deliverableCalls := 0
	delayRequestCalls := 0
	uc := expiryUsecase{
		cancelExpired: func(context.Context) ([]uuid.UUID, error) {
			calls++
			return nil, nil
		},
		autoApproveExpiredDeliverable: func(context.Context) ([]uuid.UUID, error) {
			deliverableCalls++
			if deliverableCalls == 1 {
				return nil, errors.New("database unavailable")
			}
			return nil, nil
		},
		cancelExpiredDelayRequest: func(context.Context) ([]uuid.UUID, error) {
			delayRequestCalls++
			if delayRequestCalls == 1 {
				return nil, errors.New("database unavailable")
			}
			cancel()
			return nil, nil
		},
	}
	worker := NewOrderExpiryWorker(uc, time.Millisecond, slog.New(slog.NewTextHandler(&logs, nil)))

	worker.Run(ctx)

	if delayRequestCalls < 2 {
		t.Fatalf("delay request sweep calls = %d, want retry after failure", delayRequestCalls)
	}
	if !strings.Contains(logs.String(), "delay request expiry sweep failed") {
		t.Fatalf("expected failed sweep to be logged, got %q", logs.String())
	}
	if deliverableCalls < 2 {
		t.Fatalf("deliverable sweep calls = %d, want retry after failure", deliverableCalls)
	}
	if !strings.Contains(logs.String(), "deliverables expiry sweep failed") {
		t.Fatalf("expected failed deliverable sweep to be logged, got %q", logs.String())
	}
}
