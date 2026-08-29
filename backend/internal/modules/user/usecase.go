package user

import (
	"context"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/security"
	"github.com/google/uuid"
)

type BankAccountInput struct {
	BankName      string
	AccountNumber string
}

type ArtistProfileInput struct {
	Description string
}

type RegisterInput struct {
	Username    string
	Email       string
	FirstName   string
	LastName    string
	PhoneNumber string
	Password    string
	Role        Role
	BankAccount BankAccountInput
	// Artist is required when Role is artist, and must be nil for a customer.
	Artist *ArtistProfileInput
}

type UserUsecase interface {
	// Register creates the user, bank account, and (when role is artist)
	// artist profile in one transaction. RoleAdmin is never accepted
	// (admins are seeded/ops-managed, not self-registered).
	Register(ctx context.Context, in RegisterInput) (*User, error)

	// Authenticate returns ErrInvalidCredential for both "no such user"
	// and "wrong password" so a caller cannot distinguish account existence.
	Authenticate(ctx context.Context, username, password string) (*User, error)

	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type userUsecase struct {
	repo            UserRepository
	bankRepo        BankAccountRepository
	artistRegistrar ArtistRegistrar
	tx              Transactioner
}

func NewUserUsecase(
	repo UserRepository,
	bankRepo BankAccountRepository,
	artistRegistrar ArtistRegistrar,
	tx Transactioner,
) UserUsecase {
	return &userUsecase{
		repo:            repo,
		bankRepo:        bankRepo,
		artistRegistrar: artistRegistrar,
		tx:              tx,
	}
}

func (u *userUsecase) Register(ctx context.Context, in RegisterInput) (*User, error) {
	if in.Role != RoleCustomer && in.Role != RoleArtist {
		return nil, ErrInvalidRole
	}
	if in.Role == RoleArtist {
		if in.Artist == nil || strings.TrimSpace(in.Artist.Description) == "" {
			return nil, ErrArtistDescriptionRequired
		}
	} else if in.Artist != nil {
		return nil, ErrArtistFieldsNotAllowed
	}
	if strings.TrimSpace(in.BankAccount.BankName) == "" || strings.TrimSpace(in.BankAccount.AccountNumber) == "" {
		return nil, ErrBankAccountRequired
	}

	hash, err := security.HashPassword(in.Password)
	if err != nil {
		return nil, apperror.Internal("failed to hash password", err)
	}

	now := time.Now()
	newUser := &User{
		ID:           uuid.New(),
		Username:     strings.TrimSpace(in.Username),
		Email:        strings.TrimSpace(in.Email),
		FirstName:    strings.TrimSpace(in.FirstName),
		LastName:     strings.TrimSpace(in.LastName),
		PhoneNumber:  strings.TrimSpace(in.PhoneNumber),
		PasswordHash: hash,
		Role:         in.Role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	bank := &BankAccount{
		UserID:        newUser.ID,
		BankName:      strings.TrimSpace(in.BankAccount.BankName),
		AccountNumber: strings.TrimSpace(in.BankAccount.AccountNumber),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err = u.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := u.repo.Create(ctx, newUser); err != nil {
			return err
		}
		if err := u.bankRepo.Create(ctx, bank); err != nil {
			return err
		}
		if in.Role == RoleArtist {
			return u.artistRegistrar.CreateProfile(ctx, newUser.ID, strings.TrimSpace(in.Artist.Description))
		}
		return nil
	})
	if err != nil {
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
