// Package migrations embeds the SQL migration files for use with goose.
package migrations

import "embed"

// FS holds all goose migration SQL files.
//
//go:embed *.sql
var FS embed.FS
