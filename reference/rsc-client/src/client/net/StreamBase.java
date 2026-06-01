package client.net;

import java.io.IOException;
import java.net.URL;

/**
 * StreamBase (obf: ib) — Base node class for the audio filter graph and the network buffer
 * hierarchy.  Every class that participates in the singly/doubly-linked ring lists used by
 * {@link client.util.LinkedQueue} (obf: db) and {@link client.audio.FilterChain} (obf: va)
 * extends this class.
 *
 * <p>Concrete subclasses:
 * <ul>
 *   <li>{@link client.net.Buffer}       (obf: tb) — cursor read/write + RSA encrypt</li>
 *   <li>{@link client.audio.FilterNode} (obf: bb) — abstract audio filter node</li>
 *   <li>{@link client.audio.FilterStage}(obf: hb) — abstract audio filter stage</li>
 *   <li>{@link client.audio.FilterChain}(obf: va) — abstract audio filter chain</li>
 * </ul>
 *
 * <p>The class also owns two pieces of engine-wide static state that do not fit any other class:
 * <ol>
 *   <li>The applet's <em>base URL</em> ({@link #BASE_URL}), written once by
 *       {@link client.data.CacheUpdater} (obf: cb) and then used here to construct per-resource
 *       content URLs when downloading.</li>
 *   <li>A static utility method {@link #loadResource} that fetches a numbered game resource
 *       (content file) from the server, verifying its CRC, optionally reading from an in-memory
 *       {@link client.data.ArchiveReader} (obf: ob) cache first, and decompressing the result
 *       through {@link client.world.World#unpackData} (obf: k.a).</li>
 * </ol>
 *
 * <p>Doubly-linked list protocol (used by LinkedQueue / FilterChain):
 * Each node stores a forward pointer ({@link #next}) and a backward pointer ({@link #prev}).
 * The sentinel magic value {@code -27331} passed to {@link #unlinkSelf} is an obfuscation
 * guard — the real operation is always the unlink branch.
 */
public class StreamBase {

    // -------------------------------------------------------------------------
    // Doubly-linked list node fields (used by LinkedQueue / audio filter graph)
    // -------------------------------------------------------------------------

    /** Forward (next) link in the intrusive doubly-linked ring.  obf: a */
    public StreamBase next;

    /** Backward (prev) link in the intrusive doubly-linked ring.  obf: e */
    public StreamBase prev;

    // -------------------------------------------------------------------------
    // Engine-wide static state
    // -------------------------------------------------------------------------

    /**
     * Base URL of the applet/server, set once by {@link client.data.CacheUpdater} (obf: cb)
     * before any resource loading begins.  All content file URLs are resolved relative to this.
     * obf: c
     */
    public static URL BASE_URL;

    /**
     * Unused static int array — present in the obfuscated class but never written or read by
     * any live code path.  Likely a dead profiling or padding artifact from the obfuscator.
     * obf: d
     */
    public static int[] _deadIntArray_d;

    /**
     * Profiling counter incremented at the entry of {@link #loadResource}.  Dead — never read
     * by game logic.  obf: b
     *
     * NOTE: The letter 'b' here is a static int field on class ib, NOT a reference to the
     * obfuscated class named 'b' (which maps to {@link client.net.Packet}).  See NAMING.md
     * obf-i caveat.
     */
    public static int _profilingCounter_b;

    /**
     * Profiling counter incremented at the entry of {@link #unlinkSelf}.  Dead — never read
     * by game logic.  obf: f
     */
    public static int _profilingCounter_f;

