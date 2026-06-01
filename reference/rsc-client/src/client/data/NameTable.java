package client.data;

/**
 * NameTable (obf: ub) — global string-name registry for the RSC client (rev ~233-235).
 *
 * <p>Holds two flat String arrays that are populated during game data loading and world/scene
 * setup:
 * <ul>
 *   <li>{@link #modelNames} — up to 5 000 model-name strings, indexed by numeric model ID.
 *       Populated by {@code GameData.getModelIndex(String)} (oracle: {@code GameData.modelName}).
 *       Read by {@code World} (lb) and the 3-D scene ({@code Scene}/k) to look up model meshes.
 *   <li>{@link #textureNames} — dynamically-allocated texture/sub-type name array (oracle:
 *       {@code GameData.textureName}); read by packet decode to resolve texture IDs.
 * </ul>
 *
 * <p>Also carries two parallel-sort helpers used elsewhere in the engine when an {@code int[]}
 * key array must be sorted alongside an {@code Object[]} value array (quicksort + post-sort
 * counter), and a raw-integer packet read helper ({@link #readUnsignedInt}).
 *
 * <p>The class-level XOR string pool (field {@code SIG_STRINGS}) decodes to the method-signature
 * strings used in the per-method error-reporting catch blocks:
 * <pre>
 *   SIG_STRINGS[0] = "ub.B("    // sortWithKeys sig (lookup)
 *   SIG_STRINGS[1] = "ub.C("    // readUnsignedInt sig
 *   SIG_STRINGS[2] = "ub.A("    // sortWithKeys (partition) sig
 *   SIG_STRINGS[3] = "{...}"    // non-null array placeholder
 *   SIG_STRINGS[4] = "null"     // null array placeholder
 * </pre>
 *
 * Oracle correspondence: {@code GameData.modelName[]}, {@code GameData.textureName[]},
 * {@code GameData.getUnsignedInt()}.
 */
final class NameTable {

    // -------------------------------------------------------------------------
    // Public state — read and written by Mudclient (client) and World (lb)
    // -------------------------------------------------------------------------

    /**
     * Model-name registry: up to 5 000 interned model-name strings.
     * Index 0 = "na" (null/no-model sentinel).  Populated incrementally as
     * object/NPC definitions are loaded from the game-data archive.
     * obf: c
     */
    static String[] modelNames = new String[5000];

    /**
     * Dead profiling counter — incremented once per {@link #readUnsignedInt} call.
     * Not used by game logic.
     * obf: f
     */
    static int _profileReadInt; // obf: f (dead profiling counter)

    /**
     * Texture / subtype name array; allocated and filled during game-data load
     * (oracle: {@code GameData.textureName}).  Read by packet-decode routines
     * to map a texture index onto a name string.
     * obf: b
     */
    static String[] textureNames; // obf: b

    /**
     * Dead profiling counter — incremented once per {@link #sortWithKeys} call.
     * Not used by game logic.
     * obf: e
     */
    static int _profileSort; // obf: e (dead profiling counter)

    /**
     * Integer key array associated with {@link #textureNames}; used as the
     * sort-key buffer in {@link #sortWithKeys} calls (parallel sort alongside
     * {@link #textureNames}).
     * obf: g
     */
    static int[] sortKeys; // obf: g

    /**
     * Dead profiling counter — incremented once per {@link #findById} call.
     * Not used by game logic.
     * obf: d
     */
    static int _profileFind; // obf: d (dead profiling counter)

    /**
     * 100-element string scratch array; used by Mudclient (client) as a
     * per-frame sliding window of recent chat/player-name strings (oracle:
     * the 100-slot name ring used alongside {@code pa.g[]}, {@code k.G[]},
     * {@code ja.N[]}, {@code aa.k[]} in the client's chat-line renderer).
     * Element [0] is the most-recent entry; entries are shifted down on each
     * new addition; all slots reset to {@code null} on logout/clear.
     * obf: a
     */
    static String[] recentNames = new String[100]; // obf: a

    // -------------------------------------------------------------------------
    // Private XOR string pool — decoded at class-load time
    // -------------------------------------------------------------------------

