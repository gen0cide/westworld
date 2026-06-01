package client.data;

/**
 * BZip — Canonical Huffman codec used throughout the RSC client for
 * bit-packing and decompressing data streams (chat text, sprite image data,
 * and packet payloads).
 *
 * <p>An instance encodes <em>one</em> Huffman table built from a caller-supplied
 * array of per-symbol code lengths (one byte per symbol, length 0 means "no
 * code assigned").  The constructor builds two data structures from those
 * lengths:
 * <ol>
 *   <li>{@link #codeValues} – per-symbol, left-aligned 32-bit canonical code
 *       value, used by {@link #encode}.</li>
 *   <li>{@link #trieNodes} – a flat integer array representing a binary trie,
 *       used by {@link #decode}.  Each slot holds either a non-negative child
 *       index (internal node, followed on bit=1) or a negative leaf value
 *       {@code ~symbol} (decoded symbol is {@code ~entry}).</li>
 * </ol>
 *
 * <p>Navigation rule in {@link #trieNodes}: for each incoming bit,
 * <ul>
 *   <li>bit = 1 → {@code node = trieNodes[node]} (follow stored pointer)</li>
 *   <li>bit = 0 → {@code node++}              (advance to adjacent slot)</li>
 * </ul>
 * After moving, {@code trieNodes[node] < 0} signals a leaf.
 *
 * <p>The singleton instance used for chat/text encoding is held in
 * {@link client.scene.SurfaceImageProducer#bzipCodec} (obf {@code fb.a}).
 * A second instance for sprite data is constructed by
 * {@link client.ui.MessageList} (obf {@code wb.p}) with a 256-symbol
 * code-length table.
 *
 * <p>Oracle match: {@code BZLib} in mudclient204 — the decompress path here
 * corresponds to the Huffman symbol decode that BZLib performs after the
 * BWT/MTF stages.  The <em>encode</em> direction has no oracle equivalent
 * (the server-side encoder is not shipped with the client).
 *
 * <p>Obfuscated class name: {@code aa}.  Package assigned: {@code client.data}.
 */
public final class BZip {

    // -------------------------------------------------------------------------
    // Instance fields (Huffman codec state)
    // -------------------------------------------------------------------------

    /**
     * Per-symbol canonical code value, left-aligned in 32 bits.
     * {@code codeValues[symbol]} is the canonical code for {@code symbol},
     * shifted so the MSB of the code occupies bit 31.
     * Used only during encoding ({@link #encode}).
     * <p>obf: {@code aa.g}
     */
    private int[] codeValues;   // obf: g

    /**
     * Per-symbol code length in bits (0 = symbol has no code / not in table).
     * Stored as the original byte[] passed to the constructor.
     * <p>obf: {@code aa.i}
     */
    private byte[] codeLengths; // obf: i

    /**
     * Flat binary trie used during decoding.
     * Each entry is either:
     * <ul>
     *   <li>&gt;= 0 — internal node; value is the index of the '1'-child node.
     *       The '0'-child is always at {@code index + 1} (implicit adjacency).</li>
     *   <li>&lt; 0 — leaf node; decoded symbol = {@code ~entry}
     *       (i.e. {@code -(entry + 1)}, in the range 0–255).</li>
     * </ul>
     * Slot 0 is the trie root. Grows by doubling when more nodes are needed.
     * <p>obf: {@code aa.j}
     */
    private int[] trieNodes;    // obf: j

    // -------------------------------------------------------------------------
    // Static fields — NOT related to Huffman; piggy-backed globals used by
    // other classes because the obfuscator merged many statics here.
    // -------------------------------------------------------------------------

    /**
     * Profiling counter incremented at entry of {@link #encode}.
     * Dead telemetry — never read by game logic.
     * <p>obf: {@code aa.h}
     */
    public static int encodeCallCount;             // obf: h  (profiling counter)

    /**
     * Profiling counter incremented at entry of {@link #decode}.
     * Dead telemetry — never read by game logic.
     * <p>obf: {@code aa.a}
     */
    public static int decodeCallCount;             // obf: a  (profiling counter)

    /**
     * Profiling counter incremented at entry of {@link #nextPowerOfTwo}.
     * Dead telemetry.
     * <p>obf: {@code aa.e}
     */
    public static int nextPow2CallCount;           // obf: e  (profiling counter)

