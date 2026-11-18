// Package migrations embeds the SQL migration files into the binary.
package migrations

import "embed"

// FS holds every NNNN_name.sql file in this directory, applied in filename order.
//
//go:embed *.sql
var FS embed.FS
