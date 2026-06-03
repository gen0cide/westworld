package client.util;

/**
 * DecodeBuffer — shared decode-state object for the BZip2 sprite decompressor.
 *
 * <p>This class is a plain data-holder with two static utility methods that are
 * logically unrelated to BZip2 but live here for historical bundling reasons:
 *
 * <ul>
 *   <li><b>BZip2 decode state</b> — all the mutable fields ({@link #compressedInput},
 *       {@link #symbolFlags}, {@link #huffLimits}, etc.) are read and written
 *       exclusively by {@link client.scene.SpriteDecoder} ({@code ea}) during
 *       block-by-block BZip2 decompression.  The single static instance lives in
 *       SpriteDecoder; callers pass it by reference to every helper method.</li>
 *   <li><b>Linked-list node insert</b> — {@link #insertNodeBefore} splices a
 *       {@link client.net.StreamBase} ({@code ib}) node into a doubly-linked list
 *       used by the audio pipeline ({@link client.util.LinkedQueue} / {@code db})
 *       and the filter chain.</li>
 *   <li><b>Chat-character normalization</b> — {@link #normalizeChatChar} maps
 *       accented Latin-1/Unicode chars to their ASCII base equivalents and
 *       preserves the three chat-markup control chars {@code [ ] #} intact.
 *       Used by the chat-filter / word-filter path.</li>
 * </ul>
 *
 * <p>The three static arrays ({@link #stringPool}, {@link #chatFilterCache},
 * {@link #CHAT_MARKUP_CHARS}) are accessed across the client for chat and
 * string-pool bookkeeping.
 *
 * <p>Obfuscated class name: {@code ac} (rev ~233–235 Microsoft J++ build).
 */
public final class DecodeBuffer {

    // -------------------------------------------------------------------------
    // BZip2 bit-stream reader fields
    // (populated by SpriteDecoder.decompress before calling its helpers)
    // -------------------------------------------------------------------------

    /**
     * Compressed input byte array (the raw BZip2-compressed sprite data).
     * Obf: {@code q}
     */
    public byte[] compressedInput;

    /**
     * Read cursor into {@link #compressedInput}.
     * Obf: {@code a}
     */
    public int inputPos;

    /**
     * Decompressed output byte array (caller-allocated destination buffer).
     * Obf: {@code i}
     */
    public byte[] outputBuffer;

    /**
     * Write cursor into {@link #outputBuffer}.
     * Obf: {@code o}
     */
    public int outputPos = 0;

    /**
     * Number of output bytes still requested by the caller.
     * Decremented as bytes are written to {@link #outputBuffer}.
     * Obf: {@code y}
     */
    public int outputRemaining;

    /**
     * Bit accumulator for the MSB-first bit-stream reader.
     * New bytes are shifted in from {@link #compressedInput} on demand.
     * Obf: {@code e}
     */
    public int bitAccumulator;

    /**
     * Number of valid bits currently held in {@link #bitAccumulator}.
     * Obf: {@code p}
     */
    public int bitsAvailable;

    /**
     * Running count of compressed bytes consumed (wraps at 256 for block CRC
     * synchronisation).
     * Obf: {@code c}
     */
    public int bytesConsumedMod;

    /**
     * Total number of decompressed bytes written across all BZip2 blocks.
     * Obf: {@code G}
     */
    public int totalBytesOut;

    // -------------------------------------------------------------------------
    // BZip2 block header / symbol-table fields
    // -------------------------------------------------------------------------

    /**
     * Number of BZip2 blocks processed so far (starts at 1 per block in
     * SpriteDecoder).
     * Obf: {@code f}
     */
    public int blockCount;

    /**
     * BZip2 block originator pointer (the 3-byte end-of-stream or block-start
     * magic, packed into an int).  Used as the initial index into the sorted
     * block for BWT inversion.
     * Obf: {@code E}
     */
    public int blockOriginatorPtr;

