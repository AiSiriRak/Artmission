package auth

import (
	"context"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/google/uuid"
)

// AuthUsecase defines the authentication operations available to other modules.
type AuthUsecase interface {
	Login(ctx context.Context, email string, password string) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error

	// Authenticate is used only by the HTTP auth middleware to validate
	// an incoming request; it is not part of any user-facing feature.
	Authenticate(ctx context.Context, accessToken string) (*TokenClaims, error)
}

// SessionRepository defines persistence operations for authentication sessions.
type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	FindByID(ctx context.Context, id uuid.UUID) (*Session, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

// TokenIssuer signs/parses the JWTs carrying TokenClaims. Modeled as a port
// (unlike password hashing) because unit tests fake it to avoid real
// signing and to assert claim round-tripping deterministically.
type TokenIssuer interface {
	GenerateAccessToken(claims TokenClaims, expiresAt time.Time) (string, error)
	GenerateRefreshToken(claims TokenClaims, expiresAt time.Time) (string, error)
	ParseAccessToken(token string) (*TokenClaims, error)
	ParseRefreshToken(token string) (*TokenClaims, error)
}

// UserIdentity is the subset of user operations auth needs: credential
// check on login, and rehydration on refresh.
type UserIdentity interface {
	Authenticate(ctx context.Context, email string, password string) (*user.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}
