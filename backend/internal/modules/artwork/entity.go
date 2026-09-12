// Package artwork owns artwork portfolio retrieval.
package artwork

import (
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
	Samples             []Sample
	MinimumDeadlineDays int
	PriceSatang         int64
}
