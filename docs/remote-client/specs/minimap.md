# Minimap implementation spec (F1 — `100-rsc-ui-map.md` §4.4)

Status: spec only. A mechanical implementer should need **no further investigation**.
Clone-pattern: the BANK feature (commit `ad9633c`). Touch ONLY `web/`, `cmd/cradle/`,
`docs/remote-client/specs/`.

---

## 0. Decision: client-side draw (NOT a server PNG)

The authentic RSC minimap (`drawUiTabMinimap`, Mudclient.java:3671-3756) is a 156×152
panel that **rotates with the camera** and plots colored dots for nearby
scenery/ground-items/NPCs/players around a centre white dot (self). We replicate it
**client-side on a `<canvas>`** because:

- The data is tiny (a few dozen entities) — JSON is far cheaper than re-rendering and
  re-encoding a PNG every poll.
- Rotation is a single 2D canvas transform; no 3D camera needed.
- The dot colors + rotation are pure presentation; the SPA already polls `/state` at
  400 ms (`web/src/state.ts`), so we just enrich `/state` with an `entities` block.

So the only new **backend** work is a relative-coordinate entity list added to the
existing `GET /state` response. No new endpoint is required (we extend `/state`, exactly
as the bank block was added inline in commit `ad9633c`). The minimap is a read-only
overlay with one optional interaction (left-click → walk; right-click → reuse the
existing `/pick`+`/act` menu against the dot's world tile is OUT OF SCOPE for v1, see §7).

---

## 1. Backend — `cmd/cradle/remoteclient.go` (the main new work)

### 1.1 New wire types (add near `stateBank`/`stateBankSlot`, around line 307-319)

```go
// stateEntities is the nearby-entity list the minimap (§4.4) plots. Every entity
// carries a RELATIVE tile delta (dx,dy) from the host: dx = entity.X - self.X,
// dy = entity.Y - self.Y, both in PLANE-LOCAL absolute world tiles (self and all
// entities are on the same plane; cross-plane entities are dropped). The minimap
// rotates (dx,dy) by the camera angle client-side. Radius is the Chebyshev tile
// radius these lists were gathered within (so the client can scale dot positions
// to the 152px panel). Kinds match section 4.4's dot palette.
type stateEntities struct {
	Radius int          `json:"radius"`
	Dots   []stateDot   `json:"dots"`
}

// stateDot is one minimap dot. Kind drives the color (§4.4):
//   "npc"        -> yellow  (0xFFFF00)
//   "player"     -> white   (0xFFFFFF)   (friend -> green; not backend-ready, see §5)
//   "ground_item"-> red     (0xFF0000)
//   "scenery"    -> cyan    (0x00FFFF)
// Two coordinate roles, deliberately different spaces:
//   Dx,Dy  = RELATIVE deltas in ABSOLUTE world tiles (entity.X-self.X, entity.Y-self.Y).
//            Used ONLY for minimap plotting; the renderer rotates them. Absolute Y is
//            fine here because a delta is plane-offset-invariant (both terms share the
//            same +planeOffset, which cancels).
//   X      = ABSOLUTE world X (unchanged across spaces).
//   Y      = PLANE-LOCAL world Y (entity.Y - plane*PlaneHeight) — the space host.Walk /
//            the terrain /act path expect (see §4 CRITICAL Y CONVENTION). Used ONLY for
//            click-to-walk. Do NOT use X,Y for plotting or Dx,Dy for walking.
type stateDot struct {
	Kind string `json:"kind"`
	Dx   int    `json:"dx"`
	Dy   int    `json:"dy"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Name string `json:"name,omitempty"` // tooltip; NPC/item/scenery def name when known
}
```

Add the field to `stateResponse` (struct at line 297-305), after `Bank`:

```go
type stateResponse struct {
	Self      stateSelf        `json:"self"`
	Inventory []stateInvItem   `json:"inventory"`
	Equipment []stateEquipItem `json:"equipment"`
	Chat      []chatEntry      `json:"chat"`
	Bank      *stateBank       `json:"bank,omitempty"`
	Entities  stateEntities    `json:"entities"` // NEW: nearby dots for the minimap (§4.4)
}
```

`Entities` is a value (not a pointer) and ALWAYS present — the minimap renders every
poll. `Dots` is initialized to a non-nil empty slice so JSON is `[]`, never `null`.

### 1.2 New const (add near the chatKind consts, ~line 110)

```go
// minimapRadius is the Chebyshev tile radius the /state entities block gathers
// nearby actors/items/scenery within. RSC's minimap window shows ~a 16-tile
// radius; 16 matches the authentic on-screen extent (the 152px panel at the
// default 192 zoom) and keeps the list small.
const minimapRadius = 16
```

### 1.3 Build the entity list in the `GET /state` handler

Insert this block in the `/state` handler **after** the bank block (after line 981,
`bankBlock = ...`) and **before** the `resp := stateResponse{...}` literal (line 983).
Then add `Entities: entBlock,` to that literal.

It reuses the SAME world accessors `buildLiveView` (spectate.go:189-248) and the `/pick`
path enumerate, but emits relative deltas instead of building a 3D View. `pos` (line
871) and `plane` (line 877) are already in scope.

```go
	// minimap entities (§4.4): nearby npcs/players/ground-items/scenery as
	// relative (dx,dy) dots around the host. Same world mirrors buildLiveView
	// (spectate.go) + the /pick path read; we emit deltas, not a 3D view.
	// All coords are PLANE-LOCAL absolute world tiles; cross-plane entities are
	// dropped (the minimap only shows the host's current plane). The Chebyshev
	// radius gate matches the authentic ~16-tile window.
	planeOffset := plane * world.PlaneHeight
	dots := make([]stateDot, 0, 32)
	cheb := func(ex, ey int) (dx, dy int, near bool) {
		dx, dy = ex-pos.X, ey-pos.Y
		ax, ay := dx, dy
		if ax < 0 {
			ax = -ax
		}
		if ay < 0 {
			ay = -ay
		}
		return dx, dy, ax <= minimapRadius && ay <= minimapRadius
	}

	w := host.World()

	// NPCs (yellow). NpcRecord.X/Y are absolute world tiles; TypeID -> name.
	for _, npc := range w.Npcs.All() {
		if npc.X <= 0 && npc.Y <= 0 {
			continue
		}
		if world.PlaneOf(npc.Y) != plane {
			continue
		}
		dx, dy, near := cheb(npc.X, npc.Y)
		if !near {
			continue
		}
		name := ""
		if f != nil {
			if def := f.NpcDef(npc.TypeID); def != nil {
				name = def.Name
			}
		}
		dots = append(dots, stateDot{
			Kind: "npc", Dx: dx, Dy: dy, X: npc.X, Y: npc.Y - planeOffset, Name: name,
		})
	}

	// Players (white). Index 0 is self (the centre dot, drawn client-side); skip it.
	for _, pl := range w.Players.All() {
		if pl.Index == 0 || (pl.X <= 0 && pl.Y <= 0) {
			continue
		}
		if world.PlaneOf(pl.Y) != plane {
			continue
		}
		dx, dy, near := cheb(pl.X, pl.Y)
		if !near {
			continue
		}
		dots = append(dots, stateDot{
			Kind: "player", Dx: dx, Dy: dy, X: pl.X, Y: pl.Y - planeOffset, Name: pl.Name,
		})
	}

	// Ground items (red). GroundItemRecord.X/Y absolute; ItemID -> name.
	for _, gi := range w.GroundItems.All() {
		if gi.X <= 0 && gi.Y <= 0 {
			continue
		}
		if world.PlaneOf(gi.Y) != plane {
			continue
		}
		dx, dy, near := cheb(gi.X, gi.Y)
		if !near {
			continue
		}
		name := ""
		if f != nil {
			if def := f.ItemDef(gi.ItemID); def != nil {
				name = def.Name
			}
		}
		dots = append(dots, stateDot{
			Kind: "ground_item", Dx: dx, Dy: dy, X: gi.X, Y: gi.Y - planeOffset, Name: name,
		})
	}

	// Dynamic scenery (cyan). SceneryRecord.X/Y absolute; ID -> SceneryDef name.
	for _, ds := range w.Scenery.All() {
		if ds.X <= 0 && ds.Y <= 0 {
			continue
		}
		if world.PlaneOf(ds.Y) != plane {
			continue
		}
		dx, dy, near := cheb(ds.X, ds.Y)
		if !near {
			continue
		}
		name := ""
		if f != nil {
			if def := f.SceneryDef(ds.ID); def != nil {
				name = def.Name
			}
		}
		dots = append(dots, stateDot{
			Kind: "scenery", Dx: dx, Dy: dy, X: ds.X, Y: ds.Y - planeOffset, Name: name,
		})
	}

	entBlock := stateEntities{Radius: minimapRadius, Dots: dots}
