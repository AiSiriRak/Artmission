package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

// AuthUsecase is the driving port for session lifecycle. Authenticate is
// used only by the HTTP auth middleware to validate an incoming request; it
// is not part of any user-facing feature.
type AuthUsecase interface {
	Login(ctx context.Context, username, password string) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	Authenticate(ctx context.Context, accessToken string) (*TokenClaims, error)
}

type authUsecase struct {
	userUsecase     user.UserUsecase
	sessionRepo     SessionRepository
	tokenIssuer     TokenIssuer
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	now             func() time.Time
}

func NewAuthUsecase(
	userUsecase user.UserUsecase,
	sessionRepo SessionRepository,
	tokenIssuer TokenIssuer,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) AuthUsecase {
	return &authUsecase{
		userUsecase:     userUsecase,
		sessionRepo:     sessionRepo,
		tokenIssuer:     tokenIssuer,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		now:             time.Now,
	}
}

func (u *authUsecase) Login(ctx context.Context, username, password string) (*AuthResult, error) {
	found, err := u.userUsecase.Authenticate(ctx, username, password)
	if err != nil {
		if err == user.ErrInvalidCredential {
			return nil, ErrInvalidCredential
		}
		return nil, err
	}
	return u.issueSession(ctx, found)
}

func (u *authUsecase) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	claims, err := u.tokenIssuer.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	session, err := u.sessionRepo.FindByID(ctx, claims.SessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}
	if err := u.validateSession(session, claims.UserID, refreshToken); err != nil {
		return nil, err
	}

	found, err := u.userUsecase.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	// Rotate: the old refresh token (and its session row) must not be
	// usable again once a new pair has been issued.
	if err := u.sessionRepo.DeleteByID(ctx, session.ID); err != nil {
		return nil, err
	}
	return u.issueSession(ctx, found)
}

func (u *authUsecase) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return u.sessionRepo.DeleteByID(ctx, sessionID)
}

func (u *authUsecase) Authenticate(ctx context.Context, accessToken string) (*TokenClaims, error) {
	claims, err := u.tokenIssuer.ParseAccessToken(accessToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	session, err := u.sessionRepo.FindByID(ctx, claims.SessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}
	if session.UserID != claims.UserID {
		return nil, ErrSessionNotFound
	}
	if session.ExpiresAt.Before(u.now()) {
		return nil, ErrSessionNotFound
	}
	return claims, nil
}

func (u *authUsecase) validateSession(session *Session, userID uuid.UUID, refreshToken string) error {
	if session.UserID != userID {
		return ErrSessionNotFound
	}
	if session.ExpiresAt.Before(u.now()) {
		return ErrSessionNotFound
	}
	if subtle.ConstantTimeCompare([]byte(session.RefreshTokenHash), []byte(hashToken(refreshToken))) != 1 {
		return ErrInvalidToken
	}
	return nil
}

func (u *authUsecase) issueSession(ctx context.Context, found *user.User) (*AuthResult, error) {
	now := u.now()
	sessionID := uuid.New()
	claims := TokenClaims{UserID: found.ID, SessionID: sessionID, Role: found.Role}

	accessExpiresAt := now.Add(u.accessTokenTTL)
	refreshExpiresAt := now.Add(u.refreshTokenTTL)

	accessToken, err := u.tokenIssuer.GenerateAccessToken(claims, accessExpiresAt)
	if err != nil {
		return nil, apperror.Internal("failed to generate access token", err)
	}
	refreshToken, err := u.tokenIssuer.GenerateRefreshToken(claims, refreshExpiresAt)
	if err != nil {
		return nil, apperror.Internal("failed to generate refresh token", err)
	}

	session := &Session{
		ID:               sessionID,
		UserID:           found.ID,
		RefreshTokenHash: hashToken(refreshToken),
		ExpiresAt:        refreshExpiresAt,
		CreatedAt:        now,
	}
	if err := u.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &AuthResult{
		User:                  found,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
