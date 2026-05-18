package generate

import (
	"io/fs"

	"github.com/siyuqian/gocraft/templates"
)

// EmbeddedFS returns the embedded template FS. Exposed for the `new` command
// and tests that exercise real templates.
func EmbeddedFS() fs.FS { return templates.FS }
