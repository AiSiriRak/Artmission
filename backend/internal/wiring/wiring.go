// Package wiring is the single composition root: the one place every
// adapter, usecase, and handler is constructed and connected, in
// dependency order. cmd/cmd_serve.go (the real server) and
// tests/internal/apptest (the BDD suite's in-process server) both call
// Wire so the object graph they exercise can never diverge — add a
// module here once and every caller gets it.
package wiring

import (
	"log/slog"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres"
	"github.com/AiSiriRak/Artmission/backend/internal/adapters/slip2go"
	"github.com/AiSiriRak/Artmission/backend/internal/adapters/token"
	"github.com/AiSiriRak/Artmission/backend/internal/handler/rest"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/artist"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/artwork"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/notification"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/review"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/objectstorage"
	"github.com/AiSiriRak/Artmission/backend/internal/worker"
	"github.com/uptrace/bun"
)

// Config is the wiring graph's inputs. DB, Logger, and ObjectStorage are
// live objects rather than settings, since callers build those differently.
// App and Auth are the real config package's structs, reused as-is so a
// config field renamed there can't silently drift out of sync here.
type Config struct {
	DB              *bun.DB
	Logger          *slog.Logger
	App             config.App
	Auth            config.Auth
	Slip2Go         config.Slip2Go
	ObjectStorage   *objectstorage.Client
	PaymentProvider payment.QRProvider
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
	reviewRepo := postgres.NewReviewRepository(cfg.DB)
	accountDeletionRepo := postgres.NewAccountDeletionRepository(cfg.DB)
	tokenIssuer := token.NewJWTIssuer(cfg.Auth.JWTSecret)
	tx := baserepo.NewTransactioner(cfg.DB)
	workflow := newOrderPaymentUsecases(cfg, tx)

	artistUsecase := artist.NewProfileUsecase(artistRepo, cfg.ObjectStorage.Public)
	artworkUsecase := artwork.NewUsecase(artworkRepo, tx, cfg.ObjectStorage.Public)
	userUsecase := user.NewUserUsecase(userRepo, bankRepo, artistUsecase, accountDeletionRepo, tx, cfg.ObjectStorage.Public)
	authUsecase := auth.NewAuthUsecase(userUsecase, sessionRepo, tokenIssuer, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL)
	reviewUsecase := review.NewReviewUsecase(reviewRepo, review.WithProfileImageURLResolver(cfg.ObjectStorage.Public))

	authHandler := rest.NewAuthHandler(userUsecase, authUsecase, cfg.App.BasePath, cfg.App.IsProduction, cfg.Auth.RefreshCookieDomain)
	userHandler := rest.NewUserHandler(userUsecase, authUsecase, cfg.App.BasePath, cfg.App.IsProduction, cfg.Auth.RefreshCookieDomain)
	orderHandler := rest.NewOrderHandler(workflow.order, artworkUsecase, authUsecase)
	reviewHandler := rest.NewReviewHandler(reviewUsecase, authUsecase)
	artistHandler := rest.NewArtistHandler(artistUsecase, authUsecase)
	artworkHandler := rest.NewArtworkHandler(artworkUsecase, authUsecase)
	paymentHandler := rest.NewPaymentHandler(workflow.payment, authUsecase)

	api, server := httpserver.New(cfg.App.Address, cfg.App.BasePath, cfg.App.AllowedOrigins, cfg.Logger, []httpserver.Pinger{cfg.DB, cfg.ObjectStorage})
	authHandler.Register(api)
	rest.NewNotificationHandler(workflow.notifications, authUsecase).Register(api)
	userHandler.Register(api)
	orderHandler.Register(api)
	reviewHandler.Register(api)
	artistHandler.Register(api)
	artworkHandler.Register(api)
	paymentHandler.Register(api)
	wallet := payment.NewWalletUsecase(postgres.NewWalletRepository(cfg.DB))
	rest.NewWalletHandler(wallet, authUsecase).Register(api)

	return server
}

const orderExpiryInterval = time.Minute

// WireWorkers builds the background workers. It is separate from Wire so
// the BDD suite's in-process server doesn't start timers it can't control;
// cmd_serve.go is the only caller.
func WireWorkers(cfg Config) []worker.Worker {
	workflow := newOrderPaymentUsecases(cfg, baserepo.NewTransactioner(cfg.DB))

	return []worker.Worker{
		worker.NewOrderExpiryWorker(workflow.order, orderExpiryInterval, cfg.Logger),
		worker.NewPaymentQRWorker(workflow.payment, orderExpiryInterval, cfg.Logger),
	}
}

type orderPaymentUsecases struct {
	order         order.OrderUsecase
	payment       payment.Usecase
	notifications notification.NotificationUsecase
}

func newOrderPaymentUsecases(cfg Config, tx baserepo.Transactioner) orderPaymentUsecases {
	notifications := notification.NewNotificationUsecase(postgres.NewNotificationRepository(cfg.DB))
	reader := postgres.NewOrderNotificationReader(cfg.DB)
	return orderPaymentUsecases{
		order: order.NewOrderUsecase(postgres.NewOrderRepository(cfg.DB), cfg.ObjectStorage.Private,
			order.WithNotifications(notification.NewOrderNotifier(reader, notifications), tx),
			order.WithProfileImageURLResolver(cfg.ObjectStorage.Public)),
		payment: payment.NewPaymentUsecase(postgres.NewPaymentRepository(cfg.DB), paymentProvider(cfg),
			payment.WithNotifications(notification.NewPaymentNotifier(reader, notifications), tx)),
		notifications: notifications,
	}
}

func paymentProvider(cfg Config) payment.QRProvider {
	if cfg.PaymentProvider != nil {
		return cfg.PaymentProvider
	}
	return newSlip2GoProvider(cfg.Slip2Go)
}

func newSlip2GoProvider(cfg config.Slip2Go) payment.QRProvider {
	return slip2go.NewClient(slip2go.Config{
		APIBaseURL:    cfg.APIBaseURL,
		SecretKey:     cfg.SecretKey,
		PromptPayCode: cfg.PromptPayCode,
		PromptPayType: cfg.PromptPayType,
		AccountName:   cfg.AccountName,
	})
}
