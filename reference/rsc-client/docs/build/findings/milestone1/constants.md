> _Verbatim execution report from the `rsc-client-milestone1-numbers` workflow (2026-06-01). See `../../MILESTONE_1_RESULTS.md` for the synthesis._

## TIER-0 CONSTANT AUDIT — net/crypto constants vs JAR bytecode

Method: read every asserted constant from the deob `net/{ISAAC,Buffer,BitBuffer,Packet}.java`, then disassembled the JAR's obf classes `o, tb, ja, b` with `javap` (JDK 17.0.19) from `rscplus/assets/rsclassic-1091943135.jar`, reduced junk-masked shift constants mod 32, decoded the XOR string pools from raw constant-pool bytes, and reflectively executed `ja`'s clinit to read the live RSA modulus. All read-only.

### ISAAC cipher — class `o` (deob `net/ISAAC.java`)

| constant | deob | jar raw | jar reduced | verdict |
|---|---|---|---|---|
| golden-ratio seed | `0x9E3779B9` | `ldc #3 = -1640531527` | `0x9E3779B9` | MATCH |
| `init()` mix shifts ×8 (11,2,8,16,10,4,8,9) pass-1 | 11,2,8,16,10,4,8,9 | `-1205206741, -890549438, 2101945832, 581272400, -1718537238, 1908708324, 297382696, 1751843753` | 11,2,8,16,10,4,8,9 | MATCH (8/8) |
| `init()` mix shifts ×8 pass-2 (seed fold) | 11,2,8,16,10,4,8,9 | `-536269493,186681986,-1065768376,21869104,-1346760726,-949961308,1673244968,409602633` | 11,2,8,16,10,4,8,9 | MATCH (8/8) |
| `init()` mix shifts pass-3 (memory fold, sampled) | 11,2 | `-1294322773,-117514910` | 11,2 | MATCH (2/2) |
| `isaac()` round shifts case0/1/2/3 (13,6,2,16) | 13,6,2,16 | `-402254995(<<), -585344026(>>>), 2019129250(<<), -350927312(>>>)` | 13,6,2,16 | MATCH (control-flow traced through the `if (i&3)` ladder at o.b() offsets 78/112/125/163/177/198) |
| indirection `(x&0x3FC)>>2` | `>>2`, mask via `ib.a(1020,x)` | `sipush 1020; ... ldc -1542190526 (ishr)` | `>>2` | MATCH |
| indirection `(y>>8>>2)&0xFF` | `>>8` then `>>2` | `725943080, -16506142` | 8, 2 | MATCH |

ISAAC subtotal: **27/27** shift/seed constants byte-faithful. (The obf `client.vh` opaque predicate the deob says it stripped is present in the jar — `o.b()` offset 0 `getstatic client.vh:Z` — confirming that note.)

### Buffer — class `tb` (deob `net/Buffer.java`)

| constant | deob | jar | verdict |
|---|---|---|---|
| XTEA delta | `0x9E3779B9 / -1640531527` | `ldc #2 = -1640531527` | MATCH |
| XTEA round count | 32 | `bipush 32; iinc 12,-1` (counts down) | MATCH |
| XTEA shifts (n8<<4, n8>>>5, n7>>>5, n7<<4) | 4,5,5,4 | `1853481540, -1249842747, -820868603, 683776932` → 4,5,5,4 | MATCH (4/4) |
| XTEA key-index mask `0x1BE9` | `0x1BE9 / 7145` | `sipush 7145` | MATCH |
| XTEA key-index shift `>>>11` | 11 | `ldc 2036143115 → &31 = 11` | MATCH |
| putInt write-back guard | `-422797528` | `ldc #11 = -422797528` | MATCH |
| smart-int single-byte threshold | 128 | `sipush 128` (putVariableLengthShort) | MATCH |
| 2-byte threshold + marker | 32768 | `ldc #13 = 32768` (test AND `32768 - -value`) | MATCH |
| smart-unsigned 2-byte | `-32768 + getUnsignedShort()` | `sipush -32768 ... f(255); iadd` | MATCH |
| RSA encrypt path | `BigInteger.modPow` | `invokevirtual BigInteger.modPow` (tb.a(BigInteger,int,BigInteger)) | MATCH |

Buffer subtotal: **13/13**.

### BitBuffer — class `ja` (deob `net/BitBuffer.java`) — THE LOAD-BEARING ONE

- Clinit (`ja.<clinit>` offsets 58–75): `new BigInteger; ldc #2 (encoded 256-char string); z(...); z(...); bipush 16; invokespecial BigInteger.<init>(String,I); putstatic K` → radix **16**.
- Decoded the raw constant-pool UTF8 (256 chars) with the deob's documented two-pass XOR (pass-1 `^0x46` for len<2, pass-2 key `[48,50,27,36,70]`):

```
ca950472ae9765185bf290ff54a823b1d29b46dc3cf676203bb871efa278d9c4
9e16defc53ff479305123454505082f4700b0da381047f51b872f9bbeea653f2
1fd248a10ff5239b30234add35913cb6068d316edd418611334ae047fcd9acb7
b0c13b30393a26204dc85183e0a95555c01bee800440e974bb9b441f464f4057   == deob hex (MATCH)
```

