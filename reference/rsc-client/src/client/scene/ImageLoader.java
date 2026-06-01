package client.scene;

import java.awt.Component;
import java.awt.Image;
import java.awt.image.IndexColorModel;

// Imports that won't resolve in a non-MS-JVM build — retained for fidelity:
// import client.net.ClientStream;   // da  — holds imageWidth, imageHeight, imageProducer
// import client.net.SocketFactory;  // m   — holds the shared ColorModel
// import client.scene.SurfaceImageProducer; // fb — ImageProducer implementation
// import client.world.World;         // lb  — notifyImageConsumers
// import client.net.StreamBase;      // ib  — maskByte helper
// import client.util.ErrorHandler;   // i   — rethrows with method signature string
// import client.util.Globals;        // l   — global applet/params holder
// import client.shell.LoaderThread;  // c   — carries imageWidth (field o)

/**
 * ImageLoader  (obf: pa)
 *
 * Utility class for the RSC rev-233-235 Microsoft J++ client.  Its two main
 * responsibilities are:
 *
 *  1. {@link #loadBmpImage}  — parse a raw Windows BMP byte array (indexed,
 *     8-bpp, 256-colour palette at byte offset 20) into an AWT {@link Image}
 *     backed by an {@link IndexColorModel}, then push the pixel rows (bottom-
 *     up BMP order → top-down AWT order) to the engine's
 *     {@link SurfaceImageProducer} ({@code da.db}) via three
 *     {@link World#notifyImageConsumers} calls.
 *
 *  2. {@link #buildMuLawTable} — build a 256-entry μ-law (or similar
 *     log-law) expansion table that maps signed byte indices to 8-bit
 *     amplitude values, used by the audio subsystem.
 *
 * The static initialiser also fills four shared lookup tables that are used
 * across the engine:
 *   - {@code SIN_512}  / cos side — 512-entry fixed-point trig (Q15, angle
 *     unit = 256/turn), used by the 3-D scene renderer.
 *   - {@code SIN_2048} / cos side — 2048-entry fixed-point trig (Q15, angle
 *     unit = 1024/turn), used for terrain and model transforms.
 *   - {@code BASE64_ENCODE} / {@code BASE64_DECODE} — a custom 64-symbol
 *     alphabet used for packet and cache encoding.
 *
 * Fits into the engine as a peer of {@link SpriteDecoder} (ea) and
 * {@link SpriteScaler} (ia) in the {@code client.scene} package.
 */
final class ImageLoader {

    // -----------------------------------------------------------------------
    // Static cross-class state (written by pa, read by lb/ua/fb/m)
    // -----------------------------------------------------------------------

    /**
     * Second static reference to a LoaderThread instance — unused in the
     * visible code of this class; likely set externally by the engine init
     * path.  Distinct from {@link #imageWidthCarrier} (obf: k) which is the
     * live reference used during image loading.
     * obf: b  (type: c = LoaderThread)
     */
    static client.shell.LoaderThread loaderThread;  // obf: b

    /**
     * Scratch int[100] array — purpose unclear from this class alone; likely
     * a shared work buffer for the renderer.
     * obf: g
     */
    static int[] scratchBuf = new int[100];  // obf: g

    /**
     * Shared int[] work array — external reference (set/read by other classes).
     * obf: f
     */
    static int[] sharedIntArray;  // obf: f

    /**
     * Base-64 encoding alphabet, 64 entries.
     * Layout: '0'-'9' at [0..9], 'A'-'Z' at [10..35], 'a'-'z' at [36..61],
     * 0xA3 ('£') at [62], '$' at [63].
     * obf: e  (private)
     */
    private static byte[] BASE64_ENCODE = new byte[64];  // obf: e

