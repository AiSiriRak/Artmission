package rest

import (
	"errors"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/danielgtaylor/huma/v2"
)

// mapAppError translates a framework-agnostic apperror.Error (or an
// unrecognized error) into a huma.StatusError. This is the only place in
// the codebase a domain apperror.Code becomes an HTTP status.
func mapAppError(err error) error {
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		return huma.Error500InternalServerError("internal server error", err)
	}

	switch appErr.Code {
	case apperror.CodeInvalidInput:
		return huma.Error400BadRequest(appErr.Message)
	case apperror.CodeUnauthorized:
		return huma.Error401Unauthorized(appErr.Message)
	case apperror.CodeForbidden:
		return huma.Error403Forbidden(appErr.Message)
	case apperror.CodeNotFound:
		return huma.Error404NotFound(appErr.Message)
	case apperror.CodeConflict:
		return huma.Error409Conflict(appErr.Message)
	default:
		return huma.Error500InternalServerError(appErr.Message, appErr.Cause)
	}
}
