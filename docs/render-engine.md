# Render Engine

**Status: IMPLEMENTED & actively iterated** (since 2026-05-29). A pure-Go
(WASM-capable) software renderer that reproduces a host's *perceived* RuneScape
Classic 3D scene. The engine is **`render/orsc`** — a from-scratch, line-faithful
Go port of **OpenRSC's `orsc/graphics/three`** pipeline (World/Scene/RSModel/
Shader/Polygon/Scanline), reading OpenRSC's own `.orsc` assets. Driven by
`cmd/cradle` via the `-render-view` (one PNG) and `-spectate` (live browser) flags.

> **History.** The first implementation (2026-05-29 → 05-31) was a classic
> RSCPlus-matched software rasteriser living directly in `render/`
> (`Scene`/`GameModel`/`Surface`). It was **replaced** by the faithful OpenRSC
> three/ port (`render/orsc`) and then **removed** from the tree — see git history
> for that engine. `render/` is now just the shared **sprite-compositing** layer
> (entity/item billboards) + the `View`/`Entity` value types that both `render/orsc`
> and the cradle pass around. This doc also supersedes `docs/render-port-plan.md`
> (the original pre-implementation plan, kept for historical context).

---

## 1. Why this exists

westworld bots perceive the world as decoded state (tiles, NPCs, players, objects),
not pixels. To **debug perception** and to **observe what a host sees** — and
eventually to feed a vision channel — we reconstruct the host's first-person RSC
scene from the same world-state the bot acts on, and render it to an image. The bar
is **faithfulness to the OpenRSC client**: a human (Alex) compares the live viewport
to what OpenRSC renders and reports divergences; we root-cause each against the
OpenRSC `orsc/graphics/three` sources (the engine we ported) and fix it in the port
to match the Java exactly (never work around in content).

It is a **software** renderer (no GPU, no CGo, no native window) — a line-faithful
port of OpenRSC's fixed-point three/ pipeline (`Scene`/`RSModel`/`Shader`/`Surface`
in `render/orsc`). Output is a PNG; the live viewport is a browser polling that PNG
(the browser is the only display; there is no native window). The Go port stays
faithful to the Java: where the Java and Go differ (shift precedence, `&`
precedence, `>>>` logical shifts, int32 overflow-wrap, signed-byte reads), the port
reproduces the Java's exact arithmetic rather than the "cleaner" Go default.

## 2. Decoupled by design

The renderer is an **optional, read-only consumer** of world-state. Normal bot
execution (running a routine, the net loop, the DSL) never invokes `render/`; it is
only called from the `-render-view` / `-spectate` dispatch in `cmd/cradle/main.go`.
The render path reads the *same* `world.World` + `facts.Facts` the bot uses, plus
the landscape archive, and produces an image — it never mutates game state and is
not on the control-loop hot path. So a rendering bug can never break a bot, and the
bot can run headless with zero render cost.

## 3. The pipeline

```
Authentic_Landscape.orsc ── pathfind.OpenLandscape ──► Landscape (per-tile records:
                                                       elevation, ground texture,
                                                       overlay, walls, roof, diag)
facts.Facts (defs+locs) ─┐
world.World (live state) ─┤
                          ▼
   orsc.RenderView(land, facts, models, sprites, View) ─────────────────────────┐
     1. BuildTerrain (World.generateLandscapeModel)  → terrain mesh RSModel      │
     2. LoadTexturesFromArchive (Authentic_Sprites)  → texel buffers             │
     3. placeSceneryModels (models.orsc + ob3 decode) → object models            │
     4. BuildStories (loadSections)                   → stacked walls + roofs    │
        + door pass (createWallObjectModel)           → doorframes/doors         │
     5. addViewEntities → AddEntity                    → NPC/player/self sprites  │
     6. SetCamera(host tile, pitch 912, yaw rot*4, distance) + SelfOff glide     │
     7. Scene.Render (endScene): project + sort + scanline-fill all models       │
     8. Surface.toImage()                                                        │
                                                                    ────────────┘
```

