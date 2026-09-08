// Package migrations embeds the SQL migration files, applied via goose
// (see cmd/cmd_migrate.go and cmd/cmd_serve.go).
package migrations

import "embed"

//go:embed *.sql
var Migrations embed.FS
