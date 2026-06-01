package client.world;

import client.scene.GameModel;    // ca
import client.ui.MessageList;     // wb  — holds the shared CRC lookup table
import client.data.DataStore;     // nb  — unsigned-byte helper
import client.util.ErrorHandler;  // i   — per-method exception wrapper
import client.util.LinkedQueue;   // db  — char-category helper used by trimString
import client.data.RecordLoader;  // f   — char validity check (f.a; consults ga=CharTable internally)
import client.util.DecodeBuffer;  // ac  — accent/char-folding map used by trimString

/**
 * WorldEntity — a single renderable entry in the scene/wall-object sort list.
 *
 * The Scene renderer (lb / Scene, client.scene) maintains a flat array
 * {@code y[]} of these records; each one ties one face of a GameModel (ca) to
 * its axis-aligned bounding box and sort key so that the painter's-algorithm
 * sort in Scene can sequence overlapping wall segments and scene objects
 * correctly.
 *
 * (NOTE: per NAMING.md, lb = Scene and k = World — the painter-sort owner of
 * this w[] array is the Scene 3D renderer, not the World/terrain manager.)
 *
 * Fields fall into three groups:
 *   1. GameModel reference + face index (o / i)
 *   2. AABB in world/screen space for painter-sort overlap tests
 *      (e, m  = x-axis range;  j, h  = z-axis range;  q, u  = second-axis range)
 *   3. The rotated face normal dotted with the axes (l, r, k → scalar s)
 *      used to decide which direction the face is "facing" for sorting
 *
 * Additionally this class carries three static utility methods that are
 * package-shared but logically unrelated to world entities:
 *   • trimAndValidateString  — trims a CharSequence to printable chars
 *   • computeCrc32           — CRC-32 over a byte-range using MessageList's table
 *   • readSignedShort        — decode a big-endian signed 16-bit int from a byte[]
 *
 * Obfuscated class name: {@code w}
 * Package placement:     client.world  (its w[] array is owned by lb = Scene)
 */
public final class WorldEntity {

    // -----------------------------------------------------------------------
    // Instance fields — sorted by logical group
    // -----------------------------------------------------------------------

    /**
     * Y-component of the face normal in world-space rotation.
     * Stored so the sort helper can recompute the normal dot-product without
     * re-reading the model.
     * (Clean base: Scene assigns {@code entity.l = normalY} — obf field {@code l} is normalY,
     * NOT normalX; an earlier pass had this inverted. Cross-checked against Scene.java line 846.)
     * obf: l
     */
    public int normalY;    // obf: l

    /**
     * X-component of the face normal.
     * (Clean base: Scene assigns {@code entity.r = normalX} — obf field {@code r} is normalX.)
     * obf: r
     */
    public int normalX;    // obf: r

    /**
     * Z-component of the face normal.
     * obf: k
     */
    public int normalZ;    // obf: k

    /**
     * Pre-computed dot-product of the face normal with the camera/sort axis.
     * Computed as {@code normalY * bb[v0] + normalX * H[v0] + normalZ * cc[v0]}
     * where bb/H/cc are the model's per-vertex normal components.
     * Used as the primary ordering key during the painter-sort pass.
     * obf: s
     */
    public int normalDot;  // obf: s

    /**
     * Minimum X bound of the entity's axis-aligned bounding box.
     * Set from the minimum of the face's vertex X-coords after transform.
     * obf: e
     */
    public int minX;       // obf: e

    /**
     * Maximum X bound of the entity's axis-aligned bounding box.
     * obf: m
     */
    public int maxX;       // obf: m

    /**
     * Minimum Z bound of the entity's axis-aligned bounding box.
     * obf: j
     */
    public int minZ;       // obf: j

    /**
     * Maximum Z bound of the entity's axis-aligned bounding box.
     * obf: h
     */
    public int maxZ;       // obf: h

    /**
     * Minimum value on the secondary sort axis (typically screen-Y or depth).
     * obf: q
     */
    public int minDepth;   // obf: q

    /**
     * Maximum value on the secondary sort axis.
     * obf: u
     */
    public int maxDepth;   // obf: u

