package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Category struct {
	bun.BaseModel `bun:"table:categories,alias:cat"`

	ID        uuid.UUID `bun:"id,pk"`
	Label     string    `bun:"label"`
	CreatedAt time.Time `bun:"created_at,nullzero"`
}

type Style struct {
	bun.BaseModel `bun:"table:styles,alias:sty"`

	ID        uuid.UUID `bun:"id,pk"`
	Label     string    `bun:"label"`
	CreatedAt time.Time `bun:"created_at,nullzero"`
}

type Artwork struct {
	bun.BaseModel `bun:"table:artworks,alias:art"`

	ID                  uuid.UUID `bun:"id,pk"`
	ArtistID            uuid.UUID `bun:"artist_id"`
	CategoryID          uuid.UUID `bun:"category_id"`
	Name                string    `bun:"name"`
	Category            string    `bun:"category,scanonly"`
	Description         string    `bun:"description"`
	PriceSatang         int64     `bun:"price_satang"`
	MinimumDeadlineDays int       `bun:"minimum_deadline_days"`
	CreatedAt           time.Time `bun:"created_at,nullzero"`
	UpdatedAt           time.Time `bun:"updated_at,nullzero"`
}

type ArtworkImage struct {
	bun.BaseModel `bun:"table:artwork_images,alias:ai"`

	ID               uuid.UUID `bun:"id,pk"`
	ArtworkID        uuid.UUID `bun:"artwork_id"`
	OriginalImageKey string    `bun:"original_image_key"`
	PreviewImageKey  string    `bun:"preview_image_key"`
	SortOrder        int       `bun:"sort_order"`
	CreatedAt        time.Time `bun:"created_at,nullzero"`
}

type ArtworkStyle struct {
	bun.BaseModel `bun:"table:artwork_styles,alias:aws"`

	ArtworkID uuid.UUID `bun:"artwork_id,pk"`
	StyleID   uuid.UUID `bun:"style_id,pk"`
}