    /**
     * 512-entry fixed-point sine/cosine table, Q15 (scale 32768), angle unit
     * = 256 steps per full turn (i.e. each step = 360/256 ≈ 1.40625°).
     *
     * Layout:
     *   SIN_512[i]       = (int)(sin(i * 2π/256) * 32768)   for i = 0..255
     *   SIN_512[i + 256] = (int)(cos(i * 2π/256) * 32768)   for i = 0..255
     *
     * Matches oracle {@code Scene.sin512Cache[]}.
     * obf: a
     */
    static int[] SIN_512 = new int[512];  // obf: a

    /**
     * Loader-thread instance whose {@code imageWidth} field is written during
     * BMP header parse and then read as the row stride.
     * obf: k  (type: c = LoaderThread)
     *
     * Note: the decompiler types this as {@code c} (LoaderThread) because the
     * JVM field descriptor is {@code Lc;}, but at the usage sites it is
     * treated as a carrier of an {@code int} width value via field {@code o}.
     * The field {@code c.o} in the actual bytecode is an {@code int} field
     * that Vineflower mis-types as {@code g} (ListNode) due to obfuscation.
     */
    static client.shell.LoaderThread imageWidthCarrier;  // obf: k

    /**
     * Profiling counter — incremented on entry to {@link #loadBmpImage}.
     * Dead instrumentation artefact; kept to preserve field layout.
     * obf: i  (static int)
     *
     * CAUTION: do NOT confuse with the class reference {@code i} = ErrorHandler.
     */
    static int loadBmpImageCallCount;  // obf: i  (profiling counter — dead)

    /**
     * Profiling counter — incremented on entry to {@link #buildMuLawTable}.
     * Dead instrumentation artefact; kept to preserve field layout.
     * obf: c  (static int)
     */
    static int buildMuLawTableCallCount;  // obf: c  (profiling counter — dead)

    /**
     * Timing / tick counter — purpose unclear; possibly a frame or load
     * timestamp. Read externally.
     * obf: h
     */
    static long tickCounter;  // obf: h

    /**
     * 2048-entry fixed-point sine/cosine table, Q15 (scale 32768), angle unit
     * = 1024 steps per full turn (each step = 360/1024 ≈ 0.352°).
     *
     * Layout:
     *   SIN_2048[i]        = (int)(sin(i * 2π/1024) * 32768)   for i = 0..1023
     *   SIN_2048[i + 1024] = (int)(cos(i * 2π/1024) * 32768)   for i = 0..1023
     *
     * Matches oracle {@code Scene.sin2048Cache[]}.
     * obf: j
     */
    static int[] SIN_2048 = new int[2048];  // obf: j

    /**
     * Base-64 decoding table, 256 entries.  Inverse of {@link #BASE64_ENCODE}:
     * maps ASCII character code → 6-bit index (0..63), or 0 for unmapped chars.
     * obf: d
     */
    static int[] BASE64_DECODE = new int[256];  // obf: d

    /**
     * XOR-encoded string pool — decoded at class init time.
     * Entries (decoded):
     *   z[0] = "pa.B("    — method-signature prefix for loadBmpImage error context
     *   z[1] = "null"      — null-argument label in error context
     *   z[2] = "{...}"     — non-null argument placeholder in error context
     *   z[3] = "pa.A("    — method-signature prefix for buildMuLawTable error context
     *
     * obf: z  (private static final String[])
     */
    private static final String[] OBFSTR = new String[]{
        // Decoded at runtime via the two z() helpers below.
        // z[0]: "pa.B("   — error context prefix for loadBmpImage
        decodeXorStr(decodeXorCharArr("J_\f\t%")),
        // z[1]: "null"    — null argument marker
        decodeXorStr(decodeXorCharArr("TKN'")),
        // z[2]: "{...}"   — non-null argument marker
        decodeXorStr(decodeXorCharArr("A\fep")),
        // z[3]: "pa.A("   — error context prefix for buildMuLawTable
        decodeXorStr(decodeXorCharArr("J_\f\n%"))
    };  // obf: z

