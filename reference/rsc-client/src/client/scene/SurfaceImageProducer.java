package client.scene;

import java.awt.Image;
import java.awt.image.ImageConsumer;
import java.awt.image.ImageObserver;
import java.awt.image.ImageProducer;

/**
 * SurfaceImageProducer — the singleton {@link ImageProducer}/{@link ImageObserver}
 * that feeds AWT's image pipeline for the game's software-rendered 2D surface.
 *
 * <p>In rev-235 the pixel buffer and render state live in {@code Scene} (lb) /
 * {@code Surface} (ua), but AWT's image plumbing requires a separate
 * {@code ImageProducer} object.  A single static instance of this class is
 * held in {@code ClientStream.imageProducer} (obf {@code da.db}) and passed to
 * {@code Component.createImage(ImageProducer)} so that AWT can hand a live
 * {@link Image} back to the engine.
 *
 * <p>The class also implements {@code ImageObserver} so it can be passed to
 * {@code Component.prepareImage(image, observer)} without creating a second
 * object.  {@link #imageUpdate} always returns {@code true} (keep notifying),
 * the correct no-op observer contract.
 *
 * <p>Dimension/color-model globals written by {@code ImageLoader} (pa) when
 * it decodes a BMP header:
 * <ul>
 *   <li>{@code World.surfaceWidth}  (obf {@code k.o})  — pixel columns</li>
 *   <li>{@code ClientStream.surfaceHeight} (obf {@code da.bb}) — pixel rows</li>
 *   <li>{@code SocketFactory.colorModel} (obf {@code m.d}) — 8-bit indexed
 *       palette {@code IndexColorModel} built from the BMP palette block</li>
 * </ul>
 *
 * <p>The active {@code ImageConsumer} is stored in the shared static slot
 * {@code StringCodec.imageConsumer} (obf {@code u.d}), reused by the
 * {@code Scene} (obf {@code lb}) pixel-push path that calls
 * {@code imageConsumer.setPixels(...)}.
 *
 * <p>This class also acts as a global bag for two unrelated statics that
 * happen to have no better home:
 * <ul>
 *   <li>{@link #bzip} (obf {@code aa a}) — the singleton {@code BZip}
 *       decompressor used by {@code Mudclient} to decompress game-data
 *       payloads.  Accessed as {@code fb.a} from client.java.</li>
 *   <li>{@link #errorHandler} (obf {@code i h}) — an {@code ErrorHandler}
 *       instance initialised with the decoded name {@code "RC"}.  This mirrors
 *       the pattern seen in other classes where the obfuscator planted an
 *       {@code ErrorHandler} as a class-level sentinel.</li>
 *   <li>{@link #entityIndexTableF}, {@link #entityIndexTableD},
 *       {@link #entityIndexTableC} (obf {@code int[] f/d/c}) — index/lookup
 *       arrays for game-entity tables in {@code Mudclient}.</li>
 * </ul>
 *
 * <h3>Obfuscation artefacts removed</h3>
 * <ul>
 *   <li>Per-method profiling counters: obf fields {@code e, i, j, g, b, l}
 *       (all incremented exactly once at method entry and never read for
 *       behaviour).  All six are omitted.</li>
 *   <li>{@code try/catch(RuntimeException)} wrappers that rethrow via
 *       {@code ErrorHandler.a(e, "sig(...)")}.</li>
 *   <li>XOR-encrypted string pool {@code z[]} used only in those rethrow
 *       messages (decoded values documented below).</li>
 *   <li>Static {@code boolean[] k} — 12-element array initialised to
 *       {@code false}.  Never mutated; only referenced inside dead error
 *       branches.  Omitted.</li>
 *   <li>The two private static {@code z(...)} XOR-decode helpers — no longer
 *       needed once string literals are decoded.</li>
 * </ul>
 *
 * <h3>XOR string pool — decoded values</h3>
 * <pre>
 *   Key pass 1: z(String) → char[] — XOR sole char with 0x4A only when len &lt; 2.
 *   Key pass 2: z(char[]) → String — XOR char[i] with {95,14,95,62,74}[i % 5].
 *
 *   z[0] = "null"
 *   z[1] = "{...}"
 *   z[2] = "fb.imageUpdate("
 *   z[3] = "fb.removeConsumer("
 *   z[4] = "fb.requestTopDownLeftRightResend("
 *   z[5] = "fb.isConsumer("
 *   z[6] = "fb.addConsumer("
 *   z[7] = "fb.startProduction("
 *
 *   h-init: z(z("\rM")) = "RC"    (0x0D^0x5F='R', 0x4D^0x0E='C')
 * </pre>
 *
 * obf class: {@code fb}
 */
