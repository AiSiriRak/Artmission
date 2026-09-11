package postgres

import (
	"context"
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
	Name                string    `bun:"name"`
	Category            string    `bun:"category,scanonly"`
	Description         string    `bun:"description"`
	PriceSatang         int64     `bun:"price_satang"`
	MinimumDeadlineDays int       `bun:"minimum_deadline_days"`
	CreatedAt           time.Time `bun:"created_at"`
	UpdatedAt           time.Time `bun:"updated_at"`
}

type artworkStyleModel struct {
	ArtworkID uuid.UUID `bun:"artwork_id"`
	Label     string    `bun:"label"`
}

type artworkImageModel struct {
	ArtworkID uuid.UUID `bun:"artwork_id"`
	ImageURL  string    `bun:"image_url"`
}

type artworkRepository struct {
	exec baserepo.Executor
}

var _ artwork.Repository = (*artworkRepository)(nil)

func NewArtworkRepository(db *bun.DB) artwork.Repository {
	return &artworkRepository{exec: baserepo.NewExecutor(db)}
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
