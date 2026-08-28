package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DeepAung/artmission/backend/internal/modules/user"
	"github.com/google/uuid"
)

// --- fakes ---

type fakeUserUsecase struct {
	byUsername map[string]*user.User
	byID       map[uuid.UUID]*user.User
	passwords  map[string]string
}

func newFakeUserUsecase() *fakeUserUsecase {
	return &fakeUserUsecase{
		byUsername: map[string]*user.User{},
		byID:       map[uuid.UUID]*user.User{},
		passwords:  map[string]string{},
	}
}

func (f *fakeUserUsecase) addUser(u *user.User, password string) {
	f.byUsername[u.Username] = u
	f.byID[u.ID] = u
	f.passwords[u.Username] = password
}

func (f *fakeUserUsecase) Register(context.Context, user.RegisterInput) (*user.User, error) {
	panic("not used by auth usecase tests")
}

func (f *fakeUserUsecase) Authenticate(_ context.Context, username, password string) (*user.User, error) {
	u, ok := f.byUsername[username]
	if !ok || f.passwords[username] != password {
		return nil, user.ErrInvalidCredential
	}
	return u, nil
}

func (f *fakeUserUsecase) GetByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

var _ user.UserUsecase = (*fakeUserUsecase)(nil)

type fakeSessionRepo struct {
	sessions map[uuid.UUID]*Session
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{sessions: map[uuid.UUID]*Session{}}
}

func (f *fakeSessionRepo) Create(_ context.Context, s *Session) error {
	cp := *s
	f.sessions[s.ID] = &cp
	return nil
}

func (f *fakeSessionRepo) FindByID(_ context.Context, id uuid.UUID) (*Session, error) {
	s, ok := f.sessions[id]
	if !ok {
		return nil, ErrSessionNotFound
	}
	cp := *s
	return &cp, nil
}

func (f *fakeSessionRepo) DeleteByID(_ context.Context, id uuid.UUID) error {
	if _, ok := f.sessions[id]; !ok {
		return ErrSessionNotFound
	}
	delete(f.sessions, id)
	return nil
}

var _ SessionRepository = (*fakeSessionRepo)(nil)

// fakeTokenIssuer mimics the real jwtIssuer's access/refresh type
// discrimination without real signing, so tests stay fast and deterministic.
type fakeTokenIssuer struct {
	tokens map[string]TokenClaims
	seq    int
}

func newFakeTokenIssuer() *fakeTokenIssuer {
	return &fakeTokenIssuer{tokens: map[string]TokenClaims{}}
}

func (f *fakeTokenIssuer) GenerateAccessToken(claims TokenClaims, _ time.Time) (string, error) {
	return f.issue("access", claims), nil
}

func (f *fakeTokenIssuer) GenerateRefreshToken(claims TokenClaims, _ time.Time) (string, error) {
	return f.issue("refresh", claims), nil
}

func (f *fakeTokenIssuer) issue(kind string, claims TokenClaims) string {
	f.seq++
	token := fmt.Sprintf("%s-%d", kind, f.seq)
	f.tokens[token] = claims
	return token
}

func (f *fakeTokenIssuer) ParseAccessToken(token string) (*TokenClaims, error) {
	return f.parse("access", token)
}

func (f *fakeTokenIssuer) ParseRefreshToken(token string) (*TokenClaims, error) {
	return f.parse("refresh", token)
}

func (f *fakeTokenIssuer) parse(kind, token string) (*TokenClaims, error) {
	if !strings.HasPrefix(token, kind+"-") {
		return nil, errors.New("unexpected token type")
	}
	claims, ok := f.tokens[token]
	if !ok {
		return nil, errors.New("unknown token")
	}
	return &claims, nil
}

var _ TokenIssuer = (*fakeTokenIssuer)(nil)

// --- test setup ---

func newTestAuthUsecase() (*authUsecase, *fakeUserUsecase, *fakeSessionRepo, *fakeTokenIssuer) {
	userUsecase := newFakeUserUsecase()
	sessionRepo := newFakeSessionRepo()
	tokenIssuer := newFakeTokenIssuer()
	usecase := NewAuthUsecase(userUsecase, sessionRepo, tokenIssuer, time.Hour, 7*24*time.Hour).(*authUsecase)
	return usecase, userUsecase, sessionRepo, tokenIssuer
}

var alice = &user.User{ID: uuid.New(), Username: "alice", Role: user.RoleCustomer}

// --- tests ---

