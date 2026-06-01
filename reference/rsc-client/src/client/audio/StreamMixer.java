package client.audio;

// Import stubs — com.ms.* and netscape.javascript.* won't resolve in a standard JDK, which is fine;
// the goal is faithful readable logic, not recompilation.
import client.util.LinkedQueue;   // db  — doubly-linked intrusive queue
import client.util.DecodeBuffer;  // ac  — static list-insert helper used for ordered insertion
import client.net.StreamBase;

/**
 * StreamMixer — obf: ra
 *
 * A concrete {@link FilterChain} (va) that acts as a time-scheduled audio mixer.
 * It owns two {@link LinkedQueue}s:
 *   • {@code childChains}  — the set of {@link FilterChain} (va) children that are currently
 *                            producing samples (i.e. active voices / sub-decoders such as
 *                            {@code SoundDecoder} / sb instances).
 *   • {@code stageQueue}   — a priority queue of pending {@link FilterStage} (hb) nodes sorted
 *                            by their {@code scheduledSampleOffset} (hb.g), representing future
 *                            events (note-on, envelope change, etc.) to be applied at specific
 *                            sample positions within the output stream.
 *
 * Each call to {@code mixIntoBuffer} or {@code skipSamples} advances a running sample counter
 * ({@code samplesCurrent}) and fires any {@link FilterStage} events whose scheduled position
 * has been reached.  When a stage fires, it calls {@code stage.activate(this)} to get either
 * (a) the next scheduled offset for the same stage (re-queue), or (b) a negative value meaning
 * "detach and stop".
 *
 * Child chains are mixed together by iterating {@code childChains} and delegating the
 * {@code mixIntoBuffer} / {@code skipSamples} calls to each one.  The parent (AudioChannel / sa)
 * drives this object through the {@link FilterChain} interface.
 *
 * The class is final and all public entry points are synchronized on the instance to guard the
 * shared {@code stageQueue} and {@code samplesCurrent} / {@code nextStageOffset} state from
 * concurrent access by the AudioMixer thread.
 *
 * Relationship to the engine:
 *   AudioMixer (eb) → AudioChannel (sa) → StreamMixer (ra) → [SoundDecoder (sb), ...]
 *   StreamMixer is created by AudioChannel when assembling a voice graph and populated via
 *   {@code scheduleStage(SampleBuffer, int, int)} which builds a SoundDecoder (sb) wrapped
 *   in a FilterStage (hb) and enqueues it.
 */
public final class StreamMixer extends FilterChain /* va */ {

    // -------------------------------------------------------------------------
    // Fields
    // -------------------------------------------------------------------------

    /**
     * The set of currently-active child {@link FilterChain} (va) nodes that contribute
     * samples to this mixer's output.  Iterated on every mix/skip call.
     *
     * obf: n
     */
    private LinkedQueue /* db */ childChains = new LinkedQueue();

    /**
     * Priority queue of pending {@link FilterStage} (hb) nodes ordered by their
     * {@code scheduledSampleOffset} field (hb.g, ascending).  The head of the queue
     * is the next stage event to fire.
     *
     * obf: l
     */
    private LinkedQueue /* db */ stageQueue = new LinkedQueue();

    /**
     * Running count of the number of samples that have been processed (mixed or skipped)
     * since the last time the stage queue's offsets were rebased via
     * {@code rebaseStageOffsets()}.  Used to determine when the next stage event fires.
     *
     * obf: m
     */
    private int samplesCurrent = 0;

    /**
     * Sample offset of the next pending {@link FilterStage} event, measured relative to
     * the same zero-point as {@code samplesCurrent}.  Set to {@code -1} when the queue
     * is empty (no pending stages).
     *
     * obf: k
     */
    private int nextStageOffset = -1;

    // -------------------------------------------------------------------------
    // FilterChain (va) abstract method implementations
    // -------------------------------------------------------------------------

    /**
     * Returns the next child {@link FilterChain} during a forward iteration that was
     * started by {@link #getFirstChildChain()}.  Advances the {@link LinkedQueue}
     * iteration cursor.
     *
     * Corresponds to {@code va.a()} in the obfuscated hierarchy; called as
     * {@code chain.a()} by {@link AudioChannel} (sa) after the first call to
     * {@code chain.b()} to walk remaining children.
     *
     * Calls {@code LinkedQueue.iterateNext(byte)} (db.b(byte)).
     *
     * obf: va a()
     */
    @Override
    final FilterChain /* va */ getNextChildChain() {
        // db.b(byte) — returns the node at the current cursor and advances it
        return (FilterChain) this.childChains.iterateNext((byte) 115); // obf: this.n.b((byte)115)
    }