```

Then change the response literal (line 983-989) to:

```go
	resp := stateResponse{
		Self:      selfBlock,
		Inventory: invItems,
		Equipment: equipItems,
		Chat:      chatEntries,
		Bank:      bankBlock,
		Entities:  entBlock,
	}
```

**Naming-collision note:** the handler already binds `w http.ResponseWriter` as the
handler param. The block above introduces `w := host.World()` which would shadow the
ResponseWriter for the remainder of the function. The trailing code (lines 990-992)
calls `w.Header()` on the ResponseWriter. **DO NOT shadow.** Use a distinct name:
replace `w := host.World()` with `wld := host.World()` and use `wld.Npcs`, `wld.Players`,
`wld.GroundItems`, `wld.Scenery` in the four loops. (Everything else above is otherwise
correct.)

### 1.4 Static scenery is intentionally omitted (v1)

The authentic minimap also dots **static** scenery from `facts.At`/`SceneryLocs`. The
`/pick` path reads it via `f.At(tx, absY)` per-tile (pick.go:397), but there is **no
bulk "scenery near (x,y)" facts accessor** — enumerating ~33² tiles × `f.At` every 400ms
poll is wasteful. v1 plots only **dynamic** scenery (`w.Scenery.All()`, the live
GameObject mirror: lit fires, cut/regrown trees), which is what `buildLiveView` feeds the
renderer too. Static-scenery dots are a documented follow-up (see §9). This is a
graceful, authentic-enough subset, not a missing-backend blocker.

---

## 2. Host method signatures this calls (all already exist; file:line)

The `/state` additions call ONLY pure read accessors — no Host mutators, no new methods:

| Call | Signature | Location |
|---|---|---|
| `host.World()` | `func (h *Host) World() *world.World` | runtime/host.go (existing; used throughout remoteclient.go, e.g. line 513, 871) |
| `wld.Npcs.All()` | `func (s *NpcsState) All() []NpcRecord` | world/world.go:334 |
| `wld.Players.All()` | `func (s *PlayersState) All() []PlayerRecord` | world/world.go:607 |
| `wld.GroundItems.All()` | `func (s *GroundItemsState) All() []GroundItemRecord` | world/world.go:81 |
| `wld.Scenery.All()` | `func (s *DynamicScenery) All() []SceneryRecord` | world/scenery.go:88 |
| `world.PlaneOf(y)` | `func PlaneOf(y int) int` | world/self.go (used by `Coord.Plane`, self.go:41) |
| `f.NpcDef(id)` | `func (f *Facts) NpcDef(id int) *NpcDef` | facts/facts.go:60 |
| `f.ItemDef(id)` | `func (f *Facts) ItemDef(id int) *ItemDef` | facts/facts.go:63 |
| `f.SceneryDef(id)` | `func (f *Facts) SceneryDef(id int) *SceneryDef` | facts/facts.go:54 |
| `world.PlaneHeight` | `const PlaneHeight = 944` | world/self.go:17 |

Record fields used (all exported, already imported types):
- `NpcRecord{Index, X, Y, TypeID, Heading}` — world/world.go:125
- `PlayerRecord{Index, Name, X, Y, Heading}` — world/world.go:353
- `GroundItemRecord{X, Y, ItemID}` — world/world.go:50
- `SceneryRecord{X, Y, ID}` — world/scenery.go:18
- `NpcDef.Name` facts/defs.go:64; `ItemDef.Name` facts/defs.go:81; `SceneryDef.Name` (facts/defs.go:8 — verify field is `Name`).

No new imports: `world` and `facts` are already imported in remoteclient.go (lines 43, 49).
`f` (the `*facts.Facts`) is the `serveClient` param (line 337), in scope in the handler.

**Optional (NOT required):** `*State.Near(cx, cy, radius)` variants exist
(`GroundItemsState.Near` world/world.go:93; `DynamicScenery.Near` world/scenery.go:100)
but **`NpcsState`/`PlayersState` have NO `Near` method** — so for uniformity the spec
uses `.All()` + the inline `cheb` radius gate for all four kinds. Do NOT add a `Near`
method to world/ (HARD CONSTRAINT). The `.All()`+gate approach needs zero backend edits.

---

## 3. Web types — `web/src/types.ts`

Add after `Bank` (line 103), and extend `GameState`:

```ts
export type DotKind = 'npc' | 'player' | 'ground_item' | 'scenery'

