package client.shell;

import client.data.CacheFile;      // obf: d
import client.data.CachePath;      // obf: r  (static helpers to resolve cache dirs)
import client.nativeapi.DirectDrawModes;  // obf: wa
import client.nativeapi.Win32MouseCallback; // obf: q
import client.net.StreamFactory;   // obf: na
import client.util.ListNode;       // obf: g
import client.util.Timer;          // obf: p

import java.awt.Component;
import java.awt.EventQueue;
import java.awt.Frame;
import java.awt.Point;
import java.awt.Toolkit;
import java.awt.datatransfer.Clipboard;
import java.awt.datatransfer.Transferable;
import java.io.DataInputStream;
import java.io.File;
import java.io.IOException;
import java.lang.reflect.Method;
import java.net.InetAddress;
import java.net.Socket;
import java.net.URL;

/**
 * LoaderThread (obf: c) — the privileged I/O and OS-bridge worker thread for
 * the RuneScape Classic client (rev ~233-235, Microsoft J++ / Windows build).
 *
 * <h2>Role in the engine</h2>
 * Many sensitive operations (socket connects, Thread spawning, reflection, fullscreen
 * setup, cursor hooking, clipboard access) are marshalled through this single
 * background thread to keep the actual game code free of checked exceptions and to
 * let access-control checks be centralised here.  Callers submit a {@link ListNode}
 * "work item" (with an opcode in {@code node.opcode} and arguments in
 * {@code node.intArg}/{@code node.obj}) and then wait on the node's monitor for the
 * result, which is placed in {@code node.result}.
 *
 * <h2>Platform branches</h2>
 * At construction time the class detects whether it is running under the Microsoft JVM
 * (by testing {@code java.vendor} for "microsoft", stored in {@link #isWin32}).
 * When on the Microsoft JVM it uses native {@link DirectDrawModes} and
 * {@link Win32MouseCallback}; otherwise it falls back to a reflective path through the
 * class {@code j} (RobotCursor, obf: j) and the AWT display-mode helper
 * {@code ha} (DisplayModeSetter, obf: ha) loaded by class-name.
 *
 * <h2>Cache layout opened here</h2>
 * <ul>
 *   <li>{@link #seedFile} – {@code random.dat} (PRNG seed, 25 bytes)</li>
 *   <li>{@link #dataFile} – {@code main_file_cache.dat2} (300 MB budget)</li>
 *   <li>{@link #indexFile255} – {@code main_file_cache.idx255} (master index, 1 MB)</li>
 *   <li>{@link #indexFiles}{@code [i]} – {@code main_file_cache.idx}<i>i</i> (1 MB each)</li>
 * </ul>
 *
 * <h2>Obfuscation stripped</h2>
 * <ul>
 *   <li>XOR string pool {@code z[]} decoded and inlined as literals.</li>
 *   <li>Opaque-predicate dead branches removed.</li>
 *   <li>Per-method profiling increments ({@code ++StaticCounter}) deleted.</li>
 *   <li>Exception wrappers unwrapped to bare body.</li>
 *   <li>Anti-tamper magic-constant guards on dummy params dropped.</li>
 *   <li>Junk shift-constant obfuscation reduced (e.g. {@code >> -182496008} → {@code >> 24}).</li>
 * </ul>
 */
final class LoaderThread implements Runnable {  // obf: c

    // -----------------------------------------------------------------------
    // Decoded XOR string pool  (static final String[] z — obf field name: z)
    // Two-pass decode: z(String)->char[] XORs with 0x42 if length<2;
    //                  z(char[])->String XORs each char with key[i%5] where
    //                  key = {6, 41, 18, 97, 66}.
    // -----------------------------------------------------------------------
    // z[ 0] = "jagex_"
    // z[ 1] = "c:/windows/"
    // z[ 2] = "/rscache/"
    // z[ 3] = "c:/"
    // z[ 4] = ".dat"
    // z[ 5] = "c:/rscache/"
    // z[ 6] = "_preferences"
    // z[ 7] = "_wip.dat"
    // z[ 8] = "rw"           (RandomAccessFile mode)
    // z[ 9] = "c:/winnt/"
    // z[10] = "/tmp/"
    // z[11] = "_rc.dat"
    // z[12] = "Jagex Full Screen"
    // z[13] = "exit"
    // z[14] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789?&=,.%+-_#:/*"
    // z[15] = "ha"            (class name: DisplayModeSetter)
    // z[16] = "https://"
    // z[17] = "enter"         (method name on DisplayModeSetter: enterFullscreen)
    // z[18] = "http://"
    // z[19] = "movemouse"     (method name on RobotCursor)   [NOTE: last char garbled in decode — see uncertainties]
    // z[20] = "setcustomcursor"
    // z[21] = "cmd /c start \"j\" \""   (shell-exec prefix, Windows only)
    // z[22] = "listmodes"
    // z[23] = "win"
    // z[24] = "showcursor"
    // z[25] = "setFocusTraversalKeysEnabled"
    // z[26] = "main_file_cache.idx255"
    // z[27] = "rb"            (class name: DirectSoundPlayer — loaded by name, ignored if unavailable)
    // z[28] = "java.awt.Container"
    // z[29] = "os.name"
    // z[30] = "os.version"
    // z[31] = "user.home"
    // z[32] = "java.vendor"
    // z[33] = "main_file_cache.dat2"
    // z[34] = "os.arch"
    // z[35] = "random.dat"
    // z[36] = "java.version"
    // z[37] = "~/"
    // z[38] = "setFocusCycleRoot"
    // z[39] = "Unknown"
    // z[40] = "java.awt.Component"
    // z[41] = "main_file_cache.idx"
    // z[42] = "1.1"
    // z[43] = "microsoft"

