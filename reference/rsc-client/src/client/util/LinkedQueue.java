package client.util;
import client.net.StreamBase;

/**
 * LinkedQueue -- intrusive doubly-linked circular list with a built-in iterator,
 * operating over {@link StreamBase} nodes.
 *
 * <h3>Data structure</h3>
 * <p>This is a classic <em>sentinel-headed circular doubly-linked list</em>.
 * The {@link #sentinel} node is a permanent {@link StreamBase} that acts as both
 * the logical head and tail anchor.  The two link fields in every {@link StreamBase}
 * are used directly (intrusive linking), so a node can live in at most one
 * {@link LinkedQueue} at a time.
 *
 * <p>Link field naming (canonical owner: {@code ib.java} / {@link StreamBase}; the
 * obf-to-name mapping {@code a=next, e=prev} is honored here to match StreamBase):
 * <ul>
 *   <li>{@code StreamBase.next} (obf {@code ib.a}) -- the "tailward" link (toward newer
 *       / last-inserted items)</li>
 *   <li>{@code StreamBase.prev} (obf {@code ib.e}) -- the "headward" link (toward older
 *       / first-inserted items)</li>
 * </ul>
 * After {@code n} calls to {@link #enqueue}:
 * <pre>
 *   sentinel.prev (headward) -&gt; first_enqueued -&gt; ... -&gt; last_enqueued -&gt; sentinel
 *   sentinel.next (tailward) -&gt; last_enqueued  -&gt; ... -&gt; first_enqueued -&gt; sentinel
 * </pre>
 * {@link #peekHead} returns {@code sentinel.next} (the <em>last</em> enqueued item),
 * so iteration proceeds in <em>LIFO / reverse-insertion</em> order.  This matches the
 * audio-engine usage in {@link client.audio.StreamMixer} (obf {@code ra}), where the
 * most recently added voice or stage is processed first.
 *
 * <h3>Usage in the engine (rev ~233-235)</h3>
 * <ul>
 *   <li>{@link client.audio.StreamMixer} (obf {@code ra}) holds two {@code LinkedQueue}
 *       instances: one for child {@link client.audio.FilterChain} (obf {@code va}) voices,
 *       and one for scheduled {@link client.audio.FilterStage} (obf {@code hb}) filter
 *       stages.  The stage queue is maintained in sorted order by timing value via
 *       {@code DecodeBuffer.insertNodeBefore} (obf {@code ac.a(ib,byte,ib)}).</li>
 * </ul>
 *
 * <h3>Obfuscation artifacts stripped</h3>
 * <ul>
 *   <li>Per-method profiling counter increments ({@code ++h}, {@code ++c}, {@code ++a},
 *       {@code ++e}) removed; the counter fields are retained as {@code DEAD_COUNTER_*}
 *       statics so any cross-class references still compile.</li>
 *   <li>Exception wrappers ({@code catch (RuntimeException e) { throw ErrorHandler.a(e, "sig"); }})
 *       unwrapped to bare method bodies.</li>
 *   <li>Anti-tamper junk modulo expressions in {@link #peekHead(byte)} and
 *       {@link #iterateNext(byte)} deleted:
 *       {@code 119 % ((dummyByte + 37) / 43)} (peekHead) and
 *       {@code 81  % ((-37 - dummyByte) / 51)} (iterateNext) -- results were never used.</li>
 *   <li>{@link #isSeparatorChar(int,char)} was placed in this class by the obfuscator
 *       to inflate the method count; it belongs logically to text-layout code.</li>
 *   <li>XOR string pool decoded; literal values documented in comments below.</li>
 * </ul>
 *
 * <p>Original obfuscated class name: {@code db}
 */
public final class LinkedQueue {

    // =========================================================================
    // Dead profiling counter fields  obf: h, c, a, e, g
    // The obfuscator increments a distinct static int at every method entry.
    // These fields are not read for any real purpose; retained for link compatibility.
    // =========================================================================

    /** Dead profiling counter -- incremented on every {@link #enqueue} call.      obf: h */
    public static int DEAD_COUNTER_ENQUEUE;   // obf: h

    /** Dead profiling counter -- incremented on every {@link #peekHead} call.     obf: c */
    public static int DEAD_COUNTER_PEEK;      // obf: c

