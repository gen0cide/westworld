package client.ui;

import client.net.ChatCipher;   // v  — sentinel instance used by connection-state checks
import client.data.DataStore;   // nb — supplies WIN1252_C1_MAP (nb.f), WIN1252 extra chars

/**
 * CharTable — accented/special-character table and text-input validation helper.
 *
 * <p>Three responsibilities:
 * <ol>
 *   <li>{@link #ALLOWED_CHARS} — a 53-entry {@code char[]} listing every non-alphanumeric
 *       character that the game considers valid for text input (space, underscore, dash, and a
 *       full European accented-letter set in ISO 8859-1 order).  {@code RecordLoader.isValidChar}
 *       (obf: {@code f.a}) searches this array starting at a caller-supplied {@code startIndex}
 *       so that different UI fields can restrict the subset of allowed specials.</li>
 *   <li>{@link #itemDescriptions} — a lazily-allocated {@code String[]} of per-item examine
 *       strings, sized to the item-count loaded at startup and populated from the server-data
 *       init packet.  Used by {@code Mudclient} to display examine text and to label inventory
 *       tooltips alongside the item name.</li>
 *   <li>{@link #decodeBytes} — a Windows-1252 → Java-{@code String} decoder used by
 *       {@code Buffer.readString} (obf: {@code tb}) to convert raw packet bytes into displayable
 *       chat/name strings.  Null bytes act as string terminators (they are skipped), bytes in
 *       the C1 range 0x80–0x9F are remapped through {@link DataStore#WIN1252_C1_MAP}, and all
 *       other byte values pass through as ISO 8859-1 code points (which are identical to their
 *       Unicode scalars).</li>
 * </ol>
 *
 * <p>Obfuscated class name: {@code ga} (rev ~233–235 J++ build).
 *
 * <p>Oracle cross-reference: no direct equivalent in mudclient204; closest analogues are
 * {@code GameData.itemDescription[]}, {@code Utility.isValidChar()}, and the
 * Windows-1252 decode in {@code Buffer.readString()}.
 */
public final class CharTable {

    // -------------------------------------------------------------------------
    // Fields
    // -------------------------------------------------------------------------

    /**
     * Cipher sentinel used by the connection-state registry in {@code SpriteScaler.isKnownCipher}
     * (obf: {@code ia.a(v,byte)}) to identify CharTable's cipher as a live connection.
     * Constructed with the decoded strings "WTQA", "office", "_qa" and priority 2.
     * (The string arguments are stored in the {@code ChatCipher} but appear unused in its
     * actual cipher logic; they may be a debug/identification tag.)
     *
     * // obf: ga.c
     * // XOR-decoded ctor args: z(z("D\F")) = "WTQA", z(z("=vknw7")) = "office",
     * //                         z(z("\ral"))      = "_qa"
     */
    // Initialized in the static block below (after _errorFragments are decoded).
    // Decoded ctor args: "WTQA", "office", "_qa" — identification tag; priority 2.
    public static ChatCipher cipher; // obf: c

    /**
     * Dead profiling counter incremented once per {@link #decodeBytes} call.
     * Always-zero at runtime; never read for any game logic.
     * // obf: ga.d
     */
    public static int _profilingCounter; // obf: d

