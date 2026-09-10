package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/artwork"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func newArtworkModel(item *artwork.Artwork, categoryID uuid.UUID) *pgmodel.Artwork {
	return &pgmodel.Artwork{
		ID:                  item.ID,
		ArtistID:            item.ArtistID,
		CategoryID:          categoryID,
		Name:                item.Name,
		Description:         item.Description,
		PriceSatang:         item.PriceSatang,
		MinimumDeadlineDays: item.MinimumDeadlineDays,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
}

func artworkModelToDomain(model *pgmodel.Artwork) artwork.Artwork {
	return artwork.Artwork{
		ID:                  model.ID,
		ArtistID:            model.ArtistID,
		Name:                model.Name,
		Category:            model.Category,
		Styles:              make([]string, 0),
		Description:         model.Description,
		Samples:             make([]artwork.Sample, 0),
		MinimumDeadlineDays: model.MinimumDeadlineDays,
		PriceSatang:         model.PriceSatang,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}

type artworkStyleLabel struct {
	ArtworkID uuid.UUID `bun:"artwork_id"`
	Label     string    `bun:"label"`
}

type artworkRepository struct {
	exec baserepo.Executor
}

var _ artwork.Repository = (*artworkRepository)(nil)

func NewArtworkRepository(db *bun.DB) artwork.Repository {
	return &artworkRepository{exec: baserepo.NewExecutor(db)}
}

func (repo *artworkRepository) FindOrCreateCategory(ctx context.Context, label string) (uuid.UUID, error) {
	model := &pgmodel.Category{ID: uuid.New(), Label: label}
	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		// A no-op update lets RETURNING return the existing category ID.
		_, err := idb.NewInsert().
			Model(model).
			On("CONFLICT (label) DO UPDATE").
			Set("label = EXCLUDED.label").
			Returning("id, label").
			Exec(ctx)
		return err
	})
	if err != nil {
		return uuid.Nil, apperror.Internal("failed to create artwork category", err)
	}
	return model.ID, nil
}

func (repo *artworkRepository) FindOrCreateStyles(ctx context.Context, labels []string) ([]uuid.UUID, error) {
	styleIDs := make([]uuid.UUID, len(labels))
	for index, label := range labels {
		model := &pgmodel.Style{ID: uuid.New(), Label: label}
		err := repo.exec.Run(ctx, func(idb bun.IDB) error {
			// A no-op update lets RETURNING return the existing style ID.
			_, err := idb.NewInsert().
				Model(model).
				On("CONFLICT (label) DO UPDATE").
				Set("label = EXCLUDED.label").
				Returning("id, label").
				Exec(ctx)
			return err
		})
		if err != nil {
			return nil, apperror.Internal("failed to create artwork style", err)
		}
		styleIDs[index] = model.ID
	}
	return styleIDs, nil
}

func (repo *artworkRepository) Create(ctx context.Context, item *artwork.Artwork, categoryID uuid.UUID, styleIDs []uuid.UUID) error {
	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		if _, err := idb.NewInsert().Model(newArtworkModel(item, categoryID)).Exec(ctx); err != nil {
			return err
		}

		if len(styleIDs) > 0 {
			styles := make([]pgmodel.ArtworkStyle, len(styleIDs))
			for index, styleID := range styleIDs {
				styles[index] = pgmodel.ArtworkStyle{ArtworkID: item.ID, StyleID: styleID}
			}
			if _, err := idb.NewInsert().Model(&styles).Exec(ctx); err != nil {
				return err
			}
		}

		if len(item.Samples) > 0 {
			samples := make([]pgmodel.ArtworkImage, len(item.Samples))
			for index, sample := range item.Samples {
				samples[index] = pgmodel.ArtworkImage{
					ID:        uuid.New(),
					ArtworkID: item.ID,
					ImageURL:  sample.ImageURL,
					SortOrder: sample.SortOrder,
				}
			}
			if _, err := idb.NewInsert().Model(&samples).Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return apperror.Internal("failed to create artwork", err)
	}
	return nil
}