    // -------------------------------------------------------------------------
    // XOR-decoded string pool (for error messages)
    //
    // Encoded literals are decoded at class-load time by the two private z() helpers.
    // Key table (mod-5): [81, 114, 22, 82, 98]  ('Q', 'r', 0x16, 'R', 'b')
    // Single-char fallback key: 0x62 ('b')
    //
    // Decoded values:
    //   ERRMSG[0] = "ib.OA("              — signature prefix for loadResource catch
    //   ERRMSG[1] = "{...}"               — non-null String placeholder in error message
    //   ERRMSG[2] = " len="               — label in CRC-mismatch diagnostic
    //   ERRMSG[3] = ": crc="              — label in CRC-mismatch diagnostic
    //   ERRMSG[4] = "Couldn't download file #"  — IOException prefix
    //   ERRMSG[5] = "content"             — URL path component for content files
    //   ERRMSG[6] = "null"                — null String placeholder in error message
    //   ERRMSG[7] = "ib.PA("              — signature prefix for unlinkSelf catch
    //   ERRMSG[8] = "ib.QA("              — signature prefix for bitwiseAnd catch
    // -------------------------------------------------------------------------

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /** Default constructor; subclasses call super() implicitly.  obf: ib() */
    protected StreamBase() {
    }

    // -------------------------------------------------------------------------
    // Doubly-linked list operations
    // -------------------------------------------------------------------------

    /**
     * Removes this node from the doubly-linked ring it currently belongs to.
     *
     * <p>The caller passes the sentinel value {@code -27331} as a guard (obfuscation artifact).
     * When the guard matches (always, in live code) and this node is currently linked
     * ({@code prev != null}), the node stitches its neighbours together and nulls its own
     * link pointers.
     *
     * <p>Called by {@link client.util.LinkedQueue} (obf: db) and {@link client.audio.FilterChain}
     * (obf: va) when removing a node from a queue or filter chain.
     *
     * @param sentinel  must be {@code -27331}; other values are a no-op (dead obfuscation guard).
     *                  obf param: var1
     * obf: void a(int)
     */
    public final void unlinkSelf(int sentinel) {
        // obf: ++f; — dead profiling counter, removed
        if (sentinel == -27331) {
            if (this.prev != null) {
                // Stitch neighbours together, removing this node from the ring
                this.prev.next = this.next;
                this.next.prev = this.prev;
                this.next = null;
                this.prev = null;
            }
        }
    }

    // -------------------------------------------------------------------------
    // Static utility
    // -------------------------------------------------------------------------

    /**
     * Bit-AND utility.  {@code return a & b}.
     *
     * <p>Used widely by timer, CRC, and color-math helpers across the codebase as an indirect
     * bitmask call (obfuscation indirection to break static analysis).  For example,
     * {@link client.util.Timer} (obf: p) and {@link client.data.CacheUpdater} (obf: cb) both
     * call this in place of inline {@code &} expressions.
     *
     * @param a  left operand   obf: var0 / n2
     * @param b  right operand  obf: var1 / n3
     * @return   {@code a & b}
     * obf: static int a(int, int)
     */
    public static int bitwiseAnd(int a, int b) {
        return a & b;
    }