    /**
     * Reference to the GameModel (ca) that owns this face/poly.
     * Null if the entity slot is unoccupied.
     * obf: o
     */
    public GameModel model; // obf: o  (type: ca)

    /**
     * Face index within {@link #model} that this entity represents.
     * Indexes into model.o (face vertex-index arrays), model.lb (face vertex count),
     * model.M (face flags), etc.
     * obf: i
     */
    public int faceIndex;  // obf: i

    /**
     * Depth sort key: average Z (or centroid depth) of the face, used for
     * the quicksort pass in Scene (lb).  Set to (model centroid + offset) / face count.
     * obf: t
     */
    public int sortDepth;  // obf: t

    /**
     * Object/wall tag identifier: the game-object ID encoded in this face's
     * tag field (model.faceTag), or -1 if none.
     * Used by Scene (lb) to look up adjacency flags (Hb array) during rendering.
     * obf: b
     */
    public int objectId;   // obf: b

    /**
     * Whether this entity slot is currently occupied / active.
     * Set to {@code true} when a face is assigned and {@code false} when the
     * slot is reset between scene reloads.
     * obf: c
     */
    public boolean active = false; // obf: c

    /**
     * Index of this entity within the Scene's (lb) {@code y[]} array.
     * Written once on initialisation so each slot knows its own position,
     * allowing the linked-list painter-sort to swap entries by index.
     * obf: f
     */
    public int slotIndex = 0;      // obf: f

    /**
     * Index of the entity that precedes this one in the painter-sort chain,
     * or -1 if this entity is at the front of its chain.
     * Used by Scene's (lb) insertion-sort to build the ordered render list.
     * obf: p
     */
    public int prevSortIndex = -1; // obf: p

    // -----------------------------------------------------------------------
    // Static fields
    // -----------------------------------------------------------------------

    /**
     * Dead profiling counter — incremented once per call to computeCrc32.
     * Never read by game logic; artifact of the obfuscator's method-call
     * profiling instrumentation.
     * obf: n
     */
    public static int _profileCtr_computeCrc32; // obf: n  (dead profiling counter)

    /**
     * Dead profiling counter — incremented once per call to readSignedShort.
     * obf: d
     */
    public static int _profileCtr_readSignedShort; // obf: d  (dead profiling counter)

    /**
     * Dead profiling counter — incremented once per call to trimAndValidateString.
     * obf: a  (NOTE: not to be confused with the overloaded static method a())
     */
    public static int _profileCtr_trimAndValidateString; // obf: a  (dead profiling counter)

    /**
     * Shared pixel-offset / sprite-index lookup table, length = NameHash.entityCount.
     * Allocated in SocketFactory (m) alongside nb.d, n.m, db.l, and aa.c during
     * client bootstrap; entries are filled with values from ChatCipher.a(int).
     * Read by Mudclient (client) when blitting wall-object sprites from the
     * Surface pixel buffer, indexing into SurfaceSprite's pixel arrays.
     * obf: g
     */
    public static int[] spriteOffsets; // obf: g

    // -----------------------------------------------------------------------
    // Constructor
    // -----------------------------------------------------------------------

    /**
     * Constructs a new, empty WorldEntity slot.
     * {@link #prevSortIndex} is initialised to -1 (no predecessor in sort chain)
     * and {@link #slotIndex} to 0; the Scene renderer (lb) overwrites slotIndex
     * immediately after construction.
     */
    public WorldEntity() {
        // p = -1  and  f = 0  are the only non-default initialisations.
        // (active = false and all int fields = 0 are Java defaults.)
        this.prevSortIndex = -1;
        this.slotIndex = 0;
    }

    // -----------------------------------------------------------------------
    // Static utility methods (logically unrelated to WorldEntity data;
    //   placed here by the obfuscator as static members of class w)
    // -----------------------------------------------------------------------