    /**
     * BZip2 symbol-table group presence flags.  {@code symbolGroupPresent[g]}
     * is {@code true} when the {@code g}-th 16-symbol group contributes at
     * least one byte to this block's alphabet.
     * Obf: {@code v}
     */
    public boolean[] symbolGroupPresent = new boolean[16];

    /**
     * Per-symbol in-use flags.  {@code symbolFlags[b]} is {@code true} when
     * byte value {@code b} appears in the current block.  Built from
     * {@link #symbolGroupPresent} by SpriteDecoder.
     * Obf: {@code n}
     */
    public boolean[] symbolFlags = new boolean[256];

    /**
     * Number of distinct in-use symbols in the current block (computed from
     * {@link #symbolFlags}).
     * Obf: {@code K}
     */
    public int symbolCount;

    /**
     * Packed active-symbol byte table.  {@code activeSymbols[k]} is the
     * byte value of the {@code k}-th in-use symbol (0 ≤ k < {@link #symbolCount}).
     * Obf: {@code d}
     */
    public byte[] activeSymbols = new byte[256];

    // -------------------------------------------------------------------------
    // BZip2 Huffman-tree tables
    // (6 trees × 258 entries to cover up to 258 symbols including EOB/RLE)
    // -------------------------------------------------------------------------

    /**
     * Per-tree minimum code-bit-length.  {@code huffMinLen[t]} is the shortest
     * Huffman code in tree {@code t}.
     * Obf: {@code D}
     */
    public int[] huffMinLen = new int[6];

    /**
     * Per-tree, per-level upper limit values for the Huffman canonical decoder.
     * {@code huffLimits[t][l]} is the largest code of length {@code l} in
     * tree {@code t}.  The decoder walks levels until the current code
     * ≤ huffLimits[t][level].
     * Obf: {@code u}
     */
    public int[][] huffLimits = new int[6][258];

    /**
     * Per-tree, per-level base offsets into the symbol table.
     * {@code huffBase[t][l]} is subtracted from the code to produce an index
     * into {@link #huffSymbols}.
     * Obf: {@code t}
     */
    public int[][] huffBase = new int[6][258];

    /**
     * Per-tree symbol lookup table (post-canonical expansion).
     * {@code huffSymbols[t][i]} is the decoded symbol for entry {@code i} of
     * tree {@code t}.
     * Obf: {@code J}
     */
    public int[][] huffSymbols = new int[6][258];

    /**
     * Per-tree, per-symbol code-length table as read from the compressed stream.
     * {@code huffCodeLengths[t][s]} is the bit-length of the Huffman code for
     * symbol {@code s} in tree {@code t}.
     * Obf: {@code B}
     */
    public byte[][] huffCodeLengths = new byte[6][258];

    // -------------------------------------------------------------------------
    // BZip2 selector / Huffman-tree assignment fields
    // -------------------------------------------------------------------------

    /**
     * Sequence of Huffman-tree indices for the current block's selectors
     * (after MTF decoding).  {@code huffTreeSeq[i]} names which of the 6
     * trees to use for the {@code i}-th 50-symbol group.
     * Also reused as a general-purpose large byte buffer by the BZip2 path.
     * Obf: {@code j}
     */
    public byte[] huffTreeSeq = new byte[18002];

    /**
     * Raw MTF-encoded selector lengths as read from the stream (before
     * move-to-front expansion).
     * Obf: {@code s}
     */
    public byte[] selectorLengths = new byte[18002];

    // -------------------------------------------------------------------------
    // BZip2 MTF (move-to-front) ring buffer
    // -------------------------------------------------------------------------

    /**
     * Move-to-front ring buffer used during symbol decoding.  Holds the last
     * 4096 emitted byte values, partitioned into 16 groups of 256 for fast
     * MTF lookup.
     * Obf: {@code A}
     */
    public byte[] mtfRingBuffer = new byte[4096];

    /**
     * Start-of-group pointers into {@link #mtfRingBuffer}.
     * {@code mtfGroupStart[g]} is the index of the most-recently-used position
     * within group {@code g}.
     * Obf: {@code r}
     */
    public int[] mtfGroupStart = new int[16];

