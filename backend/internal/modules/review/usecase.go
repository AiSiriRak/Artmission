package review

import (
	"context"
	"strings"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

type reviewUsecase struct {
	repo             ReviewRepository
	profileImageURLs ProfileImageURLResolver
}

// NewReviewUsecase creates a review usecase with the given repository
func NewReviewUsecase(repo ReviewRepository, options ...UsecaseOption) ReviewUsecase {
	u := &reviewUsecase{repo: repo}
	for _, option := range options {
		option(u)
	}
	return u
}

// SubmitReview validates input and normalizes the optional comment before submission.
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

	var comment *string
	if input.Comment != nil {
		trimmed := strings.TrimSpace(*input.Comment)
		if trimmed != "" {
			comment = &trimmed
		}
	}

	input.Comment = comment
	return u.repo.SubmitReview(ctx, customerID, input)
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

	page, err := u.repo.ListByArtistID(ctx, artistID, query)
	if err != nil {
		return nil, err
	}
	if err := u.resolveProfileImageURLs(page); err != nil {
		return nil, err
	}
	return page, nil
}

// GetOrderReview validates identities before requesting the order's review.
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

	item, err := u.repo.GetReviewByOrderID(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	if err := u.resolveProfileImageURL(item); err != nil {
		return nil, err
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

	page, err := u.repo.ListByArtworkID(ctx, artworkID, query)
	if err != nil {
		return nil, err
	}
	if err := u.resolveProfileImageURLs(page); err != nil {
		return nil, err
	}
	return page, nil
}

func (u *reviewUsecase) resolveProfileImageURLs(page *ReviewPage) error {
	for i := range page.Reviews {
		if err := u.resolveProfileImageURL(&page.Reviews[i]); err != nil {
			return err
		}
	}
	return nil
}

func (u *reviewUsecase) resolveProfileImageURL(item *ReviewDetail) error {
	if item == nil || item.ProfileImageKey == nil || *item.ProfileImageKey == "" {
		return nil
	}
	if u.profileImageURLs == nil {
		return ErrProfileImageURLResolverUnavailable
	}
	url := u.profileImageURLs.PublicURL(*item.ProfileImageKey)
	item.ProfileImageURL = &url
	return nil
}
