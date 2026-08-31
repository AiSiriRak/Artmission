// Package auth owns session/token concerns: issuing, validating, and
// revoking sessions.
package auth

import (
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/google/uuid"
)

// Session is the server-side record backing a refresh token. Its existence
// (and non-expiry) is what makes logout an immediate, real revocation
// instead of "wait for the JWT to expire": every access-token validation
// checks this row too (see AuthUsecase.Authenticate).
type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

// TokenClaims is embedded in both the access and refresh JWTs.
type TokenClaims struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	Role      user.Role
}
