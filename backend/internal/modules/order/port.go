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
	ViewOrders(
		ctx context.Context,
		query ListQuery,
	) (Page, error)

	// GetOrder returns an order's details if the authenticated user is a participant.
	GetOrder(
		ctx context.Context,
		participant Participant,
		participantID uuid.UUID,
		orderID uuid.UUID,
	) (*OrderDetail, error)

	// CreateOrder creates a new order on behalf of the specified customer.
	CreateOrder(
		ctx context.Context,
		customerID uuid.UUID,
		input CreateInput,
	) (*Order, error)

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

	// CancelOrder cancels an order on behalf of an authorized participant
	// and records the cancellation reason and explanation.
	CancelOrder(
		ctx context.Context,
		participant Participant,
		participantID uuid.UUID,
		input CancelInput,
	) (*OrderCancellation, error)

	// CancelExpiredOrders cancels every order whose artist-confirmation window
	// or deadline has passed, and returns the IDs it cancelled.
	CancelExpiredOrders(
		ctx context.Context,
	) ([]uuid.UUID, error)

	// CreateDeliverable creates a new deliverable version for an order
	// owned by the authenticated artist. It stores the original image,
	// generates a watermarked preview, and persists the deliverable metadata.
	CreateDeliverable(
		ctx context.Context,
		artistID uuid.UUID,
		input CreateDeliverableInput,
	) (*Deliverable, error)

	// ConfirmDeliverable accepts or rejects a deliverable on behalf of
	// the customer who owns the order.
	ConfirmDeliverable(
		ctx context.Context,
		customerID uuid.UUID,
		input ConfirmDeliverableInput,
	) (DeliverableDecision, error)

	// AutoApproveExpiredDeliverables approves pending deliverables whose
	// review window has expired, and returns the IDs of the approved deliverables.
	AutoApproveExpiredDeliverables(
		ctx context.Context,
	) ([]uuid.UUID, error)

	// CreateDelayRequest creates a deadline-extension request for an order
	// owned by the specified artist.
	CreateDelayRequest(
		ctx context.Context,
		artistID uuid.UUID,
		input CreateDelayRequestInput,
	) (*OrderDelayRequest, error)

	// ConfirmDelayRequest confirms or rejects a customer's decision on an order delay request
	// and returns the decision and the updated order deadline.
	ConfirmDelayRequest(
		ctx context.Context,
		customerID uuid.UUID,
		orderID uuid.UUID,
		input ConfirmDelayRequestInput,
	) (*ConfirmDelayRequestOutput, error)

	// CancelExpiredDelayRequest rejects delay requests whose response window
	// has expired, and returns the IDs of the rejected delay requests.
	CancelExpiredDelayRequest(
		ctx context.Context,
	) ([]uuid.UUID, error)
}

type OrderRepository interface {
	// ListOrders returns one page of orders for an already-validated query.
	ListOrders(
		ctx context.Context,
		query ListQuery,
	) (Page, error)

	// GetOrderByID returns the order's details and both participants
	// if the specified participant is authorized to access the order.
	GetOrderByID(
		ctx context.Context,
		participant Participant,
		participantID uuid.UUID,
		orderID uuid.UUID,
	) (*OrderDetailData, error)

	// Create persists a new order.
	Create(
		ctx context.Context,
		order *Order,
	) error

	// ConfirmOrder transitions a PENDING order to the given status,
	// scoped to the authenticated artist who owns the order.
	ConfirmOrder(
		ctx context.Context,
		artistID uuid.UUID,
		orderID uuid.UUID,
		status Status,
	) error

	// CancelOrder persists an order cancellation for an authorized participant.
	CancelOrder(
		ctx context.Context,
		participant Participant,
		participantID uuid.UUID,
		item *OrderCancellation,
	) error

	// CancelExpired atomically moves to CANCEL every order that is either
	// PENDING with created_at <= pendingCutoff, or in an auto-cancellable status
	// with deadline_at <= now. It returns the IDs it changed.
	CancelExpired(
		ctx context.Context,
		pendingCutoff time.Time,
		now time.Time,
	) ([]uuid.UUID, error)

	// CreateDeliverable creates a new deliverable version for an order
	// owned by the specified artist. The order must be IN_PROCESS.
	CreateDeliverable(
		ctx context.Context,
		artistID uuid.UUID,
		orderID uuid.UUID,
		item *Deliverable,
	) (*Deliverable, error)

	// ValidateOrderForDeliverable verifies that the order exists, belongs to the
	// specified artist, and is in a valid state for creating a deliverable.
	// It should be called before processing and uploading deliverable images.
	ValidateOrderForDeliverable(
		ctx context.Context,
		artistID uuid.UUID,
		orderID uuid.UUID,
	) error

	// ConfirmDeliverable accepts or rejects a deliverable on an order
	// owned by the specified customer.
	ConfirmDeliverable(
		ctx context.Context,
		customerID uuid.UUID,
		orderID uuid.UUID,
		deliverableID uuid.UUID,
		decision DeliverableDecision,
		comment string,
	) error

	// AutoApproveExpiredDeliverables approves pending deliverables
	// for orders that are still IN_PROCESS.
	AutoApproveExpiredDeliverables(
		ctx context.Context,
		expiringCutoff time.Time,
		now time.Time,
	) ([]uuid.UUID, error)

	// CreateDelayRequest persists a deadline-extension request for an order
	// owned by the specified artist.
	CreateDelayRequest(
		ctx context.Context,
		artistID uuid.UUID,
		orderID uuid.UUID,
		item *OrderDelayRequest,
	) (*OrderDelayRequest, error)

	// ConfirmDelayRequest updates the customer's decision on an order delay
	// request and returns the order's deadline after the update.
	ConfirmDelayRequest(
		ctx context.Context,
		customerID uuid.UUID,
		orderID uuid.UUID,
		delayRequestID uuid.UUID,
		delayRequestDecision DelayRequestDecision,
	) (time.Time, error)

	// CancelExpiredRequest rejects pending delay requests created before
	// expiringCutoff and returns the IDs of the rejected requests.
	CancelExpiredRequest(ctx context.Context,
		expiringCutoff time.Time,
		now time.Time,
	) ([]uuid.UUID, error)
}

type ObjectStorage interface {
	// GetPresignedURL returns a time-limited URL for the object key.
	GetPresignedURL(
		ctx context.Context,
		key string,
		ttl time.Duration,
	) (string, error)

	// Upload uploads an object from body to the specified key with the given content type.
	Upload(ctx context.Context,
		key string,
		body io.Reader,
		contentType string,
	) error

	// Delete deletes an object by the key.
	Delete(
		ctx context.Context,
		key string,
	) error
}
