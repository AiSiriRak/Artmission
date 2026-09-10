// Package cmd wires the process's CLI entrypoints:
//
//   - serve runs the HTTP API
//   - migrate drives goose
//   - seed populates local/dev fixture data.
package cmd

import (
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "artmission-api",
	Short: "Artmission backend API",
}

func init() {
	RootCmd.PersistentFlags().String("env-file", "", "path to a .env file to load")
	RootCmd.AddCommand(serveCmd)
	RootCmd.AddCommand(migrateCmd)
	RootCmd.AddCommand(seedCmd)
}

func getConfigFromCmd(cmd *cobra.Command) (config.Config, error) {
	envFile, err := cmd.Flags().GetString("env-file")
	if err != nil {
		return nil, err
	}
	return config.InitConfig(envFile)
}
