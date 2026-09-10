// Package order owns the commission order lifecycle.
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

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusNotPaid, StatusInProcess, StatusSuccess, StatusCancel:
		return true
	default:
		return false
	}
}

// Participant is the caller's relationship to an order, used to scope
// ViewOrders to only the orders where the authenticated caller is that
// participant. It is always derived from the authenticated role, never
// accepted as a request field.
type Participant string

const (
	ParticipantCustomer Participant = "customer"
	ParticipantArtist   Participant = "artist"
)

func (p Participant) IsValid() bool {
	switch p {
	case ParticipantCustomer, ParticipantArtist:
		return true
	default:
		return false
	}
}

// SortField is a ViewOrders-selectable sort key.
type SortField string

const (
	SortFieldDeadline  SortField = "deadline"
	SortFieldPrice     SortField = "price"
	SortFieldUpdatedAt SortField = "updated_at"
)

func (f SortField) IsValid() bool {
	switch f {
	case SortFieldDeadline, SortFieldPrice, SortFieldUpdatedAt:
		return true
	default:
		return false
	}
}

// SortOrder is ViewOrders' client-visible sort direction.
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

func (o SortOrder) IsValid() bool {
	switch o {
	case SortOrderAsc, SortOrderDesc:
		return true
	default:
		return false
	}
}

// Default ViewOrders query values, applied by the usecase whenever a field
// is left unset.
const (
	DefaultSort  = SortFieldUpdatedAt
	DefaultOrder = SortOrderDesc
	DefaultLimit = 20
	MaxLimit     = 100
)

type Order struct {
	ID                          uuid.UUID
	CustomerID                  uuid.UUID
	ArtistID                    uuid.UUID
	ArtworkID                   *uuid.UUID
	ArtworkNameSnapshot         string
	ArtworkDescriptionSnapshot  string
	PriceSatangSnapshot         int64
	MinimumDeadlineDaysSnapshot int
	PreviewImageURLSnapshot     string
	CustomerDescription         string
	SelectedDeadlineDays        int
	DeadlineAt                  *time.Time
	Status                      Status
	Deliverables                []Deliverable
	CompletedAt                 *time.Time
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
}

type Deliverable struct {
	ID               uuid.UUID
	OriginalImageURL string
	PreviewImageURL  string
	SortOrder        int
	CreatedAt        time.Time
}

// ListQuery is ViewOrders input
//
// Participant and ParticipantID are always derived from the authenticated
// caller, never accepted as a request field, so the participant scope
// predicate can never be broadened by request content.
type ListQuery struct {
	Participant   Participant
	ParticipantID uuid.UUID
	// Statuses filters by status with OR semantics. Empty means every status.
	Statuses []Status
	Sort     SortField
	Order    SortOrder
	Limit    int
	Offset   int
}

// Page is one ViewOrders page.
type Page struct {
	Orders []Order
	Total  int
}
