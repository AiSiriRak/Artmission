package user

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	ErrUsernameTaken     = apperror.Conflict("username is already in use")
	ErrEmailTaken        = apperror.Conflict("email is already in use")
	ErrInvalidRole       = apperror.InvalidInput("role must be customer or artist", nil)
	ErrUserNotFound      = apperror.NotFound("user not found")
	ErrInvalidCredential = apperror.Unauthorized("invalid username or password")
)
