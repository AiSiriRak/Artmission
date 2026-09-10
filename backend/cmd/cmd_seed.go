package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/migrations"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/objectstorage"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/seed"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed up|down|reseed",
	Short: "Seed the database with local/dev fixture data",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("expected an argument: up, down, or reseed")
		}

		cfg, err := getConfigFromCmd(cmd)
		if err != nil {
			return err
		}
		db, err := database.NewPostgresDB(cfg.Database())
		if err != nil {
			return err
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		goose.SetBaseFS(migrations.Migrations)
		if err := goose.SetDialect("postgres"); err != nil {
			return err
		}

		provider, err := goose.NewProvider(goose.DialectPostgres, db.DB, migrations.Migrations)
		if err != nil {
			return err
		}
		currentVersion, targetVersion, err := provider.GetVersions(ctx)
		if err != nil {
			return err
		}
		if currentVersion != targetVersion {
			return fmt.Errorf("current migration version (%d) don't match the target version (%d)", currentVersion, targetVersion)
		}

		storage, err := objectstorage.NewS3Client(ctx, cfg.S3())
		if err != nil {
			return err
		}

		switch args[0] {
		case "up":
			return seed.Up(ctx, db, storage)
		case "down":
			return seed.Down(ctx, db, storage)
		case "reseed":
			return seed.Reseed(ctx, db, storage)
		default:
			return errors.New("invalid seed argument, expected: up, down, or reseed")
		}
	},
}
