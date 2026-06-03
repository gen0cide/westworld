# RuneScape Classic client — architecture

How the deobfuscated client (`rsclassic-1091943135.jar`, **RSC revision ~233–235**, a
**Microsoft J++ build**) fits together. 71 classes, ~50k lines of readable Java in `../src/`.
Class names below are the deobfuscated names; the obfuscated single/double-letter original is in
parentheses. See `NAMING.md` for the full obf↔name map and `MUDCLIENT_SKELETON.md` for the main
class's 484 fields / 123 methods.

## Entry & lifecycle

```
java.applet.Applet
  └─ GameShell (e)                 window, AWT listeners, double-buffer, main thread loop
       └─ Mudclient (client)       the whole game: net + world + scene + UI + state
GameFrame (qb)  extends java.awt.Frame   — host window when run as an application
LoaderThread (c)                         — background bootstrap: reflection, asset/cache load
```

`GameShell.run()` drives the loop: poll input → `Mudclient` per-tick update → render via `Surface`
→ sleep to the frame budget (`Timer`/`Utility`). Standalone launches build a `GameFrame`; the
applet path uses the browser via `JSBridge` (`a`, `netscape.javascript.JSObject`).

## Subsystems (by package)

### `client.scene` — rendering (the engine core; fully expanded, see note below)
- **Surface (ua)** — software 2D rasterizer over an `int[]` pixel buffer. Sprite blitting (opaque/
  alpha/indexed/scaled/tinted), boxes/lines/gradients/circles, the bitmap font drawer (`drawstring`
  family), the minimap rotate-blit (`drawMinimapSprite`), and `ImageProducer`/`ImageObserver` so the
  buffer can be handed to AWT. All blend math is fixed-point with `0xff00ff`/`0xff00` channel-lane masks.
- **Scene (lb)** — the 3D renderer. Holds `GameModel[]`, the camera, and does the painter's-algorithm
  visible-surface sort (`polygonsQSort` + the convex screen-overlap test `polygonsOverlap`), perspective
  projection (`projectModels`), near-plane clipping, and scanline fill (`rasterize`/`generateScanlines`,
  textured + gradient). Per-face data is held in **WorldEntity (w)** (normals, bounds, sort key).
- **GameModel (ca)** — a 3D mesh: vertices, faces, per-face colour/texture/lighting, and the
  transform/lighting pipeline (`applyRotation`, lighting, `project`).
- **SurfaceSprite (ba)** — `Surface` subclass that dispatches region draws back into `Mudclient`.
- Sprite/image IO: **SpriteDecoder (ea)**, **SpriteScaler (ia)**, **ImageLoader (pa)**,
  **SurfaceImageProducer (fb)**.

### `client.world` — terrain
- **World (k)** — the landscape: tile heights/elevation, walls/roofs/floors, the map-chunk loader
  (`loadMapData`/`decodeLandscapeChunk` reading `.hei/.dat/.loc/.jm`), and tile-based pathfinding (`route`).
  Builds `GameModel` meshes that Scene then renders.
- **GameCharacter (ta)** — a player/NPC: position, 10-step movement waypoints, 12-slot equipment,
  animation/combat/chat state.

### `client.net` — protocol (rev-235; oracle: OpenRSC Payload235)
- **Packet (b) / ClientStream (da)** — packet framing over a `Socket` with a background writer thread.
- **Buffer (tb) / BitBuffer (ja)** — big-endian get/put for every width + the RSC "smart" variable-length
  int, bit-level reads, and RSA `modPow` login-block encryption (rev-235 uses a 2-byte length prefix).
- **ISAAC (o)** — the ISAAC CSPRNG; one instance per direction seeds an opcode keystream so packet
  opcodes are scrambled on the wire.
- **ChatCipher (v)**, **TextEncoder (h)** / **StringCodec (u)** — chat (de)cipher and CP1252 text framing.
- **SocketFactory (m) / ProxySocketFactory (gb)**, **StreamBase (ib)**, **DownloadWorker (jb)** — connection
  setup (incl. proxy) and async resource fetch.

### `client.data` — cache & definitions
- **CacheUpdater (cb)** downloads + CRC-checks content; **ArchiveReader (ob)** reads JAG archives;
  **CacheFile (d)** is the on-disk `RandomAccessFile` store; **CachePath (r)** resolves the cache dir.
