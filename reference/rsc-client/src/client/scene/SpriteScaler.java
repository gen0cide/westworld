package client.scene;

/**
 * SpriteScaler — scanline sprite pixel-doubling scaler for the RSC software renderer.
 *
 * <p>This class contains the performance-critical inner loop that writes scaled/zoomed
 * sprite pixels from a palette (colour-indexed sprite) into a destination pixel buffer.
 * The core method uses 8.8 fixed-point arithmetic to step through source pixels at an
 * arbitrary fractional rate, writing each source colour twice (pixel-doubling) so that
 * two consecutive destination pixels share one palette entry. An 8-pixel unrolled main
 * loop handles the bulk of each scanline, followed by a 1–7 pixel remainder loop.
 *
 * <p>The class also houses a collection of global state fields that were bundled here by
 * the obfuscator but conceptually belong to several different subsystems:
 * player-name tables used by {@code Mudclient}, a per-byte frequency table used by
 * {@code Packet}/{@code Mudclient} for protocol diagnostics, a sprite-count integer
 * used during asset loading, a UI line-height override used by {@code Panel}, a model-name
 * count used by {@code GameModel}/{@code NameTable}, and three per-method profiling
 * counters (dead code preserved for traceability).
 *
 * <p>Obfuscation layers stripped:
 * <ul>
 *   <li>Opaque predicate {@code boolean bl = client.vh} and all dead {@code if(bl)} branches.</li>
 *   <li>Profiling increments {@code ++e}, {@code ++c}, {@code ++f} (counters kept as fields).</li>
 *   <li>Exception wrappers {@code try{…}catch(RuntimeException e){throw ErrorHandler.a(…)}}.</li>
 *   <li>Anti-tamper junk parameter {@code byte var7}/{@code by} and its modulo expression
 *       {@code 69 % ((var7-27)/45)} which is dead (result unused).</li>
 *   <li>Obfuscated bit masks on palette lookups: all five distinct {@code (0xFFxx & pos)>>8}
 *       variants are identical in effect to {@code (pos >> 8) & 0xFF} for values in [0,65535];
 *       the low-byte mask bits are noise (see inline comments).</li>
 *   <li>XOR string pool {@code z[]} decoded to literals.</li>
 * </ul>
 *
 * Obfuscated class name: {@code ia}
 * Package placement: {@code client.scene} (per NAMING.md)
 * Oracle reference: {@code mudclient204/src/Surface.java} (drawSprite / plotScale methods)
 */
final class SpriteScaler {

    // -------------------------------------------------------------------------
    // Profiling counters (dead code — never read back in meaningful logic;
    // the obfuscator inserts "++<counter>" at each method entry for profiling).
    // -------------------------------------------------------------------------

    /** Profiling invocation counter for {@link #writePaletteScaledScanline}. // obf: e */
    static int profilingCounter_B;  // obf: e

    /** Profiling invocation counter for {@link #readPacketString}. // obf: c */
    static int profilingCounter_C;  // obf: c

    /** Profiling invocation counter for {@link #isChatCipherKnown}. // obf: f */
    static int profilingCounter_A;  // obf: f

    // -------------------------------------------------------------------------
    // Player-name tables  (populated/consumed by Mudclient, bundled here by obfuscator)
    // -------------------------------------------------------------------------

    /**
     * Player username list; 100 slots; written when players enter or update.
     * Indexed in parallel with {@link #playerTitles}.
     * // obf: a
     */
    static String[] playerNames = new String[100];  // obf: a

    /**
     * Player title/status list; 100 slots; parallel to {@link #playerNames}.
     * // obf: g
     */
    static String[] playerTitles = new String[100];  // obf: g

    // -------------------------------------------------------------------------
    // Asset-loading integers (set during game data initialisation by SocketFactory / m.java)
    // -------------------------------------------------------------------------

