package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/security"
	"github.com/google/uuid"
)

type fakeRepo struct {
	byUsername map[string]*user.User
	byID       map[uuid.UUID]*user.User
	createErr  error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byUsername: map[string]*user.User{}, byID: map[uuid.UUID]*user.User{}}
}

func (f *fakeRepo) Create(_ context.Context, u *user.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.byUsername[u.Username] = u
	f.byID[u.ID] = u
	return nil
}

func (f *fakeRepo) GetByUsername(_ context.Context, username string) (*user.User, error) {
	u, ok := f.byUsername[username]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

var _ user.UserRepository = (*fakeRepo)(nil)

type fakeBankRepo struct {
	byUserID  map[uuid.UUID]*user.BankAccount
	createErr error
}

func newFakeBankRepo() *fakeBankRepo {
	return &fakeBankRepo{byUserID: map[uuid.UUID]*user.BankAccount{}}
}

func (f *fakeBankRepo) Create(_ context.Context, ba *user.BankAccount) error {
	if f.createErr != nil {
		return f.createErr
	}
	cp := *ba
	f.byUserID[ba.UserID] = &cp
	return nil
}

var _ user.BankAccountRepository = (*fakeBankRepo)(nil)

type fakeArtistRegistrar struct {
	profiles map[uuid.UUID]string
	err      error
}

func newFakeArtistRegistrar() *fakeArtistRegistrar {
	return &fakeArtistRegistrar{profiles: map[uuid.UUID]string{}}
}

func (f *fakeArtistRegistrar) CreateProfile(_ context.Context, userID uuid.UUID, description string) error {
	if f.err != nil {
		return f.err
	}
	f.profiles[userID] = description
	return nil
}

var _ user.ArtistRegistrar = (*fakeArtistRegistrar)(nil)

type fakeTx struct{}

func (f *fakeTx) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

var _ user.Transactioner = (*fakeTx)(nil)

func newUsecase(repo *fakeRepo, bank *fakeBankRepo, artist *fakeArtistRegistrar) user.UserUsecase {
	return user.NewUserUsecase(repo, bank, artist, &fakeTx{})
}

func customerInput() user.RegisterInput {
	return user.RegisterInput{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
		Role:     user.RoleCustomer,
		BankAccount: user.BankAccountInput{
			BankName:      "Kasikorn",
			AccountNumber: "1234567890",
		},
	}
}

func TestRegister_CustomerCreatesUserAndBank(t *testing.T) {
	repo := newFakeRepo()
	bank := newFakeBankRepo()
	artist := newFakeArtistRegistrar()
	usecase := newUsecase(repo, bank, artist)

	got, err := usecase.Register(context.Background(), customerInput())
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}
	if got.ID == uuid.Nil {
		t.Error("Register() did not assign an id")
	}
	if got.PasswordHash == "password123" {
		t.Error("Register() stored the plaintext password instead of a hash")
	}
	if !security.VerifyPassword(got.PasswordHash, "password123") {
		t.Error("Register() stored hash does not verify against the original password")
	}
	if _, ok := repo.byUsername["alice"]; !ok {
		t.Error("Register() did not persist the user via the repository")
	}
	if ba, ok := bank.byUserID[got.ID]; !ok {
		t.Error("Register() did not persist a bank account")
	} else if ba.BankName != "Kasikorn" || ba.AccountNumber != "1234567890" {
		t.Errorf("Register() bank = %+v", ba)
	}
	if len(artist.profiles) != 0 {
		t.Errorf("Register(customer) created %d artist profiles, want 0", len(artist.profiles))
	}
}

func TestRegister_ArtistCreatesProfile(t *testing.T) {
	repo := newFakeRepo()
	bank := newFakeBankRepo()
	artist := newFakeArtistRegistrar()
	usecase := newUsecase(repo, bank, artist)

	in := customerInput()
	in.Username = "bob"
	in.Email = "bob@example.com"
	in.Role = user.RoleArtist
	in.Artist = &user.ArtistProfileInput{Description: "I paint portraits"}

	got, err := usecase.Register(context.Background(), in)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}
	if desc, ok := artist.profiles[got.ID]; !ok || desc != "I paint portraits" {
		t.Errorf("Register(artist) profile = %q, ok=%v", desc, ok)
	}
	if _, ok := bank.byUserID[got.ID]; !ok {
		t.Error("Register(artist) did not persist a bank account")
	}
}

