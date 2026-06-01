package client.util;

import java.io.DataInputStream;
import java.net.URL;

/**
 * Utility — static helper class used throughout the engine.
 *
 * <p>Three distinct responsibilities:
 * <ol>
 *   <li><b>Byte-array pool / allocation</b> — {@link #allocateByteArray(int, byte)} hands out
 *       pooled {@code byte[]} instances from three separate pools (ArchiveReader, GameShell,
 *       GameModel). Falls back to a size-keyed pool maintained in EntityDef, and finally
 *       allocates fresh arrays when all pools are exhausted.
 *   <li><b>Timing / profiling relay</b> — {@link #sleepWithProfile(int, long)} delegates to
 *       {@code StringCodec.sleep(int, long)} and is the single call-site that all game loops use
 *       to sleep; it also resets the chat-line cache ({@link #chatLines}) when called outside the
 *       normal sleep path (opcode 11200).
 *   <li><b>Client error reporting</b> — {@link #reportError(int, Throwable, String)} URL-encodes
 *       a stack trace + optional context string and posts it to
 *       {@code clienterror.ws?c=<version>&u=<user>&v1=<world>&v2=<rev>&e=<encoded-error>}.
 *   <li><b>Number / chat formatting</b> — {@link #formatNumber(int, int)} inserts comma separators
 *       and optional colour tags ({@code @gre@}, {@code @cya@}, {@code @whi@}); the 8-argument
 *       overload {@link #formatChatLine(String, String, boolean, int)} assembles chat prefix/suffix
 *       text used for tell/whisper/trade messages with various display modes.
 * </ol>
 *
 * <p><b>Engine context:</b> obfuscated class {@code mb}. Matches oracle {@code Utility} in
 * mudclient204, though this revision strips the file-I/O helpers (openFile/readFully/getBitMask)
 * which have moved into other classes (ArchiveReader, Buffer).
 *
 * <p>Obfuscation notes stripped:
 * <ul>
 *   <li>Opaque predicate {@code boolean bl = client.vh} (always false) — all dead branches removed.
 *   <li>Per-method profiling counters ({@code ++mb.j}, {@code ++mb.h}, etc.) — removed; counter
 *       fields renamed to indicate their dead-profiling purpose.
 *   <li>Exception wrapper {@code catch(RuntimeException e){throw i.a(e,"sig(...)")}} — unwrapped.
 *   <li>XOR string pool {@code z(z("…"))} — decoded; literal values written inline with comments.
 *   <li>Anti-tamper guards ({@code if (param != bigConstant) return;}) — removed.
 * </ul>
 *
 * obf: mb
 */
final class Utility {

    // -----------------------------------------------------------------------
    // Dead profiling counters (obfuscator inserted ++counter at method entry)
    // -----------------------------------------------------------------------

    /** Dead profiling counter for allocateByteArray calls. obf: mb.j */
    static int _profileAllocateByteArray;

    /** Dead profiling counter for sleepWithProfile calls. obf: mb.h */
    static int _profileSleepWithProfile;

    /** Dead profiling counter for formatNumber calls. obf: mb.f */
    static int _profileFormatNumber;

    /** Dead profiling counter for formatChatLine calls. obf: mb.d */
    static int _profileFormatChatLine;

    /** Dead profiling counter for reportError calls. obf: mb.e */
    static int _profileReportError;

    /** Dead profiling counter for computeCrc32 calls. obf: mb.c */
    static int _profileComputeCrc32;

    // -----------------------------------------------------------------------
    // Bit-mask lookup table (2^n - 1 for n = 0..31, then -1 for n=32)
    // Used by bit-field readers elsewhere in the engine (e.g. Buffer).
    // obf: mb.i
    // -----------------------------------------------------------------------

    /** Bitmask table: {@code BITMASK[n] == (1 << n) - 1} for n = 0..31, {@code BITMASK[32] == -1}. */
    static final int[] BITMASK = new int[]{
        0,
        1,
        3,
        7,
        15,
        31,
        63,
        127,
        255,
        511,
        1023,
        2047,
        4095,
        8191,
        16383,
        32767,
        65535,
        131071,
        262143,
        524287,
        1048575,
        2097151,
        4194303,
        8388607,
        16777215,
        33554431,
        67108863,
        134217727,
        268435455,
        536870911,
        1073741823,
        Integer.MAX_VALUE,
        -1
    };