    /**
     * Total number of sprites loaded from the JAG archive.
     * Set via {@code EntityDef.a(65525)} during content initialisation;
     * used to size {@code ClientStream.N[]}, {@code DecodeBuffer.l[]}, {@code Panel.K[]}.
     * // obf: h
     */
    static int spriteCount;  // obf: h

    /**
     * Count of named 3-D model entries registered in {@code NameTable.c[]}.
     * Written and read by {@code GameModel} (ca.java) during model-name lookup.
     * // obf: b
     */
    static int modelNameCount = 0;  // obf: b

    // -------------------------------------------------------------------------
    // UI / rendering state
    // -------------------------------------------------------------------------

    /**
     * Transient line-height override used by {@code Panel} (qa.java) during text layout.
     * Set to 2 before certain draw calls and reset to 0 immediately after.
     * // obf: i
     */
    static int lineHeightOverride = 0;  // obf: i

    // -------------------------------------------------------------------------
    // Protocol / cipher support
    // -------------------------------------------------------------------------

    /**
     * Per-opcode frequency histogram, 256 buckets.
     * Each entry is incremented by {@code Packet} (b.java) when a packet is finished,
     * keyed by the third byte of the packet frame.  Used for protocol diagnostics.
     * // obf: d
     */
    static int[] opcodeFrequencyTable = new int[256];  // obf: d

    // -------------------------------------------------------------------------
    // XOR-encoded error-message string pool (private; decoded at class-load time)
    // -------------------------------------------------------------------------

    /**
     * Decoded strings used in {@code ErrorHandler} rethrow messages.
     * Encoding: first pass XORs with 'W' (0x57) if length < 2;
     * second pass XORs each char with key[i % 5] = {72, 121, 74, 114, 87}.
     * <pre>
     *   z[0] = "{...}"   — non-null argument placeholder
     *   z[1] = "null"    — null argument placeholder
     *   z[2] = "ia.B("   — error prefix for writePaletteScaledScanline (method B)
     *   z[3] = "ia.C("   — error prefix for readPacketString (method C)
     *   z[4] = "ia.A("   — error prefix for isChatCipherKnown (method A)
     * </pre>
     * // obf: z
     */
    private static final String[] ERROR_STRINGS = new String[]{
        "{...}",   // obf: z[0]  — decoded from z(z("3Wd\\*"))
        "null",    // obf: z[1]  — decoded from z(z("&\f&"))
        "ia.B(",   // obf: z[2]  — decoded from z(z("!d0"))
        "ia.C(",   // obf: z[3]  — decoded from z(z("!d1"))
        "ia.A(",   // obf: z[4]  — decoded from z(z("!d3"))
    };

    // =========================================================================
    // CORE METHOD — scanline sprite scaler
    // =========================================================================

