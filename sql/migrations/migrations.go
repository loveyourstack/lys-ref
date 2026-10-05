package migrations

import "embed"

// SQLAssets is an embedded filesystem containing SQL migration files.
//
//go:embed *
var SQLAssets embed.FS
