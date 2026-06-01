package client.net;

import java.io.DataInputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.net.Socket;
import java.net.URL;

/**
 * DownloadWorker — asynchronous single-asset HTTP/JAGGRAB resource downloader (obf: jb).
 *
 * <p>Each instance downloads one asset identified by a {@link URL} into a pre-allocated
 * {@link Buffer}.  Two transport modes are attempted in order:
 * <ol>
 *   <li><b>HTTP via the applet class-loader</b> — LoaderThread enqueues a type-4 URL-fetch
 *       job; when complete, {@code httpJob.data} is a {@link DataInputStream}.</li>
 *   <li><b>Raw TCP "JAGGRAB" fallback</b> — if HTTP fails, a plain socket is opened to
 *       {@code resourceUrl.getHost():443} and the proprietary JAGGRAB request is sent:
 *       {@code 0x11} handshake byte followed by {@code "JAGGRAB /path\n\n"} encoded by
 *       {@link TextEncoder}.</li>
 * </ol>
 *
 * <p>State machine (field {@link #downloadState}):
 * <pre>
 *   0  – initial; no connection attempted
 *   1  – HTTP attempt failed; escalating to JAGGRAB
 *   2  – JAGGRAB also failed; unrecoverable error
 *   3  – download complete; buffer is ready
 * </pre>
 *
 * <p>The caller polls {@link #tick(int)} each game tick and retrieves the filled buffer
 * via {@link #getBuffer(byte)} once the state reaches 3.
 *
 * <p>This class also contains two static helpers injected by the obfuscator that have
 * nothing to do with downloading: {@link #drawTexturedSpanUnrolled} (affine texture
 * mapper used by the 3D renderer) and {@link #replaceAll} (String substitution utility).
 *
 * Obfuscated class name: {@code jb}
 */
final class DownloadWorker implements Runnable {

    // -------------------------------------------------------------------------
    // Dead profiling counters (static; incremented on every method entry;
    // never read by game logic — pure obfuscation overhead)
    // -------------------------------------------------------------------------

    /** Dead profiling counter for run(). obf: jb.m */
    static int _profRun;           // obf: m
    /** Dead profiling counter for tick(int). obf: jb.e */
    static int _profTick;          // obf: e
    /** Dead profiling counter for drawTexturedSpanUnrolled. obf: jb.d */
    static int _profDrawSpan;      // obf: d
    /** Dead profiling counter for replaceAll. obf: jb.j */
    static int _profReplaceAll;    // obf: j
    /**
     * Dead profiling counter for getBuffer(byte).
     * NOTE: obfuscated name {@code i} — same letter as the {@code ErrorHandler} class (also "i").
     * This is a static int field, not a reference to ErrorHandler; the obfuscator intentionally
     * reuses single-letter identifiers across name spaces to confuse decompilers.
     * obf: jb.i
     */
    static int _profGetBuffer;     // obf: i  (field, NOT class i = ErrorHandler)
    /** Dead profiling counter for cleanup(). obf: jb.n */
    static int _profCleanup;       // obf: n

    // -------------------------------------------------------------------------
    // Other static fields
    // -------------------------------------------------------------------------

    /**
     * Unused int array — present in original bytecode; never written by live methods.
     * May be accessed from another class via static reference.
     * obf: jb.k
     */
    static int[] unusedIntArray;   // obf: k

    /**
     * Set to 21 inside {@link #drawTexturedSpanUnrolled} when {@code halfPixelMode}
     * is true.  Acts as a shared flag between the renderer and its caller.
     * obf: jb.p  (originally declared as {@code static int o = 0})
     */
    static int halfPixelModeFlag = 0; // obf: p  (was "o" in source — renamed to avoid confusion with `o` = ISAAC class)

    // -------------------------------------------------------------------------
    // Instance fields
    // -------------------------------------------------------------------------

    /**
     * The applet's loader / thread-manager (class c = LoaderThread).
     * Used to enqueue HTTP fetch jobs, TCP socket connects, and reader threads.
     * obf: jb.b
     */
    private LoaderThread loaderThread; // obf: b  (type c = LoaderThread)

    /**
     * URL of the resource to download.
     * obf: jb.g
     */
    private URL resourceUrl; // obf: g

    /**
     * Pending HTTP fetch job node.  Returned by {@code loaderThread.fetchUrl(byte=74, url)}.
     * {@code httpJob.status} values:
     * <ul>
     *   <li>0  — pending</li>
     *   <li>1  — complete; {@code httpJob.data} is a {@link DataInputStream}</li>
     *   <li>other — failed</li>
     * </ul>
     * obf: jb.c  (type g = ListNode)
     */
    private ListNode httpJob; // obf: c