    // -----------------------------------------------------------------------
    // Fields used by allocateByteArray's small-pool bookkeeping
    // -----------------------------------------------------------------------

    /**
     * Stack pointer into the GameModel byte-array pool ({@code GameModel.bytePool}).
     * Tracks how many arrays have been reclaimed. obf: mb.b
     */
    private static int gameModelPoolTop = 0;

    // -----------------------------------------------------------------------
    // Fields used across methods
    // -----------------------------------------------------------------------

    /**
     * Size-keyed pool counts — parallel to {@code EntityDef.sizedPool} ({@code t.n}).
     * Each entry gives the number of currently available arrays of a given size class.
     * obf: mb.a  (NOTE: this is NOT the class reference "a" / JSBridge — it is a field
     * on mb itself; the obf-i caveat in NAMING.md applies here too.)
     */
    static int[] sizedPoolCounts;

    /**
     * Miscellaneous flag/state written by {@link #formatChatLine}.
     * Set to 90 when {@code rightToLeft} is false; role unclear beyond that. obf: mb.l
     */
    static int chatLineState = 0;

    /**
     * Cached array of recent chat lines, cleared by {@link #sleepWithProfile} when the
     * sleep opcode is not {@code 11200}. obf: mb.g
     */
    static String[] chatLines;

    /**
     * Tracks the size-pool slot last used by {@link #allocateByteArray}.  Not a public API.
     * obf: mb.k — exact semantics unclear beyond being a spare int[] scratch field.
     */
    static int[] k;

    // -----------------------------------------------------------------------
    // XOR-encoded string pool (decoded at class load time).
    //
    // Key: z(String)->char[]  XOR with 0x36 when len<2, else identity.
    //       z(char[])->String  XOR each char by key[i%5] where
    //       key = {56, 94, 102, 21, 54}.
    //
    // Decoded values listed below:
    //   z[ 0] = "mb.F("           — error-sig prefix for sleepWithProfile
    //   z[ 1] = "mb.C("           — error-sig prefix for allocateByteArray
    //   z[ 2] = "mb.E("           — error-sig prefix for computeCrc32
    //   z[ 3] = "null"            — null-arg indicator in error sigs
    //   z[ 4] = "{...}"           — non-null-arg indicator in error sigs
    //   z[ 5] = "You tell "       — chat prefix for mode 2 (outgoing tell)
    //   z[ 6] = " tells you: "    — chat join string for mode 1/mode 2
    //   z[ 7] = ": "              — generic chat join separator
    //   z[ 8] = " wishes to trade with you."  — trade request suffix (mode 6)
    //   z[ 9] = "mb.D("           — error-sig prefix for formatChatLine
    //   z[10] = " | "             — separator between stack trace and context
    //   z[11] = "clienterror.ws?c=" — base error-report URL path
    //   z[12] = "%40"             — URL-encoded "@"
    //   z[13] = "%23"             — URL-encoded "#"
    //   z[14] = "&e="             — error-report URL param for encoded error text
    //   z[15] = "%3a"             — URL-encoded ":"
    //   z[16] = "%26"             — URL-encoded "&"
    //   z[17] = "&v1="            — error-report URL param for world/server
    //   z[18] = "&v2="            — error-report URL param for client revision
    //   z[19] = "&u="             — error-report URL param for username
    //   z[20] = "K @whi@("        — colour-tag suffix for formatNumber (billions)
    //   z[21] = "@cya@"           — colour-tag prefix for formatNumber (millions)
    //   z[22] = " million @whi@(" — colour-tag label for formatNumber (millions)
    //   z[23] = "mb.B("           — error-sig prefix for formatNumber
    //   z[24] = "@gre@"           — colour-tag prefix for formatNumber (billions)
    // -----------------------------------------------------------------------

    // Chat-line display modes used by formatChatLine (param `mode`):
    //   0 — default / incoming message: prefix + ": " + message (or just message if no prefix)
    //   1 — incoming private tell:       prefix + " tells you: " + message
    //   2 — outgoing private tell:       "You tell " + prefix + ": " + message
    //   3 — incoming channel/clan:       prefix + ": " + message
    //   4 — tell variant (~mode == -5):  prefix + ": " + message (same as 3)
    //   5 — message-only (~mode == -6):  message only, no prefix
    //   6 — trade request:               prefix + " wishes to trade with you."
    //   7 — tell variant (~mode == -8):  prefix + ": " + message (same as 0)
    //  >=8 — (unrecognised, ~mode < -8): returns ""