final class SurfaceImageProducer implements ImageProducer, ImageObserver {

    // -------------------------------------------------------------------------
    // Global state — ErrorHandler sentinel
    // -------------------------------------------------------------------------

    /**
     * Class-level {@code ErrorHandler} sentinel initialised with the decoded
     * name {@code "RC"} and priority {@code 1}.  The decoded name comes from
     * the XOR string {@code z(z("\rM"))}: {@code 0x0D^0x5F='R'}, {@code 0x4D^0x0E='C'}.
     *
     * <p>This field is accessed by the outer engine via bytecode reference
     * {@code fb.h} but never read for semantic logic in any decompiled method.
     * Its role appears to be the standard per-class error-context tag used by
     * the obfuscator's rethrow infrastructure.
     *
     * obf: {@code static i h = new i(z(z("\rM")), 1);}
     */
    static ErrorHandler errorHandler = new ErrorHandler(
        /* name=  */ "RC",  // decoded from z(z("\rM"))
        /* param= */ 1
    );

    // -------------------------------------------------------------------------
    // Global state — BZip decompressor (unrelated to the image pipeline)
    // -------------------------------------------------------------------------

    /**
     * Singleton BZip2-style decompressor shared across the engine.
     *
     * <p>Stored here purely as a convenient global slot.  Accessed from
     * {@code Mudclient} (client.java bytecode offset 0x17010) to decompress
     * game-data payloads via {@code bzip.a(byte[], offset, byte[], …)}.
     *
     * obf: {@code static aa a;}
     */
    static BZip bzip;    // obf: aa a

    // -------------------------------------------------------------------------
    // Global state — entity-index lookup arrays (Mudclient scratch tables)
    // -------------------------------------------------------------------------

    /**
     * Game-entity index lookup table, used as an int[] array index into
     * model/entity arrays in {@code Mudclient}.
     * Exact semantics depend on surrounding context in client.java.
     *
     * obf: {@code static int[] f;}
     */
    static int[] entityIndexTableF;   // obf: int[] f

    /**
     * Game-entity index lookup table (second table).
     *
     * obf: {@code static int[] d;}
     */
    static int[] entityIndexTableD;   // obf: int[] d

    /**
     * Game-entity index lookup table (third table).
     *
     * obf: {@code static int[] c;}
     */
    static int[] entityIndexTableC;   // obf: int[] c

    // -------------------------------------------------------------------------
    // Dead/omitted fields (documented for bytecode traceability)
    // -------------------------------------------------------------------------
    //
    // The following static ints are profiling counters: each is incremented
    // once at the top of its corresponding ImageProducer method and never read
    // for any real purpose.  They are omitted from the deobfuscated body.
    //
    //   obf fb.e  (int)  — profiling counter for addConsumer
    //   obf fb.i  (int)  — profiling counter for startProduction
    //                      CAUTION: same single-letter name as class i (ErrorHandler)
    //                      in the obfuscated source — it is a plain int, not a class ref.
    //   obf fb.j  (int)  — profiling counter for requestTopDownLeftRightResend
    //   obf fb.g  (int)  — profiling counter for removeConsumer
    //   obf fb.b  (int)  — profiling counter for isConsumer
    //   obf fb.l  (int)  — profiling counter for imageUpdate
    //
    // Also omitted:
    //   obf fb.k  (boolean[12]{false,…}) — dead padding array, never mutated.

    // =========================================================================
    // ImageProducer interface
    // =========================================================================

