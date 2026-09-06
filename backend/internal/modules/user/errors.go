package user

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	ErrEmailTaken                = apperror.Conflict("email is already in use")
	ErrInvalidRole               = apperror.InvalidInput("role must be customer or artist", nil)
	ErrUserNotFound              = apperror.NotFound("user not found")
	ErrInvalidCredential         = apperror.Unauthorized("invalid email or password")
	ErrArtistDescriptionRequired = apperror.InvalidInput("artist description is required", nil)
	ErrArtistFieldsNotAllowed    = apperror.InvalidInput("artist fields are only allowed when role is artist", nil)
	ErrBankAccountRequired       = apperror.InvalidInput("bank name, account holder name, and account number are required", nil)
	ErrBankAccountNotAllowed     = apperror.Forbidden("bank account is only available for customer and artist accounts")
	ErrInvalidUsername           = apperror.InvalidInput("username must be between 3 and 20 characters", nil)
	ErrPasswordFieldsRequired    = apperror.InvalidInput("old password and new password must be provided together", nil)
	ErrInvalidNewPassword        = apperror.InvalidInput("new password must be between 8 and 16 characters", nil)
	ErrInvalidCurrentPassword    = apperror.Unauthorized("old password is incorrect")
)