`View` (in `render/render.go`) is the request: host tile `X,Y`, `Plane`, `Rotation`
(0–255, yaw = rotation*4), `Zoom`, output `W,H`, the `Entities` to draw, the host's
own appearance/heading, optional dynamic-state overrides (opened doors, depleted
scenery, dropped items), and the motion-interpolation offsets (§6). The sprite blit
+ depth sort happen inside `Scene.Render` (the m_T billboard faces), not a separate
post-pass.

### Package map

**`render/orsc`** — the renderer (the faithful three/ port):
| File | Java source | Responsibility |
|---|---|---|
| `world.go` | `World.generateLandscapeModel` | terrain mesh + overlays + bridge deck |
| `walls.go` | wall pass + `insertWallIntoModel` | building/fence walls (reads the elevation cache) |
| `roofs.go` | roof pass + `applyWallToElevationCache` | roofs + the per-story marker-stripped cache |
| `stories.go` | `World.loadSections` | multi-floor driver (accumulating elevation cache = the per-floor lift) |
| `scenery_place.go` | `World.addModels` | place `SceneryLocs` object models (ob3 decode) |
| `entity.go` | `Scene.drawSprite` / m_T faces | actor billboards (register + scaled blit) |
| `texture.go` / `textureload.go` | `loadTexturesAuthentic` | 3D texel buffers from `Authentic_Sprites.orsc` (`3225+i`) |
| `scene.go` | `Scene` (setCamera/endScene) | camera, frustum, project, painter sort, overlap |
| `model.go` | `RSModel` | vertices/faces, `rotate1024` projection, diffuse lighting |
| `raster.go` | `Scanline`/`Polygon`/`Shader` | gouraud + textured scanline fill, fog, blend |
| `surface.go` | `Surface` | framebuffer + PNG encode |
| `view.go` / `harness.go` | `mudclient` camera | live `RenderView` + `RenderBridge` entry points |
| `pick.go` | — | screen click → world tile (orsc-camera forward-project) |

**`render/`** — the shared sprite-compositing + view types (renderer-agnostic):
| File | Responsibility |
|---|---|
| `render.go` / `entity.go` | `View` / `Entity` / `DynamicSceneryItem` / `GroundItemMarker` value types |
| `entitysprite.go` / `animdefs.go` | NPC/player sprite composite from `Authentic_Sprites.orsc` (anim `number+frame`); layers/colours from `facts.NpcDef`; facing + **walk-cycle** |
| `composite_export.go` | exported `Composite{Player,NPC,PlayerAppearance}Sprite` + `NPCBillboardSize` wrappers `render/orsc` calls |
| `itemsprite.go` | item icons from `Authentic_Sprites.orsc` (`2150 + ItemDef.AppearanceID`) onto the 48×32 inventory canvas |
| `sprites.go` | the shared `Authentic_Sprites.orsc` accessor (`sprites()`) + sprite-id bases |

### Data sources
- **Landscape**: `~/Code/openrsc/server/conf/server/data/Authentic_Landscape.orsc`
  (also `Client_Base/Cache/video/`). Per-tile: ground elevation/texture, ground
  overlay (the decoration id), horizontal/vertical/diagonal wall ids, roof texture.
- **Sprites & textures**: OpenRSC's **`Authentic_Sprites.orsc`** (a plain ZIP of
  decimal-id sprite entries) is the SINGLE 2D source — 3D textures (`3225+i`),
  item icons (`2150+n`), and NPC/player animation frames (anim `number+frame`).
  Path list in `render/sprites.go spritesSearch`; `WESTWORLD_SPRITES_ORSC`
  overrides. This replaced the classic `textures17/media58/entity24/config85.jag`
  — the renderer pulls **nothing** from the eggsampler/rscdump tree.