- **EntityDef (t)**, **NameTable (ub)**, **NameHash (oa)**, **DataStore (nb)**, **RecordLoader (f)**,
  **FontWidths (n)**, **BZip (aa)** (BZip2-style decompressor for archived assets).

### `client.ui` — interface
- **Panel (qa)** the widget/text container (4 instances drive the tabs), **MessageList (wb)** chat/social
  lists, **FontBuilder (s)** rasterizes AWT fonts to bitmaps, **CharTable (ga)** validates text input.

### `client.audio` — sound
- A filter-graph engine: **AudioChannel (sa)** voices, **AudioMixer (eb)** thread, **FilterChain (va)** →
  **StreamMixer (ra)** / **SoundDecoder (sb)**, **FilterStage (hb)**, **FilterNode (bb)** / **SampleBuffer (vb)**.
- Output backends behind **AudioOutput (ma)**: **DirectSoundPlayer (rb)** (MS DirectSound) and
  **SourceLinePlayer (pb)** (Java Sound `SourceDataLine`).

### `client.nativeapi` — Microsoft VM acceleration (optional, degrades to pure Java/AWT)
- **DirectDrawModes (wa)** (DirectDraw fullscreen enum), **Win32MouseCallback (q)** (`com.ms.dll.Callback`
  + `User32.SetCursorPos`), **DisplayModeSetter (ha)** (AWT `GraphicsDevice` fullscreen).

### `client.util` — plumbing
- **ErrorHandler (i)** + **ClientRuntimeException (la)** / **ClientIOException (fa)** (the obfuscator wrapped
  every method in a catch-rethrow that funnels here), **Timer (p)**, **Utility (mb)**, **ArrayUtil (ab)**,
  **LinkedQueue (db)** / **ListNode (g)**, **Globals (l)**, **IntHolder (ka)**, **StreamFactory (na)**,
  **JSBridge (a)**, plus the two `Runnable`/thread helpers.

## Data flow (one frame)
1. **Input** → `GameShell` listeners set mouse/key state.
2. **Net** → `ClientStream` drains server packets; `Mudclient`'s packetin dispatch (the big opcode switch,
   named from Payload235) mutates `GameCharacter[]`, inventory/skills, ground items, the `World` map.
3. **Update** → `Mudclient` advances animation/movement, runs `World.route` for click-to-walk.
4. **Build** → changed regions of `World` rebuild `GameModel` meshes.
5. **Render** → `Scene` projects + painter-sorts + rasterizes models into `Surface`; `Mudclient` draws the
   2D UI (`Panel`/`MessageList`/minimap) over the top; `Surface` hands the buffer to AWT.
6. **Outgoing** → input actions build packets (packetout) encrypted via `ISAAC`/`Buffer` and queued on
   `ClientStream`'s writer thread.

## Obfuscation (stripped during deobfuscation)
The J++ obfuscator applied, per method: an opaque predicate `boolean bl = client.vh` (always false) gating
dead branches; a profiling `++counter`; a `try{…}catch(RuntimeException e){throw ErrorHandler.a(e,"sig")}`
wrapper; anti-tamper `if(param != magic) return;` guards with dummy params; and junk-masked bit-shifts
(`x >> 1599703024` ≡ `x >> 16`). It also injected **bogus overlapping exception-table ranges** (dead
`athrow` handlers) that broke Vineflower — see the pipeline note in `../README.md`; the
`strip-obf-exceptions.jar` ASM pass fixes it (238→4 method-decompile failures).

## Provenance / confidence
Names and logic were verified against three oracles: **OpenRSC Payload235** (exact rev-235 protocol
opcodes), **LeadingBot `mudclient.java`** (readable deob), and **mudclient204** (rev-204 named client).
All bodies were re-audited against a clean, complete decompilation (the exception-strip fix recovered 238
previously-undecompilable methods); ~236 logic bugs from the first (defective-base) pass were fixed. The
rendering classes were given a dedicated full-expansion pass and verified at 100% method presence with no
summarized bodies. Residual per-class uncertainties are noted in the deobfuscated files' own comments.
This is a comprehension reference, not a recompilable build (it needs long-dead Microsoft J++ stub classes).
