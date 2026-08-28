package cmd

import (
	"context"
	"database/sql"
	"os/signal"
	"syscall"
	"time"

	"github.com/DeepAung/artmission/backend/internal/adapters/postgres"
	"github.com/DeepAung/artmission/backend/internal/adapters/token"
	"github.com/DeepAung/artmission/backend/internal/handler/rest"
	"github.com/DeepAung/artmission/backend/internal/modules/auth"
	"github.com/DeepAung/artmission/backend/internal/modules/order"
	"github.com/DeepAung/artmission/backend/internal/modules/user"
	"github.com/DeepAung/artmission/backend/internal/pkg/database"
	"github.com/DeepAung/artmission/backend/internal/pkg/httpserver"
	"github.com/DeepAung/artmission/backend/internal/pkg/logger"
	"github.com/DeepAung/artmission/backend/internal/pkg/migrations"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the HTTP API",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := getConfigFromCmd(cmd)
		if err != nil {
			return err
		}
		log := logger.NewLogger(cfg.App().IsProduction)

		db, err := database.NewPostgresDB(cfg.Database())
		if err != nil {
			return err
		}
		defer db.Close()

		if err := migrateUp(db.DB); err != nil {
			return err
		}

		// --- wire adapters -> modules -> handlers ---
		userRepo := postgres.NewUserRepository(db)
		sessionRepo := postgres.NewSessionRepository(db)
		orderRepo := postgres.NewOrderRepository(db)
		tokenIssuer := token.NewJWTIssuer(cfg.Auth().JWTSecret)

		userUsecase := user.NewUserUsecase(userRepo)
		authUsecase := auth.NewAuthUsecase(userUsecase, sessionRepo, tokenIssuer, cfg.Auth().AccessTokenTTL, cfg.Auth().RefreshTokenTTL)
		orderUsecase := order.NewOrderUsecase(orderRepo)

		authHandler := rest.NewAuthHandler(userUsecase, authUsecase, cfg.App().BasePath, cfg.App().IsProduction, cfg.Auth().RefreshCookieDomain)
		orderHandler := rest.NewOrderHandler(orderUsecase, authUsecase)

		api, server := httpserver.New(cfg.App().Address, cfg.App().BasePath, cfg.App().AllowedOrigins, log, []httpserver.Pinger{db})
		rest.RegisterRoutes(api, authHandler, orderHandler)

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		return server.Start(ctx)
	},
}

func migrateUp(sqlDB *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	goose.SetBaseFS(migrations.Migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, sqlDB, ".")
}
