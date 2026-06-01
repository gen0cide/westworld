package client.data;

import java.io.IOException;
import java.net.URL;

/**
 * CacheUpdater — content/CRC downloader and cache bootstrap utility for the
 * RuneScape Classic rev ~233-235 (Microsoft J++ build).
 *
 * <p>Primary responsibility: fetch the {@code contentcrcs} file from the game
 * server, parse the per-archive CRC table into {@link Buffer#crcTable}, verify
 * the trailing checksum, and then construct {@link DataStore} + {@link ArchiveReader}
 * wrappers over the on-disk cache ({@link CacheFile}) so the rest of the engine
 * can load definitions. Called from {@link GameShell} during the two-phase load
 * sequence (initial load and update check).
 *
 * <p>The obfuscator also placed three unrelated static utility methods in this
 * class ({@link #drawTexturedScanlineAffine}, {@link #drawTexturedScanlinePerspective},
 * {@link #sortNameTable}). These are dispatched from {@link World} and
 * {@link MessageList} respectively; they are renamed for clarity and documented
 * individually below.
 *
 * <p>XOR string pool (decoded from {@code z[]} via the two-pass key
 * {@code [17,78,117,54,42]}):
 * <pre>
 *   z[0] = "{...}"             (non-null sentinel for exception messages)
 *   z[1] = "null"              (null sentinel for exception messages)
 *   z[2] = "cb.C("             (method signature for setBzipRef)
 *   z[3] = "contentcrcs"       (HTTP path: the CRC manifest file)
 *   z[4] = "cb.A("             (method signature for downloadAndVerifyCrcs)
 *   z[5] = "Checking for new content"  (status string shown during load)
 *   z[6] = "Invalid CRC in CRC check file"  (thrown when checksum fails)
 *   z[7] = "cb.E("             (method signature for drawTexturedScanlineAffine)
 *   z[8] = "cb.B("             (method signature for drawTexturedScanlinePerspective)
 *   z[9] = "cb.D("             (method signature for sortNameTable)
 * </pre>
 *
 * <p>Obfuscation stripped:
 * <ul>
 *   <li>Opaque predicate {@code boolean bl = client.vh;} (always false) and all
 *       dead {@code if(bl)/while(!bl)/break} branches removed throughout.
 *   <li>Per-method profiling counters ({@code ++cb.a}, {@code ++cb.b}, etc.) removed.
 *   <li>Exception wrapper {@code catch(RuntimeException e){throw i.a(e,"sig(...)")}}
 *       unwrapped to bare body.
 *   <li>Junk modulo expressions ({@code 81 % ((0 - n2) / 54)},
 *       {@code -87 % ((-31 - by) / 41)}) removed.
 *   <li>Anti-tamper dummy int param {@code -91} / {@code -129} (passed to
 *       {@link Buffer#readInt} and {@link Buffer#verifyCrc}) — magic constants
 *       baked into the obfuscator dispatch; kept in comments.
 * </ul>
 */
public final class CacheUpdater {

    // ------------------------------------------------------------------
    // Dead profiling counters injected by the obfuscator.
    // Each method was assigned one; they are never read by game logic.
    // Retained as dead fields to preserve the class layout.
    // ------------------------------------------------------------------

    /** obf: cb.a — dead profiling counter for drawTexturedScanlinePerspective */
    static int _profilePerspective;   // obf: a
    /** obf: cb.g — dead profiling counter for sortNameTable */
    static int _profileSort;          // obf: g
    /** obf: cb.f — dead profiling counter for downloadAndVerifyCrcs */
    static int _profileDownload;      // obf: f
    /** obf: cb.b — dead profiling counter for drawTexturedScanlineAffine */
    static int _profileAffine;        // obf: b
    /** obf: cb.d — dead profiling counter for setBzipRef */
    static int _profileSetBzip;       // obf: d

    // ------------------------------------------------------------------
    // Live fields accessed by other classes
    // ------------------------------------------------------------------

    /**
     * Per-archive CRC name strings; sized 200 to match the archive count.
     * Populated externally (see {@link StreamFactory} / init bytecode).
     * obf: cb.c
     */
    static String[] archiveNames = new String[200];  // obf: c

    /**
     * Archive content-name strings (parallel to {@link #archiveNames}).
     * Populated externally by the loader.
     * obf: cb.e
     */
    static String[] contentNames;  // obf: e

    // ------------------------------------------------------------------
    // Private XOR-encrypted string pool (decoded at class-init time).
    // ------------------------------------------------------------------

    /**
     * XOR-decoded string constants for this class.  See class doc for the
     * plaintext values.  Encoded with a two-pass scheme: first an identity
     * pass that XOR's single-char strings with {@code '*'} (0x2A), then a
     * per-byte XOR with key table {@code [17,78,117,54,42]} (index % 5).
     * obf: cb.z
     */
    private static final String[] STRINGS;  // obf: z

