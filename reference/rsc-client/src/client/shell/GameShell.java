package client.shell;

import java.applet.Applet;
import java.applet.AppletContext;
import java.awt.Color;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics;
import java.awt.Image;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.io.IOException;
import java.lang.reflect.Method;
import java.net.URL;

import client.shell.GameFrame;          // qb
import client.shell.InputState;         // kb (holds the static GameFrame `a`)
import client.shell.LoaderThread;       // c
import client.net.ClientStream;         // da (holds the static loader Applet `gb`)
import client.net.StreamBase;           // ib (resource/archive reader)
import client.scene.ImageLoader;        // pa (decodes a TGA byte[] into an Image)
import client.ui.Panel;                 // qa (font loading: Panel.createFont)
import client.ui.FontBuilder;           // s
import client.data.CacheUpdater;        // cb
import client.data.DataStore;           // nb
import client.util.Globals;             // l — global holder (params/strings)
import client.util.LinkedQueue;          // db — holds the shared global int `d` (sharedInt)
import client.util.Timer;               // p (System.currentTimeMillis wrapper)
import client.util.Utility;             // mb (sleep + logging helpers) / na (loadData)
import client.util.StreamFactory;       // na
import client.util.ErrorHandler;        // i (wraps Throwable + context string)
import client.util.JSBridge;            // a (LiveConnect / JavaScript bridge)

/**
 * GameShell — the applet/window shell that the whole RuneScape Classic client is built on.
 *
 * <p>Obfuscated name: {@code e}. It directly extends {@link java.applet.Applet} and is itself the
 * superclass of the main game class ({@code client} / Mudclient). Responsibilities:
 * <ul>
 *   <li>Hosting the AWT surface — either as a browser applet or, when {@link #startApplication}
 *       is used, inside a standalone {@link GameFrame} window.</li>
 *   <li>Owning the single game {@link Thread} and its main loop ({@link #run()}), which paces
 *       itself to a target FPS, calls {@link #handleInputs(int)} (overridden by the game) a variable
 *       number of times per frame, and then {@link #draw(boolean)}s.</li>
 *   <li>Receiving all mouse / keyboard input via the AWT listener interfaces and stashing it into
 *       simple polled fields ({@link #mouseX}, {@link #mouseButtonDown}, {@link #inputTextCurrent}
 *       …) that the game samples each loop.</li>
 *   <li>Drawing the Jagex boot/loading screen (logo + progress bar) and loading the boot archives
 *       (jagex.jag → logo, fonts###.jag → bitmap fonts).</li>
 *   <li>Forwarding applet-context calls (getParameter / getCodeBase / getGraphics …) to whichever
 *       host is active: the standalone {@link GameFrame}, an external loader {@link Applet}, or the
 *       default {@code Applet} super-implementation.</li>
 * </ul>
 *
 * <p>This particular build is the late (~rev 233–235, Microsoft J++ / 2015 re-release) variant: it
 * adds a loader-applet indirection ({@code ClientStream.applet}) and a JavaScript bridge
 * ({@link JSBridge}) used to navigate the browser on logout/crash, neither of which exists in the
 * classic rev-204 oracle.
 *
 * <p>NOTE ON DEOBFUSCATION: the original carries, in every method, (1) an opaque predicate
 * {@code boolean bl = client.vh} (always false) driving dead branches, (2) a per-method static
 * profiling counter increment, and (3) a try/catch that rewraps any RuntimeException with a textual
 * method signature via {@link ErrorHandler}. All three have been stripped here; only the real logic
 * remains. Several "setter" helpers ({@code a(int)}, {@code b(boolean)}, {@code e(int)} …) in the
 * obfuscated class are pure no-ops whose entire body is a dead profiling counter guarded by an
 * always-false condition — they exist only to host the counter and are documented as such below.
 */
public class GameShell extends Applet implements Runnable, MouseListener, MouseMotionListener, KeyListener {

    /** Characters accepted into the text/PM input buffers (printable keyboard set incl. space). */
    private static final String[] STR = decodeStringTable();

    // ------------------------------------------------------------------
    // Layout / loading state
    // ------------------------------------------------------------------

    /** Vertical origin offset applied to the graphics context and subtracted from mouse Y. (obf K) */
    public int originY;                                   // K
    /** Logo header line drawn above the progress bar (e.g. a status/error banner); null = none. (obf p) */
    public String logoHeaderText = null;                  // p
    /** Per-frame timing ring buffer (10 recent loop timestamps), used to compute FPS. (obf F) */
    public long[] timings = new long[10];                 // F

    /**
     * Loading screen archive cache: raw (decompressed) bytes for the boot archives, indexed by an
     * internal slot. Shared static scratch used by the resource layer. (obf kb)
     */
    public static byte[][] archiveCache = new byte[250][];

    /**
     * Per-entity equipment string table (obf {@code e.Mb}; clean e.java:39 {@code static String[] Mb}).
     * Allocated and filled by {@link client.net.SocketFactory#initGameData} (GameData entity tier).
     */
    public static String[] equipMb;

    /**
     * Per-bucket size key array for the {@code EntityDef.bytePool} size-keyed byte-array pool
     * (obf {@code e.wb}; clean e.java:86 {@code static int[] wb}). Read by
     * {@link client.util.Utility#allocateByteArray} to match a requested size to a pool bucket.
     */
    public static int[] wb;

    /** Main game thread running {@link #run()}. (obf z) */
    public Thread gameThread;                             // z
    /** Target frame interval = 1000 / fps; lower is faster. (obf Ib) */
    private int targetFps = 20;                    // Ib
    /** Font used for the loading text (TimesRoman 15). (obf tb) */
    public Font loadingFont;                              // tb
    /** True while the (unused in this build) "referer logo" alternate boot styling is active. (obf hb) */
    private boolean hasRefererLogo;                // hb

    /** Loading progress bar text (e.g. "Loading"). (obf B) */
    public String loadingText;                            // B
    /** Applet/canvas height in pixels (default 384). (obf a) */
    private int appletHeight = 384;                // a
    /** Set true at the top of every paint(); a "has painted at least once" marker. (obf N) */
    public boolean hasPainted = false;                    // N

    /** Decoded color-channel scratch palette buffer (256 entries), unused here. (obf nb) */
    public static int[] colorScratch = new int[512];      // nb

