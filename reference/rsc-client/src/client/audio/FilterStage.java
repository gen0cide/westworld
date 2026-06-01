package client.audio;

import client.net.StreamBase;   // obf: ib — base resource/stream node class
import client.util.LinkedQueue; // obf: db

/**
 * FilterStage — abstract audio filter stage (obf: hb).
 *
 * <p>This is the base class for individual audio processing stages that are
 * scheduled by a {@link StreamMixer} (obf: ra).  Each concrete subclass
 * represents one "event" or "note" in the audio pipeline: it knows how many
 * output samples it can produce before it is done (or needs to be
 * rescheduled), and it knows how to reset / clean itself up when removed from
 * the mixer's active queue.
 *
 * <p>Architecture sketch:
 * <pre>
 *   StreamMixer (ra)
 *     owns LinkedQueue<FilterStage>  — sorted by nextTriggerSample (g)
 *     on each mix block:
 *       pops the head FilterStage, calls a(StreamMixer) → returns next trigger
 *       re-inserts if trigger >= 0, or removes if trigger < 0
 *       calls reset() before removal so the stage can release resources
 * </pre>
 *
 * <p>Inheritance:
 * <pre>
 *   StreamBase (ib)
 *     └─ FilterStage (hb)   ← this class
 * </pre>
 *
 * <p>Known concrete subclass: {@link SoundDecoder} (obf: sb) — extends
 * {@link FilterChain} (obf: va) which in turn is a sibling branch; FilterStage
 * itself does not inherit FilterChain. The NAMING.md map lists FilterStage as a
 * direct sibling of FilterChain under StreamBase.
 *
 * <p>Obfuscation notes stripped from this file:
 * <ul>
 *   <li>The private no-arg constructor {@code hb()} that always throws
 *       {@link Error} was an anti-tamper placeholder — dropped per instructions
 *       (it simply prevents instantiation of the abstract class through
 *       reflection; the abstract keyword already enforces this in Java).</li>
 *   <li>No opaque predicates, profiling counters, or XOR string pools were
 *       present in this class (the class is extremely small — 3 members).</li>
 * </ul>
 */
abstract class FilterStage extends StreamBase { // obf: hb extends ib

    /**
     * Sample-position counter used by {@link StreamMixer} to determine when
     * this stage should next be triggered.
     *
     * <p>The mixer maintains a sorted {@link LinkedQueue} (obf: db) of active
     * FilterStage instances ordered by this value.  On each mix block the
     * mixer computes {@code offset = nextTriggerSample - currentSamplePos} and
     * fires this stage when offset falls to zero.  After firing, the return
     * value of {@link #processMixBlock(StreamMixer)} is stored back into this
     * field (if non-negative) for the next scheduling cycle.
     *
     * <p>Value semantics:
     * <ul>
     *   <li>&ge; 0 — samples until this stage's next trigger (0 = trigger
     *       immediately on next mix block).</li>
     *   <li>&lt; 0 — stage is inactive / finished; the mixer removes it from
     *       the queue and calls {@link #reset()}.</li>
     * </ul>
     *
     * obf: g
     */
    int nextTriggerSample; // obf: g

    /**
     * Process one mix-block and return the next scheduling offset.
     *
     * <p>Called by {@link StreamMixer} (obf: ra) inside a {@code synchronized}
     * block on this instance each time the stage's scheduled trigger sample is
     * reached.  The implementation should mix/generate audio into whatever
     * internal buffers it manages and then return:
     * <ul>
     *   <li>A non-negative integer — the number of samples from <em>now</em>
     *       until this stage should be triggered again.  The mixer will update
     *       {@link #nextTriggerSample} to this value and re-insert the stage
     *       into its sorted queue.</li>
     *   <li>A negative integer — signals that this stage is finished.  The
     *       mixer will set {@link #nextTriggerSample} to 0 and then remove the
     *       stage from the queue (calling {@link #reset()} to clean up).</li>
     * </ul>
     *
     * <p>The {@link StreamMixer} is passed as the argument so that the
     * stage can drive child {@link FilterChain}s (obf: va) held by the mixer,
     * e.g., by calling {@code mixer.addChildChain(chain)} or by reading the
     * mixer's current sample position.
     *
     * @param mixer the {@link StreamMixer} that owns and schedules this stage
     * @return next trigger offset in samples (&ge;0 = reschedule, &lt;0 = done)
     *
     * obf: a(ra)
     */
    abstract int processMixBlock(StreamMixer mixer); // obf: int a(ra var1)

    /**
     * Reset / clean up this stage after it has been removed from the mixer.
     *
     * <p>Called by {@link StreamMixer} immediately after dequeuing a finished
     * stage (i.e., after {@link #processMixBlock(StreamMixer)} returned a
     * negative value).  Implementations should release any resources they hold
     * (e.g., unlink from a {@link FilterChain} by calling
     * {@code StreamBase.unlinkSelf(-27331)}).
     *
     * <p>The sentinel value {@code -27331} (0xFFFF935D) is the per-method
     * anti-tamper constant used by {@link StreamBase#unlinkSelf(int)}
     * (obf: {@code ib.a(int)}) to authorise the unlink operation; it is not
     * meaningful at the application level.
     *
     * obf: a()
     */
    abstract void reset(); // obf: void a()
}