export interface MinimapDot {
  kind: DotKind
  dx: number   // entity.X - self.X (plane-local world tiles)
  dy: number   // entity.Y - self.Y
  x: number    // absolute plane-local world X (for click-to-walk)
  y: number    // absolute plane-local world Y
  name?: string
}

export interface Entities {
  radius: number
  dots: MinimapDot[]
}
```

Extend `GameState` (line 105-111):

```ts
export interface GameState {
  self: Self
  inventory: InvItem[]
  equipment: EquipItem[]
  chat: ChatEntry[]
  bank?: Bank
  entities: Entities // NEW: nearby dots for the minimap
}
```

`entities` is non-optional (server always sends it), but components MUST tolerate
`state?.entities?.dots ?? []` since `state` itself is `null` before the first poll.

---

## 4. Web API — `web/src/api.ts`

**No new fetch function is needed** — the data rides on `getState()` (already polled by
`useGameState`, state.ts:37). For click-to-walk on a dot, reuse the existing `walk(...)`
helper? **No** — `walk()` takes SCREEN px/py and runs through `render.PickTile`. The
minimap has absolute world tiles, so add a tiny new helper that calls a NEW
walk-to-tile path. Two options; pick **Option A** (no backend edit):

**Option A (preferred, zero backend change): add a `walkTile` that posts to `/act` with a
synthetic terrain target.** The `/act` LaneWalk path dispatches a `KindTerrain` /
`TargetTerrain` ref through `disp.Dispatch` → `host.Walk`. A terrain `MenuTarget`
needs `{kind:'terrain', x, y, slot:-1}`. Build the candidate's walk option id the same
way Viewport does: call `pick` is screen-based, so instead post a hand-built ref to
`/act`. Verify the dispatcher accepts a terrain ref with optionId 0 = "Walk here"
(remoteclient/target.go — terrain candidates expose a single "Walk here" option at id 0).

```ts
import type { MenuTarget } from './types'

