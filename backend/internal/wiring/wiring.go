// Package wiring is the single composition root: the one place every
// adapter, usecase, and handler is constructed and connected, in
// dependency order. cmd/cmd_serve.go (the real server) and
// tests/internal/apptest (the BDD suite's in-process server) both call
// Wire so the object graph they exercise can never diverge — add a
// module here once and every caller gets it.
package wiring

import (
	"log/slog"

	"github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres"
	"github.com/AiSiriRak/Artmission/backend/internal/adapters/token"
	"github.com/AiSiriRak/Artmission/backend/internal/handler/rest"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/artist"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/uptrace/bun"
)

// Config is the wiring graph's inputs. DB and Logger are live objects
// (a connection, a configured logger) rather than settings, since callers
// build those differently. App and Auth are the real config package's structs,
// reused as-is (not re-declared) so a config field renamed there can't
// silently drift out of sync here — the compiler catches it.
type Config struct {
	DB     *bun.DB
	Logger *slog.Logger
	App    config.App
	Auth   config.Auth
}

// Wire builds the entire object graph — adapters, usecases, handlers,
// routes — and returns the ready-to-serve HTTP server. The caller owns
// the server's lifecycle.
func Wire(cfg Config) *httpserver.Server {
	userRepo := postgres.NewUserRepository(cfg.DB)
	bankRepo := postgres.NewBankAccountRepository(cfg.DB)
	artistRepo := postgres.NewArtistRepository(cfg.DB)
	sessionRepo := postgres.NewSessionRepository(cfg.DB)
	orderRepo := postgres.NewOrderRepository(cfg.DB)
	tokenIssuer := token.NewJWTIssuer(cfg.Auth.JWTSecret)
	tx := baserepo.NewTransactioner(cfg.DB)

	artistUsecase := artist.NewProfileUsecase(artistRepo)
	userUsecase := user.NewUserUsecase(userRepo, bankRepo, artistUsecase, tx)
	authUsecase := auth.NewAuthUsecase(userUsecase, sessionRepo, tokenIssuer, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL)
	orderUsecase := order.NewOrderUsecase(orderRepo)

	authHandler := rest.NewAuthHandler(userUsecase, authUsecase, cfg.App.BasePath, cfg.App.IsProduction, cfg.Auth.RefreshCookieDomain)
	userHandler := rest.NewUserHandler(userUsecase, authUsecase)
	orderHandler := rest.NewOrderHandler(orderUsecase, authUsecase)

	api, server := httpserver.New(cfg.App.Address, cfg.App.BasePath, cfg.App.AllowedOrigins, cfg.Logger, []httpserver.Pinger{cfg.DB})
	authHandler.Register(api)
	userHandler.Register(api)
	orderHandler.Register(api)

	return server
}