    /**
     * Allocates a {@code byte[]} of the requested size, drawing from one of three object pools
     * before falling back to a size-keyed pool (EntityDef.sizedPool) and, ultimately, a fresh
     * array allocation.
     *
     * <p>Pool priority (highest to lowest):
     * <ol>
     *   <li>ArchiveReader pool ({@code ob.j}) when {@code size == 100}
     *   <li>GameShell pool ({@code e.kb}) when {@code size == 5000}
     *   <li>GameModel pool ({@code ca.tb}) when a general reclaim is possible
     *       ({@code size == ~-30001 == 30000}, i.e. the XOR trick; actually the condition
     *        decodes to {@code size == 30000}).
     *   <li>EntityDef size-keyed pool ({@code t.n[slot][v.g[slot]--]}) — scanned linearly.
     *   <li>Fresh {@code new byte[size]}.
     * </ol>
     *
     * <p>The {@code guardByte} parameter acts as a sentinel: if {@code guardByte > -97}
     * (i.e. the value is in the range [-96..127]) the call is a no-op and returns {@code null}.
     * This is an anti-tamper guard in the obfuscated build; it is never true in practice.
     *
     * <p>This method is {@code synchronized} because multiple threads (loader, game, audio) share
     * the pool arrays.
     *
     * obf: mb.a(int, byte)
     *
     * @param size      requested byte-array size (also used as a pool-selector key)
     * @param guardByte obfuscator sentinel; must be {@code <= -97} for real allocations
     * @return a {@code byte[]} of the requested length (possibly recycled)
     */
    static final synchronized byte[] allocateByteArray(int size, byte guardByte) {
        // Obfuscation stripped: opaque predicate `boolean bl = client.vh` (always false) removed.
        // Dead profiling counter (++mb.j) removed.

        // Pool 1: ArchiveReader byte-array pool (pre-allocated stacks of 1000 arrays, key == 100)
        // obf: ob.j is ArchiveReader.bytePool (byte[][] with 1000 slots)
        // obf: n.b  is FontWidths.b (int, stack pointer into ob.j)
        // Condition: ~n.b < -1  <=>  n.b > 0
        if (size == 100 && client.data.FontWidths.b > 0) {
            // Pop from ArchiveReader pool (LIFO)
            byte[] recycled = client.data.ArchiveReader.bytePool[--client.data.FontWidths.b];
            client.data.ArchiveReader.bytePool[client.data.FontWidths.b] = null; // clear slot for GC
            return recycled;
        }

        // Pool 2: GameShell byte-array pool (pre-allocated stacks of 250 arrays, key == 5000)
        // obf: e.kb is GameShell.kb (byte[][] with 250 slots) — used for large sprite buffers
        // obf: s.d  is FontBuilder.d (int, stack pointer into e.kb)
        // Condition: ~s.d < -1  <=>  s.d > 0
        if (5000 == size && client.ui.FontBuilder.d > 0) {
            byte[] recycled = client.shell.GameShell.kb[--client.ui.FontBuilder.d];
            client.shell.GameShell.kb[client.ui.FontBuilder.d] = null;
            return recycled;
        }

        // Anti-tamper sentinel guard (obfuscation artifact): callers always pass -97 or lower.
        // In the live client this branch is never taken.
        if (guardByte > -97) {
            return null;
        }

        // Pool 3: GameModel byte-array pool (pre-allocated stacks of 50 arrays)
        // Key: size == 30000  (encoded as: -30001 == ~size, i.e. ~30000 == -30001)
        // obf: ca.tb is GameModel.tb (byte[][] with 50 slots) — recycled model temp buffers
        // obf: mb.b  is Utility.gameModelPoolTop (int, stack pointer — this class field)
        if (size == 30000 && gameModelPoolTop > 0) {
            byte[] recycled = client.scene.GameModel.tb[--gameModelPoolTop];
            client.scene.GameModel.tb[gameModelPoolTop] = null;
            return recycled;
        }

        // Pool 4: EntityDef size-keyed pool — a 2D array of recycled byte arrays bucketed by size.
        // obf: t.n  is EntityDef.n   (byte[][][], outer index = size bucket)
        // obf: e.wb is GameShell.wb  (int[], the array of bucket sizes, one per slot)
        // obf: v.g  is ChatCipher.g  (int[], count of available arrays in each bucket)
        // The pool is only used when t.n has been initialised.
        if (null != client.data.EntityDef.n) {
            for (int slot = 0; slot < client.shell.GameShell.wb.length; slot++) {
                // XOR trick from bytecode: ~e.wb[slot] == ~size  iff  e.wb[slot] == size
                if (client.shell.GameShell.wb[slot] == size
                        && client.net.ChatCipher.g[slot] > 0) {
                    // Pop from this bucket (LIFO)
                    int top = client.net.ChatCipher.g[slot] - 1;
                    client.net.ChatCipher.g[slot] = top;
                    byte[] recycled = client.data.EntityDef.n[slot][top];
                    client.data.EntityDef.n[slot][top] = null; // clear slot for GC
                    return recycled;
                }
            }
        }

        // Fallback: no pool had a matching array — allocate fresh.
        return new byte[size];
    }

