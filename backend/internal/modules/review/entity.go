package review

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID         uuid.UUID
	OrderID    uuid.UUID
	CustomerID uuid.UUID
	ArtistID   uuid.UUID
	Rating     int
	Comment    *string
	CreatedAt  time.Time
}

type ReviewDetail struct {
	Review
	Username string
	Order    string
}

