package order

import (
	"context"

	"github.com/google/uuid"
)

// OrderRepository is the driven port for order persistence. Only the query
// hiring history needs exists today; Create/Cancel/etc. land with the
// Order Lifecycle epic.
type OrderRepository interface {
	ListByCustomerID(ctx context.Context, customerID uuid.UUID) ([]Order, error)
}
