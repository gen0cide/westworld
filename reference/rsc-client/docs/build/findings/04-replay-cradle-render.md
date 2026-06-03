> _Verbatim investigation report from the `rsc-client-build-options` workflow (2026-06-01). One of five parallel agents; see `../OPTIONS.md` for the synthesis._

## A4 — Replay, cradle, and render-library integration

### TL;DR (the load-bearing finding)

The westworld **cradle is a fully native, standalone Go RSC client**. It implements the RSC v235 wire protocol itself (`proto/v235/`), maintains its own `world.World`, and "sees" by rendering that world state through the Go `render/` library — which is a **line-by-line port of OUR deob `Scene.java` / `World.java`** (cross-referenced by exact file+line). **The cradle never touches rscplus, the obfuscated jar, or replay files.** rscplus and its replay subsystem are a *separate* Java track. So the two halves of the task answer to two different worlds, and the integration question between them is real and currently **unbuilt**. Concretely:

- **G3** (replays → cradle tester): there is **no existing bridge**. rscplus replays are raw ISAAC-encrypted Java-side packet logs played back over a loopback socket *into the vanilla Java client*; the cradle speaks Go and reads live server state. Bridging them means either (a) decoding replay packet bytes in Go via `proto/v235`, or (b) the simpler/native path the cradle already uses — drive scenarios live against the OpenRSC server. Replays favor *keeping rscplus* only if human-watchable visual playback is the goal; they are **not** needed for the cradle to perceive scenarios.
- **G4** (render audit): the most useful artifact is unambiguously **OUR named source** — specifically `scene/Scene.java`, `world/World.java`, `scene/GameModel.java`, `scene/Surface.java`. The Go `render/` library *already* cites these by line number. A recompiled/running client would be a nice diff oracle but is **not** required; the named source is the ground truth and is already being used as such.

---

### 1. rscplus replay subsystem

Two cooperating Java packages, both wrapping the same obfuscated jar:

**Capture (`src/Game/Replay.java`, hooked via `JClassPatcher`).** rscplus injects two callsites into the obfuscated client's raw socket I/O (`JClassPatcher.java:5138` → `Game/Replay.dumpRawInputStream([BIIII)V`, and `:5166` → `dumpRawOutputStream([BII)V`). These fire on every raw read/write **before** the client's own ISAAC decryption mutates the buffer (`Replay.java:1227` clones the bytes "since it gets modified by decryption in game logic"). So a replay records the **raw, still-ISAAC-encrypted server↔client packet stream**, framed as `[int32 timestamp][int32 length][bytes]` per burst, with `timestamp = -1` (`TIMESTAMP_EOF`) marking disconnects. It is packet-level, **not** frame/pixel capture.

**On-disk format** (a directory per replay, `ReplayEditor.importData`, `Replay.java:428-437`):
- `in.bin.gz` — gzipped incoming (server→client) timestamped packets
- `out.bin.gz` — gzipped outgoing (client→server) packets
- `keys.bin` — the per-session ISAAC keys (4 ints) needed to decrypt opcodes
- `version.bin` — `ReplayVersion` (current `VERSION=5`; `clientVersion` must be 235 for `authenticReplay()`)
- `metadata.bin` — length, dateModified, sanitization flags, optional server-extension id
- Keyboard/mouse input is also captured (`KEYBOARD_*` / `MOUSE_*` constants) for UI-faithful playback.

This is the **rscminus** format (the scraper package literally carries the rscminus license header), so it interoperates with the broader RSCPlus replay-archive tooling.

**Playback (`src/Game/ReplayServer.java`, `implements Runnable`).** This is the clever part: playback is **fully offline** — there is no live server. `ReplayServer` opens a **local loopback `ServerSocketChannel`** (`ReplayServer.java:160`), the patched client connects to it (the patcher redirects `connection_port`, `JClassPatcher.java:352`), and the server thread **streams the recorded `in.bin.gz` packets back to the real Java client** paced by a frame timer (`frame_timer += getFrameTimeSlice()`, `:421/630`; `Thread.sleep` to match the original ~50fps cadence). The vanilla client thinks it's talking to Jagex and renders normally; `out.bin.gz` is used to satisfy the client's outgoing handshake/keepalive expectations. Seeking, pause, and speed-multiplier (`fpsPlayMultiplier`) are supported. The scraper-side `ReplayReader` / `PacketBuilder` can also parse packets *without* the client, for analysis/sanitization.

