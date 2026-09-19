// Package migrations embeds the SQL migration files so the Go services can
// apply them without needing the source tree available at runtime (e.g.
// inside a distroless container image).
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
