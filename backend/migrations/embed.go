// Package migrations embeds the SQL migration files so golang-migrate can
// run them from the compiled binary.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
