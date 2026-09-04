package user

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository is the driven port for user persistence.
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}
