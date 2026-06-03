# Layer 4 — Browser UI (the SPA)

**Status:** Design (M1)
**Layer:** 4 of 4 (see `00-overview.md` §4)
**Consumes:** the Layer 3 HTTP/JSON contract (`30-http-api.md`)
**Served by:** `cmd/cradle/clientpage.go` (page asset) + `cmd/cradle/remoteclient.go` (handlers, Layer 3)

This document specifies the single-page browser application that turns the
streamed PNG viewport into a usable remote RSC client: a game viewport with
left-click default-action and right-click context menus, an inventory grid, an
equipment summary, and a chat box. It extends the frame-polling /
pixelated-scaling pattern already proven in `spectatePage`
(`cmd/cradle/spectate.go`).

The implementer of this layer should be able to write the page **mechanically**
from this doc without inventing any wire shapes. Every `fetch` call below names
its endpoint, its request, and the response JSON it destructures. Those shapes
are the contract owned by `30-http-api.md`; this doc restates them at the field
level so the UI and the handlers cannot drift. If a shape here disagrees with
`30-http-api.md` once that doc lands, **`30-http-api.md` wins** and this doc is
corrected — but they are authored together to agree.

---

## 0. Relationship to `-spectate`

The existing `-spectate` mode and its `spectatePage` stay **exactly as they
are** — do not edit `spectate.go`'s page or handlers. The full client is a new,
parallel mode behind a new `-client` flag (config `cfg.client` /
`cfg.clientAddr`, mirroring `cfg.spectate` / `cfg.spectateAddr`). Both modes
share the Go helpers `buildLiveView`, `motionTracker`, `waitForLivePosition`,
`renderOne`, and the serialized walk-worker pattern; only the page and the
extra endpoints differ. Keeping them separate means the working spectator (an
agent's read-only debugging viewport) is never put at risk by client work, and
the two pages can diverge freely (the client page is much heavier).

The client viewport reuses the spectator's frame contract verbatim
(`GET /frame?rot&zoom&w&h&anim` → PNG, `GET /pos` → `{x,y,plane}`), so the
camera/scaling/HUD logic is a direct lift from `spectatePage`.

---

## 1. Page layout

A fixed full-window grid. The viewport dominates; a right-hand sidebar stacks
inventory + equipment; the chat box spans the bottom; the HUD floats over the
viewport. No external CSS/JS — one inline `<style>` and one inline `<script>`,
authentic chunky look (`image-rendering:pixelated`, `font:13px monospace`,
greenish-on-black RSC palette).

```
┌───────────────────────────────────────────────┬──────────────────┐
│ #hud (pos/rot/zoom + hints, floats top-left)    │  #equip          │
│                                                 │   (worn items)   │
│                                                 ├──────────────────┤
│              #viewport (img#v)                  │  #inv            │
│              left-click  = default action       │   6×5 cell grid  │
│              right-click = context menu         │   (30 slots)     │
│                                                 │                  │
│              #menu  (absolutely positioned      │                  │
│               at cursor on right-click)         │                  │
│              #label (hover target name)         │                  │
│              #flash (transient toast)           │                  │
├─────────────────────────────────────────────────┤                  │
│ #chatlog  (scrollback, newest at bottom)         │                  │
│ #chatinput  <input>  (say / command / PM)        │                  │
└──────────────────────────────────────────────────┴──────────────────┘
```

CSS grid skeleton (two columns: viewport+chat stack on the left, sidebar on the
right):

```css
html,body{margin:0;height:100%;background:#000;overflow:hidden;
          font:13px monospace;color:#9f9}
#app{position:absolute;inset:0;display:grid;
     grid-template-columns:1fr 232px;            /* sidebar 232px (6 cells * 36 + pad) */
     grid-template-rows:1fr 168px;               /* chat strip 168px tall */
     grid-template-areas:"view side" "chat side"}
#view{grid-area:view;position:relative;overflow:hidden;background:#000}
#v{position:absolute;inset:0;width:100%;height:100%;
   object-fit:contain;image-rendering:pixelated;cursor:crosshair}
#chat{grid-area:chat;display:flex;flex-direction:column;
      border-top:2px solid #1a3d1a;background:#0a140a}
#side{grid-area:side;display:flex;flex-direction:column;
      border-left:2px solid #1a3d1a;background:#0a140a;overflow:hidden}
/* HUD / overlays inside #view */
#hud{position:absolute;left:8px;top:6px;white-space:pre;
     text-shadow:0 0 3px #000;pointer-events:none}
#label{position:absolute;left:8px;bottom:8px;color:#ff8;
       text-shadow:0 0 3px #000;pointer-events:none}
#flash{position:absolute;right:10px;top:8px;color:#ff8;opacity:0;
       transition:opacity .2s;pointer-events:none;font-size:15px}
/* context menu */
#menu{position:absolute;display:none;z-index:50;min-width:120px;
      background:#23282b;border:1px solid #000;
      box-shadow:2px 2px 0 #000;font:13px monospace}
#menu .hdr{background:#0a0;color:#000;padding:2px 6px;font-weight:bold}
#menu .opt{padding:2px 8px;color:#fff;cursor:pointer;white-space:nowrap}
#menu .opt:hover{background:#0a0;color:#000}
/* inventory grid: 6 columns of 36px cells */
#inv{display:grid;grid-template-columns:repeat(6,36px);gap:1px;padding:4px}
.cell{width:36px;height:36px;background:#1a1f1a;border:1px solid #000;
      position:relative;cursor:pointer;display:flex;align-items:center;
      justify-content:center;color:#cfc;font-size:11px;text-align:center}
.cell.wield{outline:1px solid #ff0}
.cell .qty{position:absolute;right:1px;bottom:0;color:#ff0;font-size:9px}
/* equipment summary: one line per worn slot */
#equip{padding:4px;border-bottom:2px solid #1a3d1a;font-size:12px;line-height:1.4}
#equip .row{display:flex;justify-content:space-between}
/* chat */
#chatlog{flex:1;overflow-y:auto;padding:4px 6px;line-height:1.35;
         white-space:pre-wrap;word-break:break-word}
#chatlog .you{color:#0ff} #chatlog .npc{color:#ff8}
#chatlog .pm{color:#f9f} #chatlog .sys{color:#9a9}
#chatinput{border:0;border-top:1px solid #1a3d1a;background:#06100a;
           color:#9f9;font:13px monospace;padding:4px 6px;outline:none}
```

Notes:
- The sidebar width (`232px`) and cell size (`36px`) are tuned so 30 inventory
  slots lay out 6×5. RSC inventory is 30 slots; we render all 30 cells always
  (empty cells are blank but still present, so the grid never reflows).
- The viewport keeps the spectator's `object-fit:contain` + pixelated scaling,
  so the render stays cheap (small `w`/`h`) while the display fills the column.
  All screen→frame-pixel mapping (un-letterboxing) is identical to
  `spectatePage`'s click handler and is the single source of `px,py` for both
  `/walk` and `/pick`.

---

## 2. Interaction model

### 2.1 Viewport — left-click = default action

Left-click maps the click to a frame pixel exactly as `spectatePage` does
(undo `object-fit:contain` letterboxing using `getBoundingClientRect` +
`scale=min(r.width/w, r.height/h)`), then POSTs `/pick` and **immediately acts
on the top (default) option of the nearest candidate** — RSC's default-action
semantics. There is no menu render on left-click.

- If `/pick` returns no candidates with any option other than "Walk here"
  (i.e. only terrain), left-click falls through to a walk: POST `/walk`
  (same as spectator). This is the common "click empty ground to move" case.
- Concretely: left-click → `pick(px,py)` → if `menu.candidates[0]` exists, call
  `act(candidates[0], candidates[0].options[0])`; the server-side default for a
  terrain-only pick is the walk, so we can also just let the top option be
  `{verb:"Walk here", ...}` and dispatch it uniformly. Either is acceptable; the
  uniform path (always `act` the top option) is preferred because it keeps one
  code path and lets Layer 2 decide walk-vs-interact.

### 2.2 Viewport — right-click = context menu

Right-click (`contextmenu` event, `e.preventDefault()`) POSTs `/pick` and
renders a context menu at the cursor:

1. `pick(px,py)` returns a depth-ordered candidate list (nearest-camera-first).
2. The menu is built as a flat list: for each candidate, its option rows are
   appended (the **first candidate's first option is the default / top row**).
   A candidate's header (its label, e.g. `Goblin (level-5)`) is shown as a
   non-clickable `.hdr` separator above its options when there is more than one
   candidate, so overlapping targets are distinguishable. With a single
   candidate the header may be collapsed into the first option for compactness.
3. Each `.opt` row's text is `option.verb + " " + candidate.label`
   (RSC reads "Attack Goblin", "Talk-to Man", "Chop Tree", "Take Logs",
   "Walk here"). The trailing `Examine` and `Walk here` options come last,
   matching authentic ordering (Layer 2 orders them; the UI does not reorder).
4. Clicking a `.opt` POSTs `/act` with that candidate's `ref` + the option's
   `id`, hides the menu, and toasts the verb via `flash()`.
5. The menu hides on: any left-click elsewhere, `Escape`, scroll, or a new
   right-click. It is clamped to stay inside the window
   (`left=min(x, innerWidth-menuW)`, same for top).

### 2.3 Viewport — hover label

On `mousemove` over the viewport (throttled, see §3) the page may show the
top target's name in `#label`. M1 keeps this cheap: hover does **not** hit the
server on every move. Instead, left/right interactions already surface labels,
and a lightweight optional hover-pick (debounced ~120ms, abortable) can fill
`#label`. Hover-pick is **stretch**; the baseline ships without it and `#label`
is populated only transiently from the last `/pick`. Implementers should gate
hover-pick behind a `const HOVER=false` so it is trivially toggled.

### 2.4 Inventory cell

Each cell is keyed by slot index 0..29. Cell content comes from `/state`
(`inventory[]`, §5.3). Rendering: if a slot has an item, show its short label
(item name truncated, or an id) and, if stacked, a `.qty` corner; wielded items
get the `.wield` outline.

- **Left-click** an occupied cell → default item command:
  POST `/act` with a synthesized inventory target ref
  (`{kind:"inventory_item", slot}`) and `option.id` = the slot's default
  command id (the cell carries the default verb from `/state`; see §5.3
  `slot.defaultOptionId`). For a wieldable item the default is Wield/Wear; for
  food it's Eat; etc. — Layer 2 derives the verb list, the UI just sends the
  default id.
- **Right-click** an occupied cell → item context menu built **client-side from
  a small fixed verb set augmented by the slot's `options[]` from `/state`**:
  the slot already carries its authentic `options` array (e.g.
  `[{id, verb:"Wield"}, {id, verb:"Drop"}, {id, verb:"Examine"}]`). The UI
  renders them in the same `#menu` widget used by the viewport, then on click
  POSTs `/act {ref:{kind:"inventory_item",slot}, option:id}`. The UI does NOT
  hardcode Use/Wield/Drop/Examine — it renders whatever the slot's `options`
  list contains, so wieldables, food, and tools each get their real verbs.
  (If `30-http-api.md` chooses to omit per-slot options from `/state` to keep
  it small, the inventory right-click instead POSTs `/pick` with a
  `slot` field and gets the menu the same way the viewport does; the UI is
  written to accept either, preferring inline `options` when present.)
- **Drag onto a target (use-item-on)** — **OPTIONAL / STRETCH for M1.** HTML5
  drag of a cell onto the viewport or onto another cell, on drop, resolves the
  drop target via `/pick` (viewport) or the slot under the cursor (cell), then
  POSTs `/act` with a use-on shape (`{ref: dropTarget, option: useWith,
  withSlot: srcSlot}`). Implement only after the core menu flow works; the page
  must function fully without it. Gate behind `const DRAG=false`.

### 2.5 Chat input

A single `<input id="chatinput">`. On `Enter` with non-empty trimmed text:

- Leading `::` or `/` → **command**: strip the prefix, POST `/chat
  {kind:"command", text}` (Layer 2 → `Host.Command`). (`::` is the RSC
  in-client command prefix; `/` is offered as an ergonomic alias.)
- Leading `@name message` → **private message**: parse the first token after
  `@` as the recipient, the rest as the body, POST `/chat {kind:"pm",
  to:name, text:body}` (Layer 2 → `Host.PrivateMessage`).
- Otherwise → **public say**: POST `/chat {kind:"say", text}` (Layer 2 →
  `Host.Say`).

After send, clear the input. The sent line is **not** echoed locally; it will
appear in the next `/state` chat snapshot if the server reflects it (and the
`you`-coloured class is applied by matching the host name). The input is not
captured while a viewport key (rotate/zoom) would fire: viewport key handlers
check `document.activeElement !== chatinput` so typing in chat doesn't rotate
the camera, and `Escape` blurs the input.

---

## 3. The polling loop

Three independent timers, mirroring `spectatePage`'s structure, each guarded so
requests never stack:

| Loop | Interval | Endpoint | Guard | Purpose |
|---|---|---|---|---|
| frame | ~33ms (≈30fps) | `GET /frame?rot&zoom&w&h&anim` → PNG | `busy` flag + `Image.onload` swap (identical to `spectatePage`) | live viewport, smooth glide |
| state | 300–500ms | `GET /state` → JSON (§5.3) | `stateBusy` flag; skip if a request is in flight | inventory, equipment, chat scrollback |
| pos | 500ms | `GET /pos` → `{x,y,plane}` | none (tiny) | HUD coords |
| anim | 160ms | (local) `anim=(anim+1)&0x3fffffff` | none | model-swap flicker pace |

Rules carried over from `spectatePage`:
- **Frame requests must not stack.** Reuse the exact `busy=true` /
  `new Image(); n.onload=()=>{img.src=n.src; busy=false}` pattern. `tick()`
  no-ops while a frame is in flight; rotate keys still advance `rot` each tick
  independent of frame readiness so spinning stays smooth.
- **`/state` is its own slower loop** so the heavier JSON (inventory + chat)
  doesn't compete with the frame cadence. Its handler is `stateBusy`-guarded
  the same way.
- `/pick` and `/act` and `/chat` are **event-driven**, not polled. They are
  POSTs fired on click / context-menu / Enter; the menu render waits on the
  `/pick` response. They share no busy flag with the loops.
- All GETs append a cache-buster (`&t=performance.now()`) and the server sends
  `Cache-Control:no-store`, exactly as today.

---

## 4. How the page is served

**Decision: a Go raw-string `const clientPage` in its own file
`cmd/cradle/clientpage.go`.** (Same mechanism as `spectatePage`, but a separate
file so the page is not buried in handler code.)

Rationale:
- **Consistency + zero new build steps.** `spectatePage` is already a raw-string
  const served via `fmt.Fprintf(w, spectatePage, ...)`. Matching that keeps one
  serving idiom across both modes; an implementer copies a known-good pattern.
- **`go:embed` buys little here and costs a file-layout constraint.** `go:embed`
  shines for many/large/binary assets or assets edited by non-Go tooling. This
  is one self-contained HTML+CSS+JS string with a handful of `%d`/`%s` config
  substitutions (host name, zoom, w, h) that are most naturally injected with
  `fmt.Fprintf` — exactly the spectator's approach. An embed would force a
  sidecar `.html` file and lose the trivial `printf` parameterization, or
  require post-load string replacement, for no real gain at this size.
- **Single-binary, no runtime file dependency.** The const compiles into the
  binary; `cradle` needs no asset directory at runtime. (`go:embed` also yields
  a single binary, so this is a wash — but the const needs no `embed` import and
  no extra file at all beyond the one `.go`.)

**File-boundary rule (enforced):** `cmd/cradle/clientpage.go` contains **only**
the page asset — `const clientPage = ` `...` `` and, if helpful, small
page-only constants (e.g. default cell count). It contains **no** `http`
handlers, no `net/http` import, no Go logic. Layer 3's
`cmd/cradle/remoteclient.go` owns the `GET /` handler and references the
`clientPage` symbol:

```go
// in remoteclient.go (Layer 3)
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" { http.NotFound(w, r); return }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    fmt.Fprintf(w, clientPage, cfg.username, cfg.renderZoom, cfg.renderW, cfg.renderH)
})
```

The `%`-format verbs in `clientPage` (and their order) are part of this
contract: `%s` host name (title), `%d` zoom, `%d` w, `%d` h — matching
`spectatePage`'s signature so the handler is a copy. Any literal `%` in the
CSS/JS must be escaped as `%%` (as `spectatePage` already does for `100%%`).

---

## 5. Concrete skeleton

### 5.1 HTML skeleton

```html
<!doctype html><html><head><meta charset="utf-8">
<title>cradle client — %s</title>
<style>/* the §1 CSS, with %% on literal percents */</style>
</head><body>
<div id="app">
  <div id="view">
    <img id="v" draggable="false">
    <div id="hud"></div>
    <div id="label"></div>
    <div id="flash"></div>
    <div id="menu"></div>
  </div>
  <div id="chat">
    <div id="chatlog"></div>
    <input id="chatinput" autocomplete="off" spellcheck="false"
           placeholder="say…  ::cmd  @name pm">
  </div>
  <div id="side">
    <div id="equip"></div>
    <div id="inv"></div>
  </div>
</div>
<script>/* the §5.4 JS */</script>
</body></html>
```

### 5.2 Wire shapes consumed (the contract this UI destructures)

These are the exact JSON shapes the UI reads. `30-http-api.md` is their owner;
they are restated here so the implementer needs no other doc to write the JS.
Field names are lowerCamel JSON.

**`POST /pick`** — request `{px, py, rot, zoom, w, h, slot?}` (slot present only
for the inventory-right-click fallback path). Response:

```jsonc
{
  "candidates": [                 // depth-ordered, nearest-camera-first
    {
      "ref": {                    // opaque-ish target identity; echoed back to /act verbatim
        "kind": "npc",            // npc|player|self|ground_item|scenery|boundary|terrain|inventory_item
        "index": 42,              // server actor index (npc/player/self)
        "defId": 0,               // scenery/boundary def id (0 if n/a)
        "itemId": 0,              // ground_item / inventory item id (0 if n/a)
        "x": 327, "y": 552,       // world tile (absolute Y)
        "dir": 0,                 // boundary direction (0 if n/a)
        "slot": -1                // inventory slot (-1 if n/a)
      },
      "label": "Goblin (level-5)",          // header / suffix for option text
      "examine": "An ugly green creature.", // examine text (no packet sent)
      "options": [                          // authentic verbs, default-first
        {"id": 0, "verb": "Attack"},
        {"id": 1, "verb": "Talk-to"},
        {"id": 2, "verb": "Examine"}
      ]
    }
    // ... more candidates (deeper targets), terrain "Walk here" last
  ]
}
```

The UI treats `ref` as **opaque**: it never constructs a ref from scratch for a
world target; it only echoes the `ref` it received back into `/act`. (The one
exception is the inventory-item synthetic ref `{kind:"inventory_item", slot}`,
which the UI may build directly when right-clicking a cell that already carries
its `options` from `/state` — see §2.4.)

**`POST /act`** — request `{ref, option}` where `ref` is a candidate's `ref`
(echoed verbatim) and `option` is the chosen option's `id`. For use-item-on
(stretch) add `withSlot`. Response `{ok: true}` or `{ok:false, error:"…"}`.
The UI does not block on the result beyond toasting success/failure; the world
effect shows up in subsequent `/frame` + `/state`.

**`POST /chat`** — request one of:
`{kind:"say", text}`, `{kind:"command", text}`, `{kind:"pm", to, text}`.
Response `{ok:true}`.

**`GET /pos`** — `{x, y, plane}` (unchanged from spectator).

**`GET /frame…`** — PNG bytes (unchanged from spectator).

### 5.3 `GET /state` shape

```jsonc
{
  "inventory": [                       // exactly 30 entries; empty slots have itemId 0
    {
      "slot": 0,
      "itemId": 517,
      "name": "Bronze sword",
      "amount": 1,                     // stack count (1 for non-stacked)
      "wielded": false,
      "defaultOptionId": 0,            // id of the left-click default verb
      "options": [                     // authentic item verbs, default-first
        {"id": 0, "verb": "Wield"},
        {"id": 1, "verb": "Drop"},
        {"id": 2, "verb": "Examine"}
      ]
    }
    // ... 30 total
  ],
  "equipment": [                       // worn-item summary (sparse: only equipped)
    {"slot": "Weapon", "name": "Bronze sword"},
    {"slot": "Body",   "name": "Leather body"}
  ],
  "stats": {                           // optional in M1; HUD may show combat level
    "combatLevel": 3, "hits": 10, "maxHits": 10
  },
  "chat": [                            // recent lines, oldest-first, capped (~50)
    {"seq": 1201, "kind": "public", "who": "Bob",  "text": "hi"},
    {"seq": 1202, "kind": "npc",    "who": "Man",  "text": "Hello there!"},
    {"seq": 1203, "kind": "private","who": "Alice","text": "wanna trade?", "in": true},
    {"seq": 1204, "kind": "system", "who": "",     "text": "Welcome to RuneScape."}
  ]
}
```

The UI uses `chat[].seq` as a monotonic cursor so it only appends **new** lines
(it remembers the last rendered `seq`); this avoids re-rendering and lets the
log scroll naturally. `kind` selects the colour class
(`public→default/you`, `npc→npc`, `private→pm`, `system→sys`); a `who` equal to
the host's own name (or `in:false` on a PM) colours it as `you`/outbound. If
Layer 3 prefers to expose chat at a separate `GET /chat` poll instead of inside
`/state`, the UI's `applyState` simply splits into two appliers; the field
shapes are unchanged.

### 5.4 JS module structure (named functions)

Plain top-level functions in one `<script>` (no modules/bundler). Grouped by
concern; each names its endpoint + the shape it touches.

**Camera / frame loop (lifted from `spectatePage`):**
- `tick()` — advance `rot` for held rotate keys; if `!busy`, kick a new
  `Image()` against `/frame?...`; on load swap `img.src`, clear `busy`,
  `drawHud()`.
- `drawHud()` — render `#hud` text from `pos`, `rot`, `zoom`, `w`, `h`, plus a
  one-line hint (`left-click act · right-click menu · drag rotate · +/- zoom`).
- `refreshPos()` — `GET /pos`, store into `pos`.
- `flash(msg)` — transient toast in `#flash` (copy from `spectatePage`).
- `screenToFrame(e)` → `{px,py}|null` — the un-letterbox math (copy
  `spectatePage`'s click mapping); the single source of frame pixels for
  `/walk`, `/pick`.

**Picking + menu:**
- `pick(px, py, extra)` → `Promise<menuJSON>` — `POST /pick` with
  `{px,py,rot,zoom,w,h, ...extra}`.
- `act(ref, optionId, extra)` — `POST /act {ref, option:optionId, ...extra}`;
  on `{ok:false}` `flash(error)`.
- `openMenu(x, y, candidates)` — build `#menu` DOM: for each candidate, append a
  `.hdr` (when >1 candidate) then one `.opt` per option whose text is
  `verb + ' ' + candidate.label`; each `.opt`'s click handler captures
  `(candidate.ref, option.id)` and calls `act(...)` then `closeMenu()`. Position
  + clamp to viewport. Set `#menu.style.display='block'`.
- `closeMenu()` — hide `#menu`, clear its children.
- `defaultAct(menuJSON)` — used by left-click: if `candidates[0]` exists, `act`
  its `options[0]`; else (terrain only) `walk`.
- `walk(px, py)` — `POST /walk?px&py&rot&zoom&w&h` (reuse spectator endpoint) for
  the empty-ground case / terrain default.

**State (inventory + equipment + chat):**
- `pollState()` — guarded by `stateBusy`; `GET /state`; on success
  `applyState(json)`; always clears `stateBusy`.
- `applyState(s)` — calls `renderInventory(s.inventory)`,
  `renderEquipment(s.equipment)`, `appendChat(s.chat)`, and updates `stats` for
  the HUD.
- `renderInventory(inv)` — ensure 30 `.cell` nodes exist (built once in
  `initInventory()`); for each, set label/qty/`.wield`; stash
  `cell.dataset` = `{slot, itemId, defaultOptionId, options:JSON}` so the
  click/contextmenu handlers read them without a closure rebuild.
- `renderEquipment(eq)` — rebuild `#equip` rows (`slot` → `name`).
- `appendChat(lines)` — append only lines with `seq > lastChatSeq`; set colour
  class by `kind`; update `lastChatSeq`; if the log was scrolled to bottom,
  keep it pinned to bottom.

**Inventory interaction:**
- `initInventory()` — create the 30 cells once; attach delegated `click` and
  `contextmenu` listeners on `#inv`.
- inventory `click` handler → read `cell.dataset`; if occupied,
  `act({kind:'inventory_item', slot}, defaultOptionId)`.
- inventory `contextmenu` handler → `e.preventDefault()`; if the cell carries
  `options`, `openMenu(e.pageX, e.pageY, [{ref:{kind:'inventory_item',slot},
  label:name, options}])`; else fall back to
  `pick(0,0,{slot}).then(j=>openMenu(...))`.

**Chat input:**
- `initChat()` — `keydown` on `#chatinput`: on `Enter` call `sendChat(value)`,
  clear; on `Escape` blur.
- `sendChat(raw)` — trim; classify (`::`/`/` → command, `@name …` → pm, else
  say); `POST /chat` with the classified shape.

**Viewport input wiring (in an `init()` run at load):**
- `#v` `click` → `screenToFrame`; `pick(px,py)` → `defaultAct`.
- `#v` `contextmenu` → `e.preventDefault()`; `screenToFrame`; `pick(px,py)` →
  `openMenu(e.pageX, e.pageY, candidates)`.
- `document` `keydown`/`keyup` → rotate/zoom/view-size (copy `spectatePage`),
  **but ignore when `document.activeElement === chatinput`** so chat typing
  never drives the camera.
- `document` `click` (capture) and `scroll` and `Escape` → `closeMenu()`.
- timers: `setInterval(tick,33)`, `setInterval(pollState, 400)`,
  `setInterval(refreshPos,500)`, `setInterval(()=>anim=(anim+1)&0x3fffffff,160)`.
- initial: `initInventory(); initChat(); refreshPos(); pollState(); drawHud();`.

**Module-level state:** `let rot=64, zoom=%d, w=%d, h=%d, anim=0;`
`let busy=false, stateBusy=false; let pos={x:0,y:0,plane:0};`
`let lastChatSeq=-1; const keys={}; const ROT=4;`
`const HOVER=false, DRAG=false;` (stretch toggles).

### 5.5 Context-menu rendering detail

The menu is one reusable widget for both world picks and inventory right-clicks.
`openMenu(x,y,candidates)`:

```text
clear #menu
for ci, cand in candidates:
    if candidates.length > 1:
        append div.hdr text=cand.label
    for opt in cand.options:
        row = div.opt
        row.textContent = opt.verb + (candidates.length>1 ? "" : (" " + cand.label))
        row.onclick = () => { act(cand.ref, opt.id); closeMenu() }
        append row
place #menu at (clamp(x), clamp(y)); display:block
```

The very first option of the first candidate is the **default** — it is also
what left-click fires via `defaultAct`, so the menu's top row and left-click
always agree. The UI never sorts or filters options; ordering and verb text are
authored by Layer 2 and rendered as-is, preserving authentic RSC menu order
(action verbs first, `Examine`/`Walk here` last).

---

## 6. Acceptance (UI-level)

The page is correct for M1 when, against a live `cradle -client`:

1. Viewport streams + glides identically to `-spectate` (same frame contract).
2. Left-click on an NPC attacks/talks (its default); left-click on empty ground
   walks there.
3. Right-click anywhere shows a context menu of authentic verbs for every
   overlapping target, ordered nearest-first, `Examine`/`Walk here` last;
   clicking an option performs it (visible in the next frames/state).
4. Inventory grid reflects `/state` within ~½s of a change; left-click fires the
   item default; right-click shows the item's real verb menu; Drop/Wield work.
5. Equipment summary lists worn items.
6. Chat log shows public/npc/private/system lines with distinct colours and
   stays pinned to the newest; typing `hello`↵ says it, `::cmd`↵ runs a command,
   `@bob hi`↵ PMs bob.
7. Typing in chat never rotates/zooms the camera; `Escape` closes any open menu
   and blurs chat.
8. No request type stacks: frames are `busy`-gated, `/state` is `stateBusy`-
   gated, `/pos` is tiny.

Stretch (not required to pass M1): hover labels (`HOVER`), drag-to-use-on
(`DRAG`).
```
