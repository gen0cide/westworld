# 2 — The wire protocol: byte truth and the cost of "looks reasonable"

> **The rules, up front.** Decoders fail loudly or they hide cascades. One weird
> field means re-derive the frame from the server source — dump the next 50 raw
> bytes, never clamp. Byte-identical payloads are still two opcodes. Read
> repeated records to payload exhaustion unless the server's generator proves a
> single record. Desync is fatal by policy — recovery loops that resync past
> corruption are how you ship corruption. A clean EOF is not a clean shutdown —
> ask *who* closed. Echo the server's exact lookup keys. Length prefixes count
> what the codec counts. And design events against what the wire actually
> carries — abstain over misattribute.

RSC's v235 wire is a 2001-era protocol with twenty years of accumulated
quirks: conditional fields keyed on item type, three different "smart"
integer encodings, bit-packed coordinate deltas, a tail-byte that wanders
inside the frame, and an opcode byte that is the *only* encrypted thing on the
wire. Every story in this chapter is the same story at a different offset: the
decoder produced something that *looked* plausible, nothing complained, and the
corruption surfaced one or more layers up wearing a costume — a phantom
inventory, a spell instead of an item, an invisible pickaxe, a fleet that
quietly died in its sleep. The fix, every time, was to stop reasoning about the
symptom layer and re-derive the byte truth from the server source
(`Payload235Generator.java` / `Payload235Parser.java` / `GameStateUpdater.java`
— the authority hierarchy this project settled on; see chapter 1
([1_PATHING.md](1_PATHING.md)) for why the reference bot is *never* the byte
oracle).

## The problem as experienced

**The key that wasn't a key (2026-05-28).** While wiring `use(key, door)`, the
server kept receiving `Item(166, 1, false)` for what should have been the key
at slot 1. Upstream, the inventory mirror showed entries like `bronze Axe
x10879108`. The decode bug: `decodeInventory` read a uint32 amount for *every*
slot, but RSC sends the amount **only for stackables** — an unstackable item
emits just 2 bytes (id, with the wielded flag in the high bit). Those 4 stolen
bytes were slot `i+1`'s id, so every slot after slot 0 was garbage (commit
`295e93d`, 2026-05-28). It took ~30 minutes of multi-layer debugging because
nothing in the system *complained* about an amount of 10,879,108
(`docs/lang/protocol-debug-strategy.md:80-84`).

**The use() that cast a spell (2026-05-28 → 05-29).** `use(item, npc)` shipped
wired to opcode **50** (commit `5e005e2`). Opcode 50 is `CAST_ON_NPC`;
`NPC_USE_ITEM` is **135**. Both payloads are byte-identical 4-byte
`[u16][u16]` frames, distinguished *only* by the opcode — so the wrong literal
silently turned every item-on-NPC interaction into a bogus spell-cast with the
inventory slot read as a spell ID (`docs/lang/protocol.md:437-445`;
`action/use.go:20`, `action/magic.go:14`). No error, no refusal: the server
just did a different thing.

**The hidden bronze pickaxe (2026-05-27 → 06-09).** The Phase-1 ground-item
decoder (opcode 99, commit `5431b33`) read **one entry per frame** — the code
even carried a "Phase 2 will return a slice" comment — and bailed to
`UnknownPacket` whenever the first byte was `0xFF`. But the server streams
**every in-view ground item per update** as a sequence of entries
(`GameStateUpdater.java:1128-1175`). For thirteen days, any item the server
happened to list second or later was invisible to perception. The flagship
casualty: the free Bronze Pickaxe (id 156) on the Barbarian-Village table —
hidden by the wire bug while Delores hunted "where do I buy a pickaxe?",
walking to Bob's Axes on a pure name-match in the episode that became the
epistemic-gap saga (`docs/lang/protocol.md:224-229`;
`docs/world-knowledge-and-learning.md` §1.1).

