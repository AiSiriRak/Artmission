// Package artist owns the artist profile (1:1 with a user whose role is artist).
package artist

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID          uuid.UUID
	ArtistName      string
	ProfileImageKey *string
	ProfileURL      *string
	Description     *string
	Categories      []Category
	Styles          []Style
	MinPriceSatang  *int64
	MaxPriceSatang  *int64
	ReviewScore     *float64
	Reviews         []Review
	Total           int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

const (
	DefaultReviewLimit  = 20
	MaxReviewLimit      = 100
	MaxProfileImageSize = 5 * 1024 * 1024
)

type Category struct {
	ID    uuid.UUID
	Label string
}

type Style struct {
	ID    uuid.UUID
	Label string
}

type Review struct {
	Username string
	Order    string
	Rating   int
}

type ProfileQuery struct {
	Limit  int
	Offset int
}