    /**
     * Sleeps the current thread for up to {@code millis} milliseconds, with profiling/pacing
     * relay through {@link StringCodec#sleep(int, long)} (obf: {@code u.a}).
     *
     * <p>When called with opcode {@code != 11200}, the shared {@link #chatLines} cache is cleared
     * (set to null). This allows the caller to signal a "reset" without a separate method.
     *
     * <p>The sleep is broken into two StringCodec calls when {@code millis} is divisible by 10:
     * <pre>
     *   StringCodec.sleep(opcode - 11200, millis - 1);  // partial sleep
     *   StringCodec.sleep(0, 1L);                        // final 1ms tick
     * </pre>
     * This is the normal sleep path (opcode == 11200, millis % 10 == 0 is common in game loops).
     * Non-divisible-by-10 values call {@code StringCodec.sleep(0, millis)} directly.
     *
     * <p>If {@code millis <= 0} the method returns immediately.
     *
     * obf: mb.a(int, long)
     *
     * @param opcode opcode selector (11200 = normal game-loop sleep; other values reset chatLines)
     * @param millis sleep duration in milliseconds; ignored if {@code <= 0}
     */
    static final void sleepWithProfile(int opcode, long millis) {
        // Dead profiling counter (++mb.h) removed.

        // Clear chatLines cache when called outside normal sleep path
        if (opcode != 11200) {
            chatLines = null;
        }

        if (millis <= 0L) {
            return;
        }

        if (millis % 10L == 0L) {
            // Split the sleep into two legs so the game loop can interleave work.
            // StringCodec.a(int, long) (obf: u.a) calls Thread.sleep when its first arg == 0.
            // When first arg != 0 it appears to be a no-op (returns immediately) — the subtraction
            // (opcode - 11200) yields 0 only when opcode == 11200, i.e. the normal path.
            StringCodec.sleepMs(opcode - 11200, millis - 1L); // first leg: millis-1
            StringCodec.sleepMs(0, 1L);                        // second leg: 1 ms tick
        } else {
            StringCodec.sleepMs(0, millis);
        }
    }