    /**
     * Writes a single scaled/zoomed scanline from a palette-indexed sprite into a
     * destination pixel buffer, doubling each source pixel horizontally.
     *
     * <p><b>Fixed-point convention:</b> {@code srcStep} and {@code srcPos} use an
     * 8.8 fixed-point format: the high 8 bits are the integer source-pixel index
     * (0–255), and the low 8 bits are the sub-pixel fraction (nearest-neighbour,
     * so the fraction is not used for interpolation — only the integer part matters
     * for the palette lookup).
     *
     * <p><b>Pixel doubling:</b> The method internally doubles {@code srcStep} via
     * {@code srcStep <<= 1}, then advances the source position by this doubled step
     * once per pair of output pixels.  Each palette entry therefore produces exactly
     * two consecutive identical output pixels, implementing 2× nearest-neighbour
     * zoom for the caller's effective step value.
     *
     * <p><b>Main loop (8-pixel unroll):</b> Handles {@code -(pixelCount / 8)} groups
     * of 8 output pixels (4 source pixels per group). The loop is manually unrolled
     * and the four source advances are interleaved with the eight output writes to
     * reduce pipeline stalls on the Microsoft JVM's JIT.
     *
     * <p><b>Obfuscation note — varied bitmasks:</b> The five palette lookups inside
     * the main loop use five different low-byte AND masks (0xFFB9, 0xFFDF, 0xFF6F,
     * 0xFF16, 0xFF6A) before the {@code >> 8} right-shift. All masks have the high
     * byte = 0xFF, so once shifted right by 8 the low-byte mask bits are irrelevant
     * (they only ever affect bits 0–7 which are shifted out). All five expressions
     * are therefore equivalent to {@code (srcPos >> 8) & 0xFF}.
     *
     * <p><b>Obfuscation note — dropped parameter:</b> The original method had a
     * trailing {@code byte var7} parameter that appeared only in a dead junk modulo
     * expression {@code 69 % ((var7-27)/45)}; it has been removed from this
     * deobfuscated version.
     *
     * @param srcStep   source advance per output pixel pair, 8.8 fixed-point.
     *                  Doubled internally; caller passes half the desired step.
     *                  // obf: var0
     * @param currentPixel  initially unused; immediately overwritten on entry with
     *                      the first palette lookup result. // obf: var1
     * @param palette   RGB colour table (256 entries) for the sprite.
     *                  Indexed by the high byte of {@code srcPos}. // obf: var2
     * @param srcPos    current source position in 8.8 fixed-point.
     *                  High byte = palette index; low byte = fractional position.
     *                  // obf: var3
     * @param dstOffset write index into {@code dst}. // obf: var4
     * @param dst       destination pixel buffer (screen or off-screen surface).
     *                  // obf: var5
     * @param negPixelCount  NEGATIVE total pixel count to write (loop guard style:
     *                  loops while {@code 0 > negPixelCount}).  E.g. pass -128 to
     *                  write 128 output pixels. // obf: var6
     *
     * Obfuscated method signature: {@code static final void a(int,int,int[],int,int,int[],int,byte)}
     * (the trailing {@code byte} was a dead anti-tamper param, dropped here)
     */
    static final void writePaletteScaledScanline(
            int srcStep,        // obf: var0
            int currentPixel,   // obf: var1  (immediately overwritten)
            int[] palette,      // obf: var2
            int srcPos,         // obf: var3
            int dstOffset,      // obf: var4
            int[] dst,          // obf: var5
            int negPixelCount   // obf: var6
    ) {
        // obf: dropped: byte var7 (dead anti-tamper param; only appears in
        //      junk expression "69 % ((var7-27)/45)" whose result is unused)

        // Guard: only proceed if there are pixels to write (caller passes negative count).
        if (negPixelCount >= 0) {
            return;
        }

        // --- Initialise ---

        // Load the very first source pixel colour from the palette.
        // The mask (0xFFB9 & srcPos) >> 8 is equivalent to (srcPos >> 8) & 0xFF;
        // the low-byte mask bits (0xB9) are obfuscation noise — they are shifted out.
        currentPixel = palette[(srcPos >> 8) & 0xFF];  // obf: palette[(65465 & var3) >> 8]

        // Double the step so that each source advance covers two output pixels.
        srcStep <<= 1;  // obf: var0 <<= 1

        // Advance source position by the (now doubled) step.
        srcPos += srcStep;

        // --- Main loop: 8-pixel (4-source-pixel) unrolled groups ---

        // nMainGroups is negative (var6/8 with var6 < 0), loop runs while 0 > nMainGroups.
        int nMainGroups = negPixelCount / 8;  // obf: var17 = var6 / 8

        for (int grp = nMainGroups; grp < 0; grp++) {  // obf: var9 starts at var17, increments

            // --- Output pixels 0 & 1: from preloaded currentPixel ---
            dst[dstOffset++] = currentPixel;
            dst[dstOffset++] = currentPixel;

            // Load source pixel A; advance source.
            // Mask 0xFFDF: obfuscation noise (≡ (srcPos >> 8) & 0xFF)
            currentPixel = palette[(srcPos >> 8) & 0xFF];  // obf: palette[(65503 & var3) >> 8]
            srcPos += srcStep;

            // --- Output pixels 2 & 3: from source pixel A ---
            dst[dstOffset++] = currentPixel;
            dst[dstOffset++] = currentPixel;

            // Load source pixel B (no advance yet — advance is interleaved with writes below).
            // Mask 0xFF6F: obfuscation noise (≡ (srcPos >> 8) & 0xFF)
            currentPixel = palette[(srcPos >> 8) & 0xFF];  // obf: palette[(var3 & 65391) >> 8]

            // --- Output pixel 4: from source pixel B ---
            dst[dstOffset++] = currentPixel;

            // Advance source *between* the two B-pixel writes (pipeline scheduling artefact).
            srcPos += srcStep;

            // --- Output pixel 5: still from source pixel B ---
            dst[dstOffset++] = currentPixel;

            // Load source pixel C; advance source.
            // Mask 0xFF16: obfuscation noise (≡ (srcPos >> 8) & 0xFF)
            currentPixel = palette[(srcPos >> 8) & 0xFF];  // obf: palette[(65302 & var3) >> 8]
            srcPos += srcStep;

            // --- Output pixels 6 & 7: from source pixel C ---
            dst[dstOffset++] = currentPixel;
            dst[dstOffset++] = currentPixel;

            // Preload source pixel D (will be used as 'currentPixel' in the next iteration).
            // Mask 0xFF6A: obfuscation noise (≡ (srcPos >> 8) & 0xFF)
            currentPixel = palette[(srcPos >> 8) & 0xFF];  // obf: palette[(var3 & 65386) >> 8]
            srcPos += srcStep;

            // (grp++ happens at for-loop increment)
        }

        // --- Remainder loop: 0–7 leftover pixels ---

        // nRemainder = number of pixels not covered by the 8-pixel groups.
        // negPixelCount % 8 is in range (-7, 0], so negating gives [0, 7).
        int nRemainder = -(negPixelCount % 8);  // obf: var17 = -(var6 % 8)

        // Loop condition:  ~rem > ~count  ≡  count < rem  (via ~x = -x-1 arithmetic)
        // i.e. iterate while rem_index < nRemainder.
        for (int remIdx = 0; remIdx < nRemainder; remIdx++) {  // obf: var9 = 0; ~var9 > ~var17

            // Write current pixel.
            dst[dstOffset++] = currentPixel;

            // On every odd iteration, advance the source position and load the next
            // palette colour.  On even iterations the same colour is written again
            // (pixel doubling). This mirrors the main loop's 2-output-per-source behaviour.
            if ((remIdx & 1) == 1) {
                currentPixel = palette[(srcPos >> 8) & 0xFF];  // obf: palette[var3 >> 8 & 0xFF]
                srcPos += srcStep;
            }
        }
    }

