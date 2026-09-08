package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artist"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type artistProfileModel struct {
	bun.BaseModel `bun:"table:artist_profiles,alias:ap"`

	UserID         uuid.UUID `bun:"user_id,pk"`
	Description	   string    `bun:"description,nullzero"`
	ArtistName     string    `bun:"artist_name,scanonly"`
	MinPriceSatang *int64    `bun:"min_price_satang"`
	MaxPriceSatang *int64    `bun:"max_price_satang"`
	ReviewScore    *float64  `bun:"review_score"`
	CreatedAt      time.Time `bun:"created_at,nullzero"`
	UpdatedAt      time.Time `bun:"updated_at,nullzero"`
}

func newArtistProfileModel(p *artist.Profile) *artistProfileModel {
	return &artistProfileModel{
		UserID:         p.UserID,
		Description:    p.Description,
		MinPriceSatang: p.MinPriceSatang,
		MaxPriceSatang: p.MaxPriceSatang,
		ReviewScore:    p.ReviewScore,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func (m *artistProfileModel) toDomain() *artist.Profile {
	return &artist.Profile{
		UserID:         m.UserID,
		ArtistName:     m.ArtistName,
		Description:    m.Description,
		MinPriceSatang: m.MinPriceSatang,
		MaxPriceSatang: m.MaxPriceSatang,
		ReviewScore:    m.ReviewScore,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

type artistStyleModel struct {
	bun.BaseModel `bun:"table:artist_styles,alias:ars"`

	ArtistID uuid.UUID `bun:"artist_id,pk"`
	StyleID  uuid.UUID `bun:"style_id,pk"`
}

type referenceModel struct {
	ID    uuid.UUID `bun:"id"`
	Label string    `bun:"label"`
}

type artistRepository struct {
	exec baserepo.Executor
}

var _ artist.ProfileRepository = (*artistRepository)(nil)

func NewArtistRepository(db *bun.DB) artist.ProfileRepository {
	return &artistRepository{exec: baserepo.NewExecutor(db)}
}

func (r *artistRepository) Create(ctx context.Context, p *artist.Profile) error {
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		_, err := idb.NewInsert().Model(newArtistProfileModel(p)).Exec(ctx)
		return err
	})
	if err != nil {
		return apperror.Internal("failed to create artist profile", err)
	}
	return nil
}

func (r *artistRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*artist.Profile, error) {
	model := new(artistProfileModel)
	var categories []referenceModel
	var styles []referenceModel

	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		if err := idb.NewSelect().
			Model(model).
			ColumnExpr("ap.*").
			ColumnExpr("u.username AS artist_name").
			Join("JOIN users AS u ON u.id = ap.user_id AND u.deleted_at IS NULL").
			Where("ap.user_id = ?", userID).
			Scan(ctx); err != nil {
			return err
		}

		if err := idb.NewSelect().
			TableExpr("categories AS c").
			ColumnExpr("DISTINCT c.id").
			ColumnExpr("c.label").
			Join("JOIN artworks AS a ON a.category_id = c.id").
			Where("a.artist_id = ?", userID).
			OrderExpr("c.label ASC, c.id ASC").
			Scan(ctx, &categories); err != nil {
			return err
		}

		return idb.NewSelect().
			TableExpr("styles AS s").
			ColumnExpr("s.id").
			ColumnExpr("s.label").
			Join("JOIN artist_styles AS ars ON ars.style_id = s.id").
			Where("ars.artist_id = ?", userID).
			OrderExpr("s.label ASC, s.id ASC").
			Scan(ctx, &styles)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, artist.ErrProfileNotFound
		}
		return nil, apperror.Internal("failed to get artist profile", err)
	}

	profile := model.toDomain()
	profile.Categories = make([]artist.Category, len(categories))
	for i, category := range categories {
		profile.Categories[i] = artist.Category{ID: category.ID, Label: category.Label}
	}
	profile.Styles = make([]artist.Style, len(styles))
	for i, style := range styles {
		profile.Styles[i] = artist.Style{ID: style.ID, Label: style.Label}
	}
	return profile, nil
}

func (r *artistRepository) UpdateByUserID(ctx context.Context, userID uuid.UUID, in artist.ProfileUpdate) error {
	model := &artistProfileModel{
		UserID:         userID,
		Description:    in.Description,
		MinPriceSatang: &in.MinPriceSatang,
		MaxPriceSatang: &in.MaxPriceSatang,
		UpdatedAt:      in.UpdatedAt,
	}
	var result sql.Result
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		var err error
		result, err = idb.NewUpdate().
			Model(model).
			Column("description", "min_price_satang", "max_price_satang", "updated_at").
			WherePK().
			Exec(ctx)
		return err
	})
	if err != nil {
		return apperror.Internal("failed to update artist profile", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return apperror.Internal("failed to inspect artist profile update", err)
	}
	if rows == 0 {
		return artist.ErrProfileNotFound
	}
	return nil
}

func (r *artistRepository) CountStylesByIDs(ctx context.Context, styleIDs []uuid.UUID) (int, error) {
	if len(styleIDs) == 0 {
		return 0, nil
	}

	var count int
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		var err error
		count, err = idb.NewSelect().Table("styles").Where("id IN (?)", bun.In(styleIDs)).Count(ctx)
		return err
	})
	if err != nil {
		return 0, apperror.Internal("failed to validate artist styles", err)
	}
	return count, nil
}

func (r *artistRepository) ReplaceStyles(ctx context.Context, userID uuid.UUID, styleIDs []uuid.UUID) error {
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		if _, err := idb.NewDelete().Model(new(artistStyleModel)).Where("artist_id = ?", userID).Exec(ctx); err != nil {
			return err
		}
		if len(styleIDs) == 0 {
			return nil
		}

		models := make([]artistStyleModel, len(styleIDs))
		for i, styleID := range styleIDs {
			models[i] = artistStyleModel{ArtistID: userID, StyleID: styleID}
		}
		_, err := idb.NewInsert().Model(&models).Exec(ctx)
		return err
	})
	if err != nil {
		return apperror.Internal("failed to replace artist styles", err)
	}
	return nil
}
