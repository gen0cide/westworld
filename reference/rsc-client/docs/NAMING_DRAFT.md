# RSC Client (rev ~233–235, Microsoft J++ build) — class naming DRAFT

Source jar: `rscplus/assets/rsclassic-1091943135.jar` — 71 obfuscated classes, default package, signed `Created-By: Jagex`.
Oracle: `mudclient204/src/*.java` (deobfuscated RSC rev 204, pure Java). This jar is a LATER revision with extra
Microsoft VM classes (`com.ms.directX`, `com.ms.dll.Callback`, `com.ms.win32.User32`, `netscape.javascript`).

This is a DRAFT hypothesis to be VERIFIED/CORRECTED against the actual normalized source + oracle.
Line counts in earlier analysis were unreliable — ignore them; re-derive from the real files if needed.

## Confirmed inheritance hierarchy (from javap — high confidence)
```
java.applet.Applet
  └─ e   [GameShell]   (implements Runnable, Mouse/MouseMotion/KeyListener)
       └─ client  [mudclient / GameApplet]  — MAIN game class; holds socket `da` + scene `k`
b   [Packet/base stream]
  └─ da  [ClientStream]  (implements Runnable; Socket + In/Out streams)
ib  [base: resource/stream/buffer base]
  ├─ bb (abstract) ─ vb
  ├─ hb (abstract)
  ├─ va (abstract) ─ ra, sb        [render/filter chain]
  └─ tb ─ ja                       [buffered input stream]
m   (abstract) ─ gb                [codec/compression base]
sa  ─ pb                           [UI component base]
ua  [Surface] (implements ImageProducer, ImageObserver) — 2D renderer; non-final base, subclassed by ba
java.awt.Frame ─ qb  [GameFrame]
java.io.IOException ─ fa           [custom checked exception wrapper]
java.lang.RuntimeException ─ la    [custom runtime exception wrapper]
com.ms.dll.Callback ─ q            [Win32 input callback]
```

## High-confidence class names (verify members against oracle)
| obf | proposed | oracle | role |
|-----|----------|--------|------|
| e | GameShell | GameShell | Applet shell: window, input listeners, main loop scaffolding |
| client | mudclient | mudclient | main game class; net + scene + game state + UI |
| ua | Surface | Surface | software 2D renderer (pixel buffer, sprites, color math) |
| ba | SurfaceSprite | SurfaceSprite | sprite-drawing extension of Surface |
| k | Scene | Scene | 3D scene/terrain: arrays of GameModel (`ca[][]`) |
| ca | GameModel | GameModel | 3D model (vertices/faces) |
| lb | World | World | world/entity/model/sprite manager |
| b | Packet | Packet | base packet/stream buffer |
| da | ClientStream | ClientStream | socket connection + network thread |
| n | GameCharacter | GameCharacter | player/npc entity (position/anim state) |
| nb | EntityHandler | EntityHandler | entity/def registry |
| o | ISAAC | ISAAC | ISAAC cipher (keystream wheel) |
| v | (de)cipher | Buffer(crypto) | RSA/XTEA-ish byte (de)cipher helper |
| h | CharFilter/Codec | — | char→byte encoding helper |
| qa | DataTables | — | large static lookup/charset tables |
| qb | GameFrame | GameFrame | AWT Frame host window |
| rb | StreamAudioPlayer/DirectSound | StreamAudioPlayer | DirectSound audio output (implements iface `ma`) |
| ma | (audio interface) | — | marker interface implemented by `rb` |
| wa | DirectXVideoModes | — | DirectX video-mode enum callback |
| q | Win32InputCallback | — | com.ms.dll.Callback (native input) |
| fa | ClientIOException | — | checked exception wrapper |
| la | ClientRuntimeException | — | runtime exception wrapper |
| i | ErrorHandler | — | wraps Throwable+context (used by the per-method try/catch obfuscation) |
| p | Timer/GameTime | — | timing helper (System.currentTimeMillis) |

## Medium/low confidence — RE-DERIVE from source
ab, ac (huffman/bzip codec?), c (loader thread), cb (GameData?), d (cache file), aa, mb (defs/data),
db, ea, f, ga, ha, ia, oa, pa, s, t, u, ub, na, ta, r, sa, tb, ja, va, ra, sb, vb, gb, m, ob, w, j, l,
ka, kb, g, jb, eb. Many small ones are data holders / utility / linked-list nodes / constant tables.

## Obfuscation patterns to STRIP for readability (present in nearly every method)
- Opaque predicate: `boolean bl = client.vh;` then scattered `if (bl) return;` / `while(!bl)` — `vh` is always false; this is dead control flow. Remove it.
- Per-method profiling: `++<staticCounter>;` near method entry — remove.
- Exception wrapper: `try { ... } catch (RuntimeException e) { throw i.a(e, "sig(" + args + ")"); }` — the catch only adds a stack-trace string; unwrap to the bare body.
- Junk bit math: masks with extra bits (`& 0xFF8DC6`) before a shift that discards them, and shift amounts already normalized mod 32 (done). Reduce `(x & 0xFF8DC6) >> 16` → `(x >> 16) & 0xFF` where equivalent.
- Dummy parameters / dead branches guarded by constant comparisons (`if (n3 != -1057205208) return;`) — these are anti-tamper guards; simplify.