/** Walk to an absolute plane-local world tile (used by the minimap). Posts a
 *  synthetic terrain target through /act (LaneWalk -> host.Walk). */
export async function walkTile(x: number, y: number): Promise<ActResponse> {
  const ref: MenuTarget = { kind: 'terrain', x, y, slot: -1 }
  return act(ref, 0) // optionId 0 == "Walk here" for a terrain candidate
}
```

> OPTION A IS CONFIRMED SOUND (verified against source):
> - `ResolveLane` (remoteclient/dispatch.go:225) does **NOT** identity-validate
>   `KindTerrain` — its validation switch (dispatch.go:236-254) has cases only for
>   `KindNPC`/`KindPlayer`/`KindInventoryItem`. A hand-built terrain ref passes straight
>   through to `LaneWalk`.
> - For a `KindTerrain` target, `BuildMenu` yields a single option `OptWalkHere` at
>   **index 0** (menu.go:129-130), so `optionId:0` == "Walk here". ✓
> - `dispatchTerrain` (dispatch.go:426-430) calls `d.host.Walk(ctx, t.X, t.Y)` directly
>   with the ref's `X,Y`. `host.Walk(ctx, x, y)` is host.go:688.
>
> **CRITICAL Y CONVENTION:** `host.Walk` / the existing `/pick`+`/act` terrain path
> operate in **plane-local-folded** Y space — `/pick` builds candidates with
> `pos.Y - plane*world.PlaneHeight` (remoteclient.go:760), and `/walk` walks to
> `ty + plane*world.PlaneHeight` only because `PickTile` returns plane-local then re-adds
> the offset. The dot's `Y` MUST therefore be **plane-local** (`entity.Y - planeOffset`),
> exactly as §1.3 emits it (`Y: npc.Y - planeOffset`, etc.). Do NOT pass raw absolute
> `entity.Y` to `walkTile` — `dot.x` is absolute X (unchanged), `dot.y` is plane-local.
> This matches `buildLiveView` (spectate.go:195 `Y: npc.Y - planeOffset`) precisely.

**Option B (fallback, ALSO zero backend change): reuse `GET /walk` with a precomputed
screen pixel.** Not viable from the minimap (no screen projection available client-side).
Therefore if Option A fails the dispatcher check, the v1 minimap is **display-only**
(no click-to-walk) — drop the `onClick` walk in §5 and ship dots-only. This is an
acceptable v1 (the authentic minimap's primary purpose is the radar display).

> If neither walk path is wired, NOTE the gap: "minimap click-to-walk needs a
> walk-to-absolute-tile endpoint; `/walk` is screen-pixel only and terrain `/act` dispatch
> requires a live candidate." Do NOT add a backend route (HARD CONSTRAINT) — ship
> display-only and file the follow-up.

---

## 5. Component — `web/src/components/Minimap.tsx` (new file)

A 152×152 `<canvas>` viewport overlay (top-right of `#view`, like RSC's minimap tab
position). Rotates dots by the camera angle so "up" on the minimap is the camera's
forward direction. Centre = white self dot.