    // -----------------------------------------------------------------------
    // Work-queue opcodes used in run() dispatch
    // (value stored in ListNode.opcode / g.g)
    // -----------------------------------------------------------------------
    /** opcode 1  — open TCP Socket to (host String, port int) */
    private static final int OP_OPEN_SOCKET     =  1;
    /** opcode 2  — spawn a daemon Thread at a given priority */
    private static final int OP_START_THREAD    =  2;
    /** opcode 3  — reverse-DNS: int-packed IPv4 → hostname String */
    private static final int OP_REVERSE_DNS     =  3;
    /** opcode 4  — open a DataInputStream from a URL */
    private static final int OP_OPEN_URL        =  4;
    /** opcode 5  — list DirectDraw display modes (or via reflection) */
    private static final int OP_LIST_MODES       =  5;
    /**
     * opcode 6  — enter fullscreen: create a borderless Frame and set display mode.
     * Bytecode: ~n == -7, if_icmpeq → jumps to new Frame(...) block.
     */
    private static final int OP_ENTER_FULLSCREEN =  6;   // ~n == -7
    /** opcode 7  — exit fullscreen / restore display (Frame arg) */
    private static final int OP_EXIT_FULLSCREEN  =  7;
    /** opcode 8  — get declared method via reflection */
    private static final int OP_GET_METHOD       =  8;
    /** opcode 9  — get declared field via reflection  (~n == -10 → n == 9) */
    private static final int OP_GET_FIELD        =  9;
    /**
     * opcode 12 — open CacheFile using cacheBasePath as prefix directory.
     * (~n == -13 → n == 12)
     */
    private static final int OP_OPEN_CACHE_PREFIXED = 12;
    /** opcode 13 — open CacheFile for named store (no prefix dir). (~n == -14 → n == 13) */
    private static final int OP_OPEN_CACHE_BARE  = 13;
    /**
     * opcode 14 — move cursor to (x,y).  Only dispatched when {@link #isWin32} is true.
     * (Bytecode: isWin32 AND n2 == 14 → SetCursorPos via Win32MouseCallback.)
     */
    private static final int OP_MOVE_CURSOR      = 14;
    /**
     * opcode 15 — show or hide the OS cursor.  Only dispatched when {@link #isWin32} is true.
     * (~n == -16 → n == 15)
     */
    private static final int OP_SHOW_CURSOR      = 15;
    /** opcode 16 — shell-exec a URL string (Windows only, strict allow-list) */
    private static final int OP_SHELL_EXEC       = 16;
    /** opcode 17 — set custom cursor image via RobotCursor (non-Win32 path) */
    private static final int OP_SET_CURSOR       = 17;
    /** opcode 18 — get clipboard contents */
    private static final int OP_GET_CLIPBOARD    = 18;
    /** opcode 19 — set clipboard contents */
    private static final int OP_SET_CLIPBOARD    = 19;
    /** opcode 21 — DNS lookup: String hostname → byte[] address */
    private static final int OP_DNS_LOOKUP       = 21;
    /** opcode 22 — open TCP Socket via proxy/SSL (StreamFactory path) */
    private static final int OP_OPEN_PROXY_SOCK  = 22;

    // -----------------------------------------------------------------------
    // Static / shared state
    // -----------------------------------------------------------------------

    /**
     * Cache-store subdirectory index used by {@link CachePath#getCacheFile(int,String)}.
     * Set to the constructor {@code storeIndex} param.
     * obf: b
     */
    static int cacheStoreIndex;                   // obf: b

    /**
     * Java vendor string (lower-cased), used to detect Microsoft JVM.
     * obf: g  (NOTE: same letter as class g = ListNode — different namespace)
     */
    private static String javaVendorLower;        // obf: g

    /**
     * Java version string from {@code System.getProperty("java.version")}.
     * obf: k
     */
    static String javaVersion;                    // obf: k

    /**
     * OS name string from {@code System.getProperty("os.name")}.
     * obf: h
     */
    private static String osName;                 // obf: h

    /**
     * Cache base path prefix: value of {@code user.home} with trailing "/",
     * or {@code "~/"} if the property is absent.
     * obf: m
     */
    private static String userHomePath;           // obf: m

    /**
     * Cache prefix path supplied by caller (e.g. applet codebase dir).
     * obf: p
     */
    private static String cacheBasePath;          // obf: p

    /**
     * Static query-string / URL parameter scratch.  Set to {@code "Unknown"} at
     * construction, then overridden by {@code System.getProperty("java.vendor")}.
     * obf: q
     */
    static String vendorQuery;                    // obf: q

    /**
     * Monotonically-increasing timestamp anchor; compared against
     * {@link Timer#currentTimeMillis(int)} to enforce cooldown windows on certain
     * blocking operations (reverse-DNS, socket-open).
     * obf: d  (volatile long — note the obf name 'd' is also class d = CacheFile;
     *          this is a field of LoaderThread, not a reference to CacheFile)
     */
    private static volatile long lastOpTimestamp = 0L;  // obf: d

    /**
     * Reflected {@link Method} for {@code java.awt.Component.setFocusTraversalKeysEnabled(boolean)}.
     * Used by GameShell to suppress Tab/Shift-Tab focus traversal.
     * obf: u
     */
    static Method setFocusTraversalKeysEnabled;   // obf: u

    /**
     * Reflected {@link Method} for {@code java.awt.Container.setFocusCycleRoot(boolean)}.
     * obf: y
     */
    static Method setFocusCycleRoot;              // obf: y

    // -----------------------------------------------------------------------
    // Instance state
    // -----------------------------------------------------------------------

    /** True once {@link #shutdown()} has been called; causes run() to exit. obf: e */
    private boolean stopped;                      // obf: e

    /**
     * Win32 mouse/cursor native callback (Microsoft JVM path).
     * Non-null only when {@link #isWin32} is true.
     * obf: t
     */
    private Win32MouseCallback win32MouseCallback; // obf: t  (type: q = Win32MouseCallback)

    /** Background worker thread (runs this Runnable). obf: r */
    private Thread workerThread;                   // obf: r

    /**
     * RobotCursor instance loaded by reflection ({@code Class.forName("j").newInstance()}).
     * Used on the non-Win32 path for custom cursor and cursor show/hide.
     * obf: x
     */
    private Object robotCursor;                    // obf: x

    /**
     * True when running on the Microsoft JVM (java.vendor contains "microsoft").
     * Selects between Win32-native paths (DirectDraw, Win32MouseCallback) and the
     * AWT-reflection fallback path.
     * obf: j (instance boolean — different from the class j = RobotCursor)
     */
    private boolean isWin32;                       // obf: j

