package user

import (
	"context"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
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
	UpdateAccount(ctx context.Context, id uuid.UUID, in UpdateAccountInput) (*User, error)
	UpdateBankAccount(ctx context.Context, userID uuid.UUID, role Role, in BankAccountInput) (*BankAccount, error)
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateAccountByID(ctx context.Context, id uuid.UUID, in AccountUpdate) (*User, error)
}

type BankAccountRepository interface {
	Create(ctx context.Context, ba *BankAccount) error
	UpsertByUserID(ctx context.Context, ba *BankAccount) (*BankAccount, error)
}

type AccountDeletionRepository interface {
	LockUserByIDForDeletion(ctx context.Context, userID uuid.UUID) error
	HasOrdersInStatuses(ctx context.Context, userID uuid.UUID, statuses []order.Status) (bool, error)
	DeleteBankAccountByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error
	SoftDeleteUserByID(ctx context.Context, userID uuid.UUID) error
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

type UpdateAccountInput struct {
	Username    string
	OldPassword *string
	NewPassword *string
}

type AccountUpdate struct {
	Username     string
	PasswordHash *string
	UpdatedAt    time.Time
}

type ArtistProfileInput struct {
	Description string
}