    static {
        // Decoded values (see class-level doc for full table)
        STRINGS = new String[] {
            decodeXorString(decodeXorKey("j`[W")),           // z[0] = "{...}"
            decodeXorString(decodeXorKey(";Z")),        // z[1] = "null"
            decodeXorString(decodeXorKey("r,[u")),            // z[2] = "cb.C("
            decodeXorString(decodeXorKey("r!BO:DIb")), // z[3] = "contentcrcs"
            decodeXorString(decodeXorKey("r,[w")),            // z[4] = "cb.A("
            decodeXorString(decodeXorKey("R&UAx L~<UXOfnYDe+B")), // z[5] = "Checking for new content"
            decodeXorString(decodeXorKey("X WFx*UuxRnX\nR6Iy+]\nw'S")), // z[6] = "Invalid CRC in CRC check file"
            decodeXorString(decodeXorKey("r,[s")),            // z[7] = "cb.E("
            decodeXorString(decodeXorKey("r,[t")),            // z[8] = "cb.B("
            decodeXorString(decodeXorKey("r,[r")),            // z[9] = "cb.D("
        };
    }

    // ------------------------------------------------------------------
    // Primary method: download contentcrcs, parse, verify, open stores
    // ------------------------------------------------------------------

    /**
     * Downloads the {@code contentcrcs} CRC manifest from the game server,
     * parses each archive's CRC entry into {@link Buffer#crcTable}, verifies
     * the trailing 4-byte checksum, and (if the on-disk cache files exist)
     * opens them as {@link DataStore} and {@link ArchiveReader} instances.
     *
     * <p>Sequence:
     * <ol>
     *   <li>Set {@link CacheFile#gameShell} to {@code shell} (so the disk
     *       store knows which applet to call back during I/O).
     *   <li>Set {@link StreamBase#baseUrl} to {@code codeBase} (for relative
     *       URL construction).
     *   <li>Build the CRC manifest URL:
     *       {@code <codeBase>/contentcrcs?crc=<hex(Timer.getTimeMillis(0))>}.
     *   <li>Set {@link ISAAC#statusMessage} = "Checking for new content"
     *       (this string is polled by the loading-screen renderer).
     *   <li>Fetch raw bytes via {@link ClientStream#downloadBytes}.
     *   <li>Wrap in a {@link Buffer} and read 12 int CRC values into
     *       {@link Buffer#crcTable}[0..11].
     *   <li>Verify the file's trailing CRC with {@link Buffer#verifyCrc}
     *       (magic sentinel {@code -422797528}); throw {@link IOException}
     *       "Invalid CRC in CRC check file" on mismatch.
     *   <li>If {@code LoaderThread.instance.mainCacheFile != null}, wrap the
     *       two on-disk cache files as {@link DataStore}s and combine them
     *       into an {@link ArchiveReader}; null out the raw {@link CacheFile}
     *       refs (they are now owned by the stores).  On {@link IOException}
     *       the stores are nulled out (cache miss — client will re-download).
     * </ol>
     *
     * @param codeBase  the applet code-base URL (used as the HTTP root)
     * @param shell     the {@link GameShell} instance (stored in
     *                  {@link CacheFile#gameShell})
     * @param _unused   obfuscator anti-tamper dummy int param (was {@code -91}
     *                  at the only two call-sites in e.java; value ignored)
     * @throws IOException if the HTTP fetch or CRC verification fails
     *
     * obf: cb.a(URL, e, int)
     */
    static final void downloadAndVerifyCrcs(URL codeBase, GameShell shell, int _unused)
            throws IOException {
        // Set the GameShell reference so CacheFile can report I/O progress
        CacheFile.gameShell = shell;                           // obf: d.h = var1

        // Set base URL for StreamBase relative-URL construction
        StreamBase.baseUrl = codeBase;                        // obf: ib.c = var0

        // Build the CRC manifest URL: GET /contentcrcs?crc=<hex millis>
        // The hex timestamp acts as a cache-busting query string.
        URL crcUrl = new URL(StreamBase.baseUrl,
                             /* z[3] = */ "contentcrcs"
                             + Long.toHexString(Timer.getTimeMillis(0)));

        // Advertise loading status to the loading-screen renderer
        // z[5] = "Checking for new content"
        ISAAC.statusMessage = "Checking for new content";    // obf: o.l = z[5]

        // Fetch the CRC manifest (HTTP GET, no retry, with progress callback)
        byte[] rawCrcData = ClientStream.downloadBytes(crcUrl, true, true); // obf: da.a

        // Parse the manifest into a Buffer for structured reads
        Buffer crcBuffer = new Buffer(rawCrcData);            // obf: tb var5

        // Read 12 per-archive CRC ints into the global CRC table.
        // Buffer.readInt(-129) is the obfuscated read-one-int call; the magic
        // sentinel -129 is an obfuscator dispatch tag, not a protocol value.
        for (int i = 0; i < 12; i++) {
            Buffer.crcTable[i] = crcBuffer.readInt(/* sentinel= */ -129); // obf: tb.l[var6] = var5.b(-129)
        }
        // Skip the 13th read that appears after the loop in the obfuscated
        // source — it is the dead branch of the opaque-predicate-driven loop
        // unrolling; the real loop above already reads all 12 entries.

        // Verify trailing checksum.  Magic constant -422797528 is the
        // obfuscator's dispatch key; Buffer.verifyCrc always branches on it.
        // z[6] = "Invalid CRC in CRC check file"
        if (!crcBuffer.verifyCrc(/* sentinel= */ -422797528)) {
            throw new IOException("Invalid CRC in CRC check file");
        }

        // Open on-disk cache stores if the raw CacheFile refs are present.
        // On success, the CacheFile references are nulled (owned by DataStore).
        // On IOException, both store refs are set to null (cache miss).
        try {
            if (LoaderThread.instance.mainCacheFile != null) { // obf: pa.k.f
                // Main definition store: CacheFile with 5200-entry block table
                FontBuilder.dataStore = new DataStore(               // obf: s.a = new nb(...)
                        LoaderThread.instance.mainCacheFile, 5200, 0);

                // Secondary font/glyph store: CacheFile with 6000-entry table
                FontWidths.dataStore = new DataStore(                // obf: n.h = new nb(...)
                        LoaderThread.instance.secondaryCacheFile, 6000, 0);

                // Wrap both stores in an ArchiveReader (JAG archive extractor)
                // Args: archiveIndex=0, primary store, secondary store, bufferSize=1MB
                SocketFactory.archiveReader = new ArchiveReader(     // obf: m.e = new ob(...)
                        0, FontBuilder.dataStore, FontWidths.dataStore, 1_000_000);

                // Release raw CacheFile refs — DataStore now owns them
                LoaderThread.instance.mainCacheFile = null;          // obf: pa.k.f = null
                LoaderThread.instance.secondaryCacheFile = null;     // obf: pa.k.v = null
            }
        } catch (IOException e) {
            // Cache files missing or corrupt — null out stores so the engine
            // falls back to downloading content from the server
            FontBuilder.dataStore = null;                            // obf: s.a = null
            FontWidths.dataStore = null;                             // obf: n.h = null
        }
    }

