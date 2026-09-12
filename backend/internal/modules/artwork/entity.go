// Package artwork owns artist portfolio artwork.
package artwork

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type Artwork struct {
	ID                  uuid.UUID
	ArtistID            uuid.UUID
	Name                string
	Category            string
	Styles              []string
	Description         string
	Samples             []Sample
	MinimumDeadlineDays int
	PriceSatang         int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Sample struct {
	ImageURL  string
	SortOrder int
}

type CreateInput struct {
	ArtistID            uuid.UUID
	Name                string
	Category            string
	Styles              []string
	Description         string
	SampleFiles         []io.Reader
	MinimumDeadlineDays int
	PriceSatang         int64
}

const MaxSampleImageSize = 5 * 1024 * 1024

type UpdateInput struct {
	ArtworkID         uuid.UUID
	DeletedSampleURLs []string
	CreateInput
}
