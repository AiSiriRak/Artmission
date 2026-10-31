package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Review represents a customer review stored in the reviews table
type Review struct {
	bun.BaseModel `bun:"table:reviews,alias:r"`

	ID         uuid.UUID `bun:"id,pk"`
	OrderID    uuid.UUID `bun:"order_id"`
	CustomerID uuid.UUID `bun:"customer_id"`
	ArtistID   uuid.UUID `bun:"artist_id"`
	Rating     int       `bun:"rating"`
	Comment    *string   `bun:"comment"`
	CreatedAt  time.Time `bun:"created_at,nullzero"`
}