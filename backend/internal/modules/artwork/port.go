package artwork

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type Usecase interface {
	ListByArtistID(ctx context.Context, artistID uuid.UUID) ([]Artwork, error)
	Create(ctx context.Context, input CreateInput) (*Artwork, error)
	Update(ctx context.Context, input UpdateInput) (*Artwork, error)
	Delete(ctx context.Context, artistID, artworkID uuid.UUID) error
}

type Repository interface {
	ListByArtistID(ctx context.Context, artistID uuid.UUID) ([]Artwork, error)
	FindOrCreateCategory(ctx context.Context, label string) (uuid.UUID, error)
	FindOrCreateStyles(ctx context.Context, labels []string) ([]uuid.UUID, error)
	Create(ctx context.Context, artwork *Artwork, categoryID uuid.UUID, styleIDs []uuid.UUID) error
	UpdateOwnedBy(ctx context.Context, artwork *Artwork, categoryID uuid.UUID, styleIDs []uuid.UUID) ([]string, error)
	DeleteOwnedBy(ctx context.Context, artworkID, artistID uuid.UUID) ([]string, error)
}

type ObjectStorage interface {
	UploadPublic(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	DeletePublicURL(ctx context.Context, rawURL string) error
	PublicURL(key string) string
}

type Transactioner interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}
