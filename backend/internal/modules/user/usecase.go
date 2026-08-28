package user

import (
	"context"
	"time"

	"github.com/DeepAung/artmission/backend/internal/pkg/apperror"
	"github.com/DeepAung/artmission/backend/internal/pkg/security"
	"github.com/google/uuid"
)

type RegisterInput struct {
	Username string
	Email    string
	Phone    string
	Password string
	Role     Role
}

// UserUsecase is the driving port other layers (HTTP handlers, other
// modules) call into.
type UserUsecase interface {
	// Register creates a new account. Role must be RoleCustomer or
	// RoleArtist; RoleAdmin is never accepted here (admins are
	// seeded/ops-managed, not self-registered).
	Register(ctx context.Context, in RegisterInput) (*User, error)

	// Authenticate verifies credentials and returns the matching user.
	// Returns ErrInvalidCredential for both "no such user" and "wrong
	// password" so a caller cannot distinguish account existence from a
	// timing/response difference.
	Authenticate(ctx context.Context, username, password string) (*User, error)

	// GetByID looks up a user by id, used by modules/auth to rehydrate the
	// user for a valid session (e.g. on refresh).
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type userUsecase struct {
	repo UserRepository
}

func NewUserUsecase(repo UserRepository) UserUsecase {
	return &userUsecase{repo: repo}
}

func (u *userUsecase) Register(ctx context.Context, in RegisterInput) (*User, error) {
	if in.Role != RoleCustomer && in.Role != RoleArtist {
		return nil, ErrInvalidRole
	}

	hash, err := security.HashPassword(in.Password)
	if err != nil {
		return nil, apperror.Internal("failed to hash password", err)
	}

	now := time.Now()
	newUser := &User{
		ID:           uuid.New(),
		Username:     in.Username,
		Email:        in.Email,
		Phone:        in.Phone,
		PasswordHash: hash,
		Role:         in.Role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := u.repo.Create(ctx, newUser); err != nil {
		return nil, err
	}
	return newUser, nil
}

func (u *userUsecase) Authenticate(ctx context.Context, username, password string) (*User, error) {
	found, err := u.repo.GetByUsername(ctx, username)
	if err != nil {
		if err == ErrUserNotFound {
			return nil, ErrInvalidCredential
		}
		return nil, err
	}

	if !security.VerifyPassword(found.PasswordHash, password) {
		return nil, ErrInvalidCredential
	}
	return found, nil
}

func (u *userUsecase) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return u.repo.GetByID(ctx, id)
}
