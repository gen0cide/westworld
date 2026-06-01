package client.data;

/**
 * NameHash — JAG archive entry lookup by name-hash.
 *
 * Holds the static tables that let the engine locate a named file inside a
 * binary JAG archive (the same format used for sound/sprite/model caches).
 * The key operation is {@link #getFileOffset}: it uppercases the requested
 * filename, folds it into a 32-bit hash with the polynomial {@code h = h*61 + ch - 32},
 * then scans the 10-byte-per-entry header table stored in a raw {@code byte[]}
 * to find a matching entry and returns the byte offset at which that file's
 * data begins.
 *
 * This is the RSC rev-235 equivalent of {@code Utility.getDataFileOffset} in the
 * rev-204 oracle (mudclient204/src/Utility.java lines 162-179).
 *
 * The two {@code String[][]} fields ({@link #entryNames} and {@link #uiStrings}) are
 * populated by the SocketFactory / initialiser (class {@code m}) at startup alongside
 * the main {@code int[][]} id table ({@link #idTable}). {@link #maxModelId} is a
 * rolling maximum maintained by the Scene renderer ({@code lb}).
 *
 * obf: oa
 */
public final class NameHash {

    // -------------------------------------------------------------------------
    // Static fields
    // -------------------------------------------------------------------------

    /**
     * Per-entity id table: {@code idTable[entityIndex][slot] = id}.
     * Allocated by the initialiser ({@code m}/SocketFactory): the outer array is
     * sized {@code [fa.b][]} (entity-count rows) and each row's length is read from
     * the cache buffer, then every slot is filled with a 2-byte value read
     * sequentially from the cache buffer via {@code t.a(65525)} (EntityDef helper,
     * {@code d.a(ka.b, ..., b.v)} — a {@code CacheFile} unsigned-short read that
     * advances the {@code ka.b} cursor).  Read back by {@code client} (Mudclient)
     * as a lookup from model-instance index to entity id.
     *
     * obf: oa.d
     */
    public static int[][] idTable;

    /**
     * Dead profiling counter incremented once per call to {@link #getFileOffset}.
     * Retained as a field stub for binary compatibility; never read by live code.
     *
     * obf: oa.e
     */
    public static int _profileCounter;

    /**
     * UI string table — index 0 holds the stake-dialog prompt:
     *   "Enter number of items to stake and press enter"
     * Decoded from the XOR string pool at class-load time.
     *
     * obf: oa.c
     */
    public static String[] uiStrings = new String[]{
        // z(z("R\x037^U7\x036VEr\x1fcTA7\x047^JdM7T\x07d\x19\"PB7\x0c-_\x07g\x1f&HT7\x08-OBe"))
        // decoded: "Enter number of items to stake and press enter"
        "Enter number of items to stake and press enter"
    };

    /**
     * Entity / model name table: parallel to {@link #idTable}, length equals
     * the entity-count upper bound ({@code fa.b}).  Populated by the initialiser
     * ({@code m}) from ISAAC cipher output ({@code o.a(byte)}).
     *
     * obf: oa.a
     */
    public static String[] entryNames;

    /**
     * Rolling maximum model-id seen so far; written and read by the Scene renderer
     * ({@code lb}) — and read by GameModel ({@code ca}) — during scene construction
     * to track the highest entity id in the current view.
     *
     * obf: oa.b
     */
    public static int maxModelId;

    // -------------------------------------------------------------------------
    // XOR string-pool constants (private, used only for ErrorHandler signatures)
    // -------------------------------------------------------------------------

    /**
     * Decoded error-signature fragments used by the {@code ErrorHandler} ({@code i})
     * catch block inside {@link #getFileOffset}:
     *   z[0] = "oa.A("   — method signature prefix
     *   z[1] = "null"    — null-argument token
     *   z[2] = "{...}"   — non-null-argument token
     *
     * obf: oa.z
     */
    private static final String[] ERROR_SIG_PARTS = new String[]{
        // z(z("x\fmz\x0f"))  -> "oa.A("
        "oa.A(",
        // z(z("y\x18/W"))    -> "null"
        "null",
        // z(z("lCm\x15Z"))   -> "{...}"
        "{...}"
    };

    // -------------------------------------------------------------------------
    // Public API
    // -------------------------------------------------------------------------

