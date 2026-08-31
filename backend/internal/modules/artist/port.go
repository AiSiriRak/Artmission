package artist

import (
	"context"

	"github.com/google/uuid"
)

type ProfileUsecase interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, description string) error
}

type ProfileRepository interface {
	Create(ctx context.Context, p *Profile) error
}