Props:

```ts
import type { Camera4 } from './Viewport'
import type { Entities } from '../types'

interface MinimapProps {
  entities: Entities | undefined
  rot: number            // camera.rot, 0..255 (RSC 256-step compass)
  selfHeading?: number   // optional: draw a heading tick (0..7); from state.self.heading
}
```

### 5.1 Geometry / rotation

- Canvas is `SIZE = 152` px. Centre `c = SIZE/2 = 76`.
- `radius` tiles map to `SIZE/2` px: `scale = (SIZE/2 - PAD) / radius` where `PAD = 6`.
  Use `entities?.radius ?? 16`.
- Rotation: RSC's minimap rotates by `(cameraRotation) & 0xFF` (256-step). The
  rendered frame uses `rot` in the same 256 space (App.tsx ROT step = 4, range 0..255).
  Convert to radians: `theta = (rot / 256) * 2*Math.PI`. The frame's camera yaw =
  `rot*4` in the 2048-step `sine11` space (pick.go:38), i.e. `rot/256` of a full turn —
  consistent. Sign: RSC rotates the map so the world spins opposite the camera. Use
  `theta` first; if dots spin the wrong way, negate (`-theta`). Document the chosen sign.
- For each dot `(dx, dy)`:
  ```
  // world dx,dy -> screen: +dx is east (world +X), +dy is south (world +Y).
  // Minimap: north(world -Y) should point "up" before rotation. Map world (dx,dy)
  // to canvas (right=+x, down=+y): px0 = dx, py0 = dy. Then rotate by theta.
  const px = c + (dx * Math.cos(theta) - dy * Math.sin(theta)) * scale
  const py = c + (dx * Math.sin(theta) + dy * Math.cos(theta)) * scale
  ```
  Clip dots whose `(px,py)` fall outside the circle of radius `SIZE/2` (draw only if
  `Math.hypot(px-c, py-c) <= SIZE/2 - 1`).

### 5.2 Colors (§4.4 — use the existing CSS palette hexes)

| kind | color | CSS var (styles.css:7-13) |
|---|---|---|
| self (centre) | `#ffffff` | `--rsc-white` |
| `npc` | `#ffff00` | `--rsc-yellow` |
| `player` | `#ffffff` | `--rsc-white` (friend → `#00ff00` `--rsc-green`; NOT backend-ready, see note) |
| `ground_item` | `#ff0000` | `--rsc-red` |
| `scenery` | `#00ffff` | cyan (note: `--rsc-cyan` is `#c8ffff`; §4.4 specifies pure `0x00FFFF` — use literal `#00ffff` for fidelity) |

Hardcode the hex strings in a `const DOT_COLOR: Record<DotKind,string>` plus
`SELF_COLOR='#ffffff'` (canvas can't read CSS vars without `getComputedStyle`; the hexes
are stable, mirror §4.4 exactly).

