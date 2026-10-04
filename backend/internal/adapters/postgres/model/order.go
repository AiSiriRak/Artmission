package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Order struct {
	bun.BaseModel `bun:"table:orders,alias:o"`

	ID                  uuid.UUID  `bun:"id,pk"`
	CustomerID          uuid.UUID  `bun:"customer_id"`
	ArtistID            uuid.UUID  `bun:"artist_id"`
	Name                string     `bun:"name"`
	ArtworkID           *uuid.UUID `bun:"artwork_id"`
	PriceSatangOrder    int64      `bun:"price_satang_order"`
	CustomerDescription string     `bun:"customer_description"`
	DeadlineAt          time.Time  `bun:"deadline_at"`
	Status              string     `bun:"status"`
	CompletedAt         *time.Time `bun:"completed_at"`
	CreatedAt           time.Time  `bun:"created_at,nullzero"`
	UpdatedAt           time.Time  `bun:"updated_at,nullzero"`
}

type OrderDeliverable struct {
	bun.BaseModel `bun:"table:order_deliverables,alias:od"`

	ID               uuid.UUID `bun:"id,pk"`
	OrderID          uuid.UUID `bun:"order_id"`
	Version          int       `bun:"version"`
	Decision         *string   `bun:"decision,nullzero"`
	Comment          *string   `bun:"comment,nullzero"`
	OriginalImageKey string    `bun:"original_image_key"`
	PreviewImageKey  string    `bun:"preview_image_key"`
	CreatedAt        time.Time `bun:"created_at,nullzero"`
}
