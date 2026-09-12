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
	"github.com/AiSiriRak/Artmission/backend/internal/modules/artwork"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/objectstorage"
	"github.com/uptrace/bun"
)

// Config is the wiring graph's inputs. DB, Logger, and ObjectStorage are
// live objects rather than settings, since callers build those differently.
// App and Auth are the real config package's structs, reused as-is so a
// config field renamed there can't silently drift out of sync here.
type Config struct {
	DB            *bun.DB
	Logger        *slog.Logger
	App           config.App
	Auth          config.Auth
	ObjectStorage *objectstorage.Client
}

// Wire builds the entire object graph — adapters, usecases, handlers,
// routes — and returns the ready-to-serve HTTP server. The caller owns
// the server's lifecycle.
func Wire(cfg Config) *httpserver.Server {
	userRepo := postgres.NewUserRepository(cfg.DB)
	bankRepo := postgres.NewBankAccountRepository(cfg.DB)
	artistRepo := postgres.NewArtistRepository(cfg.DB)
	artworkRepo := postgres.NewArtworkRepository(cfg.DB)
	sessionRepo := postgres.NewSessionRepository(cfg.DB)
	orderRepo := postgres.NewOrderRepository(cfg.DB)
	accountDeletionRepo := postgres.NewAccountDeletionRepository(cfg.DB)
	tokenIssuer := token.NewJWTIssuer(cfg.Auth.JWTSecret)
	tx := baserepo.NewTransactioner(cfg.DB)

	artistUsecase := artist.NewProfileUsecase(artistRepo, cfg.ObjectStorage)
	artworkUsecase := artwork.NewUsecase(artworkRepo, tx, cfg.ObjectStorage)
	userUsecase := user.NewUserUsecase(userRepo, bankRepo, artistUsecase, accountDeletionRepo, tx)
	authUsecase := auth.NewAuthUsecase(userUsecase, sessionRepo, tokenIssuer, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL)
	orderUsecase := order.NewOrderUsecase(orderRepo, cfg.ObjectStorage)

	authHandler := rest.NewAuthHandler(userUsecase, authUsecase, cfg.App.BasePath, cfg.App.IsProduction, cfg.Auth.RefreshCookieDomain)
	userHandler := rest.NewUserHandler(userUsecase, authUsecase, cfg.App.BasePath, cfg.App.IsProduction, cfg.Auth.RefreshCookieDomain)
	orderHandler := rest.NewOrderHandler(orderUsecase, authUsecase)
	artistHandler := rest.NewArtistHandler(artistUsecase, authUsecase)
	artworkHandler := rest.NewArtworkHandler(artworkUsecase, authUsecase)

	api, server := httpserver.New(cfg.App.Address, cfg.App.BasePath, cfg.App.AllowedOrigins, cfg.Logger, []httpserver.Pinger{cfg.DB, cfg.ObjectStorage})
	authHandler.Register(api)
	userHandler.Register(api)
	orderHandler.Register(api)
	artistHandler.Register(api)
	artworkHandler.Register(api)

	return server
}