**Friend (green) not backend-ready:** `is_friend` is a stub (`runtime/views_world.go:368`
returns `false`; no server-side friends-list mirror exists). v1 draws ALL players white.
NOTE this in output; do NOT add friend tracking to world/ or runtime/.

### 5.3 Draw routine (inside a `useEffect` / `useLayoutEffect` keyed on `[entities, rot]`)

```tsx
import { useEffect, useRef } from 'react'
import { walkTile } from '../api'
import { useUI } from '../ui'
import type { Camera4 } from './Viewport'
import type { DotKind, Entities, MinimapDot } from '../types'

const SIZE = 152, PAD = 6
const DOT_COLOR: Record<DotKind, string> = {
  npc: '#ffff00', player: '#ffffff', ground_item: '#ff0000', scenery: '#00ffff',
}
const SELF_COLOR = '#ffffff'

export function Minimap({ entities, rot }: { entities: Entities | undefined; rot: number }) {
  const ref = useRef<HTMLCanvasElement | null>(null)
  const ui = useUI()
  const dots = entities?.dots ?? []
  const radius = entities?.radius ?? 16

  useEffect(() => {
    const cv = ref.current
    if (!cv) return
    const g = cv.getContext('2d')
    if (!g) return
    const c = SIZE / 2
    const scale = (SIZE / 2 - PAD) / radius
    const theta = (rot / 256) * Math.PI * 2 // sign per §5.1; flip if dots counter-rotate
    // dark circular field
    g.clearRect(0, 0, SIZE, SIZE)
    g.fillStyle = '#0a140a'
    g.beginPath(); g.arc(c, c, SIZE / 2, 0, Math.PI * 2); g.fill()
    // dots
    for (const d of dots) {
      const px = c + (d.dx * Math.cos(theta) - d.dy * Math.sin(theta)) * scale
      const py = c + (d.dx * Math.sin(theta) + d.dy * Math.cos(theta)) * scale
      if (Math.hypot(px - c, py - c) > SIZE / 2 - 1) continue
      g.fillStyle = DOT_COLOR[d.kind] ?? '#fff'
      g.fillRect(Math.round(px) - 1, Math.round(py) - 1, 3, 3)
    }
    // self (centre, slightly larger)
    g.fillStyle = SELF_COLOR
    g.fillRect(c - 1, c - 2, 3, 4)
  }, [dots, radius, rot])

  // hit-test a canvas click back to the nearest dot (within ~5px); else walk to the
  // world tile under the click (inverse of the projection above).
  const onClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const cv = ref.current; if (!cv) return
    const r = cv.getBoundingClientRect()
    const mx = (e.clientX - r.left) * (SIZE / r.width)
    const my = (e.clientY - r.top) * (SIZE / r.height)
    const c = SIZE / 2, scale = (SIZE / 2 - PAD) / radius
    const theta = (rot / 256) * Math.PI * 2
    // nearest dot first (left-click "walk to that thing's tile")
    let best: MinimapDot | null = null, bestD = 6 * 6
    for (const d of dots) {
      const px = c + (d.dx * Math.cos(theta) - d.dy * Math.sin(theta)) * scale
      const py = c + (d.dx * Math.sin(theta) + d.dy * Math.cos(theta)) * scale
      const dd = (px - mx) ** 2 + (py - my) ** 2
      if (dd < bestD) { bestD = dd; best = d }
    }
    if (best) { void walkTile(best.x, best.y).then((res) => { if (res.message) ui.flash(res.message) }); return }
    // empty space: inverse-rotate the click back to a world delta, then walk there.
    const rx = (mx - c) / scale, ry = (my - c) / scale
    const inv = -theta
    const dx = Math.round(rx * Math.cos(inv) - ry * Math.sin(inv))
    const dy = Math.round(rx * Math.sin(inv) + ry * Math.cos(inv))
    // self.x/self.y unknown here; pass dx,dy is insufficient. See §5.4.
  }

  return (
    <canvas
      ref={ref} width={SIZE} height={SIZE} className="minimap"
      onClick={onClick}
      onContextMenu={(e) => e.preventDefault()}
      title="minimap"
    />
  )
}
```

