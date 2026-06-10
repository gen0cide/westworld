package cradle

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/index.html
var indexHTML []byte

// v2FS embeds the v2 SPA tree (web/v2/*). The all: prefix keeps every asset,
// including dot/underscore-prefixed files, so the UI never needs a build step.
//
//go:embed all:web/v2
var v2FS embed.FS

// handleIndex serves the single-page web UI. It consumes the same JSON API as
// cradle-ctl, so the UI and the CLI never drift.
func (a *API) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(indexHTML)
}

// v2Handler serves the embedded v2 SPA under /v2/: /v2/ resolves to index.html
// and /v2/app.js etc resolve through the same embedded tree. v1 at / is
// untouched.
func v2Handler() http.Handler {
	sub, err := fs.Sub(v2FS, "web/v2")
	if err != nil {
		panic(err) // go:embed guarantees web/v2 exists at compile time
	}
	return http.StripPrefix("/v2/", http.FileServerFS(sub))
}
