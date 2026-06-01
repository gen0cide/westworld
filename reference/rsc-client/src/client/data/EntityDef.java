package client.data;
import client.util.IntHolder;
import client.net.Packet;
import client.scene.SpriteDecoder;

/**
 * EntityDef — a generic definition record used by several engine subsystems.
 *
 * <p>Instances are held in arrays by {@code MessageList} (wb) to back UI list
 * entries (chat/social/quest panels). The 12-field {@link #setFields} method
 * populates each record from caller-supplied name, examine, and integer data.
 *
 * <p>The class also owns three engine-wide static resources shared by other
 * classes:
 * <ul>
 *   <li>{@link #modelCount} / {@link #modelNames} — model-name registry,
 *       initialised from the data stream in {@code SocketFactory} (m).</li>
 *   <li>{@link #bytePool} — a tiered pool of recycled byte arrays used by
 *       {@code Utility} (mb) as a lightweight allocator.</li>
 *   <li>The JAG-archive extractor {@link #extractArchiveEntry} (equiv.
 *       to {@code Utility.unpackData()} in the 204 oracle).</li>
 * </ul>
 *
 * <p>Two static pixel-blitter helpers ({@link #fillPixelColumns16},
 * {@link #readUnsignedShortFromStream}) are co-located here by the compiler;
 * both serve the scene/renderer path.
 *
 * <p>Obfuscated class name: {@code t}.
 */
public final class EntityDef {

    // -------------------------------------------------------------------------
    // Instance fields — one record's worth of definition data
    // -------------------------------------------------------------------------

    /** Generic integer field A; role depends on context (e.g. sprite/model idx). // obf: e */
    public int fieldIntA;

    /** Generic integer field B; role depends on context. // obf: i */
    public int fieldIntB;

    /** Generic integer field C; role depends on context. // obf: j */
    public int fieldIntC;

    /** Generic integer field D (width / size). // obf: l */
    public int fieldIntD;

    /** Generic integer field E (height / size or type). // obf: m */
    public int fieldIntE;

    /** Generic integer field F; role depends on context. // obf: d */
    public int fieldIntF;

    /** Primary display name string (e.g. entity or list-entry label). // obf: p */
    public String name;

    /** Secondary text string (e.g. examine text or sub-label). // obf: b */
    public String examineText;

    /** Optional tertiary text string; null by default. // obf: o */
    public String extraText = null;

    // -------------------------------------------------------------------------
    // Static engine-wide fields
    // -------------------------------------------------------------------------

    /**
     * Dead profiling counter — incremented once on entry to {@link #setFields}.
     * Always written, never read for logic. // obf: f
     */
    public static int _profilerSetFields;

    /**
     * Dead profiling counter — incremented once on entry to
     * {@link #readUnsignedShortFromStream}. // obf: c
     */
    public static int _profilerReadShort;

    /**
     * Dead profiling counter — incremented once on entry to
     * {@link #fillPixelColumns16}. // obf: a
     */
    public static int _profilerFillPixels;

    /**
     * Dead profiling counter — incremented once on entry to
     * {@link #extractArchiveEntry}. // obf: k
     */
    public static int _profilerExtract;

    /**
     * Total number of loaded models; used as the upper bound for
     * {@link #modelNames}. Set in {@code SocketFactory} (m) via
     * {@link #readUnsignedShortFromStream}. // obf: g
     */
    public static int modelCount;

    /**
     * Model-name registry; {@code modelNames[i]} is the name of model {@code i}.
     * Allocated in {@code SocketFactory} (m) as {@code new String[modelCount]}.
     * Temporarily nulled by the anti-tamper guard inside
     * {@link #readUnsignedShortFromStream} when the magic token is wrong. // obf: h
     */
    public static String[] modelNames;

    /**
     * Tiered byte-array pool used by {@code Utility} (mb) as a recycling
     * allocator. {@code bytePool[sizeClass][slot]} holds a previously freed
     * byte[] that can be reused instead of {@code new byte[n]}. // obf: n
     */
    public static byte[][][] bytePool;

