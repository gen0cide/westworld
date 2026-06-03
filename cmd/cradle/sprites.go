package main

// sprites.go — GET /sprite: authentic RSC sprite PNGs for the browser UI.
//
// Item icons are decoded by render.ItemSpritePNG and UI/interface ("media")
// sprites by render.MediaSpritePNG (both lazily open Authentic_Sprites.orsc via
// the render package's WESTWORLD_SPRITES_ORSC search paths). Icons are static,
// so responses are immutable-cached. kind=item needs the item-def table (f);
// kind=media (tab icons etc) does not.

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
		if kind != "item" && kind != "media" {
			http.Error(w, "unsupported kind (only 'item' or 'media')", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(q.Get("id"))
		if err != nil || id < 0 {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}
		// media = UI interface sprites (tab icons etc), keyed off spriteMedia and
		// needing no item-def table; item = inventory icons via the AppearanceID table.
		var (
			pngBytes []byte
			ok       bool
			etag     string
		)
		if kind == "media" {
			pngBytes, ok = render.MediaSpritePNG(id)
			etag = fmt.Sprintf(`"media-%d"`, id)
		} else {
			pngBytes, ok = render.ItemSpritePNG(f, id)
			etag = fmt.Sprintf(`"item-%d"`, id)
		}
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("ETag", etag)
		_, _ = w.Write(pngBytes)
	})
	log.Debug("registered GET /sprite (item + media icons)")
}
