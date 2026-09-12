package artist

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	ErrProfileNotFound               = apperror.NotFound("artist profile not found")
	ErrNoProfileChanges              = apperror.InvalidInput("at least one profile field must be updated", nil)
	ErrConflictingProfileImageChange = apperror.InvalidInput("profile_image and remove_profile_image cannot be used together", nil)
	ErrProfileImageTooLarge          = apperror.InvalidInput("profile image must not exceed 5 MiB", nil)
	ErrInvalidProfileImage           = apperror.InvalidInput("profile image must be a JPEG, PNG, or WebP image", nil)
)
