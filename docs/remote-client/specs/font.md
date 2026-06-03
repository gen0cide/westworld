# Spec: Authentic bitmap font (C1)

**Status:** ready to implement. **Build complexity:** moderate (asset prep + CSS,
no React logic).
**Backend:** NONE. This is pure web chrome. No `/state` field, no Host method, no
action endpoint, no `cmd/cradle` change. Everything lives under `web/`.

> Implementer: read this whole file. Every value below is copy-pasteable. The only
> judgement call left is "which font binary do I download/install" (§4 step 1), and
> two acceptable answers are given.

---

## 0. What the real client does (ground truth)

`reference/rsc-client/src/client/ui/Panel.java` (READ ONLY) builds every glyph at
runtime by rasterizing **`new Font("Helvetica", style, pointSize)`**
(`Panel.java:268` `FONT_NAME_HELVETICA = "Helvetica"`, `Panel.java:1913`) into a
packed greyscale-alpha bitmap via `FontBuilder.rasterizeGlyph`
(`FontBuilder.java:167`). On the JVM "Helvetica" is a *logical* font that maps to
the platform sans-serif — in practice **Adobe Helvetica / Arial / Liberation Sans /
Nimbus Sans**, i.e. the metrics the original screenshots were captured with.

The renderer addresses **8 font slots** (ids 0–7). The authoritative line-height
table is hard-coded in `reference/rsc-client/src/client/scene/Surface.java:2350-2359`
(`textHeight`):

| font id | line height (px) | classic Helvetica spec | style | typical use |
|--------:|-----------------:|------------------------|-------|-------------|
| 0 | 12 | `helvetica11` plain  | plain | small UI text, inventory qty, chat history |
| 1 | 14 | `helvetica12` **bold** | bold | default UI / menu options / chat input |
| 2 | 14 | `helvetica12` plain  | plain | secondary UI |
| 3 | 15 | `helvetica13` **bold** | bold | headers |
| 4 | 15 | `helvetica14` **bold** | bold | headers |
| 5 | 19 | `helvetica16` **bold** | bold | section titles |
| 6 | 24 | `helvetica20` **bold** | bold | big labels (`Mudclient.java:2786` uses `Font("Helvetica",BOLD,20)`) |
| 7 | 29 | `helvetica24` **bold** | bold | huge banners (`Mudclient.java:2736/2750`, BOLD 20 title screen) |

Key facts that drive the web approach:

- **It is Helvetica/Arial-metrics sans-serif, NOT a bespoke pixel font.** The
  "bitmap" look is just small Helvetica rasterized at integer point sizes with the
  AWT renderer's anti-aliasing (`FontBuilder.java:382`, the `30<alpha<230` AA
  detection). RSC at these sizes is mostly **1-bit, no anti-alias** on the small
  fonts (font 0/1) and lightly AA on the big ones.
- Each printable glyph is one of the 95 chars in
  `Panel.VALID_CHARS` (`Panel.java:262`):
  `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!"£$%^&*()-_=+[{]};:'@#~,<.>/?\| ` + space.
- Text draws with a **1px (+1,+1) black drop shadow** on the non-AA bitmap fonts
  when `colour != 0` (`Surface.java:2468-2472`). AA fonts skip the shadow.
- Inline colour codes `@xxx@` (`Surface.java:2429-2449`) — already handled in the
  chat layer; not a font concern.

The current SPA uses `font: 13px monospace` (`web/src/styles.css:24`) and
`11px/10px/12px monospace` in tab/inv/stats rules — visibly wrong (monospace,
wrong metrics, no shadow).

---

## 1. Decision: which of the three options

| option | fidelity | effort | downside |
|---|---|---|---|
| **A. Ship a matching webfont + `image-rendering` + tuned `font-size`/`line-height`** | high | low–med | webfont hinting ≠ AWT raster, but at 1× it's near-identical; scales cleanly |
| B. Bake `FontBuilder` glyph bitmaps into a sprite sheet + canvas/CSS text renderer | exact | high | needs a JVM/AWT extraction pass, a glyph-atlas asset, and a custom canvas text-layout engine in React; brittle; only worth it for 1:1 frame diffs |
| C. Close-enough monospace tuned to the metrics table | low | trivial | never matches — RSC Helvetica is proportional, not monospaced |

### RECOMMENDATION: **Option A** — ship a self-hosted Helvetica-metric webfont,
render normal DOM/CSS text, tune `font-size`/`line-height` per font slot to the
table above, and add the authentic 1px black text-shadow.