    /**
     * Formats an integer as a comma-separated decimal string, optionally annotated with
     * RuneScape colour-chat tags for large values.
     *
     * <p>Comma insertion: a comma is added every 3 digits from the right (standard thousands
     * separating), e.g. {@code 1234567} → {@code "1,234,567"}.
     *
     * <p>Colour tags (applied when the number is large enough to overflow the 8-char field):
     * <ul>
     *   <li>If the formatted string has more than 8 characters (i.e. 9+ chars, ≥ 100 million):
     *       result becomes {@code "@gre@" + prefix + " million @whi@(" + full + ")"}.
     *   <li>If it has more than 4 characters after trimming (> 4 chars, ≥ 10,000):
     *       result becomes {@code "@cya@" + trimmedPrefix + "K @whi@(" + full + ")"}.
     *   <li>Otherwise: the raw comma-separated string is returned unchanged.
     * </ul>
     *
     * <p>The {@code guardParam} parameter {@code 131071} (== {@code BITMASK[17]}) is the
     * anti-tamper sentinel. When {@code guardParam != 131071}, the method calls
     * {@code allocateByteArray(null, -74, 53)} (a dead anti-tamper path); that path is
     * never reached in the live client.
     *
     * obf: mb.a(int, int)
     *
     * @param value      the integer value to format
     * @param guardParam anti-tamper sentinel; always 131071 in live code
     * @return formatted string, possibly with colour tags
     */
    static final String formatNumber(int value, int guardParam) {
        // Dead profiling counter (++mb.f) removed.

        // Build initial decimal string
        String result = "" + value;

        // Insert comma separators every 3 digits from the right
        // position moves from (len - 3) down by 3 each iteration
        int insertPos = result.length() - 3;
        while (insertPos > 0) {
            result = result.substring(0, insertPos) + "," + result.substring(insertPos);
            insertPos -= 3;
        }

        // Anti-tamper dead-code path (guardParam check) — never fires in live client:
        // if (guardParam != 131071) { allocateByteArray(null, -74, 53); }

        // Apply colour tags for large numbers
        // String length > 8 means ≥ 100 million (nine chars including comma)
        if (result.length() > 8) {
            // e.g. "123,456,789" -> "@gre@123,456 million @whi@(123,456,789)"
            // Decoded: z[24]="@gre@", z[22]=" million @whi@("
            result = "@gre@"
                   + result.substring(0, result.length() - 8)
                   + " million @whi@("
                   + result + ")";
            return result;
        }

        // String length > 4 means ≥ 10,000 (five chars including comma)
        if (result.length() > 4) {
            // e.g. "12,345" -> "@cya@12K @whi@(12,345)"
            // Decoded: z[21]="@cya@", z[20]="K @whi@("
            result = "@cya@"
                   + result.substring(0, result.length() - 4)
                   + "K @whi@("
                   + result + ")";
        }

        return result;
    }

    /**
     * Assembles a chat-line string by combining a main message and an optional prefix/sender
     * string according to the specified display mode.
     *
     * <p>Display modes (value of {@code mode}):
     * <table border="1">
     *   <tr><th>mode</th><th>Result (prefix = {@code from}, message = {@code msg})</th></tr>
     *   <tr><td>0 (default)</td><td>if from non-null and non-empty: {@code from + ": " + msg}; else {@code msg}</td></tr>
     *   <tr><td>1</td><td>if from non-null/non-empty: {@code from + " tells you: " + msg}; else {@code msg}</td></tr>
     *   <tr><td>2</td><td>if from non-null/non-empty: {@code "You tell " + from + ": " + msg}; else {@code msg}</td></tr>
     *   <tr><td>3</td><td>if from non-null/non-empty: {@code from + ": " + msg}; else {@code msg}</td></tr>
     *   <tr><td>4 (~4 == -5)</td><td>if from non-null/non-empty: {@code from + ": " + msg}; else {@code msg}</td></tr>
     *   <tr><td>5 (~5 == -6)</td><td>always returns {@code msg} unchanged</td></tr>
     *   <tr><td>6</td><td>returns {@code from + " wishes to trade with you."}</td></tr>
     *   <tr><td>7 (~7 == -8)</td><td>if from non-null/non-empty: {@code from + ": " + msg}; else {@code msg}</td></tr>
     *   <tr><td>&gt;= 8</td><td>returns {@code ""} (empty string)</td></tr>
     * </table>
     *
     * <p>The {@code rightToLeft} flag, when {@code false}, sets {@link #chatLineState} to 90
     * (likely a RTL layout indicator for the UI renderer). When {@code true} it is ignored.
     *
     * obf: mb.a(String, String, boolean, int)
     *
     * @param msg         the main message body (displayed after any prefix)
     * @param from        the sender/prefix string (may be null or empty)
     * @param rightToLeft if false, sets chatLineState = 90 (RTL display mode)
     * @param mode        display mode selector (see table above)
     * @return assembled chat string
     */
    static final String formatChatLine(String msg, String from, boolean rightToLeft, int mode) {
        // Dead profiling counter (++mb.d) removed.

        // RTL layout flag
        if (!rightToLeft) {
            chatLineState = 90;
        }

        // Dispatch on mode; opaque-predicate dead branches stripped.
        //
        // The bytecode is an explicit if-else chain, not a tableswitch.
        // XOR trick annotations: ~mode == -5  =>  mode == 4
        //                        ~mode == -6  =>  mode == 5
        //                        ~mode == -8  =>  mode == 7
        //
        // Mode 0 falls to the bottom of the chain (the "default" path).

        if (mode == 1) {
            // Incoming private tell: "from tells you: msg"
            // Decoded z[6] = " tells you: "
            if (from == null || from.length() == 0) {
                return msg;
            }
            return from + " tells you: " + msg;
        }

        if (mode == 2) {
            // Outgoing private tell: "You tell from: msg"
            // Decoded z[5] = "You tell ", z[7] = ": "
            if (from == null || from.length() == 0) {
                return msg;
            }
            return "You tell " + from + ": " + msg;
        }

        if (mode == 3) {
            // Incoming channel variant: "from: msg"
            if (from == null || from.length() == 0) {
                return msg;
            }
            return from + ": " + msg;
        }

        if (mode == 4) {  // ~mode == -5
            // Clan/quest tell variant: "from: msg"
            if (from == null || from.length() == 0) {
                return msg;
            }
            return from + ": " + msg;
        }

        if (mode == 5) {  // ~mode == -6
            // Message-only (no prefix): return msg unchanged
            return msg;
        }

        if (mode == 6) {
            // Trade request: "from wishes to trade with you."
            // Decoded z[8] = " wishes to trade with you."
            return from + " wishes to trade with you.";
        }

        if (mode == 7) {  // ~mode == -8
            // Another "from: msg" variant
            if (from == null || from.length() == 0) {
                return msg;
            }
            return from + ": " + msg;
        }

        // mode == 0 (default): "from: msg" or just "msg"
        // This path also handles mode < 0 or mode > 7 if the caller misbehaves.
        // Bytecode: modes that do NOT match 1-7 fall to here OR to the empty-string path.
        // Mode 0 explicitly goes to the bottom-of-chain path at offset 0x0aa which returns
        // "from: msg"; mode >= 8 falls to offset 0x1f5 which returns "".
        if (mode == 0) {
            // Default/system message: "from: msg" or just "msg"
            // Decoded z[7] = ": "
            if (from == null || from.length() == 0) {
                return msg;
            }
            return from + ": " + msg;
        }

        // mode >= 8 (or any other unrecognised value): return empty string
        // Bytecode offset 0x1f5: ldc "" / areturn
        return "";
    }