**Garbled and exploding PMs (two bites, 2026-05-27 and 06-07).** PMs first
arrived as garbage because the body was sent without its smart char-count
prefix — the server read the first ciphertext byte as the length and
decompressed bogus characters (fixed in `772434d`). Then they came back to
bite: the prefix was written as `len(message)` — Go *bytes* — but the RSC
Huffman codec emits one code per **rune**. The first time the LLM put a curly
quote or em-dash in a host's speech, byte-count exceeded char-count and the
server's `StringEncryption.decryptString` walked off the end of its buffer
(commit `4d58de4`, 2026-06-07).

**The fleet that bled to zero.** Long soaks quietly lost hosts: a routine
server idle-kick is a clean TCP EOF, the read loop stored **no error**,
`host.Run` returned nil, and the supervisor marked the host cleanly `Stopped`
— restart policy never fired. A multi-day soak silently bled the fleet to zero
(commit `0214a0b`, 2026-06-09; `session/conn.go:251-258`).

**Actions that vanished into the void.** Boundary interactions (doors, webs)
silently no-oped when the `(x, y, dir)` triple didn't match the server's
record exactly: `getWallObjectWithDir` returns null, the server calls
`setSuspiciousPlayer(true, …)`, and nothing comes back on the wire
(`docs/lang/protocol.md:279-285`). A "reasonable-looking" canonicalization
(aliasing a north edge to the south edge of the tile above) is wrong — there
is exactly one canonical address per boundary, and the server's is the only
one that works.

## The false leads

**False lead #1: "the weird value is localized — filter it."** The
10,879,108 amount looked like "we forgot to gate the amount read on
stackability, so we read junk *into the amount field*." That diagnosis is
seductive because it's half right — the stackability condition *was* the
trigger — but it frames the damage as one wrong field. The reality: the entire
offset chain was off-by-N and **every subsequent slot was corrupted**, which is
why a door interaction failed three layers up. "Looks reasonable" decoding —
clamp the outlier, keep the rest — would have shipped the cascade
(`docs/lang/protocol-debug-strategy.md:183-194`).

**False lead #2: trusting a payload shape to identify a packet.** Opcodes 50
and 135 carry indistinguishable bytes. There is no way to catch the swap from
the payload, the symptom, or "it seems to work" testing — the server accepts
both and acts on both. An early open question even floated that byte 50 might
be one opcode disambiguated by length; the research killed it: the
opcode→meaning map is **per-revision** (v235: 50/135; v198: 49/160; v196:
185/22 — always two distinct opcodes), and the only authority is the server's
parser case map (`docs/lang/protocol.md:437-448,863`).

**False lead #3: decode the first record, ship it, plan to finish later.** The
single-entry ground-item read worked in every early test — items usually
arrived one per frame near spawn. The deferred "Phase 2" multi-entry work
didn't fail loudly; it failed by *omission*, hiding world state for two weeks
in a way no log line could surface. Same family: punting the `0xFF` first-byte
case to `UnknownPacket` instead of resolving what it meant (it's a 3-byte
coord-only clear — `[255][offX][offY]`, **no id** — for out-of-range removals;
in-range removal is the `0x8000` high bit on the id;
`docs/lang/protocol.md:204-229,875`).

**False lead #4: self-healing past corruption.** OpenRSC's own decoder ships a
recovery loop that re-rolls up to **256 ISAAC states** hunting for a plausible
opcode (`RSCProtocolDecoder.java:180-212`) — institutionalized "keep going" in
the face of a desynced cipher. We deliberately refused to port it. A resync
loop converts a hard, diagnosable protocol bug into an undiagnosable stream of
*almost*-correct frames — the worst possible failure mode for a system whose
perception layer feeds an LLM (`docs/protocol.md:57-61`).