    /** Mouse X in client coordinates (raw X minus {@link #originX}). (obf I) */
    public int mouseX = 0;                                // I
    /** ChatCipher / crypto helper instance (cross-class type {@code v}). (obf i — static field) */
    public static ChatCipher chatCipher;                  // i (static field, NOT ErrorHandler)
    /** Mouse-idle countdown (frames since last mouse activity); some draw code uses this. (obf sb) */
    public int mouseActionTimeout = 0;                    // sb

    /** Horizontal origin offset applied to graphics + subtracted from mouse X. (obf Eb) */
    public int originX;                                   // Eb

    /** Width of the applet/canvas in pixels (default 512). (obf m) */
    private int appletWidth = 512;                 // m
    /** Maximum number of handleInputs() calls allowed per draw before forcing a frame. (obf S) */
    private int maxLogicPerFrame = 1000;           // S
    /** Rolling interlace pacing counter; toggles {@link #interlace} when the loop falls behind. (obf b) */
    private int interlaceTimer = 0;                // b

    /**
     * Stop / shutdown countdown state machine:
     *  >0  = number of frames left before {@link #closeProgram} (set by {@link #stop()}),
     *   0  = running normally,
     *  -1  = destroy() requested, force-close after grace period,
     *  -2  = closing in progress. (obf vb)
     */
    private int stopTimeout = 0;                   // vb

    /** Loading progress percent last drawn (0–100). (obf V) */
    private int loadingPercent = 0;                // V

    /** Decoded TGA palette/pixel scratch buffer, unused here. (obf wb) */
    public static int[] tgaScratch;                       // wb

    /** Bold Helvetica 13 — Jagex credit lines on the loading screen. (obf X) */
    public Font creditFont;                               // X
    /** Plain Helvetica 12 — alternate credit line. (obf Jb) */
    public Font creditFontSmall;                          // Jb
    /** The live drawing context (a translated copy of the host Graphics). (obf u) */
    public Graphics graphics;                             // u
    /** The Jagex logo image decoded from logo.tga. (obf C) */
    public Image logoImage;                               // C

    /** Mouse Y in client coordinates (raw Y minus {@link #originY}). (obf xb) */
    public int mouseY = 0;                                // xb
    /** Current text-input accumulator (login/username field), capped at 20 chars. (obf e) */
    public String inputTextCurrent = "";                  // e
    /** True once the browser has been navigated away (logout/crash) so we only do it once. (obf Kb) */
    private boolean navigatedAway = false;         // Kb
    /** Current private-message input accumulator, capped at 80 chars. (obf x) */
    public String inputPmCurrent = "";                    // x
    /** F1 interlace toggle (skip every other scanline when rendering). (obf U) */
    public boolean interlace = false;                     // U

    /** Which mouse button is currently down: 0=none, 1=left, 2=right(meta). (obf Bb) */
    public int mouseButtonDown = 0;                       // Bb
    /** True while CTRL is held (bit 1 of input modifiers). (obf gb) */
    public boolean ctrlDown = false;                      // gb
    /** Snapshot of {@link #inputTextCurrent} taken when ENTER is pressed. (obf Cb) */
    public String inputTextFinal = "";                    // Cb
    /** True while the RIGHT arrow navigation key is held; rotates the camera right in the game. (obf E) */
    public boolean keyRight = false;                       // E  (VK_RIGHT == 39; the game reads this for camera rotation)
    /** Minimum thread sleep floor (ms) applied when the loop is ahead of schedule. (obf Q) */
    public int minSleep = 1;                              // Q
    /** True while ALT is held (bit 2 of input modifiers). (obf bb) */
    public boolean altDown = false;                       // bb
    /** True while the LEFT arrow navigation key is held. (obf Z) */
    public boolean keyLeft = false;                       // Z
    /** Mouse button captured on the last press (for click-vs-drag disambiguation). (obf Qb) */
    public int lastMouseButtonDown = 0;                   // Qb
    /** Snapshot of {@link #inputPmCurrent} taken when ENTER is pressed. (obf Ob) */
    public String inputPmFinal = "";                      // Ob

    // ------------------------------------------------------------------
    // Public static engine flags (referenced cross-class). Names kept generic.
    // ------------------------------------------------------------------
    public static int publicFlagAb;                // Ab
    public static boolean publicFlagT;             // T
    public static boolean publicFlagH;             // H
    public static int publicFlagZb;                // zb
    public static int publicFlagV;                 // v
    public static boolean publicFlagIb;            // ib
    public static int publicFlagLb;                // lb

    // ------------------------------------------------------------------
    // Profiling counters (obf static ints). Each is bumped once per call of its method in the
    // original build; kept as fields for faithfulness but no longer incremented in real logic.
    // ------------------------------------------------------------------
    static int profGetGraphics, profKeyTyped, profGetParameter, profStartApplication, profGetDocumentBase,
            profStartApplet, profSetTargetFps, profStartThread, profDrawString, profGetCodeBase,
            profResetTimings, profSetTargetFpsB, profDestroy, profMouseExited, profIsDisplayable,
            profMouseEntered, profUpdate, profGetAppletContext, profGetSize, profMouseClicked,
            profRun, profGetGraphicsB, profCloseProgram, profKeyPressed, profStart, profMouseMoved,
            profMousePressed, profStop, profDecodeImage, profReadDataFile, profProvideLoader,
            profDrawLoading, profMouseReleased, profMouseDragged, profShowProgress, profCreateImage,
            profKeyReleased, profLoadJagex, profSetSleep, profInputModifiers, profHandleKeyPress,
            profHandleMouseDown, profDummyA, profDummyB, profDummyC, profDummyD, profDummyE,
            profDummyF, profDummyG, profStartGame, profDraw, profOnClosing, profDummyH;

    // ==================================================================
    // Applet-context delegation: route to GameFrame, loader applet, or super.
    // ==================================================================

    /** Returns an applet parameter, delegating to the loader applet when running in browser-loader mode. */
    public final String getParameter(String name) {
        // In standalone (GameFrame) mode there are no applet params.
        if (InputState.gameFrame != null) {
            return null;
        }
        if (ClientStream.applet != null) {
            return ClientStream.applet.getParameter(name);
        }
        return super.getParameter(name);
    }