    /**
     * Reports a client-side error to the Jagex error-reporting endpoint
     * ({@code clienterror.ws}).
     *
     * <p>Steps:
     * <ol>
     *   <li>If a {@link Throwable} is provided, its stack trace is captured via
     *       {@link ProxySocketFactory#stackTraceToString(boolean, Throwable)}
     *       (obf: {@code gb.a(false, throwable)}).
     *   <li>An optional context string is appended after {@code " | "} separator.
     *   <li>The combined string is logged to stdout via
     *       {@link FontWidths#println(byte, String)} (obf: {@code n.a(-93, text)}).
     *   <li>Special characters are URL-encoded by
     *       {@link DownloadWorker#replaceAll(boolean, String, String, String)}
     *       (obf: {@code jb.a(true, encoded, original, text)}):
     *       {@code ":"} → {@code "%3a"}, {@code "@"} → {@code "%40"},
     *       {@code "&"} → {@code "%26"}, {@code "#"} → {@code "%23"}.
     *   <li>A GET request is made to:
     *       <pre>
     *         &lt;codebase&gt;/clienterror.ws?c=&lt;revision&gt;&amp;u=&lt;username&gt;&amp;v1=&lt;world&gt;&amp;v2=&lt;clientRev&gt;&amp;e=&lt;encodedError&gt;
     *       </pre>
     *       using {@link ImageLoader#b}.{@link ListNode#open(byte, URL)}
     *       (obf: {@code pa.b.a(74, url)}).
     *   <li>The method polls the resulting {@link ListNode} ({@code g}) waiting for the HTTP
     *       response ({@code ~g.b == -1} means not yet done; {@code ~g.b == -2} means success).
     *   <li>On success ({@code ~g2.b == -2}), reads and closes the response stream.
     * </ol>
     *
     * <p>Any exceptions during reporting are silently swallowed (the outer {@code catch(Exception)}
     * does nothing) — errors in error-reporting must not crash the client.
     *
     * <p>The {@code guardParam} value {@code 2097151} ({@code == BITMASK[21]}) is the
     * anti-tamper sentinel that gates a dead {@link #formatChatLine} call; ignored here.
     *
     * obf: mb.a(int, Throwable, String)
     *
     * @param guardParam  anti-tamper sentinel (always 2097151 in live code)
     * @param throwable   exception to report (may be null)
     * @param context     additional context string appended after stack trace (may be null)
     */
    static final void reportError(int guardParam, Throwable throwable, String context) {
        // Dead profiling counter (++mb.e) removed.
        // Anti-tamper dead path: if (guardParam != 2097151) { formatChatLine(null, null, true, 27); }
        //   — never fires; stripped.

        try {
            String errorText = "";

            // Step 1: capture stack trace from the throwable (if provided).
            // obf: gb.a(false, throwable)  →  ProxySocketFactory.stackTraceToString(false, throwable)
            // ProxySocketFactory.stackTraceToString walks up ClientRuntimeException.cause
            // chains and calls printStackTrace() into a StringWriter.
            if (throwable != null) {
                errorText = ProxySocketFactory.stackTraceToString(false, throwable);
            }

            // Step 2: append optional context string after " | " separator.
            // The " | " is only inserted when we also have a stack trace (throwable != null).
            // Decoded z[10] = " | "
            if (context != null) {
                if (throwable != null) {
                    errorText = errorText + " | ";
                }
                errorText = errorText + context;
            }

            // Step 3: echo the raw error to stdout for local debugging.
            // obf: n.a((byte)-93, text)  →  FontWidths.println((byte)-93, text)
            // FontWidths.println just wraps System.out.println with a newline-replace.
            FontWidths.println((byte) -93, errorText);

            // Step 4: URL-encode characters that would break the query string.
            // obf: jb.a(true, replacement, original, text)  →  DownloadWorker.replaceAll(...)
            // DownloadWorker.replaceAll replaces all occurrences of `original` with `replacement`.
            // Decoded strings: z[15]="%3a"(:), z[12]="%40"(@), z[16]="%26"(&), z[13]="%23"(#)
            errorText = DownloadWorker.replaceAll(true, "%3a", ":", errorText);
            errorText = DownloadWorker.replaceAll(true, "%40", "@", errorText);
            errorText = DownloadWorker.replaceAll(true, "%26", "&", errorText);
            errorText = DownloadWorker.replaceAll(true, "%23", "#", errorText);

            // Step 5: bail out if not running inside an applet (no codebase available).
            // obf: l.b  →  Globals.applet  (java.applet.Applet)
            if (null == Globals.applet) {
                return;
            }

            // Step 6: determine the username to embed in the URL.
            //   Use IntHolder.a (a static username String set by the login flow) if available;
            //   otherwise fall back to ImageLoader.h (the numeric user ID).
            // obf: ka.a  →  IntHolder.a   (String; the logged-in username, or null)
            // obf: pa.h  →  ImageLoader.h (int; numeric user/player ID)
            String usernameParam;
            if (IntHolder.a != null) {
                usernameParam = IntHolder.a;
            } else {
                usernameParam = "" + ImageLoader.h;
            }

            // Step 7: build the full error-report URL and open an HTTP GET.
            //
            // URL template (decoded from XOR pool):
            //   <codebase>/clienterror.ws?c=<revision>&u=<username>&v1=<world>&v2=<clientRev>&e=<errorText>
            //
            // obf: db.d  →  LinkedQueue.d   (int; client revision / content CRC version)
            // obf: c.q   →  LoaderThread.q  (String; world/server identifier)
            // obf: c.k   →  LoaderThread.k  (String; client version string)
            // Decoded: z[11]="clienterror.ws?c=", z[19]="&u=", z[17]="&v1=",
            //           z[18]="&v2=", z[14]="&e="
            URL codeBase = Globals.applet.getCodeBase();
            URL reportUrl = new URL(
                codeBase,
                "clienterror.ws?c=" + LinkedQueue.d
                + "&u=" + usernameParam
                + "&v1=" + LoaderThread.q
                + "&v2=" + LoaderThread.k
                + "&e=" + errorText
            );

            // Open the URL via ImageLoader's loader-thread connection manager.
            // obf: pa.b        →  ImageLoader.b   (LoaderThread c; the active loader thread)
            // obf: pa.b.a(74, url) opens a non-blocking GET; returns a ListNode (g) tracking state
            // obf: g.b (volatile int): 0 = pending, 1 = success data ready, ~g.b==-1 means pending,
            //                          ~g.b==-2 means complete.
            // obf: g.d (Object) → the DataInputStream of the HTTP response body
            ListNode requestNode = ImageLoader.b.openUrl((byte) 74, reportUrl);

            // Step 8: busy-wait until the HTTP response arrives (non-blocking GET).
            // ~requestNode.b == -1  <=>  requestNode.b == 0  (still pending)
            while (~requestNode.b == -1) {
                sleepWithProfile(11200, 1L); // sleep 1ms per poll tick
                // Opaque predicate dead branches removed
            }

            // Step 9: on success, drain (read one byte) and close the response stream.
            // ~requestNode.b == -2  <=>  requestNode.b == 1  (completed successfully)
            if (~requestNode.b == -2) {
                DataInputStream responseStream = (DataInputStream) requestNode.d;
                responseStream.read();
                responseStream.close();
            }

        } catch (Exception ignored) {
            // Errors during error-reporting are silently swallowed — we must not enter a
            // crash loop if the reporting endpoint itself throws.
        }
    }

