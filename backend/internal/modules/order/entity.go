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

// IsValid reports whether s is a supported order status.
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusNotPaid, StatusInProcess, StatusSuccess, StatusCancel:
		return true
	default:
		return false
	}
}

// ArtworkSnapshot contains the artwork information captured when the order
// was created. It is used by OrderDetail to display the artwork state at the
// time of the order, even if the original artwork is changed later.
type ArtworkSnapshot struct {
	ArtworkName string
	CategoryID  uuid.UUID
	StyleIDs    []uuid.UUID
}

// DeliverableDecision represents the decision made on a deliverable.
type DeliverableDecision string

const (
	DeliverableDecisionWait     DeliverableDecision = "WAIT"
	DeliverableDecisionApproved DeliverableDecision = "APPROVED"
	DeliverableDecisionRejected DeliverableDecision = "REJECTED"
)

// IsValid reports whether the deliverable decision is a supported value.
func (d DeliverableDecision) IsValid() bool {
	switch d {
	case DeliverableDecisionWait,
		DeliverableDecisionApproved,
		DeliverableDecisionRejected:
		return true
	default:
		return false
	}
}

// Participant is the caller's relationship to an order. It is derived
// from the authenticated user's role and is used to scope order access.
// It is never accepted as a request field.
type Participant string

const (
	ParticipantCustomer Participant = "customer"
	ParticipantArtist   Participant = "artist"
)

// IsValid reports whether p identifies a supported order participant.
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

// IsValid reports whether f is a supported ViewOrders sort field.
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

// IsValid reports whether o is a supported sort direction.
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
	ID                  uuid.UUID
	CustomerID          uuid.UUID
	ArtistID            uuid.UUID
	Name                string
	ArtworkID           *uuid.UUID
	PriceSatangOrder    int64
	CustomerDescription string
	DeadlineAt          time.Time
	Status              Status
	// DeliverablePreviewKey is the private-bucket object key of the most
	// recently submitted deliverable version, regardless of order status;
	// nil if the artist hasn't submitted one yet. Populated by
	// OrderRepository.ListOrders.
	DeliverablePreviewKey *string
	// DeliverablePreviewURL is DeliverablePreviewKey resolved to a
	// short-lived presigned GET URL by ViewOrders.
	DeliverablePreviewURL *string
	CompletedAt           *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// Deliverable represents a version of an order's submitted artwork.
type Deliverable struct {
	ID               uuid.UUID
	Version          int
	Decision         DeliverableDecision
	Comment          *string
	OriginalImageKey string
	PreviewImageKey  string
	CreatedAt        time.Time
	UpdatedAt        time.Time
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

// OrderParty contains the public information of the other participant
// in an order. ArtistReviewScore is populated when the other participant
// is an artist and is nil when the other participant is a customer.
type OrderParty struct {
	ID                uuid.UUID
	Name              string
	Email             string
	ArtistReviewScore *float64
}

// OrderDetailData contains all data required to build an OrderDetail.
// The repository returns both order participants, the artwork snapshot,
// and the order's deliverables; the usecase selects the opposite
// participant based on the authenticated user's role.
type OrderDetailData struct {
	Order           Order
	ArtworkSnapshot ArtworkSnapshot
	Customer        OrderParty
	Artist          OrderParty
	Deliverables    []Deliverable
}

// OrderDetail is the detailed view of an order for the authenticated
// participant. It includes the artwork snapshot captured when the order
// was created, the participant on the opposite side of the order, and
// all deliverable versions submitted for the order.
type OrderDetail struct {
	Order
	ArtworkSnapshot ArtworkSnapshot
	OtherParty      OrderParty
	Deliverables    []Deliverable
}
