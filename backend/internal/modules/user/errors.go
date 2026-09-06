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
	ErrPasswordFieldsRequired    = apperror.InvalidInput("old password and new password must be provided together", nil)
	ErrInvalidCurrentPassword    = apperror.Unauthorized("old password is incorrect")
	ErrActiveOrders              = apperror.Conflict("account cannot be deleted while an order is pending, awaiting payment, or in progress")
)