func TestRegister_ArtistMissingDescription(t *testing.T) {
	usecase := newUsecase(newFakeRepo(), newFakeBankRepo(), newFakeArtistRegistrar())

	in := customerInput()
	in.Role = user.RoleArtist
	_, err := usecase.Register(context.Background(), in)
	if !errors.Is(err, user.ErrArtistDescriptionRequired) {
		t.Errorf("Register() error = %v, want ErrArtistDescriptionRequired", err)
	}
}

func TestRegister_CustomerRejectsArtistFields(t *testing.T) {
	usecase := newUsecase(newFakeRepo(), newFakeBankRepo(), newFakeArtistRegistrar())

	in := customerInput()
	in.Artist = &user.ArtistProfileInput{Description: "should not be here"}
	_, err := usecase.Register(context.Background(), in)
	if !errors.Is(err, user.ErrArtistFieldsNotAllowed) {
		t.Errorf("Register() error = %v, want ErrArtistFieldsNotAllowed", err)
	}
}

func TestRegister_RejectsMissingBankAccount(t *testing.T) {
	usecase := newUsecase(newFakeRepo(), newFakeBankRepo(), newFakeArtistRegistrar())

	in := customerInput()
	in.BankAccount = user.BankAccountInput{}
	_, err := usecase.Register(context.Background(), in)
	if !errors.Is(err, user.ErrBankAccountRequired) {
		t.Errorf("Register() error = %v, want ErrBankAccountRequired", err)
	}
}

func TestRegister_RejectsAdminRole(t *testing.T) {
	usecase := newUsecase(newFakeRepo(), newFakeBankRepo(), newFakeArtistRegistrar())

	in := customerInput()
	in.Username = "root"
	in.Email = "root@example.com"
	in.Role = user.RoleAdmin
	_, err := usecase.Register(context.Background(), in)
	if !errors.Is(err, user.ErrInvalidRole) {
		t.Errorf("Register() error = %v, want ErrInvalidRole", err)
	}
}

func TestRegister_PropagatesDuplicateFromRepository(t *testing.T) {
	repo := newFakeRepo()
	repo.createErr = user.ErrEmailTaken
	usecase := newUsecase(repo, newFakeBankRepo(), newFakeArtistRegistrar())

	_, err := usecase.Register(context.Background(), customerInput())
	if !errors.Is(err, user.ErrEmailTaken) {
		t.Errorf("Register() error = %v, want ErrEmailTaken", err)
	}
}

func TestAuthenticate_Success(t *testing.T) {
	repo := newFakeRepo()
	hash, err := security.HashPassword("correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	stored := &user.User{ID: uuid.New(), Username: "alice", PasswordHash: hash, Role: user.RoleCustomer}
	repo.byUsername["alice"] = stored

	usecase := newUsecase(repo, newFakeBankRepo(), newFakeArtistRegistrar())
	got, err := usecase.Authenticate(context.Background(), "alice", "correct-horse")
	if err != nil {
		t.Fatalf("Authenticate() error = %v, want nil", err)
	}
	if got.ID != stored.ID {
		t.Errorf("Authenticate() returned user %v, want %v", got.ID, stored.ID)
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	repo := newFakeRepo()
	hash, _ := security.HashPassword("correct-horse")
	repo.byUsername["alice"] = &user.User{ID: uuid.New(), Username: "alice", PasswordHash: hash}

	usecase := newUsecase(repo, newFakeBankRepo(), newFakeArtistRegistrar())
	_, err := usecase.Authenticate(context.Background(), "alice", "wrong-password")
	if !errors.Is(err, user.ErrInvalidCredential) {
		t.Errorf("Authenticate() error = %v, want ErrInvalidCredential", err)
	}
}

func TestAuthenticate_UnknownUsernameLooksLikeWrongPassword(t *testing.T) {
	usecase := newUsecase(newFakeRepo(), newFakeBankRepo(), newFakeArtistRegistrar())

	_, err := usecase.Authenticate(context.Background(), "nobody", "whatever")
	if !errors.Is(err, user.ErrInvalidCredential) {
		t.Errorf("Authenticate() error = %v, want ErrInvalidCredential (must not leak account existence)", err)
	}
}