    /**
     * Trims a CharSequence to its printable, valid game-character content.
     *
     * Strips leading and trailing whitespace (via db.a / LinkedQueue.isSpace),
     * then iterates the remaining window: each char is first validated by
     * f.a (RecordLoader.a, which consults ga = CharTable), then folded to
     * ASCII via the ac codec (ac.a = DecodeBuffer.a).  Characters whose folded
     * result is 0 (NUL) are silently dropped.  Returns null if the result is
     * empty or if the input is null or too short.
     *
     * Called by Mudclient when rendering player chat messages over characters
     * (e.g., ta.C player-name strings trimmed to at least 82 chars minimum).
     *
     * Decoded error-string prefix: "w.A("  (z[4] in obf string pool)
     *
     * @param text  the raw CharSequence to trim (may be null)
     * @param minLength  minimum acceptable length threshold (inputs with
     *                   length <= 47 return null unconditionally)
     * @return the trimmed, validated string, or null if empty / invalid
     * obf: static final String a(CharSequence, byte)
     */
    public static final String trimAndValidateString(CharSequence text, byte minLength) {
        // Dead profiling counter removed: ++_profileCtr_trimAndValidateString
        // Anti-tamper guard: minLength <= 47 → null (obfuscator dummy param check)
        if (minLength <= 47) {
            return null;
        }
        if (text == null) {
            return null;
        }

        // Trim leading whitespace characters.
        int start = 0;
        int end = text.length();
        // LinkedQueue.isSpace (db.a) returns true for space-like chars at position.
        while (start < end && db.a(32, text.charAt(start))) {
            start++;
        }
        // Trim trailing whitespace.
        while (end > start && db.a(32, text.charAt(end - 1))) {
            end--;
        }

        int trimmedLen = end - start;
        // Reject empty or implausibly long (> 12) windows (obf range check).
        if (trimmedLen < 1 || trimmedLen > 12) {
            return null;
        }

        StringBuilder sb = new StringBuilder(trimmedLen);
        for (int i = start; i < end; i++) {
            char ch = text.charAt(i);
            // f.a (RecordLoader.a) validates the char is an acceptable game char.
            if (f.a(ch, 0)) {
                // ac.a (DecodeBuffer.a) folds extended/accented chars to ASCII;
                // a mapped result of 0 (NUL) means drop the char.
                // obf: append iff (~mapped != -1), i.e. (mapped != 0).
                char mapped = ac.a(ch, -194);
                if (mapped != 0) {
                    sb.append(mapped);
                }
            }
        }
        return sb.length() == 0 ? null : sb.toString();
    }

    /**
     * Computes a CRC-32 checksum over a range of bytes using the precomputed
     * CRC table stored in {@link MessageList#crcTable} (wb.q[]).
     *
     * Algorithm: standard CRC-32 with initial value -1 and final inversion (~crc).
     *
     * Usage in the codebase: called by GameModel (ca) when verifying model data
     * integrity, and by Utility (mb) for client-error reporting.
     *
     * Decoded error-string prefix: "w.C("  (z[3] in obf string pool)
     *
     * @param length   number of bytes to checksum (exclusive end: processes
     *                 data[offset] .. data[offset + length - 1])
     * @param unused   anti-tamper dummy parameter — ignored; junk modulo
     *                 expression {@code 123 / ((unused - 23) / 63)} is dead
     * @param data     byte array to checksum
     * @param offset   starting index into {@code data}
     * @return         the CRC-32 value of the specified range
     * obf: static final int a(int, int, byte[], int)
     */
    public static final int computeCrc32(int length, int unused, byte[] data, int offset) {
        // Dead profiling counter removed: ++_profileCtr_computeCrc32
        // Junk anti-tamper: `123 / ((unused - 23) / 63)` is dead — removed.

        int crc = -1; // CRC-32 init value
        for (int i = offset; i < length; i++) {
            // Standard CRC-32 table lookup step.
            // wb.q = MessageList.crcTable[256]
            crc = MessageList.crcTable[(data[i] ^ crc) & 0xFF] ^ (crc >>> 8);
        }
        return ~crc; // Final XOR inversion
    }