    /**
     * Shutdown sequence: mark closing, run {@link #onClosing()} (overridden by the game), pause
     * briefly, dispose the frame and (if launched standalone) exit the JVM. obf {@code b(int)}.
     */
    private void closeProgram(int marker) {
        stopTimeout = -2;
        System.out.println(STR[28]); // "Closing program"
        onClosing();
        Utility.sleepWithProfile(11200, 1000L);
        if (marker != 100) {
            // obf: this.e(27) — e(int) is the handleInputs() hook (overridden in Mudclient).
            // Dead in practice: closeProgram is only ever invoked with marker == 100, so this
            // never runs. Modeled as the no-op shell default.
            this.handleInputs(27);
        }
        if (InputState.gameFrame == null) {
            return;
        }
        InputState.gameFrame.dispose();
        System.exit(0);
    }

    /**
     * Main game loop. On first entry it performs the one-time boot (load Jagex logo + fonts, then
     * call {@link #startGame()}); thereafter it paces itself to {@link #targetFps}, runs
     * {@link #handleInputs(int)} a variable number of times to "catch up" on game logic, and draws once
     * per frame. Honors {@link #stopTimeout} for graceful shutdown.
     *
     * <p>(Reconstructed from the rev-204 oracle; the obfuscated bytecode failed to decompile.)
     */
    @Override
    public final void run() {
        try {
            // ---- one-time boot ----
            if (loadingStepIsOne()) {
                loadingStep = 2;
                while (!isDisplayable()) {
                    // Wait for the host window/peer to be realized before drawing. While waiting,
                    // still honor a pending stop/destroy request.
                    if (stopTimeout < 0) {
                        break;
                    }
                    if (stopTimeout > 0) {
                        if (--stopTimeout == 0) {
                            closeProgram(100);
                            gameThread = null;
                            return;
                        }
                    }
                    Utility.sleepWithProfile(11200, targetFps);
                }
                if (stopTimeout < 0) {
                    if (stopTimeout == -1) {
                        closeProgram(100);
                    }
                    gameThread = null;
                    return;
                }
                if (loadJagex()) {
                    setLoaderApplet((byte) -92); // marker; no-op when arg == -92
                    loadingStep = 0;
                } else {
                    // loadJagex() failed: close unless we are already closing (obf: if (~vb != 1) → vb != -2).
                    if (stopTimeout != -2) {
                        closeProgram(100);
                    }
                    gameThread = null;
                    return;
                }
            }

            // Register input listeners against whichever host is active.
            if (InputState.gameFrame != null) {
                InputState.gameFrame.addMouseListener(this);
                InputState.gameFrame.addMouseMotionListener(this);
                InputState.gameFrame.addKeyListener(this);
            } else if (ClientStream.applet != null) {
                ClientStream.applet.addMouseListener(this);
                ClientStream.applet.addMouseMotionListener(this);
                ClientStream.applet.addKeyListener(this);
            } else {
                addMouseListener(this);
                addMouseMotionListener(this);
                addKeyListener(this);
            }

            // ---- steady-state loop ----
            int fpsScale = 256;   // 256 == "on schedule"; scaled FPS estimate
            int sleep = 1;
            int logicAccumulator = 0;
            for (int i = 0; i < 10; i++) {
                timings[i] = Timer.currentTimeMillisCorrected(0);
            }

            int ringIndex = 0;
            while (stopTimeout >= 0) {
                if (stopTimeout > 0) {
                    if (--stopTimeout == 0) {
                        closeProgram(100);
                        gameThread = null;
                        return;
                    }
                }

                int prevFpsScale = fpsScale;
                int prevSleep = sleep;
                fpsScale = 300;
                sleep = 1;
                long now = Timer.currentTimeMillisCorrected(0);
                if (timings[ringIndex] == 0L) {
                    // No baseline yet for this slot — keep previous estimates.
                    fpsScale = prevFpsScale;
                    sleep = prevSleep;
                } else if (now > timings[ringIndex]) {
                    // fpsScale = 2560 * targetFps / elapsed-over-10-frames.
                    fpsScale = (int) ((long) (targetFps * 2560) / (now - timings[ringIndex]));
                }

                if (fpsScale < 25) {
                    fpsScale = 25;
                }
                if (fpsScale > 256) {
                    // Running faster than target: clamp and compute a real sleep to slow down.
                    fpsScale = 256;
                    sleep = (int) ((long) targetFps - (now - timings[ringIndex]) / 10L);
                    if (sleep < minSleep) {
                        sleep = minSleep;
                    }
                }

                Utility.sleepWithProfile(11200, sleep);
                timings[ringIndex] = now;
                ringIndex = (ringIndex + 1) % 10;

                if (sleep > 1) {
                    // We slept, so shift all baselines forward by that amount.
                    for (int i = 0; i < 10; i++) {
                        if (timings[i] != 0L) {
                            timings[i] += sleep;
                        }
                    }
                }

                int logicCalls = 0;
                while (logicAccumulator < 256) {
                    handleInputs(119); // obf: this.e(119)
                    logicAccumulator += fpsScale;
                    if (++logicCalls > maxLogicPerFrame) {
                        // Logic is falling behind: force a frame and bump the interlace pacer.
                        logicAccumulator = 0;
                        interlaceTimer += 6;
                        if (interlaceTimer > 25) {
                            interlaceTimer = 0;
                            interlace = true;
                        }
                        break;
                    }
                }
                interlaceTimer--;
                logicAccumulator &= 0xFF;
                draw(false); // obf: this.b(false)
            }

            if (stopTimeout == -1) {
                closeProgram(100);
            }
            gameThread = null;
        } catch (Exception ex) {
            // Fatal: log and show a "crash" banner on the loading screen.
            Utility.reportError(0x1FFFFF, ex, null);
            navigateAway(STR[12], true); // "crash"
        }
    }

    /** Key-up handler: clears the held-arrow flags. (obf keyReleased) */
    @Override
    public final synchronized void keyReleased(KeyEvent event) {
        captureModifiers(event, (byte) -128);
        char chr = event.getKeyChar();
        int code = event.getKeyCode();
        // The original had a series of empty if-blocks here for keys whose press/release handling
        // was stubbed out (page nav, 'N'/'M', '{', etc.). Only the two live ones remain:
        if (code == KeyEvent.VK_RIGHT) {    // code == 39 (obf tested ~code == -40)
            keyRight = false;
        }
        if (code == KeyEvent.VK_LEFT) {     // code == 37
            keyLeft = false;
        }
    }

    /**
     * Repaints the boot/loading screen. Only meaningful while {@code loadingStep == 2} and the logo
     * has been decoded; once the game proper is running it is effectively a no-op. (obf paint)
     */
    @Override
    public final void paint(Graphics g) {
        hasPainted = true;
        if (loadingStep == 2 && logoImage != null) {
            drawLoadingScreen(loadingText, loadingPercent, 126);
        }
        // loadingStep == 0 fell through to a stubbed call in the original; nothing to do.
    }

