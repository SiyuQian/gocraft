package templates

import "embed"

//go:embed all:base all:http all:async all:obs
var FS embed.FS
