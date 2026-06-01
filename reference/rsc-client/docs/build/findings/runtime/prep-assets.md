> _Boot-asset staging report (2026-06-01)._

# Boot-Asset Staging Report (M1/M2)

## Key finding (corrects the task premise)
This is the **rev 233-235 J++ `rsclassic` client**, not the classic mudclient. Two assumptions in the plan are wrong for this codebase:

1. **`jagex.jag` does NOT need to be built — it already exists as content index 3.** `GameShell.loadJagex()` calls `readDataFile("jagex.jag", 0, 3, 85)` → `StreamBase.loadResource(-101, "jagex.jag", 0, 3)`. The name string is **ignored**; the resource is indexed by the int `3`. Content index 3 (`content3_5181c9f5`) decompresses to a 25204-byte JAG archive containing exactly one entry — `logo.tga` (281×85, 8bpp color-mapped TGA, hash `-1752591533`). I verified content3's CRC32 = `0x5181c9f5` = its filename suffix = `contentcrcs[3]`, so the client's CRC check passes.

2. **The 8 bitmap fonts do NOT come from any `.jag` and the `.jf` files are unused.** `Panel.loadFont` parses the font name ("h11p".."h24b") only for point-size/style, then **rasterizes 95 glyphs at runtime from `java.awt.Font("Helvetica", style, size)`** via `FontBuilder.rasterizeGlyph`. I confirmed Helvetica resolves with valid metrics for all 8 sizes under JDK 17 (even headless). `rscplus/assets/jf/*.jf` are dead for this client.

`loadJagex()` is boot-critical: if it returns false, `GameShell.run()` calls `closeProgram()`. It returns false only if content-3 fails to load or any `loadFont` fails — both satisfied.

## Asset formats
- **Content pack (outer):** 6-byte header = rawLen(3B BE) + compLen(3B BE), then an RSC-BZip2 stream with the 4-byte `BZh1` header stripped (`World.unpackData`: if rawLen==compLen the body is verbatim, else prepend `BZh1` and bunzip2). All 12 packs decompress cleanly.
- **JAG archive (inner):** `numEntries`(2B BE) + numEntries×10-byte dir records `[hash:int32][usize:3B][csize:3B]` + concatenated entry data; name-hash = `hash*61+(ch-' ')` over the uppercased name; if usize≠csize the entry is individually BZip2'd (`EntityDef.extractArchiveEntry`).
- **`contentcrcs`** (58B): 12 big-endian CRC32 ints + a trailing verify-CRC at offset 48 (`0x61a8182d` = standard CRC32 over the first 48 bytes — validated; `verifyCrc` passes) + 6 zero pad bytes.
- **HTTP fetch:** `DownloadWorker`→`LoaderThread.openUrl` uses plain `URL.openStream()` (a GET on the literal path). Negative CRCs are URL-formatted via `Long.toHexString(int)` → sign-extended 16-hex-digit form (e.g. `content9_ffffffffe0e19e2c`) — which is exactly how the rscplus assets are already named.

## Inputs gathered (`rscplus/assets`)
- `content/`: **content0..11 + contentcrcs all present** (the matched rev-235 set; all CRCs consistent). No `jagex.jag`/`logo.tga` loose files (not needed — it's content3).
- `jf/`: 10 `.jf` fonts present but **unused** by this client.

## Staged under `/tmp/rsc-run`
- `/tmp/rsc-run/cache/` — all 12 content packs + `contentcrcs` copied verbatim, plus `jagex.jag` (a convenience copy of content3; byte-identical — confirmed via `cmp`).
- `/tmp/rsc-run/content-host.py` — custom static host. A plain `python3 -m http.server` is **insufficient**: the CRC manifest is fetched as `/contentcrcs<hex-millis>` (cache-buster suffix), which 404s on a vanilla server. This handler serves content packs verbatim and collapses any `/contentcrcs*` → the `contentcrcs` file. Smoke-tested end-to-end: `/content3_5181c9f5` returns matching sha256, `/contentcrcs18f3abcd` returns the 58-byte manifest, `/content9_ffffffff...` serves verbatim, missing files 404.
- `/tmp/rsc-run/start-content-host.sh` — launcher (default port 43594).
- `/tmp/rsc-run/work/jag.py`, `logo.tga`, `FontProbe.java` — verification tooling/artifacts.
- (Server prep `server-*.{log,txt}`, `start-server.sh` were placed by the parallel 0a workflow.)

## Content delivery decision: HTTP host (path a) — required, not optional
The pre-stage + skip path (`doUpdate <= 20`) does **NOT** work without a code change: `Buffer._junkArray` (the CRC table that `loadResource` uses to build URLs and check CRCs) is populated **only** by `CacheUpdater.downloadAndVerifyCrcs`, which runs **only** when `doUpdate > 20`. Skipping the update leaves `_junkArray` all-zero → `loadResource` requests `content3_0` and checks CRC against 0 → fails. So: run the HTTP host and pass `doUpdate > 20` with `BASE_URL = http://127.0.0.1:<port>` (`startApplication` hardcodes host `127.0.0.1`). This needs **zero changes** to the asset filenames. (Alternatively, a tiny `main()` change could pre-fill `_junkArray` from `contentcrcs` and skip HTTP — but the host is simpler and the recommended path.)

## Minimal-to-boot set
- `loadJagex` (M1, window+logo+fonts): content index **3** only (logo) + AWT fonts (no content).
- **Title/login screen (M2):** the theoretical minimum is index **0** (GameData) + **8** (2d graphics for `drawLoginScreen`). **But** `loadGameConfig` runs the full load sequence as one block before drawing, bailing on any `fatalLoadError`: it pulls **0** (GameData/`drawOptionsTab`), **8** (`loadMedia2d`), **1**(+**2** members) (`loadEntitySprites`), **11** (`loadTextures`), **9** (`loadModelDefs`), **4,5,6,7** (`loadMaps`). So in practice **all 12 indices must resolve to reach the title screen.** All 12 are staged, so this is satisfied.
- Live frame (M4): same set 9/11/6/4/1 — already covered.

## Blockers
None for assets. `jagex.jag` did NOT need a faithful hand-build (it's content3, staged and CRC-verified). The only hard requirement uncovered: the content host is mandatory (pre-stage+skip alone fails due to the `_junkArray` dependency). M1/M2 remain gated on the parallel Mudclient compile (M0), which is out of scope here.

## ASSETS STATUS: READY
