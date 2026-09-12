package artwork

import (
	"context"

	"github.com/google/uuid"
)

type Usecase interface {
	ListByArtistID(ctx context.Context, artistID uuid.UUID) ([]Artwork, error)
	Create(ctx context.Context, input CreateInput) (*Artwork, error)
	Delete(ctx context.Context, artistID, artworkID uuid.UUID) error
}

type Repository interface {
	ListByArtistID(ctx context.Context, artistID uuid.UUID) ([]Artwork, error)
	FindOrCreateCategory(ctx context.Context, label string) (uuid.UUID, error)
	FindOrCreateStyles(ctx context.Context, labels []string) ([]uuid.UUID, error)
	Create(ctx context.Context, artwork *Artwork, categoryID uuid.UUID, styleIDs []uuid.UUID) error
	DeleteOwnedBy(ctx context.Context, artworkID, artistID uuid.UUID) error
}

type Transactioner interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}