    /**
     * Reads a big-endian signed 16-bit integer (signed short) from two
     * consecutive bytes in an array.
     *
     * If the decoded value is >= 32768 it is adjusted down by 65536 to produce
     * the correct negative two's-complement value — equivalent to:
     * {@code (short)((data[offset] & 0xFF) << 8 | (data[offset+1] & 0xFF))}.
     *
     * The {@code guardParam} argument is an anti-tamper dummy: if it is not -1
     * the method returns the junk value 71 immediately.  In all real call sites
     * it is passed as -1 (literal or via opaque constant).
     *
     * Decoded error-string prefix: "w.B("  (z[0] in obf string pool)
     *
     * @param data       byte array containing the encoded short
     * @param guardParam anti-tamper guard; must be -1 for real execution
     * @param offset     byte index of the high byte within {@code data}
     * @return the signed 16-bit value at {@code data[offset..offset+1]},
     *         or 71 if {@code guardParam != -1}
     * obf: static final int a(byte[], int, int)
     */
    public static final int readSignedShort(byte[] data, int guardParam, int offset) {
        // Dead profiling counter removed: ++_profileCtr_readSignedShort
        // Anti-tamper guard: only the path where guardParam == -1 is real.
        if (guardParam != -1) {
            return 71; // dead / unreachable in practice
        }
        // nb.a(255, byte) = byte & 0xFF  (unsigned byte read)
        int value = 256 * DataStore.unsignedByte(255, data[offset])
                  +       DataStore.unsignedByte(255, data[1 + offset]);
        // Two's-complement wrap for values in [32768..65535]
        if (value >= 32768) {
            value -= 65536;
        }
        return value;
    }

    // -----------------------------------------------------------------------
    // XOR string-pool helpers (obfuscator-generated; not game logic)
    // -----------------------------------------------------------------------

    /**
     * First-pass string deobfuscation: converts the obfuscated literal to a
     * char array.  For arrays shorter than 2 chars the single element is
     * XOR'd with 0x1B (27); for longer arrays the chars are left as-is for
     * the second pass.
     *
     * obf: private static char[] z(String)
     */
    private static char[] decodeStep1(String encoded) {
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 0x1B);
        }
        return chars;
    }

    /**
     * Second-pass string deobfuscation: XOR-decodes the char array using the
     * 5-byte rotating key {102, 125, 122, 106, 27} (indices mod 5), then
     * interns the result.
     *
     * Decoded strings (used as error-message method signatures):
     *   z[0] = "w.B("     (readSignedShort error prefix)
     *   z[1] = "{...}"    (non-null argument placeholder)
     *   z[2] = "null"     (null argument placeholder)
     *   z[3] = "w.C("     (computeCrc32 error prefix)
     *   z[4] = "w.A("     (trimAndValidateString error prefix)
     *
     * obf: private static String z(char[])
     */
    private static String decodeStep2(char[] chars) {
        // Rotating XOR key: positions 0..4 → 102, 125, 122, 106, 27
        int len = chars.length;
        for (int i = 0; i < len; i++) {
            int key;
            switch (i % 5) {
                case 0:  key = 102; break;
                case 1:  key = 125; break;
                case 2:  key = 122; break;
                case 3:  key = 106; break;
                default: key =  27; break;
            }
            chars[i] = (char) (chars[i] ^ key);
        }
        return new String(chars).intern();
    }

    // -----------------------------------------------------------------------
    // Decoded static string pool (error-message method signatures)
    // -----------------------------------------------------------------------

    /**
     * Error-message method-signature strings, decoded from the obfuscated
     * XOR pool at class load time:
     *   z[0] = "w.B("   — readSignedShort
     *   z[1] = "{...}"  — non-null arg placeholder
     *   z[2] = "null"   — null arg placeholder
     *   z[3] = "w.C("   — computeCrc32
     *   z[4] = "w.A("   — trimAndValidateString
     * obf: private static final String[] z
     */
    private static final String[] ERROR_SIGS = new String[]{
        // z(z("S8B"))  → "w.B("
        decodeStep2(decodeStep1("S8B")),
        // z(z("STDf")) → "{...}"
        decodeStep2(decodeStep1("STDf")),
        // z(z("\b\b")) → "null"
        decodeStep2(decodeStep1("\b\b")),
        // z(z("S9B"))  → "w.C("
        decodeStep2(decodeStep1("S9B")),
        // z(z("S;B"))  → "w.A("
        decodeStep2(decodeStep1("S;B")),
    };
}
