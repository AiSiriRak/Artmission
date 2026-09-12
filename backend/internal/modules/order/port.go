package order

import (
	"context"
	"time"
)

type OrderUsecase interface {
	// ViewOrders lists query's page, then resolves each order's
	// DeliverablePreviewKey (if any) into a presigned DeliverablePreviewURL.
	// Orders with no submitted deliverable get a nil URL.
	ViewOrders(ctx context.Context, query ListQuery) (Page, error)
}

type OrderRepository interface {
	// ListOrders returns one page of orders for an already-validated query.
	ListOrders(ctx context.Context, query ListQuery) (Page, error)
}

type ObjectStorage interface {
	GetPresignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}
