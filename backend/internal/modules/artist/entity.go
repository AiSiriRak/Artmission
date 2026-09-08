// Package artist owns the artist profile (1:1 with a user whose role is artist).
package artist

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID      uuid.UUID
	Description string
	ReviewScore *float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
