package artwork

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	ErrArtistNotFound      = apperror.NotFound("artist profile not found")
	ErrArtworkNotFound     = apperror.NotFound("artwork not found")
	ErrInvalidSampleImage  = apperror.InvalidInput("artwork samples must be JPEG, PNG, or WebP images", nil)
	ErrSampleImageTooLarge = apperror.InvalidInput("artwork sample image exceeds the 5 MiB limit", nil)
)
