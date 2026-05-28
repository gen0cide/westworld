# Protocol decode debug + profile strategy

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

## Detection strategies (build these as we go)

### Tier 1 — anomaly assertions in the decoder

Add cheap range assertions inline. When a field decodes to a value
that's *clearly* impossible (item amount > 28 for an unstackable;
coords > 1000 in the Lumbridge sector; npc index > 5000), log a
WARN with the raw bytes that produced it. The next person reading
logs sees the smoking gun without having to hand-decode.

```go
// Example: in decodeInventory
if amount > 65535 && !isStackable(itemID) {
    log.Warn("inventory decode anomaly",
        "slot", i, "id", itemID, "amount", amount,
        "raw", hex.EncodeToString(payload[start:b.rpos]),
    )
}
```

Cost: ~5 LOC per check. The bug-find time saved is enormous —
the inventory bug took 30 minutes of multi-layer debugging because
nothing in our system *complained* about a field decoding to
10879108.

### Tier 2 — Ground-truth pcap comparison

OpenRSC server already writes pcaps per player session
(`pcap for Delores at ... Delores`). The Java client's decoder is
the reference. Build a `proto/conformance/` test that:

1. Loads a server-side pcap of a known session (e.g. login +
   spawn key + use key).
2. Runs our decoder over each frame in order.
3. Asserts the produced events match a checked-in golden file
   of expected event shapes.

Updating the golden when wire format intentionally changes is a
deliberate diff review. Silent regressions get caught.

### Tier 3 — Side-by-side mode

Run our cradle in "shadow" mode against the same server connection
as a real Java client. Compare decoded events frame-by-frame.
Where Java and ours diverge, dump both interpretations.

Heavier setup but invaluable for unexplored opcodes (the magic
spell packets, the trade state machine, the bank UI).

### Tier 4 — Fuzzing the decoders

For each opcode handler in `proto/v235/inbound.go`, generate
random payloads and check the decoder doesn't panic / over-read.
Catches buffer-bound bugs but not interpretation bugs. Lowest
ROI; do last.

## Profile / observability hooks

When debugging a "wrong-field" bug in the field:

- `cradle -v` prints DEBUG-level slog including raw packet bytes
  (currently only on errors; should add `-v` toggle for every
  inbound frame's `(opcode, len, hex)` triple).
- Add a `delos` event stream marker for "decoded event of type
  X" so we can grep traces for specific opcodes.
- For ad-hoc debugging: a `cradle -dump-opcode=NNN` flag that
  prints raw + decoded for every frame of that opcode.

These are small tactical additions; none blocks current work,
but each saves an hour next time a decoder bug bites.

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
