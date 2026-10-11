// Package token implements auth.TokenIssuer with signed JWTs.
package token

import (
	"errors"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
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

// NewJWTIssuer returns a JWT-backed issuer that signs and validates access and refresh tokens with the supplied secret.
func NewJWTIssuer(secret string) auth.TokenIssuer {
	return &jwtIssuer{secret: []byte(secret)}
}

// GenerateAccessToken signs a new access JWT with the given claims and expiry.
func (i *jwtIssuer) GenerateAccessToken(claims auth.TokenClaims, expiresAt time.Time) (string, error) {
	return i.sign(claims, kindAccess, expiresAt)
}

// GenerateRefreshToken signs a new refresh JWT with the given claims and expiry.
func (i *jwtIssuer) GenerateRefreshToken(claims auth.TokenClaims, expiresAt time.Time) (string, error) {
	return i.sign(claims, kindRefresh, expiresAt)
}

// ParseAccessToken validates an access token and returns the embedded claims.
func (i *jwtIssuer) ParseAccessToken(tokenString string) (*auth.TokenClaims, error) {
	return i.parse(tokenString, kindAccess)
}

// ParseRefreshToken validates a refresh token and returns the embedded claims.
func (i *jwtIssuer) ParseRefreshToken(tokenString string) (*auth.TokenClaims, error) {
	return i.parse(tokenString, kindRefresh)
}

// sign creates a JWT carrying the token kind, user data, and expiry.
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

// parse verifies the JWT signature and expected kind, then decodes the claims.
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