    // =========================================================================
    // STRING / PACKET HELPER — delegates to Mudclient
    // =========================================================================

    /**
     * Reads a length-prefixed string from a {@code Buffer} packet and returns it as a
     * {@code String}.  Optionally clears the {@link #playerNames} array before decoding.
     *
     * <p>Delegates entirely to {@code Mudclient.a(int, Buffer, int)} (the static
     * packet-string reader in the main client class), passing {@code type=0} and
     * {@code maxLen=32767 (Short.MAX_VALUE)}.
     *
     * <p>This method is logically part of the player-name protocol handler and landed
     * in this class due to the obfuscator's method-distribution pass.
     *
     * @param buf     the packet buffer to read from. // obf: var0 (type tb = Buffer)
     * @param clearNames  if {@code true}, sets {@link #playerNames} to {@code null}
     *                    before reading (used to reset the name table on reinit).
     *                    // obf: var1
     * @return the decoded string from the buffer.
     *
     * Obfuscated method signature: {@code static final String a(tb, boolean)}
     */
    static final String readPacketString(client.net.Buffer buf, boolean clearNames) {
        // obf: ++c  (dead profiling counter, kept as profilingCounter_C)

        if (clearNames) {
            // Caller signals a name-table reset.
            playerNames = null;  // obf: a = null
        }

        // Delegate to the static packet-string decoder in Mudclient.
        // Mudclient.a(0, buf, Short.MAX_VALUE) reads a 2-byte length then that many chars.
        return client.Mudclient.a(0, buf, Short.MAX_VALUE);  // obf: client.a(0, var0, 32767)
    }

