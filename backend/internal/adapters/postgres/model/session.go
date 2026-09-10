package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID               uuid.UUID `bun:"id,pk"`
	UserID           uuid.UUID `bun:"user_id"`
	RefreshTokenHash string    `bun:"refresh_token_hash"`
	ExpiresAt        time.Time `bun:"expires_at"`
	CreatedAt        time.Time `bun:"created_at,nullzero"`
}
