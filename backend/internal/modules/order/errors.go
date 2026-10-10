package order

import (
	"fmt"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
)

var (
	ErrOrderNotFound                     = apperror.NotFound("order not found")
	ErrUnsupportedParticipantRole        = apperror.Forbidden("unsupported participant role")
	ErrMissingParticipantID              = apperror.InvalidInput("missing participant id", nil)
	ErrMissingOrderID                    = apperror.InvalidInput("missing order id", nil)
	ErrInvalidOrderStatus                = apperror.Conflict("order status does not allow this operation")
	ErrFailedToPresignDeliverablePreview = apperror.Internal("failed to presign deliverable preview image", nil)
	ErrInvalidDeliverableImage           = apperror.InvalidInput("deliverable image must be a JPEG, PNG, or WebP image", nil)
	ErrDeliverableImageTooLarge          = apperror.InvalidInput(fmt.Sprintf("deliverable image exceeds the %d MiB limit", MaxDeliverableImageSize/(1024*1024)), nil)
	ErrMaxDeliverableVersionsReached     = apperror.Conflict("maximum number of deliverable versions reached")
	ErrArtistGetterUnavailable           = apperror.Internal("artist identity provider is not configured", nil)
	ErrArtistNotFound                    = apperror.NotFound("artist not found")
)
