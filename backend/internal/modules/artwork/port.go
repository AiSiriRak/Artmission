package artwork

import (
	"context"

	"github.com/google/uuid"
)

type Usecase interface {
	ListByArtistID(ctx context.Context, artistID uuid.UUID) ([]Artwork, error)
}

type Repository interface {
	ListByArtistID(ctx context.Context, artistID uuid.UUID) ([]Artwork, error)
}
