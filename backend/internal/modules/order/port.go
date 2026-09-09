package order

import (
	"context"

	"github.com/google/uuid"
)

type OrderUsecase interface {
	// ViewHiringHistory lists every order belonging to customerID, most
	// recent first. An empty slice (no orders yet) is a valid result, not
	// an error.
	ViewHiringHistory(ctx context.Context, customerID uuid.UUID) ([]Order, error)
}

// OrderRepository is the driven port for order persistence. Only the query
// hiring history needs exists today; Create/Cancel/etc. land with the
// Order Lifecycle epic.
type OrderRepository interface {
	ListByCustomerID(ctx context.Context, customerID uuid.UUID) ([]Order, error)
}