Rationale: the real client *is* small Helvetica; a self-hosted **Liberation Sans**
(or **Arimo**, both metric-compatible with Arial/Helvetica, OFL/free to embed) at
the matching px sizes reproduces the look without an AWT extraction pipeline, keeps
text selectable/accessible, scales with the responsive viewport, and is a
~`web/`-only change. Option B is reserved for a future "exact frame-diff" task; it
is explicitly NOT this task and is documented in §7 as a follow-up.

`image-rendering: pixelated` (already used on `#view img` and `.cell .icon`) does
**not** affect DOM text rendering — it only matters if we ever rasterize text to a
`<canvas>` or use a bitmap-image font. For Option A the crispness comes from
integer px sizes + disabling sub-pixel/AA smoothing via CSS (`§3`).

---

## 2. Backend additions

**None.** Confirmed against `cmd/cradle/remoteclient.go`: fonts are presentation
only. Do not add a `/state` field, a Host call, or a route. The bank template's Go
struct/handler pattern does **not** apply here. (If a future maintainer wants the
server to *serve* the font binary instead of bundling it, that is a `/sprite`-style
static route — out of scope; the Vite `public/` bundle in §4 is simpler and is what
this spec uses.)

No Host method signatures are called. No `world/` or `runtime/` dependency.

---

## 3. Web changes

### 3.1 Asset location

Vite copies `web/public/**` verbatim to the site root and into `web/dist/` (which
`web/embed.go` `//go:embed all:dist` then bundles into the cradle binary). Create:

```
web/public/fonts/rsc-sans.woff2      ← the Helvetica-metric webfont (see §4 step 1)
web/public/fonts/LICENSE.txt         ← the font's license (Liberation = OFL/GPL+exc; Arimo = Apache-2.0)
```

These ship as `/fonts/rsc-sans.woff2` at runtime and are auto-embedded — no
`embed.go` edit needed (it already globs `all:dist`).

### 3.2 `web/src/styles.css` — the only code file that MUST change

Add the `@font-face` + RSC font CSS vars at the **top** of the file (after the
header comment, before `:root`). Then retune the px/family rules. Concrete diff
units below; apply each.

**(a) Add at top of file:**

```css
/* Authentic RSC bitmap font (C1): the client rasterizes "Helvetica"
 * (reference/rsc-client/src/client/ui/Panel.java:1913) at 8 integer point sizes
 * (line-height table: Surface.java:2350-2359). We self-host a Helvetica-metric
 * webfont and match those px sizes; the 1px black drop-shadow reproduces the
 * non-AA bitmap shadow (Surface.java:2468-2472). */
@font-face {
  font-family: 'RSCSans';
  src: url('/fonts/rsc-sans.woff2') format('woff2');
  font-weight: 400 700;     /* one variable/static file covering plain+bold */
  font-style: normal;
  font-display: block;      /* avoid FOUT flashing the wrong metrics */
}
```

**(b) Add these vars inside the existing `:root { … }` block** (alongside the
palette vars at `styles.css:6-17`):

```css
  /* RSC font stack + per-slot sizes (px = the slot's point size from §0 table).
   * Stack falls back to Arial/sans so text still renders before the woff2 loads
   * or if it 404s. */
  --rsc-font: 'RSCSans', Arial, 'Liberation Sans', 'Helvetica Neue', sans-serif;
  --rsc-fs-small: 11px;   /* font 0  (lh 12) */
  --rsc-fs-base:  12px;   /* font 1/2 (lh 14) */
  --rsc-fs-head:  13px;   /* font 3  (lh 15) */
  --rsc-fs-title: 16px;   /* font 5  (lh 19) */
  --rsc-fs-big:   20px;   /* font 6  (lh 24) */
  --rsc-lh-small: 12px;
  --rsc-lh-base:  14px;
  --rsc-lh-head:  15px;
  --rsc-lh-title: 19px;
  --rsc-lh-big:   24px;
  /* authentic 1px black drop shadow (Surface.java +1,+1). */
  --rsc-text-shadow: 1px 1px 0 #000;
}
```

**(c) Replace every hard-coded `monospace` / pixel size** with the vars. Exact
edits (old → new):