    /**
     * Pending JAGGRAB TCP socket job node.  Returned by
     * {@code loaderThread.openSocket(host, 443, -68)}.
     * {@code tcpJob.status} values:
     * <ul>
     *   <li>-1 — still connecting</li>
     *   <li>1  — connected; {@code tcpJob.data} is a {@link Socket}</li>
     *   <li>other — failed</li>
     * </ul>
     * obf: jb.f  (type g = ListNode)
     */
    private ListNode tcpJob; // obf: f

    /**
     * Job node for the background reader thread that executes {@link #run()}.
     * Returned by {@code loaderThread.startThread(true, this, 5)}.
     * {@code readerThreadJob.status == -1} while the thread is still running.
     * obf: jb.l  (type g = ListNode)
     */
    private ListNode readerThreadJob; // obf: l

    /**
     * The stream from which asset bytes are read into {@link #buffer}.
     * Sourced either from the HTTP job ({@code httpJob.data}) or constructed
     * from a raw JAGGRAB socket's InputStream.
     * obf: jb.q
     */
    private DataInputStream inputStream; // obf: q

    /**
     * Destination byte buffer; allocated to the expected file size in the constructor.
     * Filled completely by the reader thread.
     * obf: jb.h  (type tb = Buffer)
     */
    private Buffer buffer; // obf: h

    /**
     * Current download state; see class-level javadoc for the state machine.
     * obf: jb.a
     */
    private int downloadState; // obf: a

    // -------------------------------------------------------------------------
    // Decoded XOR string pool  (obf: jb.z — static final String[])
    // Two-pass decode:
    //   pass 1: z(String) → char[] — XORs single-char strings with 0x2F; longer strings are toCharArray()
    //   pass 2: z(char[]) → String — XORs each char with rotating key [39,56,26,30,47][i%5]
    //
    //   z[0]  = "HG1: "           — prefix for premature-EOF error message in run()
    //   z[1]  = "jb.run()"        — error-handler context string for run()
    //   z[2]  = "JAGGRAB "        — JAGGRAB request verb
    //   z[3]  = "jb.B("           — error-handler context for tick(int)
    //   z[4]  = "\n\n"            — JAGGRAB request line terminator
    //   z[5]  = "jb.D("           — error-handler context for drawTexturedSpanUnrolled
    //   z[6]  = "{...}"           — non-null placeholder in diagnostic strings
    //   z[7]  = "null"            — null placeholder in diagnostic strings
    //   z[8]  = "jb.<init>("      — error-handler context for constructor
    //   z[9]  = "jb.finalize()"   — error-handler context for cleanup()
    //   z[10] = "jb.C("           — error-handler context for getBuffer(byte)
    // -------------------------------------------------------------------------

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /**
     * Creates a new DownloadWorker that will fill a buffer of {@code bufferSize} bytes
     * by fetching from {@code resourceUrl}.
     *
     * @param loaderThread the applet's loader/thread manager
     * @param resourceUrl  URL of the asset to fetch
     * @param bufferSize   exact expected size of the asset in bytes;
     *                     a {@link Buffer} of this size is pre-allocated
     *
     * obf: jb(c, URL, int)
     */
    DownloadWorker(LoaderThread loaderThread, URL resourceUrl, int bufferSize) {
        this.resourceUrl   = resourceUrl;
        this.loaderThread  = loaderThread;
        this.buffer        = new Buffer(bufferSize); // tb(int) — allocates byte[bufferSize]
    }

    // -------------------------------------------------------------------------
    // Runnable — background read loop
    // -------------------------------------------------------------------------