    /**
     * Computes the CRC-32 of a byte array region, delegating to
     * {@link ChatCipher#crc32(int, int, byte[], int)} (obf: {@code w.a(offset, -49, data, 0)}).
     *
     * <p>The {@code guardZero} parameter must be {@code 0} for the CRC to be computed; any
     * other value causes an early return of {@code 6} (anti-tamper sentinel — never fires in
     * live code).
     *
     * obf: mb.a(byte[], int, int)
     *
     * @param data      byte array to checksum
     * @param offset    starting offset within {@code data}
     * @param guardZero must be 0; any other value returns 6 (anti-tamper)
     * @return CRC-32 of {@code data[offset..data.length-1]}
     */
    static final int computeCrc32(byte[] data, int offset, int guardZero) {
        if (guardZero != 0) {
            // Anti-tamper dead path — never reached in live client
            return 6;
        }
        // Dead profiling counter (++mb.c) removed.
        // obf: w.a(offset, -49, data, 0)  →  WorldEntity.crc32(offset, -49, data, 0)
        // WorldEntity (obf: w) hosts a static CRC-32 helper used by the archive/cache layer.
        // param1 = -49: the anti-tamper guard in w.a checks param1 != -1 → returns 71 early,
        // but -49 != -1 so execution continues to the actual CRC table loop over wb.q[].
        return WorldEntity.a(offset, -49, data, 0);
    }

