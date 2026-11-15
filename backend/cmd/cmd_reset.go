package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/database"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/migrations"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/objectstorage"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset --yes",
	Short: "Wipe ALL data: empty both S3 buckets, truncate the database, and re-run every migration",
	Long: "Destroys every row and object in the configured environment, then rebuilds the schema.\n" +
		"Buckets are emptied before the database, so a failed run never leaves rows pointing at deleted objects.\n" +
		"Re-running after a failure is safe. Requires --yes.\n",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if yes, err := cmd.Flags().GetBool("yes"); err != nil {
			return err
		} else if !yes {
			return errors.New("refusing to wipe all data without --yes")
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

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		storage, err := objectstorage.NewS3Client(ctx, cfg.S3())
		if err != nil {
			return err
		}
		if err := storage.PingContext(ctx); err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		s3cfg := cfg.S3()
		fmt.Fprintf(out, "wiping buckets %q and %q\n", s3cfg.PublicBucketName, s3cfg.PrivateBucketName)
		n, err := storage.Public.DeleteAll(ctx)
		fmt.Fprintf(out, "public bucket: deleted %d objects\n", n)
		if err != nil {
			return err
		}
		n, err = storage.Private.DeleteAll(ctx)
		fmt.Fprintf(out, "private bucket: deleted %d objects\n", n)
		if err != nil {
			return err
		}

		// goose's down scripts fail while rows exist that reference the
		// migration-seeded catalog (artworks -> categories/styles are
		// ON DELETE RESTRICT), so empty every user-created row first. All
		// of it hangs off users by foreign key; the seeded catalog rows
		// aren't truncated and are removed and re-added by the migrations.
		fmt.Fprintln(out, "truncating data")
		if _, err := db.ExecContext(ctx, "TRUNCATE TABLE users CASCADE"); err != nil {
			return fmt.Errorf("truncate data: %w", err)
		}

		goose.SetBaseFS(migrations.Migrations)
		if err := goose.SetDialect("postgres"); err != nil {
			return err
		}
		fmt.Fprintln(out, "resetting migrations")
		if err := goose.ResetContext(ctx, db.DB, "."); err != nil {
			return err
		}
		if err := goose.UpContext(ctx, db.DB, "."); err != nil {
			return err
		}
		fmt.Fprintln(out, "done")
		return nil
	},
}

func init() {
	resetCmd.Flags().Bool("yes", false, "confirm that all data in the configured environment may be destroyed")
}
