package orsc

import (
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// stories.go ports World.loadSections (World.java:1476-1495) — the multi-FLOOR
// driver. When the host is on the ground plane it builds three stacked stories
// (planes 0,1,2); on any upper plane it builds only that plane (you only see your
// own floor when upstairs):
//
//	generateLandscapeModel(worldX, 122, true,  plane, worldZ);   // host plane
//	if (plane == 0) {
//	    generateLandscapeModel(worldX, 112, false, 1, worldZ);   // 2nd story
//	    generateLandscapeModel(worldX, 69,  false, 2, worldZ);   // 3rd story
//	}
//
// THE PER-FLOOR VERTICAL LIFT lives in the elevation cache, NOT in any explicit
// offset. World.java seeds tileElevationCache = getTileElevation ONLY for the
// host plane (it sits inside the plane-0-only `if (showWallOnMinimap)` block,
// World.java:820). For the two upper stories the cache is NEVER re-seeded:
// `resetFaceVertHead` (World.java:822) clears the per-plane MODEL geometry but
// leaves the cache alone. So the ground story's roof pass raises the cache by the
// ground-floor wall heights (applyWallToElevationCache + the roof rise), then the
// 2nd story's wall pass reads those raised heights as its FEET (insertWallIntoModel
// reads tileElevationCache) and stacks on top, and the 3rd story stacks again.
// Upper-floor tiles carry groundElevation==0, so without this accumulation every
// story would collapse onto the ground — which is exactly the bug this fixes.
//
// Faithfully reproducing that means: one shared cache, seeded once for the host
// plane, threaded un-reset through every story's wallPass (reads it) then roofPass
// (mutates it), in plane order. Geometry stays per-story (separate models, like
// modelWallGrid[plane] / modelRoofGrid[plane]).

// storyPlanes returns the planes loadSections builds for a host on hostPlane.
func storyPlanes(hostPlane int) []int {
	if hostPlane == 0 {
		return []int{0, 1, 2}
	}
	return []int{hostPlane}
}

// BuildStories builds the stacked wall + roof models for every story visible from
// hostPlane, threading the single accumulating elevation cache through them so
// upper floors lift onto the lower floors' wall-tops (World.loadSections). The
// returned models (non-empty only) are added to the scene by the caller; terrain
// is built separately for the host plane only (BuildTerrain). Returns nil if facts
// is unavailable.
func BuildStories(land *pathfind.Landscape, f *facts.Facts, baseX, baseY, hostPlane int) []*Model {
	if f == nil {
		return nil
	}
	var ec elevCache // shared, never reset between stories — the per-floor lift
	var models []*Model

	// Authentic 0x80 under-roof cull (Mudclient.java:8527 / clean client.java:6952):
	// when the host stands on the GROUND floor INSIDE a roofed building (its tile's
	// ground-overlay type == 2), the client hides the host-plane roof AND the upper
	// stories' walls+roofs so the interior is lit/walkable and the host is not
	// occluded by his own roof — only the ground-floor walls (+doors) remain. Both
	// render callers centre the host at window-local windowCentreTile, so the host
	// world tile is baseX/baseY + windowCentreTile.
	underRoof := hostUnderRoof(land, baseX, baseY, baseX+windowCentreTile, baseY+windowCentreTile, hostPlane)

	for i, sp := range storyPlanes(hostPlane) {
		// Under roof: only the ground floor's walls survive — skip every upper
		// story (the i>0 iterations) entirely (mirrors removeModel of
		// wallModels[1]/[2] + roofModels[1]/[2] never re-added when the 0x80 bit
		// is set).
		if underRoof && i > 0 {
			break
		}
		b := &terrainBuilder{land: land, baseX: baseX, baseY: baseY, plane: sp}
		if i == 0 {
			// World.java:820 — seed the cache from the host plane's terrain. Only
			// the first (showWallOnMinimap) story does this; the rest inherit it.
			b.seedElevationCache(&ec)
		}

		// Wall pass reads the cache for foot heights (World.java:822-896).
		walls := NewModel(wallModelCapacity, wallModelCapacity)
		b.wallPass(f, walls, &ec)
		if walls.faceHead > 0 {
			walls.setDiffuseLightAndColor(-50, -10, -50, 60, 24, false, 122) // World.java:891
			models = append(models, walls)
		}

		// Door pass: the openable boundaries (doorframes/doors/gates) the static
		// wall pass skips, drawn as OpenRSC's door-object panels (createWallObjectModel,
		// mudclient.java:2578). Reads the same clean cache as the walls (before the
		// roof pass mutates it) so upper-story doors stack. Separate model + the
		// door diffuse light (-95).
		doors := NewModel(wallModelCapacity, wallModelCapacity)
		b.doorPass(f, doors, &ec)
		if doors.faceHead > 0 {
			doors.setDiffuseLightAndColor(-50, -10, -50, 60, 24, false, -95) // createWallObjectModel
			models = append(models, doors)
		}

		// Roof pass MUTATES the cache (raises it by this story's wall heights +
		// roof rise) and emits roof faces (World.java:897-1145). This raise is what
		// the NEXT story's wall pass stacks on. We ALWAYS run it (it mutates the
		// shared elevation cache the next story's wall feet read), but when the host
		// is under a roof we DROP the emitted faces from the host-plane roof model —
		// the authentic client removeModel's roofModels[yj] and only re-adds it when
		// the 0x80 bit is clear (Mudclient.java:8527). Suppressing only the faces
		// keeps the cache mutation intact for any (unbuilt-here) story that would
		// stack on it, and matches OURS suppressing the plane-0 roof append.
		roofs := NewModel(wallModelCapacity, wallModelCapacity)
		b.roofPass(f, roofs, &ec)
		if roofs.faceHead > 0 && !underRoof {
			roofs.setDiffuseLightAndColor(-50, -10, -50, 50, 50, true, -98) // World.java:1145
			models = append(models, roofs)
		}
	}
	return models
}
