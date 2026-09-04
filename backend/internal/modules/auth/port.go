package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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
