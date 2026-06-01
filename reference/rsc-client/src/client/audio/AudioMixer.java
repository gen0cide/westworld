package client.audio;

import client.shell.LoaderThread;
import client.util.ErrorHandler;
import client.util.ClientRuntimeException;
import client.util.Utility;
import client.net.ChatCipher;

/**
 * AudioMixer (obf: eb) — the audio mixing thread.
 *
 * Holds a fixed-size array of {@link AudioChannel} voices (capacity 2) and
 * implements {@link Runnable} so it can run on a dedicated daemon Thread
 * managed by {@link LoaderThread} (obf: c).  On each iteration of its tight
 * loop it calls {@link AudioChannel#tick()} on every non-null voice, then
 * sleeps for ~11 200 ms worth of timing budget via
 * {@link Utility#sleep(int, long)}, and finally posts a completion event
 * back to the AWT event queue via {@link SurfaceSprite#notifyAwtQueue}.
 *
 * Lifecycle:
 *   1. {@link AudioChannel#createChannel} creates an {@code AudioMixer}
 *      (singleton, stored as {@code AudioChannel.mixer}) when the first
 *      channel is opened, if none exists yet.
 *   2. {@link LoaderThread#startThread} spawns the daemon Thread.
 *   3. {@link AudioChannel#close} signals {@code stopRequested = true}
 *      (and also sets {@code eb.a = true} on the singleton) to break the
 *      main loop, then busy-waits on {@code isRunning} until the thread
 *      exits.
 *
 * Decoded XOR strings (eb's per-class string pool, keys [61,57,104,12,8]):
 *   z(z("qp>I"))       → "LIVE"          — ErrorHandler tag for e field
 *   z(z("jm!"))        → "WTI"           — ChatCipher constructor arg0
 *   z(z("R_\x0eekX"))  → "office"        — ChatCipher constructor arg1
 *   z(z("bN\x1ce"))    → "_wti"          — ChatCipher constructor arg2
 *   z(z("X[F~}S\x11A")) → "eb.run()"    — obfuscation exception site tag
 *
 * Oracle cross-reference: mudclient204 does not have an equivalent threading
 * wrapper around audio channels (it uses a single StreamAudioPlayer); the
 * structure here is unique to the rev-233+ DirectSound / SourceDataLine build.
 */
final class AudioMixer implements Runnable {

    // -------------------------------------------------------------------------
    // Instance state
    // -------------------------------------------------------------------------

    /**
     * Set to {@code true} by {@link AudioChannel#close} to signal the loop
     * to stop on the next iteration (after completing any in-flight tick).
     * obf: a
     */
    volatile boolean stopRequested = false;   // obf: a

    /**
     * The two audio voice slots.  Slot 0 and slot 1 correspond to the two
     * possible concurrent {@link AudioChannel} instances.  A slot is
     * {@code null} when no channel is assigned.
     * obf: f
     */
    volatile AudioChannel[] voices = new AudioChannel[2];   // obf: f (was sa[])

    /**
     * Back-reference to the {@link LoaderThread} that owns this mixer thread;
     * used by {@link SurfaceSprite#notifyAwtQueue} to post AWT events after
     * each audio tick cycle.
     * obf: g  (type c → LoaderThread)
     */
    LoaderThread loaderThread;   // obf: g  (type c)

    /**
     * Flag set to {@code true} while the {@link #run()} body is executing,
     * and cleared in the finally block.  {@link AudioChannel#close} spins on
     * this to know when it is safe to null out the singleton.
     * obf: i
     */
    volatile boolean isRunning = false;   // obf: i

    // -------------------------------------------------------------------------
    // Static state
    // -------------------------------------------------------------------------

    /**
     * Dead profiling counter incremented once per {@link #run()} invocation.
     * Never read back; left as a no-op artifact.
     * obf: h
     */
    static int _profileCounter;   // obf: h  (dead profiling counter, never read)

    /**
     * Static {@link ErrorHandler} instance tagged "LIVE"
     * (decoded: z(z("qp>I")) → "LIVE").
     * Used only by the per-class exception wrapper; not part of the mixer's
     * logical state.
     * obf: e  (type i → ErrorHandler)
     */
    // NOTE: obf name 'e' here is a *field*, not the class 'e' (GameShell).
    // NAMING.md caveat: the counter field letter may shadow a same-letter class
    // reference — confirmed distinct: this is field eb.e of type i (ErrorHandler).
    static ErrorHandler errorHandlerTag =
        new ErrorHandler("LIVE" /* z(z("qp>I")) */, 0);   // obf: e