    /**
     * Registers an {@link ImageConsumer} and pushes the image metadata it
     * needs before pixel delivery can begin.
     *
     * <p>The consumer is stored in the shared slot
     * {@code StringCodec.imageConsumer} (obf {@code u.d}), which is also
     * read by the {@code Scene} (obf {@code lb}) pixel-push path:
     * {@code u.d.setPixels(0, 0, k.o, da.bb, m.d, pixels, 0, k.o)}.
     *
     * <p>Metadata pushed to the consumer:
     * <ol>
     *   <li><b>Dimensions:</b> {@code World.surfaceWidth × ClientStream.surfaceHeight}
     *       (obf {@code k.o × da.bb}), set by {@code ImageLoader} from BMP
     *       header bytes 12–15.</li>
     *   <li><b>Properties:</b> {@code null} — no custom property map.</li>
     *   <li><b>Color model:</b> {@code SocketFactory.colorModel} (obf {@code m.d}),
     *       an {@code IndexColorModel(8, 256, r[], g[], b[])} built by
     *       {@code ImageLoader} from the BMP palette block (offsets 18–20 per
     *       palette entry).</li>
     *   <li><b>Hints:</b> {@code 14} = {@code TOPDOWNLEFTRIGHT}(4) |
     *       {@code COMPLETESCANLINES}(2) | {@code SINGLEFRAME}(8).
     *       Tells AWT the entire frame arrives in one top-down scan.</li>
     * </ol>
     *
     * obf: {@code fb.addConsumer(ImageConsumer)}
     */
    @Override
    public final synchronized void addConsumer(ImageConsumer consumer) {
        // Store consumer globally for use by Scene's pixel-push path (lb → u.d).
        StringCodec.imageConsumer = consumer;         // obf: u.d = var1

        // Surface dimensions from the BMP header parsed by ImageLoader (pa).
        consumer.setDimensions(
            World.surfaceWidth,            // obf: k.o  — BMP width  (bytes 12-13)
            ClientStream.surfaceHeight     // obf: da.bb — BMP height (bytes 14-15)
        );

        // No custom property map needed.
        consumer.setProperties(null);

        // 8-bit indexed palette color model (256-entry, from BMP palette).
        consumer.setColorModel(SocketFactory.colorModel);  // obf: m.d

        // Hints: TOPDOWNLEFTRIGHT(4) | COMPLETESCANLINES(2) | SINGLEFRAME(8) = 14.
        // Informs AWT the frame is delivered as a single contiguous top-down pass.
        consumer.setHints(14);
    }

    /**
     * Starts pixel production by registering the consumer and pushing metadata.
     *
     * <p>This producer has no background delivery thread; pixels are pushed
     * synchronously by the render loop via
     * {@code StringCodec.imageConsumer.setPixels(...)}.  So {@code startProduction}
     * is simply {@link #addConsumer}: register and send the metadata.
     *
     * obf: {@code fb.startProduction(ImageConsumer)}
     */
    @Override
    public final void startProduction(ImageConsumer consumer) {
        this.addConsumer(consumer);
    }

    /**
     * No-op top-down-left-right resend request.
     *
     * <p>The oracle ({@code Surface.requestTopDownLeftRightResend}) logs
     * "TDLR" and does nothing.  This version is the same: pixel data is always
     * delivered in top-down order by the render loop, so resending is
     * unnecessary.
     *
     * obf: {@code fb.requestTopDownLeftRightResend(ImageConsumer)}
     */
    @Override
    public final void requestTopDownLeftRightResend(ImageConsumer consumer) {
        // Intentional no-op — pixel delivery is always top-down; resend not needed.
    }

    /**
     * Removes the given consumer if it matches the currently registered one.
     *
     * <p>Performs an identity check against
     * {@code StringCodec.imageConsumer} (obf {@code u.d}); if equal, clears
     * the slot to {@code null}.
     *
     * obf: {@code fb.removeConsumer(ImageConsumer)}
     */
    @Override
    public final synchronized void removeConsumer(ImageConsumer consumer) {
        if (StringCodec.imageConsumer == consumer) {   // obf: u.d != var1 → early return
            StringCodec.imageConsumer = null;           // obf: u.d = null
        }
    }