    /**
     * Table of non-alphanumeric characters accepted by the game's text-input validator.
     *
     * <p>Layout (53 entries, indices 0–52):
     * <pre>
     *  [0]  ' '  (SPACE, U+0020)
     *  [1]  ' '  (NON-BREAKING SPACE, U+00A0)  ← second space slot; different Unicode point
     *  [2]  '_'
     *  [3]  '-'
     *  [4]  'à'  [5]  'á'  [6]  'â'  [7]  'ä'  [8]  'ã'   — a-variants (lower)
     *  [9]  'À'  [10] 'Á'  [11] 'Â'  [12] 'Ä'  [13] 'Ã'   — A-variants (upper)
     *  [14] 'è'  [15] 'é'  [16] 'ê'  [17] 'ë'              — e-variants (lower)
     *  [18] 'È'  [19] 'É'  [20] 'Ê'  [21] 'Ë'              — E-variants (upper)
     *  [22] 'í'  [23] 'î'  [24] 'ï'                         — i-variants (lower)
     *  [25] 'Í'  [26] 'Î'  [27] 'Ï'                         — I-variants (upper)
     *  [28] 'ò'  [29] 'ó'  [30] 'ô'  [31] 'ö'  [32] 'õ'   — o-variants (lower)
     *  [33] 'Ò'  [34] 'Ó'  [35] 'Ô'  [36] 'Ö'  [37] 'Õ'   — O-variants (upper)
     *  [38] 'ù'  [39] 'ú'  [40] 'û'  [41] 'ü'              — u-variants (lower)
     *  [42] 'Ù'  [43] 'Ú'  [44] 'Û'  [45] 'Ü'              — U-variants (upper)
     *  [46] 'ç'  [47] 'Ç'                                    — cedilla
     *  [48] 'ÿ'  [49] 'Ÿ'                                    — y-umlaut
     *  [50] 'ñ'  [51] 'Ñ'                                    — n-tilde
     *  [52] 'ß'                                               — German sharp-s
     * </pre>
     *
     * <p>Callers pass a {@code startIndex} to slice the array: e.g. {@code startIndex=2} skips
     * the two space entries, restricting acceptance to underscore, dash, and the accented set.
     * {@code RecordLoader.isValidChar} (obf: {@code f.a}) performs a linear scan from
     * {@code startIndex} to the end of this array.
     *
     * // obf: ga.a
     */
    public static char[] ALLOWED_CHARS = new char[]{  // obf: a
        /* [0]  */ ' ',
        /* [1]  */ ' ', // NON-BREAKING SPACE (displayed as ' ' but different code point)
        /* [2]  */ '_',
        /* [3]  */ '-',
        /* [4]  */ 'à', /* [5]  */ 'á', /* [6]  */ 'â', /* [7]  */ 'ä', /* [8]  */ 'ã',
        /* [9]  */ 'À', /* [10] */ 'Á', /* [11] */ 'Â', /* [12] */ 'Ä', /* [13] */ 'Ã',
        /* [14] */ 'è', /* [15] */ 'é', /* [16] */ 'ê', /* [17] */ 'ë',
        /* [18] */ 'È', /* [19] */ 'É', /* [20] */ 'Ê', /* [21] */ 'Ë',
        /* [22] */ 'í', /* [23] */ 'î', /* [24] */ 'ï',
        /* [25] */ 'Í', /* [26] */ 'Î', /* [27] */ 'Ï',
        /* [28] */ 'ò', /* [29] */ 'ó', /* [30] */ 'ô', /* [31] */ 'ö', /* [32] */ 'õ',
        /* [33] */ 'Ò', /* [34] */ 'Ó', /* [35] */ 'Ô', /* [36] */ 'Ö', /* [37] */ 'Õ',
        /* [38] */ 'ù', /* [39] */ 'ú', /* [40] */ 'û', /* [41] */ 'ü',
        /* [42] */ 'Ù', /* [43] */ 'Ú', /* [44] */ 'Û', /* [45] */ 'Ü',
        /* [46] */ 'ç', /* [47] */ 'Ç',
        /* [48] */ 'ÿ', /* [49] */ 'Ÿ', // Ÿ — capital Y with diaeresis
        /* [50] */ 'ñ', /* [51] */ 'Ñ',
        /* [52] */ 'ß'  // German sharp-s / eszett
    };

    /**
     * Per-item examine/description strings, parallel to the item-name array
     * ({@code DecodeBuffer.itemNames}, obf: {@code ac.x}).
     *
     * <p>Allocated to {@code itemCount} (obf: {@code gb.p}) entries during startup by the
     * data-init code in {@code SocketFactory} (obf: {@code m}).  Each entry is set either from
     * the server data packet or to an empty string default.  Read by {@code Mudclient} when the
     * player examines an item or when an inventory tooltip is rendered:
     * {@code itemName + separator + itemDescriptions[itemId]}.
     *
     * // obf: ga.b
     */
    public static String[] itemDescriptions; // obf: b