    /**
     * Returns the first (head) child {@link FilterChain} and resets the iteration
     * cursor so subsequent calls to {@link #getNextChildChain()} walk the remainder.
     * Part of the {@link FilterChain} iterator contract used by {@link AudioChannel}
     * (sa) when walking the voice tree.
     *
     * Corresponds to {@code va.b()} in the obfuscated hierarchy; always called first
     * in sa's mix loop before {@code a()}.
     *
     * Calls {@code LinkedQueue.iterateFirst(byte)} (db.a(byte)).
     *
     * obf: va b()
     */
    @Override
    final FilterChain /* va */ getFirstChildChain() {
        // db.a(byte) — returns the head node and seeds the iteration cursor
        return (FilterChain) this.childChains.iterateFirst((byte) 35); // obf: this.n.a((byte)35)
    }

    /**
     * Returns {@code 0} because StreamMixer does not impose a finite sample length —
     * it runs for as long as stages are enqueued.
     *
     * obf: int d()
     */
    @Override
    public final int getSampleLength() {
        return 0;
    }

    // -------------------------------------------------------------------------
    // Core mix / skip path (synchronized — called from AudioMixer thread)
    // -------------------------------------------------------------------------

    /**
     * Mix {@code sampleCount} samples into {@code outputBuf[offset .. offset+sampleCount-1]},
     * firing any pending {@link FilterStage} events that fall within the range being mixed.
     *
     * The method processes the output buffer in segments: for each segment it mixes from
     * all active child chains, then checks whether the next stage event falls within the
     * segment; if so, it truncates the segment at that boundary, fires the stage, and
     * continues with the remainder.
     *
     * Synchronized to protect {@code stageQueue}, {@code samplesCurrent}, and
     * {@code nextStageOffset} from concurrent modification.
     *
     * @param outputBuf   accumulation buffer (samples are summed in — not overwritten)
     * @param offset      start index in {@code outputBuf}
     * @param sampleCount number of samples to mix
     *
     * obf: synchronized void b(int[], int, int)
     */
    @Override
    public final synchronized void mixIntoBuffer(int[] outputBuf, int offset, int sampleCount) {
        // While there is at least one pending stage event within reach:
        while (this.nextStageOffset >= 0) {
            if (this.samplesCurrent + sampleCount < this.nextStageOffset) {
                // The entire sampleCount fits before the next stage event — mix it all and return.
                this.samplesCurrent += sampleCount;
                this.mixChildrenIntoBuffer(outputBuf, offset, sampleCount);
                return;
            }

            // Mix up to (but not including) the stage boundary.
            int samplesUntilStage = this.nextStageOffset - this.samplesCurrent;
            this.mixChildrenIntoBuffer(outputBuf, offset, samplesUntilStage);
            offset += samplesUntilStage;
            sampleCount -= samplesUntilStage;
            this.samplesCurrent += samplesUntilStage;

            // Rebase all stage offsets so that the current position is zero again.
            this.rebaseStageOffsets();

            // Fire the head stage: dequeue it and either re-schedule or detach.
            FilterStage /* hb */ stage = (FilterStage) this.stageQueue.peekHead((byte) 106); // obf: this.l.a((byte)106)
            synchronized (stage) {
                int nextOffset = stage.activate(this); // obf: hb.a(ra) — returns next sample offset, or <0 to stop
                if (nextOffset < 0) {
                    // Stage is done: offset 0, then detach from this mixer.
                    stage.scheduledSampleOffset = 0;              // obf: hb.g = 0
                    this.detachStage(stage);
                } else {
                    // Re-schedule: insert with new offset, maintaining sorted order.
                    stage.scheduledSampleOffset = nextOffset;     // obf: hb.g = nextOffset
                    this.insertStageSorted(stage.nextStage /* ib.a */, stage);
                }
            }

            if (sampleCount == 0) {
                return;
            }
        }

        // No more pending stage events — just mix remaining samples straight through.
        this.mixChildrenIntoBuffer(outputBuf, offset, sampleCount);
    }