    /**
     * Full loading-screen redraw: clears to black, blits the logo, draws the progress-bar frame +
     * fill, the progress text and the Jagex credit lines (or the optional {@link #logoHeaderText}).
     * obf {@code a(String,int,int)}. The third arg only gated a stubbed mouseReleased call.
     */
    private void drawLoadingScreen(String text, int percent, int unused) {
        try {
            int x = (appletWidth - 281) / 2;
            int y = (appletHeight - 148) / 2;
            graphics.setColor(Color.black);
            graphics.fillRect(0, 0, appletWidth, appletHeight);
            if (!hasRefererLogo) {
                graphics.drawImage(logoImage, x, y, this);
            }
            x += 2;
            loadingPercent = percent;
            y += 90;
            loadingText = text;
            if (unused <= 97) {
                // Stubbed in original (mouseReleased(null)); preserved as a guarded no-op.
                this.mouseReleased((MouseEvent) null);
            }
            graphics.setColor(new Color(132, 132, 132));
            if (hasRefererLogo) {
                graphics.setColor(new Color(220, 0, 0));
            }
            graphics.drawRect(x - 2, y - 2, 280, 23);
            graphics.fillRect(x, y, 277 * percent / 100, 20);
            graphics.setColor(new Color(198, 198, 198));
            if (hasRefererLogo) {
                graphics.setColor(new Color(255, 255, 255));
            }
            drawCenteredString(loadingFont, text, 10 + y, true, 138 + x, graphics);
            if (!hasRefererLogo) {
                drawCenteredString(creditFont, STR[24], 30 + y, true, x + 138, graphics); // "Created by JAGeX..."
                drawCenteredString(creditFont, STR[23], y + 44, true, x + 138, graphics); // "(c) 2001-2015 Jagex Ltd"
            } else {
                graphics.setColor(new Color(132, 132, 152));
                drawCenteredString(creditFontSmall, STR[23], appletHeight - 20, true, 138 + x, graphics);
            }
            if (logoHeaderText != null) {
                graphics.setColor(Color.white);
                drawCenteredString(creditFont, logoHeaderText, y - 120, true, x + 138, graphics);
            }
        } catch (Exception ignored) {
        }
    }

    /** Mouse button released: latch coordinates and clear the pressed button. (obf mouseReleased) */
    @Override
    public final synchronized void mouseReleased(MouseEvent event) {
        captureModifiers(event, (byte) -128);
        this.mouseX = event.getX() - this.originX;
        this.mouseY = event.getY() - this.originY;
        this.mouseButtonDown = 0;
    }

    /**
     * In-place progress-bar update (no logo / credits redraw) used during archive download to show
     * incremental percentages. obf {@code a(int,byte,String)}. The byte arg only gated a stub.
     */
    public final void showLoadingProgress(int percent, byte clearPm, String text) {
        try {
            int x = (appletWidth - 281) / 2;
            x += 2;
            int y = (appletHeight - 148) / 2;
            this.loadingText = text;
            this.loadingPercent = percent;
            y += 90;
            int filled = 277 * percent / 100;
            this.graphics.setColor(new Color(132, 132, 132));
            if (this.hasRefererLogo) {
                this.graphics.setColor(new Color(220, 0, 0));
            }
            this.graphics.fillRect(x, y, filled, 20);
            this.graphics.setColor(Color.black);
            this.graphics.fillRect(filled + x, y, 277 - filled, 20);
            this.graphics.setColor(new Color(198, 198, 198));
            if (this.hasRefererLogo) {
                this.graphics.setColor(new Color(255, 255, 255));
            }
            this.drawCenteredString(this.loadingFont, text, 10 + y, true, 138 + x, this.graphics);
        } catch (Exception ignored) {
        }
        if (clearPm > -96) {
            this.inputPmCurrent = null;
        }
    }

    /** Decodes a raw TGA byte[] into an AWT Image via {@link ImageLoader}. obf {@code a(byte[],byte)}. */
    private Image decodeImage(byte[] tga, byte marker) {
        if (marker != -54) {
            this.inputPmFinal = null; // dead anti-tamper assignment
        }
        return ImageLoader.loadBmpImage(79, this, tga);
    }

    /** Installs the external loader applet that hosts this client inside a browser. obf static {@code provideLoaderApplet}. */
    public static final void provideLoaderApplet(Applet applet) {
        ClientStream.applet = applet;
    }

    /** No-op profiling stub: {@code a(int)} in the original holds only a dead counter. */
    public final void profileA(int dummy) {
        // body was: if (dummy < -54) profDecodeImage++;  — pure profiling, no effect.
    }

    /** Returns the document base, delegating to the loader applet in browser-loader mode. */
    public final URL getDocumentBase() {
        if (InputState.gameFrame != null) {
            return null;
        }
        return ClientStream.applet != null
                ? ClientStream.applet.getDocumentBase()
                : super.getDocumentBase();
    }

