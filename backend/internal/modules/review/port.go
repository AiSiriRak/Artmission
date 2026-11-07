package review

import (
	"context"

	"github.com/google/uuid"
)

type ProfileImageURLResolver interface {
	PublicURL(key string) string
}

type UsecaseOption func(*reviewUsecase)

func WithProfileImageURLResolver(resolver ProfileImageURLResolver) UsecaseOption {
	return func(u *reviewUsecase) {
		u.profileImageURLs = resolver
	}
}

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
	// SubmitReview checks order eligibility, saves the review, and refreshes the score atomically.
	SubmitReview(ctx context.Context, customerID uuid.UUID, input SubmitReviewInput) (*Review, error)
	ListByArtistID(ctx context.Context, artistID uuid.UUID, query ListQuery) (*ReviewPage, error)
	GetReviewByOrderID(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*ReviewDetail, error)
	ListByArtworkID(ctx context.Context, artworkID uuid.UUID, query ListQuery) (*ReviewPage, error)
}
