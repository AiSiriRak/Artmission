package artist

import (
	"context"
	"io"
	"time"

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
	UploadPublic(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	DeletePublic(ctx context.Context, key string) error
	PublicURL(key string) string
}

type UpdateProfileInput struct {
	Description        *string
	ProfileImage       io.Reader
	RemoveProfileImage bool
}

type ProfileUpdate struct {
	DescriptionSet     bool
	Description        *string
	ProfileImageKeySet bool
	ProfileImageKey    *string
	UpdatedAt          time.Time
}
