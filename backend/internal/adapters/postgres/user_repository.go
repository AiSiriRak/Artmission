// Package postgres implements every module's repository port against
// Postgres via bun. Each file owns a bun-tagged row model plus the
// conversions to/from its module's plain domain type — the domain layer
// never imports bun.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"
)

type userModel struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           uuid.UUID  `bun:"id,pk"`
	Username     string     `bun:"username"`
	Email        string     `bun:"email"`
	PasswordHash string     `bun:"password_hash"`
	Role         string     `bun:"role"`
	CreatedAt    time.Time  `bun:"created_at,nullzero"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero"`
	DeletedAt    *time.Time `bun:"deleted_at,soft_delete,nullzero"`
}

func newUserModel(u *user.User) *userModel {
	return &userModel{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         string(u.Role),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func (m *userModel) toDomain() *user.User {
	return &user.User{
		ID:           m.ID,
		Username:     m.Username,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Role:         user.Role(m.Role),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

type userRepository struct {
	base baserepo.BaseRepo[userModel]
	exec baserepo.Executor
}

var _ user.UserRepository = (*userRepository)(nil)

func NewUserRepository(db *bun.DB) user.UserRepository {
	return &userRepository{
		base: baserepo.NewBaseRepo[userModel](db, "user"),
		exec: baserepo.NewExecutor(db),
	}
}

func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	err := r.base.Create(ctx, newUserModel(u))
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		switch pgErr.ConstraintName {
		case "users_username_key":
			return user.ErrUsernameTaken
		case "users_email_key":
			return user.ErrEmailTaken
		}
	}
	return apperror.Internal("failed to create user", err)
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	model := new(userModel)
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().Model(model).Where("username = ?", username).Scan(ctx)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, apperror.Internal("failed to look up user by username", err)
	}
	return model.toDomain(), nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	model, err := r.base.FindByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, user.ErrUserNotFound
		}
		return nil, apperror.Internal("failed to look up user by id", err)
	}
	return model.toDomain(), nil
}