    /** Dead profiling counter -- incremented on every {@link #iterateNext} call.  obf: a */
    public static int DEAD_COUNTER_ITERATE;   // obf: a

    /** Dead profiling counter -- incremented on every {@link #isSeparatorChar} call. obf: e */
    public static int DEAD_COUNTER_IS_SEP;    // obf: e

    /**
     * Unused static int, initialised to 0 in the static initialiser.
     * No read site was found anywhere in the client.  obf: g
     */
    public static int DEAD_G = 0;             // obf: g

    // =========================================================================
    // Other static fields from the obfuscated class
    // (Present in the original bytecode; their consumers are in other classes.)
    // =========================================================================

    /** Shared int array -- purpose determined by callers in other classes.  obf: j */
    public static int[] sharedIntArray;       // obf: j

    /** Shared int -- purpose determined by callers in other classes.  obf: d */
    public static int sharedInt;              // obf: d

    /** Secondary shared int array.  obf: l */
    public static int[] sharedIntArray2;      // obf: l

    /** Secondary shared byte array.  obf: i */
    public static byte[] sharedByteArray;     // obf: i

    /**
     * Shared reference to the {@link ErrorHandler} singleton (obf class {@code i}).
     * Other classes read this field to obtain a handle on the error-handling utility.
     * obf: f
     */
    public static client.util.ErrorHandler errorHandler;  // obf: f

    // =========================================================================
    // Instance fields
    // =========================================================================

    /**
     * Sentinel node -- a permanent {@link StreamBase} that anchors the circular list.
     * <pre>
     *   sentinel.prev (headward) -&gt; first_enqueued -&gt; ... -&gt; last_enqueued -&gt; sentinel
     *   sentinel.next (tailward) -&gt; last_enqueued  -&gt; ... -&gt; first_enqueued -&gt; sentinel
     * </pre>
     * When the list is empty: {@code sentinel.next == sentinel}
     * and {@code sentinel.prev == sentinel}.
     * obf: k
     */
    public final StreamBase sentinel = new StreamBase();  // obf: k

    /**
     * Iterator cursor -- points to the node that {@link #iterateNext} will return next.
     * Set to {@code null} when no iteration is in progress or the list is exhausted.
     * The cursor always holds {@code current.next} after each step, so the traversal
     * moves headward (from newest toward oldest).
     * obf: b
     */
    private StreamBase iterCursor;                 // obf: b

    // =========================================================================
    // Constructor
    // =========================================================================

    /**
     * Creates an empty {@code LinkedQueue}.
     * Initialises the sentinel to point to itself in both link directions, producing
     * the canonical empty circular doubly-linked list.
     *
     * <p>XOR string pool entry used in error reporting: z[5] = "db.&lt;init&gt;()"
     */
    public LinkedQueue() {
        // Circular empty list: sentinel <-> sentinel
        sentinel.prev = sentinel;   // obf: this.k.e = this.k   (ib.e = prev)
        sentinel.next = sentinel;   // obf: this.k.a = this.k   (ib.a = next)
    }

    // =========================================================================
    // Mutation -- append to tail
    // =========================================================================

