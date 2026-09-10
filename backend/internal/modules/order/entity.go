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