**Could it be driven programmatically?** Partly. Playback is deterministic and headless-ish at the socket layer, but it still **requires the vanilla Java client (the jar) to be running and rendering** to interpret the bytes — the replay file alone is just encrypted opcodes + the ISAAC keys. A Go consumer would have to re-implement decryption + opcode decoding (which westworld *already has* in `proto/v235`, see §2/§4).

---

### 2. westworld cradle + scenarios

**The cradle is a standalone single-bot Go RSC client** (`cmd/cradle/main.go`). It talks **directly to the OpenRSC server** over TCP — not to rscplus, not via the jar:

- Native protocol stack: `proto/v235/` implements framing, the **ISAAC** cipher, **RSA** login, and per-opcode encoders/decoders ("bit-for-bit wire compatibility with the OpenRSC server," every fn citing `server/src/com/openrsc/server/net/`).
- `session/` (`conn.go`, `handshake.go`) wraps the socket; `runtime/host.go`'s `Host.Connect` does `session.Dial(... h.opts.Server ...)` then `conn.Login(...)`.
- `Host.handleFrame` decodes inbound packets and applies them to an in-process `world.World` (`h.world.Apply(...)` for own-player coords, NPC/player updates, scenery, ground items). **That `world.World` IS the cradle's perception of the game** — it is built from decoded server packets, exactly like a real client.

**Scenarios/routines** are `.routine` files written in westworld's own DSL (`runtime "1.0"; routine name() { ... }`), parsed and executed by `dsl/interp`. They use high-level verbs (`walk_to`, `open_boundary`, `command(...)`, `say`, `world.boundaries.at(x,y,dir)`), `require`/`abort` predicates, and event waits (`select { on boundary_changed(...) ... timeout 8s ... }`). `examples/routines/` are building blocks; `examples/scenarios/{combat,edges,...}/` are the ~196 integration tests, each annotated with category/hosts/admin/timeout/wiki-grounding headers. They run via `cradle -routine <file>` (see `scripts/dev-launch.sh`, which sources secrets and execs the cradle). "Drones" are named admin/test accounts (alex, delores, bernard) the cradle logs in as; `-reset-on-exit` cleans drone state between scenarios.

**Where the cradle perceives the world:** entirely from its own decoded `world.World`. It does **not** embed the Java client, does **not** transpile rendering, and does **not** talk to rscplus. For *visual* perception it feeds `world.World` into the Go `render/` library: `cmd/cradle/spectate.go:buildLiveView()` reads `host.World().Npcs/Players/Scenery/GroundItems/Boundaries` and packs a `render.View`, then `render.RenderView(land, facts, bundle, v)` produces a PNG (used by `-render-view`, `-spectate` browser viewport, and the `SnapshotFromCradle → RenderView` path the LLM tester reads from `/tmp/render_out/`).

---

### 3. westworld render library and the deob ground truth

`render/` is a **pure-Go 3D renderer ported from the RSC client**. Lineage per the MEMORY note + code: originally from eggsampler's deob (v204) and OpenRSC's `Client_Base`, **now actively re-grounded against OUR deob** — the comments cite `Scene.java`, `World.java`, `GameModel.java`, `Surface.java`, `mudclient.java` by exact line number throughout. It builds the same scene a real client would: terrain heightmap, walls/fences/doors (boundaries), roofs, scenery models, NPC/player billboards, ground-item sprites; then projects, painter-sorts, and rasterizes with the client's scanline filler.

**The colleague's perception concerns map directly to ported deob methods:**

| Concern (perception bug) | Go file | Ported from OUR deob |
|---|---|---|
| **Walls / doors** (can't walk through a door the cradle can't perceive) | `render/boundary.go` `BuildBoundaries` | `World.method422` (`World.java:740`), wall-edge data read off `Tile.HorizontalWall/VerticalWall/DiagonalWalls` exactly like `loadSection` (`World.java:1006-1060`) |
| **Roofs / under-roof culling** | `render/roof.go` | `World.method428` wall-top bump (`World.java:725/1050-1060`) + `mudclient` under-roof 0x80 cull |
| **Bridges** ("all messed up" bridge) | `render/terrain.go` (bridgeOverlay 250 remap) | `World.loadSection` bridge pre-pass + `getTerrainHeight` (`World.java:188, 839-996`) |
| **Floors / water / terrain mesh** | `render/terrain.go` `BuildTerrain` | `World.loadSection` / `getTerrainHeight` (vertices at `x*128, -height, y*128`) |
| **TRANSPARENT collision-only faces** (Lumbridge arch "blank") | `render/roof.go`/`boundary.go` `sceneTransparent=12345678` | `Scene.TRANSPARENT` skip (`Scene.java:2666`) |
| **Projection / face sort / fill** | `render/model.go`, `render/scene.go`, `render/rasterize.go` | `Scene.method293` normals (`Scene.java:2326-2363`), `GameModel`, `Surface` |

