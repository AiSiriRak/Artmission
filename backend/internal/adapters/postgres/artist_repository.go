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

	UserID          uuid.UUID `bun:"user_id,pk"`
	Description     *string   `bun:"description"`
	ProfileImageKey *string   `bun:"profile_image_key"`
	ArtistName      string    `bun:"artist_name,scanonly"`
	MinPriceSatang  *int64    `bun:"min_price_satang,scanonly"`
	MaxPriceSatang  *int64    `bun:"max_price_satang,scanonly"`
	ReviewScore     *float64  `bun:"review_score,scanonly"`
	CreatedAt       time.Time `bun:"created_at,nullzero"`
	UpdatedAt       time.Time `bun:"updated_at,nullzero"`
}

func newArtistProfileModel(profile *artist.Profile) *artistProfileModel {
	return &artistProfileModel{
		UserID:          profile.UserID,
		Description:     profile.Description,
		ProfileImageKey: profile.ProfileImageKey,
		CreatedAt:       profile.CreatedAt,
		UpdatedAt:       profile.UpdatedAt,
	}
}

func (model *artistProfileModel) toDomain() *artist.Profile {
	return &artist.Profile{
		UserID:          model.UserID,
		ArtistName:      model.ArtistName,
		ProfileImageKey: model.ProfileImageKey,
		Description:     model.Description,
		MinPriceSatang:  model.MinPriceSatang,
		MaxPriceSatang:  model.MaxPriceSatang,
		ReviewScore:     model.ReviewScore,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}

type referenceModel struct {
	ID    uuid.UUID `bun:"id"`
	Label string    `bun:"label"`
}

type artistReviewModel struct {
	bun.BaseModel `bun:"table:reviews,alias:r"`

	Username string `bun:"username,scanonly"`
	Order    string `bun:"order_name,scanonly"`
	Rating   int    `bun:"rating"`
}

type artistRepository struct {
	exec baserepo.Executor
}

var _ artist.ProfileRepository = (*artistRepository)(nil)

func NewArtistRepository(db *bun.DB) artist.ProfileRepository {
	return &artistRepository{exec: baserepo.NewExecutor(db)}
}

func (r *artistRepository) Create(ctx context.Context, profile *artist.Profile) error {
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		_, err := idb.NewInsert().Model(newArtistProfileModel(profile)).Exec(ctx)
		return err
	})
	if err != nil {
		return apperror.Internal("failed to create artist profile", err)
	}
	return nil
}

func (r *artistRepository) GetByUserID(ctx context.Context, userID uuid.UUID, query artist.ProfileQuery) (*artist.Profile, error) {
	model := new(artistProfileModel)
	categories := make([]referenceModel, 0)
	styles := make([]referenceModel, 0)
	reviews := make([]artistReviewModel, 0)
	total := 0

	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		if err := idb.NewSelect().
			Model(model).
			ColumnExpr("ap.user_id, ap.description, ap.profile_image_key, ap.created_at, ap.updated_at").
			ColumnExpr("u.username AS artist_name").
			ColumnExpr("(SELECT MIN(a.price_satang) FROM artworks AS a WHERE a.artist_id = ap.user_id) AS min_price_satang").
			ColumnExpr("(SELECT MAX(a.price_satang) FROM artworks AS a WHERE a.artist_id = ap.user_id) AS max_price_satang").
			ColumnExpr("(SELECT ROUND(AVG(r.rating)::numeric, 1)::double precision FROM reviews AS r WHERE r.artist_id = ap.user_id) AS review_score").
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

		if err := idb.NewSelect().
			TableExpr("styles AS s").
			ColumnExpr("DISTINCT s.id").
			ColumnExpr("s.label").
			Join("JOIN artwork_styles AS aws ON aws.style_id = s.id").
			Join("JOIN artworks AS a ON a.id = aws.artwork_id").
			Where("a.artist_id = ?", userID).
			OrderExpr("s.label ASC, s.id ASC").
			Scan(ctx, &styles); err != nil {
			return err
		}

		reviewQuery := idb.NewSelect().
			Model(&reviews).
			ColumnExpr("reviewer.username AS username").
			ColumnExpr("o.artwork_name_snapshot AS order_name").
			ColumnExpr("r.rating").
			Join("JOIN users AS reviewer ON reviewer.id = r.customer_id").
			Join("JOIN orders AS o ON o.id = r.order_id").
			Where("r.artist_id = ?", userID).
			OrderExpr("r.created_at DESC, r.id DESC").
			Limit(query.Limit).
			Offset(query.Offset)
		var err error
		total, err = reviewQuery.ScanAndCount(ctx)
		return err
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
	profile.Reviews = make([]artist.Review, len(reviews))
	for i, review := range reviews {
		profile.Reviews[i] = artist.Review{Username: review.Username, Order: review.Order, Rating: review.Rating}
	}
	profile.Total = total
	return profile, nil
}

func (r *artistRepository) UpdateByUserID(ctx context.Context, userID uuid.UUID, update artist.ProfileUpdate) error {
	model := &artistProfileModel{
		UserID:          userID,
		Description:     update.Description,
		ProfileImageKey: update.ProfileImageKey,
		UpdatedAt:       update.UpdatedAt,
	}
	columns := []string{"updated_at"}
	if update.DescriptionSet {
		columns = append(columns, "description")
	}
	if update.ProfileImageKeySet {
		columns = append(columns, "profile_image_key")
	}

	var result sql.Result
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		var err error
		result, err = idb.NewUpdate().Model(model).Column(columns...).WherePK().Exec(ctx)
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
