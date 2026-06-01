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
 * <p>The obfuscator also placed unrelated static utility methods in this class
 * ({@link #drawTexturedScanlineAffine}, {@link #drawTexturedScanlinePerspective},
 * {@link #setBzipRef}, {@link #sortNameTable}). Verified dispatch sources (from the
 * clean decompilation): the two scanline rasterisers are called from {@link Scene}
 * ({@code lb}), {@link #setBzipRef} from {@code Mudclient} ({@code client}), and
 * {@link #sortNameTable} from {@link MessageList} ({@code wb}). They are renamed for
 * clarity and documented individually below.
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
     *   <li>If {@code ImageLoader.imageWidthCarrier.dataFile != null} (obf
     *       {@code pa.k.f}), wrap the
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
        // obf: while loop runs while (-13 < ~var6) ⇔ var6 < 12, i.e. indices 0..11.
        for (int i = 0; i < 12; i++) {
            Buffer.crcTable[i] = crcBuffer.readInt(/* sentinel= */ -129); // obf: tb.l[var6] = var5.b(-129)
        }
        // 13th read: after the loop the obfuscated source executes ONE more
        // `var5.b(-129)` (when var6 reaches 12 the loop guard fails and control
        // falls through to this read before `break`).  Its *result* is discarded,
        // but the read is NOT dead: it advances the Buffer cursor (Buffer.w) by 4
        // bytes past the 12 entries, positioning it at the trailing CRC.  verifyCrc
        // below does `w -= 4`, CRCs the preceding bytes, and compares against the
        // int it re-reads — so dropping this read would CRC the wrong byte range.
        crcBuffer.readInt(/* sentinel= */ -129);  // obf: var5.b(-129) (post-loop)

        // Verify trailing checksum.  Magic constant -422797528 is the
        // obfuscator's dispatch key; Buffer.verifyCrc always branches on it.
        // z[6] = "Invalid CRC in CRC check file"
        if (!crcBuffer.verifyCrc(/* sentinel= */ -422797528)) {
            throw new IOException("Invalid CRC in CRC check file");
        }

        // Open on-disk cache stores if the raw CacheFile refs are present.
        // On success, the CacheFile references are nulled (owned by DataStore).
        // On IOException, both store refs are set to null (cache miss).
        //
        // NOTE: the obfuscated source reads the LoaderThread instance from the
        // static field pa.k = ImageLoader.imageWidthCarrier (type c = LoaderThread),
        // NOT from any "LoaderThread.instance" self-reference.  Its CacheFile fields
        // are pa.k.f = imageWidthCarrier.dataFile and pa.k.v = imageWidthCarrier.indexFile255.
        try {
            if (ImageLoader.imageWidthCarrier.dataFile != null) { // obf: pa.k.f
                // Main definition store: CacheFile with 5200-entry block table
                FontBuilder.dataStore = new DataStore(               // obf: s.a = new nb(...)
                        ImageLoader.imageWidthCarrier.dataFile, 5200, 0);

                // Secondary font/glyph store: CacheFile with 6000-entry table
                FontWidths.dataStore = new DataStore(                // obf: n.h = new nb(...)
                        ImageLoader.imageWidthCarrier.indexFile255, 6000, 0);

                // Wrap both stores in an ArchiveReader (JAG archive extractor)
                // Args: archiveIndex=0, primary store, secondary store, bufferSize=1MB
                SocketFactory.archiveReader = new ArchiveReader(     // obf: m.e = new ob(...)
                        0, FontBuilder.dataStore, FontWidths.dataStore, 1_000_000);

                // Release raw CacheFile refs — DataStore now owns them
                ImageLoader.imageWidthCarrier.dataFile = null;       // obf: pa.k.f = null
                ImageLoader.imageWidthCarrier.indexFile255 = null;   // obf: pa.k.v = null
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
    // (dispatched from Scene; placed here by the obfuscator)
    // ------------------------------------------------------------------

    /**
     * Draws one perspective-correct textured horizontal scanline into the pixel
     * buffer.  This is a highly-optimised software rasteriser inner loop used
     * by {@link Scene} for rendering textured floor and ceiling tiles in the
     * 3D view.  The algorithm uses affine sub-spans of 16 pixels each, recomputing
     * the perspective-correct UV at the span boundary (16-pixel-wide affine
     * approximation, a classic RSC technique matching Scanline.java in rev 204).
     *
     * <p>Body reconstructed faithfully from the clean decompilation
     * ({@code normalized-clean/cb.java}). The meaningful fixed-point shifts are
     * {@code <<6}/{@code >>6} for the Q6 texture coords (mask {@code 0xFC0}) and
     * {@code >>20}/{@code &0xC0000} for the atlas-page select. The earlier deob
     * mis-mapped the running U/V accumulators onto the wrong parameters, so its
     * per-pixel advances were wrong; that has been corrected.
     *
     * @param spanWidth       number of pixels in the scanline (exits immediately if ≤ 0)
     * @param reciprocalZ      running perspective denominator; stepped by {@code dz}
     *                         each 16-pixel span (obf var1)
     * @param scratch          obfuscator-declared param reused purely as a per-pixel
     *                         texel scratch (obf var2; original param value ignored)
     * @param textureIndex     method returns immediately unless this equals {@code 25}
     *                         (anti-tamper sentinel baked in by the obfuscator)
     * @param textureU         running texture U accumulator (obf var4; re-seeded from
     *                         the perspective U at each span boundary)
     * @param numeratorUStep   per-span increment added to {@code numeratorU} (obf var5)
     * @param dz               per-span increment added to {@code reciprocalZ} (obf var6)
     * @param depthStep        atlas-page depth increment; pre-shifted by 2 (obf var7)
     * @param texels           source texture pixel array (indexed by
     *                         {@code (textureV >> 6) + (textureU & 0xFC0)})
     * @param pixels           destination pixel buffer (linear, written sequentially)
     * @param pixelOffset      write index into {@code pixels} (obf var10)
     * @param numeratorU       numerator for perspective U (divided by reciprocalZ) (obf var11)
     * @param textureV         running texture V accumulator (obf var12; re-seeded from
     *                         the perspective V at each span boundary)
     * @param numeratorVStep   per-span increment added to {@code numeratorV} (obf var13)
     * @param atlasAccum       atlas-page accumulator; {@code >>20} gives the shift,
     *                         {@code &0xC0000} the page bits (obf var14)
     * @param numeratorV       numerator for perspective V (divided by reciprocalZ) (obf var15)
     *
     * obf: cb.a(int,int,int,byte,int,int,int,int,int[],int[],int,int,int,int,int,int)
     */
    static final void drawTexturedScanlinePerspective(
            int spanWidth,       // obf: var0
            int reciprocalZ,     // obf: var1  (running Z denom; += dz per span)
            int scratch,         // obf: var2  (reused as per-pixel texel scratch)
            byte textureIndex,   // obf: var3  (==25 gate)
            int textureU,        // obf: var4  (running U accumulator)
            int numeratorUStep,  // obf: var5  (numeratorU += this per span)
            int dz,              // obf: var6  (reciprocalZ += this per span)
            int depthStep,       // obf: var7  (<<2; atlasAccum += this)
            int[] texels,        // obf: var8
            int[] pixels,        // obf: var9
            int pixelOffset,     // obf: var10
            int numeratorU,      // obf: var11
            int textureV,        // obf: var12 (running V accumulator)
            int numeratorVStep,  // obf: var13 (numeratorV += this per span)
            int atlasAccum,      // obf: var14 (>>20 shift, &0xC0000 page)
            int numeratorV       // obf: var15
    ) {
        if (spanWidth <= 0) {
            return;
        }

        // Compute perspective-correct starting UV from the reciprocal-Z denominator.
        int perspU = 0;   // obf: var16 — perspective U at span start
        int perspV = 0;   // obf: var17 — perspective V at span start

        if (reciprocalZ != 0) {
            perspU = (numeratorU / reciprocalZ) << 6;   // obf: var16 = var11/var1 << 6
            perspV = (numeratorV / reciprocalZ) << 6;   // obf: var17 = var15/var1 << 6
        }

        // Pre-shift depthStep (×4) so the inner loop can add it without a multiply.
        depthStep <<= 2;   // obf: var7 <<= 2

        // Clamp perspU to [0, 4032]
        if (perspU < 0) {
            perspU = 0;
        } else if (perspU > 4032) {
            perspU = 4032;
        }

        // Anti-tamper sentinel: the obfuscator only allows textureIndex == 25.
        if (textureIndex != 25) {
            return;
        }

        // Process spans of 16 pixels (affine approximation per span).
        // obf: var20 = var0; while loop drains 16 pixels at a time.
        int remaining = spanWidth;  // obf: var20
        while (true) {
            // Inner condition: ~remaining < -1  ⇔  remaining > 0.  When remaining
            // drops to <= 0 the obf method returns out of the whole routine.
            if (~remaining >= -1) {
                return;
            }

            // Re-seed the U/V accumulators from the previous span's perspective UV.
            textureU = perspV;   // obf: var4  = var17
            textureV = perspU;   // obf: var12 = var16

            // Step the perspective numerators / denominator for the next boundary.
            numeratorU  += numeratorUStep;  // obf: var11 += var5
            reciprocalZ += dz;              // obf: var1  += var6
            numeratorV  += numeratorVStep;  // obf: var15 += var13

            // Compute new perspective UV at span end (if Z != 0).
            if (reciprocalZ != 0) {
                perspV = (numeratorV / reciprocalZ) << 6;  // obf: var17 = var15/var1 << 6
                perspU = (numeratorU / reciprocalZ) << 6;  // obf: var16 = var11/var1 << 6
            }

            // Clamp new perspU to [0, 4032]
            // obf: if (~var16 > -1) var16 = 0; else if (~var16 < -4033) var16 = 4032;
            if (perspU < 0) {
                perspU = 0;
            } else if (perspU > 4032) {
                perspU = 4032;
            }

            // Per-pixel affine UV deltas across the 16-pixel span (difference / 16).
            int deltaU = (perspV - textureU) >> 4;  // obf: var19 = (var17 - var4) >> 4
            int deltaV = (perspU - textureV) >> 4;  // obf: var18 = (var16 - var12) >> 4

            // Advance atlas-page position.
            textureV += 0xC0000 & atlasAccum;        // obf: var12 += 786432 & var14
            int atlasShift = atlasAccum >> 20;       // obf: var21 = var14 >> 20
            atlasAccum += depthStep;                 // obf: var14 += var7

            // Fast path: a full 16-pixel span.  obf: if (-17 >= ~var20) ⇔ remaining >= 16.
            if (remaining >= 16) {
                // Pixels 1-4
                scratch = texels[(textureV >> 6) + (textureU & 0xFC0)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                pixelOffset++; textureU += deltaU; textureV += deltaV;

                scratch = texels[(textureV >> 6) + (textureU & 0xFC0)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                textureU += deltaU; pixelOffset++; textureV += deltaV;

                scratch = texels[(textureV >> 6) + (textureU & 0xFC0)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                textureU += deltaU; pixelOffset++; textureV += deltaV;

                scratch = texels[(textureU & 0xFC0) + (textureV >> 6)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                pixelOffset++; textureV += deltaV; textureU += deltaU;

                // Pixels 5-8: advance atlas page
                atlasShift = atlasAccum >> 20;
                textureV = (0xC0000 & atlasAccum) + (0xFFF & textureV);
                atlasAccum += depthStep;
                scratch = texels[(textureV >> 6) + (textureU & 0xFC0)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                pixelOffset++; textureV += deltaV; textureU += deltaU;

                scratch = texels[(textureU & 0xFC0) + (textureV >> 6)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                pixelOffset++; textureV += deltaV; textureU += deltaU;

                scratch = texels[(textureU & 0xFC0) + (textureV >> 6)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                pixelOffset++; textureV += deltaV; textureU += deltaU;

                scratch = texels[(textureU & 0xFC0) + (textureV >> 6)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                textureU += deltaU; pixelOffset++; textureV += deltaV;

                // Pixels 9-12: advance atlas page
                textureV = (textureV & 0xFFF) + (atlasAccum & 0xC0000);
                atlasShift = atlasAccum >> 20;
                scratch = texels[(textureV >> 6) + (textureU & 0xFC0)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                atlasAccum += depthStep;
                pixelOffset++; textureV += deltaV; textureU += deltaU;

                scratch = texels[(textureV >> 6) + (textureU & 0xFC0)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                textureU += deltaU; textureV += deltaV; pixelOffset++;

                scratch = texels[(textureV >> 6) + (textureU & 0xFC0)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                textureV += deltaV; textureU += deltaU; pixelOffset++;

                scratch = texels[(textureU & 0xFC0) + (textureV >> 6)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                pixelOffset++; textureV += deltaV; textureU += deltaU;

                // Pixels 13-16: advance atlas page
                atlasShift = atlasAccum >> 20;
                textureV = (atlasAccum & 0xC0000) + (textureV & 0xFFF);
                atlasAccum += depthStep;
                scratch = texels[(textureU & 0xFC0) + (textureV >> 6)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                textureU += deltaU; textureV += deltaV; pixelOffset++;

                scratch = texels[(textureU & 0xFC0) + (textureV >> 6)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                textureV += deltaV; pixelOffset++; textureU += deltaU;

                scratch = texels[(textureU & 0xFC0) + (textureV >> 6)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                pixelOffset++; textureV += deltaV; textureU += deltaU;

                scratch = texels[(0xFC0 & textureU) + (textureV >> 6)] >>> atlasShift;
                if (scratch != 0) pixels[pixelOffset] = scratch;
                pixelOffset++;
            } else {
                // Slow path: remaining < 16; draw the leftover pixels one at a time.
                for (int j = 0; remaining > j; j++) {
                    scratch = texels[(textureV >> 6) + (textureU & 0xFC0)] >>> atlasShift;
                    if (scratch != 0) {
                        pixels[pixelOffset] = scratch;
                    }
                    pixelOffset++;
                    textureV += deltaV;
                    textureU += deltaU;
                    // Every 4 pixels: advance atlas page
                    if ((j & 3) == 3) {
                        atlasShift = atlasAccum >> 20;
                        textureV = (0xFFF & textureV) + (atlasAccum & 0xC0000);
                        atlasAccum += depthStep;
                    }
                }
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
     * {@link SurfaceImageProducer#bzip}, making it available to the image
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
        SurfaceImageProducer.bzip = bzip;
    }

    // ------------------------------------------------------------------
    // Affine textured scanline rasteriser (15-param variant)
    // (dispatched from Scene; placed here by the obfuscator)
    // ------------------------------------------------------------------

    /**
     * Draws one affine (non-perspective) textured horizontal scanline blended
     * with a background colour array.  Used by {@link Scene} for translucent
     * floor overlays and water-surface tiles.  The source texture and background
     * colours are blended per-pixel using a masking approach (mask
     * {@code 0xFEFEFE}/{@code 0x7F7F7F} halve the colour channels before
     * adding, a standard "half-alpha" trick).
     *
     * <p>This body was reconstructed faithfully from the clean Vineflower
     * decompilation ({@code normalized-clean/cb.java}). The earlier deob (built
     * against a defective decompile) mis-stated the fixed-point scale: the real
     * code uses {@code <<7}/{@code >>7} for the texture coords (Q7, atlas stride
     * 16256) and {@code >>23} for the atlas-page shift — NOT {@code <<6}/{@code >>6}/
     * {@code >>20}. Those have been corrected here. {@code ib.a(x,y)} is a plain
     * bitwise-AND ({@link StreamBase#bitwiseAnd}), used both to mask texture
     * coordinates and to halve background colour channels for the blend.
     *
     * @param pixelDst         write index into {@code bgPixels} (obf var0)
     * @param numeratorU       numerator for U (divided by reciprocalZ); += uStep per group (obf var1)
     * @param numeratorV       numerator for V (divided by reciprocalZ); += vStep per group (obf var2)
     * @param textureU         running texture U accumulator; re-seeded each group (obf var3)
     * @param atlasAccum       atlas-page accumulator; {@code >>23} shift, {@code &0x600000}
     *                         page bits, {@code += atlasStep} (obf var4)
     * @param uStep            per-group increment added to {@code numeratorU} (obf var5)
     * @param textureV         running texture V accumulator; re-seeded each group (obf var6)
     * @param reciprocalZ      running perspective denominator; += zStep per group (obf var7)
     * @param vStep            per-group increment added to {@code numeratorV} (obf var8)
     * @param atlasStep        per-quad increment added to {@code atlasAccum} (obf var9)
     * @param texels           source texture pixel array (indexed by
     *                         {@code (textureV >> 7) + (textureU & 16256)}) (obf var10)
     * @param pixelCount       number of pixels to draw (obf var11)
     * @param zStep            per-group increment added to {@code reciprocalZ} (obf var12)
     * @param bgPixels         destination buffer; read back for the half-brightness blend (obf var13)
     * @param _dummy           obfuscator anti-tamper byte; if {@code <= 97} a junk
     *                         self-call is emitted (state ignored) (obf var14)
     *
     * obf: cb.a(int,int,int,int,int,int,int,int,int,int,int[],int,int,int[],byte)
     */
    static final void drawTexturedScanlineAffine(
            int pixelDst,       // obf: var0
            int numeratorU,     // obf: var1  (/ reciprocalZ → textureU)
            int numeratorV,     // obf: var2  (/ reciprocalZ → textureV)
            int textureU,       // obf: var3  (running U accumulator)
            int atlasAccum,     // obf: var4  (>>23 shift, &0x600000 page, += atlasStep)
            int uStep,          // obf: var5  (numeratorU += this)
            int textureV,       // obf: var6  (running V accumulator)
            int reciprocalZ,    // obf: var7  (running Z denom; += zStep)
            int vStep,          // obf: var8  (numeratorV += this)
            int atlasStep,      // obf: var9  (atlasAccum += this)
            int[] texels,       // obf: var10
            int pixelCount,     // obf: var11
            int zStep,          // obf: var12 (reciprocalZ += this)
            int[] bgPixels,     // obf: var13 (destination + blend source)
            byte _dummy         // obf: var14 (<=97 junk-call gate)
    ) {
        // Exit immediately if no pixels to draw
        if (pixelCount <= 0) {
            return;
        }

        int perspU = 0;  // obf: var16 — perspective U at span boundary (numeratorU / Z)
        int perspV = 0;  // obf: var15 — perspective V at span boundary (numeratorV / Z)

        // Obfuscator junk self-call: when the dummy byte is <= 97 the build emits a
        // recursive call into the perspective variant with sentinel args.  It never
        // affects this method's state (all-null texture arrays, perspV<=0 path).
        if (_dummy <= 97) {
            drawTexturedScanlinePerspective(
                    -65, -47, -42, (byte) -16,
                    62, 50, -59, -91,
                    null, null,
                    71, -91, -16, -29, 110, 81);
        }

        // Compute starting texture coords from the perspective denominator.
        if (reciprocalZ != 0) {
            textureU = (numeratorU / reciprocalZ) << 7;   // obf: var3 = var1/var7 << 7
            textureV = (numeratorV / reciprocalZ) << 7;   // obf: var6 = var2/var7 << 7
        }

        // Accumulate the depth step.
        reciprocalZ += zStep;                 // obf: var7 += var12

        // Clamp textureV to [0, 16256]
        // obf: if (~var6 > -1) var6 = 0; else if (~var6 < -16257) var6 = 16256;
        if (textureV < 0) {
            textureV = 0;
        } else if (textureV > 16256) {
            textureV = 16256;
        }

        // Step the perspective numerators.
        numeratorU += uStep;   // obf: var1 += var5
        numeratorV += vStep;   // obf: var2 += var8

        // Compute the perspective UV at the end of the first span.
        if (reciprocalZ != 0) {
            perspV = (numeratorV / reciprocalZ) << 7;  // obf: var15 = var2/var7 << 7
            perspU = (numeratorU / reciprocalZ) << 7;  // obf: var16 = var1/var7 << 7
        }

        // Clamp perspV to [0, 16256]
        // obf: if (~var15 > -1) var15 = 0; else if (~var15 < -16257) var15 = 16256;
        if (perspV < 0) {
            perspV = 0;
        } else if (perspV > 16256) {
            perspV = 16256;
        }

        // Per-pixel affine UV deltas across the 16-pixel span (difference / 16).
        int stepV = (perspV - textureV) >> 4;  // obf: var17 = (var15 - var6) >> 4
        int stepU = (perspU - textureU) >> 4;  // obf: var18 = (var16 - var3) >> 4

        // Number of full 16-pixel groups.  obf: var20 = var11 >> 4.
        int groups = pixelCount >> 4;          // obf: var20
        int atlasShift = 0;                    // obf: var19 — atlas page shift (hoisted for tail)

        // Main loop: while (~var20 < -1) ⇔ groups > 0.
        while (true) {
            if (~groups >= -1) {
                break;
            }

            // 16 pixels per group; each blends a texel with the (half-brightness)
            // background.  ib.a(x,y) is a bitwise AND (StreamBase.bitwiseAnd):
            //   - (bg & 0xFEFEFE) >> 1 / (bg >> 1 & 0x7F7F7F) = half-brightness bg
            //   - texels[ (textureV >> 7) + (textureU & 16256) ] >>> atlasShift = source texel
            // 0x600000 (6291456) selects the atlas page; >>23 yields the page shift.

            // Pixels 1-4
            atlasShift = atlasAccum >> 23;                     // obf: var23 = var4 >> 23
            textureV += atlasAccum & 0x600000;                 // obf: var6 += var4 & 6291456
            atlasAccum += atlasStep;                           // obf: var4 += var9
            bgPixels[pixelDst++] = StreamBase.bitwiseAnd(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[(textureV >> 7) + StreamBase.bitwiseAnd(textureU, 16256)] >>> atlasShift);
            textureU += stepU;
            textureV += stepV;
            bgPixels[pixelDst++] = StreamBase.bitwiseAnd(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[(textureV >> 7) + StreamBase.bitwiseAnd(16256, textureU)] >>> atlasShift);
            textureV += stepV;
            textureU += stepU;
            bgPixels[pixelDst++] = (texels[StreamBase.bitwiseAnd(16256, textureU) + (textureV >> 7)] >>> atlasShift)
                    + (StreamBase.bitwiseAnd(0xFEFEFE, bgPixels[pixelDst]) >> 1);
            textureU += stepU;
            textureV += stepV;
            bgPixels[pixelDst++] = (StreamBase.bitwiseAnd(bgPixels[pixelDst], 0xFEFEFE) >> 1)
                    + (texels[StreamBase.bitwiseAnd(textureU, 16256) + (textureV >> 7)] >>> atlasShift);
            textureU += stepU;
            textureV += stepV;

            // Pixels 5-8 (advance atlas page)
            atlasShift = atlasAccum >> 23;                     // obf: var24 = var4 >> 23
            textureV = (textureV & 16383) + (atlasAccum & 0x600000);
            atlasAccum += atlasStep;
            bgPixels[pixelDst++] = StreamBase.bitwiseAnd(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[StreamBase.bitwiseAnd(16256, textureU) + (textureV >> 7)] >>> atlasShift);
            textureU += stepU;
            textureV += stepV;
            bgPixels[pixelDst++] = (texels[(textureV >> 7) + StreamBase.bitwiseAnd(textureU, 16256)] >>> atlasShift)
                    + StreamBase.bitwiseAnd(bgPixels[pixelDst] >> 1, 0x7F7F7F);
            textureU += stepU;
            textureV += stepV;
            bgPixels[pixelDst++] = (StreamBase.bitwiseAnd(bgPixels[pixelDst], 0xFEFEFF) >> 1)
                    + (texels[StreamBase.bitwiseAnd(textureU, 16256) + (textureV >> 7)] >>> atlasShift);
            textureV += stepV;
            textureU += stepU;
            bgPixels[pixelDst++] = (texels[(textureV >> 7) + StreamBase.bitwiseAnd(16256, textureU)] >>> atlasShift)
                    + (StreamBase.bitwiseAnd(bgPixels[pixelDst], 0xFEFEFF) >> 1);
            textureV += stepV;
            textureU += stepU;

            // Pixels 9-12 (advance atlas page)
            textureV = (16383 & textureV) + (atlasAccum & 0x600000);
            atlasShift = atlasAccum >> 23;                     // obf: var25 = var4 >> 23
            bgPixels[pixelDst++] = (StreamBase.bitwiseAnd(16711423, bgPixels[pixelDst]) >> 1)
                    + (texels[(textureV >> 7) + StreamBase.bitwiseAnd(textureU, 16256)] >>> atlasShift);
            atlasAccum += atlasStep;
            textureU += stepU;
            textureV += stepV;
            bgPixels[pixelDst++] = StreamBase.bitwiseAnd(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[(textureV >> 7) + StreamBase.bitwiseAnd(16256, textureU)] >>> atlasShift);
            textureV += stepV;
            textureU += stepU;
            bgPixels[pixelDst++] = (texels[StreamBase.bitwiseAnd(textureU, 16256) + (textureV >> 7)] >>> atlasShift)
                    + StreamBase.bitwiseAnd(8355711, bgPixels[pixelDst] >> 1);
            textureV += stepV;
            textureU += stepU;
            bgPixels[pixelDst++] = (StreamBase.bitwiseAnd(16711423, bgPixels[pixelDst]) >> 1)
                    + (texels[StreamBase.bitwiseAnd(16256, textureU) + (textureV >> 7)] >>> atlasShift);
            textureU += stepU;
            textureV += stepV;

            // Pixels 13-16 (advance atlas page)
            textureV = (textureV & 16383) + (atlasAccum & 0x600000);
            atlasShift = atlasAccum >> 23;                     // obf: var19 = var4 >> 23
            bgPixels[pixelDst++] = StreamBase.bitwiseAnd(8355711, bgPixels[pixelDst] >> 1)
                    + (texels[(textureV >> 7) + StreamBase.bitwiseAnd(textureU, 16256)] >>> atlasShift);
            atlasAccum += atlasStep;
            textureV += stepV;
            textureU += stepU;
            bgPixels[pixelDst++] = StreamBase.bitwiseAnd(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[(textureV >> 7) + StreamBase.bitwiseAnd(16256, textureU)] >>> atlasShift);
            textureV += stepV;
            textureU += stepU;
            bgPixels[pixelDst++] = (texels[StreamBase.bitwiseAnd(textureU, 16256) + (textureV >> 7)] >>> atlasShift)
                    + StreamBase.bitwiseAnd(bgPixels[pixelDst] >> 1, 0x7F7F7F);
            textureV += stepV;
            textureU += stepU;
            bgPixels[pixelDst++] = StreamBase.bitwiseAnd(bgPixels[pixelDst] >> 1, 0x7F7F7F)
                    + (texels[(textureV >> 7) + StreamBase.bitwiseAnd(16256, textureU)] >>> atlasShift);

            // Advance the depth and world-space numerators, then re-seed the running
            // accumulators from the perspective UV of the previous span boundary.
            reciprocalZ += zStep;   // obf: var7 += var12
            numeratorU  += uStep;   // obf: var1 += var5
            numeratorV  += vStep;   // obf: var2 += var8
            textureU = perspU;      // obf: var3 = var16
            textureV = perspV;      // obf: var6 = var15

            // Recompute perspective UV at the next 16-pixel boundary.
            if (reciprocalZ != 0) {
                perspU = (numeratorU / reciprocalZ) << 7;  // obf: var16 = var1/var7 << 7
                perspV = (numeratorV / reciprocalZ) << 7;  // obf: var15 = var2/var7 << 7
            }

            // Clamp perspV to [0, 16256]
            // obf: if (var15 >= 0){ if (~var15 >= -16257) break; var15 = 16256; } else var15 = 0;
            if (perspV < 0) {
                perspV = 0;
            } else if (perspV > 16256) {
                perspV = 16256;
            }

            // Update affine deltas for the next 16-pixel span.
            stepU = (perspU - textureU) >> 4;  // obf: var18 = (var16 - var3) >> 4
            stepV = (perspV - textureV) >> 4;  // obf: var17 = (var15 - var6) >> 4
            groups--;                          // obf: var20--
        }

        // Tail loop: draw the leftover (pixelCount & 15) pixels.
        // obf: groups reset to 0; loop while ~var20 > ~(var11 & 15) ⇔ var20 < (var11 & 15).
        int tail = pixelCount & 15;            // obf: ~(var11 & 15) bound
        for (int j = 0; j < tail; j++) {
            // Advance atlas page every 4 pixels.  obf: if (~(var20 & 3) == -1) ⇔ (var20 & 3) == 0.
            if ((j & 3) == 0) {
                textureV = (atlasAccum & 0x600000) + (textureV & 16383);
                atlasShift = atlasAccum >> 23;   // obf: var19 = var4 >> 23
                atlasAccum += atlasStep;
            }
            bgPixels[pixelDst++] = (texels[StreamBase.bitwiseAnd(textureU, 16256) + (textureV >> 7)] >>> atlasShift)
                    + (StreamBase.bitwiseAnd(bgPixels[pixelDst], 0xFEFEFE) >> 1);
            textureV += stepV;
            textureU += stepU;
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
     * work is delegated to {@link NameTable#sortWithKeys}.
     *
     * @param trigger      must equal {@code (byte)-70} for the sort to execute
     * @param nameStrings  array of name strings (parallel to {@code keyIndices})
     * @param keyIndices   sort-key index array; sorted in-place
     *
     * obf: cb.a(byte, Object[], int[])
     */
    static final void sortNameTable(byte trigger, Object[] nameStrings, int[] keyIndices) {
        if (trigger == -70) {
            // Sort keyIndices[0..length-1] using nameStrings as the parallel
            // value array.  Sentinel byte -128 and inclusive range 0..(length-1).
            // obf: ub.a(var2, (byte)-128, 0, var2.length - 1, var1)
            NameTable.sortWithKeys(keyIndices, (byte) -128, 0, keyIndices.length - 1, nameStrings);
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
