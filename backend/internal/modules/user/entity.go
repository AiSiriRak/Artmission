// Package user owns account identity and credentials: the User entity, its
// persistence port, and the usecases for registration and credential
// verification. Session/token concerns belong to modules/auth, which
// depends on this package's UserUsecase rather than duplicating password
// logic.
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
	Phone        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
