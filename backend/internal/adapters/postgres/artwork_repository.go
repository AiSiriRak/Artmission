package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artwork"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type artworkModel struct {
	bun.BaseModel `bun:"table:artworks,alias:a"`

	ID                  uuid.UUID `bun:"id,pk"`
	ArtistID            uuid.UUID `bun:"artist_id"`
	CategoryID          uuid.UUID `bun:"category_id"`
	Name                string    `bun:"name"`
	Category            string    `bun:"category,scanonly"`
	Description         string    `bun:"description"`
	PriceSatang         int64     `bun:"price_satang"`
	MinimumDeadlineDays int       `bun:"minimum_deadline_days"`
	CreatedAt           time.Time `bun:"created_at"`
	UpdatedAt           time.Time `bun:"updated_at"`
}

type artworkStyleModel struct {
	bun.BaseModel `bun:"table:artwork_styles,alias:aws"`

	ArtworkID uuid.UUID `bun:"artwork_id,pk"`
	StyleID   uuid.UUID `bun:"style_id,pk"`
	Label     string    `bun:"label,scanonly"`
}

type artworkImageModel struct {
	bun.BaseModel `bun:"table:artwork_images,alias:ai"`

	ID        uuid.UUID `bun:"id,pk"`
	ArtworkID uuid.UUID `bun:"artwork_id"`
	ImageURL  string    `bun:"image_url"`
	SortOrder int       `bun:"sort_order"`
}

type artworkCategoryModel struct {
	bun.BaseModel `bun:"table:categories,alias:c"`

	ID    uuid.UUID `bun:"id,pk"`
	Label string    `bun:"label"`
}

type artworkReferenceStyleModel struct {
	bun.BaseModel `bun:"table:styles,alias:s"`

	ID    uuid.UUID `bun:"id,pk"`
	Label string    `bun:"label"`
}

type artworkRepository struct {
	exec baserepo.Executor
}

var _ artwork.Repository = (*artworkRepository)(nil)

func NewArtworkRepository(db *bun.DB) artwork.Repository {
	return &artworkRepository{exec: baserepo.NewExecutor(db)}
}

func (repo *artworkRepository) FindOrCreateCategory(ctx context.Context, label string) (uuid.UUID, error) {
	model := &artworkCategoryModel{ID: uuid.New(), Label: label}
	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
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
		model := &artworkReferenceStyleModel{ID: uuid.New(), Label: label}
		err := repo.exec.Run(ctx, func(idb bun.IDB) error {
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
		if _, err := idb.NewInsert().Model(&artworkModel{
			ID:                  item.ID,
			ArtistID:            item.ArtistID,
			CategoryID:          categoryID,
			Name:                item.Name,
			Description:         item.Description,
			PriceSatang:         item.PriceSatang,
			MinimumDeadlineDays: item.MinimumDeadlineDays,
		}).Exec(ctx); err != nil {
			return err
		}

		if len(styleIDs) > 0 {
			styles := make([]artworkStyleModel, len(styleIDs))
			for index, styleID := range styleIDs {
				styles[index] = artworkStyleModel{ArtworkID: item.ID, StyleID: styleID}
			}
			if _, err := idb.NewInsert().Model(&styles).Exec(ctx); err != nil {
				return err
			}
		}

		if len(item.Samples) > 0 {
			samples := make([]artworkImageModel, len(item.Samples))
			for index, sample := range item.Samples {
				samples[index] = artworkImageModel{
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

func (repo *artworkRepository) DeleteOwnedBy(ctx context.Context, artworkID, artistID uuid.UUID) error {
	var result sql.Result
	err := repo.exec.Run(ctx, func(idb bun.IDB) error {
		var err error
		result, err = idb.NewDelete().
			Model(new(artworkModel)).
			Where("id = ? AND artist_id = ?", artworkID, artistID).
			Exec(ctx)
		return err
	})
	if err != nil {
		return apperror.Internal("failed to delete artwork", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return apperror.Internal("failed to inspect artwork deletion", err)
	}
	if rows == 0 {
		return artwork.ErrArtworkNotFound
	}
	return nil
}

func (repo *artworkRepository) ListByArtistID(ctx context.Context, artistID uuid.UUID) ([]artwork.Artwork, error) {
	models := make([]artworkModel, 0)
	styleModels := make([]artworkStyleModel, 0)
	imageModels := make([]artworkImageModel, 0)

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
			ColumnExpr("a.id, a.artist_id, a.name, a.description, a.price_satang, a.minimum_deadline_days, a.created_at, a.updated_at").
			ColumnExpr("c.label AS category").
			Join("JOIN categories AS c ON c.id = a.category_id").
			Where("a.artist_id = ?", artistID).
			OrderExpr("a.created_at DESC, a.id DESC").
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
			Where("aws.artwork_id IN (?)", bun.In(artworkIDs)).
			OrderExpr("aws.artwork_id ASC, s.label ASC, s.id ASC").
			Scan(ctx, &styleModels); err != nil {
			return err
		}

		return idb.NewSelect().
			TableExpr("artwork_images AS ai").
			ColumnExpr("ai.artwork_id, ai.image_url").
			Where("ai.artwork_id IN (?)", bun.In(artworkIDs)).
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
	for index, model := range models {
		artworks[index] = artwork.Artwork{
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
		indexByID[model.ID] = index
	}
	for _, style := range styleModels {
		index := indexByID[style.ArtworkID]
		artworks[index].Styles = append(artworks[index].Styles, style.Label)
	}
	for _, image := range imageModels {
		index := indexByID[image.ArtworkID]
		artworks[index].Samples = append(artworks[index].Samples, artwork.Sample{ImageURL: image.ImageURL})
	}
	return artworks, nil
}
