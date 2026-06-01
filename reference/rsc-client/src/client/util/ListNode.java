package client.util;

/**
 * ListNode — generic singly-linked task/request node used as the unit of work
 * in the {@code LoaderThread} (obf: {@code c}) background-worker queue.
 *
 * <p>Every call into the loader thread allocates one {@code ListNode}, sets its
 * {@link #type}, {@link #intArg}/{@link #intArg2}, and {@link #payload} fields,
 * then appends it to the FIFO queue ({@code c.a}/{@code c.o} are head/tail).
 * The loader thread's {@code run()} loop dequeues nodes, dispatches on
 * {@link #type}, writes the outcome into {@link #result}, sets {@link #status}
 * to 1 (done) or 2 (error), then {@code notify()}s the node so callers that
 * blocked on it wake up.
 *
 * <p>The class is also referenced by {@link DownloadWorker} (obf: {@code jb})
 * which holds a few live node references to track in-flight URL downloads.
 *
 * <p>There are no methods here; all logic lives in {@code LoaderThread.run()}.
 * The {@code volatile} modifier on {@link #status} and {@link #result} ensures
 * that threads reading the "done" state after the notify see the written values.
 *
 * <h3>Request type codes (field {@link #type})</h3>
 * <pre>
 *  -1  open raw Socket          payload=String host, intArg=port        → result: Socket
 *   1  open CacheFile (sync)    payload=String path, intArg=fileId       → result: CacheFile (d)
 *  22  open CacheFile (async)   same as 1, daemon background open
 *   2  start daemon Thread      payload=Runnable, intArg=priority         → result: Thread
 *   3  reverse DNS (packed IP)  intArg=packed int IP                      → result: String "a.b.c.d"
 *   4  URL.openStream           payload=URL                               → result: DataInputStream
 *   5  query DirectDraw modes   (payload unused)                          → result: int[]
 *   6  DNS → byte[] address     payload=String hostname                   → result: byte[4]
 *   7  hide/show cursor         payload=Component, intArg=0/1             → (no result)
 *   8  getDeclaredMethod        payload=Object[]{Class,String,Class[]}    → result: Method
 *  -9  getDeclaredField         payload=Object[]{Class,String}            → result: Field
 *  12  open CacheFile (alt-A)   payload=String path (alternative branch)  → result: CacheFile
 * -12  open CacheFile (alt-B)   payload=String path (alternative branch)  → result: CacheFile
 *  13  new Frame (fullscreen)   (title from static pool)                  → result: Frame
 *  14  set Win32 cursor pos     intArg=x, intArg2=y                       → (no result)
 *  15  movemouse (Robot)        intArg=x, intArg2=y                       → (no result)
 *  16  Runtime.exec             payload=String command                    → (no result)
 *  17  set cursor (DirectX)     payload=Object[]{Component,int[],…}       → (no result)
 *  18  get clipboard contents   (no args)                                 → result: Transferable
 *  19  set clipboard contents   payload=Transferable                      → (no result)
 *  21  hostname → InetAddress   payload=String hostname                   → result: InetAddress
 * </pre>
 *
 * <p>obf: {@code g}
 */
public class ListNode {

    // -------------------------------------------------------------------------
    // Queue linkage
    // -------------------------------------------------------------------------

    /**
     * Next node in the loader-thread's singly-linked FIFO queue.
     * The queue head is {@code LoaderThread.a}; the tail is {@code LoaderThread.o}.
     * Null when this node is the last (or only) item.
     *
     * obf: {@code g.a}
     */
    public ListNode next; // obf: a

    // -------------------------------------------------------------------------
    // Request parameters
    // -------------------------------------------------------------------------

    /**
     * Request type code; selects the operation executed by {@code LoaderThread.run()}.
     * See the class-level Javadoc for the full dispatch table.
     *
     * obf: {@code g.g}
     */
    public int type; // obf: g

    /**
     * Primary integer argument — port number, thread priority, packed IPv4 address,
     * frame width, or cursor X coordinate, depending on {@link #type}.
     *
     * Declared {@code public} in the original bytecode (read directly across
     * classes, e.g. by {@code DownloadWorker} / {@code LoaderThread}).
     *
     * obf: {@code g.e}
     */
    public int intArg; // obf: e

    /**
     * Secondary integer argument — frame height, cursor Y coordinate, or the
     * low 16-bits of a packed dimension word.  Not used by all request types.
     *
     * obf: {@code g.c}
     */
    public int intArg2; // obf: c

    /**
     * Object input payload — the request-specific object input.  Cast by the
     * loader thread according to {@link #type}: String (hostname/path/command),
     * URL, Runnable, java.awt.Component, java.awt.datatransfer.Transferable,
     * Object[] (for reflection requests), etc.
     *
     * Not declared volatile because it is written before the node is enqueued
     * (the enqueue itself is synchronized on the LoaderThread monitor) and is
     * only ever read by the loader thread worker — so the monitor guarantees
     * visibility.
     *
     * obf: {@code g.f}
     */
    public Object payload; // obf: f

    // -------------------------------------------------------------------------
    // Response / completion
    // -------------------------------------------------------------------------

    /**
     * Completion status, written by the loader thread after the operation
     * finishes.  Declared {@code volatile} so that callers polling
     * {@code while (~node.status == -1)} (i.e. {@code status == 0}) without
     * holding a lock see the write promptly.
     *
     * <ul>
     *   <li>0 = pending (initial value set by field initialiser)</li>
     *   <li>1 = completed successfully</li>
     *   <li>2 = completed with error (see {@link #result} for exception message)</li>
     * </ul>
     *
     * obf: {@code g.b}
     */
    public volatile int status = 0; // obf: b

    /**
     * Result object written by the loader thread after the operation completes.
     * Declared {@code volatile} for the same visibility reason as {@link #status}.
     * The concrete type depends on {@link #type}:
     * <ul>
     *   <li>Socket, Thread, DataInputStream, CacheFile ({@code d}),
     *       java.lang.reflect.Method/Field, int[], byte[], Frame, Transferable, …</li>
     *   <li>On error (status == 2) this may be null or hold an error-message String.</li>
     * </ul>
     *
     * obf: {@code g.d}
     */
    public volatile Object result; // obf: d

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /**
     * Default constructor — all fields left at their zero/null defaults.
     * Callers (exclusively {@code LoaderThread.a(int,int,byte,int,Object)})
     * set the fields directly after construction before enqueuing.
     *
     * obf: {@code g()}
     */
    public ListNode() {
        // intentionally empty; fields initialised by the factory in LoaderThread
    }
}