    /**
     * Scratch variable written only when the opaque predicate
     * {@code client.vh} is true (never at runtime). Effectively dead.
     * <p>obf: {@code aa.b}
     */
    public static int deadScratch;                 // obf: b

    /**
     * Screen-layout ratio constant: 114.  Used by {@link client.ui.Panel}
     * (obf {@code qa}) in an ISAAC coordinate scaling formula
     * ({@code ua.C * n4 / 114}), and by the world-viewport calculation.
     * The value 114 represents a fixed-point scale factor for RSC's 176×114
     * UI viewport height.
     * <p>obf: {@code aa.d}
     */
    public static int VIEWPORT_HEIGHT_RATIO = 114; // obf: d

    /**
     * Entity/player count limit; parsed at startup from the applet parameter
     * and used throughout {@link client.Mudclient} (obf {@code client}) to
     * size world-entity arrays.
     * <p>obf: {@code aa.l}
     */
    public static int entityLimit;                 // obf: l

    /**
     * Render visibility flag; toggled by {@link client.scene.Scene}
     * (obf {@code lb}) and tested by {@link client.scene.GameModel}
     * (obf {@code ca}).
     * <p>obf: {@code aa.f}
     */
    public static int renderFlag;                  // obf: f

    /**
     * Shared 100-slot player/NPC name string array, indexed by entity slot.
     * Populated and cleared by {@link client.Mudclient}.
     * <p>obf: {@code aa.k}
     */
    public static String[] entityNames = new String[100]; // obf: k

    /**
     * Entity presence/count array; allocated by {@link client.net.SocketFactory}
     * (obf {@code m}) to {@code na.e} elements and used as a per-entity flag
     * table in {@link client.Mudclient}.
     * <p>obf: {@code aa.c}
     */
    public static int[] entityFlags;               // obf: c