- **Models**: `~/Code/openrsc/Client_Base/Cache/video/models.orsc` (`.ob3`).
- **Defs**: `facts.Facts` from OpenRSC's `server/conf/server/defs/` —
  `TileDef.xml` (overlays), `DoorDef.xml` (boundaries), `GameObjectDef.xml`
  (scenery), `ItemDefs.json` (incl. `appearanceID` icon index), `NpcDefs.json`
  (incl. `sprites1..12` + hair/top/bottom/skin colours + camera dims), and the
  sparse loc files.

## 4. Authentic-client invariants (the load-bearing facts)

These are RSC-client truths that the engine must honour; each was a bug until ported.

- **`12345678` = `Scene.TRANSPARENT`** (`RSModel.m_Vb`). A face whose visible-side
  texture is this sentinel is *built but never drawn* (`Scene.java:2666`). The
  "blank" boundary family (collision-only / open archways, e.g. the Lumbridge castle
  entrance) uses it — render nothing, not a grey quad.
- **Green `0x00ff00` is a texture transparency key.** The texture loader converts
  every pure-green texel to the magenta key on load (it encodes see-through
  openings — doorways). And the span filler skips only the *stored key* texel
  (`tex[ti]==0`), never a legitimately-dark shaded texel.
- **Window patterns are a texture SUBTYPE composite.** A window/arrowslit/
  stained-glass wall's texture is a base sprite ("wall") with a subtype sprite
  (arrowslit/stainedglass/window/…) drawn over it where the subtype's palette index
  ≠ 0. `textures.go` composites them.
- **TileDef.colour convention**: `>= 0` is a TEXTURE id; `< 0` is a `method305` flat
  colour — the same convention `scene.go` face fills use. `tileType` (the
  `<unknown>` tag in `TileDef.xml`): **2** = indoor floor, **3** = outdoor textured,
  **4** = water-class (forced to the water texture *and* gets a raised second-pass
  plank deck — §5), **5** = bridge sentinel.
- **Bridge overlay `250`** is remapped before terrain build → `9` (plank edge at
  sector seams) / `2` (water) — open swamp/ocean water, distinct from §5's road
  bridges.
- **NPC colours are RAW 24-bit dye values; PLAYER colours are palette INDICES**
  (`composite(..., rawColours bool)` — NPCs `true`, players `false`).
- **`.orsc` edge-axis swap**: the `HorizontalWall` byte runs along Y/Z, `VerticalWall`
  along X — inverse of the names (verified vs raw data).
- **Under-roof cull**: when the host stands on an overlay-type-2 tile, the covering
  roof + upper-storey walls are freed so an interior/passage isn't boxed in.

## 5. The two-pass water/bridge model

A `tileType-4` ("water-class") tile is rendered **twice** (World.java:704-723):
1. a **flattened water quad at y=0** (river level) — the type-4 colour-force = water
   texture; its vertices are flattened with the surrounding water.
2. a **second quad at the tile's REAL (un-flattened) elevation**, painted with the
   tile's deck texture (overlay-4 colour 3 = planks) — the raised **plank deck** of
   the Lumbridge→Varrock road bridge.

`terrain.go` keeps a parallel un-flattened `rawH[]` grid + a `deck[]` flag for this.
The wooden railings are separate scenery (`def 45 "railing"`).

## 6. The live spectator (`cmd/cradle/spectate.go`)

`cradle -spectate -spectate-addr localhost:8089` logs the host in and serves a
browser viewport that follows it. The browser is the display; it polls a PNG.

**HTTP endpoints**
| route | purpose |
|---|---|
| `/` | the HTML+JS page (continuous rotation on held `<`/`>`, `+`/`-` zoom, click-to-walk, `p`/`c` capture, `#flash`) |
| `/frame?rot=&zoom=&w=&h=&anim=` | one freshly-rendered PNG of the host's view |
| `/pos` | host tile `{x,y,plane}` (JSON) |
| `/walk?px=&py=&rot=&zoom=&w=&h=` | screen click → `PickTile` → serialized `host.Walk` |
| `/shot`, `/clip` | save captures for an agent to read (see below) |