### 5.4 Click-to-walk needs the self tile

To walk to **empty-space** clicks the component needs `self.x/self.y` to turn a relative
delta into an absolute tile. Two clean options:

- **Preferred:** add a `self` prop `{ x: number; y: number }` (from `state.self`). Then
  empty-space walk = `walkTile(self.x + dx, self.y + dy)`. Dots already carry absolute
  `x,y` so they don't need it.
- **Simpler v1:** support ONLY dot-click walk (dots carry absolute `x,y`); for
  empty-space clicks do nothing (or `ui.flash('click a dot to walk')`). Drop the
  inverse-rotation block. This avoids the extra prop and is acceptable for v1.

Choose the simpler v1 unless self is trivially available — App already has
`state.self`, so passing `self={{x,y}}` is one line; prefer the full version.

Final props (full version):

```ts
export function Minimap({ entities, rot, self }: {
  entities: Entities | undefined
  rot: number
  self?: { x: number; y: number }
})
```

### 5.5 Interaction summary

- **Left-click on a dot** → `walkTile(dot.x, dot.y)` → flash the result message.
- **Left-click empty** → `walkTile(self.x + dx, self.y + dy)` (full version) or no-op (v1).
- **Right-click** → `e.preventDefault()` only (suppress browser menu). A full
  right-click verb menu (Attack NPC / Pick up item) is OUT OF SCOPE — it would require a
  world-tile `/pick` (the existing `/pick` is screen-pixel only). Documented follow-up §9.

---

## 6. Wiring — `web/src/App.tsx` + styles

### 6.1 App.tsx

Import and mount the minimap as a viewport overlay. The minimap should sit ON TOP of the
streamed frame in the top-right corner. Mount it as a sibling of `<Hud>` inside
`<Viewport hud={...}>` so it shares the `#view` stacking context, OR as an absolutely
positioned overlay in `#app`. Cleanest: pass it through the existing `hud` slot by
composing both into a fragment.

Add import (after line 14):

```tsx
import { Minimap } from './components/Minimap'
```

Change the `<Viewport hud=...>` prop (line 94) to compose Hud + Minimap:

```tsx
<Viewport
  camera={camera}
  animRef={animRef}
  hud={
    <>
      <Hud self={state?.self ?? null} camera={camera} />
      <Minimap
        entities={state?.entities}
        rot={camera.rot}
        self={state?.self ? { x: state.self.x, y: state.self.y } : undefined}
      />
    </>
  }
/>
```

`Viewport` already renders `{hud}` after the `<img>` (Viewport.tsx:81), so the fragment
lands in the `#view` relative container — the minimap positions itself top-right via CSS.

### 6.2 styles.css

Add after the `.hud` block (after line 60), inside the viewport-overlay group:

```css
/* ---- minimap (§4.4): top-right viewport overlay, 152px round radar ---- */
.minimap {
  position: absolute;
  right: 8px;
  top: 6px;
  width: 152px;
  height: 152px;
  border-radius: 50%;
  border: 2px solid #1a3d1a;
  box-shadow: 0 0 0 2px #000;
  background: #0a140a;
  image-rendering: pixelated;
  cursor: crosshair;
  z-index: 30;          /* above the frame img, below flash(60)/menu(50)/modal(40) */
  pointer-events: auto; /* override .hud's pointer-events:none if nested near it */
}
```

Note `.hud` sets `pointer-events:none`; the minimap is a SIBLING (not a child) of `.hud`
inside the fragment, so it keeps its own `pointer-events:auto`. Verify the canvas
receives clicks after build.

---

## 7. Authentic-fidelity notes (§4.4 reconciliation)

- §4.4 panel is 156×152 with a ~40px left inset (the tab strip). Our SPA has no native
  tab strip, so we use a clean 152×152 round radar in the viewport corner. Acceptable
  divergence (documented).
- §4.4 zoom is `192 + jitter`; we map `radius` tiles → panel px directly (16 tiles ≈ the
  192-zoom extent). No jitter (the authentic ±2 jitter is cosmetic dither; skip).