    // -------------------------------------------------------------------------
    // XOR-decoded string pool (error-handler method signatures)
    // -------------------------------------------------------------------------

    /**
     * XOR-encoded string literals decoded at class load time.
     * Indices:
     *   0 = "{...}"  (non-null placeholder in error messages)
     *   1 = "null"   (null placeholder in error messages)
     *   2 = "t.B("   (setFields signature prefix)
     *   3 = "t.C("   (fillPixelColumns16 signature prefix)
     *   4 = "t.D("   (extractArchiveEntry signature prefix)
     *   5 = "t.A("   (readUnsignedShortFromStream signature prefix)
     *
     * Decode: z(String) XORs char[0] with 't' if len<2 (single-char literals),
     * then z(char[]) XORs each char with key[i%5] = {66,83,113,59,116}.
     */
    private static final String[] z = new String[]{
        z(z("9}_\t")),   // z[0] = "{...}"
        z(z(",&W")),      // z[1] = "null"
        z(z("6}3")),      // z[2] = "t.B("
        z(z("6}2")),      // z[3] = "t.C("
        z(z("6}5")),      // z[4] = "t.D("
        z(z("6}0")),      // z[5] = "t.A("
    };

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /**
     * Constructs an empty EntityDef. {@code name} starts null and must be
     * populated via {@link #setFields} before the record is used. // obf: t()
     */
    public EntityDef() {
        this.name = null;
    }

    // -------------------------------------------------------------------------
    // Instance methods
    // -------------------------------------------------------------------------

    /**
     * Populates this record's fields from the caller-supplied arguments.
     *
     * <p>If {@code enableThreshold} > 69 (always true from the {@code MessageList}
     * caller which passes 100), all fields are written. Otherwise only
     * {@link #name}, {@link #fieldIntD}, and {@link #fieldIntE} are set and the
     * method returns early — this appears to be an anti-tamper stub guard.
     *
     * <p>Params {@code unusedStr2} and {@code unusedStr3} are accepted by the
     * signature but never stored; they exist to match the caller's 12-arg
     * convention (possibly retained from a broader GameData load path).
     *
     * <p>Equivalent to a GameData record setter in the 204 oracle. // obf: a(String,int,int,int,int,String,int,int,String,String,int,String)
     *
     * @param name            primary display name (stored in {@link #name})
     * @param intD            integer field D — width/size (stored in {@link #fieldIntD})
     * @param intA            integer field A (stored in {@link #fieldIntA})
     * @param intE            integer field E — height/type (stored in {@link #fieldIntE})
     * @param intB            integer field B (stored in {@link #fieldIntB})
     * @param unusedStr2      unused string argument (not stored; padding param)
     * @param enableThreshold guard value; fields written only when {@code > 69}
     * @param intF            integer field F (stored in {@link #fieldIntF})
     * @param extraText       optional tertiary text (stored in {@link #extraText})
     * @param unusedStr3      unused string argument (not stored; padding param)
     * @param intC            integer field C (stored in {@link #fieldIntC})
     * @param examineText     examine/description text (stored in {@link #examineText})
     */
    public final void setFields(
            String name,
            int    intD,
            int    intA,
            int    intE,
            int    intB,
            String unusedStr2,
            int    enableThreshold,
            int    intF,
            String extraText,
            String unusedStr3,
            int    intC,
            String examineText) {

        // obf: profiling counter ++f; — dead, removed
        this.name     = name;
        this.fieldIntD = intD;
        this.fieldIntE = intE;

        // Anti-tamper stub: caller always passes enableThreshold=100, so this
        // branch is always taken in practice. A forged/replayed call with a
        // value ≤ 69 would silently skip filling the remaining fields.
        if (enableThreshold <= 69) {
            return;
        }

        this.fieldIntF  = intF;
        this.fieldIntA  = intA;
        this.extraText  = extraText;
        this.fieldIntB  = intB;
        this.fieldIntC  = intC;
        this.examineText = examineText;
    }

    // -------------------------------------------------------------------------
    // Static methods
    // -------------------------------------------------------------------------

