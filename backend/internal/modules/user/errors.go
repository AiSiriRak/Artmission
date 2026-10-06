package user

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	ErrEmailTaken                    = apperror.Conflict("email is already in use")
	ErrInvalidRole                   = apperror.InvalidInput("role must be customer or artist", nil)
	ErrUserNotFound                  = apperror.NotFound("user not found")
	ErrInvalidCredential             = apperror.Unauthorized("invalid email or password")
	ErrArtistFieldsNotAllowed        = apperror.InvalidInput("artist fields are only allowed when role is artist", nil)
	ErrBankAccountRequired           = apperror.InvalidInput("bank name, account holder name, and account number are required", nil)
	ErrBankAccountNotAllowed         = apperror.Forbidden("bank account is only available for customer and artist accounts")
	ErrBankAccountNotFound           = apperror.NotFound("bank account not found")
	ErrPasswordFieldsRequired        = apperror.InvalidInput("old password and new password must be provided together", nil)
	ErrInvalidCurrentPassword        = apperror.Unauthorized("old password is incorrect")
	ErrActiveOrders                  = apperror.Conflict("account cannot be deleted while an order is pending, awaiting payment, or in progress")
	ErrNoProfileChanges              = apperror.InvalidInput("at least one profile field must be updated", nil)
	ErrConflictingProfileImageChange = apperror.InvalidInput("profile_image and remove_profile_image cannot be used together", nil)
	ErrProfileImageTooLarge          = apperror.InvalidInput("profile image must not exceed 5 MiB", nil)
	ErrInvalidProfileImage           = apperror.InvalidInput("profile image must be a JPEG, PNG, or WebP image", nil)
)
