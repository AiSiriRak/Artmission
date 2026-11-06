package review

import (
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
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

type SubmitReviewInput struct {
	OrderID uuid.UUID
	Rating  int
	Comment *string
}

type OrderForReview struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	ArtistID   uuid.UUID
	Status     order.Status
}

type ListQuery struct {
	Limit  int
	Offset int
}

type ReviewPage struct {
	Reviews []ReviewDetail
	Total   int
}
