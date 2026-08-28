package rest

import (
	"context"

	"github.com/DeepAung/artmission/backend/internal/modules/user"
	"github.com/google/uuid"
)

// AuthInfo is what the auth middleware injects into the request context
// after validating an access token, and what handlers/role guards read
// back out.
type AuthInfo struct {
	UserID    uuid.UUID
	Role      user.Role
	SessionID uuid.UUID
}

type authInfoKey struct{}

func withAuthInfo(ctx context.Context, info AuthInfo) context.Context {
	return context.WithValue(ctx, authInfoKey{}, info)
}

// authInfoFromContext returns the AuthInfo set by the auth middleware. ok
// is false if the route has no auth middleware or it hasn't run.
func authInfoFromContext(ctx context.Context) (AuthInfo, bool) {
	info, ok := ctx.Value(authInfoKey{}).(AuthInfo)
	return info, ok
}