    // -------------------------------------------------------------------------
    // XOR string pool (private, only used at class-init time)
    // -------------------------------------------------------------------------

    /**
     * Per-class XOR-encrypted error-message fragments, decoded once in the static initialiser.
     * Decodes (mod-5 byte-key table, keys [82,16,13,7,20]) to:
     *   z[0] = "null"    — null-argument placeholder in exception signatures
     *   z[1] = "{...}"   — non-null array/object placeholder
     *   z[2] = "ga.A("   — method signature prefix for decodeBytes error reporting
     *
     * // obf: ga.z
     */
    private static final String[] _errorFragments; // obf: z

    // -------------------------------------------------------------------------
    // Static initialiser
    // -------------------------------------------------------------------------

    static {
        // XOR-decode the per-class error-message string pool.
        // Original obf: z = new String[]{z(z("<eak")), z(z(")>#)i")), z(z("5q#F<"))};
        //   z[0] = "null"    -- null-argument placeholder in exception signatures
        //   z[1] = "{...}"   -- non-null array/object placeholder
        //   z[2] = "ga.A("   -- method-signature prefix used when rethrowing from decodeBytes
        _errorFragments = new String[]{"null", "{...}", "ga.A("};

        // Construct the ChatCipher sentinel for this class.
        // Original obf: c = new v(z(z("\u0005D\\F")), z(z("=vknw7")), z(z("\ral")), 2)
        // XOR-decoded args: "WTQA", "office", "_qa" -- identification tag strings; priority 2.
        cipher = new ChatCipher("WTQA", "office", "_qa", 2);
    }

    // -------------------------------------------------------------------------
    // Public API
    // -------------------------------------------------------------------------

