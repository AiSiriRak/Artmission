package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// User represents a system user account, including authentication credentials and profile information.
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID              uuid.UUID  `bun:"id,pk"`
	Username        string     `bun:"username"`
	Email           string     `bun:"email"`
	PasswordHash    string     `bun:"password_hash"`
	Role            string     `bun:"role"`
	ProfileImageKey *string    `bun:"profile_image_key"`
	CreatedAt       time.Time  `bun:"created_at,nullzero"`
	UpdatedAt       time.Time  `bun:"updated_at,nullzero"`
	DeletedAt       *time.Time `bun:"deleted_at,soft_delete,nullzero"`
}

// BankAccount represents financial payout account details associated with a user.
type BankAccount struct {
	bun.BaseModel `bun:"table:bank_accounts,alias:ba"`

	UserID            uuid.UUID `bun:"user_id,pk"`
	BankName          string    `bun:"bank_name"`
	AccountHolderName string    `bun:"account_holder_name"`
	AccountNumber     string    `bun:"account_number"`
	CreatedAt         time.Time `bun:"created_at,nullzero"`
	UpdatedAt         time.Time `bun:"updated_at,nullzero"`
}

// ArtistProfile represents extended profile details, showcase stats, and metadata for artist accounts.
type ArtistProfile struct {
	bun.BaseModel `bun:"table:artist_profiles,alias:ap"`

	UserID          uuid.UUID `bun:"user_id,pk"`
	Description     *string   `bun:"description"`
	ProfileImageKey *string   `bun:"profile_image_key,scanonly"`
	ArtistName      string    `bun:"artist_name,scanonly"`
	MinPriceSatang  *int64    `bun:"min_price_satang,scanonly"`
	MaxPriceSatang  *int64    `bun:"max_price_satang,scanonly"`
	ReviewScore     *float64  `bun:"review_score,scanonly"`
	CreatedAt       time.Time `bun:"created_at,nullzero"`
	UpdatedAt       time.Time `bun:"updated_at,nullzero"`
}
