package user

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	ErrEmailTaken                = apperror.Conflict("email is already in use")
	ErrInvalidRole               = apperror.InvalidInput("role must be customer or artist", nil)
	ErrUserNotFound              = apperror.NotFound("user not found")
	ErrInvalidCredential         = apperror.Unauthorized("invalid email or password")
	ErrArtistDescriptionRequired = apperror.InvalidInput("artist description is required", nil)
	ErrArtistFieldsNotAllowed    = apperror.InvalidInput("artist fields are only allowed when role is artist", nil)
	ErrBankAccountRequired       = apperror.InvalidInput("bank name and account number are required", nil)
)
