package artist

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ProfileUsecase interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, description string) error
}

type profileUsecase struct {
	repo ProfileRepository
}

func NewProfileUsecase(repo ProfileRepository) ProfileUsecase {
	return &profileUsecase{repo: repo}
}

func (u *profileUsecase) CreateProfile(ctx context.Context, userID uuid.UUID, description string) error {
	now := time.Now()
	return u.repo.Create(ctx, &Profile{
		UserID:      userID,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}
