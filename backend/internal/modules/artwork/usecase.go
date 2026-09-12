package artwork

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

type usecase struct {
	repo    Repository
	tx      Transactioner
	storage ObjectStorage
}

func NewUsecase(repo Repository, tx Transactioner, storage ObjectStorage) Usecase {
	return &usecase{repo: repo, tx: tx, storage: storage}
}

func (u *usecase) Create(ctx context.Context, input CreateInput) (*Artwork, error) {
	created, err := normalizeArtwork(uuid.New(), input)
	if err != nil {
		return nil, err
	}
	uploadedURLs, err := u.uploadSamples(ctx, created, input.SampleFiles)
	if err != nil {
		return nil, err
	}

	err = u.tx.Transaction(ctx, func(ctx context.Context) error {
		categoryID, err := u.repo.FindOrCreateCategory(ctx, created.Category)
		if err != nil {
			return err
		}
		styleIDs, err := u.repo.FindOrCreateStyles(ctx, created.Styles)
		if err != nil {
			return err
		}
		return u.repo.Create(ctx, created, categoryID, styleIDs)
	})
	if err != nil {
		u.deleteSampleURLs(ctx, uploadedURLs)
		return nil, err
	}
	return created, nil
}

func (u *usecase) Update(ctx context.Context, input UpdateInput) (*Artwork, error) {
	if input.ArtworkID == uuid.Nil {
		return nil, apperror.InvalidInput("artwork id must not be empty", nil)
	}

	updated, err := normalizeArtwork(input.ArtworkID, input.CreateInput)
	if err != nil {
		return nil, err
	}
	uploadedURLs, err := u.uploadSamples(ctx, updated, input.SampleFiles)
	if err != nil {
		return nil, err
	}
	var replacedURLs []string
	err = u.tx.Transaction(ctx, func(ctx context.Context) error {
		categoryID, err := u.repo.FindOrCreateCategory(ctx, updated.Category)
		if err != nil {
			return err
		}
		styleIDs, err := u.repo.FindOrCreateStyles(ctx, updated.Styles)
		if err != nil {
			return err
		}
		replacedURLs, err = u.repo.UpdateOwnedBy(ctx, updated, categoryID, styleIDs)
		return err
	})
	if err != nil {
		u.deleteSampleURLs(ctx, uploadedURLs)
		return nil, err
	}
	u.deleteSampleURLs(ctx, replacedURLs)
	return updated, nil
}

func (u *usecase) Delete(ctx context.Context, artistID, artworkID uuid.UUID) error {
	if artistID == uuid.Nil || artworkID == uuid.Nil {
		return apperror.InvalidInput("artwork id and artist id must not be empty", nil)
	}
	var deletedURLs []string
	err := u.tx.Transaction(ctx, func(ctx context.Context) error {
		var err error
		deletedURLs, err = u.repo.DeleteOwnedBy(ctx, artworkID, artistID)
		return err
	})
	if err != nil {
		return err
	}
	u.deleteSampleURLs(ctx, deletedURLs)
	return nil
}

func normalizeArtwork(id uuid.UUID, input CreateInput) (*Artwork, error) {
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
	if input.MinimumDeadlineDays <= 0 {
		return nil, apperror.InvalidInput("minimum deadline days must be greater than zero", nil)
	}
	if input.PriceSatang < 0 {
		return nil, apperror.InvalidInput("price satang must not be negative", nil)
	}

	now := time.Now()
	return &Artwork{
		ID:                  id,
		ArtistID:            input.ArtistID,
		Name:                name,
		Category:            category,
		Styles:              styles,
		Description:         description,
		Samples:             make([]Sample, 0, len(input.SampleFiles)),
		MinimumDeadlineDays: input.MinimumDeadlineDays,
		PriceSatang:         input.PriceSatang,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
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

func (u *usecase) uploadSamples(ctx context.Context, item *Artwork, files []io.Reader) ([]string, error) {
	uploadedURLs := make([]string, 0, len(files))
	for index, reader := range files {
		content, contentType, extension, err := readSampleImage(reader)
		if err != nil {
			u.deleteSampleURLs(ctx, uploadedURLs)
			return nil, err
		}
		key := fmt.Sprintf("artworks/%s/%s/%s.%s", item.ArtistID, item.ID, uuid.New(), extension)
		if err := u.storage.UploadPublic(ctx, key, bytes.NewReader(content), int64(len(content)), contentType); err != nil {
			u.deleteSampleURLs(ctx, uploadedURLs)
			return nil, apperror.Internal("failed to upload artwork sample", err)
		}
		imageURL := u.storage.PublicURL(key)
		item.Samples = append(item.Samples, Sample{ImageURL: imageURL, SortOrder: index})
		uploadedURLs = append(uploadedURLs, imageURL)
	}
	return uploadedURLs, nil
}

func (u *usecase) deleteSampleURLs(ctx context.Context, urls []string) {
	for _, imageURL := range urls {
		_ = u.storage.DeletePublicURL(ctx, imageURL)
	}
}

func readSampleImage(reader io.Reader) ([]byte, string, string, error) {
	content, err := io.ReadAll(io.LimitReader(reader, MaxSampleImageSize+1))
	if err != nil {
		return nil, "", "", apperror.InvalidInput("failed to read artwork sample image", err)
	}
	if len(content) > MaxSampleImageSize {
		return nil, "", "", ErrSampleImageTooLarge
	}
	switch {
	case len(content) >= 3 && bytes.Equal(content[:3], []byte{0xff, 0xd8, 0xff}):
		return content, "image/jpeg", "jpg", nil
	case len(content) >= 8 && bytes.Equal(content[:8], []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}):
		return content, "image/png", "png", nil
	case len(content) >= 12 && string(content[:4]) == "RIFF" && string(content[8:12]) == "WEBP":
		return content, "image/webp", "webp", nil
	default:
		return nil, "", "", ErrInvalidSampleImage
	}
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
