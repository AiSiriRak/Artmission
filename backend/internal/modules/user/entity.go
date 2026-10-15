// Package user owns account identity and credentials.
package user

import (
	"io"
	"time"

	"github.com/google/uuid"
)

// Role is fixed at registration; an account is exactly one of these, never
// more than one.
type Role string

const (
	RoleCustomer Role = "customer"
	RoleArtist   Role = "artist"

	// RoleAdmin is reserved for seeded or operations-managed accounts
	// and cannot be selected during registration.
	RoleAdmin Role = "admin"
)

const (
	// MaxProfileImageSize is the maximum allowed profile image size in bytes.
	MaxProfileImageSize = 5 * 1024 * 1024
)

// User represents an account and its profile information.
type User struct {
	ID              uuid.UUID
	Username        string
	Email           string
	PasswordHash    string
	Role            Role
	ProfileImageKey *string
	ProfileImageURL *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// BankAccount represents a bank account associated with a user.
type BankAccount struct {
	UserID            uuid.UUID
	BankName          string
	AccountHolderName string
	AccountNumber     string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// RegisterInput contains the information required to register a user,
// including their bank account and optional artist profile information.
type RegisterInput struct {
	Username    string
	Email       string
	Password    string
	Role        Role
	BankAccount BankAccountInput
	// Artist may be nil when Role is artist (description is optional); must be nil for a customer.
	Artist *ArtistProfileInput
}

// BankAccountInput contains the bank account information supplied
// during registration or bank account updates.
type BankAccountInput struct {
	BankName          string
	AccountHolderName string
	AccountNumber     string
}

// UpdateAccountInput contains the information required to update
// an account's username and optionally change its password.
type UpdateAccountInput struct {
	Username    string
	OldPassword *string
	NewPassword *string
}

// ArtistProfileInput contains the information used to create
// or update an artist profile.
type ArtistProfileInput struct {
	Description *string
}

// UpdateProfileImageInput specifies whether to upload a new profile image
// or remove the existing profile image.
type UpdateProfileImageInput struct {
	ProfileImage       io.Reader
	RemoveProfileImage bool
}

// AccountUpdate contains the account fields that can be updated
// by the repository, including the update timestamp.
type AccountUpdate struct {
	Username     string
	PasswordHash *string
	UpdatedAt    time.Time
}
