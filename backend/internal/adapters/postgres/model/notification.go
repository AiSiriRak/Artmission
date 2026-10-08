package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Notification struct {
	bun.BaseModel `bun:"table:notifications,alias:n"`

	ID        uuid.UUID `bun:"id,pk"`
	UserID    uuid.UUID `bun:"user_id"`
	Title     string    `bun:"title"`
	Message   string    `bun:"message"`
	Type      string    `bun:"type"`
	IsRead    bool      `bun:"is_read"`
	CreatedAt time.Time `bun:"created_at,nullzero"`
}