    /**
     * DisplayModeSetter instance ({@code ha}) loaded by reflection — non-Win32 path.
     * Provides AWT {@code GraphicsDevice.setDisplayMode()} fullscreen.
     * obf: i
     */
    private Object displayModeSetter;              // obf: i

    /** AWT system event queue, captured at construction time. obf: n */
    EventQueue systemEventQueue;                   // obf: n

    /** True when running on the Microsoft JVM (Win32 native path active). obf: c */
    private boolean isDirectDraw;                  // obf: c
    // NOTE: isWin32 (field j) and isDirectDraw (field c) are set together — both are
    // true iff java.vendor.toLowerCase().contains("microsoft").

    /**
     * DirectDraw display-mode enumerator and fullscreen helper.
     * Non-null only when {@link #isDirectDraw} is true.
     * obf: w
     */
    private DirectDrawModes directDraw;            // obf: w  (type: wa = DirectDrawModes)

    /** Random seed cache file ({@code random.dat}, max 25 bytes). obf: s */
    CacheFile seedFile;                            // obf: s  (type: d = CacheFile)

    /** Main data file ({@code main_file_cache.dat2}, max 300 MB). obf: f */
    CacheFile dataFile;                            // obf: f  (type: d = CacheFile)

    /** Master index file ({@code main_file_cache.idx255}, max 1 MB). obf: v */
    CacheFile indexFile255;                        // obf: v  (type: d = CacheFile)

    /** Per-archive index files ({@code main_file_cache.idx0..n}, 1 MB each). obf: l */
    private CacheFile[] indexFiles;               // obf: l  (type: d[] = CacheFile[])

    // Work-queue: singly-linked list of ListNode items.
    /** Head of the pending-work queue (next item to dequeue). obf: a (instance g field) */
    private ListNode queueHead;                    // obf: a  (type: g = ListNode)
    /** Tail of the pending-work queue (last item enqueued). obf: o */
    private ListNode queueTail;                    // obf: o  (type: g = ListNode)

    // -----------------------------------------------------------------------
    // Constructor
    // -----------------------------------------------------------------------

    /**
     * Initialise the loader, probe the runtime environment, open cache files,
     * and start the background dispatch thread.
     *
     * @param storeIndex   cache store index passed to {@link CachePath} methods
     * @param basePath     applet/application cache base path (stored in {@link #cacheBasePath})
     * @param numIndexFiles number of per-archive index files to open
     * @param useWin32Native true if the caller knows we are on a Microsoft JVM
     *                       (normally re-detected here from {@code java.vendor})
     */
    LoaderThread(int storeIndex, String basePath, int numIndexFiles, boolean useWin32Native) throws Exception {
        // obf: c(int, String, int, boolean)
        // Anti-tamper dummy params removed (none in this constructor).
        // Profiling counter ++staticCounter removed.

        this.stopped    = false;
        this.isDirectDraw = false;
        this.indexFile255 = null;
        this.queueTail  = null;

        // z[42] = "1.1",  z[39] = "Unknown"
        javaVersion     = "1.1";          // obf: k = z[42]
        cacheStoreIndex = storeIndex;     // obf: b = var1
        vendorQuery     = "Unknown";      // obf: q = z[39]
        cacheBasePath   = basePath;       // obf: p = var2
        this.isWin32    = useWin32Native; // obf: j = var4

        // Probe system properties; failures are silently swallowed.
        try {
            // java.vendor → detect Microsoft JVM
            vendorQuery = System.getProperty("java.vendor");   // z[32]
            // java.version
            javaVersion = System.getProperty("java.version");  // z[36]
        } catch (Exception ignored) {}

        // If java.vendor (lower-cased) contains "microsoft" → Win32 native path.
        // ~indexOf("microsoft") == 0  means indexOf returned -1 (not found), so
        // the condition breaks out WITHOUT setting isDirectDraw — the fall-through sets it.
        if (vendorQuery.toLowerCase().indexOf("microsoft") != -1) {  // z[43]
            this.isDirectDraw = true;
            // isWin32 (field j) is also used as a security gate for reflection ops
            // and is set from the constructor argument; isDirectDraw mirrors it at runtime.
        }

        try {
            osName       = System.getProperty("os.name");      // z[29]
        } catch (Exception ignored) {
            osName       = "Unknown";                          // z[39]
        }
        javaVendorLower = osName.toLowerCase(); // obf: g = h.toLowerCase()

        // Read os.version and os.arch for diagnostics (results unused here).
        try { System.getProperty("os.arch").toLowerCase(); }    catch (Exception ignored) {} // z[34]
        try { System.getProperty("os.version").toLowerCase(); } catch (Exception ignored) {} // z[30]

        // Resolve user home directory for cache path.
        try {
            userHomePath = System.getProperty("user.home");    // z[31]
            if (userHomePath != null) {
                userHomePath = userHomePath + "/";
            }
        } catch (Exception ignored) {}

        if (userHomePath == null) {
            userHomePath = "~/";  // z[37]
        }

        // Capture AWT system event queue (used by GameShell for event pumping).
        try {
            systemEventQueue = Toolkit.getDefaultToolkit().getSystemEventQueue();
        } catch (Throwable ignored) {}

        // On non-Win32 path: look up focus-management methods by reflection so we
        // can disable keyboard tab-traversal from the game without a compile-time dep.
        if (!this.isDirectDraw) {
            try {
                // java.awt.Component.setFocusTraversalKeysEnabled(boolean)  — z[40], z[25]
                setFocusTraversalKeysEnabled = Class.forName("java.awt.Component")
                        .getDeclaredMethod("setFocusTraversalKeysEnabled", boolean.class);
            } catch (Exception ignored) {}

            try {
                // java.awt.Container.setFocusCycleRoot(boolean)  — z[28], z[38]
                setFocusCycleRoot = Class.forName("java.awt.Container")
                        .getDeclaredMethod("setFocusCycleRoot", boolean.class);
            } catch (Exception ignored) {}
        }

        // Notify CachePath of the store index + base path so it can build file paths.
        // obf: r.a(b, (byte)101, p)  — CachePath.init(storeIndex, guard, basePath)
        // The (byte)101 is an anti-tamper guard inside CachePath.a; we pass it through.
        CachePath.init(cacheStoreIndex, (byte)101, cacheBasePath);

        // Open cache files (only when the full Win32 native subsystem is active).
        if (this.isWin32) {
            // random.dat  — entropy seed, tiny (25 bytes max)
            // r.a(b, null, z[35], 0) = CachePath.resolveFile(storeIndex, null, "random.dat", 0)
            this.seedFile = new CacheFile(
                    CachePath.resolveFile(cacheStoreIndex, null, "random.dat", 0),
                    "rw", 25L);

            // main_file_cache.dat2  — main data blob, 300 MB budget
            // r.a(2, z[33]) = CachePath.getCacheFile(2, "main_file_cache.dat2")
            this.dataFile = new CacheFile(
                    CachePath.getCacheFile(2, "main_file_cache.dat2"),
                    "rw", 314572800L);

            // main_file_cache.idx255  — master index, 1 MB
            this.indexFile255 = new CacheFile(
                    CachePath.getCacheFile(2, "main_file_cache.idx255"),
                    "rw", 1048576L);

            // main_file_cache.idx0..n  — per-archive index files, 1 MB each
            this.indexFiles = new CacheFile[numIndexFiles];
            for (int i = 0; i < numIndexFiles; i++) {
                // z[41] + i = "main_file_cache.idx" + i
                this.indexFiles[i] = new CacheFile(
                        CachePath.getCacheFile(2, "main_file_cache.idx" + i),
                        "rw", 1048576L);
            }

            // Attempt to load DirectSoundPlayer (obf: rb) by name to trigger
            // any static class initialisation it needs.  Failure silently ignored.
            if (this.isDirectDraw) {
                try {
                    Class.forName("rb").newInstance(); // z[27] = "rb" = DirectSoundPlayer
                } catch (Throwable ignored) {}
            }

            // Initialise the display-mode backend.
            try {
                if (this.isDirectDraw) {
                    // Win32 path: wrap DirectDraw
                    this.directDraw = new DirectDrawModes();  // obf: wa
                } else {
                    // AWT path: load DisplayModeSetter (obf: ha) by name
                    // z[15] = "ha"
                    this.displayModeSetter = Class.forName("ha").newInstance();
                }
            } catch (Throwable ignored) {}

            // Initialise the cursor/mouse backend.
            try {
                if (!this.isDirectDraw) {
                    // AWT path: load RobotCursor (obf: j) by name
                    this.robotCursor = Class.forName("j").newInstance();
                } else {
                    // Win32 path: native mouse callback
                    this.win32MouseCallback = new Win32MouseCallback();  // obf: q
                }
            } catch (Throwable ignored) {}
        }

        // Start the background dispatch thread at max priority.
        this.stopped = false;
        this.workerThread = new Thread(this);
        this.workerThread.setPriority(10);
        this.workerThread.setDaemon(true);
        this.workerThread.start();
    }

