package main

// sprites.go — GET /sprite: authentic RSC sprite PNGs for the browser UI.
//
// Item icons are decoded by render.ItemSpritePNG (which lazily opens
// media58/config85/entity24 from the WESTWORLD_*_JAG env vars or the render
// package's search paths). Icons are static, so responses are immutable-cached.

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/render"
)

// registerSpriteRoutes adds GET /sprite to the mux. Kept in its own file so the
// serveClient mux setup only gains a single call. f supplies the item-def
// AppearanceID table that render.ItemSpritePNG resolves into the sprite archive.
func registerSpriteRoutes(mux *http.ServeMux, log *slog.Logger, f *facts.Facts) {
	mux.HandleFunc("/sprite", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		kind := q.Get("kind")
		if kind == "" {
			kind = "item"
		}
		if kind != "item" {
			http.Error(w, "unsupported kind (only 'item' for now)", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(q.Get("id"))
		if err != nil || id < 0 {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}
		pngBytes, ok := render.ItemSpritePNG(f, id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("ETag", fmt.Sprintf(`"item-%d"`, id))
		_, _ = w.Write(pngBytes)
	})
	log.Debug("registered GET /sprite (item icons)")
}