    // -----------------------------------------------------------------------
    // Static initialiser — fills trig and base-64 tables
    // -----------------------------------------------------------------------

    static {
        // --- 512-entry sine/cosine table (256 steps/turn, Q15) ---
        // Angle unit: 1 step = 2π/256 rad ≈ 1.406°
        // Constant 0.02454369 = 2π/256 ≈ 6.2832/256
        // obf: static block, first while loop
        for (int i = 0; i < 256; i++) {
            SIN_512[i]       = (int)(32768.0 * Math.sin(0.02454369  * (double)i));
            SIN_512[256 + i] = (int)(32768.0 * Math.cos(0.02454369  * (double)i));
        }

        // --- 2048-entry sine/cosine table (1024 steps/turn, Q15) ---
        // Angle unit: 1 step = 2π/1024 rad ≈ 0.352°
        // Constant 0.00613592315 ≈ 2π/1024
        // obf: static block, second while loop  (-1025 < ~var0  ↔  var0 < 1024)
        for (int i = 0; i < 1024; i++) {
            SIN_2048[i]        = (int)(Math.sin((double)i * 0.00613592315) * 32768.0);
            SIN_2048[i + 1024] = (int)(Math.cos((double)i * 0.00613592315) * 32768.0);
        }

        // --- Base-64 encoding alphabet ---
        // Digits '0'-'9' → indices 0-9
        // obf: static block, third loop  (-11 < ~var0  ↔  var0 < 10)
        for (int i = 0; i < 10; i++) {
            BASE64_ENCODE[i] = (byte)('0' + i);  // 48..57
        }

        // Uppercase 'A'-'Z' → indices 10-35
        // obf: static block, fourth loop  (~var0 > -27  ↔  var0 < 26)
        for (int i = 0; i < 26; i++) {
            BASE64_ENCODE[i + 10] = (byte)('A' + i);  // 65..90
        }

        // Lowercase 'a'-'z' → indices 36-61
        // obf: static block, fifth loop
        for (int i = 0; i < 26; i++) {
            BASE64_ENCODE[i + 36] = (byte)('a' + i);  // 97..122
        }

        // Special symbols
        BASE64_ENCODE[63] = '$';        // 0x24
        BASE64_ENCODE[62] = (byte)0xA3; // '£' (163 as signed byte = -93)

        // --- Base-64 decoding table (inverse mapping) ---
        // '0'-'9' → 0-9
        // obf: static block, sixth loop  (-11 < ~var0  ↔  var0 < 10)
        for (int i = 0; i < 10; i++) {
            BASE64_DECODE['0' + i] = i;       // BASE64_DECODE[48+i] = i
        }

        // 'A'-'Z' → 10-35
        // obf: static block, seventh loop  (-27 < ~var0  ↔  var0 < 26)
        for (int i = 0; i < 26; i++) {
            BASE64_DECODE['A' + i] = i + 10;  // BASE64_DECODE[65+i] = i+10
        }

        // 'a'-'z' → 36-61
        // obf: static block, eighth loop
        for (int i = 0; i < 26; i++) {
            BASE64_DECODE['a' + i] = 36 + i;  // BASE64_DECODE[97+i] = 36+i
        }

        // '$' (0x24=36) → index 63
        BASE64_DECODE[36]  = 63;
        // '£' (0xA3=163) → index 62
        BASE64_DECODE[163] = 62;
    }

    // -----------------------------------------------------------------------
    // loadBmpImage — obf: a(int, Component, byte[])
    // -----------------------------------------------------------------------