    /**
     * Locate a named file inside a JAG archive byte-array and return the byte
     * offset at which its data begins, or {@code 0} if no entry matches.
     *
     * The JAG archive header layout (little-endian multi-byte values stored
     * big-endian per Jagex convention) is:
     * <pre>
     *   bytes 0-1         : numEntries (big-endian unsigned short)
     *   bytes 2 .. 2+numEntries*10-1 : entry table (10 bytes each):
     *       +0 .. +1      : (unused / padding in this rev)
     *       +2 .. +5      : nameHash  (big-endian int, 61-polynomial)
     *       +6 .. +8      : (unused / compressedSize?)
     *       +9 .. +11     : fileSize  (big-endian 3-byte uint)
     *   bytes 2+numEntries*10 .. : concatenated file data
     * </pre>
     *
     * Corresponds to {@code Utility.getDataFileOffset(String, byte[])} in the
     * rev-204 oracle (mudclient204/src/Utility.java:162).
     *
     * Note: the {@code dummy} parameter is always passed as {@code (byte)68} ('D')
     * by every call-site; the guard {@code if (dummy != 68) return 71} is an
     * anti-tamper dead branch and has been removed.  The unused {@code dummy}
     * parameter is documented here for traceability but dropped from the clean body.
     *
     * obf: oa.a(String, byte, byte[])
     *
     * @param name    the entry filename to look up (case-insensitive)
     * @param dummy   always 68; anti-tamper guard (dropped)
     * @param archive raw bytes of the JAG archive (header + data)
     * @return byte offset of the file's data within {@code archive}, or 0 if not found
     */
    public static final int getFileOffset(String name, byte dummy, byte[] archive) {
        // Read the 2-byte big-endian entry count from offset 0.
        // CacheFile.readUnsignedShort(0, archive) = (archive[0]<<8) | archive[1]
        // obf: var15 = d.a(0, (byte)48, var2)
        int numEntries = CacheFile.readUnsignedShort(0, archive);

        // Compute the 32-bit name hash: h = h*61 + uppercaseChar - 32
        // This is identical to Utility.getDataFileOffset in the rev-204 oracle.
        // obf: var4
        int wantedHash = 0;
        name = name.toUpperCase();
        for (int i = 0; i < name.length(); i++) {
            // Polynomial: multiply by 61 then add (char - 32).  The '-32' maps
            // the printable ASCII range so that ' ' (0x20) contributes 0.
            wantedHash = wantedHash * 61 + name.charAt(i) - 32;
        }

        // Data begins immediately after the header table (2 bytes for count +
        // 10 bytes per entry).
        // obf: var5 (reused as offset after the hash loop)
        int offset = 2 + numEntries * 10;

        // Scan each 10-byte entry for a matching hash.
        // obf: var6
        for (int entry = 0; entry < numEntries; entry++) {
            int base = entry * 10;

            // Entry name-hash: bytes [base+2 .. base+5], big-endian 32-bit int.
            // obf: var7
            int entryHash = ((archive[base + 2] & 0xFF) << 24)
                          | ((archive[base + 3] & 0xFF) << 16)
                          | ((archive[base + 4] & 0xFF) <<  8)
                          |  (archive[base + 5] & 0xFF);

            // Entry file size: bytes [base+9 .. base+11], big-endian 24-bit uint.
            // Used to advance the data offset past entries that don't match.
            // obf: var8
            int fileSize = ((archive[base +  9] & 0xFF) << 16)
                         | ((archive[base + 10] & 0xFF) <<  8)
                         |  (archive[base + 11] & 0xFF);

            if (entryHash == wantedHash) {
                // Found: return the byte offset where this file's data starts.
                return offset;
            }

            // Not a match — advance past this entry's data block.
            offset += fileSize;
        }

        // Entry not found in archive.
        return 0;
    }

    // -------------------------------------------------------------------------
    // XOR string-pool helpers (private; identical boilerplate in every class)
    // -------------------------------------------------------------------------

    /**
     * First-pass XOR decode: converts a raw obfuscated string literal to a
     * char array.  For strings of length less than 2, XORs the single character
     * with {@code 0x27} (the key for the 1-char variant).  For longer strings
     * the array is returned as-is for the second pass.
     *
     * obf: oa.z(String)
     */
    private static char[] xorDecodePass1(String raw) {
        char[] arr = raw.toCharArray();
        if (arr.length < 2) {
            // Single-char pool entry: key is 0x27 ('\'' )
            arr[0] = (char)(arr[0] ^ 0x27);
        }
        return arr;
    }

    /**
     * Second-pass XOR decode: applies a 5-byte rotating key table
     * {@code [23, 109, 67, 59, 39]} to the char array produced by
     * {@link #xorDecodePass1}, then interns and returns the result.
     *
     * obf: oa.z(char[])
     */
    private static String xorDecodePass2(char[] arr) {
        // Key schedule: position mod 5 selects the XOR byte.
        // key = { 23=0x17, 109=0x6D, 67=0x43, 59=0x3B, 39=0x27 }
        for (int i = 0; i < arr.length; i++) {
            int key;
            switch (i % 5) {
                case 0:  key = 23;  break;
                case 1:  key = 109; break;
                case 2:  key = 67;  break;
                case 3:  key = 59;  break;
                default: key = 39;  break;
            }
            arr[i] = (char)(arr[i] ^ key);
        }
        return new String(arr).intern();
    }
}
