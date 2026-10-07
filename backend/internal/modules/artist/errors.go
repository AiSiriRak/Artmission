package artist

import (
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
)

var (
	ErrProfileNotFound               = apperror.NotFound("artist profile not found")
	ErrNoProfileChanges              = user.ErrNoProfileChanges
	ErrConflictingProfileImageChange = user.ErrConflictingProfileImageChange
	ErrProfileImageTooLarge          = user.ErrProfileImageTooLarge
	ErrInvalidProfileImage           = user.ErrInvalidProfileImage
)
