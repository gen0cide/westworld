# RSC Client (rev ~233–235, Microsoft J++ build) — class naming MAP (VERIFIED)

Source jar: `rscplus/assets/rsclassic-1091943135.jar` — 71 obfuscated classes, default package, signed `Created-By: Jagex`.
Oracle: `mudclient204/src/*.java` (deobfuscated RSC rev 204, pure Java).

This is a LATER revision than the oracle. Extra Microsoft VM classes pull in `com.ms.directX` (DirectDraw/DirectSound),
`com.ms.dll.Callback`, `com.ms.win32.User32`, `com.ms.awt.WComponentPeer`, `netscape.javascript.JSObject`. The whole
client is also wrapped in heavy control-flow obfuscation (opaque predicate `client.vh`, per-method profiling counters,
per-method try/catch that rethrows via `i.a(e,"sig(...)")`, and per-class XOR-encrypted string pools decoded by two
`z(...)` helpers).

Names were verified by reading the normalized Vineflower output, the CFR cross-ref, decoding each class's XOR string
pool, and matching behavior to the oracle. Where the obfuscated build's structure differs from the oracle (e.g. the
audio engine, the socket/proxy factory, the cache/CRC downloader, DirectX paths), the name reflects the actual code.

> **CORRECTION (post-deob):** the deobfuscation agents found `k` and `lb` were swapped in the original map.
> Truth (verified from method bodies): **`k` = World** (terrain/landscape, `client.world`), **`lb` = Scene** (3D renderer, `client.scene`).
> The two giant files were written under the swapped identities (`scene/Scene.java` holds World logic, `world/World.java` holds Scene logic)
> and still need to be swapped + cross-references reconciled in the consistency-review pass.

## Final map

