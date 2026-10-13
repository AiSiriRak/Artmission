// Package artwork owns artist portfolio artwork.
package artwork

import (
	"io"
	"time"

	"github.com/google/uuid"
)

// Artwork represents an artwork in an artist's portfolio,
// including its category, styles, samples, pricing, and deadline requirements.
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

// Category represents a category that can be assigned to an artwork.
type Category struct {
	ID    uuid.UUID
	Label string
}

// Style represents an artistic style that can be assigned to an artwork.
type Style struct {
	ID    uuid.UUID
	Label string
}

// Sample represents an artwork sample image and its display order.
type Sample struct {
	ImageURL  string
	SortOrder int
}

// CreateInput contains the information required to create an artwork,
// including its category, styles, sample images, pricing, and deadline.
type CreateInput struct {
	ArtistID            uuid.UUID
	Name                string
	CategoryID          uuid.UUID
	StyleIDs            []uuid.UUID
	Description         string
	SampleFiles         []io.Reader
	MinimumDeadlineDays int
	PriceSatang         int64
}

// UpdateInput contains the information required to update an existing artwork.
// It embeds CreateInput to reuse the artwork fields and includes the sample
// image URLs to delete.
type UpdateInput struct {
	ArtworkID         uuid.UUID
	DeletedSampleURLs []string
	CreateInput
}

const (
	// MaxSampleImageSize is the maximum allowed sample image size in bytes.
	MaxSampleImageSize = 5 * 1024 * 1024
)

// SearchSort defines the supported sort orders for artwork search results.
type SearchSort string

const (
	SearchSortNameAsc         SearchSort = "name_asc"
	SearchSortPriceAsc        SearchSort = "price_asc"
	SearchSortPriceDesc       SearchSort = "price_desc"
	SearchSortReviewScoreAsc  SearchSort = "review_score_asc"
	SearchSortReviewScoreDesc SearchSort = "review_score_desc"
	DefaultSearchSort         SearchSort = SearchSortNameAsc
	SearchPageSize                       = 20
)

// IsValid reports whether s is one of the supported artwork search sort orders.
func (s SearchSort) IsValid() bool {
	switch s {
	case SearchSortNameAsc, SearchSortPriceAsc, SearchSortPriceDesc, SearchSortReviewScoreAsc, SearchSortReviewScoreDesc:
		return true
	default:
		return false
	}
}

// ArtistSummary contains the artist information displayed alongside
// an artwork in search results.
type ArtistSummary struct {
	ID              uuid.UUID
	Name            string
	ProfileImageKey *string
	ProfileURL      *string
	ReviewScore     *float64
}

// SearchItem represents an artwork and its artist information
// in the search results.
type SearchItem struct {
	Artwork Artwork
	Artist  ArtistSummary
}

// SearchQuery contains the filters, sort order, and pagination parameters
// for searching artworks.
type SearchQuery struct {
	ArtistName     string
	Category       string
	Styles         []string
	MinPriceSatang *int64
	MaxPriceSatang *int64
	MinReviewScore *float64
	Sort           SearchSort
	Page           int
	Limit          int
	Offset         int
}

// SearchPage contains a page of artwork search results and the total
// number of matching artworks.
type SearchPage struct {
	Items []SearchItem
	Total int
	Page  int
}
