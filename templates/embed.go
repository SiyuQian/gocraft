package templates

import "embed"

//go:embed all:base all:http all:async
var FS embed.FS