| obf | name | package | role | oracle / evidence |
|-----|------|---------|------|-------------------|
| e | GameShell | client.shell | Applet shell: window, Thread main loop, mouse/key listeners, getParameter, paint | `extends Applet implements Runnable,Mouse/Key…`; matches oracle `GameShell` |
| client | Mudclient | client | Main game class; net + scene + world + UI + game state; `extends e`; holds opaque `vh` | `extends e`; oracle `mudclient`/`GameConnection` merged here |
| qb | GameFrame | client.shell | AWT `Frame` host window; holds `e g`; routes MouseEvents | `extends Frame`; matches oracle `GameFrame` |
| j | RobotCursor | client.shell | java.awt.Robot mouse/custom-cursor controller (`showcursor`/`movemouse`/`setcustomcursor`) | uses `Robot`, `createCustomCursor` |
| kb | InputState | client.shell | static input/framebuffer holder: `qb a`, key/pixel arrays | holds `qb` + `byte[]`/`int[]` used by e/client/o |
| c | LoaderThread | client.shell | init/loader Thread: reflection bootstrap, `EventQueue`/`Toolkit`, owns `q` mouse callback, `wa`, `d[]` | `implements Runnable`; AWT/reflection |
| jb | DownloadWorker | client.net | download/read worker Thread: `DataInputStream` → fills a `tb` buffer | `implements Runnable`; `q.read` into `tb.F` |
| ua | Surface | client.scene | software 2D renderer (pixel buffer, sprites, color math, drawString); `ImageProducer/ImageObserver` | matches oracle `Surface` |
| ba | SurfaceSprite | client.scene | `extends ua`; holds `client dc`; dispatches region draws (minimap/UI) back into `client` | matches oracle `SurfaceSprite` |
| k | World | client.world | landscape/terrain: tiles, elevation, walls/roofs, map load (`loadMapData`/`decodeLandscapeChunk`), `route`/pathfinding | oracle `World` — **CORRECTED** (map originally mislabeled `k` as Scene); verified by `getElevation`/`getWall*`/terrain methods in the deobbed file |
| ca | GameModel | client.scene | 3D model (vertices/faces/lighting/transforms) | matches oracle `GameModel` |
| ea | SpriteDecoder | client.scene | decodes sprite/image stream into `ua.Mb` pixel array via `ac` byte stream | writes Surface pixels; `c(ac)`/`d(ac)` byte reads |
| pa | ImageLoader | client.scene | decodes indexed BMP bytes → AWT `Image` (`IndexColorModel`, palette @ off 20) | builds `Image` from byte[] palette |
| ia | SpriteScaler | client.scene | scanline sprite/texture scaler: writes scaled pixels from palette with fractional step | tight pixel-doubling loop into `int[]` |
| fb | SurfaceImageProducer | client.scene | `ImageProducer/ImageObserver` feeding `da`/Surface; holds `aa` bzip ref | `addConsumer/imageUpdate`; `static aa a` |
| lb | Scene | client.scene | 3D scene renderer: `ca` (GameModel) arrays, camera, polygon painter-sort (`polygonsQSort`/`polygonsOverlap`), `render`/`rasterize`/`projectModels` | oracle `Scene` — **CORRECTED** (map originally mislabeled `lb` as World); verified by render/raster methods. NOTE: `w` (currently WorldEntity) is likely this Scene's per-polygon entry (Polygon) — flagged for review |
| w | WorldEntity | client.world | scene/wall/object holder: many int coords + `ca o` (GameModel) | used as `w[]` by `lb` |
| ta | GameCharacter | client.world | player/npc entity: `int[12]` equip, `int[10]` waypoints, name/message, anim/combat | matches oracle `GameCharacter` |
| b | Packet | client.net | base network packet/stream: two ISAAC ciphers (`o z`,`o j`), `ja f`, read framing | matches oracle `Packet` |
| da | ClientStream | client.net | `extends b`; Socket + writer Thread (`run`), read/write stream bytes | matches oracle `ClientStream` |
| tb | Buffer | client.net | `extends ib`; cursor read/write (putByte/getInt/getLong/getString) + RSA `modPow` encrypt | matches oracle `Buffer` |
| ja | BitBuffer | client.net | `extends tb`; bit-level reader; holds RSA modulus `BigInteger K` | bit reads over `tb.F` |
| o | ISAAC | client.net | ISAAC stream cipher (golden-ratio const `-1640531527`, keystream wheel) | matches oracle `ISAAC` |
| m | SocketFactory | client.net | abstract socket factory; `abstract Socket a(byte)`, `new Socket(host,port)` | `abstract Socket a(...)` |
| gb | ProxySocketFactory | client.net | `extends m`; proxy-aware connect (`ProxySelector`,`InetSocketAddress`,reflection) | proxy `connect()` path |
| u | StringCodec | client.net | writes length-prefixed strings into a `tb` buffer via `h`+`aa`; id registry lookup | `a(int,tb,String)` packet write |
| h | TextEncoder | client.net | `CharSequence → byte[]` encoder using char-map tables (`int[] b`,`int[] c`) | `a(CharSequence,byte):byte[]` |
| v | ChatCipher | client.net | byte (de)cipher helper (`a(byte[],int,int,int):byte[]`); reused as small string holder `v(...)` | XTEA/char (de)cipher |
| s | FontBuilder | client.ui | rasterizes AWT `Font` glyphs to bitmaps (`Font`,`FontMetrics`,`e`) | `a(int,Font,…,FontMetrics,…)` |
| qa | Panel | client.ui | UI panel/widget container & text rendering (Helvetica font); 4 instances in `client` | `new qa` ×4 → 7 client fields |
| wb | MessageList | client.ui | multi-string list/message panel (`a(String,String,String,int,String,byte)`); 3 instances | `new wb` ×3 (chat/social tabs) |
| ga | CharTable | client.ui | accented/special-character table + `a(char,int)` validation for text input | `char[]{à,á,â,…}` table |
| nb | DataStore | client.data | in-memory archive/def store; URL/InputStream load, `a(long,int)` record access (`b.q`) | def/record handler used by Packet |
| ob | ArchiveReader | client.data | JAG archive entry reader: `a(int,int,int,byte[])`, `a(int,int):byte[]`, toString | archive extraction |
| cb | CacheUpdater | client.data | content/CRC downloader: "contentcrcs", "Checking for new content", "Invalid CRC…"; builds `ob`+`nb` | CRC verify + download |
| d | CacheFile | client.data | on-disk `RandomAccessFile` cache (`.file_store`), `finalize`, record read/write | RandomAccessFile store |
| r | CachePath | client.data | resolves cache dir/file ("user.home","/rscache/",".jagex_cache_","c:/windows/") | File path resolver |
| t | EntityDef | client.data | definition record/loader (name+examine strings + ints; decode `a(byte[],…)`) | item/object def record |
| ub | NameTable | client.data | name registry: `String[5000]` (model names) + parallel-sort helpers | `String[5000]`, read by scene/model |
| oa | NameHash | client.data | name→id hash lookup over `d[][]` (uppercase `61*h+ch`); reads cache `d` | hashed name lookup |
| f | RecordLoader | client.data | loads a 24-byte cache record from `nb` (`b.q`) into a `tb` packet | `b.q.a(...)` → `tb` |
| n | FontWidths | client.data | char-width table `a[256]=9*idx` from a glyph string; `static nb h` | charset width lookup |
| aa | BZip | client.data | BZip2-style decompressor `a(int,byte[],int,byte[],int,int)` | matches oracle `BZLib` |
| ac | DecodeBuffer | client.util | byte/bit decode buffer (`byte[] j/q/i/d/s`,`boolean[] n`) + list-insert + chat-markup chars `[ ] #` | used by `ea` decode; `ac.a(ib,byte,ib)` |
| rb | DirectSoundPlayer | client.nativeapi | DirectSound output (`DSBufferDesc`,`DSCursors`,`DirectSound`,`WaveFormatEx`); `implements ma` | `com.ms.directX.*` |
| pb | SourceLinePlayer | client.audio | Java Sound output (`SourceDataLine`,`AudioFormat`); `extends sa` | `LineUnavailableException` |
| sa | AudioChannel | client.audio | abstract audio voice/stream base; owns `va` filter chain; built by `eb` mixer | `va`/`bb` chain, `eb`/`c` refs |
| eb | AudioMixer | client.audio | mixer Thread: iterates `sa[] f` voices | `implements Runnable`, `sa[] f` |
| ma | AudioOutput | client.audio | marker interface for audio output backend (implemented by `rb`) | empty interface |
| ib | StreamBase | client.net | base resource/stream class (static URL loader); superclass of bb/hb/va/tb | `class ib` base |
| bb | FilterNode | client.audio | abstract audio filter node `extends ib` (`int g`) | base of `vb` |
| vb | SampleBuffer | client.audio | `extends bb`; raw sample bytes (`byte[] l`,`int i/h/k`) | constructed with byte[] |
| hb | FilterStage | client.audio | abstract filter stage `extends ib` (`abstract int a(ra)`, `abstract void a()`) | scheduled by `ra` |
| va | FilterChain | client.audio | abstract audio filter chain `extends ib` (linked `va j`,`bb h`; mix/skip) | base of ra/sb |
| ra | StreamMixer | client.audio | `extends va`; schedules `hb` stages over a `db` queue; mixes child `va`s | `new ra` in client |
| sb | SoundDecoder | client.audio | `extends va`; large bit-decode tables (sound/voice decode into samples) | huffman-ish decode of `vb` |
| db | LinkedQueue | client.util | doubly-linked list/queue over `ib` nodes (`a(ib,bool)` add, `a/b(byte)` iterate) | iterator state |
| g | ListNode | client.util | generic linked node/cache entry (`g a`, ints, `Object f/d`); used everywhere | hashtable node |
| ka | IntHolder | client.util | small holder (`int[] c`,`int b`,`String a`) | shared scratch |
| l | Globals | client.util | global holder: `Applet b`, `String[] a`, `String[100] c` (params/strings) | used by every class |
| i | ErrorHandler | client.util | wraps Throwable+context: `a(Throwable,String):la`; toString | per-method catch sink |
| la | ClientRuntimeException | client.util | `extends RuntimeException`; carries `Throwable e`,`String h` | thrown by `i.a` |
| fa | ClientIOException | client.util | `extends IOException`; checked exception wrapper | `extends IOException` |
| mb | Utility | client.util | timing (`a(int,long)` sleep/profile), error report ("clienterror.ws?c="), chat formatting | matches oracle `Utility` |
| p | Timer | client.util | `System.currentTimeMillis()` timing (synchronized) | time helper |
| ab | ArrayUtil | client.util | array helpers: `a(int[],int,int)`, `a(byte[],int,byte[],int,int)` (fill/copy) | tiny static |
| na | StreamFactory | client.util | factory: `a(String,int,byte[],int):byte[]` (→`t`), `a(int,int,String):m` (→`gb`) | builds codecs/sockets |
| a | JSBridge | client.util | `netscape.javascript.JSObject.getWindow(applet).call(...)` browser bridge | `JSObject` call |
| ha | DisplayModeSetter | client.nativeapi | fullscreen via AWT `GraphicsDevice.setDisplayMode(DisplayMode)` | `DisplayMode`/`GraphicsDevice` |
| wa | DirectDrawModes | client.nativeapi | DirectDraw display-mode enumerator + fullscreen setup; `IEnumModesCallback` | `com.ms.directX.DirectDraw` |
| q | Win32MouseCallback | client.nativeapi | `extends com.ms.dll.Callback`; `User32.SetCursorPos`, window-long hooks | `com.ms.win32.User32` |