func TestLogin_Success(t *testing.T) {
	authUsecase, userUsecase, sessionRepo, _ := newTestAuthUsecase()
	userUsecase.addUser(alice, "secret")

	result, err := authUsecase.Login(context.Background(), "alice", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}
	if result.User.ID != alice.ID {
		t.Errorf("Login() user = %v, want %v", result.User.ID, alice.ID)
	}
	if !strings.HasPrefix(result.AccessToken, "access-") {
		t.Errorf("Login() access token = %q, want access- prefix", result.AccessToken)
	}
	if !strings.HasPrefix(result.RefreshToken, "refresh-") {
		t.Errorf("Login() refresh token = %q, want refresh- prefix", result.RefreshToken)
	}
	if len(sessionRepo.sessions) != 1 {
		t.Errorf("Login() created %d sessions, want 1", len(sessionRepo.sessions))
	}
}

func TestLogin_WrongPasswordCreatesNoSession(t *testing.T) {
	uc, userUsecase, sessionRepo, _ := newTestAuthUsecase()
	userUsecase.addUser(alice, "secret")

	_, err := uc.Login(context.Background(), "alice", "wrong")
	if !errors.Is(err, ErrInvalidCredential) {
		t.Errorf("Login() error = %v, want ErrInvalidCredential", err)
	}
	if len(sessionRepo.sessions) != 0 {
		t.Errorf("Login() with bad credentials created %d sessions, want 0", len(sessionRepo.sessions))
	}
}

func TestLogout_InvalidatesSessionImmediately(t *testing.T) {
	uc, userUsecase, _, tokenIssuer := newTestAuthUsecase()
	userUsecase.addUser(alice, "secret")

	result, err := uc.Login(context.Background(), "alice", "secret")
	if err != nil {
		t.Fatal(err)
	}

	// Confirm the access token is valid before logout.
	if _, err := uc.Authenticate(context.Background(), result.AccessToken); err != nil {
		t.Fatalf("Authenticate() before logout error = %v, want nil", err)
	}

	claims, _ := tokenIssuer.ParseAccessToken(result.AccessToken)
	if err := uc.Logout(context.Background(), claims.SessionID); err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}

	// The access token's JWT is still technically unexpired, but the
	// session backing it is gone: Authenticate must reject it immediately.
	if _, err := uc.Authenticate(context.Background(), result.AccessToken); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("Authenticate() after logout error = %v, want ErrSessionNotFound", err)
	}
}

func TestRefresh_RotatesSessionAndInvalidatesOldToken(t *testing.T) {
	uc, userUsecase, sessionRepo, _ := newTestAuthUsecase()
	userUsecase.addUser(alice, "secret")

	first, err := uc.Login(context.Background(), "alice", "secret")
	if err != nil {
		t.Fatal(err)
	}

	second, err := uc.Refresh(context.Background(), first.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v, want nil", err)
	}
	if second.RefreshToken == first.RefreshToken {
		t.Error("Refresh() returned the same refresh token instead of rotating it")
	}
	if len(sessionRepo.sessions) != 1 {
		t.Errorf("Refresh() left %d sessions, want exactly 1 (old rotated out)", len(sessionRepo.sessions))
	}

	// The old refresh token's session is gone; reusing it must fail.
	if _, err := uc.Refresh(context.Background(), first.RefreshToken); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("Refresh() with rotated-out token error = %v, want ErrSessionNotFound", err)
	}
}

func TestRefresh_RejectsAccessTokenUsedAsRefreshToken(t *testing.T) {
	uc, userUsecase, _, _ := newTestAuthUsecase()
	userUsecase.addUser(alice, "secret")

	result, err := uc.Login(context.Background(), "alice", "secret")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := uc.Refresh(context.Background(), result.AccessToken); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Refresh(accessToken) error = %v, want ErrInvalidToken", err)
	}
}

func TestAuthenticate_RejectsExpiredSession(t *testing.T) {
	usecase, userUsecase, _, _ := newTestAuthUsecase()
	userUsecase.addUser(alice, "secret")

	// Freeze time in the past relative to login so the session is already
	// expired the moment it's created.
	base := time.Now()
	usecase.now = func() time.Time { return base }
	usecase.refreshTokenTTL = -time.Hour // ExpiresAt = base - 1h, already expired

	result, err := usecase.Login(context.Background(), "alice", "secret")
	if err != nil {
		t.Fatal(err)
	}

	usecase.now = func() time.Time { return base } // Authenticate checks against "now" >= ExpiresAt
	if _, err := usecase.Authenticate(context.Background(), result.AccessToken); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("Authenticate() with expired session error = %v, want ErrSessionNotFound", err)
	}
}
