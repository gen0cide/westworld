package orsc

import (
	"github.com/gen0cide/westworld/assets"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// scenery_place.go wires SceneryLocs (trees, railings, wells, signs, …) into the
// three/ scene — the World.addLoginScreenModels path (World.java:162-216), but
// driven by the OpenRSC server's SceneryLocs.json (via facts) the same way the
// existing render/ package does, rather than the landscape's 48000+ inline
// markers (which OpenRSC superseded with SceneryLocs).

// modelNameSuffixes mirror render/scenery.go's ModelCache.Get fallbacks: the
// def's model name may be stored bare, with a ".ob3" extension, or with a "1"
// frame suffix in models.orsc.
var modelNameSuffixes = []string{"", ".ob3", "1", "1.ob3"}

// decodeOb3 finds + decodes a named model from the models.orsc archive, trying
// the known name suffixes. Returns nil if absent/undecodable.
func decodeOb3(arc *assets.Archive, name string) *assets.Model {
	for _, suf := range modelNameSuffixes {
		cand := name + suf
		if arc.Has(cand) {
			if b, err := arc.Get(cand); err == nil {
				return assets.DecodeModel(b, 0)
			}
		}
	}
	return nil
}

// placeSceneryModels places every SceneryLoc inside the
// [baseX,baseX+96)×[baseY,baseY+96) window into the scene. The decoded .ob3
// geometry is cached per model name; each placement builds a FRESH *Model (its
// own transform via PlaceScenery's Orient+Translate). Missing models/defs are
// skipped; never panics.
func placeSceneryModels(scene *Scene, arc *assets.Archive, land *pathfind.Landscape, f *facts.Facts, baseX, baseY, plane int) {
	if f == nil || arc == nil {
		return
	}
	cache := map[string]*assets.Model{}
	get := func(name string) *assets.Model {
		if m, ok := cache[name]; ok {
			return m
		}
		m := decodeOb3(arc, name)
		cache[name] = m
		return m
	}
	for _, loc := range f.SceneryLocs {
		if loc.X < baseX || loc.X >= baseX+worldWindowTiles ||
			loc.Y < baseY || loc.Y >= baseY+worldWindowTiles {
			continue
		}
		def := f.SceneryDef(loc.DefID)
		if def == nil || def.Model == "" {
			continue
		}
		am := get(def.Model)
		if am == nil || am.NumVertices == 0 {
			continue
		}
		m := FromAssets(am.NumVertices, am.NumFaces, am.VertexX, am.VertexY, am.VertexZ,
			am.FaceVertices, am.FaceNumVertices, am.FaceFillFront, am.FaceFillBack, am.FaceIntensity)
		PlaceScenery(m, land, baseX, baseY, plane, loc.X-baseX, loc.Y-baseY, def.Width, def.Height, loc.Direction, loc.DefID)
		scene.AddModel(m)
	}
}
