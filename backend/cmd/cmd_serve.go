package cmd

import (
	"context"
	"database/sql"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/logger"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/migrations"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/objectstorage"
	"github.com/AiSiriRak/Artmission/backend/internal/wiring"
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

		objectStorage, err := objectstorage.NewS3Client(context.Background(), cfg.S3())
		if err != nil {
			return err
		}

		wireConfig := wiring.Config{DB: db, Logger: log, App: cfg.App(), Auth: cfg.Auth(), ObjectStorage: objectStorage}
		server := wiring.Wire(wireConfig)

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		workersCtx, stopWorkers := context.WithCancel(ctx)
		var workers sync.WaitGroup
		for _, backgroundWorker := range wiring.WireWorkers(wireConfig) {
			workerToRun := backgroundWorker
			workers.Add(1)
			go func() {
				defer workers.Done()
				workerToRun.Run(workersCtx)
			}()
		}
		defer func() {
			stopWorkers()
			workers.Wait()
		}()

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
