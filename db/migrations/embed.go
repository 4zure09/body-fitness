package migrations

import "embed"

// Files contains every versioned SQL migration used by the deployment runner.
//
//go:embed *.sql
var Files embed.FS
