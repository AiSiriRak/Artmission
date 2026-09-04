// Package user owns account identity and credentials.
package user

import (
	"time"

	"github.com/google/uuid"
)

// Role is fixed at registration; an account is exactly one of these, never
// more than one.
type Role string

const (
	RoleCustomer Role = "customer"
	RoleArtist   Role = "artist"
	RoleAdmin    Role = "admin" // seeded/ops-managed only; not selectable at registration.
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	FirstName    string
	LastName     string
	PhoneNumber  string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type BankAccount struct {
	UserID        uuid.UUID
	BankName      string
	AccountNumber string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