    // =========================================================================
    // CHAT CIPHER IDENTITY CHECK — enum replacement
    // =========================================================================

    /**
     * Returns {@code true} if the given {@code ChatCipher} instance is one of the five
     * well-known static cipher singletons used by the engine.
     *
     * <p>In pre-Java-5 RSC, Java enums did not exist. Instead, singleton instances of
     * {@code ChatCipher} (class {@code v}) were stored as static fields in various
     * classes; this method replaces what would today be an {@code EnumSet} membership
     * test.  The five recognised singletons are:
     * <ul>
     *   <li>{@code ClientStream.O}  (obf: {@code da.O})</li>
     *   <li>{@code CharTable.c}     (obf: {@code ga.c})</li>
     *   <li>{@code GameCharacter.f} (obf: {@code ta.f})</li>
     *   <li>{@code AudioMixer.d}    (obf: {@code eb.d})</li>
     *   <li>{@code ProxySocketFactory.n} (obf: {@code gb.n})</li>
     * </ul>
     *
     * @param cipher  the {@code ChatCipher} instance to test. // obf: param0 (type v)
     * @return {@code true} if {@code cipher} is one of the five known singletons.
     *
     * Obfuscated method signature: {@code static final boolean a(v, byte)}
     * (the trailing {@code byte} was a dead anti-tamper param, dropped here)
     */
    static final boolean isChatCipherKnown(client.net.ChatCipher cipher) {
        // obf: dropped: byte param1 (dead param; only in junk: "3 / ((-39 - by) / 54)")
        // obf: ++f  (dead profiling counter, kept as profilingCounter_A)

        // Identity comparisons against the five known ChatCipher singletons.
        if (client.net.ClientStream.O == cipher)        return true;  // obf: da.O
        if (client.ui.CharTable.c    == cipher)         return true;  // obf: ga.c
        if (client.world.GameCharacter.f == cipher)     return true;  // obf: ta.f
        if (cipher == client.audio.AudioMixer.d)        return true;  // obf: eb.d
        if (client.net.ProxySocketFactory.n == cipher)  return true;  // obf: gb.n

        return false;
    }

    // =========================================================================
    // XOR STRING-POOL DECODERS (private helpers, preserved for completeness)
    // =========================================================================

    /**
     * First decoding pass: converts a raw String to a char array, XOR-ing the single
     * character with {@code 'W'} (0x57) if the string has fewer than 2 characters.
     *
     * // obf: private static char[] z(String)
     */
    private static char[] decodePass1(String encoded) {  // obf: z(String)
        char[] arr = encoded.toCharArray();
        if (arr.length < 2) {
            arr[0] = (char)(arr[0] ^ 'W');  // 'W' = 0x57
        }
        return arr;
    }

    /**
     * Second decoding pass: XOR-decodes a char array using the 5-byte rotating key
     * {@code {72, 121, 74, 114, 87}} (i.e. {@code "HyJrW"}) and interns the result.
     *
     * <pre>
     *   key[0] = 72  ('H')
     *   key[1] = 121 ('y')
     *   key[2] = 74  ('J')
     *   key[3] = 114 ('r')
     *   key[4] = 87  ('W')
     * </pre>
     *
     * // obf: private static String z(char[])
     */
    private static String decodePass2(char[] arr) {  // obf: z(char[])
        final int[] KEY = {72, 121, 74, 114, 87};  // "HyJrW"
        for (int i = 0; i < arr.length; i++) {
            arr[i] = (char)(arr[i] ^ KEY[i % 5]);
        }
        return new String(arr).intern();
    }
}