    /**
     * Parse a raw 8-bpp Windows BMP byte array and return a fully prepared
     * AWT {@link Image}.
     *
     * BMP layout assumed by this method (matches the header format the RSC
     * client receives from the server/JAG archive):
     * <pre>
     *   Offset  Size  Field
     *    12      2    image width  (little-endian uint16)
     *    14      2    image height (little-endian uint16)
     *    18    768    palette: 256 × (B, G, R) bytes  (note: BGR order)
     *   786     w×h   pixel index data, bottom row first (standard BMP order)
     * </pre>
     *
     * The pixel rows are re-ordered from bottom-up (BMP) to top-down (AWT)
     * before being fed to the {@link IndexColorModel}.
     *
     * After constructing the {@link Image} via {@code component.createImage(da.db)}
     * (i.e. from the engine's {@link SurfaceImageProducer}), the method calls
     * {@link World#notifyImageConsumers} three times with {@code prepareImage}
     * in between, following the same pattern as {@link Surface#init} — this
     * forces the AWT to materialise the image before returning.
     *
     * @param unusedGuard  obf: var0 — dummy parameter used only in the
     *                     obfuscated anti-tamper junk expression
     *                     {@code 113 / ((var0 - 10) / 53)}, which is deleted
     *                     here.  The engine always passes a value that makes
     *                     the expression non-zero; the result is never stored.
     * @param component    AWT component used to create and prepare the image.
     * @param bmpData      raw BMP byte array (at least 786 + width*height bytes).
     * @return             the prepared AWT {@link Image}.
     *
     * obf: a(int, Component, byte[])  — obf signature: pa.B(int,Component,byte[])
     */
    static final Image loadBmpImage(int unusedGuard, Component component, byte[] bmpData) {
        // --- Parse BMP header ---
        // Width at bytes 12-13 (little-endian), stored into the shared
        // imageWidth carrier's field 'o'.
        // obf: k.o = var2[12] + var2[13] * 256
        imageWidthCarrier.imageWidth = bmpData[12] + bmpData[13] * 256;

        // Height at bytes 14-15 (little-endian), stored into ClientStream.imageHeight.
        // obf: da.bb = var2[14] + 256 * var2[15]
        client.net.ClientStream.imageHeight = bmpData[14] + 256 * bmpData[15];

        // (Profiling counter — dead artifact)
        // obf: ++i;   removed — dead instrumentation

        final int width  = imageWidthCarrier.imageWidth;
        final int height = client.net.ClientStream.imageHeight;

        // --- Decode 256-colour palette ---
        // Palette starts at byte offset 18; each entry is 3 bytes in BGR order:
        //   offset 18 + idx*3 + 0 = Blue
        //   offset 18 + idx*3 + 1 = Green
        //   offset 18 + idx*3 + 2 = Red
        // IndexColorModel(8, 256, reds, greens, blues) expects separate R/G/B arrays.
        // obf: var15=red, var4=green, var5=blue (from indices 20+idx*3, 19+idx*3, 18+idx*3)
        byte[] reds   = new byte[256];
        byte[] greens = new byte[256];
        byte[] blues  = new byte[256];

        // obf: while (-257 < ~var6)  ↔  var6 < 256
        for (int idx = 0; idx < 256; idx++) {
            reds[idx]   = bmpData[20 + idx * 3];  // R at offset 20 + idx*3
            greens[idx] = bmpData[19 + idx * 3];  // G at offset 19 + idx*3
            blues[idx]  = bmpData[18 + idx * 3];  // B at offset 18 + idx*3
        }

        // Store the IndexColorModel in the shared SocketFactory slot.
        // obf: m.d = new IndexColorModel(...)
        // m = SocketFactory; its static ColorModel d is shared with fb (SurfaceImageProducer).
        client.net.SocketFactory.colorModel = new IndexColorModel(8, 256, reds, greens, blues);

        // --- Flip pixel rows from BMP bottom-up to AWT top-down ---
        // BMP stores rows starting at the bottom; AWT wants top-first.
        // Pixel data starts at offset 786 (= 14-byte file header + 40-byte
        // DIB header + 768-byte palette).
        // obf: var16 = pixel buffer, var7 = dest cursor,
        //      var8 counts rows from (height-1) down to 0
        byte[] pixels = new byte[width * height];
        int destIdx = 0;

        // obf: for (var8 = -1 + da.bb; -1 >= ~var8; var8--)
        //      ↔  for (row = height - 1; row >= 0; row--)
        for (int row = height - 1; row >= 0; row--) {
            for (int col = 0; col < width; col++) {
                // Source offset: 786 + row * width + col
                pixels[destIdx++] = bmpData[width * row + 786 + col];
            }
        }

        // --- Create the AWT Image from the engine's ImageProducer ---
        // da.db is a SurfaceImageProducer (fb) instance, which implements
        // ImageProducer and uses the ColorModel stored above in m.d.
        // obf: var17 = component.createImage(da.db)
        Image image = component.createImage(client.net.ClientStream.imageProducer);

        // Anti-tamper junk expression — DELETE:
        // obf: int var18 = 113 / ((var0 - 10) / 53);  ← result never used,
        // param var0 is a dummy guard.  Removed entirely.

        // --- Push pixel data to ImageConsumer (3× as in Surface.init) ---
        // Each lb.a(true, pixels) call invokes:
        //   ImageConsumer.setPixels(0, 0, width, height, ColorModel, pixels, 0, width)
        //   ImageConsumer.imageComplete(STATICIMAGEDONE=3)
        // Followed by component.prepareImage() to force AWT realisation.
        // Three rounds matches the oracle Surface.init() pattern.
        // obf: lb.a(true, var16);  var1.prepareImage(var17, da.db);  (×3)
        client.world.World.notifyImageConsumers(true, pixels);
        component.prepareImage(image, client.net.ClientStream.imageProducer);

        client.world.World.notifyImageConsumers(true, pixels);
        component.prepareImage(image, client.net.ClientStream.imageProducer);

        client.world.World.notifyImageConsumers(true, pixels);
        component.prepareImage(image, client.net.ClientStream.imageProducer);

        return image;
    }