    /**
     * Appends {@code node} to the <em>tail</em> of this queue (just before the sentinel
     * in the tailward direction), implementing an enqueue-at-back operation.
     *
     * <p>If {@code node} is already linked into another {@code LinkedQueue} (its
     * {@code prev} field is non-null), {@link StreamBase#unlinkSelf(int)} is called with
     * the magic constant {@code -27331} to detach it first, preventing cross-list
     * pointer corruption.
     *
     * <p>After the call:
     * <ul>
     *   <li>{@code sentinel.next} now points to {@code node} (it becomes the new tail).</li>
     *   <li>The next call to {@link #peekHead} will return {@code node}.</li>
     * </ul>
     *
     * @param node      the {@link StreamBase} node to append; must not be {@code null}
     * @param resetIter if {@code true}, advances the iterator cursor by one step
     *                  (calls {@link #iterateNext} with the magic byte {@code 78}).
     *                  Used by {@link client.audio.StreamMixer} to keep an in-progress
     *                  iteration valid after a concurrent insert into the same queue.
     *                  obf param: var2 / bl
     *
     * <p>XOR string pool: z[4] = "db.C(",  z[2] = "{...}",  z[3] = "null"
     *
     * obf method: {@code final void a(ib, boolean)} -- internal sig: "db.C("
     */
    public final void enqueue(StreamBase node, boolean resetIter) {
        // (obf profiling removed: ++h)

        // If this node is already in some list, unlink it first.
        // StreamBase.unlinkSelf(-27331) is the self-removal method; -27331 is the magic
        // opcode it checks before performing the pointer surgery.
        if (node.prev != null) {             // obf: if (null != var1.e)   (ib.e = prev)
            node.unlinkSelf(-27331);         // obf: var1.a(-27331)
        }

        // Link node at the tail (just before sentinel in tailward direction):
        //
        //   Before: ... <- oldTail <- sentinel
        //                  oldTail -> sentinel
        //
        //   After:  ... <- oldTail <- node <- sentinel
        //                  oldTail -> node -> sentinel
        //
        node.prev       = sentinel;          // obf: var1.e = this.k     (ib.e = prev)
        node.next       = sentinel.next;     // obf: var1.a = this.k.a   (ib.a = next; = oldTail)
        node.prev.next  = node;              // obf: var1.e.a = var1     (sentinel.next = node)
        node.next.prev  = node;              // obf: var1.a.e = var1     (oldTail.prev  = node)

        if (resetIter) {
            iterateNext((byte) 78);          // obf: this.b((byte)78)
        }
    }

    // =========================================================================
    // Iteration -- two-phase: peekHead starts, iterateNext advances
    // =========================================================================

    /**
     * Begins an iteration pass: returns the <em>tail element</em> (the most recently
     * enqueued node) and primes the internal cursor for subsequent {@link #iterateNext}
     * calls.
     *
     * <p>Traversal order is <em>LIFO / reverse-insertion</em> (newest to oldest).
     *
     * <p>Returns {@code null} and clears the cursor when the queue is empty.
     *
     * <p>The {@code dummyByte} parameter is a pure anti-tamper artifact.  The original
     * bytecode computed {@code 119 % ((dummyByte + 37) / 43)} into a local that was
     * never read afterwards.  Callers in the engine pass arbitrary byte constants such
     * as {@code -87}, {@code 106}, {@code 35}, {@code -80}, {@code -88}.
     *
     * @param dummyByte anti-tamper operand; value is entirely ignored.  obf param: var1 / by
     * @return the last-enqueued (tail) node, or {@code null} if the queue is empty
     *
     * <p>XOR string pool: z[6] = "db.D("
     *
     * obf method: {@code final ib a(byte)} -- internal sig: "db.D("
     */
    public final StreamBase peekHead(byte dummyByte) {
        // (obf profiling removed: ++c)
        // (anti-tamper junk removed: int n = 119 % ((dummyByte + 37) / 43))

        StreamBase last = sentinel.next;     // obf: ib var2 = this.k.a   (ib.a = next)

        if (sentinel == last) {
            // sentinel.next == sentinel means the list is empty.
            iterCursor = null;               // obf: this.b = null
            return null;
        }

        // Prime the cursor one step headward of 'last' so that the first
        // iterateNext() call returns the second-most-recent item.
        iterCursor = last.next;              // obf: this.b = var2.a   (ib.a = next)
        return last;                         // obf: return var2
    }

    /**
     * Advances the iteration cursor and returns the next (older) node.
     *
     * <p>Must only be called after {@link #peekHead(byte)} has returned a non-null
     * value.  Each call moves one step headward through the list.  Returns {@code null}
     * and clears the cursor once the sentinel is reached (all elements have been visited).
     *
     * <p>The {@code dummyByte} parameter is an anti-tamper artifact exactly like the
     * one in {@link #peekHead}: the expression {@code 81 % ((-37 - dummyByte) / 51)}
     * was computed but its result discarded.  Callers pass arbitrary byte constants
     * such as {@code 111}, {@code 122}, {@code 115}, {@code 80}, {@code 78}.
     *
     * <p>Typical usage pattern (mirroring {@link client.audio.StreamMixer}):
     * <pre>
     *   for (FilterStage s = (FilterStage) queue.peekHead((byte) -87);
     *        s != null;
     *        s = (FilterStage) queue.iterateNext((byte) 111)) {
     *       s.process();
     *   }
     * </pre>
     *
     * @param dummyByte anti-tamper operand; value is entirely ignored.  obf param: var1 / by
     * @return the next older node, or {@code null} if iteration is complete
     *
     * <p>XOR string pool: z[1] = "db.A("
     *
     * obf method: {@code final ib b(byte)} -- internal sig: "db.A("
     */
    public final StreamBase iterateNext(byte dummyByte) {
        // (obf profiling removed: ++a)
        // (anti-tamper junk removed: int n = 81 % ((-37 - dummyByte) / 51))

        StreamBase current = iterCursor;     // obf: ib var2 = this.b

        if (sentinel == current) {
            // Cursor has reached the sentinel -- every real node has been visited.
            iterCursor = null;               // obf: this.b = null
            return null;
        }

        // Advance the cursor one step headward (toward older items).
        iterCursor = current.next;           // obf: this.b = var2.a   (ib.a = next)
        return current;                      // obf: return var2
    }