    // -----------------------------------------------------------------------
    // Background dispatch loop
    // -----------------------------------------------------------------------

    /**
     * Work-item dispatch loop.  Runs on {@link #workerThread}.
     *
     * The thread waits on {@code this} for items enqueued via {@link #submitWork}.
     * Each dequeued {@link ListNode} carries an opcode ({@code node.opcode}) and
     * argument fields.  On completion the result is stored in {@code node.result}
     * and the node is notified so the caller can wake.
     *
     * Bytecode was partially garbled by Vineflower; this reconstruction is from the
     * CFR output cross-referenced with the bytecode listing.
     */
    @Override
    public final void run() {
        // obf: run()  (implements Runnable)
        // Vineflower couldn't decompile run(); reconstructed from CFR + raw bytecode.
        while (true) {
            ListNode workItem;

            // Dequeue the next work item, waiting if the queue is empty.
            synchronized (this) {
                while (true) {
                    if (this.stopped) {
                        return; // shutdown() was called
                    }
                    if (this.queueHead != null) {
                        workItem = this.queueHead;
                        // Advance head pointer
                        this.queueHead = this.queueHead.next;  // obf: g.a = next ListNode
                        if (this.queueHead == null) {
                            this.queueTail = null; // queue now empty; clear tail too
                        }
                        break;
                    }
                    try {
                        this.wait();
                    } catch (InterruptedException ignored) {}
                }
            }

            // Dispatch on opcode stored in workItem.opcode (obf: g.g).
            try {
                dispatch(workItem);
                workItem.status = 1; // SUCCESS
            } catch (ThreadDeath td) {
                throw td;
            } catch (Throwable ex) {
                workItem.status = 2; // FAILED
            }

            // Wake the caller waiting on the work item.
            synchronized (workItem) {
                workItem.notify();
            }
        }
    }

