package cmd

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/DeepAung/artmission/backend/internal/pkg/database"
	"github.com/DeepAung/artmission/backend/internal/pkg/migrations"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate up|down|reset|create",
	Short: "Migrate the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("expected an argument: up, down, reset, or create")
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

		switch args[0] {
		case "up":
			return goose.UpContext(ctx, db.DB, ".")
		case "down":
			return goose.DownContext(ctx, db.DB, ".")
		case "reset":
			err := goose.ResetContext(ctx, db.DB, ".")
			if err != nil && strings.Contains(err.Error(), "failed to get status of migrations") {
				return nil
			}
			return err
		case "create":
			name := ""
			if len(args) >= 2 {
				name = args[1]
			}
			// goose.Create doesn't honor SetBaseFS, so point it at the real
			// migrations directory on disk instead of the embedded FS.
			return goose.Create(db.DB, "./internal/pkg/migrations", name, "sql")
		default:
			return errors.New("invalid migrate argument, expected: up, down, reset, or create")
		}
	},
}
