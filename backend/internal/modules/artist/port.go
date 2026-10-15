package artist

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type ProfileUsecase interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, description *string) error
	GetProfile(ctx context.Context, userID uuid.UUID, query ProfileQuery) (*Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (*Profile, error)
}

type ProfileRepository interface {
	Create(ctx context.Context, p *Profile) error
	GetByUserID(ctx context.Context, userID uuid.UUID, query ProfileQuery) (*Profile, error)
	UpdateByUserID(ctx context.Context, userID uuid.UUID, in ProfileUpdate) error
}

type ObjectStorage interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}
