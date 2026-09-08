// Package order owns the commission order lifecycle. This slice only reads
// orders (customer hiring history); creation and the rest of the lifecycle
// land in a later increment, but the schema is modeled fully now to avoid
// a breaking migration later.
package order

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusNotPaid   Status = "NOT_PAID"
	StatusInProcess Status = "IN_PROCESS"
	StatusSuccess   Status = "SUCCESS"
	StatusCancel    Status = "CANCEL"
)

type Category struct {
	ID    uuid.UUID
	Label string
}

type Style struct {
	ID    uuid.UUID
	Label string
}

type Order struct {
	ID          uuid.UUID
	CustomerID  uuid.UUID
	ArtistID    uuid.UUID
	Description string
	Category    *Category
	Style       *Style
	// Price is a nullable pointer: 0 is itself a valid price. float64 for now
	// since this slice never writes it; revisit as fixed-point/decimal once
	// order creation and payment (EPIC 6/7) do money arithmetic here.
	Price       *float64
	Status      Status
	Deadline    *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