    /**
     * Inner dispatch: execute the work described by {@code node} and store the
     * result in {@code node.result}.
     *
     * Opcodes are plain integers; see the OP_* constants above.  The heavy nesting
     * in the original bytecode is entirely due to obfuscation; here each opcode is
     * a clean case.
     */
    private void dispatch(ListNode node) throws Throwable {
        // obf: part of run() (the giant switch-like block inside the try in run())

        int opcode = node.opcode;  // obf: g.g

        // ------------------------------------------------------------------
        // opcode 1 (OP_OPEN_SOCKET): open a raw TCP socket.
        // node.obj  = host name (String)
        // node.intArg  = port
        // Checks lastOpTimestamp cooldown before connecting.
        // ------------------------------------------------------------------
        if (opcode == OP_OPEN_SOCKET) {
            if (Timer.currentTimeMillis(0) < lastOpTimestamp) {
                throw new IOException(); // rate-limited
            }
            node.result = new Socket(
                    InetAddress.getByName((String) node.obj),
                    node.intArg);
            return;
        }

        // ------------------------------------------------------------------
        // opcode 22 (OP_OPEN_PROXY_SOCK): open a socket via proxy/SSL.
        // node.intArg  = secondary port
        // node.obj     = host name (String)
        // Uses StreamFactory (obf: na) which may produce a ProxySocketFactory.
        // Port 4718 is the fixed Jagex login/game SSL port.
        // ------------------------------------------------------------------
        if (opcode == OP_OPEN_PROXY_SOCK) {
            if ((lastOpTimestamp ^ -1L) > (Timer.currentTimeMillis(0) ^ -1L)) {
                throw new IOException(); // rate-limited
            }
            try {
                node.result = StreamFactory.createSocketFactory(4718, node.intArg, (String) node.obj)
                                           .openSocket((byte) 50);
            } catch (client.util.ClientIOException ex) {
                // On ClientIOException store the message string and rethrow.
                node.result = ex.getMessage();
                throw ex;
            }
            return;
        }

        // ------------------------------------------------------------------
        // opcode 2 (OP_START_THREAD): spawn and start a daemon Thread.
        // node.obj     = Runnable to wrap
        // node.intArg  = Thread priority
        // ------------------------------------------------------------------
        if (opcode == OP_START_THREAD) {
            Thread t = new Thread((Runnable) node.obj);
            t.setDaemon(true);
            t.start();
            t.setPriority(node.intArg);
            node.result = t;
            return;
        }

        // ------------------------------------------------------------------
        // opcode 4 (OP_OPEN_URL): open a DataInputStream from a URL.
        // node.obj = URL
        // Checks lastOpTimestamp cooldown.
        // ------------------------------------------------------------------
        if (opcode == OP_OPEN_URL) {
            if (Timer.currentTimeMillis(0) < lastOpTimestamp) {
                throw new IOException();
            }
            node.result = new DataInputStream(((URL) node.obj).openStream());
            return;
        }

        // ------------------------------------------------------------------
        // opcode 8 (OP_GET_METHOD): getDeclaredMethod() via reflection.
        // node.obj = Object[]{Class targetClass, String methodName, Class[] paramTypes}
        // Security check: on Win32 (isWin32), bootstrap classes (null classloader)
        // are blocked UNLESS targetClass also has a null classloader (i.e. is a
        // bootstrap class itself).  That logic is actually inverted here vs. normal:
        // if (isWin32 && targetClass.getClassLoader() == null) throw SecurityException.
        // ------------------------------------------------------------------
        if (opcode == OP_GET_METHOD) {
            Object[] args = (Object[]) node.obj;
            Class<?> targetClass = (Class<?>) args[0];
            if (this.isWin32 && targetClass.getClassLoader() == null) {
                throw new SecurityException();
            }
            node.result = targetClass.getDeclaredMethod(
                    (String) args[1],
                    (Class<?>[]) args[2]);
            return;
        }

        // ------------------------------------------------------------------
        // opcode 9 (OP_GET_FIELD): getDeclaredField() via reflection.
        // node.obj = Object[]{Class targetClass, String fieldName}
        // Same security gate as OP_GET_METHOD.
        // ------------------------------------------------------------------
        if (opcode == OP_GET_FIELD) {
            Object[] args = (Object[]) node.obj;
            Class<?> targetClass = (Class<?>) args[0];
            if (this.isWin32 && targetClass.getClassLoader() == null) {
                throw new SecurityException();
            }
            node.result = targetClass.getDeclaredField((String) args[1]);
            return;
        }

        // ------------------------------------------------------------------
        // Guard: opcodes below this point require Win32 / trusted mode.
        // Bytecode throws Exception("") if !isWin32 and opcode falls through to here.
        // ------------------------------------------------------------------
        if (!this.isWin32) {
            throw new Exception("");
        }

        // ------------------------------------------------------------------
        // opcode 3 (OP_REVERSE_DNS): IPv4 int → hostname String.
        // node.intArg = 32-bit IP address (network byte order packed in int):
        //   octet0 = (intArg >> 24) & 0xFF
        //   octet1 = (intArg >> 16) & 0xFF
        //   octet2 = (intArg >>  8) & 0xFF   [original used 0xFFC0 mask — obf junk, simplified]
        //   octet3 =  intArg        & 0xFF
        // Checks lastOpTimestamp cooldown.
        // ------------------------------------------------------------------
        if (opcode == OP_REVERSE_DNS) {
            if (lastOpTimestamp > Timer.currentTimeMillis(0)) {
                throw new IOException();
            }
            // Reconstruct dotted-decimal from int-packed IP.
            // Original junk mask on octet2: (0xFFC0 & intArg) >> 8 — simplified to (intArg >> 8) & 0xFF.
            String dotted = ((node.intArg >> 24) & 0xFF) + "."
                          + ((node.intArg >> 16) & 0xFF) + "."
                          + ((node.intArg >>  8) & 0xFF) + "."
                          + ( node.intArg        & 0xFF);
            node.result = InetAddress.getByName(dotted).getHostName();
            return;
        }

        // ------------------------------------------------------------------
        // opcode 21 (OP_DNS_LOOKUP): hostname → byte[] raw address.
        // node.obj = host name (String)
        // Cooldown check (opposite direction: p.a(0) must be >= d).
        // ------------------------------------------------------------------
        if (opcode == OP_DNS_LOOKUP) {
            if (Timer.currentTimeMillis(0) < lastOpTimestamp) {
                throw new IOException();
            }
            node.result = InetAddress.getByName((String) node.obj).getAddress();
            return;
        }

        // ------------------------------------------------------------------
        // opcode 7 (OP_EXIT_FULLSCREEN): leave fullscreen, restore display.
        // node.obj = Frame to restore
        // ------------------------------------------------------------------
        if (opcode == OP_EXIT_FULLSCREEN) {
            if (this.isDirectDraw) {
                // Win32 path: restore DirectDraw display mode
                this.directDraw.restoreDisplayMode((Frame) node.obj, 0);
            } else {
                // AWT reflection path: DisplayModeSetter.exit()  — z[13] = "exit"
                Class.forName("ha")  // z[15]
                     .getMethod("exit")  // z[13]
                     .invoke(this.displayModeSetter);
            }
            return;
        }

        // ------------------------------------------------------------------
        // opcode 13 (OP_OPEN_CACHE_BARE): open a named CacheFile, no prefix dir.
        // node.obj = filename String (e.g. "jagex_preferences_foo.dat")
        // Uses cacheStoreIndex, empty prefix ("").
        // ------------------------------------------------------------------
        if (opcode == OP_OPEN_CACHE_BARE) {
            // static helper: a(storeIndex, prefixDir, filename, false)
            // prefixDir="" means no prefix subdirectory
            node.result = openPreferencesFile(cacheStoreIndex, "", false, (String) node.obj);
            return;
        }

        // ------------------------------------------------------------------
        // opcode 12 (OP_OPEN_CACHE_PREFIXED): open a named CacheFile using
        // cacheBasePath as the prefix directory.  (~n == -13 → n == 12)
        // node.obj = filename String (suffix portion)
        // ------------------------------------------------------------------
        if (opcode == OP_OPEN_CACHE_PREFIXED) {
            node.result = openPreferencesFile(cacheStoreIndex, cacheBasePath, false, (String) node.obj);
            return;
        }

        // ------------------------------------------------------------------
        // opcode 14 (OP_MOVE_CURSOR): move the mouse cursor to (x, y).
        // Only reachable when isWin32 is true (bytecode guard: if this.j && n2==14).
        // node.intArg = x coordinate  (obf: g.e)
        // node.auxArg = y coordinate  (obf: g.c)
        // ------------------------------------------------------------------
        if (this.isWin32 && opcode == OP_MOVE_CURSOR) {
            int x = node.intArg;  // obf: g.e
            int y = node.auxArg;  // obf: g.c
            if (this.isDirectDraw) {
                // Win32 path: Win32MouseCallback.moveMouse(0x5BE9, y, x)
                // 0x5BE9 = 23529 — Win32MouseCallback opcode for User32.SetCursorPos
                this.win32MouseCallback.moveMouse(23529, y, x);
            } else {
                // AWT reflection path: j.movemouse(x, y)  — z[19] = "movemouse"
                Class.forName("j")
                     .getDeclaredMethod("movemouse", Integer.TYPE, Integer.TYPE)  // z[19]
                     .invoke(this.robotCursor, x, y);
            }
            return;
        }

        // ------------------------------------------------------------------
        // opcode 15 (OP_SHOW_CURSOR): show or hide the OS cursor.
        // Only reachable when isWin32 is true (~n == -16 → n == 15).
        // node.intArg = 0 → hide, nonzero → show
        //   (original: ~g2.e != -1, i.e. g2.e != 0 → visible)
        // node.obj    = Component whose HWND is used on the Win32 path
        // ------------------------------------------------------------------
        if (this.isWin32 && opcode == OP_SHOW_CURSOR) {
            // visible = (node.intArg != 0), i.e. ~intArg != -1
            boolean visible = (node.intArg ^ -1) != -1; // node.intArg != 0
            Component comp  = (Component) node.obj;
            if (this.isDirectDraw) {
                // Win32 path: Win32MouseCallback.showCursor(-4, comp, visible)
                this.win32MouseCallback.showCursor(-4, comp, visible);
            } else {
                // AWT reflection path: j.showcursor(comp, visible)  — z[24]
                Class.forName("j")
                     .getDeclaredMethod("showcursor", Component.class, Boolean.TYPE)  // z[24]
                     .invoke(this.robotCursor, comp, visible);
            }
            return;
        }

        // ------------------------------------------------------------------
        // opcode 16 (OP_SHELL_EXEC): launch a URL or file path via cmd.exe.
        // SECURITY: restricted to Win32 ("win" prefix), http/https URLs only,
        // path sanitised against strict allow-list of chars.
        // node.obj = target URL/path String
        // ------------------------------------------------------------------
        if (opcode == OP_SHELL_EXEC) {
            // Must be running on Windows.
            if (!javaVendorLower.startsWith("win")) {  // z[23] = "win"
                throw new Exception();
            }
            String target = (String) node.obj;
            // Must start with http:// or https://.
            if (!target.startsWith("http://") && !target.startsWith("https://")) {  // z[18], z[16]
                throw new Exception();
            }
            // Every character must be in the URL allow-list.
            // z[14] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789?&=,.%+-_#:/*"
            String allowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789?&=,.%+-_#:/*";
            for (int i = 0; i < target.length(); i++) {
                if (allowedChars.indexOf(target.charAt(i)) == -1) {
                    throw new Exception();
                }
            }
            // Launch via Windows shell.  z[21] = "cmd /c start \"j\" \""
            Runtime.getRuntime().exec("cmd /c start \"j\" \"" + target + "\"");
            node.result = null;
            return;
        }

        // ------------------------------------------------------------------
        // opcode 17 (OP_SET_CURSOR): set a custom cursor image.
        // node.obj     = Object[]{Component comp, int[] hotspotOrPixels, Point hotspot}
        // node.intArg  = cursor width
        // node.auxArg  = cursor height  (obf: g.c)
        // Only available on non-Win32 AWT path (isDirectDraw = false).
        // ------------------------------------------------------------------
        if (!this.isDirectDraw && opcode == OP_SET_CURSOR) {
            Object[] args = (Object[]) node.obj;
            // j.setcustomcursor(comp, pixels, width, height, hotspot)  — z[20]
            Class.forName("j")
                 .getDeclaredMethod("setcustomcursor",
                         Component.class, int[].class, Integer.TYPE, Integer.TYPE, Point.class)
                 .invoke(this.robotCursor,
                         args[0],           // Component
                         args[1],           // int[] pixel data
                         node.intArg,       // width   (obf: g.e)
                         node.auxArg,       // height  (obf: g.c)
                         args[2]);          // Point hotspot
            return;
        }

        // ------------------------------------------------------------------
        // opcode 5 (OP_LIST_MODES): query available display modes.
        // Returns int[] packed as {width, height, bitDepth, refreshRate, ...}.
        // ------------------------------------------------------------------
        if (opcode == OP_LIST_MODES) {
            if (this.isDirectDraw) {
                // Win32 path: enumerate via DirectDraw
                node.result = this.directDraw.listModes((byte) -100);
            } else {
                // AWT reflection path: DisplayModeSetter.listmodes()  — z[22], z[15]
                node.result = Class.forName("ha")
                        .getMethod("listmodes")
                        .invoke(this.displayModeSetter);
            }
            return;
        }

        // ------------------------------------------------------------------
        // opcode 6 (OP_ENTER_FULLSCREEN): create a borderless Frame and enter
        // fullscreen at the requested display mode.
        // Bytecode: ~n2 == -7 (n2 == 6) → jumps forward to the new Frame(...) block.
        //
        // node.intArg  = packed (screenWidth << 16) | screenHeight  — obf: g.e
        // node.auxArg  = packed (displayX    << 16) | displayY      — obf: g.c
        //
        // Unpacking (all shifts mod 32; obfuscated junk constants reduced):
        //   width      = (node.intArg >>> 16) & 0xFFFF   [obf: g.e >>> -1397573296 → >>> 16]
        //   height     =  node.intArg          & 0xFFFF
        //   displayX   =  node.auxArg  >> 16             [obf: g.c  >> -747878896  →   >> 16]
        //   displayY   =  node.auxArg  & 0xFFFF
        // ------------------------------------------------------------------
        if (opcode == OP_ENTER_FULLSCREEN) {
            Frame frame = new Frame("Jagex Full Screen");  // z[12]
            node.result = frame;
            frame.setResizable(false);

            // Unpack dimensions from int args.
            int screenW   = (node.intArg >>> 16) & 0xFFFF;  // obf shift: -1397573296 mod 32 = 16
            int screenH   =  node.intArg          & 0xFFFF;
            int displayX  =  node.auxArg >> 16;              // obf shift: -747878896 mod 32 = 16
            int displayY  =  node.auxArg  & 0xFFFF;
            // NOTE: DirectDraw refreshRate/bitDepth packed into node.intArg high bits:
            //   bitDepth    = (node.intArg >>> 16) … same field, alternate interpretation in wa
            //   refreshRate =  node.intArg  & 0xFFFF
            // The DirectDraw path passes (frame, x, y, w, h, guard=77).

            if (this.isDirectDraw) {
                // Win32 path: DirectDrawModes.enterFullscreen(frame, x, y, w, h, guard)
                // obf args: g.c >> -747878896 = displayX, g.c & 0xFFFF = displayY,
                //           g.e & 0xFFFF = screenW, g.e >>> 831913136 = refreshRate (>>> 16)
                int refreshRate = (node.intArg >>> 16) & 0xFFFF;  // obf: g.e >>> 831913136 mod 32 = 16
                this.directDraw.enterFullscreen(frame, displayX, displayY, screenW, refreshRate, (byte) 77);
            } else {
                // AWT reflection path: DisplayModeSetter.enter(frame, w, h, x, y)  — z[17], z[15]
                // obf args: g.e >>> -1397573296 → 16 = width, g.e & 0xFFFF = height,
                //           g.c >> -1159913680 → 16 = displayX, g.c & 0xFFFF = displayY
                Class.forName("ha")
                     .getMethod("enter",       // z[17]
                             Frame.class, Integer.TYPE, Integer.TYPE, Integer.TYPE, Integer.TYPE)
                     .invoke(this.displayModeSetter,
                             frame,
                             screenW,
                             screenH,
                             displayX,
                             displayY);
            }
            return;
        }

        // ------------------------------------------------------------------
        // opcode 18 (OP_GET_CLIPBOARD): get the system clipboard contents.
        // ------------------------------------------------------------------
        if (opcode == OP_GET_CLIPBOARD) {
            Clipboard clipboard = Toolkit.getDefaultToolkit().getSystemClipboard();
            node.result = clipboard.getContents(null);
            return;
        }

        // ------------------------------------------------------------------
        // opcode 19 (OP_SET_CLIPBOARD): set the system clipboard contents.
        // node.obj = Transferable to place on clipboard
        // ------------------------------------------------------------------
        if (opcode == OP_SET_CLIPBOARD) {
            Transferable transferable = (Transferable) node.obj;
            Clipboard clipboard = Toolkit.getDefaultToolkit().getSystemClipboard();
            clipboard.setContents(transferable, null);
            return;
        }

        // Unrecognised opcode.
        throw new Exception("");
    }