    /**
     * Reads bytes from {@link #inputStream} into {@link #buffer} until the buffer
     * is completely filled, then marks the download complete.
     *
     * <p>Runs on a background thread managed by LoaderThread.  The loop calls
     * {@link DataInputStream#read(byte[], int, int)} repeatedly, accumulating bytes
     * via {@code buffer.offset} (the write cursor).  Receiving -1 (EOF) before the
     * buffer is full is treated as failure.
     *
     * <p>On success: {@code downloadState} is set to 3 (STATE_COMPLETE).
     * On any error: {@link #cleanup()} is called and {@code downloadState} is incremented.
     *
     * obf: jb.run()
     */
    @Override
    public final void run() {
        // NOTE on obfuscated bytecode semantics in the original run():
        //
        // The Vineflower/CFR decompilations both show a peculiar structure:
        //   while (w < len) { read(); if (bytesRead == -1) break; w += bytesRead; }
        //   if (w == len) throw new Exception("HG1: " + len + " " + url);
        //   synchronized { finalize(); a = 3; }
        //
        // Superficially this throws on FULL read and succeeds on partial/EOF — backwards.
        // However, the opaque-predicate machinery wraps a dead `if (bl) break;` inside
        // the success sub-block (bl is always false), so Vineflower's break label target
        // may be mis-identified.  The INTENDED semantics (consistent with the rest of the
        // class: a==3 = "buffer ready", a<2 = "in progress") are:
        //   - Full read (w reaches len) → a = 3 (STATE_COMPLETE, buffer is ready).
        //   - EOF before full (w < len) → throw → caught → a++ (transport failed, retry).
        //
        // We implement the INTENDED semantics here.  See uncertainty notes for details.
        try {
            // Fill buffer.data[0..length) from inputStream, accumulating via buffer.offset.
            while (buffer.offset < buffer.data.length) {
                int bytesRead = inputStream.read(
                        buffer.data,
                        buffer.offset,
                        buffer.data.length - buffer.offset);

                if (bytesRead == -1) {
                    // Premature EOF — stream ended before buffer was fully filled.
                    // z[0] = "HG1: " is the original error-message prefix.
                    throw new IOException("HG1: " + buffer.data.length + " " + resourceUrl);
                }

                buffer.offset += bytesRead;
            }

            // Buffer completely filled — download succeeded.
            synchronized (this) {
                cleanup();         // close streams; data is safely in buffer.data
                downloadState = 3; // STATE_COMPLETE
            }

        } catch (Exception ex) {
            // I/O error or premature EOF → record failure and close connections.
            synchronized (this) {
                cleanup();
                downloadState++; // 0→1 escalate to JAGGRAB, or 1→2 give up
            }
        }
    }

    // -------------------------------------------------------------------------
    // State-machine tick  (called from game main loop each tick)
    // -------------------------------------------------------------------------

