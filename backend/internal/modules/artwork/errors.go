package artwork

import (
	"fmt"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
)

var (
	ErrArtistNotFound      = apperror.NotFound("artist profile not found")
	ErrArtworkNotFound     = apperror.NotFound("artwork not found")
	ErrInvalidSampleImage  = apperror.InvalidInput("artwork samples must be JPEG, PNG, or WebP images", nil)
	ErrSampleImageTooLarge = apperror.InvalidInput(fmt.Sprintf("artwork sample image exceeds the %d MiB limit", MaxSampleImageSize/(1024*1024)), nil)
	ErrSampleNotOwned      = apperror.InvalidInput("deleted artwork sample does not belong to the artwork", nil)
)