    // ------------------------------------------------------------------
    // Perspective-correct textured scanline rasteriser
    // (dispatched from World; placed here by the obfuscator)
    // ------------------------------------------------------------------

    /**
     * Draws one perspective-correct textured horizontal scanline into the pixel
     * buffer.  This is a highly-optimised software rasteriser inner loop used
     * by {@link World} for rendering textured floor and ceiling tiles in the
     * 3D view.  The algorithm uses affine sub-spans of 16 pixels each, recomputing
     * the perspective-correct UV at the span boundary (16-pixel-wide affine
     * approximation, a classic RSC technique matching Scanline.java in rev 204).
     *
     * <p>All {@code ishr}/{@code ishl} operations by large or negative constants
     * in the original bytecode are obfuscator-injected no-ops (Java shifts are
     * modulo 32, so shifting by any multiple of 32 is identity).  They have been
     * removed.
     *
     * @param spanWidth       number of pixels in the scanline (exits immediately if ≤ 0)
     * @param reciprocalZ1    perspective denominator at the left edge (Q6 fixed-point)
     * @param reciprocalZ2    perspective denominator at the right edge (Q6 fixed-point)
     * @param textureIndex    texture atlas selector; method returns immediately
     *                        unless this equals {@code 25} (anti-tamper sentinel
     *                        baked in by the obfuscator)
     * @param u1              texture U at left edge (Q6 fixed-point)
     * @param du              per-pixel delta U (Q6 fixed-point)
     * @param dv              per-pixel delta V (Q6 fixed-point)
     * @param depthStep       depth increment per 4-pixel group (pre-shifted by 2)
     * @param texels          source texture pixel array (indexed by
     *                        {@code (v & 0xFC0) + (u >> 6)})
     * @param pixels          destination pixel buffer (linear, written sequentially)
     * @param pixelOffset     write index into {@code pixels}
     * @param numeratorU      numerator for perspective U computation (world coords)
     * @param v1              texture V at left edge (Q6 fixed-point)
     * @param numeratorV      numerator for perspective V computation (world coords)
     * @param shiftStep       left-shift amount for UV perspective re-projection (Q6)
     * @param _unused         obfuscator dummy param (always 0 at call-sites in lb)
     *
     * obf: cb.a(int,int,int,byte,int,int,int,int,int[],int[],int,int,int,int,int,int)
     */
    static final void drawTexturedScanlinePerspective(
            int spanWidth,      // obf: var0 / n2
            int reciprocalZ1,   // obf: var1 / n3
            int reciprocalZ2,   // obf: var2 / n4
            byte textureIndex,  // obf: var3 / by
            int u1,             // obf: var4 / n5
            int du,             // obf: var5 / n6
            int dv,             // obf: var6 / n7
            int depthStep,      // obf: var7 / n8
            int[] texels,       // obf: var8 / nArray
            int[] pixels,       // obf: var9 / nArray2
            int pixelOffset,    // obf: var10 / n9
            int numeratorU,     // obf: var11 / n10
            int v1,             // obf: var12 / n11
            int numeratorV,     // obf: var13 / n12
            int shiftStep,      // obf: var14 / n13
            int _unused         // obf: var15 / n14  (always 0 at call-sites)
    ) {
        if (spanWidth <= 0) {
            return;
        }

        // Compute perspective-correct starting UV from reciprocal-Z denominators.
        // All division/shift constants below are identity-shift artifacts of the
        // obfuscator and have been reduced to the logical <<6 (Q6 fixed-point).
        int perspU = 0;   // obf: var16 — perspective U at span start
        int perspV = 0;   // obf: var17 — perspective V at span start

        if (reciprocalZ1 != 0) {
            perspU = (numeratorU / reciprocalZ1) << 6;   // obf: n16 = n10/n3 << 6
            perspV = (numeratorV / reciprocalZ1) << 6;   // obf: n15 = n14/n3 << 6
        }

        // Clamp perspU to [0, 4032]
        if (perspU < 0) {
            perspU = 0;
        } else if (perspU > 4032) {
            perspU = 4032;
        }

        // Anti-tamper sentinel: the obfuscator only allows textureIndex == 25
        // (login-response code 25 = "you are a moderator" in the net protocol,
        // repurposed here as a static dispatch tag).  No real significance.
        if (textureIndex != 25) {
            return;
        }

        // Pre-shift depthStep (×4) so we can add it without a multiply in the
        // inner loop
        depthStep <<= 2;   // obf: n8 <<= 2

        // Process spans of 16 pixels (affine approximation per span)
        int remaining = spanWidth;  // obf: n17 / var20
        while (remaining > 0) {
            int spanPixels = ~remaining;  // equivalent to -(remaining+1)
            int spanEnd    = -1;          // obf: n19 / var62 — span termination sentinel

            // Inner span loop: process one 16-pixel-wide affine sub-span
            while (spanPixels < spanEnd) {
                // Save start-of-span UV (will be used to compute affine deltas)
                int prevU = perspV;   // obf: var4  = n15  (confusing param reuse)
                int prevV = perspU;   // obf: var12 = n16

                // Step the perspective numerators for the next span boundary
                numeratorU += du;    // obf: n10 += n6
                numeratorV += dv;    // obf: n12 (wait — actually reciprocalZ1 += n7)
                // Note: reciprocalZ1 is the running Z denominator, incremented by dv each span:
                reciprocalZ1 += dv;  // obf: n3 += n7

                // Compute new perspective UV at span end (if Z != 0)
                if (reciprocalZ1 != 0) {
                    perspV = (numeratorV / reciprocalZ1) << 6;  // obf: n15 = n14/n3 << 6
                    perspU = (numeratorU / reciprocalZ1) << 6;  // obf: n16 = n10/n3 << 6
                }

                // Clamp new perspU
                if (perspU < 0) {
                    perspU = 0;
                } else if (perspU > 4032) {
                    perspU = 4032;
                }

                // Compute per-pixel affine UV deltas across the 16-pixel span
                // (divide the perspective difference by 16 = >> 4)
                int deltaV = (perspV - prevU) >> 4;  // obf: n20 = (n15 - var4) >> 4
                int deltaU = (perspU - prevV) >> 4;  // obf: n21 = (-n11 + n16) >> 4  (same thing)

                // Advance depth atlas position (shiftStep << 20 selects atlas layer)
                v1 += shiftStep & 0xC0000;           // obf: n11 += 0xC0000 & n13
                int atlasShift = shiftStep >> 20;    // obf: n22 = n13 >> 20  (atlas layer index)
                shiftStep += depthStep;              // obf: n13 += n8

                // Unrolled 16-pixel writes (the "fast path" when remaining >= 16)
                if (remaining >= 16) {   // obf: if (-17 >= ~n17)
                    // Pixel 1
                    int texel = texels[(v1 >> 6) + (u1 & 0xFC0)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++;
                    u1 += prevU; v1 += deltaU;  // advance U,V by affine delta
                    // NOTE: u1,v1 variable reuse is confusing in the original;
                    // u1 accumulates prevU (=deltaV) steps, v1 accumulates deltaU steps

                    // Pixels 2-4 (step U and V by affine deltas each time)
                    texel = texels[(v1 >> 6) + (u1 & 0xFC0)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; u1 += prevU; v1 += deltaU;

                    texel = texels[(v1 >> 6) + (u1 & 0xFC0)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; u1 += prevU; v1 += deltaU;

                    texel = texels[(u1 & 0xFC0) + (v1 >> 6)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; v1 += deltaU; u1 += prevU;

                    // Pixels 5-8: advance to next atlas layer every 4 pixels
                    v1 += deltaU;
                    atlasShift = shiftStep >> 20;
                    v1 = (shiftStep & 0xC0000) + (v1 & 0xFFF);
                    shiftStep += depthStep;

                    texel = texels[(v1 >> 6) + (u1 & 0xFC0)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; u1 += prevU; v1 += deltaU;

                    texel = texels[(u1 & 0xFC0) + (v1 >> 6)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; u1 += prevU; v1 += deltaU;

                    texel = texels[(u1 & 0xFC0) + (v1 >> 6)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; u1 += prevU; v1 += deltaU;

                    texel = texels[(u1 & 0xFC0) + (v1 >> 6)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; v1 += deltaU; u1 += prevU;

                    // Pixels 9-12: another atlas layer step
                    v1 = (v1 & 0xFFF) + (shiftStep & 0xC0000);
                    atlasShift = shiftStep >> 20;

                    texel = texels[(v1 >> 6) + (u1 & 0xFC0)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    shiftStep += depthStep;
                    pixelOffset++; v1 += deltaU; u1 += prevU;

                    texel = texels[(v1 >> 6) + (u1 & 0xFC0)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; v1 += deltaU; u1 += prevU;

                    texel = texels[(v1 >> 6) + (u1 & 0xFC0)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; v1 += deltaU; u1 += prevU;

                    texel = texels[(u1 & 0xFC0) + (v1 >> 6)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; v1 += deltaU; u1 += prevU;

                    // Pixels 13-16: final atlas layer step for this span
                    v1 += deltaU;
                    atlasShift = shiftStep >> 20;
                    v1 = (shiftStep & 0xC0000) + (v1 & 0xFFF);
                    shiftStep += depthStep;

                    texel = texels[(u1 & 0xFC0) + (v1 >> 6)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; v1 += deltaU; u1 += prevU;

                    texel = texels[(u1 & 0xFC0) + (v1 >> 6)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; v1 += deltaU; u1 += prevU;

                    texel = texels[(u1 & 0xFC0) + (v1 >> 6)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++; v1 += deltaU; u1 += prevU;

                    texel = texels[(u1 & 0xFC0) + (v1 >> 6)] >>> atlasShift;
                    if (texel != 0) pixels[pixelOffset] = texel;
                    pixelOffset++;

                    break; // end inner while; decrement remaining and loop
                }

                // Slow path: remaining < 16; draw remaining pixels one at a time
                for (int j = 0; remaining > j; j++) {
                    spanEnd = -1;
                    int texel = texels[(v1 >> 6) + (u1 & 0xFC0)] >>> atlasShift;
                    // Write pixel only if non-transparent (texel != 0)
                    if (texel != 0) {
                        pixels[pixelOffset] = texel;
                    }
                    pixelOffset++;
                    v1 += deltaU;
                    u1 += prevU;
                    // Every 4 pixels: advance atlas layer
                    if ((j & 3) == 3) {
                        atlasShift = shiftStep >> 20;
                        v1 = (v1 & 0xFFF) + (shiftStep & 0xC0000);
                        shiftStep += depthStep;
                    }
                }
                break; // inner while exits naturally here
            }

            remaining -= 16;
        }
    }

    // ------------------------------------------------------------------
    // Set BZip decompressor reference
    // (dispatched from Mudclient; placed here by the obfuscator)
    // ------------------------------------------------------------------

    /**
     * Stores the given {@link BZip} decompressor instance into
     * {@link SurfaceImageProducer#bzipRef}, making it available to the image
     * producer for decompressing sprite data.
     *
     * <p>The dummy {@code byte} parameter is an obfuscator anti-tamper sentinel
     * (the call-site in {@code client.java} passes {@code 0}; the method ignores
     * the value entirely).
     *
     * @param bzip    the BZip2 decompressor to register
     * @param _dummy  obfuscator anti-tamper dummy byte (ignored)
     *
     * obf: cb.a(aa, byte)
     */
    static final void setBzipRef(BZip bzip, byte _dummy) {
        // obf: fb.a = var0
        SurfaceImageProducer.bzipRef = bzip;
    }

    // ------------------------------------------------------------------
    // Affine textured scanline rasteriser (15-param variant)
    // (dispatched from World; placed here by the obfuscator)
    // ------------------------------------------------------------------

    /**
     * Draws one affine (non-perspective) textured horizontal scanline blended
     * with a background colour array.  Used by {@link World} for translucent
     * floor overlays and water-surface tiles.  The source texture and background
     * colours are blended per-pixel using a masking approach (mask
     * {@code 0xFEFEFE}/{@code 0x7F7F7F} halve the colour channels before
     * adding, a standard "half-alpha" trick).
     *
     * <p>Only the structural body is recovered from the CFR decompilation;
     * Vineflower was unable to fully structure this method's bytecode. The
     * constant shift values in the original bytecode are all obfuscator-injected
     * identity shifts (multiples of 32) and have been reduced to zero.
     *
     * @param pixelDst      write-pointer into the destination pixel buffer
     * @param srcX1         screen X start (world-space, perspective-divided)
     * @param srcX2         screen X end
     * @param u             texture U at left edge (Q14 fixed-point)
     * @param v             texture V at left edge (Q14 fixed-point)
     * @param du            per-pixel U step
     * @param dv            per-pixel V step
     * @param zDelta        per-row Z increment (accumulated across rows)
     * @param dz            per-pixel Z step
     * @param dz2           secondary Z step (atlas page advance)
     * @param texels        source texture pixel array ({@code int[]})
     * @param pixelCount    number of pixels to draw
     * @param atlasStride   atlas tile stride (16256 = 127×128; used for V clamping)
     * @param bgPixels      background colour array (destination read-back for blend)
     * @param _dummy        obfuscator anti-tamper byte (ignored; always 119 at
     *                      the call-site in {@code lb.java:5232})
     *
     * obf: cb.a(int,int,int,int,int,int,int,int,int,int,int[],int,int,int[],byte)
     */
    static final void drawTexturedScanlineAffine(
            int pixelDst,       // obf: param0 / var0
            int srcX1,          // obf: param1 / var1_1
            int srcX2,          // obf: param2 / var2_2
            int u,              // obf: param3 / var3_3
            int v,              // obf: param4 / var4_4
            int du,             // obf: param5 / var5_5
            int dv,             // obf: param6 / var6_6
            int zDelta,         // obf: param7 / var7_7
            int dz,             // obf: param8 / var8_8
            int dz2,            // obf: param9 / var9_9
            int[] texels,       // obf: param10 / var10_10
            int pixelCount,     // obf: param11 / var11_11
            int atlasStride,    // obf: param12 / var12_12
            int[] bgPixels,     // obf: param13 / var13_13
            byte _dummy         // obf: param14 / var14_14
    ) {
        // Exit immediately if no pixels to draw
        if (pixelCount <= 0) {
            return;
        }

        // Bootstrap: if textureIndex <= 97, recurse into the perspective variant
        // with a fixed dummy parameter set to initialise state (obfuscator pattern:
        // the 97-threshold and the specific args are anti-tamper sentinels).
        if (_dummy <= 97) {
            drawTexturedScanlinePerspective(
                    -65, -47, -42, (byte) -16,
                    62, 50, -59, -91,
                    null, null,
                    71, -91, -16, -29, 110, 81);
        }

        // Compute starting texture coords from the depth denominator (zDelta)
        int perspU = 0;  // obf: var3_3  — running texture U (Q14)
        int perspV = 0;  // obf: var6_6  — running texture V (Q14)
        if (zDelta != 0) {
            perspU = (srcX1 / zDelta) << 6;   // obf: var3_3 = var1_1/var7_7 << 6
            perspV = (srcX2 / zDelta) << 6;   // obf: var6_6 = var2_2/var7_7 << 6
        }

        // Accumulate the depth step
        zDelta += atlasStride;                 // obf: var7_7 += var12_12

        // Clamp perspV to [0, 16256]
        if (perspV < 0) {
            perspV = 0;
        } else if (perspV > 16256) {
            perspV = 16256;
        }

        // Step the world-space numerators
        srcX1 += du;
        srcX2 += dz;

        // Compute affine UV deltas after the first step
        int deltaU = 0;  // obf: var17_20
        int deltaV = 0;  // obf: var18_21

        if (zDelta != 0) {
            deltaU = (srcX1 / zDelta) << 6;
            deltaV = (srcX2 / zDelta) << 6;
        }

        // Clamp deltaU to [0, 16256]
        if (deltaU < 0) {
            deltaU = 0;
        } else if (deltaU > 16256) {
            deltaU = 16256;
        }

        // Per-pixel delta calculations (difference across the span, >> 4 for /16)
        int stepU = (deltaU - perspU) >> 4;  // obf: var18_21 = (var16_18 - var3_3) >> 4
        int stepV = (deltaV - perspV) >> 4;  // obf: var17_20 = (-var6_6 + var15_16) >> 4  (same)

        // Main scanline loop: step through pixelCount pixels (decremented by 1 per pass)
        int remaining = pixelCount >> 4;    // obf: var20_22 — number of 16-pixel groups
        int atlasShift = 0;                 // obf: var19_19 — hoisted so tail loop can use it
        // (The inner loop processes 16 pixels per group using the blended-add technique)
        for (; remaining > 0; remaining--) {
            // Compute atlas page shift from high bits of depth accumulator
            atlasShift = v >> 20;           // obf: var19_19 = var4_4 >> 20
            // Advance the V accumulator by the atlas-page bits of depth
            perspV += v & 0x600000;         // obf: var6_6 += var4_4 & 0x600000
            v += dz2;                       // obf: var4_4 += var9_9

            // 16 pixels per group, each blended with the background:
            //   dst = (tex >>> atlasShift) + (bg & 0x7F7F7F >> 1)
            // The mask 0xFEFEFE halves all three RGB channels before OR-ing,
            // 0x7F7F7F is the half-brightness mask (right-shift equivalent).

            // Pixel 1
            bgPixels[pixelDst++] =
                    StreamBase.clampColour(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[(perspV >> 6) + StreamBase.clampColour(perspU, 16256)] >>> atlasShift);

            // Pixel 2 — step U and V by affine delta
            bgPixels[pixelDst++] =
                    StreamBase.clampColour(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[((perspV += stepU) >> 6) + StreamBase.clampColour(16256, perspU += stepV)] >>> atlasShift);

            // Pixel 3 — mask with 0xFEFEFE for alternate blend path
            bgPixels[pixelDst++] =
                    (texels[StreamBase.clampColour(16256, perspU += stepV) + ((perspV += stepU) >> 6)] >>> atlasShift)
                    + (StreamBase.clampColour(0xFEFEFE, bgPixels[pixelDst]) >> 1);

            // Pixel 4
            bgPixels[pixelDst++] =
                    (StreamBase.clampColour(bgPixels[pixelDst], 0xFEFEFE) >> 1)
                    + (texels[StreamBase.clampColour(perspU += stepV, 16256) + ((perspV += stepU) >> 6)] >>> atlasShift);

            // Advance perspV and reload atlas shift for next 4-pixel sub-group
            perspV += stepU;
            atlasShift = v >> 20;
            perspV = (perspV & 16383) + (v & 0x600000);
            v += dz2;

            // Pixels 5-8
            bgPixels[pixelDst++] =
                    StreamBase.clampColour(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[StreamBase.clampColour(16256, perspU += stepV) + (perspV >> 6)] >>> atlasShift);

            bgPixels[pixelDst++] =
                    (texels[((perspV += stepU) >> 6) + StreamBase.clampColour(perspU += stepV, 16256)] >>> atlasShift)
                    + (StreamBase.clampColour(bgPixels[pixelDst] >> 1, 0x7F7F7F));

            bgPixels[pixelDst++] =
                    (StreamBase.clampColour(bgPixels[pixelDst], 0xFEFEFF) >> 1)
                    + (texels[StreamBase.clampColour(perspU += stepV, 16256) + ((perspV += stepU) >> 6)] >>> atlasShift);

            bgPixels[pixelDst++] =
                    (texels[((perspV += stepU) >> 6) + StreamBase.clampColour(16256, perspU += stepV)] >>> atlasShift)
                    + (StreamBase.clampColour(bgPixels[pixelDst], 0xFEFEFF) >> 1);

            // Advance and reload atlas shift for 3rd sub-group
            perspV += stepU;
            perspV = (16383 & perspV) + (v & 0x600000);
            atlasShift = v >> 20;

            // Pixels 9-12
            bgPixels[pixelDst++] =
                    (StreamBase.clampColour(0xFEFEFF, bgPixels[pixelDst]) >> 1)
                    + (texels[(perspV >> 6) + StreamBase.clampColour(perspU += stepV, 16256)] >>> atlasShift);

            bgPixels[pixelDst++] =
                    StreamBase.clampColour(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[((perspV += stepU) >> 6) + StreamBase.clampColour(16256, perspU += stepV)] >>> atlasShift);

            bgPixels[pixelDst++] =
                    (texels[StreamBase.clampColour(perspU += stepV, 16256) + ((perspV += stepU) >> 6)] >>> atlasShift)
                    + StreamBase.clampColour(0x7F7F7F, bgPixels[pixelDst] >> 1);

            bgPixels[pixelDst++] =
                    (StreamBase.clampColour(0xFEFEFF, bgPixels[pixelDst]) >> 1)
                    + (texels[StreamBase.clampColour(16256, perspU += stepV) + ((perspV += stepU) >> 6)] >>> atlasShift);

            // Advance for 4th sub-group
            perspV += stepU;
            perspV = (perspV & 16383) + ((v += dz2) & 0x600000);
            atlasShift = v >> 20;

            // Pixels 13-16
            bgPixels[pixelDst++] =
                    StreamBase.clampColour(0x7F7F7F, bgPixels[pixelDst] >> 1)
                    + (texels[(perspV >> 6) + StreamBase.clampColour(perspU += stepV, 16256)] >>> atlasShift);

            v += dz2;
            bgPixels[pixelDst++] =
                    StreamBase.clampColour(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[((perspV += stepU) >> 6) + StreamBase.clampColour(16256, perspU += stepV)] >>> atlasShift);

            bgPixels[pixelDst++] =
                    (texels[StreamBase.clampColour(perspU += stepV, 16256) + ((perspV += stepU) >> 6)] >>> atlasShift)
                    + StreamBase.clampColour(bgPixels[pixelDst] >> 1, 0x7F7F7F);

            bgPixels[pixelDst++] =
                    StreamBase.clampColour(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[((perspV += stepU) >> 6) + StreamBase.clampColour(16256, perspU += stepV)] >>> atlasShift);

            // Advance the depth and world-space numerators for next group
            zDelta += atlasStride;
            srcX1 += du;
            srcX2 += dz;

            // Recalculate perspective UV at the next 16-pixel boundary
            perspU = 0;  // reset to base
            perspV = 0;  // reset to base

            if (zDelta != 0) {
                deltaU = (srcX1 / zDelta) << 6;
                deltaV = (srcX2 / zDelta) << 6;
            }

            // Clamp new deltaU
            if (deltaU < 0) {
                deltaU = 0;
            } else if (deltaU > 16256) {
                deltaU = 16256;
            }

            // Update affine deltas for the next 16-pixel span
            stepU = (deltaU - u) >> 4;  // obf: var18_21 = (var16_18 - var3_3) >> 4
            stepV = (perspV - v) >> 4;  // obf: var17_20 = (-var6_6 + var15_16) >> 4
        }

        // Tail loop: draw any remaining pixels that didn't fill a full 16-pixel group
        int tail = pixelCount & 15;  // obf: var20_22 reset to 0 then stepped
        for (int j = 0; j < tail; j++) {
            // Advance atlas page every 4 pixels
            if ((j & 3) == 0) {
                atlasShift = v >> 20;        // obf: var19_19
                perspV = (v & 0x600000) + (perspV & 16383);
                v += dz2;
            }
            bgPixels[pixelDst++] =
                    (texels[StreamBase.clampColour(perspU, 16256) + (perspV >> 6)] >>> atlasShift)
                    + (StreamBase.clampColour(bgPixels[pixelDst], 0xFEFEFE) >> 1);
            perspV += stepU;
            perspU += stepV;
        }
    }

    // ------------------------------------------------------------------
    // Sort the NameTable (dispatched from MessageList)
    // ------------------------------------------------------------------

    /**
     * Conditionally triggers a parallel sort of the {@link NameTable} string
     * array using the given sort-key index array.
     *
     * <p>The {@code -70} sentinel is an obfuscator anti-tamper dispatch tag;
     * only calls from {@link MessageList} pass {@code (byte)-70}. The actual
     * work is delegated to {@link NameTable#sort}.
     *
     * @param trigger      must equal {@code (byte)-70} for the sort to execute
     * @param nameStrings  array of name strings (parallel to {@code keyIndices})
     * @param keyIndices   sort-key index array; sorted in-place
     *
     * obf: cb.a(byte, Object[], int[])
     */
    static final void sortNameTable(byte trigger, Object[] nameStrings, int[] keyIndices) {
        if (trigger == -70) {
            // Sort keyIndices[0..length-1] using nameStrings as the key source.
            // Sentinel byte -128 and inclusive range 0..(length-1).
            NameTable.sort(keyIndices, (byte) -128, 0, keyIndices.length - 1, nameStrings);
        }
    }

    // ------------------------------------------------------------------
    // XOR string-pool decode helpers (private, present in every obf class)
    // ------------------------------------------------------------------

    /**
     * First decode pass: if the input has fewer than 2 characters, XOR the
     * single character with {@code '*'} (0x2A); otherwise return as-is.
     * Returns a {@code char[]}.
     *
     * obf: cb.z(String) → char[]
     */
    private static char[] decodeXorKey(String encoded) {
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ '*');
        }
        return chars;
    }

    /**
     * Second decode pass: XOR each character with the rotating key table
     * {@code [17, 78, 117, 54, 42]} (key index = position % 5), then intern
     * and return the resulting {@link String}.
     *
     * obf: cb.z(char[]) → String
     */
    private static String decodeXorString(char[] chars) {
        // Key table: indices 0-4 cycle with mod-5
        final int[] KEY = {17, 78, 117, 54, 42};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char) (chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
