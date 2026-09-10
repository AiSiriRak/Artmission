package artist

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	ErrProfileNotFound        = apperror.NotFound("artist profile not found")
	ErrPriceMustBeNonNegative = apperror.InvalidInput("artist prices must not be negative", nil)
	ErrInvalidPriceRange      = apperror.InvalidInput("minimum price must not exceed maximum price", nil)
	ErrStyleNotFound          = apperror.InvalidInput("one or more styles do not exist", nil)
)
