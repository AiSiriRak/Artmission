// Package artist owns the artist profile (1:1 with a user whose role is artist).
package artist

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID         uuid.UUID
	ArtistName     string
	Description    string
	Categories     []Category
	Styles         []Style
	MinPriceSatang *int64
	MaxPriceSatang *int64
	ReviewScore    *float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Category struct {
	ID    uuid.UUID
	Label string
}

type Style struct {
	ID    uuid.UUID
	Label string
}