    /**
     * Per-class XOR-encoded signature strings for error reporting.
     * Decoded by the two {@link #decodeKey} / {@link #decodePool} helpers.
     * Key table (mod-5): [98, 108, 35, 8, 120] ('b','l','#',0x08,'x').
     * Decoded values:
     *   [0] "ub.B("  [1] "ub.C("  [2] "ub.A("  [3] "{...}"  [4] "null"
     * obf: z
     */
    private static final String[] SIG_STRINGS = new String[]{
        decodePool(decodeKey("\rJP")),  // "ub.B("
        decodePool(decodeKey("\rKP")),  // "ub.C("
        decodePool(decodeKey("\rIP")),  // "ub.A("
        decodePool(decodeKey("B\r&")),  // "{...}"
        decodePool(decodeKey("\fOd"))         // "null"
    };

    // =========================================================================
    // Methods
    // =========================================================================

    /**
     * In-place quicksort of {@code keys[lo..hi]} with a parallel {@code values} array
     * that is reordered to match the sorted key order (obf: a(int[],byte,int,int,Object[])).
     *
     * <p>This is a 3-median-of-1 Lomuto-style partition:
     * <ol>
     *   <li>Swap the middle element with the high element to choose a pivot.
     *   <li>Walk {@code lo..hi-1}: any element strictly less than the pivot
     *       (with an obfuscated {@code (i&flag)+pivot} comparison — see below)
     *       is swapped into the left partition.
     *   <li>Place pivot between the two partitions.
     *   <li>Recurse on left and right sub-ranges.
     * </ol>
     *
     * <p><b>Obfuscation note on the comparison target:</b> the JARifier inserts
     * {@code flag = (pivot == Integer.MAX_VALUE) ? 0 : 1} and compares
     * {@code (i & flag) + pivot} rather than bare {@code pivot}.  When
     * {@code pivot != Integer.MAX_VALUE}, {@code flag=1} and the comparison
     * target alternates between {@code pivot} (even i) and {@code pivot+1}
     * (odd i), which is equivalent to "less than or equal to pivot" for odd
     * positions and "strictly less than pivot" for even positions.  This
     * asymmetry is an anti-analysis artefact; the partition remains correct
     * because quicksort only requires a consistent total order, and the
     * MAX_VALUE guard prevents signed overflow.  We preserve the semantics
     * faithfully.
     *
     * <p>The {@code byteGuard} parameter is an anti-tamper dummy: the method
     * body only executes when {@code byteGuard <= -82}; recursive calls always
     * pass {@code (byte)-124} and {@code (byte)-123}.  The guard is stripped.
     *
     * <p>After each return from a recursive call, the profiling counter
     * {@link #_profileSort} is incremented (dead).
     *
     * @param keys      integer key array to sort (sorted in-place)
     * @param byteGuard obfuscation guard; must be {@code <= -82} for execution
     *                  (dropped anti-tamper param — real callers use -124/-123)
     * @param lo        inclusive lower bound of sub-range
     * @param hi        inclusive upper bound of sub-range
     * @param values    object array sorted in parallel with {@code keys}
     *
     * obf: a(int[], byte, int, int, Object[])
     */
    static final void sortWithKeys(int[] keys, byte byteGuard, int lo, int hi, Object[] values) {
        // Strip: boolean bl = client.vh; — opaque predicate, always false
        // Strip: try/catch wrapping — unwrapped to bare body below

        // Anti-tamper guard: only execute when byteGuard <= -82.
        // Recursive calls pass (byte)-124 and (byte)-123, both satisfy this.
        if (byteGuard > -82) {
            return;
        }

        // Base case: sub-range has 0 or 1 elements — nothing to sort.
        // ~lo > ~hi  ⟺  lo < hi  (bitwise NOT reverses order)
        if (lo >= hi) {
            // Dead profiling: ++_profileSort happens on return path
            _profileSort++;
            return;
        }

        // --- Pivot selection: median of (middle, high) by swapping mid → hi ---
        int mid = (lo + hi) / 2;
        int leftPtr = lo;           // next slot for a "less-than-pivot" element

        int pivot = keys[mid];
        keys[mid] = keys[hi];
        keys[hi] = pivot;

        Object pivotValue = values[mid];
        values[mid] = values[hi];
        values[hi] = pivotValue;

        // Obfuscated comparison flag:
        //   flag = 0  when pivot == Integer.MAX_VALUE (to avoid pivot+1 overflow)
        //   flag = 1  otherwise (comparison alternates ±1 around pivot)
        // See method-level Javadoc for full analysis.
        int compFlag = (Integer.MIN_VALUE == ~pivot) ? 0 : 1; // ~pivot==MIN_VALUE ⟺ pivot==MAX_VALUE

        // --- Partition: sweep lo..hi-1, move elements < target to the left ---
        for (int idx = lo; idx < hi; idx++) {
            // Obfuscated comparison target: (idx & compFlag) + pivot
            //   compFlag=1: alternates between pivot (even idx) and pivot+1 (odd idx)
            //   compFlag=0: always pivot (when pivot==MAX_VALUE, prevents overflow)
            int compTarget = (idx & compFlag) + pivot;
            int elem = keys[idx];

            if (compTarget <= elem) {
                // element >= target: leave it in the right partition
                continue;
            }

            // element < target: move it into the left partition
            int tmp = keys[idx];
            keys[idx] = keys[leftPtr];
            keys[leftPtr] = tmp;

            Object tmpVal = values[idx];
            values[idx] = values[leftPtr];
            values[leftPtr++] = tmpVal;
        }

        // Place pivot at the boundary between the two partitions
        keys[hi] = keys[leftPtr];
        keys[leftPtr] = pivot;
        values[hi] = values[leftPtr];
        values[leftPtr] = pivotValue;

        // Recurse on left sub-range  [lo .. leftPtr-1]
        sortWithKeys(keys, (byte) -124, lo, leftPtr - 1, values);
        // Recurse on right sub-range [leftPtr+1 .. hi]
        sortWithKeys(keys, (byte) -123, 1 + leftPtr, hi, values);

        // Dead profiling counter increment (present on every return path).
        _profileSort++;
    }

