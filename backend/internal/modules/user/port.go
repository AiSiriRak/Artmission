package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type BankAccountRepository interface {
	Create(ctx context.Context, ba *BankAccount) error
}

// ArtistRegistrar is implemented by the artist module and injected at wiring
// time so this package never imports artist (artist will later import user).
type ArtistRegistrar interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, description string) error
}

type Transactioner interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}