All 71 obfuscated classes (`a`..`w`, `aa`..`wa`, `ab`,`ac`,`bb`..`wb`, `client`) appear exactly once above.

## Inheritance hierarchy (verified from declarations)

```
java.applet.Applet
  └─ e   [GameShell]   (Runnable, Mouse/MouseMotion/KeyListener)
       └─ client  [Mudclient]   — MAIN; net (da) + scene (lb=Scene) + world (k=World) + UI (qa/wb)

java.awt.Frame
  └─ qb  [GameFrame]

b   [Packet]          (ISAAC o z / o j, ja f, nb q)
  └─ da [ClientStream] (Runnable; Socket + writer thread)

ib  [StreamBase]      (static URL loader)
  ├─ tb [Buffer]      (cursor read/write + RSA encrypt)
  │    └─ ja [BitBuffer]  (bit reader, RSA modulus K)
  ├─ bb [FilterNode] (abstract)
  │    └─ vb [SampleBuffer]
  ├─ hb [FilterStage] (abstract)
  └─ va [FilterChain] (abstract)
       ├─ ra [StreamMixer]
       └─ sb [SoundDecoder]

m   [SocketFactory] (abstract; abstract Socket a(byte))
  └─ gb [ProxySocketFactory]

sa  [AudioChannel]
  └─ pb [SourceLinePlayer]   (SourceDataLine)

ua  [Surface]  (ImageProducer, ImageObserver)
  └─ ba [SurfaceSprite]      (holds client dc)

java.io.IOException     ─ fa [ClientIOException]
java.lang.RuntimeException ─ la [ClientRuntimeException]
com.ms.dll.Callback     ─ q  [Win32MouseCallback]

interfaces:
  ma [AudioOutput]            implemented by rb [DirectSoundPlayer]
  IEnumModesCallback          implemented by wa [DirectDrawModes]
  ImageProducer/ImageObserver implemented by ua [Surface], fb [SurfaceImageProducer]
  Runnable                    implemented by e, da, c [loader thread], eb [AudioMixer], jb [download worker]
```

## Obfuscation patterns to strip when reading

- Opaque predicate `boolean bl = client.vh;` (always false) gating dead `if/while/break` — remove.
- Per-method profiling `++<staticCounter>;` near entry — remove.
- Exception wrapper `catch (RuntimeException e){ throw i.a(e,"sig("+args+")"); }` — unwrap to bare body.
- Per-class XOR string pools: `z(z("…"))` → decode with the two `z` helpers (1-char key + mod-5 byte table).
- Anti-tamper guards `if (n != <magic>) return;` / junk modulo expressions (`109 % ((var0+8)/63)`) — dead, remove.