    /**
     * Advance the mixer by {@code sampleCount} samples without producing output
     * (i.e. the upstream decided the voice is silent / muted for this window).
     * Fires stage events just as {@link #mixIntoBuffer} would.
     *
     * Synchronized for the same reason as {@link #mixIntoBuffer}.
     *
     * @param sampleCount number of samples to skip
     *
     * obf: synchronized void b(int)
     */
    @Override
    public final synchronized void skipSamples(int sampleCount) {
        while (this.nextStageOffset >= 0) {
            if (this.samplesCurrent + sampleCount < this.nextStageOffset) {
                this.samplesCurrent += sampleCount;
                this.skipChildrenSamples(sampleCount);
                return;
            }

            int samplesUntilStage = this.nextStageOffset - this.samplesCurrent;
            this.skipChildrenSamples(samplesUntilStage);
            sampleCount -= samplesUntilStage;
            this.samplesCurrent += samplesUntilStage;

            this.rebaseStageOffsets();

            FilterStage /* hb */ stage = (FilterStage) this.stageQueue.peekHead((byte) -80); // obf: this.l.a((byte)-80)
            synchronized (stage) {
                int nextOffset = stage.activate(this); // obf: hb.a(ra)
                if (nextOffset < 0) {
                    stage.scheduledSampleOffset = 0;             // obf: hb.g = 0
                    this.detachStage(stage);
                } else {
                    stage.scheduledSampleOffset = nextOffset;    // obf: hb.g = nextOffset
                    this.insertStageSorted(stage.nextStage /* ib.a */, stage);
                }
            }

            if (sampleCount == 0) {
                return;
            }
        }

        this.skipChildrenSamples(sampleCount);
    }

    // -------------------------------------------------------------------------
    // Public schedule entry-point
    // -------------------------------------------------------------------------

    /**
     * Build a {@link SoundDecoder} (sb) from the given raw sample buffer and schedule
     * it as the initial stage of this mixer.  Called by {@link AudioChannel} (sa) or
     * the client when setting up a new audio voice.
     *
     * Delegates to {@code SoundDecoder.create(SampleBuffer, int, int)} (sb.a) to construct
     * the decoder, then enqueues the resulting stage.
     *
     * @param sampleBuf   the raw PCM/encoded sample data  (obf: vb)
     * @param pitch       playback pitch / rate parameter
     * @param envelope    initial envelope value
     *
     * obf: void a(vb, int, int)
     */
    public final void scheduleStage(SampleBuffer /* vb */ sampleBuf, int pitch, int envelope) {
        // sb.a(vb, int, int) — factory: creates SoundDecoder or returns null if buffer empty.
        // The resulting SoundDecoder (sb extends va / FilterChain) is added as an active child.
        this.addChildChain(SoundDecoder.create(sampleBuf, pitch, envelope)); // obf: this.a(sb.a(...))
    }

    // -------------------------------------------------------------------------
    // Private helpers
    // -------------------------------------------------------------------------

    /**
     * Rebase all stage offsets in {@code stageQueue} by subtracting {@code samplesCurrent},
     * then reset {@code samplesCurrent} to zero.  This prevents the counters from growing
     * without bound and keeps the arithmetic tractable.
     *
     * Also subtracts {@code samplesCurrent} from {@code nextStageOffset}.
     *
     * Only called when {@code samplesCurrent > 0}.
     *
     * obf: private void e()
     */
    private void rebaseStageOffsets() {
        if (this.samplesCurrent > 0) {
            // Walk all stages in the queue and subtract the elapsed sample count.
            for (FilterStage /* hb */ stage = (FilterStage) this.stageQueue.iterateFirst((byte) -87);    // obf: this.l.a((byte)-87)
                 stage != null;
                 stage = (FilterStage) this.stageQueue.iterateContinue((byte) 111)) {                    // obf: this.l.b((byte)111)
                stage.scheduledSampleOffset -= this.samplesCurrent; // obf: var1.g -= this.m
            }
            this.nextStageOffset -= this.samplesCurrent;
            this.samplesCurrent = 0;
        }
    }

    /**
     * Insert {@code stage} into the {@code stageQueue} at the correct sorted position
     * (sorted ascending by {@code scheduledSampleOffset}), starting the search from
     * {@code searchFrom}.  After insertion, refresh {@code nextStageOffset} from the
     * new queue head.
     *
     * Uses {@code DecodeBuffer.insertBefore(ib, byte, ib)} (ac.a) to do the intrusive
     * list insertion.
     *
     * @param searchFrom linked-list node to begin the search from (typically
     *                   {@code stage.nextStage} = the node that was just after it, or
     *                   the first node of the queue)
     * @param stage      the stage to insert
     *
     * obf: private void a(ib, hb)
     */
    private void insertStageSorted(StreamBase /* ib */ searchFrom, FilterStage /* hb */ stage) {
        // Walk forward from searchFrom until we find a node whose offset is greater than
        // stage's, or we reach the sentinel (stageQueue.sentinel / db.k).
        while (searchFrom != this.stageQueue.sentinel /* db.k */
               && ((FilterStage) searchFrom).scheduledSampleOffset <= stage.scheduledSampleOffset) {
            searchFrom = searchFrom.nextNode; // obf: var1 = var1.a
        }

        // Insert stage before searchFrom (so the list stays sorted ascending).
        // ac.a(ib node, byte magic, ib insertBefore) — intrusive doubly-linked insert
        DecodeBuffer.insertBefore(stage, (byte) 34, searchFrom); // obf: ac.a(var2, (byte)34, var1)

        // Update nextStageOffset: the new earliest event is the node just after the sentinel.
        // stageQueue.sentinel.nextNode (db.k.a) is the head (smallest offset) in this queue layout.
        this.nextStageOffset = ((FilterStage) this.stageQueue.sentinel.nextNode).scheduledSampleOffset;
        // obf: this.k = ((hb)this.l.k.a).g
    }

