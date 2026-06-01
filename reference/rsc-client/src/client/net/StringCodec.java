package client.net;

import java.awt.image.ImageConsumer;
import client.util.ErrorHandler;
import client.scene.SurfaceImageProducer;   // fb — hosts the shared BZip encoder (fb.a)

/**
 * StringCodec (obf: u) — Miscellaneous static helpers for the network layer.
 *
 * Three responsibilities are bundled here by the obfuscator into a single class:
 *
 *   1. {@link #writeString writeString(Buffer, String)} — Encodes a Java String
 *      to RSC's on-wire representation and appends it length-prefixed into a
 *      {@link Buffer} (tb). The encoding is delegated to {@link TextEncoder} (h),
 *      which maps Unicode special chars to Windows-1252-compatible bytes. The raw
 *      byte copy into the buffer is performed by {@link BZip} (aa) via the global
 *      instance {@link SurfaceImageProducer#bzipRef} (fb.a) with flag 119 (pass-
 *      through / plain-copy mode). Returns the total byte count appended.
 *
 *   2. {@link #lookupRegisteredEntry lookupRegisteredEntry(boolean, int)} — Linear
 *      scan of the global 3-element id-registry (returned by
 *      {@link ProxySocketFactory#getRegistry getRegistry(69)}) for an
 *      {@link ErrorHandler} (i) entry whose {@code id} field matches a target int.
 *      If not found and {@code mustExist} is true, writes {@code -2} to
 *      {@link #STATUS_NOT_FOUND}.
 *
 *   3. {@link #sleepIfZero sleepIfZero(int, long)} — Conditional thread-sleep:
 *      sleeps for the given milliseconds only when the guard argument is 0.
 *      Matches oracle {@code Utility.sleep()}.
 *
 * Package context: client.net. All three methods are called from packet-building and
 * initialisation code in Mudclient (client) and ClientStream (da).
 *
 * Obfuscation notes for this class:
 *   - Static fields {@link #DEAD_INT_ARRAY}, {@link #DEAD_IMAGE_CONSUMER}, and
 *     {@link #DEAD_STRING_ARRAY} are written only inside opaque-predicate dead
 *     branches and carry no runtime meaning.
 *   - {@link #writeStringCallCount}, {@link #lookupCallCount}, and
 *     {@link #sleepCallCount} are per-method profiling counters (incremented once
 *     per call; never read for logic).
 *   - The XOR string pool {@code z[]} (5 entries) encodes only the per-method
 *     error-context strings used when re-throwing exceptions through ErrorHandler.
 *     Decoded values are shown below.
 */
public final class StringCodec {

    // -------------------------------------------------------------------------
    // Dead anti-tamper / opaque-predicate fields (no runtime meaning)
    // -------------------------------------------------------------------------

    /**
     * Dead anti-tamper field — written {@code null} only in a dead branch gated by
     * the opaque predicate {@code if (param0 <= 10)}.
     * obf: u.a  (static int[])
     */
    public static int[] DEAD_INT_ARRAY;                         // obf: a

    /**
     * The active ImageConsumer registered with {@link SurfaceImageProducer}.
     * Despite being declared in this class by the obfuscator, this field is
     * owned conceptually by SurfaceImageProducer (fb): it is written in
     * {@code fb.addConsumer}, cleared in {@code fb.removeConsumer}, and tested
     * in {@code fb.isConsumer}. StringCodec itself never reads or writes it.
     * The obfuscator scattered fields across classes as a disassembly deterrent.
     * obf: u.d  (static ImageConsumer)
     */
    public static ImageConsumer DEAD_IMAGE_CONSUMER;            // obf: d

    /**
     * Dead anti-tamper field — never written from real code paths.
     * obf: u.b  (static String[])
     */
    public static String[] DEAD_STRING_ARRAY;                   // obf: b

    // -------------------------------------------------------------------------
    // Live state
    // -------------------------------------------------------------------------

    /**
     * Sentinel set to {@code -2} by {@link #lookupRegisteredEntry} when the
     * requested id is not found in the registry and {@code mustExist} was true.
     * Initialized to 0 (no error). callers can inspect this field after a null
     * return to distinguish "not registered" from a normal absent result.
     * obf: u.g  (static int)
     */
    public static int STATUS_NOT_FOUND = 0;                     // obf: g

    // -------------------------------------------------------------------------
    // Per-method profiling counters (incremented once at entry; never used for
    // logic; all deletable)
    // -------------------------------------------------------------------------

    /** obf: u.f — call counter for {@link #writeString}. */
    public static int writeStringCallCount;                     // obf: f

    /** obf: u.c — call counter for {@link #lookupRegisteredEntry}. */
    public static int lookupCallCount;                          // obf: c

    /** obf: u.e — call counter for {@link #sleepIfZero}. */
    public static int sleepCallCount;                           // obf: e

