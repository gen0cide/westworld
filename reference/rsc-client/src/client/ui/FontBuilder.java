package client.ui;

import java.awt.Color;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics;
import java.awt.Image;
import java.awt.image.PixelGrabber;
import java.math.BigInteger;

/**
 * FontBuilder — obf: {@code s}
 *
 * <p>Rasterizes individual AWT font glyphs into a packed, greyscale bitmap representation
 * used by the RSC software renderer ({@link client.scene.Surface}). For each glyph it:
 * <ol>
 *   <li>Creates an off-screen AWT image, draws the character white-on-black.</li>
 *   <li>Grabs the raw ARGB pixel array via {@link PixelGrabber}.</li>
 *   <li>Trims the tight ink bounding-box (skipping fully-black border rows/cols).</li>
 *   <li>Writes a 9-byte header record into {@link client.shell.GameFrame#fontData} at slot
 *       {@code charSlot * 9}, followed by the raw per-pixel alpha bytes at the current
 *       write cursor {@link client.net.Packet#writePos}.</li>
 *   <li>Marks {@link client.scene.SurfaceImageProducer#hasAntiAlias}[fontId] if any pixel
 *       is in the anti-aliasing grey range (30 < pixel < 230).</li>
 * </ol>
 *
 * <p>The glyph records produced here are consumed by
 * {@code Surface.drawCharacter()} / {@code Surface.drawString()} and the per-char
 * width table in {@link client.data.FontWidths}.
 *
 * <h2>Glyph record layout (9 bytes at {@code fontData[charSlot * 9]})</h2>
 * <pre>
 *   [0]  bitmapOffsetHi   = writePos / 16384          (signed byte)
 *   [1]  bitmapOffsetMid  = (writePos / 128) &amp; 0x7F   (signed byte)
 *   [2]  bitmapOffsetLo   = writePos &amp; 0x7F           (signed byte)
 *   [3]  glyphWidth       = rightInk  - leftInk        (signed byte)
 *   [4]  glyphHeight      = bottomInk - topInk         (signed byte)
 *   [5]  xBearing         = leftInk                    (signed byte, left bearing)
 *   [6]  yBearing         = maxAscent - topInk         (signed byte, top bearing from baseline)
 *   [7]  advanceWidth     = fontMetrics.charWidth(ch)  (signed byte, unchanged by italic)
 *   [8]  lineHeight       = fontMetrics.getHeight()    (signed byte)
 * </pre>
 * Immediately following the header (starting at {@code fontData[writePos]}): raw greyscale
 * alpha bytes, one per pixel in the ink bbox, row-major from top-left.
 *
 * <h2>Static fields</h2>
 * Several static fields on this class ({@code e}, {@code c}, {@code a}, {@code d}, {@code f})
 * are obfuscator-injected statics that do not belong to FontBuilder's own logic — they appear
 * to be artefacts of the single-file-per-class obfuscation pass sharing string pools and RSA
 * constants across classes. They are retained here as-is to match the compiled class exactly.
 *
 * <p>Package: {@code client.ui} (obf: default package, class {@code s})
 */
public final class FontBuilder {

    // -------------------------------------------------------------------------
    // Obfuscator-injected statics (not used by rasterizeGlyph; kept for fidelity)
    // -------------------------------------------------------------------------

    /**
     * Per-method profiling counter (obfuscation artifact). obf: {@code s.b}
     * Incremented once at the top of every method call; never read by game logic.
     * Kept as dead field to match the compiled class layout.
     */
    // obf: b  (dead profiling counter — do NOT confuse with class 'b' = Packet)
    public static int callCount;

    /**
     * XOR-encoded string pool entry. obf: {@code s.e}
     * Decoded value: {@code "Enter number of items to offer and press enter"}
     * This string is not used anywhere in FontBuilder's own code; it is an
     * obfuscator-injected initialiser that piggy-backs on this class's static block.
     * obf: {@code e}
     */
    // obf: e  (decoded: "Enter number of items to offer and press enter")
    public static String[] injectedStrings = new String[]{
        "Enter number of items to offer and press enter"   // z(z(".;..")) decoded
    };

