package rest

import (
	"github.com/danielgtaylor/huma/v2"
)

// RegisterRoutes attaches every REST operation to api.
//
// Liveness/readiness are registered separately, unprefixed and outside huma/OpenAPI — see internal/pkg/httpserver.
func RegisterRoutes(api huma.API, authHandler *AuthHandler, orderHandler *OrderHandler) {
	authHandler.Register(api)
	orderHandler.Register(api)
}