func (repo *artworkRepository) UpdateOwnedBy(ctx context.Context, item *artwork.Artwork, categoryID uuid.UUID, styleIDs []uuid.UUID, deletedSampleURLs []string) ([]string, error) {
	deletedURLs := make([]string, 0, len(deletedSampleURLs))
	timestamps := struct {
		CreatedAt time.Time `bun:"created_at"`
		UpdatedAt time.Time `bun:"updated_at"`
	}{}
	model := newArtworkModel(item, categoryID)
	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		err := idb.NewUpdate().
			Model(model).
			Column("category_id", "name", "description", "price_satang", "minimum_deadline_days", "updated_at").
			Where("id = ? AND artist_id = ?", item.ID, item.ArtistID).
			Returning("created_at, updated_at").
			Scan(ctx, &timestamps)
		if errors.Is(err, sql.ErrNoRows) {
			return artwork.ErrArtworkNotFound
		}
		if err != nil {
			return err
		}
		existingImages := make([]pgmodel.ArtworkImage, 0)
		if err := idb.NewSelect().
			Model(&existingImages).
			Column("id", "artwork_id", "image_url", "sort_order").
			Where("artwork_id = ?", item.ID).
			OrderExpr("sort_order ASC, id ASC").
			Scan(ctx); err != nil {
			return err
		}

		deleteSet := make(map[string]struct{}, len(deletedSampleURLs))
		for _, imageURL := range deletedSampleURLs {
			deleteSet[imageURL] = struct{}{}
		}
		retainedSamples := make([]artwork.Sample, 0, len(existingImages))
		for _, image := range existingImages {
			if _, deleted := deleteSet[image.ImageURL]; deleted {
				deletedURLs = append(deletedURLs, image.ImageURL)
				delete(deleteSet, image.ImageURL)
				continue
			}
			retainedSamples = append(retainedSamples, artwork.Sample{ImageURL: image.ImageURL})
		}
		if len(deleteSet) > 0 {
			return artwork.ErrSampleNotOwned
		}

		uploadedSamples := append([]artwork.Sample(nil), item.Samples...)
		item.Samples = append(retainedSamples, uploadedSamples...)
		for index := range item.Samples {
			item.Samples[index].SortOrder = index
		}

		if _, err := idb.NewDelete().Model(new(pgmodel.ArtworkStyle)).Where("artwork_id = ?", item.ID).Exec(ctx); err != nil {
			return err
		}
		if _, err := idb.NewDelete().Model(new(pgmodel.ArtworkImage)).Where("artwork_id = ?", item.ID).Exec(ctx); err != nil {
			return err
		}

		if len(styleIDs) > 0 {
			styles := make([]pgmodel.ArtworkStyle, len(styleIDs))
			for index, styleID := range styleIDs {
				styles[index] = pgmodel.ArtworkStyle{ArtworkID: item.ID, StyleID: styleID}
			}
			if _, err := idb.NewInsert().Model(&styles).Exec(ctx); err != nil {
				return err
			}
		}
		if len(item.Samples) > 0 {
			samples := make([]pgmodel.ArtworkImage, len(item.Samples))
			for index, sample := range item.Samples {
				samples[index] = pgmodel.ArtworkImage{
					ID:        uuid.New(),
					ArtworkID: item.ID,
					ImageURL:  sample.ImageURL,
					SortOrder: sample.SortOrder,
				}
			}
			if _, err := idb.NewInsert().Model(&samples).Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, artwork.ErrArtworkNotFound) {
			return nil, artwork.ErrArtworkNotFound
		}
		if errors.Is(err, artwork.ErrSampleNotOwned) {
			return nil, artwork.ErrSampleNotOwned
		}
		return nil, apperror.Internal("failed to update artwork", err)
	}
	item.CreatedAt = timestamps.CreatedAt
	item.UpdatedAt = timestamps.UpdatedAt
	return deletedURLs, nil
}