    // -------------------------------------------------------------------------
    // XOR-decoded string pool (used only in error messages; decoded at class load)
    //
    // Original obfuscated literals each pass through z(String)→char[] then
    // z(char[])→String using per-class XOR key table {86, 45, 51, 97, 12}:
    //   z[0]: "aa.<init>("   — constructor error prefix
    //   z[1]: "{...}"        — non-null array representation
    //   z[2]: "null"         — null array representation
    //   z[3]: "aa.B("        — encode() error prefix
    //   z[4]: "aa.A("        — decode() error prefix
    //   z[5]: "aa.C("        — nextPowerOfTwo() error prefix
    // -------------------------------------------------------------------------
    private static final String[] z = new String[]{
        "aa.<init>(",  // obf: z(z("7L\x1d]e8DG_$"))
        "{...}",       // obf: z(z("-\x03\x1dOq"))
        "null",        // obf: z(z("8X_\r"))
        "aa.B(",       // obf: z(z("7L\x1d#$"))
        "aa.A(",       // obf: z(z("7L\x1d $"))
        "aa.C("        // obf: z(z("7L\x1d\"$"))
    };

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /**
     * Builds a Huffman codec from a symbol code-length table.
     *
     * <p>Each element {@code codeLengthTable[symbol]} specifies how many bits
     * the canonical Huffman code for {@code symbol} should have.  A value of
     * {@code 0} means the symbol is not present in the alphabet and will be
     * silently skipped.
     *
     * <p>The constructor performs two passes over the length table:
     * <ol>
     *   <li><strong>Canonical code assignment</strong> — uses a running
     *       accumulator array {@code canonicalAcc[0..32]} (one slot per code
     *       length) to assign left-aligned 32-bit code values in standard
     *       canonical order (shorter codes come first, same-length codes are
     *       assigned in symbol order).  The result is stored in
     *       {@link #codeValues}.</li>
     *   <li><strong>Trie construction</strong> — walks the bits of each symbol's
     *       canonical code from MSB to LSB, threading through (and growing)
     *       {@link #trieNodes}.  For each bit:
     *       <ul>
     *         <li>bit = 0 → advance node pointer by 1 ({@code ++node})</li>
     *         <li>bit = 1 → follow stored child pointer
     *             ({@code node = trieNodes[node]}); allocate a new subtree root
     *             if the slot is uninitialized (zero).</li>
     *       </ul>
     *       After all bits, writes {@code ~symbol} into the final node slot.</li>
     * </ol>
     *
     * <p>obf: constructor {@code aa.<init>(byte[])}
     *
     * @param codeLengthTable  byte array; {@code codeLengthTable[s]} = code length
     *                         for symbol {@code s}, 0 means absent.
     */
    public BZip(byte[] codeLengthTable) {
        // obf artifact stripped: ++aa.a (profiling), outer try/catch error wrapper

        int numSymbols = codeLengthTable.length;
        this.codeValues  = new int[numSymbols];   // g
        this.codeLengths = codeLengthTable;        // i
        this.trieNodes   = new int[8];             // j — grows on demand

        // canonicalAcc[L] holds the current running canonical code value at
        // length L, left-aligned in 32 bits.  Slot 0 is unused (0-length codes
        // are skipped); slots 1–32 correspond to bit lengths 1–32.
        int[] canonicalAcc = new int[33]; // obf local: nArray

        int nextFreeNode = 0;  // high-water mark in trieNodes[]    (obf: n3)
        int symbol = 0;        // current symbol being processed     (obf: n4)

        while (symbol < numSymbols) {
            int codeLen = codeLengthTable[symbol]; // obf: by (loaded as byte, used as int length)

            if (codeLen != 0) {
                // --- Canonical code assignment ---

                // Bitmask for the position of the MSB of a code of length codeLen
                // in a 32-bit word (bit 31-codeLen set to 1).
                // Example: codeLen=3 → msb = 0x20000000 (bit 29).
                int msb = 1 << (32 - codeLen);  // obf: n10

                // Retrieve (and store in g[]) the current canonical code value
                // at this length level.
                int currentCode = canonicalAcc[codeLen]; // obf: n9
                this.codeValues[symbol] = currentCode;   // g[symbol] = canonical code

                // Compute the NEXT canonical code at this length level:
                int nextCode;  // obf: n8
                if ((msb & currentCode) == 0) {
                    // MSB position is still free — we need to propagate a carry
                    // through shorter-length accumulator slots.  Walk backwards
                    // through lengths < codeLen, merging adjacent codes.
                    int len = codeLen - 1;  // obf: n5
                    while (len >= 1 && currentCode == canonicalAcc[len]) {
                        // Bitmask for the MSB position at this shorter length
                        int shorterMsb = 1 << (32 - len);  // obf: n6
                        int slotVal    = canonicalAcc[len]; // obf: n7/n11

                        if ((slotVal & shorterMsb) != 0) {
                            // Carry consumed — borrow next value from the slot above
                            canonicalAcc[len] = canonicalAcc[len - 1];
                            break;
                        }
                        // Flip the bit: set MSB of this length, keep lower bits
                        canonicalAcc[len] = (shorterMsb | slotVal); // obf: d.a(n6, slotVal) — CacheFile.bitwiseOr is simply n2 | n3
                        len--;
                    }
                    nextCode = currentCode | msb;  // set the MSB for this length
                } else {
                    // MSB already set — use the value from the length-1 level as next
                    nextCode = canonicalAcc[codeLen - 1];
                }

                // Update accumulator at this length
                canonicalAcc[codeLen] = nextCode;

                // Propagate the update to longer lengths that were still sharing
                // the old value (they inherit the new base code).
                // NOTE: clean base uses `continue` (skip mismatched slots and keep
                // scanning to length 32), NOT `break` — a slot that diverged earlier
                // does not stop the scan of higher lengths.
                for (int l = codeLen + 1; l <= 32; l++) {
                    if (canonicalAcc[l] == currentCode) { // obf: if(~n9 != ~nArray[l]) continue; nArray[l]=n8;
                        canonicalAcc[l] = nextCode;
                    }
                }

                // --- Trie insertion ---
                // Walk the trie from the root (node 0), following the bits of
                // currentCode from MSB down to the (codeLen)th bit.

                int node = 0;  // obf: n5 (reused)
                for (int bitIdx = 0; bitIdx < codeLen; bitIdx++) {
                    // Extract bit bitIdx from the MSB end of currentCode.
                    // Integer.MIN_VALUE = 0x80000000; >>> bitIdx shifts it right unsigned.
                    int bitMask = Integer.MIN_VALUE >>> bitIdx;  // obf: n6

                    if ((bitMask & currentCode) == 0) {
                        // Bit is 0: '0'-child is the adjacent slot (implicit)
                        node++;
                    } else {
                        // Bit is 1: '1'-child is at the pointer stored in trieNodes[node]
                        if (this.trieNodes[node] == 0) {
                            // Slot uninitialized — allocate a new subtree root
                            this.trieNodes[node] = nextFreeNode;
                        }
                        node = this.trieNodes[node]; // follow 1-child
                    }

                    // Grow the trie array if the current node index is out of bounds
                    if (node >= this.trieNodes.length) {
                        int[] grown = new int[this.trieNodes.length * 2];
                        for (int i = 0; i < this.trieNodes.length; i++) {
                            grown[i] = this.trieNodes[i];
                        }
                        this.trieNodes = grown;
                    }
                }

                // Update high-water mark
                if (nextFreeNode <= node) {
                    nextFreeNode = node + 1;
                }

                // Store leaf: ~symbol (negative, decoded during decode as ~entry = symbol)
                this.trieNodes[node] = ~symbol;
            }

            symbol++;
        }
    }