**Motion smoothing** — actors only move on the ~600ms server walk tick, so a naive
render snaps tile-to-tile. A spectator-side `motionTracker` (keyed on the server's
stable actor index) records each actor's prev→cur tile + change time and emits a
**sub-tile glide offset** (`render.Entity.OffX/OffZ`, `View.SelfOff*`) interpolated
over `walkGlideDur` (600ms). bernard's glide shifts both the camera and his own
billboard so he stays centred while the world scrolls. On top of the glide, the
**walk-cycle** selects sprite frame `i2*3 + npcWalkModel[phase]` (`StepPhase`) so
legs actually move. Non-adjacent jumps (teleport/re-appear) snap without streaking.

**Captures** (for an AI agent to read) land in `/tmp/render_out/spectator/`:
`/shot` → `shot_HHMMSS.png` (+ `shot.png` latest alias); `/clip` → a 12-frame burst
tiled into `clip_sheet_HHMMSS.png` (+ per-frame `clip_00..11.png`). Filenames are
timestamped so successive captures don't overwrite each other.

## 7. Feature status

**Feature-complete as of 2026-05-31**, with bridge rendering being the last subsystem
hardened: under-deck water consistency, the raised-deck second pass, the plank overhang
skirt, entity/camera elevation on the deck, and railing placement were all fixed and
**verified live** at the 104,655 bridge (`cradle -spectate`) — the host stands on the
deck between two clean railings with the river flowing underneath. Every implemented
feature renders faithfully, verified across the static `rendertest` scenes (castle
interior/exterior, straight + diagonal doors, doorframes, windows, wooden + swamp
bridges, gatehouse, multi-storey roofs, scenery, NPC/player sprites) **and** live
`cradle -spectate` captures at Lumbridge (church + bridge + stained glass + water).
The remaining open item (lighting parity) is optional polish, not a missing feature.

| Area | Status |
|---|---|
| Terrain heightmap, ground colours, path/overlay colour-split | ✅ |
| Water (river flatten + texture), shoreline diagonal, swamp 250 | ✅ |
| Walls / fences (textured, directional light), open doorways | ✅ |
| Door / doorframe **panels** (closed-door/arch from openable defs) | ✅ |
| Multi-storey roofs (3-wide ridge symmetry, cross-tile gouraud) | ✅ |
| Under-roof cull (gatehouse/castle passages render lit + open) | ✅ |
| Interior/overlay floor **textures** (chapel planks, marble, pentagram) | ✅ |
| Windows: arrowslit / stained-glass / timber / desert (subtype composite) | ✅ |
| Scenery 3D models (trees, wells, signs, railings, …) | ✅ |
| NPC + player + self **sprites** (real appearance, facing, raw vs index colour) | ✅ |
| **Motion glide** + **walk-cycle** leg animation | ✅ |
| Ground-item icons; model-swap anim (fires/torches) | ✅ |
| **Bridge plank deck** (tileType-4 raised second pass) | ✅ — under-deck water seats at y=0 via the authentic `isTileType2` **class** colour-split (so the river stays uniform under the deck, no fabricated shoreline wedges); the deck quad uses raw (un-flattened) elevation. |
| Castle **door** geometry (straight + diagonal, open + closed) | ✅ — the once-reported "out-of-place door / magenta sliver" did **not** reproduce in extensive static or live captures (2026-05-31); the symptom matches the green/transparency-key bug since fixed (`d8595e3` / `e5a95d6` / `9cfd7f7`). Re-open only if it recurs: walk a host to it in `-spectate`, hit `p`, read the shot. |
| Bridge plank **skirt/overhang** onto road-approach tiles (World.java:724-809) | ✅ implemented (Pass B in `terrain.go`): every non-tileType-3 tile bordering a deck draws one coplanar plank quad in the neighbour deck's colour, so the deck meets the bank instead of ending over the river gap. Gated `tileType==3 → skip`, so it extends only onto dry approaches, never across the N/S railings. |
| **Entity/camera elevation on bridges** (stand ON the deck, not in the river) | ✅ — feet + camera anchor to raw `elevationAt` (raised deck height), not the water-flattened height grid. Verified live at the 104,655 bridge: host stands on the deck between the railings. |
| **Bridge railings** (def 45 `woodenrailing`, outer edges of a 3-wide deck) | ✅ — placement verified empirically (baked-vertex dump): dir-0 → outer-low-Z edge, dir-4 (roll 128) → outer-high-Z edge; both land on the outside of the 1st/3rd deck rows, matching v204. The earlier "rail on the inside" symptom was the under-deck water + entity-elevation bugs above (since fixed), not railing placement. |
| Lighting vs RSCPlus | 🔶 **polish (optional, not a feature gap)** — the `SetLight` formula is authentic (`256-ambience*4`, `(64-diffuse)*16+128`); per-model ambience/diffuse are deliberately tuned (walls intentionally higher-contrast). Final pixel-match to RSCPlus is a human-in-the-loop tune against reference screenshots. |