func (repo *artworkRepository) DeleteOwnedBy(ctx context.Context, artworkID, artistID uuid.UUID) ([]string, error) {
	deletedURLs := make([]string, 0)
	var result sql.Result
	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		if err := idb.NewSelect().
			Model(new(pgmodel.ArtworkImage)).
			Column("ai.image_url").
			Join("JOIN artworks AS a ON a.id = ai.artwork_id").
			Where("ai.artwork_id = ? AND a.artist_id = ?", artworkID, artistID).
			Scan(ctx, &deletedURLs); err != nil {
			return err
		}
		var err error
		result, err = idb.NewDelete().
			Model(new(pgmodel.Artwork)).
			Where("id = ? AND artist_id = ?", artworkID, artistID).
			Exec(ctx)
		return err
	})
	if err != nil {
		return nil, apperror.Internal("failed to delete artwork", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, apperror.Internal("failed to inspect artwork deletion", err)
	}
	if rows == 0 {
		return nil, artwork.ErrArtworkNotFound
	}
	return deletedURLs, nil
}

func (repo *artworkRepository) ListByArtistID(ctx context.Context, artistID uuid.UUID) ([]artwork.Artwork, error) {
	models := make([]pgmodel.Artwork, 0)
	styleLabels := make([]artworkStyleLabel, 0)
	imageModels := make([]pgmodel.ArtworkImage, 0)

	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		exists, err := idb.NewSelect().
			TableExpr("artist_profiles AS ap").
			Join("JOIN users AS u ON u.id = ap.user_id AND u.deleted_at IS NULL").
			Where("ap.user_id = ?", artistID).
			Exists(ctx)
		if err != nil {
			return err
		}
		if !exists {
			return artwork.ErrArtistNotFound
		}

		if err := idb.NewSelect().
			Model(&models).
			ColumnExpr("art.id, art.artist_id, art.name, art.description, art.price_satang, art.minimum_deadline_days, art.created_at, art.updated_at").
			ColumnExpr("c.label AS category").
			Join("JOIN categories AS c ON c.id = art.category_id").
			Where("art.artist_id = ?", artistID).
			OrderExpr("art.created_at DESC, art.id DESC").
			Scan(ctx); err != nil {
			return err
		}
		if len(models) == 0 {
			return nil
		}

		artworkIDs := make([]uuid.UUID, len(models))
		for index := range models {
			artworkIDs[index] = models[index].ID
		}

		if err := idb.NewSelect().
			TableExpr("artwork_styles AS aws").
			ColumnExpr("aws.artwork_id").
			ColumnExpr("s.label").
			Join("JOIN styles AS s ON s.id = aws.style_id").
			Where("aws.artwork_id IN (?)", bun.List(artworkIDs)).
			OrderExpr("aws.artwork_id ASC, s.label ASC, s.id ASC").
			Scan(ctx, &styleLabels); err != nil {
			return err
		}

		return idb.NewSelect().
			TableExpr("artwork_images AS ai").
			ColumnExpr("ai.artwork_id, ai.image_url").
			Where("ai.artwork_id IN (?)", bun.List(artworkIDs)).
			OrderExpr("ai.artwork_id ASC, ai.sort_order ASC, ai.id ASC").
			Scan(ctx, &imageModels)
	})
	if err != nil {
		if errors.Is(err, artwork.ErrArtistNotFound) {
			return nil, artwork.ErrArtistNotFound
		}
		return nil, apperror.Internal("failed to list artist artworks", err)
	}

	artworks := make([]artwork.Artwork, len(models))
	indexByID := make(map[uuid.UUID]int, len(models))
	for index := range models {
		artworks[index] = artworkModelToDomain(&models[index])
		indexByID[models[index].ID] = index
	}
	for _, style := range styleLabels {
		index := indexByID[style.ArtworkID]
		artworks[index].Styles = append(artworks[index].Styles, style.Label)
	}
	for _, image := range imageModels {
		index := indexByID[image.ArtworkID]
		artworks[index].Samples = append(artworks[index].Samples, artwork.Sample{ImageURL: image.ImageURL})
	}
	return artworks, nil
}