    // -------------------------------------------------------------------------
    // BZip2 frequency / BWT inversion fields
    // -------------------------------------------------------------------------

    /**
     * Frequency counts for each decoded symbol value (0–255) in the current
     * block.  Used to build the cumulative-frequency table for BWT inversion.
     * Obf: {@code m}
     */
    public int[] symbolFrequency = new int[256];

    /**
     * Cumulative-frequency table for BWT (Burrows-Wheeler) inversion.
     * {@code cumulativeFreq[b+1] = sum of symbolFrequency[0..b]}.
     * Size 257 to allow a sentinel at index 0.
     * Obf: {@code w}
     */
    public int[] cumulativeFreq = new int[257];

    // -------------------------------------------------------------------------
    // BZip2 RLE (run-length encoding) output-side state
    // -------------------------------------------------------------------------

    /**
     * RLE run count still to emit for the current run.  0 = no active run.
     * Obf: {@code F}
     */
    public int rleRunCount;

    /**
     * Most-recently-emitted byte value (used to detect RLE run continuations).
     * Obf: {@code g}
     */
    public byte rleCurrentByte;

    // -------------------------------------------------------------------------
    // BZip2 BWT traversal state (used during e() / bwtEmit())
    // -------------------------------------------------------------------------

    /**
     * Current position/node in the BWT follow-on chain (packed: high bits =
     * next index, low byte = current byte value in the chain).
     * Obf: {@code H}
     */
    public int bwtChainNode;

    /**
     * Number of BWT bytes emitted in the current block so far.
     * Obf: {@code L}
     */
    public int bwtBytesEmitted;

    /**
     * Current byte value being extracted from the BWT chain.
     * Obf: {@code h}
     */
    public int bwtCurrentByte;

    /**
     * Total number of decompressed bytes in the current block (set after the
     * Huffman pass completes).
     * Obf: {@code b}
     */
    public int blockSize;

    // -------------------------------------------------------------------------
    // Static / shared fields
    // -------------------------------------------------------------------------

    /**
     * Landscape face-flag / tile-property array, allocated by the model/landscape
     * def loader ({@code m}) to {@code SpriteScaler.numModels} ({@code ia.h})
     * entries — side-by-side with {@code da.N} and {@code qa.K} — and read by
     * {@link client.world.World} ({@code k}) when building terrain tile geometry
     * (e.g. {@code if (landscapeFaceFlags[idx - 1] != 0) ...} flagging a tile face).
     *
     * <p>NOTE: {@code k} is {@link client.world.World} (terrain/landscape), NOT
     * Scene — per the post-deob correction in docs/NAMING.md ({@code k}/{@code lb}
     * were swapped in the original map). Scene is {@code lb}.
     *
     * <p>This field has nothing to do with BZip2 decoding; it was placed in
     * {@code ac} by the obfuscator to avoid static-field name collisions across
     * the single default package.
     *
     * Obf: {@code static int[] l}
     */
    public static int[] landscapeFaceFlags;

    /**
     * Profiling counter incremented at each call to {@link #normalizeChatChar}.
     * Dead instrumentation artefact — never read for logic.
     * Obf: {@code C}
     */
    public static int normCallCount;

    /**
     * Profiling counter incremented inside {@link #insertNodeBefore}.
     * Dead instrumentation artefact — never read for logic.
     * Obf: {@code k}
     */
    public static int insertCallCount;

    /**
     * Global interned string pool (200 slots) shared across the client.
     * Used by the chat / word-filter sub-system.
     * Obf: instance field {@code z} (static String[] z = new String[200]).
     * NOTE: The name {@code z} also appears as the two private XOR-decode helpers
     * inside each class; this is the only *field* named {@code z} in {@code ac}.
     */
    public static String[] stringPool = new String[200];

    /**
     * Cached word-filter string array.  Nulled whenever a chat message is
     * submitted without a "quotation" flag (byte param ≠ 34 in
     * {@link #insertNodeBefore}), forcing a rebuild on next use.
     * Obf: {@code x}
     */
    public static String[] chatFilterCache;

