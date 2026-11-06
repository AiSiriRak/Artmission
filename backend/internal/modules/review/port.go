package review

import (
	"context"

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
		query ListQuery,
	) (*ReviewPage, error)

	GetOrderReview(
		ctx context.Context,
		userID uuid.UUID,
		orderID uuid.UUID,
	) (*ReviewDetail, error)

	GetReviewsByArtwork(
		ctx context.Context,
		artworkID uuid.UUID,
		query ListQuery,
	) (*ReviewPage, error)
}

type ReviewRepository interface {
	GetOrderForReview(ctx context.Context, orderID uuid.UUID) (*OrderForReview, error)
	ExistsByOrderID(ctx context.Context, orderID uuid.UUID) (bool, error)
	Create(ctx context.Context, review *Review) error
	ListByArtistID(ctx context.Context, artistID uuid.UUID, query ListQuery) (*ReviewPage, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*ReviewDetail, error)
	ListByArtworkID(ctx context.Context, artworkID uuid.UUID, query ListQuery) (*ReviewPage, error)
}