## 8. How to run + iterate

```bash
cd ~/Code/westworld

# Static, no-login renders (reads the landscape directly) — fast iteration:
go run ./cmd/rendertest/        # → /tmp/render_out/*.png (cal_/castle_/door_/bridge_/wbridge_/…)

# One live frame of a logged-in host:
go build -o /tmp/cradle-rv ./cmd/cradle
set -a; . ~/Code/westworld/.local.env; set +a   # provides WESTWORLD_PASSWORD (never echo it)
/tmp/cradle-rv -username bernard -server localhost:43594 -facts ~/Code/openrsc \
  -render-view -render-out /tmp/x.png -render-rotation -1   # -1 = 8-way sweep

# Live browser spectator:
nohup /tmp/cradle-rv -username bernard -server localhost:43594 -facts ~/Code/openrsc \
  -spectate -spectate-addr localhost:8089 >/tmp/cradle-rv.log 2>&1 & disown
# → open http://localhost:8089/   (note: ~30fps poll pegs ~1 core; lower the page tick interval if needed)
```

**Diagnosing a new visual divergence** (the proven loop): reproduce it
(`rendertest` static, or capture the live spectator with `p`/`c`), then read the
**authentic client sources** to find the invariant we're missing — don't guess.
For anything subtle, fan out a read-only diagnosis + adversarial verification before
implementing (several render fixes that "looked obvious" were wrong on first pass).

### Reference clients (ground truth)
- **OpenRSC** client `~/Code/openrsc/Client_Base/src/orsc/` — **the port source +
  faithfulness target**: `graphics/three/World.java` (terrain/walls/roofs/bridge),
  `Scene.java` (camera/project/sort/scanline + TRANSPARENT), `RSModel.java`,
  `Shader.java` (fill), `mudclient.java` (camera/door objects). Server defs in
  `~/Code/openrsc/server/conf/server/defs/`.
- **deob** `~/Code/rscdump.com-runescape-classic-dump/` —
  `LeadingBot/mudclient.java` (readable), `deob106/`, `eggsampler-rsc-204-*/` (the
  classic three/ lineage OpenRSC derives from; useful for cross-checking intent).
- **RSCPlus** `~/Code/rscplus` — a secondary visual sanity-check only (the original
  classic renderer targeted it; the current port matches OpenRSC).

### Useful coords (Lumbridge)
castle ~`(116-120, 648-659)`; Lum→Varrock road bridge deck ~`(106-107, 655-657)`;
swamp open water ~`(140-190, 720-760)`.

## 9. Security

The OpenRSC test password is in the `WESTWORLD_PASSWORD` env var
(`~/Code/westworld/.local.env`). **Never** print, log, commit, or embed it; grep
staged diffs for it before committing. Pushing/signing is the human's step.