    // -------------------------------------------------------------------------
    // Encode: symbol bytes → bit-packed Huffman codes
    // -------------------------------------------------------------------------

    /**
     * Encodes {@code count} symbol bytes from {@code inBuf[inPos..inPos+count)}
     * into a bit-packed Huffman code stream written into {@code outBuf} starting
     * at <em>bit</em> offset {@code outBase * 8}.
     *
     * <p>Each symbol is looked up in {@link #codeValues} and {@link #codeLengths};
     * its canonical Huffman code bits are packed MSB-first into {@code outBuf}.
     * The output buffer may span up to 5 bytes per symbol (for long codes).
     *
     * <p>Returns the number of output <em>bytes</em> written (ceiling of bits).
     *
     * <p>Call sites (after obf stripping):
     * <ul>
     *   <li>{@code u.a}: encodes a string into a {@code tb} (Buffer) for packet
     *       transmission — {@code buf.w += fb.a.encode(buf.F, buf.w, byArray, 0, len)}</li>
     * </ul>
     *
     * <p>obf: {@code aa.a(int, byte[], int, byte[], int, int)} — the {@code int}
     * dummy 6th parameter ({@code var6}) is an anti-tamper guard; it must be
     * ≥ 99 or a recursive self-call is made.  That guard is stripped here.
     *
     * @param count     number of symbols to encode (obf: var1, mutated to end)
     * @param outBuf    output byte buffer receiving bit-packed codes  (obf: var2)
     * @param outBase   byte offset in {@code outBuf} to start writing  (obf: var3)
     * @param inBuf     input symbol byte array                        (obf: var4)
     * @param inPos     start index in {@code inBuf}                   (obf: var5)
     * @return number of output bytes written from {@code outBase}
     */
    public final int encode(int count, byte[] outBuf, int outBase, byte[] inBuf, int inPos) {
        // obf artifacts stripped:
        //   boolean bl = client.vh;       (opaque predicate, always false)
        //   ++h;                          (profiling counter, dead)
        //   if (var6 < 99) this.a(-58, null, 69, null, -39, 22); (anti-tamper guard, dead)
        //   try{} catch(RuntimeException e){ throw i.a(e,...); } (unwrapped)

        ++encodeCallCount; // obf: ++h  (dead profiling counter; kept for field completeness)

        int endPos    = inPos + count;  // exclusive end in inBuf    (obf: var1 after +=)
        int bitPos    = outBase << 3;   // current bit offset in outBuf (obf: var8; outBase*8)
        int accumBits = 0;              // carry-over bits from previous symbol (obf: var24)

        while (inPos < endPos) {
            int symbol   = inBuf[inPos] & 0xFF; // unsigned symbol value (obf: var9)
            int code     = this.codeValues[symbol];  // left-aligned 32-bit code (obf: var10 = g[symbol])
            byte codeLen = this.codeLengths[symbol];  // code length in bits     (obf: var11 = i[symbol])

            // Validate: symbol must have a non-zero code length
            if (codeLen == 0) {
                throw new RuntimeException("" + symbol); // no code for this symbol
            }

            // --- Bit-pack the code into outBuf ---
            //
            // Current state: outBuf is written up to bit offset bitPos.
            //   bytePos   = bitPos / 8      = index of the byte currently being filled
            //   bitOffset = bitPos % 8      = how many bits are already used in that byte (0 = fresh)
            //
            // We need to write `codeLen` bits starting at bit `bitOffset` within byte `bytePos`.
            // This may span up to ceil((codeLen + bitOffset) / 8) bytes.

            int bytePos   = bitPos >> 3;               // current output byte index (obf: var12)
            int bitOffset = bitPos & 7;                // bits already filled in current byte (obf: var13)

            // If bitOffset == 0, the current byte is fresh and accumBits should be 0.
            // Otherwise keep the carry from the previous symbol.
            // Trick: (-bitOffset >> 31) is 0 when bitOffset==0, -1 (all ones) otherwise.
            accumBits &= (-bitOffset >> 31); // zero out if byte-aligned (obf: var25 = var24 & -var13>>31)

            // Last output byte index that will be written by this symbol
            int lastByte = bytePos + ((codeLen + bitOffset - 1) >> 3); // obf: var14

            bitPos += codeLen;  // advance global bit position

            // The code is stored left-aligned in a 32-bit int.
            // To align it with the current bit offset, shift right by (32 - bitOffset).
            // Then merge with any carry bits from the previous symbol.
            // Note: bitOffset is first incremented by 24 to use the upper 8 bits range.
            bitOffset += 24; // now bitOffset ∈ [24..31], so code>>>bitOffset gives top byte fragment

            // Write first output byte (merged with carry)
            accumBits = (accumBits | (code >>> bitOffset)); // obf: d.a(var25, code>>>bitOffset) — CacheFile.bitwiseOr = n2|n3
            outBuf[bytePos] = (byte) accumBits;

            if (bytePos < lastByte) {
                bytePos++;
                bitOffset -= 8;
                accumBits = code >>> bitOffset;
                outBuf[bytePos] = (byte) accumBits;

                if (bytePos < lastByte) {
                    bytePos++;
                    bitOffset -= 8;
                    accumBits = code >>> bitOffset;
                    outBuf[bytePos] = (byte) accumBits;

                    if (bytePos < lastByte) {
                        bytePos++;
                        bitOffset -= 8;
                        accumBits = code >>> bitOffset;
                        outBuf[bytePos] = (byte) accumBits;

                        if (bytePos < lastByte) {
                            // Fifth byte — code spans 5 output bytes (very long code + misaligned)
                            bytePos++;
                            bitOffset -= 8;
                            // Left-shift into top of next byte: code << (-bitOffset) = code << (8-bitOffset)
                            accumBits = code << -bitOffset;
                            outBuf[bytePos] = (byte) accumBits;
                        }
                    }
                }
            }

            inPos++;
        }

        // Return the number of output BYTES written from outBase.
        // (bitPos+7)/8 gives the ceiling byte index; subtract outBase.
        return ((bitPos + 7) >> 3) - outBase;
    }

