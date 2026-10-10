package user

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/security"
	"github.com/google/uuid"
)

type userUsecase struct {
	repo            UserRepository
	bankRepo        BankAccountRepository
	artistRegistrar ArtistRegistrar
	deletionRepo    AccountDeletionRepository
	tx              Transactioner
	storage         ObjectStorage
}

func NewUserUsecase(
	repo UserRepository,
	bankRepo BankAccountRepository,
	artistRegistrar ArtistRegistrar,
	deletionRepo AccountDeletionRepository,
	tx Transactioner,
	storage ObjectStorage,
) UserUsecase {
	return &userUsecase{
		repo:            repo,
		bankRepo:        bankRepo,
		artistRegistrar: artistRegistrar,
		deletionRepo:    deletionRepo,
		tx:              tx,
		storage:         storage,
	}
}

func (u *userUsecase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	blockingStatuses := []order.Status{order.StatusPending, order.StatusNotPaid, order.StatusInProcess}

	return u.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := u.deletionRepo.LockUserByIDForDeletion(ctx, userID); err != nil {
			return err
		}
		hasActiveOrders, err := u.deletionRepo.HasOrdersInStatuses(ctx, userID, blockingStatuses)
		if err != nil {
			return err
		}
		if hasActiveOrders {
			return ErrActiveOrders
		}
		if err := u.deletionRepo.DeleteBankAccountByUserID(ctx, userID); err != nil {
			return err
		}
		if err := u.deletionRepo.DeleteSessionsByUserID(ctx, userID); err != nil {
			return err
		}
		return u.deletionRepo.SoftDeleteUserByID(ctx, userID)
	})
}

