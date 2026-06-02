// Package web embeds the built Vite/React frontend (web/dist) so the cradle
// binary serves the remote-client UI with no external asset dependency.
//
// The dist/ directory is produced by `npm run build` in this directory. It must
// exist for the package to compile (go:embed is a build-time directive); a
// committed build keeps `go build ./...` green on a fresh checkout.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Dist returns the production build rooted so index.html sits at the FS root.
func Dist() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("web: embedded dist missing (run `npm run build` in web/): " + err.Error())
	}
	return sub
}
