> _Verbatim render-completeness audit from the render-deob-completeness-audit workflow (2026-06-01). See ../../RENDER_DEOB_GAPS.md for the consolidated plan._

# READ-ONLY Render-Completeness Audit — Surface + sprite/raster helpers (2D rasterizer)

Area: `src/client/scene/{Surface, SurfaceSprite, SpriteDecoder, SpriteScaler, ImageLoader, SurfaceImageProducer, Scanline}.java`. Ground truth: `decompiled/normalized-clean/{ua, ba, ea, ia, pa, fb, n}.java`.

Headline: this is the strongest-transcribed render area I would expect to find. Six of seven files are fully deobfuscated, fully expanded, and behaviorally faithful. The gaps are concentrated in two cross-class identity collisions that the obfuscator created around the single-letter name `k` and the merged class `n`, plus the (already-known) Scanline reconstruction question. One of these is a genuine HIGH-severity behavioral break.

## Scanline.java vs clean n.java — the `n` class is MERGED (Scanline + FontWidths)

The task said "verify Scanline.java matches clean n.java exactly." It does NOT, and cannot — because **clean obf `n` is not a Scanline class at all in the NAMING map; it is a dual-purpose merged class.** Evidence (`n.java`):
- Instance fields `int e; int k; int d; int l;` (n.java:3,5,10,11) — the per-row span struct (`Scene`'s `n[] x`; `ca`/`lb` field accesses `sl.d/sl.k/sl.e/sl.l`).
- PLUS static FontWidths members: `static int[] a = new int[256]` (n.java:4), `static nb h` (n.java:6), `static int[] j` (n.java:12), `static int[] m` (n.java:14), and the `<clinit>` (n.java:32-52) that fills `a[256] = 9*charset.indexOf(c)` from an XOR-decoded charset.

How the deob split it (both halves verified correct):
1. `Scanline.java` reconstructs ONLY the 4 instance fields (`d,k,e,l`, all `public int`). This is the correct struct for the Scene rasterizer — verified against `decompiled/.../lb.java:22` (`private n[] x`), `:3806` (`new n[...]`), `:3814` (`new n()`), and the deob `Scene.java:172/494/496` + field accesses `Scene.java:1853-2054` (`sl.d/sl.k/sl.e/sl.l`). Field order/types match.
2. The static FontWidths side (`n.a`) is reproduced INLINE in `Surface.java:2763-2787` (`GameFonts.buildCharacterWidth()`). I decoded the clean XOR charset string from `n.java:35` and diffed the resulting `a[256]` against the deob's hard-coded charset: **0 of 256 entries differ** — the char-width table is byte-exact.

Classification: this is a **MISSING-CLASS-MEMBER (documentation/architecture) gap, not a render-logic gap.** The reconstructed `Scanline.java` is missing the `n.a/n.h/n.j/n.m` static members and the `<clinit>`, but for the render path those are satisfied elsewhere (`Surface.GameFonts`), and the non-render consumers (`n.h/n.j/n.m` read by `cb/client/ra/m/mb/sa`) are out of scope. Fix sketch: add a doc note on `Scanline.java` that obf `n` was a merged class and the static FontWidths half lives in `Surface.GameFonts`; no logic change needed for rendering.

## HIGH-severity BEHAVIORAL — `k.o` (BMP image width) written to the wrong storage in ImageLoader

The obfuscator overloaded the name `k`: it is both the class `k` (=World, with `static int o`) and a static field `k` in several classes. The clean decompile of `pa.java` declares `static c k;` (pa.java:11) and then does `k.o = var2[12] + var2[13]*256` (pa.java:21), plus reads `k.o` for stride (pa.java:35,39,40). Vineflower rendered the `getstatic k.o`/`putstatic k.o` (World's `static int o`) as a field access on `pa`'s own `static c k` because that field exists.

Proof the field interpretation is wrong (javap on the jar): `World` (k) has `static int o;` (k.java:50); `LoaderThread` (c) has `private g o;` (c.java:40) — a `g`/ListNode, not an int, and `private` (inaccessible from `pa`). So `pa.k.o = <int>` would be both a type error and an access violation; the only valid reading is `World.o = <int>`. And the same slot is read by `fb.java:35` `setDimensions(k.o, da.bb)` and `lb.java:3770` `setPixels(0,0,k.o,da.bb,…)`.

The two deob files diverge on this single slot:
- `SurfaceImageProducer.addConsumer` (`SurfaceImageProducer.java:216`) reads it CORRECTLY as `World.surfaceWidth`.
- `ImageLoader.loadBmpImage` (`ImageLoader.java:106,287,296`) writes it INCORRECTLY to a fabricated `public static LoaderThread imageWidthCarrier; … imageWidthCarrier.imageWidth = bmpData[12]+bmpData[13]*256`. The value never reaches `World.surfaceWidth`.

Effect: the BMP's internal pixel flip works (ImageLoader reads back its own field), but the width that AWT and `Scene.setPixels` use (`World.surfaceWidth`) is never updated → every BMP-backed image (login/loading screens) is produced with stale/zero dimensions. Clean evidence: `pa.java:21` (write) and `fb.java:35` / `lb.java:3770` (read) all target the same `World.o`.
- Class/method: `ImageLoader.loadBmpImage`, gap type BEHAVIORAL.
- Fix sketch: replace `imageWidthCarrier.imageWidth` (the made-up LoaderThread carrier) with `client.world.World.surfaceWidth` for both the write (`ImageLoader.java:287`) and the two reads (`:296` `width`, and the pixel-loop stride). Drop the `imageWidthCarrier` field; keep `ClientStream.surfaceHeight` (=`da.bb`) as-is (that one is correct).

## Per-file findings

### Surface.java (ua) — comprehensive, 2 minor notes
- Coverage: ~66 distinct clean methods vs ~55 named deob methods + helpers; every clean signature has an `obf:` annotation (148 total). No MISSING-METHOD found. Blitter family verified present: opaque/clip (`spriteClipping` overloads at :1528/:1662/:1820), alpha (`drawSpriteAlpha`:1253, `blitSpriteAlpha`:1326, `blitSpriteAlphaIndexed`:1361), indexed (`blitSpriteIndexed` ×2 :1210/:1235), scaled/tinted (`plotScale`:1489, `plotScaleTinted`:1614, `transparentScale`:1769, `spriteClippingTinted`:1401), boxes/lines/gradient/circle (`drawBox`:575, `drawBoxEdge`:401, `drawGradientDirect`:426, `drawBoxAlpha`:481, `drawLineHoriz/Vert`:618/644, `setPixel`:669, `drawCircle`:691, `buildShadeRamp`:751), the drawstring family (:2373-2567 incl. full `@col@` table and `~ddd~` jump), and `drawMinimapSprite` rotate-blit (:2143) with span fillers `drawMinimap`/`drawMinimapTranslate` (:2313/:2326).
- Verified BEHAVIORAL-correct spots: `textWidth` glyph advance `fontData[characterWidth()[ch]+7]` (matches clean ua.java:1110/2159/4420); `drawstring` `~ddd~` digit guard `c>='0'&&c<='9'` (clean ua.java:2140 `-49>=~var11`); `centrepara` gates the `%` break on `forcePercentBreak` (clean behavior, noted in deob comment :2524-2526); `drawMinimapSprite` corner rotation `>>22`, `spanRightX init 0xfa0a1f01 == -99999999` (clean ua.java:2937-2938), texture sample `>>17` (clean ua.java:1153).
- Note 1 (SIG-DRIFT, benign): `Surface.GameFonts.data` (`m.b` gameFonts) and `antialiased` (`fb.k`) are local stand-ins (Surface.java:2765/2772) rather than the real `SocketFactory`/`SurfaceImageProducer` owners. Documented as such; not a logic gap but a cross-class wiring stub the integrator must reconcile (`fontAntialiased` should read `SurfaceImageProducer.k[font]`, currently a private 50-elem array always-false).
- Note 2 (cosmetic): `decodeStage2` (:2749) builds via `StringBuilder` instead of the clean in-place XOR + `intern()`; output identical, dead code anyway.

### SurfaceSprite.java (ba) — faithful
- `spriteClipping` entity dispatch (`SurfaceSprite.java:122`) matches clean `ba.a(...)` (ba.java:35) ranges 50000/40000/20000/5000 and the `flags!=29 → socialNames=null` clear (ba.java:72). `formatIpAddress` (:90) matches ba.java:24. `flushEventQueue` (:162) reconstructs the `$VF: Couldn't be decompiled` method (ba.java:80-192) from the provided bytecode (50-iteration peek loop, 1ms sleep, synthetic ActionEvent) — a correct, non-stub reconstruction of an otherwise-undecompilable method. No gaps.

### SpriteDecoder.java (ea) — faithful BZip2/BWT decoder
- `decompress`/`decompressBlocks`/`emitOutput`/`buildDecodeTables`/`makeMaps`/`readByte`/`readBit`/`readBits` map 1:1 to clean `a(byte[],…)`/`b`/`e`/`a(int[]…)`/`a(ac)`/`c`/`d`/`a(int,ac)`. Verified the error-prone RLE1 state machine (`emitOutput`:544-673) against clean `e(ac)`:444-563 — run lookahead 2→3→(4+extra) and byte advances match exactly. Output target `ua.Mb` correctly mapped to `Surface.tt` (declared `Surface.java:91`). The two "elided" comments (:662,:816) faithfully reproduce empty `if` blocks in clean ea.java:551-552/571-572. No gaps.

### SpriteScaler.java (ia) — faithful scanline scaler
- `writePaletteScaledScanline` (:195) matches clean `ia.a(...)` (ia.java:15): `srcStep<<=1`, 8-px unrolled main loop with the 5 noise masks correctly normalized to `(srcPos>>8)&0xFF`, remainder loop `0>negPixelCount` guard and odd-iteration advance (clean ia.java:62-77). `readPacketString`/`isChatCipherKnown` match ia.java:103/116. No gaps.

### ImageLoader.java (pa) — one HIGH bug (above), else faithful
- `buildMuLawTable` (:399) matches clean `pa.a(int)` μ-law decode (pa.java:58-89) incl. `~n` complement, `(mant<<(exp+2))-132`, sign negate on `var4!=0`. Static `<clinit>` trig+base64 tables match pa.java:91-132. Palette BGR decode and bottom-up→top-down row flip match pa.java:28-42.
- The `k.o` bug (BEHAVIORAL, HIGH) is the only render-affecting gap; see above.

### SurfaceImageProducer.java (fb) — faithful
- All `ImageProducer`/`ImageObserver` methods match clean fb.java: `addConsumer` stores `StringCodec.imageConsumer` (=`u.d`), `setDimensions(World.surfaceWidth, ClientStream.surfaceHeight)`, `setColorModel(SocketFactory.colorModel)`, `setHints(14)` (fb.java:31-42). `startProduction`→`addConsumer`, `removeConsumer`/`isConsumer` identity checks, `imageUpdate`→true. No gaps. (This file is the one that reads `k.o` correctly as `World.surfaceWidth`, exposing the ImageLoader divergence.)

## GAPS

Per-class counts by type:

| Class (obf) | MISSING-METHOD | MISSING-CLASS | SUMMARIZED/STUB | OBF-DRIFT | SIG-DRIFT | BEHAVIORAL |
|---|---|---|---|---|---|---|
| Surface (ua) | 0 | 0 | 0 | 0 | 1 (GameFonts stand-ins for `m.b`/`fb.k`) | 0 |
| SurfaceSprite (ba) | 0 | 0 | 0 | 0 | 0 | 0 |
| SpriteDecoder (ea) | 0 | 0 | 0 | 0 | 0 | 0 |
| SpriteScaler (ia) | 0 | 0 | 0 | 0 | 0 | 0 |
| ImageLoader (pa) | 0 | 0 | 0 | 0 | 0 | **1 (`k.o` width → wrong storage)** |
| SurfaceImageProducer (fb) | 0 | 0 | 0 | 0 | 0 | 0 |
| Scanline (n) | 0 | 0 | 1 (static FontWidths half not in this file; satisfied in Surface.GameFonts) | 0 | 0 | 0 |

Totals: 0 MISSING-METHOD, 0 MISSING-CLASS (Scanline already reconstructed; no other missing render class found), 1 doc-level SUMMARIZED, 0 OBF-DRIFT, 1 SIG-DRIFT (benign cross-class stub), 1 BEHAVIORAL.

Highest-severity items (ordered):
1. **HIGH — BEHAVIORAL — `ImageLoader.loadBmpImage` writes BMP width to a fabricated `imageWidthCarrier` (LoaderThread) instead of `World.surfaceWidth`** (`ImageLoader.java:287/296/106`). Clean `pa.java:21` `k.o=…` is `World.o` (proven via javap: `World.o` is `static int`; `LoaderThread.o` is `private g`), the same slot read by `SurfaceImageProducer.addConsumer` (`fb.java:35`) and `Scene.setPixels` (`lb.java:3770`). Result: BMP-backed images get stale/zero AWT dimensions. Fix: write/read `World.surfaceWidth`, delete `imageWidthCarrier`.
2. **MEDIUM — SIG-DRIFT/wiring — `Surface.fontAntialiased` reads a private local `GameFonts.antialiased[]` (always false) instead of `SurfaceImageProducer.k[font]`** (`Surface.java:2722/2769`; clean ua.java:2160/2164 `fb.k[var7]`). With it stuck false, anti-aliased fonts wrongly get the +1,+1 drop-shadow and the 1-bit plot. Same applies to `GameFonts.data` vs the real `m.b`. Fix at integration: point these at the real owners.
3. **LOW — doc/architecture — `Scanline.java` omits the static FontWidths half of merged obf class `n`** (`n.java:4,6,12,14` + `<clinit>`). Render path is fine (the `n.a` table is byte-exact in `Surface.GameFonts`); only add a note that obf `n` was Scanline+FontWidths merged and the static half is reproduced in Surface.