    // -------------------------------------------------------------------------
    // XOR-encrypted error-context string pool  (obf: z[])
    //
    // Decoded by two-pass XOR at class load time:
    //   Pass 1 (z(String)->char[]): single-char inputs XOR'd with 'R' (0x52);
    //                               multi-char inputs pass through unchanged.
    //   Pass 2 (z(char[])->String): each char[i] XOR'd with KEY[i % 5] where
    //                               KEY = {112, 13, 56, 50, 82}.
    //
    // Decoded values:
    //   z[0]  =  "u.C("   — error context prefix for lookupRegisteredEntry
    //   z[1]  =  "{...}"  — non-null object placeholder in error messages
    //   z[2]  =  "null"   — null-argument placeholder in error messages
    //   z[3]  =  "u.B("   — error context prefix for writeString
    //   z[4]  =  "u.A("   — error context prefix for sleepIfZero
    // -------------------------------------------------------------------------

    /**
     * Error-context string pool, decoded at static initialization from the
     * XOR-obfuscated literals. Only used in the exception re-throw wrappers.
     * obf: z  (private static final String[])
     */
    private static final String[] ERROR_CTX = new String[]{
        /* [0] obf: z(z("#{"))  -> */ "u.C(",
        /* [1] obf: z(z("#/")) -> */ "{...}",
        /* [2] obf: z(z("xT^"))       -> */ "null",
        /* [3] obf: z(z("#z"))  -> */ "u.B(",
        /* [4] obf: z(z("#y"))  -> */ "u.A(",
    };

    // =========================================================================
    // Public API
    // =========================================================================

    /**
     * Encodes {@code text} as bytes and writes it into {@code buf} as a
     * length-prefixed field at the current cursor position.
     *
     * On-wire layout (in {@code buf.data}):
     * <pre>
     *   [smart-length: 1 byte if 0..127, else 2 bytes with high-bit flag set]
     *   [encoded body: length bytes]
     * </pre>
     *
     * The encoding step ({@link TextEncoder#encode TextEncoder.encode(text, 30)})
     * maps Unicode special characters (curly quotes U+2018/U+2019/U+201C/U+201D,
     * euro sign U+20AC, en-dash U+2013, em-dash U+2014, ellipsis U+2026, etc.)
     * to their Windows-1252 / RSC-protocol byte equivalents; plain ASCII/Latin-1
     * bytes pass through directly.
     *
     * The body-copy step uses {@link SurfaceImageProducer#bzipRef} (fb.a) —
     * the shared {@link BZip} instance — invoked with flag 119, which selects its
     * plain memory-copy branch (no actual BZip2 processing).  The method returns
     * the number of bytes it copied, which is added to {@code buf.cursor}.
     *
     * obf: u.a(int, tb, String)
     *   — the leading {@code int} param (var0) is an anti-tamper guard:
     *     {@code if (var0 <= 10) DEAD_INT_ARRAY = null;} — fully dead, dropped here.
     *   — the trailing return value was {@code buf.cursor - startOffset}.
     *
     * @param buf   the output {@link Buffer} to write into
     * @param text  the string to encode and append
     * @return      total byte count appended to {@code buf} (length prefix + body)
     */
    public static final int writeString(Buffer buf, String text) {
        // obf: f++ (profiling counter — deleted)

        // Snapshot the cursor so we can report total bytes written at the end.
        // obf var1.w = Buffer.offset (declared; map row "cursor" stale).
        int startOffset = buf.offset;               // obf: int var9 = var1.w

        // Encode the Java String to RSC's custom byte encoding.
        // TextEncoder.encode maps Unicode special chars to CP-1252 equivalents.
        byte[] encodedBytes = TextEncoder.encode(text, (byte) 30);  // obf: h.a(var2, (byte)30)

        // Write the encoded length as a compact "smart int":
        //   0..127   → 1 byte  as-is (0x00..0x7F)
        //   128..32767 → 2 bytes with the high bit of the first byte set as a flag
        // obf var1.b(int,byte) is declared Buffer.putVariableLengthShort (byte guard dropped).
        buf.putVariableLengthShort(encodedBytes.length);  // obf: var1.b(var4.length, (byte)-88)

        // Copy the encoded bytes into the buffer's underlying byte array.
        // SurfaceImageProducer.bzip is the shared BZip instance (fb.a : aa).
        // obf fb.a.a(count, dest, destOff, src, srcOff, 119) is declared BZip.encode
        // (the 6th arg, the anti-tamper guard 119, was dropped from the deob signature):
        //   encode(srcLen, destBuf, destOff, srcBuf, srcOff)
        //   → copies srcBuf[srcOff..srcOff+srcLen) into destBuf[destOff..]
        //   → returns the number of bytes copied (= srcLen)
        buf.offset = buf.offset
                + SurfaceImageProducer.bzip.encode(  // obf: fb.a.a(...)
                        encodedBytes.length,   // number of bytes to copy
                        buf.data,              // obf: var1.F — destination array
                        buf.offset,            // obf: var1.w — destination offset
                        encodedBytes,          // obf: var4   — source array
                        0);                    // source offset (guard 119 dropped)

        // Return bytes consumed (smart-length prefix + body).
        return buf.offset - startOffset;       // obf: return var1.w - var9
    }