    /**
     * A {@link ChatCipher} instance constructed with decoded string args
     * "WTI", "office", "_wti" and integer 5.
     * (decoded: z(z("jm!"))→"WTI", z(z("R_\x0eekX"))→"office",
     *           z(z("bN\x1ce"))→"_wti")
     * This appears to be an anti-tamper cipher used by the obfuscation layer
     * to validate strings; it is not called from any live audio logic.
     * obf: d  (type v → ChatCipher)
     */
    static ChatCipher tamperCipher =
        new ChatCipher("WTI" /* z(z("jm!")) */,
                        "office" /* z(z("R_\x0eekX")) */,
                        "_wti" /* z(z("bN\x1ce")) */,
                        5);   // obf: d

    /**
     * Spare int-array scratch buffer; allocated but not used in the mixer
     * itself — set externally by channel management code.
     * obf: b
     */
    static int[] scratchBuffer;   // obf: b

    // -------------------------------------------------------------------------
    // Per-class XOR string pool (obfuscation infrastructure, not audio logic)
    // -------------------------------------------------------------------------

    /**
     * Identity character decode table (256 chars, indices 0-255).
     * Characters 33,34,42-47,59,61,92,124 are re-assigned to their own
     * ASCII values (no-op; the static block below is purely defensive).
     * obf: c  (char[])
     */
    private static char[] charDecodeTable = new char[256];   // obf: c

    /**
     * Decoded exception-site tag string: "eb.run()" (z(z("X[F~}S\x11A"))).
     * Passed to {@link ErrorHandler#wrap} in the outermost catch to annotate
     * stack traces with the originating method signature.
     * obf: z
     */
    private static final String EXCEPTION_SITE_TAG =
        decodeXorString(decodeXorChars("X[F~}SA"));   // obf: z  → "eb.run()"

    // -------------------------------------------------------------------------
    // Static initialiser — char-table setup
    // -------------------------------------------------------------------------

    static {
        // Fill identity table: charDecodeTable[i] == (char)i for all i.
        for (int i = 0; i < 256; i++) {
            charDecodeTable[i] = (char) i;
        }
        // Redundant re-assignments (each value equals its index; no-op).
        charDecodeTable[45]  = '-';
        charDecodeTable[59]  = ';';
        charDecodeTable[42]  = '*';
        charDecodeTable[124] = '|';
        charDecodeTable[43]  = '+';
        charDecodeTable[33]  = '!';
        charDecodeTable[34]  = '"';
        charDecodeTable[47]  = '/';
        charDecodeTable[46]  = '.';
        charDecodeTable[61]  = '=';
        charDecodeTable[92]  = '\\';
        charDecodeTable[44]  = ',';
    }

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /** Default constructor — both flags start false. obf: eb() */
    AudioMixer() {
        // stopRequested and isRunning default to false via field initializers.
    }

    // -------------------------------------------------------------------------
    // Runnable implementation
    // -------------------------------------------------------------------------