    /**
     * Reads one unsigned 16-bit value from the shared packet stream
     * ({@code Packet.rawData} / {@code b.v}) at offset {@code IntHolder.offset}
     * ({@code ka.b}) and advances the cursor by 2.
     *
     * <p>Used by {@code SocketFactory} (m) to read count fields during data
     * loading, e.g. {@code t.g = t.a(65525)} to read {@code modelCount}.
     *
     * <p>Anti-tamper: if {@code magicToken != 65525} the engine nulls
     * {@link #modelNames} before proceeding (dead-data poisoning).
     *
     * <p>Equivalent to {@code Utility.getUnsignedShort()} in the 204 oracle,
     * but reading directly from the static stream cursor. // obf: a(int)
     *
     * @param magicToken anti-tamper constant; callers always pass {@code 65525}
     * @return unsigned 16-bit value read from {@code Packet.rawData[IntHolder.offset..+1]}
     */
    public static final int readUnsignedShortFromStream(int magicToken) {
        // Anti-tamper guard: poison modelNames if the token doesn't match.
        // In real execution magicToken is always 65525, so this never fires.
        if (magicToken != 65525) {
            modelNames = null;
        }

        // obf: profiling counter ++c; — dead, removed
        // CacheFile.a(offset, (byte)100, data): reads big-endian unsigned short
        // from Packet.rawData[IntHolder.offset .. IntHolder.offset+1]
        int value = CacheFile.readUnsignedShort(IntHolder.offset, (byte) 100, Packet.rawData);
        IntHolder.offset += 2;
        return value;
    }

    /**
     * Extracts a named entry from a JAG archive byte array.
     *
     * <p>JAG archive format (from the 204 oracle {@code Utility.unpackData}):
     * <pre>
     *   bytes 0..1:  numEntries (big-endian unsigned short)
     *   bytes 2..N:  directory — numEntries × 10-byte records; per-record
     *                offsets (relative to the start of each 10-byte record):
     *     [0..3]  fileHash:          big-endian int32
     *     [4..6]  uncompressed size: 3-byte big-endian int
     *     [7..9]  compressed size:   3-byte big-endian int
     *   bytes N+1..: concatenated (optionally BZip2-compressed) file data
     * </pre>
     *
     * <p>The file is located by hashing {@code filename.toUpperCase()} with the
     * same polynomial used for model names:
     * {@code hash = hash * 61 + (ch - ' ')}.
     *
     * <p>If {@code fileSize != compressedSize}, {@code SpriteDecoder.decompress}
     * ({@code ea.a}) is called to BZip2-decompress the data into {@code dest};
     * otherwise the raw bytes are copied directly.
     *
     * <p>Equivalent to {@code Utility.unpackData()} in the 204 oracle. // obf: a(byte[],int,int,String,byte[])
     *
     * @param archiveData  raw JAG archive bytes (header + directory + file data)
     * @param destOffset   offset within {@code dest} at which to write the file
     *                     (passed as the {@code int i} extra-allocation hint)
     * @param allocHint    extra bytes to pre-allocate in {@code dest} when
     *                     creating a new buffer (dest = new byte[fileSize + allocHint])
     * @param filename     name of the file to find (matched case-insensitively)
     * @param dest         optional pre-allocated output buffer; allocated here if null
     * @return populated {@code dest} buffer, or {@code null} if not found
     */
    public static final byte[] extractArchiveEntry(
            byte[] archiveData,
            int    destOffset,
            int    allocHint,
            String filename,
            byte[] dest) {

        // obf: profiling counter ++k; — dead, removed

        // Read the number of directory entries from bytes 0..1
        int numEntries = (archiveData[0] & 0xFF) * 256
                       + (archiveData[1] & 0xFF);

        // Compute the 32-bit name hash: hash = hash * 61 + (ch - ' ')
        filename = filename.toUpperCase();
        int wantedHash = 0;
        for (int i = 0; i < filename.length(); i++) {
            wantedHash = 61 * wantedHash + (filename.charAt(i) - ' ');
        }

        // Scan the directory. File data begins at byte 2 + numEntries*10.
        int dataOffset = 2 + 10 * numEntries;

        for (int entry = 0; entry < numEntries; entry++) {
            // Directory entry layout (10 bytes, 0-indexed within this entry):
            //   [0..3]  fileHash:         bytes 2..5 of archive  (big-endian int32)
            //   [4..6]  uncompressed size: bytes 6..8 of archive  (3-byte big-endian)
            //   [7..9]  compressed size:   bytes 9..11 of archive (3-byte big-endian)
            int fileHash = (archiveData[2  + entry * 10] & 0xFF) * 0x1000000
                         + (archiveData[3  + entry * 10] & 0xFF) * 0x10000
                         + (archiveData[4  + entry * 10] & 0xFF) * 256
                         + (archiveData[5  + entry * 10] & 0xFF);

            int fileSize       = (archiveData[6  + entry * 10] & 0xFF) * 0x10000
                               + (archiveData[7  + entry * 10] & 0xFF) * 256
                               + (archiveData[8  + entry * 10] & 0xFF);

            int compressedSize = (archiveData[9  + entry * 10] & 0xFF) * 0x10000
                               + (archiveData[10 + entry * 10] & 0xFF) * 256
                               + (archiveData[11 + entry * 10] & 0xFF);

            if (fileHash == wantedHash) {
                // Found the entry — allocate dest if not provided
                if (dest == null) {
                    dest = new byte[fileSize + allocHint];
                }

                if (fileSize != compressedSize) {
                    // BZip2-compressed: decompress into dest
                    // ea = SpriteDecoder, which hosts the BZip2 decompressor
                    SpriteDecoder.decompress(dest, fileSize, archiveData, compressedSize, dataOffset);
                } else {
                    // Uncompressed: raw copy
                    for (int j = 0; j < fileSize; j++) {
                        dest[j] = archiveData[dataOffset + j];
                    }
                }
                return dest;
            }

            // Advance past this entry's compressed data in the data region
            dataOffset += compressedSize;
        }

        return null; // entry not found
    }

