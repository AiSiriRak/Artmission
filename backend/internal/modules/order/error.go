package order

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var (
	// ErrOrderNotFound indicates the order does not exist or is not visible to the caller.
	ErrOrderNotFound = apperror.NotFound("order not found")

	// ErrUnsupportedParticipantRole indicates the caller's role cannot access orders.
	ErrUnsupportedParticipantRole = apperror.Forbidden("unsupported participant role")

	// ErrMissingParticipantID indicates the authenticated participant ID is empty.
	ErrMissingParticipantID = apperror.InvalidInput("missing participant id", nil)

	// ErrMissingOrderID indicates the requested order ID is empty.
	ErrMissingOrderID = apperror.InvalidInput("missing order id", nil)
)