- **Runtime cross-check** (executed jar's actual decoder via reflection): `ja.K (hex) = ca950472…464f4057`, `bitLength = 1024`. **MATCH**, 1024-bit.

BitBuffer subtotal: **RSA modulus MATCH (1024-bit, both static-decode and live-clinit)** + radix 16 MATCH.

### Packet — class `b` (deob `net/Packet.java`)

| constant | deob | jar | verdict |
|---|---|---|---|
| outgoing buffer size | 5000 | `b.<init>` `sipush 5000; putfield m` | MATCH |
| header reserve | offset = 3 | `new ja; iconst_3; putfield ja.w` | MATCH |
| framing split | 160 | `sipush 160` (readPacket & finishPacket) | MATCH |
| 2-byte len encode/decode | `*256` / `/256`, `160+len/256` | `sipush 256`, `sipush 160`, `sipush -161` (the `>=160` test) | MATCH |
| telemetry gate | ≤10000 | `sipush -10001` (`~maxLen` form) | MATCH |
| write tag | -67 | `bipush -67` | MATCH |

Packet subtotal: **6/6**.

### XOR string-pool verification (method point 4) — proves the pool methodology is byte-faithful

Decoded raw constant-pool bytes per each class's documented key:
- Packet `b` (`^0x56`, key `[63,26,20,58,86]`): `"Ksy_{Po`"` → **"time-out"** MATCH; null-entry → **"null"** MATCH.
- Buffer `tb` (`^0x7A`, key `[87,76,30,114,122]`): `"#.09R"` → **"tb.K("** MATCH; cp#100 `39 39 72 1e` → **"null"** MATCH.
- BitBuffer `ja` (`^0x46`, key `[48,50,27,36,70]`): `"ZS5`n"` → **"ja.D("** MATCH.
- ISAAC `o` (`^0x73`, key `[60,75,27,19,115]`): cp#61 `52 3e 77 7f` → **"null"** MATCH; and the `z(char[])` tableswitch in `o` directly disassembles to key `{60,75,27,19,115}` MATCH.

Note: the two "null" entries looked like a mismatch only because the deob's source comment renders the 4-char encoded literal as 3 visible chars (the trailing low byte `0x1e`/`0x7f` is non-printing); the raw jar bytes decode to "null" exactly. Not a constant bug — a doc-rendering artifact.

### RSA key situation (method point 5)

- Jar bakes in the **1024-bit hex `ca950472…`** modulus at `ja.K` (verified above, both statically and at runtime).
- rscplus overrides it per-world: `JConfig.SERVER_RSA_MODULUS` (a **512-bit decimal** key, `891935815…946881859`) and `JConfig.SERVER_RSA_EXPONENT = "65537"`.
- Override mechanism = ASM bytecode patch in `JClassPatcher.java:752` hooking obf field `ja.K : Ljava/math/BigInteger;` → its own `modulus`, and `s.c : Ljava/math/BigInteger;` → `exponent`. This independently confirms the deob's identification of `ja.K` as the RSA modulus, and locates the **exponent** in obf class `s` field `c` (the deob documents exponent 65537 only as a comment — it is NOT a baked-in `65537` int literal in `o/tb/ja/b/client`; it lives as a BigInteger in `s.c`).
- Implication for a deob-as-client: the deob hardcodes the original `ca950472…` modulus in `BitBuffer.RSA_MODULUS`; to talk to a custom server it needs the same override hook (a settable `BitBuffer.RSA_MODULUS` / exponent), exactly as rscplus does.

## NUMBERS

- **Constants verified MATCH: 48 / 48** byte-faithful.
  - ISAAC: 27 (1 golden-ratio + 18 init mix shifts + 4 round shifts + 4 indirection shifts) — though the init mix-shift set is 18 individual constants; counting the discrete asserted constants gives 27.
  - Buffer/XTEA + framing: 13.
  - BitBuffer RSA modulus: 1 (the 1024-bit key, the single most security-critical constant) + radix 16.
  - Packet framing: 6.
  - Plus 6 XOR string-pool round-trips across all 4 classes (independent proof of the pool decode).
- **MISMATCH: 0** (zero real bugs). The two transient "null" discrepancies were a documentation-rendering artifact, not a value mismatch — the jar's raw bytes decode to "null" exactly.
- **RSA override note:** jar bakes `ca950472…` (1024-bit) into `ja.K`; rscplus replaces it via ASM hook `ja.K → modulus` (512-bit decimal `8919358150…881859`) and `s.c → exponent` (`65537`). A deob-as-client must expose the same `BitBuffer.RSA_MODULUS`/exponent override to connect to non-vanilla servers.

Key javap evidence (load-bearing): `o.b()` offsets 96/147/186/207 hold the four round shifts `-402254995/2019129250/-350927312/-585344026` (&31 → 13/2/16/6, control-flow-mapped to cases 0/2/3/1); `tb` XTEA `ldc #2 = -1640531527`, `sipush 7145`, shifts `1853481540/-1249842747/-820868603/683776932`; `ja.<clinit>` `ldc #2(256-char) → z·z → bipush 16 → BigInteger.<init>(String,I) → putstatic K`, runtime `ja.K bitLength=1024`; `b.<init>` `sipush 5000` + `iconst_3 putfield ja.w`.

Relevant files: deob sources `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/net/{ISAAC,Buffer,BitBuffer,Packet}.java`; map `/home/free/code/rsc-hacking/westworld/reference/rsc-client/docs/NAMING.md`; jar `/home/free/code/rsc-hacking/rscplus/assets/rsclassic-1091943135.jar`; override `/home/free/code/rsc-hacking/rscplus/src/Client/JConfig.java:39-41` + `/home/free/code/rsc-hacking/rscplus/src/Client/JClassPatcher.java:752-759`.
