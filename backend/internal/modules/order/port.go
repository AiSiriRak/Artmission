package order

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OrderUsecase interface {
	// ViewOrders lists query's page, then resolves each order's
	// DeliverablePreviewKey (if any) into a presigned DeliverablePreviewURL.
	// Orders with no submitted deliverable get a nil URL.
	ViewOrders(ctx context.Context, query ListQuery) (Page, error)

	// GetOrder returns an order's details if the authenticated user is a participant.
	GetOrder(
		ctx context.Context,
		participant Participant,
		participantID uuid.UUID,
		orderID uuid.UUID,
	) (*OrderDetail, error)

	// ConfirmOrder accepts or rejects an order on behalf of its artist.
	// The order must belong to the artist and currently be PENDING.
	// Accept = true transitions the order's status to NOT_PAID.
	// Accept = false transitions the order's status to CANCEL.
	ConfirmOrder(
		ctx context.Context,
		artistID uuid.UUID,
		orderID uuid.UUID,
		input ConfirmOrderInput,
	) (Status, error)
}

type OrderRepository interface {
	// ListOrders returns one page of orders for an already-validated query.
	ListOrders(ctx context.Context, query ListQuery) (Page, error)

	// GetOrderByID returns the order's details and both participants if the caller is a participant.
	GetOrderByID(
		ctx context.Context,
		participant Participant,
		participantID uuid.UUID,
		orderID uuid.UUID,
	) (*OrderDetailData, error)

	// ConfirmOrder transitions a PENDING order to the given status,
	// scoped to the authenticated artist who owns the order.
	ConfirmOrder(
		ctx context.Context,
		artistID uuid.UUID,
		orderID uuid.UUID,
		status Status,
	) error
}

type ObjectStorage interface {
	// GetPresignedURL returns a time-limited URL for the object key.
	GetPresignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}

type ConfirmOrderInput struct {
	Accept bool
}