    /**
     * Looks up a singleton {@code ChatCipher} (v) object by its numeric ID from the engine's
     * global cipher registry (obf: a(int, byte)).
     *
     * <p>The byte parameter {@code magicGuard} is an anti-tamper sentinel: only when
     * {@code magicGuard == 24} is {@link #modelNames} left intact; any other value
     * nullifies it (dead anti-tamper side-effect).  Callers always pass {@code (byte)24}.
     *
     * <p>Internally calls {@code ErrorHandler.lookupCiphers(magicGuard - 735)} which
     * decodes to {@code ErrorHandler.lookupCiphers(-711)} — a hardcoded registry key
     * returning the global array of ChatCipher singletons.  Iterates the array
     * linearly to find the one whose {@code id} field matches {@code cipherId}.
     *
     * @param cipherId  numeric cipher/entity ID to find (matched against {@code v.i})
     * @param magicGuard anti-tamper byte sentinel; callers always pass {@code (byte)24}
     * @return the matching {@code ChatCipher} (v) singleton, or {@code null} if not found
     *
     * obf: a(int, byte)
     */
    static final client.net.ChatCipher findById(int cipherId, byte magicGuard) {
        // Strip: boolean bl = client.vh; — opaque predicate
        // Strip: try/catch wrapper

        // Dead profiling counter (anti-tamper side effect on the counter field)
        // ++_profileFind; — dropped (profiling, dead)

        // Anti-tamper: nullify modelNames if called with wrong guard byte.
        // Real callers always pass (byte)24, so this branch is never taken.
        if (magicGuard != 24) {
            modelNames = null; // dead anti-tamper side-effect; never actually reached
        }

        // Registry lookup: ErrorHandler.lookupCiphers(magicGuard - 735)
        // With magicGuard==24: lookupCiphers(24-735) = lookupCiphers(-711)
        // i.a(-711) returns the global ChatCipher singleton array.
        // obf: v[] var2 = i.a(var1 + -735);
        client.net.ChatCipher[] ciphers = client.util.ErrorHandler.lookupCiphers(magicGuard - 735);

        for (int idx = 0; idx < ciphers.length; idx++) {
            client.net.ChatCipher cipher = ciphers[idx];
            if (cipherId == cipher.id) { // obf: cipher.i
                return cipher;
            }
        }

        return null;
    }