**Ground-truth files the Go lib must match (in `reference/rsc-client/src/client/`):**
- `world/World.java` — terrain heightmap, boundary/wall meshing (`method422/428`), bridge/overlay remaps, `loadSection`. **This is the authority for floors/walls/bridges/roof bases.**
- `scene/Scene.java` — projection, frustum/clip, face normals & sorting, `TRANSPARENT` handling, scanline fill. **Authority for what is drawn vs occluded.**
- `scene/GameModel.java` — vertex/face model, lighting/normals.
- `scene/Surface.java` — the 2D raster target / span fillers.

Note a subtlety relevant to G4: the deob is documented to have had a **World/Scene swap resolved** during correctness audit (commit `960310c`). The Go port's line citations should be re-validated against the *current* `World.java`/`Scene.java` to ensure they didn't pin to pre-swap line numbers.

---

### 4. Integration conclusions

**G3 — cradle tester "seeing" scenarios:**

The cradle **already perceives scenarios live** without rscplus: it connects directly to OpenRSC, decodes packets, and renders its own `world.World` to PNGs for a Claude-Code tester to inspect (the `SnapshotFromCradle → RenderView → /tmp/render_out` path; `-spectate` for a live browser viewport). So **rscplus replays are not required for the cradle to see scenarios.**

What rscplus replays *add* is a **deterministic, offline, recorded** scenario stream — valuable if you want a Claude tester to evaluate a *captured human/bot session* repeatedly without a live server. To wire replays into the cradle you'd need a Go-side replay consumer:
- Decode `keys.bin` → ISAAC, then feed `in.bin.gz` packet bursts through `proto/v235/inbound.go` to drive `world.World.Apply(...)` (the cradle's existing decode path). westworld already has the ISAAC + opcode decoders, so this is feasible **without** running any Java.
- The simpler, already-supported alternative is **scripted live scenarios** (the current `.routine` model): deterministic enough for integration testing, and it exercises the *real* server (matching G3's "drone accounts doing integration tests against the server").

**Verdict for G3:** This **does not favor keeping rscplus** for the cradle's perception. The cradle is self-sufficient in Go. rscplus is worth keeping for what it's uniquely good at — **human-watchable visual replay capture/playback against the vanilla client, plus its mature replay archive format** — and as a *capture source* you could later decode in Go. If a replay→cradle bridge is wanted, build a small Go replay reader on top of `proto/v235`; do **not** try to pipe the cradle through rscplus's loopback ReplayServer (that path only feeds the Java client).

**G4 — render audit artifact:**

The most useful artifact is **OUR named deob source**, specifically the four files above — and it's already the chosen oracle (the Go port cites them by line). Recommended audit form:
1. Use `world/World.java` + `scene/Scene.java` as the **textual ground truth**; gap-analyze each Go function against the cited method (re-verifying line numbers post World/Scene-swap).
2. A **doc map** (extend `NAMING.md` / a new `render-parity.md`) listing each `render/*.go` function ↔ deob method ↔ "ported / partial / TODO" status would make the audit tractable; the Go comments already contain most of this prose informally.
3. A **recompiled+running client to diff against** would be the gold standard for pixel/behavioral parity (render the same scene in both and diff), **but the deob is explicitly a comprehension reference, not buildable** (dead J++ native stubs). Getting it to the pure-Java/AWT path it's documented to degrade to would be a sizable separate effort. It is a *nice-to-have* oracle, **not** the critical artifact — the named source already resolves the colleague's specific door/bridge/roof perception questions.

**Language boundary summary:** rscplus + jar + replays = Java (loopback-socket, ISAAC-encrypted packet logs, feeds the vanilla client). Cradle + render + proto = Go (native protocol, own world model, own renderer). They share **only the protocol revision (235) and the deob as a shared specification** — not code or runtime. The deob serves G3 as the *protocol/world spec* the Go `proto/v235` and `world` packages were ported from, and serves G4 as the *render spec* the Go `render` package was ported from. rscplus is orthogonal to both except as an optional replay-capture feeder.