    /**
     * The three chat-markup control characters used by the client's rich-text
     * chat renderer: {@code '['} (open colour/effect tag), {@code ']'} (close
     * tag), {@code '#'} (colour delimiter).
     * These are passed through unchanged by {@link #normalizeChatChar}.
     * Obf: {@code I}
     */
    public static final char[] CHAT_MARKUP_CHARS = new char[]{'[', ']', '#'};

    // -------------------------------------------------------------------------
    // XOR string-pool (error-message fragments, decoded at class-init time)
    // -------------------------------------------------------------------------

    /**
     * Error-message prefix/argument strings for the two static methods.
     * Decoded from XOR-obfuscated literals:
     *   M[0] = "ac.A("  — normalizeChatChar error prefix
     *   M[1] = "null"   — null-argument representation
     *   M[2] = "ac.B("  — insertNodeBefore error prefix
     *   M[3] = "{...}"  — non-null argument representation
     * Obf: {@code M}
     */
    private static final String[] ERROR_MSG_PARTS;

    static {
        // Decoded via the two z() helpers (1-char XOR key for len<2, then
        // mod-5 byte-key table {9,60,65,38,13} for the char[] pass).
        //   z(z("h_og%"))            -> "ac.A("
        //   z(z("gI-J"))             -> "null"
        //   z(z("h_od%"))            -> "ac.B("
        //   z(z("rop")) -> "{...}"
        // Decoded strings written directly as literals for readability.
        // The decodeXorChars/decodeXorString helpers below document the
        // obfuscator mechanism but are not called at runtime.
        ERROR_MSG_PARTS = new String[]{
            "ac.A(", // obf: z(z("h_og%"))
            "null",  // obf: z(z("gI-J"))
            "ac.B(", // obf: z(z("h_od%"))
            "{...}"  // obf: z(z("r\x12o\x08p"))
        };
    }

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /** Allocates all BZip2 decode buffers to their fixed sizes. */
    public DecodeBuffer() {
        // (all fields already initialized inline above; this matches the
        //  original constructor which re-assigned them with identical values)
    }

    // -------------------------------------------------------------------------
    // Static utility: doubly-linked list node insertion
    // -------------------------------------------------------------------------

    /**
     * Inserts {@code node} into the doubly-linked list immediately before
     * {@code insertBefore}.
     *
     * <p>The doubly-linked list is defined on {@link client.net.StreamBase}
     * ({@code ib}) via its fields {@code a} (next) and {@code e} (prev).
     * This insert pattern is shared by the audio filter chain
     * ({@link client.audio.FilterChain}/{@code va}) and the
     * {@link client.util.LinkedQueue}/{@code db}) queue.
     *
     * <p>The {@code flag} byte controls whether the chat-filter string cache
     * ({@link #chatFilterCache}) is invalidated.  When {@code flag == 34}
     * (ASCII {@code '"'}, the "quoted message" sentinel used by the chat path)
     * the cache is preserved; any other value clears it, forcing a rebuild.
     *
     * <p>If {@code node} already belongs to a list it is detached first
     * (via {@link client.net.StreamBase#unlinkSelf} with sentinel
     * {@code -27331}).
     *
     * <p>Obfuscated name: {@code ac.a(ib, byte, ib)}.
     *
     * @param node         the node to insert
     * @param flag         controls chatFilterCache invalidation (34 = preserve)
     * @param insertBefore the existing node before which {@code node} is placed
     */
    public static final void insertNodeBefore(
            client.net.StreamBase node,
            byte flag,
            client.net.StreamBase insertBefore) {
        // Detach from any existing list position first.
        // StreamBase.prev is the "prev" link; when non-null the node is in a list.
        if (null != node.prev) {                // obf ib.e -> StreamBase.prev
            node.unlinkSelf(-27331);            // obf ib.a(int) -> StreamBase.unlinkSelf; detach sentinel
        }

        // ++insertCallCount;  -- dead profiling counter (obf: k++), dropped

        // Invalidate the chat-filter string cache unless flag is the "quoted
        // message" sentinel (byte 34 = '"').
        if (flag != 34) {
            chatFilterCache = null;             // obf: x = null
        }

        // Standard doubly-linked list splice: insert node just before insertBefore.
        // StreamBase uses field 'next' (obf a) and field 'prev' (obf e).
        node.next = insertBefore;               // obf ib.a -> StreamBase.next   (node.next = insertBefore)
        node.prev = insertBefore.prev;          // obf ib.e -> StreamBase.prev   (node.prev = insertBefore.prev)
        node.prev.next = node;                  // obf ib.e.a -> .prev.next       (node.prev.next = node)
        node.next.prev = node;                  // obf ib.a.e -> .next.prev       (insertBefore.prev = node)
    }

