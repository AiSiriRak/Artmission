// Package postgres implements every module's repository port against
// Postgres via bun. Each file owns the conversions to/from its module's
// plain domain type — the domain layer never imports bun. Row shapes
// themselves live in the model subpackage, shared with the seed package.
package postgres

import (
	"context"
	"database/sql"
	"errors"

	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"
)

func newUserModel(u *user.User) *pgmodel.User {
	return &pgmodel.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         string(u.Role),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func userModelToDomain(m *pgmodel.User) *user.User {
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
	base baserepo.BaseRepo[pgmodel.User]
	exec baserepo.Executor
}

var _ user.UserRepository = (*userRepository)(nil)

func NewUserRepository(db *bun.DB) user.UserRepository {
	return &userRepository{
		base: baserepo.NewBaseRepo[pgmodel.User](db, "user"),
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
		case "users_email_key":
			return user.ErrEmailTaken
		}
	}
	return apperror.Internal("failed to create user", err)
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	model := new(pgmodel.User)
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		return idb.NewSelect().Model(model).Where("email = ?", email).Scan(ctx)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, apperror.Internal("failed to look up user by email", err)
	}
	return userModelToDomain(model), nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	model, err := r.base.FindByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, user.ErrUserNotFound
		}
		return nil, apperror.Internal("failed to look up user by id", err)
	}
	return userModelToDomain(model), nil
}

func (r *userRepository) UpdateAccountByID(ctx context.Context, id uuid.UUID, in user.AccountUpdate) (*user.User, error) {
	model := &pgmodel.User{
		ID:        id,
		Username:  in.Username,
		UpdatedAt: in.UpdatedAt,
	}
	err := r.exec.Run(ctx, func(idb bun.IDB) error {
		query := idb.NewUpdate().
			Model(model).
			Column("username", "updated_at").
			WherePK().
			Returning("*")
		if in.PasswordHash != nil {
			model.PasswordHash = *in.PasswordHash
			query = query.Column("password_hash")
		}
		return query.Scan(ctx)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, apperror.Internal("failed to update account", err)
	}
	return userModelToDomain(model), nil
}