    // -------------------------------------------------------------------------
    // Decode: bit-packed Huffman codes → symbol bytes
    // -------------------------------------------------------------------------

    /**
     * Decodes Huffman-coded bits from {@code inBuf[inBase..]} into
     * {@code outBuf[outPos..outPos+count)}, producing exactly {@code count}
     * decoded bytes (or until the output buffer is full).
     *
     * <p>Each input byte is consumed one bit at a time (MSB first), navigating
     * the trie stored in {@link #trieNodes}:
     * <ul>
     *   <li>bit = 1 → {@code node = trieNodes[node]}</li>
     *   <li>bit = 0 → {@code node++}</li>
     * </ul>
     * When {@code trieNodes[node] < 0}, a leaf has been reached and
     * {@code ~trieNodes[node]} (i.e. the original symbol) is written to
     * {@code outBuf}.  The node pointer resets to 0 (root) for the next symbol.
     *
     * <p>Returns the number of <em>input bytes consumed</em>.
     *
     * <p>Call sites:
     * <ul>
     *   <li>{@code client.java}: {@code buf.w += fb.a.decode(buf.F, byArray, n2, buf.w, -1, n4)}
     *       — decodes packet chat text.</li>
     * </ul>
     *
     * <p>obf: {@code aa.a(byte[], byte[], int, int, int, int)}
     * The 5th parameter ({@code n4}) is an anti-tamper guard (must equal -1).
     * Stripped here.
     *
     * @param inBuf   source of bit-packed Huffman data            (obf: byArray / param1)
     * @param outBuf  destination for decoded symbol bytes         (obf: byArray2 / param2)
     * @param outPos  start index in {@code outBuf}               (obf: n2 / param3)
     * @param inBase  start index in {@code inBuf}                (obf: n3 / param4; used as cursor)
     * @param count   number of decoded bytes to produce          (obf: n5 / param6)
     * @return number of input bytes consumed
     */
    public final int decode(byte[] inBuf, byte[] outBuf, int outPos, int inBase, int count) {
        // obf artifacts stripped:
        //   boolean bl = client.vh;              (opaque predicate)
        //   ++a;                                 (profiling counter)
        //   if (n5 == 0) return 0;               (early exit — kept)
        //   if (n4 != -1) this.a(105,...);       (anti-tamper guard, stripped)
        //   do { ... } while (!bl);              (!bl always true → just a while(true) with internal breaks)
        //   try{} catch(...){ throw i.a(...); }  (unwrapped)

        ++decodeCallCount; // obf: ++a  (dead profiling counter)

        if (count == 0) {
            return 0;
        }

        int outEnd  = outPos + count;  // exclusive end in outBuf (obf: n5 after +=n2)
        int inCursor = inBase;         // current position in inBuf (obf: n7 = n3 initially)
        int node     = 0;              // current trie node index, 0 = root (obf: n6)

        // Main decode loop: consume input bytes until output is full
        while (true) {
            int inputByte = inBuf[inCursor]; // read next input byte (obf: n9, signed)
            int nodeVal;                      // obf: n8 = this.j[node]

            // Process all 8 bits of inputByte, MSB first.
            // For each bit: navigate trie (1→follow pointer, 0→++node),
            //               then check if we landed on a leaf.

            // --- Bit 7 (0x80) ---
            if (inputByte < 0) {
                // Bit 7 is set (sign bit of signed byte): follow 1-child
                node = this.trieNodes[node];
            } else {
                // Bit 7 is clear: 0-child is adjacent
                node++;
            }
            nodeVal = this.trieNodes[node];
            if (nodeVal < 0) { // leaf
                outBuf[outPos++] = (byte) ~nodeVal; // emit decoded symbol
                if (outPos >= outEnd) return inCursor - inBase + 1;
                node = 0; // reset to trie root
            }

            // --- Bit 6 (0x40) ---
            if ((inputByte & 0x40) != 0) {
                node = this.trieNodes[node]; // bit=1: follow pointer
            } else {
                node++; // bit=0: adjacent 0-child
            }
            nodeVal = this.trieNodes[node];
            if (nodeVal < 0) {
                outBuf[outPos++] = (byte) ~nodeVal;
                if (outPos >= outEnd) return inCursor - inBase + 1;
                node = 0;
            }

            // --- Bit 5 (0x20) ---
            if ((inputByte & 0x20) != 0) {
                node = this.trieNodes[node];
            } else {
                node++;
            }
            nodeVal = this.trieNodes[node];
            if (nodeVal < 0) {
                outBuf[outPos++] = (byte) ~nodeVal;
                if (outPos >= outEnd) return inCursor - inBase + 1;
                node = 0;
            }

            // --- Bit 4 (0x10) ---
            if ((inputByte & 0x10) != 0) {
                node = this.trieNodes[node];
            } else {
                node++;
            }
            nodeVal = this.trieNodes[node];
            if (nodeVal < 0) {
                outBuf[outPos++] = (byte) ~nodeVal;
                if (outPos >= outEnd) return inCursor - inBase + 1;
                node = 0;
            }

            // --- Bit 3 (0x08) ---
            if ((inputByte & 0x08) != 0) {
                node = this.trieNodes[node];
            } else {
                node++;
            }
            nodeVal = this.trieNodes[node];
            if (nodeVal < 0) {
                outBuf[outPos++] = (byte) ~nodeVal;
                if (outPos >= outEnd) return inCursor - inBase + 1;
                node = 0;
            }

            // --- Bit 2 (0x04) ---
            if ((inputByte & 0x04) != 0) {
                node = this.trieNodes[node];
            } else {
                node++;
            }
            nodeVal = this.trieNodes[node];
            if (nodeVal < 0) {
                outBuf[outPos++] = (byte) ~nodeVal;
                if (outPos >= outEnd) return inCursor - inBase + 1;
                node = 0;
            }

            // --- Bit 1 (0x02) ---
            if ((inputByte & 0x02) != 0) {
                node = this.trieNodes[node];
            } else {
                node++;
            }
            nodeVal = this.trieNodes[node];
            if (nodeVal < 0) {
                outBuf[outPos++] = (byte) ~nodeVal;
                if (outPos >= outEnd) return inCursor - inBase + 1;
                node = 0;
            }

            // --- Bit 0 (0x01) ---
            if ((inputByte & 0x01) != 0) {
                node = this.trieNodes[node];
            } else {
                node++;
            }
            nodeVal = this.trieNodes[node];
            if (nodeVal < 0) {
                outBuf[outPos++] = (byte) ~nodeVal;
                if (outPos >= outEnd) return inCursor - inBase + 1;
                node = 0;
            }

            // Advance to next input byte
            inCursor++;
        }
        // Note: the Vineflower/CFR obfuscated source shows the loop as
        //   do { ... } while (!bl) with bl=client.vh (always false),
        // which is equivalent to an infinite while(true) with internal return-on-full.
        // The return below (dead code in original) reconstructs the final return
        // that the bytecode shows at offset 0x2b6: return -inBase + 1 + inCursor.
    }