    /**
     * Downloads and returns a numbered game content file (resource), with CRC verification.
     *
     * <p>Lookup / download sequence:
     * <ol>
     *   <li>If {@link client.util.ClientRuntimeException ClientRuntimeException}.g[resourceId]
     *       (obf: {@code la.g}) is already non-null, return it immediately from the in-memory
     *       cache.</li>
     *   <li>If the anti-tamper guard {@code priority > -73} fires, return {@code null} (this
     *       path is a dead obfuscation guard — priority is always &le; -73 in live calls).</li>
     *   <li>If a live {@link client.data.ArchiveReader} (obf: {@code m.e}, type {@code ob})
     *       exists, try to read the resource from it via
     *       {@code ArchiveReader.load(9395, resourceId)}.  If the CRC matches
     *       {@code Buffer._junkArray[resourceId]} (obf: {@code tb.l[resourceId]}), decompress
     *       via {@code World.unpackData(128, true, data)} and cache it.</li>
     *   <li>Otherwise build the URL
     *       {@code BASE_URL + "content" + resourceId + "_" + toHexString(expectedCrc)},
     *       and attempt to download it up to 3 times via
     *       {@link client.net.ClientStream#downloadFile} (obf: {@code da.a(URL,true,true)}).
     *       On each attempt verify the CRC; on success store in the archive cache (if present)
     *       and decompress into the in-memory cache.</li>
     *   <li>On final failure throw {@link IOException} with a diagnostic message that includes
     *       the resource id, expected CRC, actual length, and first few data bytes.</li>
     * </ol>
     *
     * <p>Obfuscation notes stripped:
     * <ul>
     *   <li>{@code boolean bl = client.vh;} opaque predicate (always false) — all dead branches
     *       removed.</li>
     *   <li>{@code ++b;} profiling counter at entry — removed.</li>
     *   <li>{@code if (param0 > -73) return null;} anti-tamper guard — removed (note: param0
     *       was an unused dummy int dropped here).</li>
     *   <li>Loop {@code while (~n5 > -4)} = {@code while (n5 < 3)} — 3 download retries.</li>
     *   <li>CRC comparison {@code ~mb.a(data,len,0) != ~tb.l[id]} = CRC mismatch check
     *       (double-negation is an obfuscation no-op).</li>
     *   <li>Inner byte-dump loop {@code while (~data.length < ~n6 && -6 >= ~n6)} iterates
     *       the first min(data.length, 5) bytes for the diagnostic string.</li>
     *   <li>Inner {@code try/catch (IOException)} retry guard wraps ONLY the archive-store +
     *       decompress + return; the download ({@code da.a}) and CRC check ({@code mb.a}) sit
     *       OUTSIDE it.  A download IOException therefore propagates immediately and is never
     *       retried.</li>
     * </ul>
     *
     * @param _unusedPriority  dummy anti-tamper guard parameter (always &le; -73 in practice);
     *                         dropped from renamed signature  obf: param0 / n2
     * @param _unusedString    dummy string parameter (written to {@code o.l} / ISAAC.statusMsg
     *                         for debugging); dropped from renamed signature  obf: param1 / string
     * @param statusCode       written to {@link client.data.DataStore} (obf: {@code nb.q}) as a
     *                         download-status tag; obf: param2 / n3
     * @param resourceId       index into the content file table (0–11 in this build);
     *                         used for cache lookup, URL construction, and CRC lookup.
     *                         obf: param3 / n4
     * @return the raw (decompressed) resource bytes, cached in
     *         {@code ClientRuntimeException.g[resourceId]} (obf: {@code la.g[resourceId]})
     * @throws IOException if all download attempts fail or the CRC never matches
     * obf: static byte[] a(int, String, int, int)
     */
    public static final byte[] loadResource(int _unusedPriority, String _unusedString,
                                     int statusCode, int resourceId) throws IOException {
        // obf: ++b; — dead profiling counter, removed

        // Step 1: Return from in-memory cache if already loaded
        // obf: la.g = ClientRuntimeException.g (byte[][] of cached content files)
        if (client.util.ClientRuntimeException.byteRowScratch[resourceId] != null) {
            return client.util.ClientRuntimeException.byteRowScratch[resourceId];
        }

        // Step 2: Anti-tamper guard — in live code _unusedPriority is always <= -73.
        // The original: if (n2 > -73) return null;
        // Kept as documentation; in practice this branch is never taken.
        // (Dropped dummy param from renamed sig above.)

        // Write download-status tag for debugging
        // obf: nb.q = DataStore.unused_q; o.l = ISAAC.unusedL
        client.data.DataStore.unused_q = statusCode;
        client.net.ISAAC.unusedL = _unusedString;

        // Step 3: Try archive (ob = ArchiveReader, m.e = SocketFactory.globalArchive)
        // obf: m.e is the global ArchiveReader instance held on SocketFactory
        client.data.ArchiveReader archiveReader = client.net.SocketFactory.globalArchive;
        if (archiveReader != null) {
            // obf: ob.a(9395, resourceId) — read archive entry with magic id 9395
            byte[] cached = archiveReader.load(9395, resourceId);
            if (cached != null) {
                // Verify CRC: mb.a(data, len, 0) computes CRC via w.a()
                // obf: tb.l = Buffer._junkArray (int[] of expected CRC values per resource)
                if (client.util.Utility.computeCrc32(cached, cached.length, 0)
                        == client.net.Buffer._junkArray[resourceId]) {
                    // Decompress and cache
                    // obf: k.a(128, true, data) = World.unpackData (strips 6-byte header,
                    // else BZip-decompresses).  k = World per NAMING.md (NOT Scene = lb).
                    client.util.ClientRuntimeException.byteRowScratch[resourceId] =
                            client.world.World.unpackData(128, true, cached);
                    return client.util.ClientRuntimeException.byteRowScratch[resourceId];
                }
            }
        }

        // Step 4: Download from server
        // URL pattern: BASE_URL + "content" + resourceId + "_" + toHexString(expectedCrc)
        // e.g.: "http://world1.runescape.com/content3_1a2b3c4d"
        // obf: z[5] = "content"
        URL resourceUrl = new URL(BASE_URL,
                "content" + resourceId + "_"
                + Long.toHexString(client.net.Buffer._junkArray[resourceId]));

        byte[] data = null;
        // obf: while (~n5 > -4)  <==>  while (n5 < 3)  — up to 3 attempts
        for (int attempt = 0; attempt < 3; attempt++) {
            // NOTE: in the original bytecode, the download (da.a) and the CRC
            // comparison (mb.a) are OUTSIDE the inner try/catch.  Only the
            // archive-store + decompress + return are wrapped by the IOException
            // retry-catch.  An IOException thrown by the *download itself*
            // therefore propagates straight out of loadResource (it is NOT
            // retried and NOT caught by the outer RuntimeException wrapper).
            // Verified against decompiled/cfr/ib.java lines 81-103 and the
            // two split try-regions in the clean-base bytecode listing.

            // obf: da.a(URL, true, true) = ClientStream.downloadFile(url, followRedirects, retry)
            // (outside the try — IOException here is not caught/retried)
            data = client.net.ClientStream.downloadFile(resourceUrl, true, true);

            // Verify CRC of downloaded data (also outside the try)
            // obf: ~mb.a(data,data.length,0) != ~tb.l[resourceId]
            //      double-bit-NOT is identity; this is just an inequality check
            //      with negation to defeat simple pattern matching.  Equivalent to:
            //      mb.a(data,data.length,0) != tb.l[resourceId]
            int downloadedCrc = client.util.Utility.computeCrc32(data, data.length, 0);
            int expectedCrc   = client.net.Buffer._junkArray[resourceId];
            if (downloadedCrc != expectedCrc) {
                // obf: break block32 — CRC mismatch, advance to next attempt
                continue;
            }

            try {
                // CRC matches — store in archive cache if available
                // obf: ob.a(resourceId, data.length, -97, data) — write entry back to archive
                if (archiveReader != null) {
                    archiveReader.store(resourceId, data.length, -97, data);
                }

                // Decompress and cache
                // obf: k.a(128, true, data) = World.unpackData (k = World per NAMING.md)
                client.util.ClientRuntimeException.byteRowScratch[resourceId] =
                        client.world.World.unpackData(128, true, data);
                return client.util.ClientRuntimeException.byteRowScratch[resourceId];

            } catch (IOException retryEx) {
                // obf: if (~n5 == -3) throw; else break block32 (retry)
                // ~n5 == -3  <==>  n5 == 2  — last attempt, re-throw
                if (attempt == 2) {
                    throw retryEx;
                }
                // else: silently retry (fall through to ++attempt)
            }
        }

        // Step 5: All attempts exhausted — throw with diagnostic
        if (data != null) {
            // Build diagnostic: "Couldn't download file #N: crc=X len=Y b0 b1 b2 b3 b4"
            // obf: z[4]="Couldn't download file #", z[3]=": crc=", z[2]=" len="
            StringBuilder diag = new StringBuilder(
                    "Couldn't download file #" + resourceId
                    + ": crc=" + client.net.Buffer._junkArray[resourceId]);
            diag.append(" len=").append(data.length);
            // Dump first min(data.length, 5) bytes for diagnostics
            // obf: while (~data.length < ~n6 && -6 >= ~n6)
            //      <==>  while (n6 < data.length && n6 < 5)
            for (int i = 0; i < data.length && i < 5; i++) {
                diag.append(" ").append(data[i]);
            }
            throw new IOException(diag.toString());
        }

        // data is still null — download never produced any bytes at all
        // obf: z[4]="Couldn't download file #", z[3]=": crc="
        throw new IOException(
                "Couldn't download file #" + resourceId
                + ": crc=" + client.net.Buffer._junkArray[resourceId]);
    }
}
