package artist

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

type profileUsecase struct {
	repo ProfileRepository
	tx   Transactioner
}

func NewProfileUsecase(repo ProfileRepository, tx Transactioner) ProfileUsecase {
	return &profileUsecase{repo: repo, tx: tx}
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

func (u *profileUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	return u.repo.GetByUserID(ctx, userID)
}

func (u *profileUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (*Profile, error) {
	description := strings.TrimSpace(in.Description)
	if description == "" {
		return nil, ErrDescriptionRequired
	}
	if in.MinPriceSatang < 0 || in.MaxPriceSatang < 0 {
		return nil, ErrPriceMustBeNonNegative
	}
	if in.MinPriceSatang > in.MaxPriceSatang {
		return nil, ErrInvalidPriceRange
	}

	styleIDs, err := normalizeStyleIDs(in.StyleIDs)
	if err != nil {
		return nil, err
	}

	err = u.tx.Transaction(ctx, func(ctx context.Context) error {
		count, err := u.repo.CountStylesByIDs(ctx, styleIDs)
		if err != nil {
			return err
		}
		if count != len(styleIDs) {
			return ErrStyleNotFound
		}

		if err := u.repo.UpdateByUserID(ctx, userID, ProfileUpdate{
			Description:    description,
			MinPriceSatang: in.MinPriceSatang,
			MaxPriceSatang: in.MaxPriceSatang,
			UpdatedAt:      time.Now(),
		}); err != nil {
			return err
		}
		return u.repo.ReplaceStyles(ctx, userID, styleIDs)
	})
	if err != nil {
		return nil, err
	}

	return u.repo.GetByUserID(ctx, userID)
}

func normalizeStyleIDs(styleIDs []uuid.UUID) ([]uuid.UUID, error) {
	seen := make(map[uuid.UUID]struct{}, len(styleIDs))
	normalized := make([]uuid.UUID, 0, len(styleIDs))
	for _, styleID := range styleIDs {
		if styleID == uuid.Nil {
			return nil, apperror.InvalidInput("style id must not be empty", nil)
		}
		if _, exists := seen[styleID]; exists {
			continue
		}
		seen[styleID] = struct{}{}
		normalized = append(normalized, styleID)
	}
	slices.SortFunc(normalized, func(a, b uuid.UUID) int {
		return strings.Compare(a.String(), b.String())
	})
	return normalized, nil
}