    // -------------------------------------------------------------------------
    // Static utility: next power of two
    // -------------------------------------------------------------------------

    /**
     * Returns the smallest power of two that is ≥ {@code n}.
     *
     * <p>Standard bit-fill trick: decrement, fill all lower bits by cascading
     * OR-shifts (×1, ×2, ×4, ×8, ×16), then add 1.
     *
     * <p>Used by {@link client.audio.SourceLinePlayer} (obf {@code pb}) to
     * round audio buffer sizes up to a power of two.
     *
     * <p>The {@code boolean} parameter in the obfuscated source is a dummy
     * anti-tamper guard that writes to the dead {@link #deadScratch} field
     * when true (never at runtime since {@code client.vh} is always false).
     * It is dropped here.
     *
     * <p>obf: {@code static aa.a(int, boolean)}
     *
     * @param n value to round up (must be &gt; 0)
     * @return smallest power of two ≥ {@code n}
     */
    public static final int nextPowerOfTwo(int n) {
        // obf artifacts stripped:
        //   ++e;             (profiling counter, dead)
        //   if (bl) b = -4; (opaque guard, dead)
        //   try{} catch(...){ throw i.a(...); } (unwrapped)

        ++nextPow2CallCount; // obf: ++e  (dead profiling counter)

        n--;
        n |= n >>> 1;
        n |= n >>> 2;
        n |= n >>> 4;
        n |= n >>> 8;
        n |= n >>> 16;
        return n + 1; // = n - (-1) in original (n - -1 == n+1)
    }