    /**
     * Secondary string pool (unused by FontBuilder). obf: {@code s.f}
     */
    // obf: f
    public static String[] injectedStrings2;

    /**
     * RSA public exponent injected by the obfuscator. obf: {@code s.c}
     * Value: {@code BigInteger("10001", 16)} = 65537 (standard RSA e).
     * Used elsewhere in the codebase (e.g. {@link client.net.BitBuffer}), not by FontBuilder.
     */
    // obf: c
    public static BigInteger rsaPublicExponent = new BigInteger("10001", 16);

    /**
     * DataStore reference (obfuscator-injected). obf: {@code s.a} (field, not method)
     * Shadowed by the static method {@code a()} below; the field is always null.
     */
    // obf: a (field)  — NOTE: same letter as the rasterizeGlyph method; field is null always
    public static client.data.DataStore dataStore = null;   // obf type: nb

    /**
     * Integer scratch (obfuscator-injected). obf: {@code s.d}
     */
    // obf: d
    public static int scratch = 0;

    // -------------------------------------------------------------------------
    // Private string pool used by the ErrorHandler signature builder
    // -------------------------------------------------------------------------

    /**
     * Strings used to build the error-handler method signature string in
     * the catch block. Decoded values:
     *   z[0] = "s.a("    — method signature prefix
     *   z[1] = "{...}"   — non-null object placeholder
     *   z[2] = "null"    — null object placeholder
     * obf: {@code s.z}
     */
    // obf: z
    private static final String[] SIG_STRINGS = new String[]{
        "s.a(",    // z(z("{#+"))   decoded
        "{...}",   // z(z("{L-")) decoded
        "null"     // z(z(" o")) decoded
    };

    // -------------------------------------------------------------------------
    // Core method
    // -------------------------------------------------------------------------