- §4.4 compass sprite (`spriteMedia+24`) is omitted v1 (no sprite pipeline for media
  sprites yet — `/sprite` is item-only, sprites.go:27). A north tick can be drawn as a
  small white notch at the top of the ring if desired (`g.fillRect(c-1, 1, 2, 4)` before
  rotation), but it does not rotate with content. Optional.
- Dot size: RSC plots 2–3px squares; we use 3×3. Self is 3×4 to read as the centre.

---

## 8. Test plan

Backend (`cmd/cradle`):
1. `go build ./...` from repo root — compiles with the new struct + handler block (watch
   the `wld`-not-`w` shadow fix from §1.3).
2. `curl -s localhost:8090/state | jq '.entities'` while logged in near NPCs/items →
   expect `{ "radius": 16, "dots": [ {kind,dx,dy,x,y,name}, ... ] }`, never `null` dots.
3. Sanity: every dot satisfies `abs(dx) <= 16 && abs(dy) <= 16`. `dot.x == self.x + dot.dx`
   holds (both X are absolute). For Y the spaces differ on purpose: `dot.y` is
   plane-local (for walking) while `dot.dy` is the absolute-space delta — so
   `dot.y + plane*944 == self_abs.Y + dot.dy` should hold (i.e. folding `dot.y` back to
   absolute and subtracting the absolute self-Y recovers `dot.dy`).
4. Stand on a different plane / upstairs → dots for other-plane entities disappear
   (PlaneOf filter).
5. Drop an item (`/chat` `::` command or in-game) → a red dot appears within one poll
   (~400 ms); pick it up → dot clears.

Web (`npm --prefix web run build` then `go build`):
6. `npm --prefix web run build` — TS compiles (new types, component, App wiring).
7. Load `http://localhost:8090/` — round minimap top-right of the viewport; white centre
   dot; colored dots for nearby entities.
8. Rotate the camera (`< >` / arrow keys) — dots rotate around the centre, staying
   locked to world positions (verify a stationary NPC's dot orbits the centre as the
   camera spins, i.e. its WORLD position is fixed). If it counter-rotates vs the 3D
   frame, flip the `theta` sign (§5.1) and rebuild.
9. Left-click a dot → host walks toward that tile (flash shows "Walk here ..." or the
   dispatch message); the on-screen frame confirms movement.
10. Right-click the minimap → no browser context menu (preventDefault).
11. `npx tsc --noEmit` (or the build's typecheck) passes with no `any`/unused warnings.

---

## 9. Follow-ups (explicitly out of v1 scope; do NOT implement now)

- Static-scenery cyan dots (needs a bulk `facts` "scenery near (x,y)" accessor — not
  present; per-tile `f.At` ×33² per poll is too costly). File against facts/.
- Friend (green) player dots — needs a server-side friends-list mirror; `is_friend` is a
  stub (`runtime/views_world.go:368`). Do NOT add to world/runtime (HARD CONSTRAINT);
  note the gap.
- Right-click verb menu on dots — needs a world-tile `/pick` (current `/pick` is
  screen-pixel only). A `/pickTile` that runs `remoteclient.BuildCandidates` against an
  absolute tile would be the clean addition; out of scope.
- Compass sprite (`spriteMedia+24`) — needs a media-sprite branch in `/sprite`
  (sprites.go is item-only). Optional.

---

## 10. File-change checklist (touch ONLY these)

| File | Change |
|---|---|
| `cmd/cradle/remoteclient.go` | add `stateEntities`/`stateDot` structs (§1.1); add `minimapRadius` const (§1.2); add entity-gathering block + `Entities` field in `/state` (§1.3), using `wld` to avoid shadowing `w http.ResponseWriter` |
| `web/src/types.ts` | add `DotKind`, `MinimapDot`, `Entities`; extend `GameState` (§3) |
| `web/src/api.ts` | add `walkTile(x,y)` (§4 Option A) — or omit if shipping display-only |
| `web/src/components/Minimap.tsx` | NEW component (§5) |
| `web/src/App.tsx` | import + mount `<Minimap>` via the `hud` slot fragment (§6.1) |
| `web/src/styles.css` | add `.minimap` (§6.2) |
| `docs/remote-client/specs/minimap.md` | this file |

NO edits to `reference/`, `world/`, or `runtime/`. NO new backend route (extend `/state`).
NO `git commit`. NO server restart.