    /**
     * Advances the download state machine one step and returns whether the
     * download has finished (successfully or with failure).
     *
     * <p>Must be called repeatedly from the game's main thread. Each invocation
     * is non-blocking: if a job is still pending it returns {@code false}
     * immediately and the caller tries again next tick.
     *
     * <p>State 0 — Attempt HTTP fetch via the applet class-loader:
     * <ol>
     *   <li>Lazily enqueue a URL-fetch job via {@code loaderThread.fetchUrl(74, url)}.</li>
     *   <li>Poll {@code httpJob.status}; 0 = still pending → return false.</li>
     *   <li>Status 1 = success; fall through to stream setup.</li>
     *   <li>Other = failure → null out job, increment state to 1, return false.</li>
     * </ol>
     *
     * <p>State 1 — JAGGRAB TCP fallback:
     * <ol>
     *   <li>Lazily enqueue a socket-connect job via {@code loaderThread.openSocket(host, 443)}.</li>
     *   <li>Status -1 = still connecting → return false.</li>
     *   <li>Status 1 = connected; fall through to stream setup.</li>
     *   <li>Other = failure → null out job, increment state to 2, return false.</li>
     * </ol>
     *
     * <p>Stream setup (both states):
     * <ul>
     *   <li>For HTTP: extract the {@link DataInputStream} from {@code httpJob.data}.</li>
     *   <li>For JAGGRAB: get the {@link Socket} from {@code tcpJob.data}, set a 10 s SO_TIMEOUT,
     *       send the JAGGRAB handshake ({@code 0x11} + encoded {@code "JAGGRAB /path\n\n"}),
     *       then wrap the socket's InputStream in a DataInputStream.</li>
     *   <li>Reset the buffer write cursor to 0.</li>
     *   <li>Start the reader thread via {@code loaderThread.startThread(true, this, 5)}.</li>
     * </ul>
     *
     * <p>Reader-thread polling:
     * <ul>
     *   <li>{@code readerThreadJob.status == -1} → thread still running → return false.</li>
     *   <li>{@code readerThreadJob.status == expectedThreadState} → thread finished cleanly
     *       (run() set downloadState=3 on success) → return false (caller checks isDone separately).</li>
     *   <li>Status mismatch → unexpected termination → cleanup and increment state.</li>
     * </ul>
     *
     * @param expectedThreadState the status code the reader thread should report on normal
     *                            completion (caller-supplied; typically the thread's assigned
     *                            priority or a sentinel)
     * @return {@code true} when the download has terminated (state ≥ 2, or state == 3);
     *         {@code false} while still in progress
     *
     * obf: jb.a(int)  — renamed: tick(int)
     */
    final synchronized boolean tick(int expectedThreadState) {
        // States 2 and 3: done (failed or complete).
        if (downloadState >= 2) {
            return true;
        }

        // ---- State 0: HTTP fetch via applet class-loader ----
        if (downloadState == 0) {
            if (httpJob == null) {
                // Enqueue type-4 URL fetch job.  Magic byte 74 is the job-type discriminator
                // in LoaderThread's dispatcher (c.a(byte=74, URL)).
                httpJob = loaderThread.fetchUrl((byte) 74, resourceUrl); // obf: b.a((byte)74, g)
            }

            if (httpJob.status == 0) {
                return false; // still pending
            }

            if (httpJob.status != 1) {
                // HTTP failed — escalate to JAGGRAB.
                httpJob = null;
                downloadState++; // 0 → 1
                return false;
            }
            // httpJob.status == 1: fetch complete; fall through to DataInputStream setup.
        }

        // ---- State 1: JAGGRAB TCP fallback ----
        if (downloadState == 1) {
            if (tcpJob == null) {
                // Enqueue a raw TCP connect to port 443.
                // The original call passes -68 as a third int param (anti-tamper dummy; dropped).
                tcpJob = loaderThread.openSocket(resourceUrl.getHost(), 443); // obf: b.a(String, 443, (int)-68)
            }

            if (tcpJob.status == -1) {
                return false; // still connecting
            }

            if (tcpJob.status != 1) {
                // TCP connection failed — both transports exhausted.
                tcpJob = null;
                downloadState++; // 1 → 2
                return false;
            }
            // tcpJob.status == 1: socket connected; fall through to stream setup.
        }

        // ---- Set up DataInputStream (once per attempt) ----
        if (inputStream == null) {
            try {
                if (downloadState == 0) {
                    // HTTP path: the DataInputStream was produced by the loader.
                    inputStream = (DataInputStream) httpJob.data; // obf: c.d
                }

                if (downloadState == 1) {
                    // JAGGRAB path: configure socket, send request, wrap InputStream.
                    Socket socket = (Socket) tcpJob.data; // obf: f.d
                    socket.setSoTimeout(10000); // 10-second read timeout

                    OutputStream out = socket.getOutputStream();

                    // JAGGRAB handshake:
                    //   Byte 0x11 (decimal 17) — proprietary connection-type identifier.
                    out.write(17); // 0x11 = JAGGRAB magic

                    // Request line: "JAGGRAB /filename.jag\n\n"
                    // z[2]="JAGGRAB ", z[4]="\n\n" — decoded from XOR pool above.
                    // TextEncoder.encode(CharSequence, byte=-104) produces a Latin-1 byte array.
                    out.write(TextEncoder.encode(
                            "JAGGRAB " + resourceUrl.getFile() + "\n\n",
                            (byte) -104)); // obf: h.a(CharSequence, (byte)-104)

                    inputStream = new DataInputStream(socket.getInputStream());
                }

                // Reset buffer write cursor to the beginning.
                buffer.offset = 0; // obf: h.w = 0

            } catch (IOException ex) {
                // Transport setup failed; close everything and advance.
                cleanup();
                downloadState++;
            }
        }

        // ---- Start the background reader thread (once) ----
        if (readerThreadJob == null) {
            // daemon=true, runnable=this, priority=5
            readerThreadJob = loaderThread.startThread(true, this, 5); // obf: b.a(true, Runnable, 5)
        }

        // Poll reader thread status.
        if (readerThreadJob.status == -1) {
            return false; // thread still running
        }

        // If the thread finished with a status code different from what the caller expects,
        // something went wrong (e.g. wrong server response).
        if (readerThreadJob.status != expectedThreadState) {
            cleanup();
            downloadState++;
        }

        return false;
    }

    // -------------------------------------------------------------------------
    // Buffer retrieval
    // -------------------------------------------------------------------------

    /**
     * Returns the completed {@link Buffer} if the download succeeded, or {@code null}
     * if still in progress or failed.
     *
     * <p>The {@code hint} parameter is a dummy byte injected by the obfuscator as an
     * anti-tamper guard.  When {@code hint > -110} the method runs the download loop
     * synchronously on the calling thread — this path is never taken in practice because
     * all callers pass a negative byte value.
     *
     * @param hint anti-tamper dummy; if &gt; -110 forces a synchronous run() call
     * @return the filled {@link Buffer} when {@code downloadState == 3}, else {@code null}
     *
     * obf: jb.a(byte)  — renamed: getBuffer(byte)
     */
    final Buffer getBuffer(byte hint) {
        if (hint > -110) {
            // Synchronous fallback path — dead in practice.
            run();
        }
        return (downloadState == 3) ? buffer : null;
    }

    // -------------------------------------------------------------------------
    // Cleanup  (called from run() and tick() on error or completion)
    // -------------------------------------------------------------------------