    /**
     * Sprite column blitter — fills a run of pixels in {@code destPixels} from
     * a palette lookup table {@code palette}, reading texture coordinates at a
     * fixed stride.
     *
     * <p>The inner loop is 16-wide unrolled (4 groups of 4 identical pixels),
     * which is twice the unroll of the 8-wide version in {@code SpriteScaler}
     * (ia). Each palette lookup reads {@code palette[texCoord >> 8 & 0xFF]},
     * i.e. the high byte of the 16-bit fixed-point texture coordinate selects
     * the palette entry; the low byte is the sub-texel fraction.
     *
     * <p>Entry condition: {@code height} must be negative (encodes the pixel
     * count as {@code -height}; the method returns immediately if
     * {@code height >= 0}).
     *
     * <p>Anti-tamper: if {@code magicToken != 418609192} the method tail-calls
     * itself with dummy arguments (recursive dead branch, unreachable in normal
     * execution). // obf: a(int,int,int,int[],int[],int,int,int)
     *
     * @param color      initial palette color index (overwritten from palette on first use)
     * @param stride     texture coordinate advance per pixel (fixed-point, << 2 applied internally)
     * @param height     negative pixel count; returns immediately if >= 0
     * @param destPixels destination pixel array (written in groups of 4)
     * @param palette    color palette lookup table (indexed by texCoord >> 8 & 0xFF)
     * @param texCoord   starting texture coordinate (16-bit fixed-point)
     * @param destIdx    starting write index into {@code destPixels}
     * @param magicToken anti-tamper token; callers always pass {@code 418609192}
     */
    public static final void fillPixelColumns16(
            int   color,
            int   stride,
            int   height,
            int[] destPixels,
            int[] palette,
            int   texCoord,
            int   destIdx,
            int   magicToken) {

        // obf: profiling counter ++a; — dead, removed

        // Guard: height must be negative to encode a pixel count
        if (height >= 0) {
            return;
        }

        // Load first palette entry; multiply stride by 4 (unroll factor)
        // palette index = texCoord >> 8 & 0xFF  (shift -1455207192 & 31 = 8)
        color  = palette[0xFF & (texCoord >> 8)];
        stride <<= 2; // stride *= 4: each loop iteration advances 4 sub-pixels
        texCoord += stride;

        // Main loop: process 16 pixels per iteration. The compiler interleaves
        // palette reloads / texCoord advances at irregular boundaries; the exact
        // ordering below is reproduced verbatim from the clean base (the partial
        // groups of 1 and 3 pixels are deliberate).
        int remaining = height / 16;  // negative quotient; loop while remaining < 0

        for (; remaining < 0; remaining++) {
            // Group 1: 4 identical pixels from current palette entry
            destPixels[destIdx++] = color;
            destPixels[destIdx++] = color;
            destPixels[destIdx++] = color;
            destPixels[destIdx++] = color;

            // Reload then advance; (texCoord & 0xFFF5) >> 8
            color = palette[(texCoord & 0xFFF5) >> 8];
            texCoord += stride;

            // Group 2: 4 identical pixels
            destPixels[destIdx++] = color;
            destPixels[destIdx++] = color;
            destPixels[destIdx++] = color;
            destPixels[destIdx++] = color;

            // Reload BEFORE advancing; (texCoord & 0xFFBD) >> 8
            color = palette[(texCoord & 0xFFBD) >> 8];

            // Group 3a: 1 pixel written with the just-reloaded color, then advance
            destPixels[destIdx++] = color;
            texCoord += stride;

            // Group 3b: 3 more identical pixels (same color)
            destPixels[destIdx++] = color;
            destPixels[destIdx++] = color;
            destPixels[destIdx++] = color;

            // Reload then advance; 0xFF & (texCoord >> 8)
            color = palette[0xFF & (texCoord >> 8)];
            texCoord += stride;

            // Group 4: 4 identical pixels
            destPixels[destIdx++] = color;
            destPixels[destIdx++] = color;
            destPixels[destIdx++] = color;
            destPixels[destIdx++] = color;

            // Reload then advance for next outer iteration; (texCoord & 0xFF32) >> 8
            color = palette[(texCoord & 0xFF32) >> 8];
            texCoord += stride;
        }

        // Anti-tamper recursive self-call guard — never fires in real execution
        if (magicToken != 418609192) {
            // magicToken mismatch: call self with junk args (dead branch)
            EntityDef.fillPixelColumns16(76, -63, -59, null, null, 124, -66, -99);
        }

        // Tail: fill remainder pixels (-height % 16, up to 15 pixels)
        int tailCount = -(height % 16);
        for (int i = 0; i < tailCount; i++) {
            destPixels[destIdx++] = color;
            // Reload palette every 4th pixel (when low 2 bits of i == 3)
            if ((i & 3) == 3) {
                // (texCoord >> 8 & 0xFF): standard sub-texel read
                color = palette[0xFF & (texCoord >> 8)];
                texCoord += stride;
            }
        }
    }