    // -------------------------------------------------------------------------
    // Static utility: chat-character normalization
    // -------------------------------------------------------------------------

    /**
     * Normalises a single character for use in the client's chat filter/display
     * pipeline.
     *
     * <p>Rules (applied in order):
     * <ol>
     *   <li>Space (0x20), non-breaking space (0xA0), underscore (0x5F), and
     *       hyphen (0x2D) are all replaced with {@code '_'} (underscore).
     *   <li>The three chat-markup control characters {@code '['} (0x5B),
     *       {@code ']'} (0x5D), and {@code '#'} (0x23) are returned unchanged
     *       (they correspond to {@link #CHAT_MARKUP_CHARS}).</li>
     *   <li>Accented Latin-1 / Latin Extended-A code-points are mapped to their
     *       ASCII base letter (à→a, è→e, í→i, ò→o, ù→u, ç→c, ÿ→y, ñ→n,
     *       ß→b), covering both lower-case and upper-case variants.</li>
     *   <li>All other characters fall through to
     *       {@link Character#toLowerCase(char)}.</li>
     * </ol>
     *
     * <p>The {@code unusedAntiTamperParam} argument exists only as an anti-tamper
     * guard: the obfuscated build checks it against the constant {@code -194} at
     * entry and recursively calls itself with a dummy sentinel if the check fails.
     * The parameter carries no semantic value and has been dropped from the
     * logical signature.  (Obf: param named {@code param1}, check
     * {@code if (param1 != -194) ac.a('ﾆ', 87);})
     *
     * <p>Obfuscated name: {@code ac.a(char, int)}.
     *
     * @param c                      character to normalise
     * @param unusedAntiTamperParam  must equal {@code -194} in the obfuscated
     *                               build; ignored here (anti-tamper artefact)
     * @return normalised ASCII character
     */
    public static final char normalizeChatChar(char c, int unusedAntiTamperParam) {
        // ++normCallCount;  -- dead profiling counter (obf: C++), dropped

        // --- Whitespace / punctuation -> underscore ---
        if (c == ' ')    return '_'; // 0x20
        if (c == ' ') return '_'; // 0xA0 non-breaking space
        if (c == '_')    return '_'; // 0x5F
        if (c == '-')    return '_'; // 0x2D hyphen

        // --- Chat-markup control chars: pass through unchanged ---
        if (c == '[') return c; // 0x5B  -- chat effect open-tag
        if (c == ']') return c; // 0x5D  -- chat effect close-tag
        if (c == '#') return c; // 0x23  -- chat colour delimiter

        // --- Accented 'a' variants (Latin-1 + Latin Extended-A) ---
        if (c == 'à') return 'a'; // à
        if (c == 'á') return 'a'; // á
        if (c == 'â') return 'a'; // â
        if (c == 'ä') return 'a'; // ä
        if (c == 'ã') return 'a'; // ã
        if (c == 'À') return 'a'; // À
        if (c == 'Á') return 'a'; // Á
        if (c == 'Â') return 'a'; // Â
        if (c == 'Ä') return 'a'; // Ä
        if (c == 'Ã') return 'a'; // Ã

        // --- Accented 'e' variants ---
        if (c == 'è') return 'e'; // è
        if (c == 'é') return 'e'; // é
        if (c == 'ê') return 'e'; // ê
        if (c == 'ë') return 'e'; // ë
        if (c == 'È') return 'e'; // È
        if (c == 'É') return 'e'; // É
        if (c == 'Ê') return 'e'; // Ê
        if (c == 'Ë') return 'e'; // Ë

        // --- Accented 'i' variants ---
        if (c == 'í') return 'i'; // í
        if (c == 'î') return 'i'; // î
        if (c == 'ï') return 'i'; // ï
        if (c == 'Í') return 'i'; // Í
        if (c == 'Î') return 'i'; // Î
        if (c == 'Ï') return 'i'; // Ï

        // --- Accented 'o' variants ---
        if (c == 'ò') return 'o'; // ò  (obf: ~c3==-0xf3 i.e. ~(-243))
        if (c == 'ó') return 'o'; // ó  (obf: -244==~c3)
        if (c == 'ô') return 'o'; // ô  (obf: ~c3==-0xf5)
        if (c == 'ö') return 'o'; // ö  (obf: -247==~c3)
        if (c == 'õ') return 'o'; // õ
        if (c == 'Ò') return 'o'; // Ò
        if (c == 'Ó') return 'o'; // Ó
        if (c == 'Ô') return 'o'; // Ô
        if (c == 'Ö') return 'o'; // Ö
        if (c == 'Õ') return 'o'; // Õ

        // --- Accented 'u' variants ---
        if (c == 'ù') return 'u'; // ù
        if (c == 'ú') return 'u'; // ú
        if (c == 'û') return 'u'; // û
        if (c == 'ü') return 'u'; // ü
        if (c == 'Ù') return 'u'; // Ù
        if (c == 'Ú') return 'u'; // Ú
        if (c == 'Û') return 'u'; // Û
        if (c == 'Ü') return 'u'; // Ü

        // --- Cedilla 'c' variants ---
        if (c == 'ç') return 'c'; // ç
        if (c == 'Ç') return 'c'; // Ç

        // --- 'y' variants ---
        if (c == 'ÿ') return 'y'; // ÿ
        if (c == 'Ÿ') return 'y'; // Ÿ (Latin Extended-A)

        // --- Tilde 'n' variants ---
        if (c == 'ñ') return 'n'; // ñ
        if (c == 'Ñ') return 'n'; // Ñ

        // --- German sharp-s: transliterated to 'b' (client convention) ---
        if (c == 'ß') return 'b'; // ß -> 'b'  (not 'ss'; Jagex choice)

        // Default: fold to lower-case
        return Character.toLowerCase(c);
    }

