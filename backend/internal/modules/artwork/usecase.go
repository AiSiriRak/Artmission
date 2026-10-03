package artwork

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/gabriel-vasile/mimetype"
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

func (u *usecase) Search(ctx context.Context, query SearchQuery) (SearchPage, error) {
	normalized, err := normalizeSearchQuery(query)
	if err != nil {
		return SearchPage{}, err
	}

	page, err := u.repo.Search(ctx, normalized)
	if err != nil {
		return SearchPage{}, err
	}
	if page.Items == nil {
		page.Items = make([]SearchItem, 0)
	}
	for i := range page.Items {
		if page.Items[i].Artwork.Styles == nil {
			page.Items[i].Artwork.Styles = make([]string, 0)
		}
		if page.Items[i].Artwork.Samples == nil {
			page.Items[i].Artwork.Samples = make([]Sample, 0)
		}
		u.attachSearchArtistURL(&page.Items[i].Artist)
	}
	page.Page = normalized.Page
	return page, nil
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
	deletedSampleURLs, err := normalizeDeletedSampleURLs(input.DeletedSampleURLs)
	if err != nil {
		return nil, err
	}
	uploadedURLs, err := u.uploadSamples(ctx, updated, input.SampleFiles)
	if err != nil {
		return nil, err
	}
	var deletedURLs []string
	err = u.tx.Transaction(ctx, func(ctx context.Context) error {
		categoryID, err := u.repo.FindOrCreateCategory(ctx, updated.Category)
		if err != nil {
			return err
		}
		styleIDs, err := u.repo.FindOrCreateStyles(ctx, updated.Styles)
		if err != nil {
			return err
		}
		deletedURLs, err = u.repo.UpdateOwnedBy(ctx, updated, categoryID, styleIDs, deletedSampleURLs)
		return err
	})
	if err != nil {
		u.deleteSampleURLs(ctx, uploadedURLs)
		return nil, err
	}
	u.deleteSampleURLs(ctx, deletedURLs)
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

func normalizeDeletedSampleURLs(urls []string) ([]string, error) {
	seen := make(map[string]struct{}, len(urls))
	normalized := make([]string, 0, len(urls))
	for _, imageURL := range urls {
		value, err := requiredText("deleted sample URL", imageURL)
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
		key := fmt.Sprintf("artists/%s/artworks/%s/%s.%s", item.ArtistID, item.ID, uuid.New(), extension)
		if err := u.storage.Upload(ctx, key, bytes.NewReader(content), contentType); err != nil {
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
		key, ok := u.storage.KeyFromURL(imageURL)
		if !ok {
			continue
		}
		_ = u.storage.Delete(ctx, key)
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
	switch mimetype.Detect(content).String() {
	case "image/jpeg":
		return content, "image/jpeg", "jpg", nil
	case "image/png":
		return content, "image/png", "png", nil
	case "image/webp":
		return content, "image/webp", "webp", nil
	default:
		return nil, "", "", ErrInvalidSampleImage
	}
}

func (u *usecase) attachSearchArtistURL(artist *ArtistSummary) {
	if artist.ProfileImageKey == nil {
		artist.ProfileURL = nil
		return
	}
	url := u.storage.PublicURL(*artist.ProfileImageKey)
	artist.ProfileURL = &url
}

func normalizeSearchQuery(query SearchQuery) (SearchQuery, error) {
	query.ArtistName = strings.TrimSpace(query.ArtistName)

	if query.Category != "" {
		category, err := normalizeFilterLabel("category", query.Category)
		if err != nil {
			return SearchQuery{}, err
		}
		query.Category = category
	}

	styles, err := normalizeFilterLabels("style", query.Styles)
	if err != nil {
		return SearchQuery{}, err
	}
	query.Styles = styles

	if query.MinPriceSatang != nil && *query.MinPriceSatang < 0 {
		return SearchQuery{}, apperror.InvalidInput("min price satang must not be negative", nil)
	}
	if query.MaxPriceSatang != nil && *query.MaxPriceSatang < 0 {
		return SearchQuery{}, apperror.InvalidInput("max price satang must not be negative", nil)
	}
	if query.MinPriceSatang != nil && query.MaxPriceSatang != nil && *query.MinPriceSatang > *query.MaxPriceSatang {
		return SearchQuery{}, apperror.InvalidInput("min price satang must not be greater than max price satang", nil)
	}

	if query.MinReviewScore != nil {
		if *query.MinReviewScore < 0 || *query.MinReviewScore > 5 {
			return SearchQuery{}, apperror.InvalidInput("min review score must be between 0 and 5", nil)
		}
	}

	if query.Sort == "" {
		query.Sort = DefaultSearchSort
	}
	if !query.Sort.IsValid() {
		return SearchQuery{}, apperror.InvalidInput(fmt.Sprintf("invalid sort %q", query.Sort), nil)
	}

	if query.Page == 0 {
		query.Page = 1
	}
	if query.Page < 1 {
		return SearchQuery{}, apperror.InvalidInput("page must be greater than zero", nil)
	}

	query.Limit = SearchPageSize
	query.Offset = (query.Page - 1) * SearchPageSize
	return query, nil
}

func normalizeFilterLabels(field string, labels []string) ([]string, error) {
	seen := make(map[string]struct{}, len(labels))
	normalized := make([]string, 0, len(labels))
	for _, label := range labels {
		value, err := requiredText(field, label)
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

func normalizeFilterLabel(field string, label string) (string, error) {
	value, err := requiredText(field, label)
	if err != nil {
		return "", err
	}
	return value, nil
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