    // -------------------------------------------------------------------------
    // XOR string-pool decoders (private, class-init only)
    // -------------------------------------------------------------------------

    /**
     * First-pass decoder: if the char array has fewer than 2 elements, XOR
     * element 0 with {@code 't'} (0x74); otherwise pass through unchanged.
     * Used to handle single-character obfuscated literals. // obf: z(String)
     */
    private static char[] z(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 't');
        }
        return chars;
    }

    /**
     * Second-pass decoder: XOR each character with a rotating 5-byte key
     * {@code {66, 83, 113, 59, 116}} and intern the resulting string.
     * Paired with {@link #z(String)} to double-XOR the raw literals in
     * {@link #z}. // obf: z(char[])
     */
    private static String z(char[] chars) {
        // Key table: key[i % 5] = {66='B', 83='S', 113='q', 59=';', 116='t'}
        for (int i = 0; i < chars.length; i++) {
            int keyByte;
            switch (i % 5) {
                case 0:  keyByte =  66; break; // 'B'
                case 1:  keyByte =  83; break; // 'S'
                case 2:  keyByte = 113; break; // 'q'
                case 3:  keyByte =  59; break; // ';'
                default: keyByte = 116; break; // 't'
            }
            chars[i] = (char) (chars[i] ^ keyByte);
        }
        return new String(chars).intern();
    }
}