    /**
     * Standalone (non-applet) launch path: opens a {@link GameFrame} window of the given size, wires
     * up the loader {@link LoaderThread}, optionally kicks off a content/cache update from a host,
     * and starts the game thread. obf {@code a(boolean,String,int,String,int,byte,int,int,int)}.
     *
     * @param resizable    whether the window may be resized
     * @param title        window title
     * @param loaderArg    arg passed to the LoaderThread
     * @param loaderName   name passed to the LoaderThread
     * @param port         port for the content-pack update URL
     * @param doUpdate     marker; only triggers the cache update when > 20
     * @param storeFlag    written into DataStore.flag
     * @param width        canvas width
     * @param height       canvas height
     */
    public final void startApplication(boolean resizable, String title, int loaderArg, String loaderName,
                                int port, byte doUpdate, int storeFlag, int width, int height) {
        try {
            System.out.println(STR[16]); // "Started application"
            this.appletWidth = width;
            this.appletHeight = height;
            InputState.gameFrame = new GameFrame(this, 800, 600, title, resizable, false);
            try {
                // Reflectively disable AWT focus-traversal keys (so TAB reaches the game).
                InputState.gameFrame.getClass()
                        .getMethod(STR[17], boolean.class) // "setFocusTraversalKeysEnabled"
                        .invoke(InputState.gameFrame, Boolean.FALSE);
            } catch (Exception ignored) {
            }
            // obf: db.d = storeFlag  — clean writes LinkedQueue.sharedInt (obf db.d), NOT Globals (obf l).
            // The import comment conflated obf classes db (LinkedQueue) and l (Globals); fixed receiver here.
            LinkedQueue.sharedInt = storeFlag;
            this.loadingStep = 1;
            // obf: pa.b = pa.k = new c(...).  pa.b = ImageLoader.loaderThread, pa.k = ImageLoader.imageWidthCarrier;
            // both LoaderThread refs are set in lockstep (clean e.java:719).
            ImageLoader.loaderThread = ImageLoader.imageWidthCarrier = new LoaderThread(loaderArg, loaderName, 0, true);
            try {
                if (doUpdate <= 20) {
                    return;
                }
                // Pull an HTTP content pack / cache update from the given host:port.
                CacheUpdater.downloadAndVerifyCrcs(new URL(STR[15], STR[14], port, ""), this, -91); // "http", "127.0.0.1"
            } catch (IOException ioe) {
                Utility.reportError(0x1FFFFF, ioe, null);
            }
            this.gameThread = new Thread(this);
            this.gameThread.start();
            this.gameThread.setPriority(1);
        } catch (Exception ex) {
            Utility.reportError(0x1FFFFF, ex, null);
        }
    }

    /**
     * Loads the boot archives and bitmap fonts: jagex.jag → logo.tga (the Jagex logo), then
     * fonts###.jag → eight {@code .jf} bitmap fonts registered into {@link Panel}. Returns false if
     * any required resource fails to load. obf {@code b(byte)} — the byte arg is a dead guard.
     */
    private boolean loadJagex() {
        byte[] jagexJag = readDataFile(STR[32], 0, 3, 85); // "Jagex library"
        if (jagexJag == null) {
            return false;
        }
        byte[] logoTga = StreamFactory.lookupEntityDefRecord(STR[30], 0, jagexJag); // "logo.tga"
        this.logoImage = this.decodeImage(logoTga, (byte) -54);
        // Register the bitmap fonts (h11p, h12b, h12p, h13b, h14b, h16b, h20b, h24b) into slots 0..7.
        if (!Panel.loadFont(this, STR[29], 0, 0)) return false; // h11p
        if (!Panel.loadFont(this, STR[39], 1, 0)) return false; // h12b
        if (!Panel.loadFont(this, STR[31], 2, 0)) return false; // h12p
        if (!Panel.loadFont(this, STR[36], 3, 0)) return false; // h13b
        if (!Panel.loadFont(this, STR[38], 4, 0)) return false; // h14b
        if (!Panel.loadFont(this, STR[34], 5, 0)) return false; // h16b
        if (!Panel.loadFont(this, STR[35], 6, 0)) return false; // h20b
        return Panel.loadFont(this, STR[37], 7, 0);             // h24b
    }

    /** Sets {@link #stopTimeout} to 85 when requested. obf {@code a(boolean)} — only the true branch acts. */
    public void requestStop(boolean request) {
        if (request) {
            this.stopTimeout = 85;
        }
    }

    /** Returns the applet context, delegating to the loader applet in browser-loader mode. */
    public final AppletContext getAppletContext() {
        if (InputState.gameFrame != null) {
            return null;
        }
        return ClientStream.applet != null
                ? ClientStream.applet.getAppletContext()
                : super.getAppletContext();
    }

    /** Clears the game thread when the (second) coordinate arg is small. obf {@code a(int,int,int,int)} — mostly dead. */
    public void maybeClearThread(int a, int b, int c, int d) {
        if (b < 87) {
            this.gameThread = null;
        }
    }

    /**
     * Navigates the host browser to a status page (e.g. "loggedout" or "error_game_crash") via the
     * {@link JSBridge} / LiveConnect, falling back to {@code showDocument}. Runs at most once.
     * obf {@code a(String,boolean)}.
     */
    private void navigateAway(String page, boolean once) {
        if (this.navigatedAway) {
            return;
        }
        this.navigatedAway = once;
        System.out.println(STR[63] + page); // "error_game_" + page
        try {
            // DRIFT FIX: JSBridge.call's deob signature is (String methodName, Applet applet) — the
            // dead anti-tamper byte param (obf var1 / "by") was stripped from the deob method, so the
            // callers must drop it too.  obf: a(methodName, byte, applet) → call(methodName, applet).
            if (ClientStream.applet != null) {
                JSBridge.call(STR[59], ClientStream.applet); // "loggedout"
            } else {
                JSBridge.call(STR[59], this);
            }
        } catch (Throwable ignored) {
        }
        try {
            this.getAppletContext().showDocument(
                    new URL(this.getCodeBase(), STR[63] + page + STR[62]), STR[60]); // ".ws", "_top"
        } catch (Exception ignored) {
        }
    }

    /** Mouse dragged: latch coordinates and set the button (meta = right = 2, else left = 1). (obf mouseDragged) */
    @Override
    public final synchronized void mouseDragged(MouseEvent event) {
        captureModifiers(event, (byte) -128);
        this.mouseX = event.getX() - this.originX;
        this.mouseY = event.getY() - this.originY;
        if (event.isMetaDown()) {
            this.mouseButtonDown = 2;
        } else {
            this.mouseButtonDown = 1;
        }
    }