    // -----------------------------------------------------------------------
    // Work-queue submission helpers
    // -----------------------------------------------------------------------

    /**
     * Submit a work item and return the {@link ListNode} handle immediately (async).
     * The caller may block on the node's monitor to wait for completion.
     * Appends to the tail of the queue and notifies the worker thread.
     *
     * @param opcode   operation code (one of the OP_* constants)
     * @param auxArg   secondary integer argument (obf: g.c)
     * @param guard    anti-tamper guard byte; must be -21 to proceed, else clears javaVersion
     * @param intArg   primary integer argument  (obf: g.e)
     * @param obj      object argument            (obf: g.f)
     * @return the submitted ListNode
     */
    private ListNode submitWork(int opcode, int auxArg, byte guard, int intArg, Object obj) {
        // obf: a(int, int, byte, int, Object)
        ListNode node = new ListNode();
        node.intArg = intArg;  // obf: g.e
        node.opcode = opcode;  // obf: g.g
        node.auxArg = auxArg;  // obf: g.c
        node.obj    = obj;     // obf: g.f

        synchronized (this) {
            if (this.queueTail == null) {
                // Queue was empty; head and tail both point to the new node.
                this.queueTail = this.queueHead = node;
            } else {
                this.queueTail.next = node;  // obf: g.a
                this.queueTail = node;
            }
            this.notify();
        }

        // Anti-tamper: guard byte must equal -21; otherwise null out javaVersion.
        // This is a canary — passing any other value is an obfuscation artifact.
        if (guard != -21) {
            javaVersion = null;  // obf: k = null
        }

        return node;
    }

