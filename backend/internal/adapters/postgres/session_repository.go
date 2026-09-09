package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type sessionModel struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID               uuid.UUID `bun:"id,pk"`
	UserID           uuid.UUID `bun:"user_id"`
	RefreshTokenHash string    `bun:"refresh_token_hash"`
	ExpiresAt        time.Time `bun:"expires_at"`
	CreatedAt        time.Time `bun:"created_at,nullzero"`
}

func newSessionModel(s *auth.Session) *sessionModel {
	return &sessionModel{
		ID:               s.ID,
		UserID:           s.UserID,
		RefreshTokenHash: s.RefreshTokenHash,
		ExpiresAt:        s.ExpiresAt,
		CreatedAt:        s.CreatedAt,
	}
}

func (m *sessionModel) toDomain() *auth.Session {
	return &auth.Session{
		ID:               m.ID,
		UserID:           m.UserID,
		RefreshTokenHash: m.RefreshTokenHash,
		ExpiresAt:        m.ExpiresAt,
		CreatedAt:        m.CreatedAt,
	}
}

// sessionRepository creates sessions only while a shared lock confirms the
// account is live, so an account-deletion lock cannot race a new session in.
type sessionRepository struct {
	base baserepo.BaseRepo[sessionModel]
	exec baserepo.Executor
}

var _ auth.SessionRepository = (*sessionRepository)(nil)

func NewSessionRepository(db *bun.DB) auth.SessionRepository {
	return &sessionRepository{
		base: baserepo.NewBaseRepo[sessionModel](db, "session"),
		exec: baserepo.NewExecutor(db),
	}
}

func (r *sessionRepository) Create(ctx context.Context, s *auth.Session) error {
	var rowsAffected int64
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		result, err := idb.NewRaw(`
			INSERT INTO sessions (id, user_id, refresh_token_hash, expires_at, created_at)
			SELECT ?, ?, ?, ?, ?
			FROM users
			WHERE id = ? AND deleted_at IS NULL
			FOR KEY SHARE
		`, s.ID, s.UserID, s.RefreshTokenHash, s.ExpiresAt, s.CreatedAt, s.UserID).Exec(ctx)
		if err == nil {
			rowsAffected, err = result.RowsAffected()
		}
		return err
	})
	if err != nil {
		return apperror.Internal("failed to create session", err)
	}
	if rowsAffected == 0 {
		return auth.ErrSessionNotFound
	}
	return nil
}

func (r *sessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*auth.Session, error) {
	model, err := r.base.FindByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, auth.ErrSessionNotFound
		}
		return nil, apperror.Internal("failed to look up session", err)
	}
	return model.toDomain(), nil
}

func (r *sessionRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if err := r.base.DeleteByID(ctx, id); err != nil {
		if isNotFound(err) {
			return auth.ErrSessionNotFound
		}
		return apperror.Internal("failed to delete session", err)
	}
	return nil
}

func isNotFound(err error) bool {
	var appErr *apperror.Error
	return errors.As(err, &appErr) && appErr.Code == apperror.CodeNotFound
}