    /**
     * Applet launch path: sets the canvas size, records the code base, optionally creates the
     * {@link LoaderThread}, reflectively toggles loader-applet visibility, runs a cache update from
     * the code base, and starts the game thread. obf {@code a(int,int,int,int,int)}.
     *
     * @param height   canvas height
     * @param storeFlag written into DataStore.flag
     * @param loaderArg arg for the LoaderThread
     * @param mode     2 == perform the visibility toggle + update; the update URL marker derives from it
     * @param width    canvas width
     */
    public final void startApplet(int height, int storeFlag, int loaderArg, int mode, int width) {
        try {
            System.out.println(STR[69]); // "Started applet"
            this.loadingStep = 1;
            this.appletWidth = width;
            this.appletHeight = height;
            DataStore.baseUrl = this.getCodeBase();
            // obf: db.d = storeFlag  — LinkedQueue.sharedInt (NOT Globals); see receiver note above.
            LinkedQueue.sharedInt = storeFlag;
            // obf: if (pa.k == null) { pa.b = pa.k = new c(...) }.  pa.k = ImageLoader.imageWidthCarrier;
            // pa.b and pa.k are set in lockstep (clean e.java:904-905).
            if (ImageLoader.imageWidthCarrier == null) {
                LoaderThread loader = new LoaderThread(loaderArg, null, 0, ClientStream.applet != null);
                ImageLoader.loaderThread = ImageLoader.imageWidthCarrier = loader;
            }
            if (mode != 2) {
                return;
            }
            if (ClientStream.applet != null) {
                // Reflectively show/hide loading UI on the host loader applet.
                Method showMethod = LoaderThread.setFocusCycleRoot;
                if (showMethod != null) {
                    try {
                        showMethod.invoke(ClientStream.applet, Boolean.TRUE);
                    } catch (Throwable ignored) {
                    }
                }
                Method hideMethod = LoaderThread.setFocusTraversalKeysEnabled;
                if (hideMethod != null) {
                    try {
                        hideMethod.invoke(ClientStream.applet, Boolean.FALSE);
                    } catch (Throwable ignored) {
                    }
                }
            }
            try {
                CacheUpdater.downloadAndVerifyCrcs(this.getCodeBase(), this, mode - 110);
            } catch (IOException ioe) {
                ioe.printStackTrace();
            }
            this.startThread(mode - 1, this);
        } catch (Exception ex) {
            Utility.reportError(mode ^ 0x1FFFFD, ex, null);
            this.navigateAway(STR[12], true); // "crash"
        }
    }

    /** Mouse entered the canvas: just refresh modifier state. (obf mouseEntered) */
    @Override
    public final void mouseEntered(MouseEvent event) {
        captureModifiers(event, (byte) -128);
    }

    /**
     * Creates a fresh, origin-translated drawing context, clears it to black, and draws the initial
     * "Loading..." screen. Returns false if no Graphics is available. obf {@code d(int)} — the arg
     * only gated a dead {@link #mouseActionTimeout} write.
     */
    public final boolean initGraphics(int marker) {
        Graphics hostGraphics = this.getGraphics();
        if (hostGraphics == null) {
            return false;
        }
        if (marker != 2) {
            this.mouseActionTimeout = -7; // dead anti-tamper write
        }
        this.graphics = hostGraphics.create();
        this.graphics.translate(this.originX, this.originY);
        this.graphics.setColor(Color.black);
        this.graphics.fillRect(0, 0, this.appletWidth, this.appletHeight);
        this.drawLoadingScreen(STR[56], 0, marker ^ 0x67); // "Loading..."
        return true;
    }

    /** Reads ALT (bit 2) and CTRL (bit 1) modifier state from an input event. obf {@code a(InputEvent,byte)}. */
    private void captureModifiers(InputEvent event, byte marker) {
        if (marker > -127) {
            return; // dead anti-tamper guard (marker is always -128 from callers)
        }
        int modifiers = event.getModifiers();
        this.altDown = (modifiers & 2) != 0;
        this.ctrlDown = (modifiers & 1) != 0;
    }

    /** Returns the canvas size from whichever host is active. */
    public final Dimension getSize() {
        if (InputState.gameFrame != null) {
            return InputState.gameFrame.getSize();
        }
        if (ClientStream.applet != null) {
            return ClientStream.applet.getSize();
        }
        return super.getSize();
    }

    /** Returns the host Graphics from whichever host is active. */
    public final Graphics getGraphics() {
        if (InputState.gameFrame != null) {
            return InputState.gameFrame.getGraphics();
        }
        if (ClientStream.applet != null) {
            return ClientStream.applet.getGraphics();
        }
        return super.getGraphics();
    }

    /** Returns the code base (null in standalone mode, loader applet's otherwise). */
    public final URL getCodeBase() {
        if (InputState.gameFrame != null) {
            return null;
        }
        return ClientStream.applet != null
                ? ClientStream.applet.getCodeBase()
                : super.getCodeBase();
    }

    /** Mouse clicked: refresh modifier state only. (obf mouseClicked) */
    @Override
    public final void mouseClicked(MouseEvent event) {
        captureModifiers(event, (byte) -128);
    }

    /** Creates an off-screen image via the active host. */
    public final Image createImage(int width, int height) {
        if (InputState.gameFrame != null) {
            return InputState.gameFrame.createImage(width, height);
        }
        if (ClientStream.applet != null) {
            return ClientStream.applet.createImage(width, height);
        }
        return super.createImage(width, height);
    }

    /**
     * Draws a string horizontally centered on {@code centerX} and vertically anchored at {@code y},
     * using the active host's {@link FontMetrics}. obf {@code a(Font,String,int,boolean,int,Graphics)}.
     * The boolean arg only gated a dead profiling call.
     */
    public final void drawCenteredString(Font font, String text, int y, boolean live, int centerX, Graphics g) {
        Object host = (InputState.gameFrame == null) ? this : InputState.gameFrame;
        FontMetrics metrics = ((Component) host).getFontMetrics(font);
        metrics.stringWidth(text); // (original called this once unused — kept for fidelity)
        if (!live) {
            this.profileC(68); // dead profiling-only call
        }
        g.setFont(font);
        g.drawString(text, centerX - metrics.stringWidth(text) / 2, y + metrics.getHeight() / 4);
    }

    /**
     * Applet destroy hook: request shutdown, wait up to 5s for the loop to acknowledge, then force
     * {@link #closeProgram} and stop the game thread. (obf destroy)
     */
    public final void destroy() {
        this.stopTimeout = -1;
        Utility.sleepWithProfile(11200, 5000L);
        if (this.stopTimeout != -1) {
            return; // the loop already shut down cleanly
        }
        System.out.println(STR[0]); // "5 seconds expired, forcing kill"
        this.closeProgram(100);
        if (this.gameThread != null) {
            this.gameThread.stop();
            this.gameThread = null;
        }
    }

    /** Clears modifier state by feeding a null event when the marker is small. obf {@code a(byte,int)} — mostly dead. */
    public void clearModifiers(byte marker, int dummy) {
        if (marker <= 105) {
            this.captureModifiers((InputEvent) null, (byte) 83);
        }
    }

    /** Mouse moved: latch coordinates and reset idle/button state. (obf mouseMoved) */
    @Override
    public final synchronized void mouseMoved(MouseEvent event) {
        captureModifiers(event, (byte) -128);
        this.mouseX = event.getX() - this.originX;
        this.mouseY = event.getY() - this.originY;
        this.mouseActionTimeout = 0;
        this.mouseButtonDown = 0;
    }