    /**
     * Searches the global 3-element id-registry for an {@link ErrorHandler} entry
     * whose {@code id} field matches {@code targetId}.
     *
     * The registry is fetched via {@link ProxySocketFactory#getRegistry
     * ProxySocketFactory.getRegistry(69)}, which returns:
     * <pre>
     *   { AudioMixer.registeredEntry,             // eb.e  (id = ?)
     *     SurfaceImageProducer.registeredEntry,   // fb.h  (id = 1)
     *     RecordLoader.registeredEntry }           // f.b   (id = ?)
     * </pre>
     *
     * If no entry matches:
     *   - if {@code mustExist} is {@code true}, sets {@link #STATUS_NOT_FOUND}
     *     to {@code -2} to signal a hard miss to the caller.
     *   - returns {@code null}.
     *
     * obf: u.a(boolean, int)
     *   — {@code boolean var0} was the first param (mustExist)
     *   — {@code int var1} was the second param (targetId)
     *   — opaque predicate {@code boolean var5 = client.vh} (always false) gated
     *     dead {@code break} jumps inside the loop; stripped.
     *   — profiling counter {@code ++c} at entry; stripped.
     *
     * @param mustExist  when true and not found, sets STATUS_NOT_FOUND = -2
     * @param targetId   integer id to match against {@code ErrorHandler.id} (obf: i.a)
     * @return           matching entry, or {@code null} if none found
     */
    public static final ErrorHandler lookupRegisteredEntry(boolean mustExist, int targetId) {
        // obf: boolean var5 = client.vh  — opaque predicate, always false; deleted.
        // obf: ++c (profiling counter — deleted)

        // Fetch the live registry snapshot. The key value 69 is a compile-time
        // constant that ProxySocketFactory.getRegistry checks as an anti-tamper
        // guard (it clears its own field if the value is <= 37).
        // obf gb.a(int) is declared ProxySocketFactory._junkRegistration (map "getRegistry" stale).
        ErrorHandler[] registry = ProxySocketFactory._junkRegistration(69);  // obf: gb.a(69)

        for (int i = 0; i < registry.length; i++) {
            ErrorHandler entry = registry[i];
            // obf: n4 = var4.a  (read the int id from the ErrorHandler instance)
            // obf: if (var4.a == var1) return var4;
            // obf i.a (instance) is declared ErrorHandler.instanceValue (map "id" stale).
            if (entry.instanceValue == targetId) {
                return entry;
            }
        }

        // Not found: signal miss to caller if required.
        // obf: if (var10000 != 0) g = -2;
        if (mustExist) {
            STATUS_NOT_FOUND = -2;
        }
        return null;
    }

    /**
     * Sleeps the current thread for {@code millis} milliseconds, but only if
     * {@code guard == 0}; returns immediately for any other guard value.
     *
     * {@link InterruptedException} is silently swallowed, matching the original.
     *
     * This matches the oracle {@code Utility.sleep(int, long)} pattern where the
     * first argument is a "sleep unless already done" flag.
     *
     * obf: u.a(int, long)
     *   — {@code int var0} was the guard (non-zero → skip)
     *   — {@code long var1} was the sleep duration in milliseconds
     *   — profiling counter {@code ++e} executed AFTER the sleep; stripped.
     *
     * @param guard   if non-zero, return immediately without sleeping
     * @param millis  sleep duration in milliseconds
     */
    public static final void sleepIfZero(int guard, long millis) {
        // obf: if (var0 != 0) return;
        if (guard != 0) {
            return;
        }
        try {
            Thread.sleep(millis);
        } catch (InterruptedException ignored) {
            // Intentionally swallowed — matches original empty catch block.
        }
        // obf: e++ (profiling counter after sleep — deleted)
    }

    // =========================================================================
    // Private helpers — XOR string-pool decryption
    // (Identical two-overload z() pattern used in every class in this build.)
    // =========================================================================

    /**
     * First pass of the two-pass XOR decryption used for the per-class string pool.
     * Converts a raw obfuscated {@code String} literal to a {@code char[]}.
     * If the string has fewer than 2 characters, the single character is XOR'd with
     * {@code 'R'} (0x52); longer strings pass through unchanged.
     * obf: private static char[] z(String)
     */
    @SuppressWarnings("unused")
    private static char[] decodeXorFirst(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 'R'); // 0x52
        }
        return chars;
    }

    /**
     * Second pass of the two-pass XOR decryption: XORs each {@code char[i]} with
     * {@code KEY[i % 5]} and returns the result as an interned String.
     * KEY bytes by position mod 5: {112, 13, 56, 50, 82}.
     * obf: private static String z(char[])
     */
    @SuppressWarnings("unused")
    private static String decodeXor(char[] chars) {
        // Mod-5 XOR key table; position 0..4 cycle:
        //   0 → 0x70 (112 = 'p')
        //   1 → 0x0D (13  = '\r')
        //   2 → 0x38 (56  = '8')
        //   3 → 0x32 (50  = '2')
        //   4 → 0x52 (82  = 'R')
        final byte[] KEY = {112, 13, 56, 50, 82};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char) (chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