    /**
     * Decodes a sequence of raw packet bytes into a Java {@code String} using Windows-1252
     * encoding rules.
     *
     * <p>Algorithm:
     * <ol>
     *   <li>Allocate a {@code char[length]} output buffer.</li>
     *   <li>For each byte at {@code buffer[startOffset + i]} (i in [0, length)):
     *       <ul>
     *         <li>Read the byte as an unsigned value (0–255).</li>
     *         <li>If the value is {@code 0} (null terminator), <em>skip</em> it — do not
     *             emit a character.  This lets the caller pass a fixed-size buffer that may
     *             have trailing null padding.</li>
     *         <li>If the value is in the Windows-1252 C1 range {@code 0x80–0x9F}: look up
     *             the Unicode mapping in {@link DataStore#WIN1252_C1_MAP}
     *             ({@code DataStore.f[byteVal - 128]}).  If the mapping is {@code '\0'}
     *             (the slot is undefined in Windows-1252), substitute {@code '?'} (U+003F)
     *             to avoid embedding a null character in the result.</li>
     *         <li>Otherwise pass the byte value through directly as a {@code char} — ISO 8859-1
     *             code points 0x01–0x7F and 0xA0–0xFF map 1-to-1 to Unicode.</li>
     *       </ul>
     *   </li>
     *   <li>Return {@code new String(outputBuf, 0, writeIndex)}.</li>
     * </ol>
     *
     * <p>Called by:
     * <ul>
     *   <li>{@code Buffer.readString} (obf: {@code tb}) — decodes null-terminated strings
     *       from the packet stream: {@code decodeBytes(length, junkSeed, startOffset, this.F)}</li>
     *   <li>{@code Mudclient} — decodes a compressed/stored string field similarly.</li>
     * </ul>
     *
     * @param length      number of bytes to decode (and maximum output chars)
     * @param _junkSeed   anti-tamper dummy parameter; only appears in unreachable dead-code
     *                    division after the loop; has no effect on the returned value.
     *                    Callers pass {@code (byteParam - 68)} or {@code (n2 ^ 0xFFFFFF84)}.
     * @param startOffset index into {@code buffer} of the first byte to decode
     * @param buffer      raw packet byte array
     * @return decoded String (length 0 to {@code length})
     *
     * // obf: ga.a(int,int,int,byte[])
     */
    public static final String decodeBytes(int length, int _junkSeed, int startOffset, byte[] buffer) { // obf: a
        // Dead obfuscation artifact stripped:
        //   boolean bl = client.vh;  // opaque predicate, always false
        //   ++d;                     // profiling counter, ignored
        //
        // Anti-tamper junk AFTER the loop (unreachable; stripped):
        //   n6 = -103;
        //   n5 = (-63 - _junkSeed) / 49;
        //   n7 = n6 / n5;   // would cause div-by-zero if reached
        //
        // Exception wrapper stripped:
        //   catch (RuntimeException e) { throw ErrorHandler.a(e, "ga.A(" + length + "," + ...); }

        char[] outputBuf = new char[length];
        int writeIdx = 0;

        for (int i = 0; i < length; i++) {
            // Read byte as unsigned (0–255).
            // Bytecode: byArray[param2 - (-n7)] = byArray[startOffset + i]
            int byteVal = buffer[startOffset + i] & 0xFF;

            // Skip null bytes — they act as implicit string padding/terminators.
            if (byteVal == 0) {
                continue;
            }

            // Windows-1252 C1 range (0x80–0x9F): these are NOT defined in ISO 8859-1;
            // Windows-1252 maps them to printable characters (€, ‚, ƒ, „, …, †, ‡, ˆ,
            // ‰, Š, ‹, Œ, Ž, ', ', ", ", •, –, —, ˜, ™, š, ›, œ, ž, Ÿ).
            // DataStore.WIN1252_C1_MAP (nb.f[]) stores the Unicode scalar for each.
            if (byteVal >= 0x80 && byteVal <= 0x9F) {
                // Lookup in the 32-entry C1 translation table.
                char mapped = DataStore.WIN1252_C1_MAP[byteVal - 0x80]; // obf: nb.f[byteVal - 128]
                if (mapped == '\0') {
                    // Undefined slot in the Windows-1252 C1 block (e.g., 0x81, 0x8D, 0x8F,
                    // 0x90, 0x9D) — substitute '?' rather than embedding a null character.
                    // Bytecode: if (~n10 == -1) n10 = 63;  (~0 == -1, so mapped=='\0')
                    mapped = '?';
                }
                byteVal = mapped;
            }
            // Bytes 0x01–0x7F and 0xA0–0xFF are ISO 8859-1 / Latin-1; their values are
            // identical to their Unicode code points, so no translation is needed.

            outputBuf[writeIdx++] = (char) byteVal;
        }

        return new String(outputBuf, 0, writeIdx);
    }

    // -------------------------------------------------------------------------
    // Private XOR string-pool helpers (obfuscator-generated; stripped from callers)
    // -------------------------------------------------------------------------

    /**
     * Pass-1 of the two-pass XOR string decoder.  Converts the raw literal {@code String}
     * into a mutable {@code char[]}.  If the string has fewer than 2 characters, XORs the
     * single character with {@code 0x14} (20).  For strings of 2+ chars the array is returned
     * unchanged — the real XOR work happens in {@link #_decodePool}.
     *
     * // obf: ga.z(String) -> char[]
     */
    private static char[] _encodeStep(String raw) { // obf: z(String)
        char[] chars = raw.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 0x14);
        }
        return chars;
    }

    /**
     * Pass-2 of the two-pass XOR string decoder.  XORs each character with a key selected by
     * {@code index % 5}:
     * <pre>
     *   index % 5 == 0 → key 82  (0x52)
     *   index % 5 == 1 → key 16  (0x10)
     *   index % 5 == 2 → key 13  (0x0D)
     *   index % 5 == 3 → key  7  (0x07)
     *   index % 5 == 4 → key 20  (0x14)
     * </pre>
     * Returns the decoded string interned so identical literals share one instance.
     *
     * // obf: ga.z(char[]) -> String
     */
    private static String _decodePool(char[] chars) { // obf: z(char[])
        final int[] KEY = {82, 16, 13, 7, 20};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char) (chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
