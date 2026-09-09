package artist

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ProfileUsecase interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, description string) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (*Profile, error)
}

type ProfileRepository interface {
	Create(ctx context.Context, p *Profile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error)
	UpdateByUserID(ctx context.Context, userID uuid.UUID, in ProfileUpdate) error
	CountStylesByIDs(ctx context.Context, styleIDs []uuid.UUID) (int, error)
	ReplaceStyles(ctx context.Context, userID uuid.UUID, styleIDs []uuid.UUID) error
}

type Transactioner interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type UpdateProfileInput struct {
	Description    string
	StyleIDs       []uuid.UUID
	MinPriceSatang int64
	MaxPriceSatang int64
}

type ProfileUpdate struct {
	Description    string
	MinPriceSatang int64
	MaxPriceSatang int64
	UpdatedAt      time.Time
}
