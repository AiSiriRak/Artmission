package review

import (
	"context"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

type reviewUsecase struct {
	repo ReviewRepository
}

// NewReviewUsecase creates a review usecase with the given repository
func NewReviewUsecase(repo ReviewRepository) ReviewUsecase {
	return &reviewUsecase{repo: repo}
}

// SubmitReview validates ownership, order status, and rating, then creates a review
func (u *reviewUsecase) SubmitReview(
	ctx context.Context,
	customerID uuid.UUID,
	input SubmitReviewInput,
) (*Review, error) {
	if customerID == uuid.Nil {
		return nil, apperror.Unauthorized("missing customer identity")
	}

	if input.OrderID == uuid.Nil {
		return nil, apperror.InvalidInput("order ID is required", nil)
	}

	if input.Rating < 1 || input.Rating > 5 {
		return nil, ErrInvalidRating
	}

	orderData, err := u.repo.GetOrderForReview(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if orderData == nil {
		return nil, ErrOrderNotFound
	}

	if orderData.CustomerID != customerID {
		return nil, ErrNotOrderOwner
	}

	if orderData.Status != order.StatusSuccess {
		return nil, ErrOrderNotCompleted
	}

	exists, err := u.repo.ExistsByOrderID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrReviewAlreadyExists
	}

	var comment *string
	if input.Comment != nil {
		trimmed := strings.TrimSpace(*input.Comment)
		if trimmed != "" {
			comment = &trimmed
		}
	}

	result := &Review{
		ID:         uuid.New(),
		OrderID:    orderData.ID,
		CustomerID: customerID,
		ArtistID:   orderData.ArtistID,
		Rating:     input.Rating,
		Comment:    comment,
		CreatedAt:  time.Now(),
	}

	if err := u.repo.Create(ctx, result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetArtistReviews retrieves a page of artist reviews, newest first.
func (u *reviewUsecase) GetArtistReviews(
	ctx context.Context,
	artistID uuid.UUID,
	query ListQuery,
) (*ReviewPage, error) {
	if artistID == uuid.Nil {
		return nil, apperror.InvalidInput("artist ID is required", nil)
	}

	if query.Limit == 0 {
		query.Limit = 20
	}
	if query.Limit < 1 || query.Limit > 100 {
		return nil, apperror.InvalidInput("limit must be between 1 and 100", nil)
	}
	if query.Offset < 0 {
		return nil, apperror.InvalidInput("offset must not be negative", nil)
	}

	return u.repo.ListByArtistID(ctx, artistID, query)
}

// retrieves an order's review for its customer or artist.
func (u *reviewUsecase) GetOrderReview(
	ctx context.Context,
	userID uuid.UUID,
	orderID uuid.UUID,
) (*ReviewDetail, error) {
	if userID == uuid.Nil {
		return nil, apperror.Unauthorized("missing user identity")
	}
	if orderID == uuid.Nil {
		return nil, apperror.InvalidInput("order ID is required", nil)
	}

	orderData, err := u.repo.GetOrderForReview(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if orderData == nil {
		return nil, ErrOrderNotFound
	}

	if userID != orderData.CustomerID && userID != orderData.ArtistID {
		return nil, ErrNotOrderParticipant
	}

	item, err := u.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrReviewNotFound
	}

	return item, nil
}

// GetReviewsByArtwork retrieves a page of artwork reviews, newest first.
func (u *reviewUsecase) GetReviewsByArtwork(
	ctx context.Context,
	artworkID uuid.UUID,
	query ListQuery,
) (*ReviewPage, error) {
	if artworkID == uuid.Nil {
		return nil, apperror.InvalidInput("artwork ID is required", nil)
	}

	if query.Limit == 0 {
		query.Limit = 20
	}
	if query.Limit < 1 || query.Limit > 100 {
		return nil, apperror.InvalidInput("limit must be between 1 and 100", nil)
	}
	if query.Offset < 0 {
		return nil, apperror.InvalidInput("offset must not be negative", nil)
	}

	return u.repo.ListByArtworkID(ctx, artworkID, query)
}