    // =========================================================================
    // Static helper -- word-wrap / text separator check
    // Placed in db.java by the obfuscator to scatter the method call graph.
    // Logically belongs to text-layout / rendering code.
    // =========================================================================

    /**
     * Returns {@code true} if {@code charClass == 32} and {@code ch} is one of the
     * characters that the RSC text renderer treats as a word-boundary separator:
     * <ul>
     *   <li>U+00A0 -- non-breaking space  (obf: {@code ~ch == -161})</li>
     *   <li>U+0020 -- plain space          (obf: {@code ch == ' '})</li>
     *   <li>U+005F -- underscore           (obf: {@code ~ch == -96})</li>
     *   <li>U+002D -- hyphen               (obf: {@code ~ch == -46})</li>
     * </ul>
     *
     * <p>The obfuscator uses the bitwise-NOT form {@code ~ch == -N} to hide literal
     * character comparisons: {@code ~ch == -N} is equivalent to {@code ch == (N - 1)}.
     *
     * <p>This method is called from {@link client.world.WorldEntity} (obf {@code w})
     * during entity-name / overhead-label rendering.
     *
     * @param charClass must equal {@code 32} for any {@code true} result; any other
     *                  value returns {@code false} immediately
     * @param ch        the character to test
     * @return {@code true} if {@code ch} is a word-break character under class 32
     *
     * <p>XOR string pool: z[0] = "db.B("
     *
     * obf method: {@code static final boolean a(int, char)} -- internal sig: "db.B("
     */
    public static final boolean isSeparatorChar(int charClass, char ch) {
        // (obf profiling removed: ++e)

        if (charClass != 32) {
            return false;
        }

        // Obfuscated form vs. plain equivalent:
        //   ~ch == -161  <=>  ch == 160  (U+00A0, non-breaking space)
        //   ch == ' '   direct            (U+0020, plain space)
        //   ~ch == -96   <=>  ch == 95   (U+005F, underscore '_')
        //   ~ch == -46   <=>  ch == 45   (U+002D, hyphen '-')

        if (ch == '\u00A0') return true;  // obf: if (~c2 == -161)  non-breaking space (U+00A0)
        if (ch == ' ')       return true;  // obf: if (c2 == ' ')    plain space        (U+0020)
        if (ch == '_')       return true;  // obf: if (~c2 == -96)   underscore         (U+005F)
        if (ch == '-')       return true;  // obf: if (~c2 == -46)   hyphen             (U+002D)

        return false;
    }

    // =========================================================================
    // XOR string-pool reference (decoded; not reproduced as executable code)
    // =========================================================================
    //
    // Original obfuscated field: private static final String[] z
    // Decoded values:
    //   z[0] = "db.B("        -- isSeparatorChar(int,char) error prefix
    //   z[1] = "db.A("        -- iterateNext(byte) error prefix
    //   z[2] = "{...}"        -- non-null StreamBase placeholder in enqueue error msg
    //   z[3] = "null"         -- null StreamBase placeholder in enqueue error msg
    //   z[4] = "db.C("        -- enqueue(StreamBase,boolean) error prefix
    //   z[5] = "db.<init>()"  -- constructor error prefix
    //   z[6] = "db.D("        -- peekHead(byte) error prefix
    //
    // Codec:
    //   z(String s)  -> char[]  (XOR s[0] ^= 0x17 if s.length() < 2; else no-op)
    //   z(char[] a)  -> String  (XOR a[i] by mod-5 key table {25, 111, 116, 16, 23})
}