    // -----------------------------------------------------------------------
    // XOR string-pool helpers (obfuscation infrastructure — kept for
    // completeness but not used post-decode; normally these would be private
    // static and called only at class-load time from the static initializer).
    // -----------------------------------------------------------------------

    /**
     * First pass of the two-stage XOR string decoder.
     * If the input string has fewer than 2 characters, XORs the single character with {@code 0x36}
     * ('6'). Strings of length ≥ 2 are returned as-is (as a char array).
     *
     * <p>In the obfuscated build this is called as the inner {@code z(...)} in {@code z(z("..."))}.
     *
     * obf: mb.z(String)
     *
     * @param encoded XOR-encoded string literal from the constant pool
     * @return intermediate char array for the second pass
     */
    private static char[] z(String encoded) {
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            // Single-char key: XOR with 0x36 ('6')
            chars[0] = (char) (chars[0] ^ '6');
        }
        return chars;
    }

    /**
     * Second pass of the two-stage XOR string decoder.
     * XORs each character by a key selected from the 5-element table
     * {@code {56, 94, 102, 21, 54}} based on {@code position % 5}.
     *
     * <p>In the obfuscated build this is called as the outer {@code z(...)} in {@code z(z("..."))}.
     * The result is interned so repeated decodes share the same String instance.
     *
     * obf: mb.z(char[])
     *
     * @param chars intermediate char array from the first pass
     * @return the decoded, interned String
     */
    private static String z(char[] chars) {
        // key table indexed by position % 5
        // positions: 0→56('8'), 1→94('^'), 2→102('f'), 3→21(0x15), 4→54('6')
        final int[] KEY = {56, 94, 102, 21, 54};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char) (chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
