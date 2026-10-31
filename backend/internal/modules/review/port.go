package review

import (
	"context"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/google/uuid"
)

type ReviewUsecase interface {
	SubmitReview(
		ctx context.Context,
		customerID uuid.UUID,
		input SubmitReviewInput,
	) (*Review, error)

	GetArtistReviews(
	ctx context.Context,
	artistID uuid.UUID,
	) ([]ReviewDetail, error)
}

type ReviewRepository interface {
	GetOrderForReview(ctx context.Context, orderID uuid.UUID) (*OrderForReview, error)
	ExistsByOrderID(ctx context.Context, orderID uuid.UUID) (bool, error)
	Create(ctx context.Context, review *Review) error
	
	ListByArtistID(
	ctx context.Context,
	artistID uuid.UUID,
	) ([]ReviewDetail, error)
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