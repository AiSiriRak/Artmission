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

type Category struct {
	ID    uuid.UUID
	Label string
}

type Style struct {
	ID    uuid.UUID
	Label string
}

type Sample struct {
	ImageURL  string
	SortOrder int
}

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

const MaxSampleImageSize = 5 * 1024 * 1024

type UpdateInput struct {
	ArtworkID         uuid.UUID
	DeletedSampleURLs []string
	CreateInput
}

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

func (s SearchSort) IsValid() bool {
	switch s {
	case SearchSortNameAsc, SearchSortPriceAsc, SearchSortPriceDesc, SearchSortReviewScoreAsc, SearchSortReviewScoreDesc:
		return true
	default:
		return false
	}
}

type ArtistSummary struct {
	ID              uuid.UUID
	Name            string
	ProfileImageKey *string
	ProfileURL      *string
	ReviewScore     *float64
}

type SearchItem struct {
	Artwork Artwork
	Artist  ArtistSummary
}

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

type SearchPage struct {
	Items []SearchItem
	Total int
	Page  int
}

type ArtworkDetail struct {
	ID                  uuid.UUID
	ArtistID            uuid.UUID
	Name                string
	CategoryID          uuid.UUID
	StyleIDs            []uuid.UUID
	Description         string
	Samples             []Sample
	MinimumDeadlineDays int
	PriceSatang         int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
