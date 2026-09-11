package artwork

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

type usecase struct {
	repo Repository
	tx   Transactioner
}

func NewUsecase(repo Repository, tx Transactioner) Usecase {
	return &usecase{repo: repo, tx: tx}
}

func (u *usecase) Create(ctx context.Context, input CreateInput) (*Artwork, error) {
	if input.ArtistID == uuid.Nil {
		return nil, apperror.InvalidInput("artist id must not be empty", nil)
	}

	name, err := requiredText("name", input.Name)
	if err != nil {
		return nil, err
	}
	category, err := requiredText("category", input.Category)
	if err != nil {
		return nil, err
	}
	description, err := requiredText("description", input.Description)
	if err != nil {
		return nil, err
	}
	styles, err := normalizeStyles(input.Styles)
	if err != nil {
		return nil, err
	}
	samples, err := normalizeSamples(input.Samples)
	if err != nil {
		return nil, err
	}
	if input.MinimumDeadlineDays <= 0 {
		return nil, apperror.InvalidInput("minimum deadline days must be greater than zero", nil)
	}
	if input.PriceSatang < 0 {
		return nil, apperror.InvalidInput("price satang must not be negative", nil)
	}

	now := time.Now()
	created := &Artwork{
		ID:                  uuid.New(),
		ArtistID:            input.ArtistID,
		Name:                name,
		Category:            category,
		Styles:              styles,
		Description:         description,
		Samples:             samples,
		MinimumDeadlineDays: input.MinimumDeadlineDays,
		PriceSatang:         input.PriceSatang,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	err = u.tx.Transaction(ctx, func(ctx context.Context) error {
		categoryID, err := u.repo.FindOrCreateCategory(ctx, category)
		if err != nil {
			return err
		}
		styleIDs, err := u.repo.FindOrCreateStyles(ctx, styles)
		if err != nil {
			return err
		}
		return u.repo.Create(ctx, created, categoryID, styleIDs)
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (u *usecase) Delete(ctx context.Context, artistID, artworkID uuid.UUID) error {
	if artistID == uuid.Nil || artworkID == uuid.Nil {
		return apperror.InvalidInput("artwork id and artist id must not be empty", nil)
	}
	return u.repo.DeleteOwnedBy(ctx, artworkID, artistID)
}

func requiredText(field, value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", apperror.InvalidInput(field+" must not be blank", nil)
	}
	return normalized, nil
}

func normalizeStyles(labels []string) ([]string, error) {
	seen := make(map[string]struct{}, len(labels))
	normalized := make([]string, 0, len(labels))
	for _, label := range labels {
		value, err := requiredText("style", label)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized, nil
}

func normalizeSamples(samples []Sample) ([]Sample, error) {
	normalized := make([]Sample, len(samples))
	for index, sample := range samples {
		imageURL, err := requiredText("image url", sample.ImageURL)
		if err != nil {
			return nil, err
		}
		parsed, err := url.Parse(imageURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return nil, apperror.InvalidInput(fmt.Sprintf("artwork sample %d must have an absolute HTTP or HTTPS image URL", index+1), err)
		}
		normalized[index] = Sample{ImageURL: imageURL, SortOrder: index}
	}
	return normalized, nil
}

func (u *usecase) ListByArtistID(ctx context.Context, artistID uuid.UUID) ([]Artwork, error) {
	artworks, err := u.repo.ListByArtistID(ctx, artistID)
	if err != nil {
		return nil, err
	}
	if artworks == nil {
		artworks = make([]Artwork, 0)
	}
	for artworkIndex := range artworks {
		if artworks[artworkIndex].Styles == nil {
			artworks[artworkIndex].Styles = make([]string, 0)
		}
		if artworks[artworkIndex].Samples == nil {
			artworks[artworkIndex].Samples = make([]Sample, 0)
		}
	}
	return artworks, nil
}