| selector (line) | old | new |
|---|---|---|
| `html,body,#root` (`:24`) | `font: 13px monospace;` | `font: var(--rsc-fs-base)/1 var(--rsc-font);` |
| add to `html,body,#root` | — | `-webkit-font-smoothing: none; font-smooth: never; text-rendering: optimizeSpeed;` |
| `.chatlog` (`:82-89`) | `line-height: 1.35;` | `line-height: var(--rsc-lh-base); font-size: var(--rsc-fs-base); text-shadow: var(--rsc-text-shadow);` |
| `.chatinput` (`:96-106`) | `font: 13px monospace;` | `font: var(--rsc-fs-base)/var(--rsc-lh-base) var(--rsc-font);` |
| `.tabbar button` (`:121-130`) | `font: 11px monospace;` | `font: var(--rsc-fs-small)/var(--rsc-lh-small) var(--rsc-font);` |
| `.cell` (`:141-155`) | `font-size: 10px;` | `font-size: var(--rsc-fs-small);` |
| `.cell .qty` (`:157`) | `font-size: 9px;` | `font-size: var(--rsc-fs-small); text-shadow: var(--rsc-text-shadow);` |
| `.cell .stub` (`:159`) | `font-size: 10px;` | `font-size: var(--rsc-fs-small);` |
| `.stats` (`:162`) | `font-size: 12px;` | `font-size: var(--rsc-fs-base);` |
| `.equip` (`:135`) | `font-size: 12px;` | `font-size: var(--rsc-fs-base);` |
| `.menu` (`:214-222`) | `font: 13px monospace;` | `font: var(--rsc-fs-base)/var(--rsc-lh-base) var(--rsc-font);` |
| `.bankhdr button` (`:195-202`) | `font: 12px monospace;` | `font: var(--rsc-fs-base)/var(--rsc-lh-base) var(--rsc-font);` |
| `.hud` (`:53-60`) | (keep `white-space: pre`) | add `font-size: var(--rsc-fs-base); text-shadow: var(--rsc-text-shadow);` |

> Rule: after this pass, **`grep -n monospace web/src/styles.css` must return
> nothing.** That is the acceptance check for the CSS pass.

**(d) Optional helper classes** (so future components can pick a slot by name
instead of raw px). Add at the very bottom of `styles.css`:

```css
/* RSC font-slot helpers — pick by the engine's font id (see spec §0). */
.rsc-f0 { font: var(--rsc-fs-small)/var(--rsc-lh-small) var(--rsc-font); }
.rsc-f1 { font: 700 var(--rsc-fs-base)/var(--rsc-lh-base) var(--rsc-font); }   /* default bold */
.rsc-f2 { font: 400 var(--rsc-fs-base)/var(--rsc-lh-base) var(--rsc-font); }
.rsc-f3 { font: 700 var(--rsc-fs-head)/var(--rsc-lh-head) var(--rsc-font); }
.rsc-f5 { font: 700 var(--rsc-fs-title)/var(--rsc-lh-title) var(--rsc-font); }
.rsc-f6 { font: 700 var(--rsc-fs-big)/var(--rsc-lh-big) var(--rsc-font); }
.rsc-shadow { text-shadow: var(--rsc-text-shadow); }
```

### 3.3 No new types, api, component, or App wiring

- `web/src/types.ts`: **no change** (no new wire field).
- `web/src/api.ts`: **no change** (no new endpoint).
- `web/src/App.tsx`: **no change** (font applies globally via `styles.css`, already
  imported in `web/src/main.tsx:5`).
- `web/src/components/*`: **no change required.** The CSS var swap recolors every
  existing component. (Optional polish: components that currently set inline px
  sizes can adopt the `.rsc-fN` helpers — but none currently do, so this is purely
  optional and not part of the acceptance gate.)
