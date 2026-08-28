package rest

import (
	"net/http"
	"strings"

	"github.com/DeepAung/artmission/backend/internal/modules/auth"
	"github.com/DeepAung/artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
)

// requireAuth extracts a Bearer access token, validates it, and injects the
// resulting AuthInfo into the request context for downstream handlers and
// requireRole. Attach per-operation via huma.Operation.Middlewares; it is
// intentionally not global since not every route needs authentication.
func requireAuth(api huma.API, authUsecase auth.AuthUsecase) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		header := ctx.Header("Authorization")
		scheme, token, found := strings.Cut(header, " ")
		if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "missing or malformed Authorization header")
			return
		}

		claims, err := authUsecase.Authenticate(ctx.Context(), token)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "invalid or expired access token")
			return
		}

		info := AuthInfo{UserID: claims.UserID, Role: claims.Role, SessionID: claims.SessionID}
		next(huma.WithContext(ctx, withAuthInfo(ctx.Context(), info)))
	}
}

// requireRole must run after requireAuth in the same operation's middleware
// chain. It rejects the request with 403 if the authenticated user's role
// doesn't match.
func requireRole(api huma.API, role user.Role) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		info, ok := authInfoFromContext(ctx.Context())
		if !ok || info.Role != role {
			huma.WriteErr(api, ctx, http.StatusForbidden, "insufficient permissions")
			return
		}
		next(ctx)
	}
}