    // -------------------------------------------------------------------------
    // Private XOR string-pool decode helpers
    // (These are per-class artefacts of the obfuscator; every obfuscated class
    //  has its own pair of z() helpers with different key constants.)
    // -------------------------------------------------------------------------

    /**
     * First XOR decode pass: converts a String literal to a char[].
     * If the array has only one element, XORs it with the single-char key
     * {@code 0x0D} ('\r').  Arrays of length ≥ 2 are returned as-is.
     * Obf: {@code private static char[] z(String)}
     */
    private static char[] decodeXorChars(String encoded) {
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char)(chars[0] ^ '\r'); // obf key: 0x0D
        }
        return chars;
    }

    /**
     * Second XOR decode pass: applies a mod-5 byte-key table to the char[]
     * and interns the result as a String.
     * Key table (index % 5): {9, 60, 65, 38, 13}.
     * Obf: {@code private static String z(char[])}
     */
    private static String decodeXorString(char[] chars) {
        for (int i = 0; i < chars.length; i++) {
            int key;
            switch (i % 5) {
                case 0:  key =  9; break; // 0x09
                case 1:  key = 60; break; // 0x3C
                case 2:  key = 65; break; // 0x41
                case 3:  key = 38; break; // 0x26
                default: key = 13; break; // 0x0D
            }
            chars[i] = (char)(chars[i] ^ key);
        }
        return new String(chars).intern();
    }
}