    /**
     * Reads one big-endian unsigned 32-bit integer from the current position in the
     * incoming packet buffer ({@code b.v} at offset {@code ka.b}), advancing the
     * offset by 4 bytes (obf: a(byte)).
     *
     * <p>If the raw value exceeds 99 999 999, it is treated as a negative-sentinel
     * and remapped: {@code result = -result + 99_999_999}.  This matches the oracle's
     * {@code GameData.getUnsignedInt()} clamping convention.
     *
     * <p>The byte parameter {@code magicGuard} is an anti-tamper sentinel: only when
     * {@code magicGuard == -105} is {@link #sortKeys} left intact.  Real callers
     * always pass the correct value.
     *
     * @param magicGuard anti-tamper byte sentinel (callers pass {@code (byte)-105})
     * @return the clamped unsigned 32-bit integer value from the packet buffer
     *
     * obf: a(byte)
     */
    static final int readUnsignedInt(byte magicGuard) {
        // Strip: try/catch wrapper

        // Dead profiling counter: ++_profileReadInt; — dropped

        // Read 4 bytes big-endian from packet buffer b.v at offset ka.b
        // m.a(true, ka.b, b.v) == SocketFactory.readBigEndianInt(true, offset, buf)
        // obf: int var1 = m.a(true, ka.b, b.v);
        int value = client.net.SocketFactory.readBigEndianInt(true, client.util.IntHolder.offset, client.net.Packet.buffer);

        // Clamp: if value > 99_999_999, treat as signed-negative sentinel.
        // ~value < -100_000_000  ⟺  value > 99_999_999
        // (bitwise-NOT then compare; equivalent to the oracle's > 99999999 check)
        if (~value < -100000000) {
            value = -value + 99999999;
        }

        // Anti-tamper: nullify sortKeys if called with wrong guard byte.
        // Dead side-effect — real callers always pass (byte)-105.
        if (magicGuard != -105) {
            sortKeys = null; // dead anti-tamper side-effect
        }

        // Advance packet-buffer cursor by 4 bytes
        client.util.IntHolder.offset += 4; // obf: ka.b += 4

        return value;
    }

    // =========================================================================
    // XOR string-pool helpers (private, obf: z / z)
    // =========================================================================

    /**
     * First-pass decoder: converts the raw compile-time string literal to a {@code char[]}
     * applying a single-byte XOR when the string is shorter than 2 characters.
     * For strings of length >= 2 the array is returned as-is (pass-through to
     * {@link #decodePool}).
     *
     * <p>Single-char case: {@code char ^ 0x78} ('x').
     *
     * obf: z(String)  → char[]
     */
    private static char[] decodeKey(String raw) {
        char[] chars = raw.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 'x'); // 0x78
        }
        return chars;
    }

    /**
     * Second-pass decoder: XOR-decodes a {@code char[]} using a 5-byte rotating key
     * table and interns the resulting string.
     *
     * <p>Key table (position mod 5):
     * <pre>
     *   0 → 98  ('b')
     *   1 → 108 ('l')
     *   2 → 35  ('#')
     *   3 → 8   (0x08)
     *   4 → 120 ('x')
     * </pre>
     *
     * obf: z(char[])  → String
     */
    private static String decodePool(char[] chars) {
        int len = chars.length;
        for (int i = 0; i < len; i++) {
            int key;
            switch (i % 5) {
                case 0:  key = 98;  break; // 'b'
                case 1:  key = 108; break; // 'l'
                case 2:  key = 35;  break; // '#'
                case 3:  key = 8;   break; // 0x08
                default: key = 120; break; // 'x'
            }
            chars[i] = (char) (chars[i] ^ key);
        }
        return new String(chars).intern();
    }
}
