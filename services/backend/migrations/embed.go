package migrations

import "embed"

// Files contains embedded SQL migrations for backend bootstrap.
//
//go:embed *.sql
var Files embed.FS
