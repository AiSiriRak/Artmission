// Package token implements auth.TokenIssuer with signed JWTs.
package token

import (
	"errors"
	"time"

	"github.com/DeepAung/artmission/backend/internal/modules/auth"
	"github.com/DeepAung/artmission/backend/internal/modules/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// tokenKind is embedded in the JWT so an access token can never be
// replayed as a refresh token (or vice versa) even though both are signed
// with the same secret and carry the same TokenClaims shape.
type tokenKind string

const (
	kindAccess  tokenKind = "access"
	kindRefresh tokenKind = "refresh"
)

type jwtClaims struct {
	jwt.RegisteredClaims
	Kind      tokenKind `json:"kind"`
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	Role      string    `json:"role"`
}

type jwtIssuer struct {
	secret []byte
}

var _ auth.TokenIssuer = (*jwtIssuer)(nil)

func NewJWTIssuer(secret string) auth.TokenIssuer {
	return &jwtIssuer{secret: []byte(secret)}
}

func (i *jwtIssuer) GenerateAccessToken(claims auth.TokenClaims, expiresAt time.Time) (string, error) {
	return i.sign(claims, kindAccess, expiresAt)
}

func (i *jwtIssuer) GenerateRefreshToken(claims auth.TokenClaims, expiresAt time.Time) (string, error) {
	return i.sign(claims, kindRefresh, expiresAt)
}

func (i *jwtIssuer) ParseAccessToken(tokenString string) (*auth.TokenClaims, error) {
	return i.parse(tokenString, kindAccess)
}

func (i *jwtIssuer) ParseRefreshToken(tokenString string) (*auth.TokenClaims, error) {
	return i.parse(tokenString, kindRefresh)
}

func (i *jwtIssuer) sign(claims auth.TokenClaims, kind tokenKind, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Kind:      kind,
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		Role:      string(claims.Role),
	})
	return token.SignedString(i.secret)
}

func (i *jwtIssuer) parse(tokenString string, want tokenKind) (*auth.TokenClaims, error) {
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Kind != want {
		return nil, errors.New("unexpected token kind")
	}

	return &auth.TokenClaims{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		Role:      user.Role(claims.Role),
	}, nil
}