    /**
     * Returns {@code true} if the given consumer is the one currently registered.
     *
     * <p>Simple identity comparison against
     * {@code StringCodec.imageConsumer} (obf {@code u.d}).
     *
     * obf: {@code fb.isConsumer(ImageConsumer)}
     */
    @Override
    public final synchronized boolean isConsumer(ImageConsumer consumer) {
        return consumer == StringCodec.imageConsumer;  // obf: var1 == u.d
    }

    // =========================================================================
    // ImageObserver interface
    // =========================================================================

    /**
     * Always returns {@code true}, keeping AWT image-loading notifications
     * active.
     *
     * <p>This instance is passed as the {@code ImageObserver} argument to
     * {@code Component.prepareImage(image, this)}.  Returning {@code true}
     * tells AWT to keep delivering status updates; no action is taken on the
     * individual {@code infoflags} bits because the client drives its own
     * render loop rather than reacting to observer callbacks.
     *
     * @param image     the AWT {@link Image} being tracked
     * @param infoflags combination of {@link ImageObserver} flag constants
     *                  ({@code WIDTH}, {@code HEIGHT}, {@code ALLBITS}, etc.)
     * @param x         left edge of the updated rectangle
     * @param y         top edge of the updated rectangle
     * @param width     width of the updated rectangle in pixels
     * @param height    height of the updated rectangle in pixels
     * @return always {@code true} — continue notifications
     *
     * obf: {@code fb.imageUpdate(Image, int, int, int, int, int)}
     */
    @Override
    public final boolean imageUpdate(
            Image image, int infoflags, int x, int y, int width, int height) {
        return true;
    }

    // =========================================================================
    // Private helpers — XOR string-pool decryption
    // (Identical two-overload z() pattern used in every class in this build.)
    //
    // These decode the per-class error-message string pool z[] and the "RC"
    // ErrorHandler name.  The pool itself is only consumed by the stripped
    // exception-rethrow branches, but the two methods are real members of the
    // class and are transcribed here in full.
    //
    // Decoded string pool for this class (output of the two passes below):
    //   z[0] = "null"
    //   z[1] = "{...}"
    //   z[2] = "fb.imageUpdate("
    //   z[3] = "fb.removeConsumer("
    //   z[4] = "fb.requestTopDownLeftRightResend("
    //   z[5] = "fb.isConsumer("
    //   z[6] = "fb.addConsumer("
    //   z[7] = "fb.startProduction("
    //   h-init name: z(z("\rM")) = "RC"
    // =========================================================================

    /**
     * First pass of the two-pass XOR decryption used for the per-class string
     * pool.  Converts a raw obfuscated {@code String} literal to a
     * {@code char[]}.  If the string has fewer than 2 characters, the single
     * character is XOR'd with {@code 'J'} (0x4A); longer strings pass through
     * unchanged.
     *
     * <p>obf: {@code private static char[] z(String)}
     */
    @SuppressWarnings("unused")
    private static char[] decodeXorFirst(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 0x4A); // 74 = 'J'
        }
        return chars;
    }

    /**
     * Second pass of the two-pass XOR decryption: XORs each {@code char[i]} with
     * {@code KEY[i % 5]} and returns the result as an interned String.
     *
     * <p>Mod-5 XOR key table; positions 0..4 cycle:
     * <ul>
     *   <li>0 → 0x5F (95 = '_')</li>
     *   <li>1 → 0x0E (14)</li>
     *   <li>2 → 0x5F (95 = '_')</li>
     *   <li>3 → 0x3E (62 = '&gt;')</li>
     *   <li>4 → 0x4A (74 = 'J')</li>
     * </ul>
     *
     * <p>obf: {@code private static String z(char[])}
     */
    @SuppressWarnings("unused")
    private static String decodeXor(char[] chars) {
        final byte[] KEY = {95, 14, 95, 62, 74};
        for (int i = 0; i < chars.length; i++) {
            // obf source unrolled this as a switch(i % 5) over the five key
            // constants; the table form above is the canonical equivalent.
            chars[i] = (char) (chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