    // -----------------------------------------------------------------------
    // Public-facing typed submission overloads
    // (Each strips an anti-tamper guard param or opaque-predicate from the
    //  obfuscated version and routes to submitWork.)
    // -----------------------------------------------------------------------

    /**
     * Submit a connect-socket or close request with a String argument.
     * When {@code useProxy} is false, sends opcode 1 (raw socket); when true, opcode 22 (proxy).
     *
     * @param intArg    port number
     * @param guard     anti-tamper guard (must be 81)
     * @param host      hostname String
     * @param useProxy  false = raw TCP, true = proxy/SSL path
     * obf: a(int, byte, String, boolean)
     */
    private ListNode submitSocketRequest(int intArg, byte guard, String host, boolean useProxy) {
        // guard must equal 81 (anti-tamper); if not, recurse with opcode 3/null to trigger error.
        if (guard != 81) {
            submitSocketRequest(3, (byte) -100, null, true);
        }
        int opcode = useProxy ? OP_OPEN_PROXY_SOCK : OP_OPEN_SOCKET;
        return submitWork(opcode, 0, (byte) -21, intArg, host);
    }

    /**
     * Open a TCP socket to the given host/port.
     * Returns a ListNode; caller blocks on it for the resulting Socket.
     *
     * @param intArg  port (or opaque large negative — guarded by n3 > -66 check)
     * @param host    hostname String
     * @param n3      anti-tamper: only proceeds if n3 <= -66
     * obf: a(String, int, int)
     */
    final ListNode openSocket(String host, int intArg, int n3) {
        if (n3 > -66) {
            return null; // anti-tamper guard
        }
        return submitSocketRequest(intArg, (byte) 81, host, false);
    }