    // -----------------------------------------------------------------------
    // buildMuLawTable — obf: a(int)
    // -----------------------------------------------------------------------

    /**
     * Build and return a 256-entry amplitude expansion table using a
     * log-law (μ-law style) encoding scheme.
     *
     * Each index {@code n} in the range [-128, 127] (signed byte) is decoded
     * as a custom 8-bit floating-point value:
     * <pre>
     *   raw  = ~n                    // bitwise NOT (complement of index)
     *   sign = raw &amp; 0x80           // bit 7: 0 = positive, 1 = negative
     *   exp  = (raw &gt;&gt; 4) &amp; 7       // bits 6-4: exponent (0-7)
     *   mant = (raw &amp; 0x0F) | 0x10  // bits 3-0 + implicit leading 1
     *   val  = (2 * mant + 1) &lt;&lt; (exp + 2)  - 132
     *   if (sign == 0) val = -val
     *   table[n &amp; 0xFF] = (byte)(val / 256)
     * </pre>
     *
     * This is the standard G.711 μ-law decoder structure.  The result array
     * maps byte sample indices → signed 8-bit linear amplitudes and is used
     * by the audio subsystem.
     *
     * @param unused  obf: var0 — dummy parameter; the anti-tamper guard
     *                {@code if (var0 > -125) a(-7, null, null)} is deleted.
     * @return  256-byte expansion table, indexed by unsigned byte value.
     *
     * obf: a(int)  — obf signature: pa.A(int)
     */
    static final byte[] buildMuLawTable(int unused) {
        // (Profiling counter — dead artifact)
        // obf: ++c;  removed — dead instrumentation

        // Anti-tamper guard deleted:
        // obf: if (var0 > -125) { a(-7, null, null); }
        // This would crash loadBmpImage with a null component if var0 were
        // ever non-sentinel; the guard is always satisfied by callers.

        byte[] table = new byte[256];

        // Iterate over signed byte range [-128, 127].
        // obf: var2 = -128; while (var2 < 127) { ... var2++; }
        for (int n = -128; n < 128; n++) {
            // Complement of the index gives the encoded 8-bit float.
            // obf: var3 = (byte)var2; var3 = ~var3;
            int raw  = ~((byte) n);  // bitwise NOT of the signed byte value

            // Extract sign bit (bit 7 of the complemented value).
            // obf: var4 = var3 & 128;  (~var4 == -1) means var4 == 0 → positive
            int sign = raw & 0x80;

            // Extract 3-bit biased exponent from bits 6-4.
            // obf: var5 = var3 >> 4 & 7
            // (Vineflower shows the shift constant as a large negative due to
            // obfuscated shift-amount arithmetic; CFR correctly yields >> 4.)
            int exp  = (raw >> 4) & 7;

            // Extract 4-bit mantissa from bits 3-0, then set the implicit
            // leading 1 (IEEE-like: mantissa = 1.xxxx).
            // obf: var6 = 15 & var3; var6 |= 16;
            int mant = (raw & 0x0F) | 0x10;  // range 0x10..0x1F (16..31)

            // Expand to odd integer mantissa: (2 * mant) + 1.
            // obf: var6 = 1 + (var6 << 1)
            // (Vineflower shows the shift constant as a large negative; CFR
            // and the semantic match yield << 1.)
            mant = 1 + (mant << 1);          // range 33..63 (odd numbers)

            // Shift by (exp + 2) to get the linear magnitude, then bias by -132.
            // obf: var7 = var6 << var5 + 2;  var7 = -132 + var7;
            int val = (mant << (exp + 2)) - 132;

            // Apply sign: if sign bit was clear (sign == 0) the value is
            // negative in the μ-law convention used here.
            // obf: if (~var4 == -1) break label32;  else var7 = -var7;
            // ~var4 == -1  ↔  var4 == 0  ↔  sign bit clear → skip negation.
            // sign bit SET → positive; sign bit CLEAR → negate.
            if (sign == 0) {
                val = -val;
            }

            // Store the high byte of val as the 8-bit amplitude.
            // ib.a(n, 255) = n & 255 — converts signed loop index to unsigned
            // array index (0-255).
            // obf: var1[ib.a(var2, 255)] = (byte)(var7 / 256);
            table[n & 0xFF] = (byte)(val / 256);
        }

        return table;
    }