    /**
     * Detach a completed stage from the queue: unlink it from the intrusive list, call its
     * cleanup method, then update {@code nextStageOffset}.
     *
     * If the queue becomes empty after removal, {@code nextStageOffset} is set to {@code -1}.
     *
     * @param stage the stage to remove and finalize
     *
     * obf: private void a(hb)
     */
    private void detachStage(FilterStage /* hb */ stage) {
        // Unlink stage from the intrusive doubly-linked list.
        // ib.a(-27331) is the "unlink self" sentinel value in StreamBase.
        stage.unlinkFromList(-27331); // obf: var1.a(-27331)

        // Let the stage perform any internal cleanup (e.g. stop child decoders).
        stage.cleanup(); // obf: var1.a()  — abstract void a() in hb

        // Peek at the new head: stageQueue.sentinel.nextNode
        StreamBase /* ib */ head = this.stageQueue.sentinel.nextNode; // obf: this.l.k.a
        if (head == this.stageQueue.sentinel /* db.k */) {
            // Queue is now empty — no more pending stage events.
            this.nextStageOffset = -1;
        } else {
            this.nextStageOffset = ((FilterStage) head).scheduledSampleOffset; // obf: ((hb)var2).g
        }
    }

    /**
     * Add a {@link FilterChain} (va) child to the active {@code childChains} mix set.
     * Called by {@link #scheduleStage} with a freshly-constructed {@link SoundDecoder}
     * (sb), and also available to {@link FilterStage} subclasses (hb) via the {@code ra}
     * reference they receive in {@code activate(StreamMixer)} — stages call back here to
     * install new decoder children as the voice progresses.
     *
     * Synchronized because the AudioMixer thread may be iterating {@code childChains} at
     * the same time.
     *
     * Uses {@code LinkedQueue.add(ib, false)} (db.a(ib, false)) to append at the tail
     * without triggering a wait/notify.
     *
     * obf: private synchronized void a(va)
     */
    private synchronized void addChildChain(FilterChain /* va */ child) {
        this.childChains.add(child, false); // obf: this.n.a(var1, false)
    }

    /**
     * Delegate a mix-into-buffer call to every child chain currently in
     * {@code childChains}.  Each child's {@code mixIntoBuffer} (va.a(int[],int,int))
     * accumulates its samples into the shared {@code outputBuf}.
     *
     * @param outputBuf   shared accumulation buffer
     * @param offset      start index in outputBuf
     * @param sampleCount samples to mix
     *
     * obf: private void c(int[], int, int)
     */
    private void mixChildrenIntoBuffer(int[] outputBuf, int offset, int sampleCount) {
        // db.a(byte) starts iteration; db.b(byte) advances it.
        for (FilterChain /* va */ child = (FilterChain) this.childChains.iterateFirst((byte) -88);  // obf: this.n.a((byte)-88)
             child != null;
             child = (FilterChain) this.childChains.iterateContinue((byte) 122)) {                  // obf: this.n.b((byte)122)
            // va.a(int[], int, int) — checks the 'active' flag (va.g) then calls b() or skip
            child.mixIntoBuffer(outputBuf, offset, sampleCount); // obf: var4.a(var1, var2, var3)
        }
    }

    /**
     * Delegate a skip-samples call to every child chain in {@code childChains}.
     * Used by {@link #skipSamples} to keep child decoder positions in sync.
     *
     * @param sampleCount number of samples to skip
     *
     * obf: private void c(int)
     */
    private void skipChildrenSamples(int sampleCount) {
        // db.a(byte) starts; db.b(byte) continues iteration.
        for (FilterChain /* va */ child = (FilterChain) this.childChains.iterateFirst((byte) -123);  // obf: this.n.a((byte)-123)
             child != null;
             child = (FilterChain) this.childChains.iterateContinue((byte) 80)) {                    // obf: this.n.b((byte)80)
            child.skipSamples(sampleCount); // obf: var2.b(var1)  — va.b(int)
        }
    }

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /**
     * Default constructor.  Both queues are empty; {@code nextStageOffset} starts at
     * {@code -1} (no pending stages).
     *
     * obf: ra()
     */
    public StreamMixer() {
        // Fields initialised at declaration site above; nothing else to do.
    }
}
