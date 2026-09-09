// Package apperror defines a small, framework-agnostic error taxonomy for
// usecases and repositories. Domain code (internal/modules/**) reports
// errors in terms of this package instead of importing the HTTP framework;
// the presentation layer (internal/handler/rest) is the only place that
// translates a Code into a transport-specific status.
package apperror

import "fmt"

type Code string

const (
	CodeInvalidInput Code = "INVALID_INPUT"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeInternal     Code = "INTERNAL"
)

// Error is a domain error carrying a Code the presentation layer maps to a
// transport status, a user-facing Message, and an optional wrapped cause
// for logging/introspection.
type Error struct {
	Code    Code
	Message string
	Cause   error
}

func New(code Code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

func InvalidInput(message string, cause error) *Error { return New(CodeInvalidInput, message, cause) }
func Unauthorized(message string) *Error              { return New(CodeUnauthorized, message, nil) }
func Forbidden(message string) *Error                 { return New(CodeForbidden, message, nil) }
func NotFound(message string) *Error                  { return New(CodeNotFound, message, nil) }
func Conflict(message string) *Error                  { return New(CodeConflict, message, nil) }
func Internal(message string, cause error) *Error     { return New(CodeInternal, message, cause) }
