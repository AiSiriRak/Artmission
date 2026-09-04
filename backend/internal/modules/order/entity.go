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
	StatusPending    Status = "PENDING"
	StatusInProgress Status = "IN_PROGRESS"
	StatusComplete   Status = "COMPLETE"
	StatusCancelled  Status = "CANCELLED"
)

type Order struct {
	ID          uuid.UUID
	CustomerID  uuid.UUID
	ArtistID    uuid.UUID
	Description string
	Category    string
	Style       string
	// Price is a nullable pointer (unlike Category/Style, an empty-string
	// zero-value isn't a safe stand-in for "no price": 0 is itself a valid
	// price). float64 for now since this slice never writes it (only reads
	// existing rows); revisit as a fixed-point/decimal type once order
	// creation and payment (EPIC 6/7) do money arithmetic here.
	Price       *float64
	Status      Status
	Deadline    *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
