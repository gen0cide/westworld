# Protocol decode debug + profile strategy

> **STATUS: PARTIALLY BUILT** (verified 2026-06-10 against branch
> `tidy/structure-and-docs`, HEAD `0bfa818`). The failure-mode taxonomy and the
> debug playbook below are evergreen — they are the doc's value. Of the four
> detection tiers: **Tier 1 is BUILT** (`proto/v235/anomaly.go`); **Tier 2 was
> re-planned** as Phase 5 of the Plutonium migration plan
> ([`../_research/plutonium/EXECUTION-PLAN.md`](../_research/plutonium/EXECUTION-PLAN.md),
> gated on [TODO.md](../TODO.md) **MP-5**/**MP-6**); **Tier 3 is SUPERSEDED** by
> the Plutonium comparative study
> ([`../_research/plutonium/`](../_research/plutonium/README.md)); **Tier 4 is
> DESIGN, not built**.

RSC's wire protocol has quirks every era of the client/server has
added or papered over. Some examples we've already hit, with more
guaranteed to surface as we exercise new opcodes:

- **Conditional fields by item type.** Inventory amount is sent
  ONLY for stackables. Reading unconditionally corrupted every
  later slot. (Fixed 2026-05-28.)
- **Smart-int encoding variants.** "Smart 8/16" (1 byte if <128,
  else 2-byte+offset), "unsigned short int" (2 bytes if value
  fits in short, else 4 bytes with high bit set), "smart 16/32"
  (similar tiered encoding). Easy to read one as the other —
  bytes overlap, no shape marker.
- **Bit-packed flags.** Wielded bit in high bit of item id short.
  Other packets have N-flag bytes where some bits gate optional
  following fields.
- **Direction encoding for boundaries.** 0-3 for cardinal, but
  the byte's meaning depends on whether you're reading a tile's
  HorizontalWall (north edge), VerticalWall (west edge), or
  Diagonal (>48000 = scenery in disguise; we missed that one
  too — multi-tile scenery wasn't expanded.)
- **Coordinate offsets that wrap modulo sector.** NPC coords
  arrive as 5-bit deltas relative to player, with wrap; off-by-
  one in the sign extension silently produces NPCs on the wrong
  side of the map.

Categorizing the failure modes:

1. **Off-by-N: decoder consumes too many or too few bytes per
   element, every subsequent element corrupted.** (The inventory
   bug.) Signature: one field looks fine for slot 0, slot 1+ is
   garbage; later fields wrap or out-of-range.
2. **Field misalignment: decoder reads the right number of
   bytes but interprets them as the wrong type or position.**
   Signature: per-slot values look "noisy but plausible" — e.g.
   small ints where coords should be, or vice versa.
3. **Conditional fields missed.** Signature: works for one item
   type, fails for another (stackable vs not, noted vs not,
   wielded vs not).
4. **Bit-flag misinterpretation.** Signature: half the items
   look wielded; or every NPC looks aggressive.

## Detection strategies

### Tier 1 — anomaly assertions in the decoder — BUILT

`proto/v235/anomaly.go` implements this. `flagAnomaly(field, value,
plausibleMax, raw)` logs a WARN (`"proto decode anomaly"`) with the
field label, the decoded value, the plausible ceiling, and the hex of
the raw bytes that produced it — the smoking gun is in the log line,
no hand-decoding required. The plausible-range table
(`plausibleMaxItemID=5000`, `plausibleMaxNpcIndex=0xF000`,
`plausibleMaxCoord=16384`, `plausibleMaxDamage=1000`,
`plausibleMaxAmount=2e9`) is deliberately loose: anything past it is
almost certainly a decode bug, not data.

Wired call sites today: the two inventory decoders —
`decodeInventory` checks item id + amount per slot
(`proto/v235/inbound.go:445`), `decodeInventorySlotUpdate` checks the
amount (`inbound.go:478`). These guard exactly the off-by-N cascade
that produced the original `bronze Axe x10879108` bug. The
npc-index / coord / damage constants are declared but have **no call
sites yet** — extend coverage opportunistically per step 4 of the
playbook below, whenever a decoder bites. One further Tier-1
assertion lives above the proto layer: the opcode-79 NPC positional
desync WARN at `runtime/frame.go:117` (its comment cites this doc)
catches the localNpcs order-list diverging from the server's count.

Cost is ~1 LOC per check at the call site, run once per decoded
field. The bug-find time saved is enormous — the inventory bug took
30 minutes of multi-layer debugging because nothing in our system
*complained* about a field decoding to 10879108.

### Tier 2 — Ground-truth pcap comparison — RE-PLANNED

The original sketch here (decode a server-side pcap, assert against
checked-in golden event shapes) survives, sharpened, as **Phase 5 of
[`../_research/plutonium/EXECUTION-PLAN.md`](../_research/plutonium/EXECUTION-PLAN.md)**:
capture a reference stream with the server's `want_pcap_logging`
toggled on, build `proto/v235/pcap/reader.go`, and validate decoders
against golden frames — coords (opcodes 191/79) as the first blocking
smoke gate, then the four silent-corruptor payload reshapes, each
subtest asserting the FULL record list (because every reshape
corrupts every later record in its packet).

Two corrections to the original sketch, both from the study:

- The server does **not** write pcaps by default —
  `westworld.conf` has `want_pcap_logging: false`; capture is a
  deliberate config toggle (the recipe is TODO **MP-6**).
- The byte-layout reference is the **server source**
  (`Payload235Generator.java` / `GameStateUpdater.java` /
  `Payload235Parser.java`), not the Java client and not Plutonium.

Status: planned, zero code — the whole pcap-golden program rides on
the un-ratified 10010 migration decision, TODO **MP-5** (NEEDS
OPERATOR), with **MP-6** as its protocol-independent prerequisite.

### Tier 3 — Side-by-side mode — SUPERSEDED

The original idea: run our cradle in "shadow" mode against the same
server connection as a real reference client and diff decoded events
frame-by-frame. The **Plutonium comparative study**
([`../_research/plutonium/`](../_research/plutonium/README.md) —
17 dimensions, 78 source-verified findings) delivered this tier's
goal statically, and killed the live version of the idea:

- Verdict: our protocol layer is **fine** — authentic v235,
  byte-verified against the server source. The live movement bugs
  traced to one subsystem (`pathfind/grid.go`), not decode.
- Plutonium is **not a byte-oracle**: it speaks the custom 10010
  dialect (plain length framing, no ISAAC), we speak authentic 235.
  Frame-by-frame byte comparison against it would mislead. Use
  Plutonium for algorithms and primitives; use the server source
  for wire layouts; use Tier 2 pcap goldens for byte-level proof.

### Tier 4 — Fuzzing the decoders — DESIGN, not built

For each opcode handler in `proto/v235/inbound.go`, generate
random payloads and check the decoder doesn't panic / over-read.
Catches buffer-bound bugs but not interpretation bugs. Lowest
ROI; still unbuilt, still last. (Not tracked as a TODO item; the
adjacent DSL lexer/parser fuzzing is [TODO.md](../TODO.md) DSL-15.)

## Profile / observability hooks — what exists

- `legacy-cradle -v` (`cmd/legacy-cradle/main.go:125`) flips slog to
  DEBUG. Decode failures WARN with the opcode and error at
  `runtime/frame.go:163` — but **without** raw payload hex; the
  per-frame `(opcode, len, hex)` dump this doc once wished for was
  never built. When you need raw bytes, hex-dump at the failing
  decoder (playbook step 1).
- Decoded-event tracing is solved by the **debughttp control plane**
  (`debughttp/server.go`): every bus event is recorded to an
  in-memory ring + JSONL file (default
  `/tmp/cradle_debug/<username>_events.jsonl`), queryable via
  `GET /events?since=N&kind=K&limit=L`. Reachable two ways:
  `legacy-cradle -debug-http`, or per-host through the cradle
  daemon at `/api/hosts/{name}/debug/…` (`cradle/api.go`). Grepping
  the JSONL by event `kind` replaces the old "event stream marker"
  idea wholesale.
- An ad-hoc `-dump-opcode=NNN` flag was proposed and never built.

Remaining wire-debug work items live in [TODO.md](../TODO.md) — see
**MP-3** (opcode-validity gate + ISAAC regression test), **MP-5/MP-6**
(pcap goldens), **MP-10** (protocol research tails), **MP-12**
(sleep/wire hardening).

## What to do when a new opcode surfaces a bug

1. **Capture the raw bytes.** `b.data[start:b.rpos]` hex-dumped.
   Without the raw bytes, you can't tell shape bugs from
   interpretation bugs.
2. **Compare with the Java client's decoder.** OpenRSC source is
   `~/Code/openrsc/server/src/com/openrsc/server/net/rsc/parsers/impl/Payload235Parser.java`
   and `Packet.java` for reader primitives.
3. **Look for `writeUnsignedShortInt`, `writeSmart`, conditional
   `if` branches in the SENDER side.** ActionSender.java often
   has comments explaining which fields are skipped for which
   conditions.
4. **Add a Tier 1 anomaly assertion** so the bug can't reoccur
   silently — the next person to break this gets a WARN line,
   not a "well, why doesn't use(key, door) work" hour.
5. **Backfill a Tier 2 golden-file test** if the opcode is
   important (anything in the critical-path of the host loop).

(Per the Plutonium study's authority hierarchy: when steps 2–3
disagree with anything, the server source generators —
`Payload235Generator.java` / `GameStateUpdater.java` — win.)

## Heuristic: "looks reasonable" isn't safe

The inventory bug shipped because `amount = 10879108` looks like
"oh, the stackable amount field exists, we just read junk
because we didn't filter by stackability." It was actually
proof the entire decoder offset chain was off-by-N. The eye
glosses over "huge number = wrong-but-localized"; the reality is
often "wrong AND cascading."

When a decoder produces a single weird value, the first instinct
should be to dump the next 50 bytes raw and recheck the wire
shape — not to clamp the value and move on.