    /**
     * Main mixer loop.  Runs on the daemon thread started by
     * {@link LoaderThread}.
     *
     * Bytecode-accurate structure (after stripping obfuscation):
     * <pre>
     *   isRunning = true
     *   try {
     *     while (true) {
     *       if (stopRequested) { isRunning = false; return; }  // bc offset 0x63
     *       for (int i = 0; i &lt; 2; i++) {
     *           AudioChannel ch = voices[i];
     *           if (ch != null) ch.tick();
     *       }
     *       Utility.sleep(11200, 10L);
     *       SurfaceSprite.notifyAwtQueue(null, 1, loaderThread);
     *     }
     *   } catch (Exception e) {
     *       Utility.reportError(0x1FFFFF, e, null);  // swallowed, isRunning=false in catch
     *   } finally {
     *       isRunning = false;   // bc offset 0x7e (unwind path)
     *   }
     * </pre>
     *
     * Obfuscation stripped:
     *   - {@code boolean bl = client.vh} opaque-predicate guard removed (always false).
     *   - {@code ++eb.h} profiling counter removed.
     *   - Outermost {@code catch (RuntimeException) { throw i.a(e, z); }} unwrapped.
     *   - Dead {@code if (bl)} branches and {@code do...while (!bl)} retries removed.
     *
     * Note on {@code isRunning} clearing: the bytecode clears it in three places —
     * at the {@code stopRequested} exit (offset 0x63), inside the exception catch
     * (offset 0x75), and in the unwind/finally path (offset 0x7e).  The try-finally
     * below captures all three; the early-return form preserves the bytecode intent.
     *
     * obf: run()
     */
    @Override
    public final void run() {
        // obf: boolean bl = client.vh;  — always false, removed (opaque predicate)
        // obf: ++eb.h;                  — dead profiling counter (field h), removed

        isRunning = true;   // obf: this.i = true  (bc 05-07)
        try {
            // Main mixing loop — exits only when stopRequested is set.
            while (true) {
                // Exit condition: AudioChannel.close() has signalled us to stop.
                // Bytecode: getfield eb.a Z; ifne 63 → if (a) { i=false; return }
                if (stopRequested) {                    // obf: if (this.a)  (bc 12-16)
                    isRunning = false;                  // bc offset 0x63-0x65
                    return;
                }

                // Tick each active voice slot.
                // Bytecode loop condition: iload n2; ixor -1; bipush -3; if_icmple 4c
                //   ⟺  ~n2 > -3  ⟺  n2 < 2
                for (int i = 0; i < 2; i++) {          // obf: n2  (bc 1e-49)
                    AudioChannel ch = voices[i];        // obf: sa2 = this.f[n2]  (bc 28-2e)
                    if (ch != null) {
                        ch.tick();                      // obf: sa2.a()  (bc 3c-3d)
                    }
                }

                // Sleep for one audio frame budget.
                // 11200 is the sentinel tag Utility.sleep uses to identify
                // "audio mixer" sleeps (vs. other callers of mb.a).
                Utility.sleep(11200, 10L);              // obf: mb.a(11200, 10L)  (bc 4c-52)

                // Notify the AWT event queue that one audio cycle has completed.
                // null = no source object; 1 = completion event type code.
                SurfaceSprite.notifyAwtQueue(null, 1, loaderThread);
                // obf: ba.a(null, 1, this.g)  (bc 55-5b)
            }
        } catch (Exception e) {
            // Non-fatal audio failure: report via Utility (0x1FFFFF suppresses
            // the clienterror.ws network upload — logging only).
            // obf: mb.a(0x1FFFFF, exception, null)  (bc 6c-70)
            Utility.reportError(0x1FFFFF, e, null);
            // isRunning cleared by finally below (mirrors bc offset 0x75).
        } finally {
            // Always clear isRunning so AudioChannel.close() can proceed past
            // its busy-wait on this flag.
            isRunning = false;                          // obf: this.i = false  (bc 7c-7e)
        }
        // Outer RuntimeException wrapper (unwrapped — not live logic):
        //   obf: catch (RuntimeException runtimeException) {
        //            throw i.a(runtimeException, z);
        //            // i = ErrorHandler, z = "eb.run()" (decoded EXCEPTION_SITE_TAG)
        //        }
    }

    // -------------------------------------------------------------------------
    // XOR string pool helpers (obfuscation infrastructure — not audio logic)
    // -------------------------------------------------------------------------

    /**
     * First XOR pass: converts an encoded String literal to a char[].
     * For strings shorter than 2 characters the single char is XOR'd with 8;
     * longer strings are returned as-is (the real XOR work is done by the
     * second pass).
     * obf: z(String) → char[]
     */
    private static char[] decodeXorChars(String encoded) {   // obf: z(String)
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 8);   // single-char key = 8
        }
        return chars;
    }

    /**
     * Second XOR pass: decodes a char[] produced by {@link #decodeXorChars}
     * using a 5-element rotating byte key [61, 57, 104, 12, 8] and interns
     * the result.
     * obf: z(char[]) → String
     */
    private static String decodeXorString(char[] chars) {   // obf: z(char[])
        // Rotating key table (position mod 5):
        //   0 → 61 ('='),  1 → 57 ('9'),  2 → 104 ('h'),  3 → 12 ('\f'),  4 → 8 ('\b')
        int[] XOR_KEYS = {61, 57, 104, 12, 8};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char) (chars[i] ^ XOR_KEYS[i % 5]);
        }
        return new String(chars).intern();
    }
}
