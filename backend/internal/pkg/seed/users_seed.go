package seed

import (
	"context"
	"fmt"
	"time"

	pgmodel "github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres/model"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/security"
	"github.com/google/uuid"
)

const seedPassword = "Password123!"

type userSeed struct {
	Key         string
	Username    string
	Email       string
	Role        string // "customer" | "artist"
	Description string // artist bio; empty for customers
}

var users = []userSeed{
	{Key: "customer-1", Username: "jane_customer", Email: "jane@artmission.test", Role: "customer"},
	{
		Key: "artist-1", Username: "ann_artist", Email: "ann@artmission.test", Role: "artist",
		Description: "Portrait and character illustrator with a soft, painterly style.",
	},
	{
		Key: "artist-2", Username: "bo_artist", Email: "bo@artmission.test", Role: "artist",
		Description: "Digital concept artist specializing in fantasy landscapes.",
	},
	{
		Key: "artist-3", Username: "cleo_artist", Email: "cleo@artmission.test", Role: "artist",
		Description: "Minimalist line-art and logo commissions.",
	},
}

func userID(key string) uuid.UUID { return id("user", key) }

func artistUsers() []userSeed {
	var artists []userSeed
	for _, u := range users {
		if u.Role == "artist" {
			artists = append(artists, u)
		}
	}
	return artists
}

func seedUsersUp(ctx context.Context, d deps) error {
	passwordHash, err := security.HashPassword(seedPassword)
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}

	now := time.Now()
	rows := make([]pgmodel.User, len(users))
	for i, u := range users {
		rows[i] = pgmodel.User{
			ID:           userID(u.Key),
			Username:     u.Username,
			Email:        u.Email,
			PasswordHash: passwordHash,
			Role:         u.Role,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	}
	if err := seedTable(ctx, d, rows, "id"); err != nil {
		return err
	}

	printSeedCredentials()
	return nil
}

func seedUsersDown(ctx context.Context, d deps) error {
	ids := make([]uuid.UUID, len(users))
	for i, u := range users {
		ids[i] = userID(u.Key)
	}
	return deleteByIDs[pgmodel.User](ctx, d, "id", ids)
}

func seedBankAccountsUp(ctx context.Context, d deps) error {
	now := time.Now()
	rows := make([]pgmodel.BankAccount, len(users))
	for i, u := range users {
		rows[i] = pgmodel.BankAccount{
			UserID:            userID(u.Key),
			BankName:          "Seed Bank",
			AccountHolderName: u.Username,
			AccountNumber:     fmt.Sprintf("SEED-%s", u.Key),
			CreatedAt:         now,
			UpdatedAt:         now,
		}
	}
	return seedTable(ctx, d, rows, "user_id")
}

func seedArtistProfilesUp(ctx context.Context, d deps) error {
	now := time.Now()
	artists := artistUsers()
	rows := make([]pgmodel.ArtistProfile, len(artists))
	for i, a := range artists {
		description := a.Description
		rows[i] = pgmodel.ArtistProfile{
			UserID:      userID(a.Key),
			Description: &description,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
	return seedTable(ctx, d, rows, "user_id")
}

// printSeedCredentials prints every seeded account's login so a frontend
// dev can sign in immediately after "seed up".
func printSeedCredentials() {
	fmt.Println("Seeded accounts (password for all:", seedPassword+"):")
	for _, u := range users {
		fmt.Printf("  %-10s %s\n", u.Role, u.Email)
	}
}