**False lead #5: trusting your own port by re-reading it.** Our ISAAC
implementation guards the keystream regeneration with a pre-decrement zero
check; the server's `getNextValue` uses a post-decrement (`ISAACCipher.java:
115-121`). Eyeballing the two for equivalence is exactly the kind of
off-by-one-phase judgment call that humans (and prior verification agents) get
wrong with confidence. An earlier "verified" finding in the same study nearly
introduced a real bug on the strength of pattern-matching a different dialect
(the 5-bit/6-bit near-miss — chapter 1).

**False lead #6: "no error" means "no problem" on teardown.** The read loop
treated EOF as the natural end of a connection — which it is, when *we* close.
The bug was failing to distinguish who initiated. Symmetrically, decode errors
could have been logged-and-skipped to keep the session alive; both are the
same instinct — smooth over the anomaly — and both turn infrastructure
failures into cognition mysteries.

## The determined fix

**Stackability-conditional decode + a facts seam.** `DecodeInbound` takes an
`isStackable func(itemID int) bool` callback — proto stays facts-free, the
runtime threads the lookup through — and `decodeInventory` consults it per
slot, with `readUnsignedShortIntSmart` mirroring the server's
`writeUnsignedShortInt` (2 or 4 bytes by value; commit `295e93d`;
`proto/v235/inbound.go:411`).

**Tier-1 anomaly assertions, permanent.** The same day, `proto/v235/anomaly.go`
landed (commit `5e005e2`): `flagAnomaly(field, value, plausibleMax, raw)` logs
a WARN with the field label, the decoded value, the plausible ceiling, **and
the hex of the raw bytes that produced it** — the smoking gun in the log line.
The plausible-range table is deliberately loose (item id 5000, coord 16384,
amount 2×10⁹): anything past it is a decode bug, not data. Wired at the two
inventory decoders (`inbound.go:445-446,478`) — guarding exactly the off-by-N
cascade — plus one assertion above the proto layer: the opcode-79 NPC
order-list desync WARN (`runtime/frame.go:110-122`). The companion playbook is
`docs/lang/protocol-debug-strategy.md`: capture raw bytes first, read the
server's *sender* for conditional writes, add a Tier-1 assertion so the next
break is loud, and — the heuristic worth quoting — **when a decoder produces a
single weird value, dump the next 50 bytes raw and recheck the wire shape; do
not clamp the value and move on.**

**Opcode 50/135 split, fixed and fenced.** `outUseItemOnNpc` corrected to 135
(commit `a72efc0`, 2026-05-29 — found by the wave-2 protocol research reading
the server's parser case map, *not* by debugging a live symptom). The
per-revision opcode table and the "byte-identical payloads" warning were
written into the protocol reference the same day (`604e6b9`;
`docs/lang/protocol.md:437-448`), and both constants now carry the
distinguishing comment in code (`action/use.go:20`, `action/magic.go:14`).

**Multi-entry ground items + the 0xFF clear.** `decodeGroundItem` rewritten to
read entries to payload exhaustion — 4-byte adds, `0x8000`-flagged in-range
removals, 3-byte `[255][x][y]` out-of-range clears — as step 1 of the
movement-design execution (commit `b24012c`, 2026-06-09;
`proto/v235/inbound.go:514`). The pinned regression test's decisive assertion
is literally the war story: *the Bronze Pickaxe (156) decodes even as the
second entry* (`proto/v235/ground_item_test.go:13-54`). The sibling opcode-211
bulk region-clear decoder (`decodeRemoveWorldEntity`, `inbound.go:568`) is
pinned in the same file — its offsets are signed **16-bit**, not the int8 of
the per-entity packets, because exceeding int8 range is the packet's reason to
exist.

**Desync-is-fatal, as policy.** Our `FrameDecoder` advances the ISAAC stream
exactly once per opcode and trusts sync; a decode error is stored and the
connection is torn down (`session/conn.go:235-237`), and the posture is
documented as policy: *if a decode error appears in the logs, treat it as a
real protocol bug to fix, not transient noise* (`docs/protocol.md:57-61`). The
Plutonium study's audit of this seam confirmed the residual risk — an inbound
ISAAC slip would today corrupt every subsequent opcode silently, since unknown
opcodes pass through unvalidated — and the chosen mitigation is a **validity
gate + loud diagnostic**, explicitly *not* an auto-resync
(`docs/_research/plutonium/findings.md:56-62`; tracked as `docs/TODO.md` MP-3).

**ISAAC differential verification.** The keystream was proven bit-identical to
the server's by re-deriving, not re-reading: a fresh line-for-line int32
transliteration of the *full* `ISAACCipher.java` (mirroring Java `>>` vs `>>>`
semantics), run 600 values deep across two regeneration boundaries against our
production implementation — **0 mismatches** — plus a 320-opcode round-trip.
The verifying agent's method note is the lesson: *"I did not trust the prior
agent's claim"* (`docs/_research/plutonium/findings.md:188-193`). The same
sweep byte-verified the rest of the crypto stack by hand: RSA modulus
DER-parsed from the server's pem, framing line-identical to
`RSCProtocolEncoderMain.java`, XTEA bit-identical, login layout consumed
field-for-field by `LoginPacketHandler.java` (`findings.md:47-52`). The
differential test itself was a throwaway (removed after); promoting it to a
permanent regression test rides MP-3 — until then, `proto/v235/isaac_test.go`
pins symmetry and determinism only.

**Server-EOF restart semantics.** A `closing` flag distinguishes local Close
from server hang-up; a server-initiated EOF now stores the `ErrServerClosed`
sentinel, so `host.Run` returns an error, the frame pump's exit cancels the
per-host ctx, and the supervisor's restart policy actually fires (commit
`0214a0b`; `session/conn.go:26,44,246-263`). Both behaviors are pinned by
loopback-socket tests: `TestServerEOFStoresErrServerClosed` and
`TestLocalCloseStaysClean` (`session/conn_eof_test.go`).

**PM/chat length prefix, twice-hardened.** The smart char-count prefix landed
in `772434d`; the rune-count fix (`len([]rune(message))`, never
`len(message)`) plus `sanitizeChat` — folding the LLM's habitual curly quotes,
em-dashes, ellipses to clean RSC-charset ASCII — landed in `4d58de4`, with a
regression test locking the rune-count contract on a deliberately multi-byte
message. The wire fact to keep: the prefix counts **characters as the Huffman
codec counts them**, and the codec emits one code per rune
(`docs/lang/protocol.md:760-767`).

**Exact-key boundary addressing.** The canonical rule is documented where the
verbs live: `.at(x,y,dir)` must echo the server's exact dir — no
canonicalization, no N/S aliasing; a wall between rows is *always* `(x, y, 0)`
on the higher tile, and the server's derived `WALL_SOUTH`/`WALL_WEST` flags
are internal bookkeeping, not addressable boundaries
(`docs/lang/protocol.md:271-285`, written in `604e6b9`). The empirical anchor:
the Varrock-east web is DoorDef 24 at `(208,547)` dir **2** — a `\` diagonal.

**Designing the trust ledger against the wire.** The v235 damage/death packets
carry **no attacker identity** — so "am I under attack, and by whom?" cannot
be a clean push event. This constraint shaped two layers. In the DSL,
`under_attack`-style conditions are polled `when`-predicates, not `on` events;
`damage_taken` fires with its `source` permanently empty
(`docs/lang/events.md:289-302`). In the limbic system (commits `af35919`,
`2daffee`), **all trust updates are attributed**: they fire only on signals
that carry a counterparty name on the wire (chat, PM, trade, duel), and melee
death is deliberately *not* mapped to a trust delta
(`runtime/limbic.go:194-198`). Wilderness-PK grievance uses the documented
attribution ceiling — engaged-player index, then last-attacked-player, gated
by the skull heuristic, never the NPC-contaminated combat target — and when
attribution is ambiguous it records **nothing**: *better silent than
mis-attributed, the cardinal constraint* (`runtime/limbic.go:163-178,511-522`).

## The durable rules

1. **Decoders fail loudly or they hide cascades.** Every decoded field that has
   a plausible ceiling gets a Tier-1 anomaly assertion — ~1 LOC at the call
   site buys back the half-hour the inventory bug cost, every time.
2. **One weird field means re-derive the frame.** Dump the next 50 raw bytes
   and recheck the wire shape against the server's *generator*; never clamp
   the value and move on. "Wrong but localized" is usually "wrong and
   cascading."
3. **Byte-identical payloads are still two opcodes.** Packet identity lives in
   the opcode→meaning map, which is per-revision; verify every opcode literal
   against the server's parser case map, never against "it seems to work."
4. **Read repeated records to payload exhaustion** unless the server source
   proves a single record. A single-entry read of a multi-entry frame is a
   silent perception filter — the worst bugs are the ones that hide data.
5. **Desync is fatal.** Recovery loops that resync past corruption ship
   corruption; tear down on decode error and fix the real bug. Make the
   failure observable (validity gate + diagnostics), not survivable.
6. **A clean EOF is not a clean shutdown.** Distinguish who closed the
   connection; a teardown path that stores no error converts an
   infrastructure failure into a slow fleet death.
7. **Echo the server's exact lookup keys.** `(x, y, dir)` matches exactly or
   the action silently fails (and flags you suspicious) — no canonicalization
   the server doesn't do.
8. **Length prefixes count what the codec counts** — runes for the RSC Huffman
   codec, never Go bytes. Sanitize LLM prose to the wire's charset before
   encoding.
9. **Verify crypto by differential test against a fresh transliteration of
   the server source** — not by re-reading your own port, and not by trusting
   a prior agent's verification.
10. **Design events against what the wire actually carries.** If identity
    isn't on the wire, don't synthesize it: attribute only from named signals,
    poll the rest, and abstain rather than misattribute.

## Sources

- Commits: `295e93d` (inventory off-by-N fix + isStackable seam, 2026-05-28),
  `5e005e2` (Tier-1 anomaly assertions; also *introduced* the opcode-50 bug,
  2026-05-28), `a72efc0` (opcode 135 fix, 2026-05-29), `604e6b9` (boundary
  dir table + opcode-50 conflation docs, 2026-05-29), `5431b33` (the Phase-1
  single-entry ground-item decoder, 2026-05-27), `b24012c` (multi-entry +
  0xFF clear + opcode 211, 2026-06-09), `772434d` (PM smart length prefix,
  2026-05-27), `4d58de4` (rune-count + sanitizeChat, 2026-06-07), `0214a0b`
  (ErrServerClosed + supervisor restart, 2026-06-09), `af35919` / `2daffee`
  (limbic trust attribution; duel ≠ PvP).
- `docs/lang/protocol.md` — the per-faculty opcode catalog: framing/ISAAC
  (§0), ground items + the pickaxe note (§1, lines 204-229), boundary
  exact-dir addressing (lines 251-285), opcode 50/135 (lines 437-448,
  573-574), PM gotcha (lines 760-767), resolved open questions (lines 863,
  875, 885).
- `docs/protocol.md` — transport reference; the desync-is-fatal posture
  (lines 57-61).
- `docs/lang/protocol-debug-strategy.md` — the failure-mode taxonomy, Tier-1
  mechanics, the debugging playbook, and the "looks reasonable isn't safe"
  heuristic.
- `docs/_research/plutonium/findings.md` — the byte-verification corpus
  (lines 47-52), the no-validity-gate finding + gate-not-resync
  recommendation (lines 56-62), the ISAAC differential proof (lines 188-193).
- Code: `proto/v235/anomaly.go`, `proto/v235/inbound.go:411,445-446,478,514,568`,
  `proto/v235/ground_item_test.go`, `proto/v235/isaac_test.go`,
  `action/use.go:20`, `action/magic.go:14`, `runtime/frame.go:110-122`,
  `session/conn.go:26,44,221-265`, `session/conn_eof_test.go`,
  `runtime/limbic.go:163-178,194-198,511-522`.
- Open work: `docs/TODO.md` MP-3 (opcode-validity gate + permanent ISAAC
  regression test), MP-12 (sleep/wire hardening, session-id race).
- Cross-references: [1_PATHING.md](1_PATHING.md) (the byte-oracle doctrine,
  the 5-bit near-miss); `docs/world-knowledge-and-learning.md` §1.1 (the
  epistemic saga the hidden pickaxe fed);
  `docs/lang/events.md` (the reflex/deliberate cleavage).
