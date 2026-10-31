package review

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	ErrInvalidRating = apperror.InvalidInput("rating must be between 1 and 5",nil)
	ErrOrderNotFound = apperror.NotFound("order not found")
	ErrNotOrderOwner = apperror.Forbidden("only the order's customer can submit a review")
	ErrOrderNotCompleted = apperror.Conflict("order must be successful before submitting a review")
	ErrReviewAlreadyExists = apperror.Conflict("this order already has a review")
)