    /** Resets the timing ring buffer to zero. obf {@code c(int)} — runs only for its sentinel arg. */
    public final void resetTimings(int marker) {
        if (marker == -28492) {
            for (int i = 0; i < 10; i++) {
                this.timings[i] = 0L;
            }
        }
    }

    /**
     * Reads (and decompresses) a named boot archive via the resource layer, logging failures.
     * obf {@code a(String,int,int,int)}. The final int arg only gated a dead {@link #resetTimings} call.
     *
     * @param file       archive file name (e.g. "jagex" / "fonts###")
     * @param a          loader index/flag
     * @param percent    loading-bar percentage to display while reading
     */
    public final byte[] readDataFile(String file, int a, int percent, int marker) {
        if (marker <= 53) {
            this.resetTimings(15); // dead guard
        }
        try {
            return StreamBase.loadResource(-101, file, a, percent);
        } catch (IOException ioe) {
            Utility.reportError(0x1FFFFF, ioe, STR[65] + percent); // "Unable to load content pack "
            return null;
        }
    }

    /** Key typed: refresh modifier state only (text accumulation happens in keyPressed). (obf keyTyped) */
    @Override
    public final void keyTyped(KeyEvent event) {
        captureModifiers(event, (byte) -128);
    }

    /** Applet start hook: resume the loop (clear any pending stop countdown). (obf start) */
    public final void start() {
        if (this.stopTimeout >= 0) {
            this.stopTimeout = 0;
        }
    }

    /** Mouse exited the canvas: refresh modifier state only. (obf mouseExited) */
    @Override
    public final void mouseExited(MouseEvent event) {
        captureModifiers(event, (byte) -128);
    }

    /**
     * Mouse pressed: latch coordinates and button (meta = right = 2, else left = 1), record the
     * "last button" and dispatch {@link #handleMouseDown}. obf {@code mousePressed} (reconstructed
     * from the rev-204 oracle; the obfuscated bytecode failed to decompile).
     */
    @Override
    public final synchronized void mousePressed(MouseEvent event) {
        captureModifiers(event, (byte) -128);
        this.mouseX = event.getX() - this.originX;
        this.mouseY = event.getY() - this.originY;
        if (event.isMetaDown()) {
            this.mouseButtonDown = 2;
        } else {
            this.mouseButtonDown = 1;
        }
        this.lastMouseButtonDown = this.mouseButtonDown;
        this.mouseActionTimeout = 0;
        this.handleMouseDown(this.mouseX, 94, this.mouseButtonDown, this.mouseY);
    }

    /** Clears the installed loader applet when the marker differs. obf {@code a(byte)} — only the dead branch. */
    public void setLoaderApplet(byte marker) {
        if (marker != -92) {
            provideLoaderApplet(null);
        }
    }

    /**
     * Key pressed: forwards a character to {@link #handleKeyPress}, updates arrow / interlace state,
     * and accumulates printable characters into the text and PM input buffers (BACKSPACE deletes,
     * ENTER commits both into the *Final fields). obf {@code keyPressed} (reconstructed from the
     * rev-204 oracle; the obfuscated bytecode failed to decompile).
     */
    @Override
    public final synchronized void keyPressed(KeyEvent event) {
        captureModifiers(event, (byte) -128);
        char chr = event.getKeyChar();
        int code = event.getKeyCode();
        this.handleKeyPress((byte) 126, chr);

        if (code == KeyEvent.VK_F1) {        // toggle interlace
            this.interlace = !this.interlace;
        }
        this.mouseActionTimeout = 0;
        if (code == KeyEvent.VK_RIGHT) {     // code == 39 (obf tested ~code == -40)
            this.keyRight = true;
        }
        if (code == KeyEvent.VK_LEFT) {      // code == 37
            this.keyLeft = true;
        }

        // Accept the character if it is in the printable input set.
        boolean printable = false;
        for (int i = 0; i < ChatCipher.charMap.length(); i++) {
            if (chr == ChatCipher.charMap.charAt(i)) {
                printable = true;
                break;
            }
        }
        if (printable && this.inputTextCurrent.length() < 20) {
            this.inputTextCurrent += chr;
        }
        if (printable && this.inputPmCurrent.length() < 80) {
            this.inputPmCurrent += chr;
        }
        if (chr == 8 && this.inputTextCurrent.length() > 0) { // BACKSPACE
            this.inputTextCurrent = this.inputTextCurrent.substring(0, this.inputTextCurrent.length() - 1);
        }
        if (chr == 8 && this.inputPmCurrent.length() > 0) {
            this.inputPmCurrent = this.inputPmCurrent.substring(0, this.inputPmCurrent.length() - 1);
        }
        if (chr == 10 || chr == 13) { // ENTER — commit both buffers
            this.inputTextFinal = this.inputTextCurrent;
            this.inputPmFinal = this.inputPmCurrent;
        }
    }

    /** Sets the target frame rate (stored as 1000/fps). obf {@code a(int,byte)} — byte arg gated a dead write. */
    public final void setTargetFps(int fps, byte marker) {
        this.targetFps = 1000 / fps;
        if (marker <= 104) {
            this.originX = 113; // dead anti-tamper write
        }
    }

    /** Applet stop hook: schedule a graceful shutdown after ~4 seconds of frames. (obf stop) */
    public final void stop() {
        if (this.stopTimeout >= 0) {
            this.stopTimeout = 4000 / this.targetFps;
        }
    }

    /** True once the host window/peer has been realized (so we may draw to it). (obf isDisplayable) */
    public final boolean isDisplayable() {
        if (InputState.gameFrame != null) {
            return InputState.gameFrame.isDisplayable();
        }
        if (ClientStream.applet != null) {
            return ClientStream.applet.isDisplayable();
        }
        return super.isDisplayable();
    }

    /** AWT update hook: skip the default clear and paint directly to avoid flicker. (obf update) */
    public final void update(Graphics g) {
        this.paint(g);
    }

    /** Starts {@code runnable} on a daemon thread when {@code mode == 1}. obf {@code a(int,Runnable)}. */
    public void startThread(int mode, Runnable runnable) {
        Thread thread = new Thread(runnable);
        if (mode != 1) {
            return;
        }
        thread.setDaemon(true);
        thread.start();
    }

