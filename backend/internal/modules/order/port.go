package order

import (
	"context"
	"io"
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

	CancelOrder(ctx context.Context, participant Participant, participantID uuid.UUID, orderID uuid.UUID) error

	CreateOrder(ctx context.Context, customerID uuid.UUID, input CreateInput) (*Order, error)

	// CreateDeliverable creates a new deliverable version for an order
	// owned by the authenticated artist. It stores the original image,
	// generates a watermarked preview, and persists the deliverable metadata.
	CreateDeliverable(
		ctx context.Context,
		artistID uuid.UUID,
		input CreateDeliverableInput,
	) (*Deliverable, error)
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

	CancelOrder(ctx context.Context, participant Participant, participantID uuid.UUID, orderID uuid.UUID) error
	Create(ctx context.Context, order *Order) error

	// CreateDeliverable creates a new deliverable version for an order
	// owned by the specified artist. The order must be IN_PROCESS.
	CreateDeliverable(
		ctx context.Context,
		artistID uuid.UUID,
		orderID uuid.UUID,
		orderDeliverable *Deliverable,
	) error
}

type ObjectStorage interface {
	// GetPresignedURL returns a time-limited URL for the object key.
	GetPresignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)

	// Upload uploads an object from body to the specified key with the given content type.
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error

	// Delete deletes an object by the key.
	Delete(ctx context.Context, key string) error
}

type ConfirmOrderInput struct {
	Accept bool
}