    /**
     * Start a daemon Thread running the given Runnable at a specified priority.
     *
     * @param isDaemon   must be true; if false triggers an internal null-arg enqueue
     * @param runnable   the Runnable to wrap in a new Thread
     * @param priority   Thread priority (1-10)
     * obf: a(boolean, Runnable, int)
     */
    final ListNode startThread(boolean isDaemon, Runnable runnable, int priority) {
        if (!isDaemon) {
            // Anti-tamper: enqueue a null Runnable to signal unexpected state
            submitWork(-34, 71, (byte) 60, 103, null);
        }
        return submitWork(OP_START_THREAD, 0, (byte) -21, priority, runnable);
    }

    /**
     * Open a DataInputStream from a URL.
     *
     * @param guard must be 74; if not, sets {@link #lastOpTimestamp} = -110 (anti-tamper)
     * @param url   the URL to open
     * obf: a(byte, URL)
     */
    final ListNode openUrl(byte guard, URL url) {
        if (guard != 74) {
            lastOpTimestamp = -110L; // anti-tamper side-effect
        }
        return submitWork(OP_OPEN_URL, 0, (byte) -21, 0, url);
    }

    /**
     * Submit a reverse-DNS lookup (int-packed IPv4 → hostname).
     *
     * @param packedIp  32-bit packed IPv4 address
     * @param guard     junk divisor guard (any byte; junk expression 9/((−58−guard)/56) discarded)
     * obf: a(int, byte)
     */
    final ListNode reverseDns(int packedIp, byte guard) {
        // Junk expression: 9 / ((-58 - guard) / 56) — dead, removed.
        return submitWork(OP_REVERSE_DNS, 0, (byte) -21, packedIp, null);
    }

    // -----------------------------------------------------------------------
    // Static helper: open a preferences/cache file
    // -----------------------------------------------------------------------

    /**
     * Resolve and open a preferences (or .dat/.wip) CacheFile.
     *
     * Builds the file name as {@code "jagex_preferences_" + storeSubDir + "_" + suffix + ext}
     * where {@code ext} depends on {@code storeIndex}:
     * <ul>
     *   <li>33 ({@code ~33 == -34}): {@code _rc.dat}</li>
     *   <li>34 ({@code ~34 == -35}): {@code _wip.dat}</li>
     *   <li>other:                   {@code .dat}</li>
     * </ul>
     *
     * Then searches a priority list of candidate directories for the file and opens it
     * with a 10 000-byte limit.  Returns {@code null} if no candidate directory works.
     *
     * @param storeIndex  controls the file extension (33/34/other)
     * @param storeSubDir prefix substring (e.g. "")
     * @param mustExist   if true, skip file creation and return null immediately
     * @param suffix      the unique part of the filename (e.g. applet codebase host)
     * obf: a(int, String, boolean, String)
     */
    private static CacheFile openPreferencesFile(int storeIndex, String storeSubDir,
                                                  boolean mustExist, String suffix) {
        // Build filename:  "jagex_" + storeSubDir + "_preferences" + suffix + ext
        // z[0]="jagex_", z[6]="_preferences", z[4]=".dat", z[7]="_wip.dat", z[11]="_rc.dat"
        String ext;
        if (~storeIndex == -34) {      // storeIndex == 33
            ext = "_rc.dat";           // z[11]
        } else if (~storeIndex == -35) { // storeIndex == 34
            ext = "_wip.dat";          // z[7]
        } else {
            ext = ".dat";              // z[4]
        }
        String filename = "jagex_" + storeSubDir + "_preferences" + suffix + ext;
        // z[0] + storeSubDir + z[6] + suffix + ext

        if (mustExist) {
            return null;
        }

        // Candidate directories in priority order:
        // z[5]="c:/rscache/", z[2]="/rscache/", userHomePath, z[1]="c:/windows/",
        // z[9]="c:/winnt/",   z[3]="c:/",        z[10]="/tmp/", ""
        String[] candidates = {
            "c:/rscache/",   // z[5]
            "/rscache/",     // z[2]
            userHomePath,    // m (resolved from user.home)
            "c:/windows/",   // z[1]
            "c:/winnt/",     // z[9]
            "c:/",           // z[3]
            "/tmp/",         // z[10]
            ""               // current directory
        };

        for (String dir : candidates) {
            // Skip non-empty dirs that do not exist.
            if (dir.length() > 0 && !new File(dir).exists()) {
                continue;
            }
            try {
                return new CacheFile(new File(dir, filename), "rw", 10000L);
            } catch (Exception ignored) {
                // This directory didn't work; try next.
            }
        }
        return null;
    }

    // -----------------------------------------------------------------------
    // XOR string-pool decoding helpers  (obf: z — two overloads)
    // These are only used during static initialisation of the z[] pool;
    // they are preserved here for reference but not called at runtime.
    // -----------------------------------------------------------------------

    /**
     * First decode pass: XOR a single-char string with 0x42 ('B'); strings of
     * length >= 2 are returned as-is (char array).
     * obf: z(String) → char[]
     */
    private static char[] xorDecodeStage1(String encoded) {
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 0x42);
        }
        return chars;
    }

    /**
     * Second decode pass: XOR each character with a rotating 5-byte key
     * {6, 41, 18, 97, 66}, then intern the result.
     * obf: z(char[]) → String
     */
    private static String xorDecodeStage2(char[] chars) {
        // Key table: key[i % 5]
        int[] key = {6, 41, 18, 97, 66};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char) (chars[i] ^ key[i % 5]);
        }
        return new String(chars).intern();
    }
}
