package postgres

import (
	"context"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artist"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type artistProfileModel struct {
	bun.BaseModel `bun:"table:artist_profiles,alias:ap"`

	UserID      uuid.UUID `bun:"user_id,pk"`
	Description string    `bun:"description"`
	ReviewScore *float64  `bun:"review_score"`
	CreatedAt   time.Time `bun:"created_at,nullzero"`
	UpdatedAt   time.Time `bun:"updated_at,nullzero"`
}

func newArtistProfileModel(p *artist.Profile) *artistProfileModel {
	return &artistProfileModel{
		UserID:      p.UserID,
		Description: p.Description,
		ReviewScore: p.ReviewScore,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
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
