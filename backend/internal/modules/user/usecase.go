package user

import (
	"context"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/security"
	"github.com/google/uuid"
)

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

func (u *userUsecase) Authenticate(ctx context.Context, email, password string) (*User, error) {
	found, err := u.repo.GetByEmail(ctx, email)
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

func (u *userUsecase) UpdateBankAccount(ctx context.Context, userID uuid.UUID, role Role, in BankAccountInput) (*BankAccount, error) {
	if role != RoleCustomer && role != RoleArtist {
		return nil, ErrBankAccountNotAllowed
	}

	bankName := strings.TrimSpace(in.BankName)
	accountNumber := strings.TrimSpace(in.AccountNumber)
	if bankName == "" || accountNumber == "" {
		return nil, ErrBankAccountRequired
	}

	now := time.Now()
	bank := &BankAccount{
		UserID:        userID,
		BankName:      bankName,
		AccountNumber: accountNumber,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return u.bankRepo.UpsertByUserID(ctx, bank)
}