    /**
     * Closes all open connections and nulls out job handles.  Silently swallows
     * exceptions from individual {@code close()} calls, matching the original behavior.
     *
     * <p>Closes in order:
     * <ol>
     *   <li>The {@link DataInputStream} inside {@link #httpJob} (HTTP path).</li>
     *   <li>The {@link Socket} inside {@link #tcpJob} (JAGGRAB path).</li>
     *   <li>The primary {@link #inputStream}.</li>
     * </ol>
     * Also nulls out {@link #readerThreadJob}.
     *
     * <p>The original code overrides {@link Object#finalize()} for this logic — a known
     * anti-pattern.  We delegate from {@code finalize()} here and also call this directly
     * from {@link #run()} and {@link #tick(int)}.
     *
     * obf: jb.finalize()
     */
    @Override
    protected final void finalize() {
        cleanup();
    }

    /** Internal cleanup implementation. */
    private void cleanup() {
        // Close HTTP job's DataInputStream, if any.
        if (httpJob != null) {
            if (httpJob.data != null) {
                try {
                    ((DataInputStream) httpJob.data).close();
                } catch (Exception ignored) { }
            }
            httpJob = null;
        }

        // Close JAGGRAB socket, if any.
        if (tcpJob != null) {
            if (tcpJob.data != null) {
                try {
                    ((Socket) tcpJob.data).close();
                } catch (Exception ignored) { }
            }
            tcpJob = null;
        }

        // Close the primary DataInputStream.
        if (inputStream != null) {
            try {
                inputStream.close();
            } catch (Exception ignored) { }
            inputStream = null;
        }

        // Release the reader-thread job handle.
        readerThreadJob = null;
    }

    // -------------------------------------------------------------------------
    // Static helper: replaceAll  (injected by obfuscator — unrelated to downloading)
    // -------------------------------------------------------------------------

    /**
     * Replaces all occurrences of {@code target} within {@code source} with
     * {@code replacement}, equivalent to {@link String#replace(CharSequence, CharSequence)}.
     *
     * <p>This helper was placed in this class by the obfuscator (possibly for vtable sharing).
     * The {@code allowNoop} flag, when {@code false}, triggers a heavyweight call to
     * {@link #drawTexturedSpanUnrolled} with sentinel constants — a dead anti-tamper check
     * that is always bypassed in practice.
     *
     * @param allowNoop   when false, runs an anti-tamper validation call (dead code)
     * @param replacement string to insert in place of each match of {@code target}
     * @param target      substring to search for
     * @param source      input string
     * @return {@code source} with all occurrences of {@code target} replaced
     *
     * obf: jb.a(boolean, String, String, String)  — renamed: replaceAll
     */
    static final String replaceAll(boolean allowNoop, String replacement, String target, String source) {
        // Anti-tamper check (always bypassed because allowNoop is always true at call sites):
        // if (!allowNoop) { drawTexturedSpanUnrolled(null,78,-46,-87,-87,-58,-96,-121,50,-80,null,false,-54,-83,52); }

        int pos = source.indexOf(target);
        while (pos != -1) {
            source = source.substring(0, pos) + replacement + source.substring(target.length() + pos);
            pos = source.indexOf(target, pos + replacement.length());
        }
        return source;
    }

    // -------------------------------------------------------------------------
    // Static helper: drawTexturedSpanUnrolled  (injected by obfuscator)
    // -------------------------------------------------------------------------