    // -----------------------------------------------------------------------
    // XOR string-pool helpers  (obf: private static z(String) / z(char[]))
    // -----------------------------------------------------------------------

    /**
     * First-pass XOR decoder: converts a literal String to a char[] and, if
     * the string is shorter than 2 characters, XORs the single char with 0x0D.
     * For strings of length ≥ 2 this is effectively a no-op pass whose only
     * purpose is to break simple string-constant analysis.
     *
     * @param encoded  obfuscated literal string from the constant pool.
     * @return         mutable char[] ready for the second-pass decoder.
     *
     * obf: z(String) → char[]
     */
    private static char[] decodeXorCharArr(String encoded) {
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char)(chars[0] ^ '\r');  // 0x0D
        }
        return chars;
    }

    /**
     * Second-pass XOR decoder: applies a 5-byte rotating key
     * {@code [58, 62, 34, 75, 13]} to each character in the array, then
     * interns and returns the resulting String.
     *
     * Key derivation (from switch cases):
     * <pre>
     *   i % 5 == 0 → key = 58  (':')
     *   i % 5 == 1 → key = 62  ('>')
     *   i % 5 == 2 → key = 34  ('"')
     *   i % 5 == 3 → key = 75  ('K')
     *   i % 5 == 4 → key = 13  (CR)
     * </pre>
     *
     * @param chars  mutable char[] from {@link #decodeXorCharArr}.
     * @return       decoded, interned String.
     *
     * obf: z(char[]) → String
     */
    private static String decodeXorStr(char[] chars) {
        final int[] KEY = {58, 62, 34, 75, 13};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char)(chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
