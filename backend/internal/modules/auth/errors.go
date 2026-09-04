package auth

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	ErrInvalidCredential = apperror.Unauthorized("invalid username or password")
	ErrInvalidToken      = apperror.Unauthorized("invalid or expired token")
	ErrSessionNotFound   = apperror.Unauthorized("session is invalid or has expired")
)