func (u *userUsecase) Register(ctx context.Context, in RegisterInput) (*User, error) {
	if in.Role != RoleCustomer && in.Role != RoleArtist {
		return nil, ErrInvalidRole
	}
	if in.Role != RoleArtist && in.Artist != nil {
		return nil, ErrArtistFieldsNotAllowed
	}
	if strings.TrimSpace(in.BankAccount.BankName) == "" || strings.TrimSpace(in.BankAccount.AccountHolderName) == "" || strings.TrimSpace(in.BankAccount.AccountNumber) == "" {
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
		UserID:            newUser.ID,
		BankName:          strings.TrimSpace(in.BankAccount.BankName),
		AccountHolderName: strings.TrimSpace(in.BankAccount.AccountHolderName),
		AccountNumber:     strings.TrimSpace(in.BankAccount.AccountNumber),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err = u.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := u.repo.Create(ctx, newUser); err != nil {
			return err
		}
		if err := u.bankRepo.Create(ctx, bank); err != nil {
			return err
		}
		if in.Role == RoleArtist {
			var description *string
			if in.Artist != nil {
				description = in.Artist.Description
			}
			return u.artistRegistrar.CreateProfile(ctx, newUser.ID, description)
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
	return u.attachProfileImageURL(found), nil
}

func (u *userUsecase) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	account, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return u.attachProfileImageURL(account), nil
}

func (u *userUsecase) GetArtistName(ctx context.Context, id uuid.UUID) (string, error) {
	account, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if account == nil {
		return "", ErrUserNotFound
	}
	return account.Username, nil
}

func (u *userUsecase) UpdateAccount(ctx context.Context, id uuid.UUID, in UpdateAccountInput) (*User, error) {
	username := strings.TrimSpace(in.Username)

	oldPasswordProvided := in.OldPassword != nil
	newPasswordProvided := in.NewPassword != nil
	if oldPasswordProvided != newPasswordProvided {
		return nil, ErrPasswordFieldsRequired
	}

	var passwordHash *string
	if oldPasswordProvided {
		found, err := u.repo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if !security.VerifyPassword(found.PasswordHash, *in.OldPassword) {
			return nil, ErrInvalidCurrentPassword
		}

		hash, err := security.HashPassword(*in.NewPassword)
		if err != nil {
			return nil, apperror.Internal("failed to hash password", err)
		}
		passwordHash = &hash
	}

	account, err := u.repo.UpdateAccountByID(ctx, id, AccountUpdate{
		Username:     username,
		PasswordHash: passwordHash,
		UpdatedAt:    time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return u.attachProfileImageURL(account), nil
}

func (u *userUsecase) UpdateBankAccount(ctx context.Context, userID uuid.UUID, role Role, in BankAccountInput) (*BankAccount, error) {
	if role != RoleCustomer && role != RoleArtist {
		return nil, ErrBankAccountNotAllowed
	}

	bankName := strings.TrimSpace(in.BankName)
	accountHolderName := strings.TrimSpace(in.AccountHolderName)
	accountNumber := strings.TrimSpace(in.AccountNumber)
	if bankName == "" || accountHolderName == "" || accountNumber == "" {
		return nil, ErrBankAccountRequired
	}

	now := time.Now()
	bank := &BankAccount{
		UserID:            userID,
		BankName:          bankName,
		AccountHolderName: accountHolderName,
		AccountNumber:     accountNumber,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return u.bankRepo.UpsertByUserID(ctx, bank)
}

func (u *userUsecase) GetBankAccount(ctx context.Context, userID uuid.UUID, role Role) (*BankAccount, error) {
	if role != RoleCustomer && role != RoleArtist {
		return nil, ErrBankAccountNotAllowed
	}
	return u.bankRepo.GetByUserID(ctx, userID)
}

func (u *userUsecase) UpdateProfileImage(
	ctx context.Context,
	id uuid.UUID,
	in UpdateProfileImageInput,
) (*User, error) {
	if in.ProfileImage == nil && !in.RemoveProfileImage {
		return nil, ErrNoProfileChanges
	}
	if in.ProfileImage != nil && in.RemoveProfileImage {
		return nil, ErrConflictingProfileImageChange
	}

	currentUser, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var newKey *string
	if in.ProfileImage != nil {
		content, contentType, extension, err := readProfileImage(in.ProfileImage)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf("profile-images/%s/%s.%s", id, uuid.New(), extension)
		if err := u.storage.Upload(ctx, key, bytes.NewReader(content), contentType); err != nil {
			return nil, apperror.Internal("failed to upload profile image", err)
		}
		newKey = &key
	}

	updatedUser, err := u.repo.UpdateProfileImageByID(ctx, id, newKey)
	if err != nil {
		if newKey != nil {
			_ = u.storage.Delete(ctx, *newKey)
		}
		return nil, err
	}

	// Delete the old one if there's change or delete
	if currentUser.ProfileImageKey != nil && (in.RemoveProfileImage || newKey != nil) {
		_ = u.storage.Delete(ctx, *currentUser.ProfileImageKey)
	}

	return u.attachProfileImageURL(updatedUser), nil
}

func (u *userUsecase) attachProfileImageURL(account *User) *User {
	if account.ProfileImageKey == nil {
		account.ProfileImageURL = nil
		return account
	}
	url := u.storage.PublicURL(*account.ProfileImageKey)
	account.ProfileImageURL = &url
	return account
}

func readProfileImage(reader io.Reader) ([]byte, string, string, error) {
	content, err := io.ReadAll(io.LimitReader(reader, MaxProfileImageSize+1))
	if err != nil {
		return nil, "", "", apperror.InvalidInput("failed to read profile image", err)
	}
	if len(content) > MaxProfileImageSize {
		return nil, "", "", ErrProfileImageTooLarge
	}
	contentType, extension, ok := detectProfileImageType(content)
	if !ok {
		return nil, "", "", ErrInvalidProfileImage
	}
	return content, contentType, extension, nil
}

func detectProfileImageType(content []byte) (string, string, bool) {
	switch {
	case len(content) >= 3 && bytes.Equal(content[:3], []byte{0xff, 0xd8, 0xff}):
		return "image/jpeg", "jpg", true
	case len(content) >= 8 && bytes.Equal(content[:8], []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}):
		return "image/png", "png", true
	case len(content) >= 12 && string(content[:4]) == "RIFF" && string(content[8:12]) == "WEBP":
		return "image/webp", "webp", true
	default:
		return "", "", false
	}
}
