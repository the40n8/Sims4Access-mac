// Package webui embeds the phone app. companion/web builds into dist/
// (mac/build.sh copies it here before compiling the bridge).
package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var files embed.FS

// FS is the built web app rooted at its index.html.
func FS() fs.FS {
	sub, err := fs.Sub(files, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
