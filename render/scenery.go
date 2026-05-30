package render

import (
	"github.com/gen0cide/westworld/assets"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// animatedSceneryModel maps object ids that animate by whole-model swap
// (mudclient.updateObjectAnimation) to their FIRST animation-frame model
// basename. We render the static frame-1 model. OpenRSC defs already store
// these suffixed names, but this guards a defs source carrying the bare base.
var animatedSceneryModel = map[int]string{
	51:   "torcha1",      // Torch
	97:   "firea1",       // fire
	143:  "skulltorcha1", // skull torch
	274:  "fireplacea1",  // Fireplace
	1031: "lightning1",   // lightning
	1036: "firespell1",   // flames / firespell
	1142: "clawspell1",   // clawspell
	1147: "spellcharge1", // Spellcharge
}

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
	modelName := def.Model
	if anim, ok := animatedSceneryModel[loc.DefID]; ok {
		modelName = anim // force the frame-1 animation model
	}
	am := mc.Get(modelName)
	if am == nil {
		return nil
	}
	g := FromAsset(am)
	g.SetLight(48, 48, -50, -10, -50)

	// local tile coords inside the window
	lx := loc.X - baseX
	ly := loc.Y - baseY
	// Footprint-aware centre (mudclient.java:5286-5294): a WxH object centres
	// at ((2*tile + size)*128)/2, not the fixed half-tile (+64) of a 1x1 — so a
	// 2x2 well/fountain no longer sits half a tile off. width/height swap for
	// any direction other than 0/4 (the object rotated onto its side). The
	// Width/Height==0 -> 1 guard keeps 1x1 objects at the old tile*128+64.
	w, hgt := int32(1), int32(1)
	if def.Width > 0 {
		w = int32(def.Width)
	}
	if def.Height > 0 {
		hgt = int32(def.Height)
	}
	if loc.Direction != 0 && loc.Direction != 4 {
		w, hgt = hgt, w
	}
	cx := (int32(lx)*2 + w) * 128 / 2
	cz := (int32(ly)*2 + hgt) * 128 / 2
	t := land.Tile(loc.X, loc.Y, plane)
	cy := -int32(t.GroundElevation) * 3

	// dir -> ROLL (rotation about the vertical Y axis = the object's heading),
	// NOT yaw. applyRotation maps the roll arg to the Z/X-mixing block (the
	// vertical-axis spin), matching the authentic m.addRotation(0, dir*32, 0).
	// Using yaw here tipped scenery onto its side (the well became a buried
	// grey blob).
	g.Rotate(0, 0, int32(loc.Direction)*32&0xff)
	g.Translate(cx, cy, cz)
	return g
}