    /**
     * Renders a horizontal pixel span using affine texture mapping, with a 16-pixel
     * unrolled inner loop and a residual loop for leftovers.  This is the core inner
     * loop of the software 3D rasterizer (normally lives in the renderer classes).
     *
     * <p>The method was <em>completely undecompilable by Vineflower</em>; this body is
     * reconstructed from the CFR decompilation and raw bytecode.
     *
     * <h4>Shift-constant obfuscation</h4>
     * The JVM {@code ishr}/{@code iushr} instructions use only the low 5 bits of the
     * shift amount, so large constants like {@code -496952415} and {@code 223739206}
     * collapse to their value {@code & 31}.  The constants are preserved in comments
     * for bytecode traceability.  Examples:
     * <ul>
     *   <li>{@code -496952415 & 31 = 1}  → effective right-shift of 1 (half-brightness blend)</li>
     *   <li>{@code 223739206  & 31 = 6}  → effective left-shift of 6  (fixed-point scale)</li>
     *   <li>{@code 955305670  & 31 = 6}  → effective right-shift of 6 (texture V coordinate)</li>
     * </ul>
     *
     * <h4>Fixed-point layout</h4>
     * <ul>
     *   <li>{@code texFrac} packs V index (bits 12–19, stride 4032) and U fraction (bits 0–11).</li>
     *   <li>{@code 4032 = 63 × 64} is the texture row width (6-bit texel index × 64 rows).</li>
     *   <li>{@code brightness} is the top 12 bits of texFrac used as a right-shift to dim texels.</li>
     * </ul>
     *
     * @param texels         source texture palette-mapped pixel array            (obf: param0)
     * @param texUStep       U numerator step added each sub-span                 (obf: param1)
     * @param texVStep       V denominator step added each sub-span               (obf: param2)
     * @param subdivLen      perspective sub-span length (divisor for slopes)     (obf: param3)
     * @param uStepScaled    U step pre-scaled (×4) for stride arithmetic         (obf: param4)
     * @param texFrac        packed U/V texture accumulator (16.12 fixed-point)   (obf: param5)
     * @param destIndex      current write index into destPixels                  (obf: param6)
     * @param spanLen        total pixel count to render                          (obf: param7)
     * @param texU           current U numerator                                  (obf: param8)
     * @param texV           current V accumulator (perspective denominator)      (obf: param9)
     * @param destPixels     destination framebuffer pixel array                  (obf: param10)
     * @param halfPixelMode  when true, sets {@link #halfPixelModeFlag} = 21      (obf: param11)
     * @param perspStep      perspective step added to texU each iteration        (obf: param12)
     * @param texUNumer      starting U numerator (for slope computation)         (obf: param13)
     * @param texVFrac       starting V fractional value                          (obf: param14)
     *
     * obf: jb.a(int[], int, int, int, int, int, int, int, int, int, int[], boolean, int, int, int)
     *      renamed: drawTexturedSpanUnrolled
     */
    static final void drawTexturedSpanUnrolled(
            int[]   texels,        // obf: param0
            int     texUStep,      // obf: param1
            int     texVStep,      // obf: param2
            int     subdivLen,     // obf: param3
            int     uStepScaled,   // obf: param4
            int     texFrac,       // obf: param5
            int     destIndex,     // obf: param6
            int     spanLen,       // obf: param7
            int     texU,          // obf: param8
            int     texV,          // obf: param9
            int[]   destPixels,    // obf: param10
            boolean halfPixelMode, // obf: param11
            int     perspStep,     // obf: param12
            int     texUNumer,     // obf: param13
            int     texVFrac       // obf: param14
    ) {
        // Anti-tamper sentinel: "if (~spanLen >= -1) return;" = "if (spanLen >= 0) return;"
        // In real calls spanLen is always positive, so this guard is never triggered.
        if (~spanLen >= -1) return;

        // ---- Initialise perspective-corrected U/V slopes ----
        // uSlope and vSlope are per-pixel U and V steps within the current sub-span
        // (computed by dividing the numerators by the sub-span length).
        // Effective left-shift 6 = ×64 — scales from 16.12 fixed-point to pixel stride.
        int uSlope = 0;
        int vSlope = 0;
        if (~subdivLen != -1) { // i.e. subdivLen != -1 (non-degenerate sub-span)
            uSlope = (texUNumer / subdivLen) << (223739206 & 31);  // eff <<6
            vSlope = (texU      / subdivLen) << (-902154106 & 31); // eff <<6
        }
        uStepScaled <<= 2; // pre-scale U step by 4 for stride arithmetic

        // Optional flag for caller post-processing.
        if (halfPixelMode) {
            halfPixelModeFlag = 21; // obf: jb.o = 21
        }

        // Clamp uSlope to [0, 4032] (valid texture-row index range).
        if (~uSlope > -1) { // uSlope < 0
            uSlope = 0;
        }
        if (4032 < uSlope) {
            uSlope = 4032;
        }

        // ---- Main span loop: 16 pixels per major iteration ----
        int remaining = spanLen;
        do {
            // Per-sub-span: advance perspective denominators.
            subdivLen += texVStep;
            int savedVSlope = uSlope;  // previous uSlope for interpolation delta
            texU      += perspStep;
            int savedUSlope = vSlope;  // previous vSlope
            texUNumer += texUStep;

            // Recompute slopes for this sub-span (one divide per sub-span = perspective correction).
            if (~subdivLen != -1) { // subdivLen != -1
                uSlope = (texU      / subdivLen) << (-213852602 & 31); // eff <<6
                vSlope = (texUNumer / subdivLen) << (-474023130 & 31); // eff <<6
            }

            // Clamp.
            if (~uSlope > -1) uSlope = 0;
            if (4032 < uSlope) uSlope = 4032;

            // Per-pixel increments for the 16-pixel block (linear interpolation between slopes).
            // Eff >>4 = divide by 16 (the unroll factor).
            int vInc = (vSlope - savedUSlope) >> (-1841585212 & 31); // eff >>4
            int uInc = (-savedVSlope + uSlope) >> (166246532 & 31);  // eff >>4

            // Brightness (dimming) factor derived from the upper bits of texFrac.
            // Used as an unsigned right-shift amount on texel values to darken them.
            int brightness = texFrac >> (-1249879148 & 31); // eff >>20

            // Advance texFrac's page component and step to the next sub-span.
            texFrac += texFrac & 0xC0000; // add current row stride (bits 18-19)
            texFrac += uStepScaled;

            if (remaining >= 16) {
                // ---- 16-pixel unrolled block ----
                // Each pixel blends the existing destination value with a texel from the texture.
                // StreamBase.blend(a, b) masks one value against a brightness mask from the other.
                // The large ishr constants reduce to their low 5 bits (effective shift shown in comment).

                // Pixel 1
                destPixels[destIndex++] =
                        StreamBase.blend(texels[destIndex] >> (-496952415 & 31) /* >>1 */, 0x7F7F7F)
                        + (destPixels[StreamBase.blend(4032, savedUSlope)
                                + (texFrac >> (955305670 & 31) /* >>6 */)] >>> brightness);
                texFrac += uInc;
                savedUSlope += vInc;

                // Pixel 2
                destPixels[destIndex++] =
                        (StreamBase.blend(texels[destIndex], 0xFEFEFF) >> (393665345 & 31) /* >>1 */)
                        + (destPixels[StreamBase.blend(4032, savedUSlope += vInc)
                                - -((texFrac += uInc) >> (556791558 & 31) /* >>6 */)] >>> brightness);

                // Pixel 3
                destPixels[destIndex++] =
                        (StreamBase.blend(0xFEFEFF, texels[destIndex]) >> (-2007060127 & 31) /* >>1 */)
                        + (destPixels[StreamBase.blend(savedUSlope += vInc, 4032)
                                + ((texFrac += uInc) >> (-1069632730 & 31) /* >>6 */)] >>> brightness);

                // Pixel 4
                destPixels[destIndex++] =
                        (StreamBase.blend(texels[destIndex], 0xFEFEFF) >> (-526841663 & 31) /* >>1 */)
                        + (destPixels[((texFrac += uInc) >> (-1891324570 & 31) /* >>6 */)
                                + StreamBase.blend(4032, savedUSlope += vInc)] >>> brightness);

                // Advance accumulators for pixel 4 → 5 boundary; re-page texFrac at 4-pixel mark.
                texFrac += uInc;
                savedUSlope += vInc;
                texFrac = (texFrac & 0xC0000) + (0xFFF & texFrac); // snap U/V page boundary
                brightness = texFrac >> (-580603052 & 31); // >>20 — refresh brightness

                // Pixel 5
                destPixels[destIndex++] =
                        (destPixels[StreamBase.blend(savedUSlope += vInc, 4032)
                                + (texFrac >> (1328606726 & 31) /* >>6 */)] >>> brightness)
                        + (StreamBase.blend(texels[destIndex], 0xFEFEFE) >> (604787489 & 31) /* >>1 */);
                texFrac += uInc;

                // Pixel 6
                destPixels[destIndex++] =
                        (destPixels[((texFrac += uInc) >> (-830951482 & 31) /* >>6 */)
                                + StreamBase.blend(4032, savedUSlope += vInc)] >>> brightness)
                        + (StreamBase.blend(texels[destIndex], 0xFEFEFF) >> (310428257 & 31) /* >>1 */);
                texFrac += uInc;
                savedUSlope += vInc;

                // Pixel 7  (subtract of negation = add)
                destPixels[destIndex++] =
                        (destPixels[StreamBase.blend(4032, savedUSlope += vInc)
                                + ((texFrac += uInc) >> (-1841159226 & 31) /* >>6 */)] >>> brightness)
                        - -(StreamBase.blend(texels[destIndex], 0xFEFEFF) >> (-1760233471 & 31) /* >>1 */);

                // Pixel 8
                destPixels[destIndex++] =
                        (destPixels[StreamBase.blend(4032, savedUSlope += vInc)
                                + ((texFrac += uInc) >> (1454319654 & 31) /* >>6 */)] >>> brightness)
                        - -(StreamBase.blend(0xFEFEFF, texels[destIndex]) >> (1605358369 & 31) /* >>1 */);

                // Pixels 8 → 9 boundary: re-page texFrac at 8-pixel mark.
                texFrac += uInc;
                savedUSlope += vInc;
                uStepScaled += uStepScaled; // obf step-accumulator advance
                texFrac = (0xC0000 & uStepScaled) + (0xFFF & texFrac);
                brightness = texFrac >> (1147218452 & 31); // >>20

                // Pixel 9
                destPixels[destIndex++] =
                        (destPixels[StreamBase.blend(4032, savedUSlope += vInc)
                                - -(texFrac >> (1983636742 & 31) /* >>6 */)] >>> brightness)
                        - -(StreamBase.blend(texels[destIndex], 0xFEFEFE) >> (1637168449 & 31) /* >>1 */);
                texFrac += uInc;

                // Pixel 10
                destPixels[destIndex++] =
                        (destPixels[StreamBase.blend(savedUSlope += vInc, 4032)
                                - -((texFrac += uInc) >> (1901625030 & 31) /* >>6 */)] >>> brightness)
                        - -StreamBase.blend(texels[destIndex] >> (-256795167 & 31) /* >>1 */, 0x7F7F7F);

                // Pixel 11
                destPixels[destIndex++] =
                        (destPixels[StreamBase.blend(4032, savedUSlope += vInc)
                                + ((texFrac += uInc) >> (1605754694 & 31) /* >>6 */)] >>> brightness)
                        - -(StreamBase.blend(texels[destIndex], 0xFEFEFF) >> (-216359295 & 31) /* >>1 */);

                // Pixel 12
                destPixels[destIndex++] =
                        (StreamBase.blend(0xFEFEFE, texels[destIndex]) >> (791103809 & 31) /* >>1 */)
                        + (destPixels[((texFrac += uInc) >> (371413222 & 31) /* >>6 */)
                                + StreamBase.blend(savedUSlope += vInc, 4032)] >>> brightness);

                // Pixels 12 → 13 boundary: re-page texFrac at 12-pixel mark.
                texFrac += uInc;
                savedUSlope += vInc;
                uStepScaled += uStepScaled;
                texFrac = ((uStepScaled) & 0xC0000) + (texFrac & 0xFFF);
                brightness = texFrac >> (711720340 & 31); // >>20

                // Pixel 13
                destPixels[destIndex++] =
                        (destPixels[StreamBase.blend(savedUSlope += vInc, 4032)
                                - -(texFrac >> (-2780922 & 31) /* >>6 */)] >>> brightness)
                        - -(StreamBase.blend(0xFEFEFE, texels[destIndex]) >> (962756193 & 31) /* >>1 */);
                uStepScaled += uStepScaled;

                // Pixel 14
                destPixels[destIndex++] =
                        (StreamBase.blend(texels[destIndex], 0xFEFEFF) >> (-838985215 & 31) /* >>1 */)
                        + (destPixels[StreamBase.blend(savedUSlope += vInc, 4032)
                                - -((texFrac += uInc) >> (1389805702 & 31) /* >>6 */)] >>> brightness);
                texFrac += uInc;
                savedUSlope += vInc;

                // Pixel 15
                destPixels[destIndex++] =
                        (destPixels[StreamBase.blend(4032, savedUSlope += vInc)
                                - -((texFrac += uInc) >> (-1171869722 & 31) /* >>6 */)] >>> brightness)
                        - -StreamBase.blend(texels[destIndex] >> (1072420929 & 31) /* >>1 */, 0x7F7F7F);

                // Pixel 16
                destPixels[destIndex++] =
                        (destPixels[((texFrac += uInc) >> (227032774 & 31) /* >>6 */)
                                + StreamBase.blend(4032, savedUSlope += vInc)] >>> brightness)
                        + (StreamBase.blend(texels[destIndex], 0xFEFEFF) >> (-454180287 & 31) /* >>1 */);
            }

            // ---- Residual loop: handles 0–15 leftover pixels ----
            for (int residual = 0; remaining > residual; ++residual) {
                destPixels[destIndex++] =
                        (destPixels[(texFrac >> (-208962138 & 31) /* >>6 */)
                                + StreamBase.blend(savedUSlope, 4032)] >>> brightness)
                        - -(StreamBase.blend(0xFEFEFE, texels[destIndex]) >> (-1883934911 & 31) /* >>1 */);
                savedUSlope += vInc;
                texFrac      += uInc;

                // Every 4th residual pixel: re-page texFrac and advance U step accumulator.
                if ((residual & 3) == 3) {
                    brightness   = texFrac >> (-2030987340 & 31); // >>20
                    texFrac      = (texFrac & 0xFFF) - -(0xC0000 & uStepScaled);
                    uStepScaled += uStepScaled;
                }
            }

            remaining -= 16;
        } while (remaining > 0);
    }

    // -------------------------------------------------------------------------
    // XOR string-pool decoder helpers  (private static, obf: jb.z)
    // These are dead at runtime once the static initialiser has run.
    // -------------------------------------------------------------------------

    /**
     * Pass-1 XOR decoder: {@link String} → {@code char[]}.
     * For strings with fewer than 2 characters, XORs the single character with {@code 0x2F ('/').}
     * For longer strings, equivalent to {@code s.toCharArray()}.
     * obf: jb.z(String)
     */
    private static char[] decodePass1(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ '/'); // 0x2F
        }
        return chars;
    }

    /**
     * Pass-2 XOR decoder: {@code char[]} → interned {@link String}.
     * XORs each character with the rotating 5-byte key {@code [39, 56, 26, 30, 47]}.
     * obf: jb.z(char[])
     */
    private static String decodePass2(char[] chars) {
        final byte[] KEY = {39, 56, 26, 30, 47};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char) (chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
