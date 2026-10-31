package notification

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

// ErrNotFound hides whether a notification is missing or belongs to another recipient.
var ErrNotFound = apperror.NotFound("notification not found")

var (
	ErrInvalidEventType      = apperror.InvalidInput("unsupported notification event type", nil)
	ErrInvalidOrderStatus    = apperror.InvalidInput("notification order status must be PENDING or NOT_PAID", nil)
	ErrMissingUserID         = apperror.InvalidInput("notification recipient ID is required", nil)
	ErrMissingNotificationID = apperror.InvalidInput("notification ID is required", nil)
	ErrMissingTitle          = apperror.InvalidInput("notification title must not be blank", nil)
	ErrMissingMessage        = apperror.InvalidInput("notification message must not be blank", nil)
	ErrInvalidType           = apperror.InvalidInput("notification type must be ORDER_UPDATE, REVIEW, or SYSTEM", nil)
	ErrInvalidLimit          = apperror.InvalidInput("notification limit must be between 1 and 100", nil)
	ErrInvalidOffset         = apperror.InvalidInput("notification offset must not be negative", nil)
)