    /** Builds the three AWT loading-screen fonts. (obf constructor) */
    protected GameShell() {
        this.loadingFont = new Font(STR[43], 0, 15);    // TimesRoman 15
        this.creditFont = new Font(STR[42], 1, 13);     // Helvetica bold 13
        this.creditFontSmall = new Font(STR[42], 0, 12);// Helvetica plain 12
    }

    // ==================================================================
    // Overridable game hooks (empty in the shell; the game subclass fills them in).
    // In the obfuscated build these are inherited / overridden in `client`; the shell-level
    // declarations are implied by the virtual calls above. Provided here as documented stubs.
    // ==================================================================

    /** Loading-step state: 1 = needs boot, 2 = booting (draw loading screen), 0 = running. (obf n) */
    public int loadingStep = 1;

    /** Convenience: was {@code 1 == this.n} in the run() loop. */
    private boolean loadingStepIsOne() {
        return this.loadingStep == 1;
    }

    /** No-op profiling stub: {@code c(int)} sibling used by drawCenteredString. */
    public final void profileC(int dummy) {
        // pure profiling no-op
    }

    /** Game hook: one-time setup after the loading screen, overridden by the game subclass. */
    protected void startGame() {
    }

    /**
     * Game hook: per-loop input/logic tick, overridden by the game subclass. obf {@code e(int)} — the
     * int arg is a behavior selector consumed by the override (e.g. {@code if (arg < 64) clear scratch}).
     * The run loop passes {@code 119}; the shutdown path passes {@code 27} (only reachable as dead code).
     */
    protected synchronized void handleInputs(int marker) {
    }

    /** Game hook: invoked during shutdown, overridden by the game subclass. */
    protected void onClosing() {
    }

    /**
     * Game hook: per-frame render, overridden by the game subclass. obf {@code b(boolean)} — the run
     * loop always passes {@code false}; at the shell level the body is a dead profiling counter.
     */
    protected synchronized void draw(boolean dummy) {
    }

    /** Game hook: a key was pressed (character), overridden by the game subclass. obf {@code a(byte,int)} virtual. */
    protected void handleKeyPress(byte marker, int chr) {
    }

    /** Game hook: a mouse button went down, overridden by the game subclass. obf {@code a(int,int,int,int)} virtual. */
    protected void handleMouseDown(int x, int marker, int button, int y) {
    }

    // ==================================================================
    // Obfuscated string-table decoder (two-stage XOR). Reproduces the original `z(z(...))`.
    // ==================================================================

    /**
     * Stage 1: if the literal is a single char, XOR it with 0x68. (For multi-char literals this is a
     * pass-through.) obf {@code z(String) -> char[]}.
     */
    private static char[] preXor(String s) {
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 'h'); // 0x68
        }
        return chars;
    }

    /**
     * Stage 2: XOR each char by a position-dependent key cycling through {24, 3, 72, 8, 104}.
     * obf {@code z(char[]) -> String}.
     */
    private static String descramble(char[] chars) {
        for (int i = 0; i < chars.length; i++) {
            int key;
            switch (i % 5) {
                case 0:  key = 24;  break;
                case 1:  key = 3;   break;
                case 2:  key = 72;  break;
                case 3:  key = 8;   break;
                default: key = 104;
            }
            chars[i] = (char) (chars[i] ^ key);
        }
        return new String(chars).intern();
    }

    /**
     * Builds the obfuscated string table {@code Sb}. Entries are mostly error-context labels used by
     * the (now-stripped) per-method exception wrapper, plus a handful of real resource names and UI
     * strings. The scrambled literals are reproduced verbatim from the original class.
     */
    private static String[] decodeStringTable() {
        String[] scrambled = {
            "-#;mwm,{H}{8a}gd(wq+a##at",
            "}-,mlq'q@1", "}-M@", "vv$d", "c-f&", "}- ",
            "}-#mLz8m\f0", "}-%gkfd{h-l@", "}-/m_q)x q`; A",
            "}-%gkfz\rkp-l@", "}-8ivw`", "}-:}0*", "{q){ ",
            "}-M@", ")1&X63f9", "pw<x", "Kw)z}ghiho!k\tlj'f",
            "kf<N{v;\\yu-zyomkF&i\ntf,", "}-M@",
            "}-%gkfm}b;m\f0", "}-%gkf\rf}q-l@",
            "}-/mHb:i}w-z@", "}-M@", "±#z8X).z8Y-#i}{hD|",
            "[q-i}ghj8I\tO\r@#e(qp!|Hot?&yd-pF{l%", "}-;|\tjw`!",
            "}-#mHq-{}g`", "}-M@", "[o'{vdhxwd:i", "p2yx",
            "tl/gFld)", "p2zx", "Rb/m8o!jyq1", "}-M@", "p2~j", "p1xj",
            "p2{j", "p1|j", "p2|j", "p2zj", "}-=x\fyw- ", "}-/mYs8d\rl@'f}{< A",
            "Pf$~\rlj+i", "Lj%mJl%i", "Tl)lvd", "}-%gkfg}g`",
            "}-\tL@", "}-!{,qp8d\tab*d\r0*", "}-L@", "}-I@", "}- ",
            "}-/m\\l+}}m<J\tkf`!", "}-;|h+a", "}-M@",
            "}-8znj,m$wb,mYs8d\rl+", "}-M@", "Tl)lvdf&F", "}-\rM@",
            "}-#mJf$m\tkf, ", "tl/o\r|l=|", "Gw'x", "}-M@", "6t;",
            "}q:gGd)e\rG", "}-%gkf\rplf, ",
            "Mm)j}#<gHtl)lH{l&|\rvwhx\t{hh", "}- M@", "}-/mKj2m@1",
            "}-M@", "Kw)z}ghiho-|", "}-M@", "}-+z\ryw-Ayd- ",
            "}-I@", "}-K@", "}-M@", "}-M@", "}-%gkf\fz\td-l@"
        };
        String[] out = new String[scrambled.length];
        for (int i = 0; i < scrambled.length; i++) {
            out[i] = descramble(preXor(scrambled[i]));
        }
        return out;
    }

    /**
     * Helper type placeholder for the crypto/char-map class ({@code v}) referenced statically as
     * {@code ChatCipher.charMap}. Declared here only so this file reads as valid Java; the real type
     * lives in {@code client.net.ChatCipher}.
     */
    static final class ChatCipher {
        /** Printable input character set (letters, digits, punctuation, space). */
        static String charMap =
                "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
              + "!\"£$%^&*()-_=+[{]};:'@#~,<.>/?\\| ";
    }
}
