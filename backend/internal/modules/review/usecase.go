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