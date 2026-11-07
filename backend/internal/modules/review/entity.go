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
	Username        string
	ProfileImageKey *string
	ProfileImageURL *string
	Order           string
	ArtworkName     string
}

type SubmitReviewInput struct {
	OrderID uuid.UUID
	Rating  int
	Comment *string
}

type ListQuery struct {
	Limit  int
	Offset int
}

type ReviewPage struct {
	Reviews []ReviewDetail
	Total   int
}
