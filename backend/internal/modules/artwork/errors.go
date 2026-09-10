package artwork

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	ErrArtistNotFound  = apperror.NotFound("artist profile not found")
	ErrArtworkNotFound = apperror.NotFound("artwork not found")
)