- `web/index.html`: **no change.** Do NOT add a Google-Fonts `<link>` — the font is
  self-hosted under `public/fonts/` so the embedded binary has zero external
  network dependency (matches the project's offline-binary goal).

There is no left/right-click interaction for this feature — it is non-interactive
chrome. (The bank template's click handlers have no analogue here.)

---

## 4. Implementation steps (mechanical)

1. **Obtain the webfont** (pick ONE; both are free to embed and Arial/Helvetica
   metric-compatible):
   - **Liberation Sans** (recommended; SIL OFL / GPL-with-font-exception). Source:
     the `liberation-fonts` release zip, file `LiberationSans-Regular.ttf` +
     `LiberationSans-Bold.ttf`. Convert to one woff2:
     ```
     # if only static TTFs are available, subset+convert with fonttools:
     pip install fonttools brotli
     pyftsubset LiberationSans-Regular.ttf \
       --text='ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!"£$%^&*()-_=+[{]};:'"'"'@#~,<.>/?\| ' \
       --flavor=woff2 --output-file=web/public/fonts/rsc-sans.woff2
     ```
     (Subset to the 95 `VALID_CHARS` glyphs from §0 to keep the asset tiny, <8 KB.)
   - **Arimo** (Apache-2.0; also Arial-metric) — same procedure with
     `Arimo-Regular.ttf`.

   If a combined plain+bold file is wanted, package both weights or ship a second
   `rsc-sans-bold.woff2` and add a second `@font-face` with `font-weight: 700`.
   The single-weight stack in §3.2 is acceptable (browser synth-bolds slot 1).

2. Drop the woff2 (+ `LICENSE.txt`) into `web/public/fonts/` (create the dir).

3. Apply the `web/src/styles.css` edits in §3.2 (a)–(d).

4. Build + embed:
   ```
   npm --prefix web run build
   go build -o /tmp/cradle ./cmd/cradle
   ```
   `npm run build` runs `tsc -b && vite build`; the new `public/fonts/*` lands in
   `web/dist/fonts/*` and is `//go:embed`'d automatically (`web/embed.go:14`).

5. **Do NOT commit. Do NOT restart the running :8090 server.** Verify against the
   dev server or a throwaway `/tmp/cradle` per §6.

---

## 5. Files touched (summary)

| file | change |
|---|---|
| `web/public/fonts/rsc-sans.woff2` | **new** — Helvetica-metric subset webfont |
| `web/public/fonts/LICENSE.txt` | **new** — font license |
| `web/src/styles.css` | edit — `@font-face`, `:root` font vars, swap all `monospace`/px to vars, add shadow + `.rsc-fN` helpers |
| `web/dist/**` | regenerated by `npm run build` (do not hand-edit) |

No file under `cmd/cradle/`, `world/`, `runtime/`, or `reference/` changes.

---

## 6. Test plan

**Static / build gate**
1. `grep -n monospace web/src/styles.css` → **no output** (all swapped).
2. `npm --prefix web run build` → exits 0; `web/dist/fonts/rsc-sans.woff2` exists
   (`ls web/dist/fonts/`).
3. `go build ./...` → green (embed picks up the new asset).

**Runtime (dev server, do NOT touch the live :8090 cradle)**
4. `cd web && npm run dev`, open `http://localhost:5173`. In DevTools → Network,
   `/fonts/rsc-sans.woff2` returns **200** (not 404) and `font-family` computed on
   `body` resolves to `RSCSans`.
5. Visual diff vs `reference` screenshots / the `/legacy` page: chat lines,
   inventory qty badges, side-panel tab labels, stats rows, and the right-click
   menu render in proportional Helvetica-metric text (NOT monospace) with a crisp
   1px black shadow on colored text.
6. Force a 404 (rename the woff2) and reload → text still renders via the
   `Arial, 'Liberation Sans', sans-serif` fallback (no invisible text;
   `font-display: block` then falls back after the block period).
7. Confirm `image-rendering: pixelated` on `#view img` and `.cell .icon` is
   unaffected (still set; text sharpness comes from `-webkit-font-smoothing: none`
   + integer px, not from `image-rendering`).

**Acceptance:** steps 1–5 pass; the SPA's text matches the metrics table in §0 and
no longer uses `monospace`.

---

## 7. Out of scope / follow-ups (do NOT do here)

- **Option B (exact glyph atlas).** A future "frame-perfect" task could extract the
  real packed glyphs by running `FontBuilder.rasterizeGlyph` in a tiny JVM harness
  (or porting it to Go), emitting a 95-glyph alpha atlas PNG + a per-glyph metrics
  JSON (`{advance,xBearing,yBearing,w,h}` per `FontBuilder.java:338-363`), serving
  the atlas via a `/sprite?kind=font&id=N`-style route, and writing a
  `<canvas>`/CSS text-layout component that blits glyphs by hand (honoring `@col@`
  + the +1,+1 shadow). That reproduces RSC pixels exactly but is a large, separate
  unit. This spec deliberately ships Option A first.
- UI-chrome sprite tab strip / borders (task B4 / C3) — separate.
- The `@col@`/`~ddd~` inline-markup parser already lives in the chat layer; no font
  work needed.
