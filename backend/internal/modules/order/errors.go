package order

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	ErrOrderNotFound              = apperror.NotFound("order not found")
	ErrUnsupportedParticipantRole = apperror.Forbidden("unsupported participant role")
	ErrMissingParticipantID       = apperror.InvalidInput("missing participant id", nil)
	ErrMissingOrderID             = apperror.InvalidInput("missing order id", nil)
)