    // -------------------------------------------------------------------------
    // XOR string pool helpers (class-private; used only for error messages)
    // -------------------------------------------------------------------------

    /**
     * Step 1 of XOR string pool decode: converts string to char array,
     * XOR-ing the single character with 0x0C if length &lt; 2.
     * obf: {@code static aa.z(String)}
     */
    private static char[] z(String s) {
        char[] arr = s.toCharArray();
        if (arr.length < 2) {
            arr[0] = (char)(arr[0] ^ '\f'); // '\f' = 0x0C
        }
        return arr;
    }

    /**
     * Step 2 of XOR string pool decode: XORs each character with a
     * position-dependent key from the 5-element table {86, 45, 51, 97, 12}
     * (indices 0..4 of i%5), then interns the result.
     * obf: {@code static aa.z(char[])}
     */
    private static String z(char[] arr) {
        // Key table: index % 5 → XOR byte
        // {0→86 ('V'), 1→45 ('-'), 2→51 ('3'), 3→97 ('a'), 4→12 ('\f')}
        for (int i = 0; i < arr.length; i++) {
            int key;
            switch (i % 5) {
                case 0:  key = 86; break;
                case 1:  key = 45; break;
                case 2:  key = 51; break;
                case 3:  key = 97; break;
                default: key = 12; break;
            }
            arr[i] = (char)(arr[i] ^ key);
        }
        return new String(arr).intern();
    }
}