    /**
     * Rasterizes one glyph of an AWT {@link Font} into the engine's packed font buffer.
     *
     * <p>Called by {@link client.ui.Panel#buildFont} for each printable character
     * (charSlot 0..94) in the 95-char RSC charset:
     * {@code ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!"#...}.
     *
     * <p>On success the method:
     * <ul>
     *   <li>Writes the 9-byte glyph header to
     *       {@code GameFrame.fontData[charSlot * 9 .. charSlot * 9 + 8]}.</li>
     *   <li>Writes the raw pixel bytes starting at
     *       {@code GameFrame.fontData[Packet.writePos]}, advancing
     *       {@code Packet.writePos} by {@code glyphWidth * glyphHeight}.</li>
     *   <li>Sets {@code SurfaceImageProducer.hasAntiAlias[fontId] = true} if any
     *       pixel is in the anti-aliasing grey range (exclusive: 30 &lt; alpha &lt; 230).</li>
     * </ul>
     *
     * @param fontId     index into {@code SurfaceImageProducer.hasAntiAlias[]} that marks
     *                   whether this font has anti-aliased glyphs  (obf: param0 / {@code n2})
     * @param font       the AWT {@link Font} to rasterize from                (obf: param1)
     * @param charSlot   glyph slot index (0..94); written to {@code fontData[charSlot*9]}
     *                                                                          (obf: param2 / {@code n3})
     * @param unusedGuard junk anti-tamper parameter (always -95 at call sites); ignored
     *                                                                          (obf: param3 / dropped)
     * @param applet     the {@link client.shell.GameShell} used for off-screen image creation
     *                                                                          (obf: param4, type {@code e})
     * @param ch         the character to rasterize                             (obf: param5)
     * @param metrics    {@link FontMetrics} for {@code font} obtained from the applet
     *                                                                          (obf: param6)
     * @param italic     {@code true} to render a fake-italic variant: the glyph is drawn
     *                   twice (at x=0 and x=1), and certain wide chars get +1 advance width
     *                                                                          (obf: param7)
     * @return {@code true} on success; {@code false} if image creation or pixel-grabbing failed
     * obf: {@code s.a(ILjava/awt/Font;IILe;CLjava/awt/FontMetrics;Z)Z}
     */
    // obf: a
    public static final boolean rasterizeGlyph(
            int fontId,
            Font font,
            int charSlot,
            int unusedGuard,        // anti-tamper param, always -95; dropped from logic
            client.shell.GameShell applet,  // obf type: e
            char ch,
            FontMetrics metrics,
            boolean italic) {

        // -- obfuscation stripped: ++callCount (profiling counter) --
        // -- obfuscation stripped: boolean bl = client.vh; (opaque predicate, always false) --

        // ---- 1. Measure the character ----

        // Raw advance width from the font (not affected by italic padding).
        // obf: n18 / n12 (two slots; n12 may be bumped below)
        int advanceWidth = metrics.charWidth(ch);   // obf: n18
        int bitmapWidth  = advanceWidth;            // obf: n12  (may be bumped for italic wide chars)

        // ---- 2. Italic adjustments ----
        // In italic mode certain wide chars (f, t, v, w, k, x, y, A, V, W) receive +1 bitmap
        // width so the second shifted draw doesn't clip.  The '/' character disables italic
        // for itself (the slash already looks slanted without the double-draw trick).
        if (italic) {
            // Wide chars that need an extra pixel when rendered italic.
            // Tested via both direct and bitwise-NOT comparisons to foil simple pattern matching.
            if (   ch == 'f'                  // 102
                || ch == 't'                  // 116
                || 'w' == ch                  // 119
                || ~ch == ~'v'                // 118  (~c == 0xffffff89 obf; ~118==-119)
                || ch == 'k'                  // 107
                || ~ch == ~'x'               // 120  (~c == 0xffffff87 obf; ~120==-121)
                || 'y' == ch                  // 121
                || ch == 'A'                  // 65
                || ~ch == ~'V'               // 86   (~c == 0xffffffa9)
                || ~ch == ~'W')              // 87   (~c == 0xffffffa8)
            {
                bitmapWidth++;  // extra column so double-draw doesn't clip right edge
            }

            // '/' (0x2f) is already visually slanted — disable italic for this char.
            if (~ch == ~'/')   // ~'/' == ~47 == -48  (obf: ~c == 0xffffffd0)
            {
                italic = false;
            }
        }

        // ---- 3. Font metrics ----

        // obf: n17
        int maxAscent   = metrics.getMaxAscent();
        // Total pixel height of the off-screen image = ascent + descent.
        // getMaxAscent() - (-getMaxDescent()) = ascent + descent.
        // obf: n16
        int totalHeight = metrics.getMaxAscent() + metrics.getMaxDescent();
        // obf: n15
        int lineHeight  = metrics.getHeight();

        // ---- 4. Render glyph into an off-screen image ----

        // Use the GameShell (which extends Applet) to create a compatible off-screen image.
        Image glyphImage = applet.createImage(bitmapWidth, totalHeight);  // obf: image
        if (glyphImage == null) {
            return false;   // peer not yet ready (headless / too early in init)
        }

        Graphics g = glyphImage.getGraphics();

        // Fill background black so non-ink pixels read as 0x000000.
        g.setColor(Color.black);
        g.fillRect(0, 0, bitmapWidth, totalHeight);

        // Draw the character white (alpha will be read from the blue channel of 0xFFFFFF).
        g.setColor(Color.white);
        g.setFont(font);
        g.drawString(String.valueOf(ch), 0, maxAscent);

        if (italic) {
            // Fake italic: draw again shifted one pixel to the right.
            // This produces a crude slant effect at the cost of one extra column.
            g.drawString(String.valueOf(ch), 1, maxAscent);
        }

        // ---- 5. Grab the raw pixels ----

        // Row-major pixel array: pixel[col + row * bitmapWidth]
        int[] pixels = new int[totalHeight * bitmapWidth];   // obf: nArray
        PixelGrabber grabber = new PixelGrabber(
                glyphImage, 0, 0, bitmapWidth, totalHeight,
                pixels, 0, bitmapWidth);
        try {
            grabber.grabPixels();
        } catch (InterruptedException e) {
            return false;
        }

        glyphImage.flush();
        glyphImage = null;   // release native image resources immediately

        // ---- 6. Find tight ink bounding box ----

        // Default bounds: full image (no trimming if entirely blank).
        // obf: n11=leftInk, n14=topInk, n10=rightInkExcl, n9=bottomInkExcl
        int leftInk   = 0;             // obf: n11
        int topInk    = 0;             // obf: n14
        int rightInk  = bitmapWidth;   // obf: n10  (exclusive)
        int bottomInk = totalHeight;   // obf: n9   (exclusive)

        // -- 6a. Find topmost row containing at least one non-black pixel --
        // Scans rows top-to-bottom; stops and records n14 at first ink row found.
        topScan:
        for (int row = 0; row < totalHeight; row++) {
            for (int col = 0; col < bitmapWidth; col++) {
                // Mask to RGB (ignore alpha channel in grabbed ARGB pixel).
                if ((pixels[col + row * bitmapWidth] & 0xFFFFFF) != 0) {
                    topInk = row;
                    break topScan;
                }
            }
        }

        // -- 6b. Find leftmost column containing at least one non-black pixel --
        leftScan:
        for (int col = 0; col < bitmapWidth; col++) {
            for (int row = 0; row < totalHeight; row++) {
                if ((pixels[col + row * bitmapWidth] & 0xFFFFFF) != 0) {
                    leftInk = col;
                    break leftScan;
                }
            }
        }

        // -- 6c. Find bottommost row containing a non-black pixel (scan upward) --
        // obf: the loop scans from row = totalHeight-1 downward
        // Anti-tamper stub at this point (param3 >= -86 recursive call) is stripped.
        bottomScan:
        for (int row = totalHeight - 1; row >= 0; row--) {
            for (int col = 0; col < bitmapWidth; col++) {
                if ((pixels[col + row * bitmapWidth] & 0xFFFFFF) != 0) {
                    bottomInk = row + 1;   // exclusive upper bound
                    break bottomScan;
                }
            }
        }

        // -- 6d. Find rightmost column containing a non-black pixel (scan leftward) --
        rightScan:
        for (int col = bitmapWidth - 1; col >= 0; col--) {
            for (int row = 0; row < totalHeight; row++) {
                if ((pixels[col + row * bitmapWidth] & 0xFFFFFF) != 0) {
                    rightInk = col + 1;   // exclusive upper bound
                    break rightScan;
                }
            }
        }

        // ---- 7. Write the 9-byte glyph header ----
        //
        // Header slot base: charSlot * 9  into GameFrame.fontData[]  (obf: qb.k)
        // Bitmap offset = current write cursor into fontData          (obf: b.c = Packet.writePos)
        //
        // The 21-bit bitmap offset is packed into three signed bytes using a 7-bit-per-byte
        // scheme (matching the Surface.drawCharacter() decode:
        //   offset = fontData[slot]*16384 + fontData[slot+1]*128 + fontData[slot+2]).
        //
        // ib.a(x, 127) = x & 127  (StreamBase.bitwiseAnd — strips bit 7 to keep value signed-safe).

        int bitmapOffset = client.net.Packet.writePos;   // obf: b.c — current pixel write position

        // [0] High byte: floor(offset / 16384) — bits 20..14 as signed byte
        client.shell.GameFrame.fontData[0 + 9 * charSlot] =
                (byte)(bitmapOffset / 16384);
        // [1] Mid byte: (offset / 128) & 0x7F — bits 13..7
        client.shell.GameFrame.fontData[1 + 9 * charSlot] =
                (byte)((bitmapOffset / 128) & 0x7F);   // obf: ib.a(b.c/128, 127)
        // [2] Low byte: offset & 0x7F — bits 6..0
        client.shell.GameFrame.fontData[2 + 9 * charSlot] =
                (byte)(bitmapOffset & 0x7F);            // obf: ib.a(b.c, 127)
        // [3] Glyph pixel width (ink bbox, exclusive right minus exclusive left)
        client.shell.GameFrame.fontData[3 + 9 * charSlot] =
                (byte)(rightInk - leftInk);
        // [4] Glyph pixel height (ink bbox)
        client.shell.GameFrame.fontData[4 + 9 * charSlot] =
                (byte)(bottomInk - topInk);
        // [5] X bearing: how many blank columns on the left before ink starts
        client.shell.GameFrame.fontData[5 + 9 * charSlot] =
                (byte)leftInk;
        // [6] Y bearing: distance from top of ink to the baseline (ascent - top trim)
        client.shell.GameFrame.fontData[6 + 9 * charSlot] =
                (byte)(maxAscent - topInk);
        // [7] Advance width: full horizontal advance (original charWidth, before italic padding)
        client.shell.GameFrame.fontData[7 + 9 * charSlot] =
                (byte)advanceWidth;
        // [8] Line height: total font line spacing
        client.shell.GameFrame.fontData[8 + 9 * charSlot] =
                (byte)lineHeight;

        // ---- 8. Emit per-pixel alpha bytes into the font data buffer ----
        //
        // Scan only the tight ink bounding box (topInk..bottomInk, leftInk..rightInk).
        // Each pixel contributes its low 8 bits (blue channel of the white-on-black render,
        // which equals the greyscale alpha/anti-aliasing value).
        //
        // If any pixel falls in the anti-aliasing grey range (30 < pixel < 230), mark this
        // font as having anti-aliased glyphs.

        for (int row = topInk; row < bottomInk; row++) {
            for (int col = leftInk; col < rightInk; col++) {
                // Extract low 8 bits (blue / greyscale alpha).  The image was drawn white
                // on black, so this is 0 (empty) or 255 (solid) or a grey anti-alias value.
                int alpha = 0xFF & pixels[bitmapWidth * row + col];  // obf: n6

                // Anti-aliasing detection: grey pixels (not fully off, not fully solid)
                // indicate the AWT renderer applied sub-pixel blending.
                if (alpha > 30 && alpha < 230) {   // obf: n29=30 < n30=alpha && ~alpha > -231
                    // Mark this font ID as having anti-aliased glyphs.
                    // SurfaceImageProducer.hasAntiAlias[fontId] = true
                    client.scene.SurfaceImageProducer.hasAntiAlias[fontId] = true;  // obf: fb.k[n2]
                }

                // Write the pixel byte and advance the write cursor.
                client.shell.GameFrame.fontData[client.net.Packet.writePos++] =
                        (byte)alpha;   // obf: qb.k[b.c++] = (byte)n6
            }
        }

        return true;

        // -- obfuscation stripped: RuntimeException catch wrapping i.a(e, "s.a("+args+")") --
    }

    // -------------------------------------------------------------------------
    // XOR string-pool helpers (present in every obfuscated class)
    // -------------------------------------------------------------------------

    /**
     * First-pass XOR decoder: converts a literal {@link String} to a {@code char[]} for the
     * second-pass decoder.  For strings shorter than 2 chars the single character is XOR'd
     * with {@code 'n'} (0x6E); longer strings are returned as-is.
     * obf: {@code z(Ljava/lang/String;)[C}
     */
    // obf: z(String)
    private static char[] z(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char)(chars[0] ^ 'n');   // single-char XOR key = 0x6E
        }
        return chars;
    }

    /**
     * Second-pass XOR decoder: decodes a {@code char[]} produced by {@link #z(String)}
     * using a rotating 5-byte key table {@code [107, 85, 98, 3, 110]} (k=107, U=85, b=98, ETX=3, n=110).
     * Returns the decoded string as an interned {@link String}.
     * obf: {@code z([C)Ljava/lang/String;}
     */
    // obf: z(char[])
    private static String z(char[] chars) {
        // Rotating key: indices 0..4 → XOR keys 107, 85, 98, 3, 110
        final int[] KEY = {107, 85, 98, 3, 110};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char)(chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
