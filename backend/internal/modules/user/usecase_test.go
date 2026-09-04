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

func TestRegister_Success(t *testing.T) {
	repo := newFakeRepo()
	usecase := user.NewUserUsecase(repo)

	got, err := usecase.Register(context.Background(), user.RegisterInput{
		Username: "alice",
		Email:    "alice@example.com",
		Phone:    "0123456789",
		Password: "password123",
		Role:     user.RoleCustomer,
	})
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
}

func TestRegister_RejectsAdminRole(t *testing.T) {
	usecase := user.NewUserUsecase(newFakeRepo())

	_, err := usecase.Register(context.Background(), user.RegisterInput{
		Username: "root", Email: "root@example.com", Phone: "0", Password: "password123",
		Role: user.RoleAdmin,
	})
	if !errors.Is(err, user.ErrInvalidRole) {
		t.Errorf("Register() error = %v, want ErrInvalidRole", err)
	}
}

func TestRegister_PropagatesDuplicateFromRepository(t *testing.T) {
	repo := newFakeRepo()
	repo.createErr = user.ErrEmailTaken
	usecase := user.NewUserUsecase(repo)

	_, err := usecase.Register(context.Background(), user.RegisterInput{
		Username: "alice", Email: "alice@example.com", Phone: "0", Password: "password123",
		Role: user.RoleCustomer,
	})
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

	usecase := user.NewUserUsecase(repo)
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

	usecase := user.NewUserUsecase(repo)
	_, err := usecase.Authenticate(context.Background(), "alice", "wrong-password")
	if !errors.Is(err, user.ErrInvalidCredential) {
		t.Errorf("Authenticate() error = %v, want ErrInvalidCredential", err)
	}
}

func TestAuthenticate_UnknownUsernameLooksLikeWrongPassword(t *testing.T) {
	usecase := user.NewUserUsecase(newFakeRepo())

	_, err := usecase.Authenticate(context.Background(), "nobody", "whatever")
	if !errors.Is(err, user.ErrInvalidCredential) {
		t.Errorf("Authenticate() error = %v, want ErrInvalidCredential (must not leak account existence)", err)
	}
}
