package user

import (
	"context"

	"github.com/google/uuid"
)

type UserUsecase interface {
	// Register creates the user, bank account, and (when role is artist)
	// artist profile in one transaction. RoleAdmin is never accepted
	// (admins are seeded/ops-managed, not self-registered).
	Register(ctx context.Context, in RegisterInput) (*User, error)

	// Authenticate returns ErrInvalidCredential for both "no such user"
	// and "wrong password" so a caller cannot distinguish account existence.
	Authenticate(ctx context.Context, email, password string) (*User, error)

	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateBankAccount(ctx context.Context, userID uuid.UUID, role Role, in BankAccountInput) (*BankAccount, error)
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type BankAccountRepository interface {
	Create(ctx context.Context, ba *BankAccount) error
	UpsertByUserID(ctx context.Context, ba *BankAccount) (*BankAccount, error)
}

// ArtistRegistrar is implemented by the artist module and injected at wiring
// time so this package never imports artist (artist will later import user).
type ArtistRegistrar interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, description string) error
}

type Transactioner interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type RegisterInput struct {
	Username    string
	Email       string
	Password    string
	Role        Role
	BankAccount BankAccountInput
	// Artist is required when Role is artist, and must be nil for a customer.
	Artist *ArtistProfileInput
}

type BankAccountInput struct {
	BankName          string
	AccountHolderName string
	AccountNumber     string
}

type ArtistProfileInput struct {
	Description string
}
