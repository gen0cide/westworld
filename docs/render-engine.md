# Render Engine

**Status: IMPLEMENTED & actively iterated** (since 2026-05-29). A pure-Go software
rasteriser that reproduces a host's *perceived* RuneScape Classic 3D scene, matched
pixel-faithfully to the **RSCPlus** reference client. Lives in `render/`; driven by
`cmd/cradle` via the `-render-view` (one PNG) and `-spectate` (live browser) flags.

> This supersedes `docs/render-port-plan.md` (the original pre-implementation plan,
> kept for historical context).

---

## 1. Why this exists

westworld bots perceive the world as decoded state (tiles, NPCs, players, objects),
not pixels. To **debug perception** and to **observe what a host sees** — and
eventually to feed a vision channel — we reconstruct the host's first-person RSC
scene from the same world-state the bot acts on, and render it to an image. The bar
is **pixel-faithfulness to RSCPlus**: a human (Alex) compares the live viewport to
RSCPlus screenshots and reports divergences; we root-cause each against the
authentic client sources and fix it in the engine (never work around in content).

It is a **software** rasteriser (no GPU, no CGo, no native window) — a direct port
of the classic RSC client's fixed-point 3D pipeline (`Scene`/`GameModel`/`Surface`).
Output is a PNG; the live viewport is a browser polling that PNG (the browser is the
only display; there is no native window).

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
   render.RenderView(land, facts, bundle, View) ──────────────────────────────┐
     1. TerrainHeights / buildTerrain   → terrain mesh GameModel               │
     2. BuildBoundaries                 → wall/door/fence quads                │
     3. BuildRoofs                      → multi-storey roof quads              │
     4. PlaceScenery                    → 3D object models (trees, wells, …)   │
     5. (assemble) Scene.Add(each model)                                       │
     6. camera = SetCamera(host tile, pitch 912, yaw rotation*4, distance)     │
     7. Scene.RenderTo(Surface, cx, cy) → rasterise all models to a pixel buf  │
     8. DrawEntitySprites               → composite NPC/player/self billboards │
                                          + ground-item icons over the scene   │
     9. Surface.PNG()                                                          │
                                                                    ────────────┘
```

`View` (in `render/render.go`) is the request: host tile `X,Y`, `Plane`, `Rotation`
(0–255, yaw = rotation*4), `Zoom`, output `W,H`, the `Entities` to draw, the host's
own appearance/heading, optional dynamic-state overrides (opened doors, depleted
scenery, dropped items), and the motion-interpolation offsets (§6).

### Package map (`render/`)
| File | Responsibility |
|---|---|
| `render.go` | `RenderView` orchestration, the camera, `DrawEntitySprites`, `View`/`Entity` types |
| `terrain.go` | heightmap mesh; ground-overlay classification; water flatten; **bridge deck** (§5) |
| `palette.go` | the 25-entry **`tileDefs`** table (overlay → colour+tileType) + ground-colour palette + `method305` |
| `boundary.go` | walls / doors / fences from tile edge records; directional lighting; under-roof cull |
| `roof.go` | multi-storey roofs (apex raise, ridge split, cross-tile gouraud via vertex dedup) |
| `textures.go` | decode `textures17.jag`; the **subtype composite** (windows) + green-key transparency |
| `entity.go` / `entitysprite.go` | NPC/player sprite composite from `entity24.jag`/`config85.jag`; facing + **walk-cycle** |
| `scene.go` / `rasterize.go` | the `Scene`/`GameModel` rasteriser: projection, painter sort, textured/flat spans, fog |
| `pick.go` | screen click → world tile (forward-project window tiles, nearest to click) |
| `scenery.go`, `model*.go` | 3D object model decode + placement; model cache |

### Data sources
- **Landscape**: `~/Code/openrsc/server/conf/server/data/Authentic_Landscape.orsc`
  (also `Client_Base/Cache/video/`). Per-tile: ground elevation/texture, ground
  overlay (the decoration id), horizontal/vertical/diagonal wall ids, roof texture.
- **Textures**: the classic `textures17.jag` (path list in
  `render/textures.go textureJagSearch`; `WESTWORLD_TEXTURES_JAG` overrides).
- **Defs**: `facts.Facts` from OpenRSC's `server/conf/server/defs/` —
  `TileDef.xml` (overlays), `DoorDef.xml` (boundaries), `GameObjectDef.xml`
  (scenery), plus NPC/item defs and the sparse loc files.
- **Sprites**: `entity24.jag` (body-part frames) + `config85.jag` (NPC table).

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
| **Bridge plank deck** (tileType-4 raised second pass) | ✅ |
| Out-of-place castle **door** (specific geometry) | ⛔ PENDING — not reproduced; needs a fresh capture |
| Bridge plank **skirt** onto road-approach tiles (World.java:724-799) | ⛔ optional |
| Lighting fine-tune (shadow depth vs RSCPlus) | ⛔ approximate |

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
- **OpenRSC** client `~/Code/openrsc/Client_Base/src/orsc/` —
  `graphics/three/World.java` (terrain/walls/roofs/bridge), `Scene.java`
  (rasteriser + TRANSPARENT), `RSModel.java`, `mudclient.java`. Server defs in
  `~/Code/openrsc/server/conf/server/defs/`.
- **deob** `~/Code/rscdump.com-runescape-classic-dump/` —
  `LeadingBot/mudclient.java` (readable), `deob106/`, `eggsampler-rsc-204-*/`.
- **RSCPlus** `~/Code/rscplus` — Alex's reference client (the pixel target).

### Useful coords (Lumbridge)
castle ~`(116-120, 648-659)`; Lum→Varrock road bridge deck ~`(106-107, 655-657)`;
swamp open water ~`(140-190, 720-760)`.

## 9. Security

The OpenRSC test password is in the `WESTWORLD_PASSWORD` env var
(`~/Code/westworld/.local.env`). **Never** print, log, commit, or embed it; grep
staged diffs for it before committing. Pushing/signing is the human's step.
