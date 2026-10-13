// Package artist owns the artist profile (1:1 with a user whose role is artist).
package artist

import (
	"io"
	"time"

	"github.com/google/uuid"
)

// Profile represents an artist's profile and its associated information,
// including categories, styles, pricing, and reviews.
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
	// DefaultReviewLimit is the default maximum number of reviews returned
	// when no limit is specified.
	DefaultReviewLimit = 20

	// MaxReviewLimit is the maximum number of reviews that can be returned
	// in a single request.
	MaxReviewLimit = 100

	// MaxProfileImageSize is the maximum allowed profile image size in bytes.
	MaxProfileImageSize = 5 * 1024 * 1024
)

// Category represents a category associated with an artist.
type Category struct {
	ID    uuid.UUID
	Label string
}

// Style represents an artistic style associated with an artist.
type Style struct {
	ID    uuid.UUID
	Label string
}

// Review represents a customer's review of an artist's completed order.
type Review struct {
	Username string
	Order    string
	Rating   int
}

// ProfileQuery contains pagination parameters for retrieving artist profiles
// or their associated data.
type ProfileQuery struct {
	Limit  int
	Offset int
}

// ProfileQuery contains pagination parameters for retrieving artist profiles
// or their associated data.
type UpdateProfileInput struct {
	Description        *string
	ProfileImage       io.Reader
	RemoveProfileImage bool
}

// ProfileUpdate contains the fields to update in an artist's profile.
// The Set fields indicate whether the corresponding value should be updated,
// allowing nil values to be distinguished from fields that are not provided.
type ProfileUpdate struct {
	DescriptionSet     bool
	Description        *string
	ProfileImageKeySet bool
	ProfileImageKey    *string
	UpdatedAt          time.Time
}
