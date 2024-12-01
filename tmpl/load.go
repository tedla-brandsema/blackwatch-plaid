package tmpl

import (
	"embed"
	"io/fs"
)

var (
	//go:embed *
	tmplFS embed.FS
)

func FileSystem() fs.FS {
	return tmplFS
}
