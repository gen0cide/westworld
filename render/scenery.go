package render

import (
	"github.com/gen0cide/westworld/assets"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// ModelCache lazily decodes named .ob3 models from the models.orsc archive.
type ModelCache struct {
	arc   *assets.Archive
	cache map[string]*assets.Model
}

func NewModelCache(arc *assets.Archive) *ModelCache {
	return &ModelCache{arc: arc, cache: map[string]*assets.Model{}}
}

// Get decodes (and caches) the named model, trying a few name suffixes the
// client uses (".ob3", "1"). Returns nil if not present.
func (mc *ModelCache) Get(name string) *assets.Model {
	if m, ok := mc.cache[name]; ok {
		return m
	}
	for _, cand := range []string{name, name + ".ob3", name + "1", name + "1.ob3"} {
		if mc.arc.Has(cand) {
			b, err := mc.arc.Get(cand)
			if err == nil {
				m := assets.DecodeModel(b, 0)
				mc.cache[name] = m
				return m
			}
		}
	}
	mc.cache[name] = nil
	return nil
}

// PlaceScenery builds a placed GameModel for one SceneryLoc relative to the
// terrain mesh origin (baseX, baseY). It copies the named model, rotates by
// dir*32 (256-step yaw space), and translates to the tile centre at the local
// terrain height (mudclient.java:5285 placement). worldHeight returns the
// terrain elevation*3 used as the model's vertical anchor.
func PlaceScenery(mc *ModelCache, f *facts.Facts, land *pathfind.Landscape,
	baseX, baseY, plane int, loc facts.SceneryLoc) *GameModel {

	def := f.SceneryDef(loc.DefID)
	if def == nil || def.Model == "" {
		return nil
	}
	am := mc.Get(def.Model)
	if am == nil {
		return nil
	}
	g := FromAsset(am)
	g.SetLight(48, 48, -50, -10, -50)

	// local tile coords inside the window
	lx := loc.X - baseX
	ly := loc.Y - baseY
	// centre of the footprint, in world units (tile*128, +64 to centre)
	cx := int32(lx)*128 + 64
	cz := int32(ly)*128 + 64
	t := land.Tile(loc.X, loc.Y, plane)
	cy := -int32(t.GroundElevation) * 3

	g.Rotate(int32(loc.Direction)*32&0xff, 0, 0)
	g.Translate(cx, cy, cz)
	return g
}
