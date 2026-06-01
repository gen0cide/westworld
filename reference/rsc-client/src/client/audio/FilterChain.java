package client.audio;

import client.net.StreamBase;   // ib — base resource/stream node class

/**
 * FilterChain (obf: va) — Abstract audio filter-chain node.
 *
 * <p>Sits at the heart of the RSC audio pipeline. Every playing sound is
 * represented as a {@code FilterChain} node that the mixer ({@link StreamMixer})
 * walks each mix-tick. The class is a doubly-intrusive singly-linked list:
 *
 * <ul>
 *   <li>{@link #next} links a {@code FilterChain} to the one scheduled behind it
 *       in the same priority-slot (forms the per-slot list maintained by
 *       {@link AudioChannel} / {@link StreamMixer}).</li>
 *   <li>{@link #filterNode} is an optional {@link FilterNode} (obf: bb) attached
 *       to this chain, carrying a mutable sample-count/budget ({@code FilterNode.g})
 *       that the mixer uses for priority/budget accounting.</li>
 *   <li>{@link #priority} is the absolute sample-position at which this node
 *       is due to fire (set by {@link AudioChannel} when the node is enqueued).</li>
 *   <li>{@link #active} gates which mix-path is taken: when {@code true} the full
 *       PCM-mix path ({@link #mixInto(int[], int, int)}) is called; when
 *       {@code false} the skip path ({@link #skipSamples(int)}) is called
 *       instead, advancing the read head without writing samples.</li>
 * </ul>
 *
 * <p>Two concrete subclasses exist:
 * <ul>
 *   <li>{@link StreamMixer} (obf: ra) — schedules {@link FilterStage} (obf: hb)
 *       events over a {@link LinkedQueue} (obf: db) and mixes child
 *       {@code FilterChain} nodes.</li>
 *   <li>{@link SoundDecoder} (obf: sb) — large bit-decode / Huffman-ish sound
 *       decoder that reads raw sample bytes from a {@link SampleBuffer}
 *       (obf: vb) and writes PCM into the mix buffer.</li>
 * </ul>
 *
 * <p>Inheritance: {@code StreamBase (ib) → FilterChain (va)}
 * <br>Subclasses:  {@code FilterChain → StreamMixer (ra), SoundDecoder (sb)}
 *
 * <p>Usage path (per mix-tick in {@link AudioChannel} / obf: sa):
 * <ol>
 *   <li>Mixer calls {@link #mix(int[], int, int)} on each active chain.</li>
 *   <li>{@code mix} dispatches to either {@link #mixInto} (PCM path) or
 *       {@link #skipSamples} (drain/skip path) depending on {@link #active}.</li>
 *   <li>After mixing, the mixer calls {@link #firstChild()} / {@link #nextChild()}
 *       to walk any children that need to be enqueued.</li>
 * </ol>
 */
public abstract class FilterChain extends StreamBase /* ib */ {

    // -----------------------------------------------------------------------
    // Fields
    // -----------------------------------------------------------------------

    /**
     * Next node in the same priority-bucket linked-list.
     * Maintained by {@link AudioChannel} (sa) and {@link StreamMixer} (ra).
     * {@code null} when this node is the last in its bucket or not enqueued.
     *
     * obf: j  (field type: va → FilterChain)
     */
    public FilterChain next;

    /**
     * Optional FilterNode (obf: bb) associated with this chain.
     * Carries a mutable sample-count budget {@code FilterNode.g} that the
     * mixer reads/writes for priority accounting and sound-end detection.
     * May be {@code null} for nodes without a budget (e.g. the root
     * {@link StreamMixer}).
     *
     * obf: h  (field type: bb → FilterNode)
     */
    FilterNode /* bb */ filterNode;

    /**
     * Priority / scheduled-sample-position of this node.
     * Set by {@link AudioChannel#enqueue(FilterChain, int)} to the absolute
     * sample counter at which this node should next fire.
     * Used by the scheduler in {@link StreamMixer} to order events.
     *
     * obf: i
     */
    public int priority;

    /**
     * Active flag: {@code true} → mix PCM samples into the output buffer;
     * {@code false} → just advance the read head (skip mode).
     *
     * <p>The mixer drives this flag via the static helper
     * {@code AudioChannel.resetChain(FilterChain)} (obf: {@code sa.b(va)}):
     * it sets {@code active = false} and zeros {@code filterNode.g} on the
     * chain and all of its children before re-scheduling, then the chain
     * goes back to {@code active = true} when the {@link AudioChannel}
     * dispatches the next buffer.
     *
     * Declared {@code volatile} because {@link AudioMixer} (eb) reads it from
     * its own thread while {@link AudioChannel} may write it from a caller
     * thread.
     *
     * obf: g
     */
    public volatile boolean active = true;

    // -----------------------------------------------------------------------
    // Abstract API — subclasses must implement all five
    // -----------------------------------------------------------------------

    /**
     * Skip-mode path: advance the read position by {@code sampleCount}
     * samples without writing any PCM output.
     * Called by {@link #mix} when {@link #active} is {@code false}.
     *
     * obf: b(int)
     *
     * @param sampleCount number of samples to skip
     */
    public abstract void skipSamples(int sampleCount);

    /**
     * Return the number of samples this node has decoded/advanced since it
     * was last reset. Used by {@link AudioChannel} to compute budget
     * accounting for the {@link FilterNode}.
     *
     * obf: d()
     *
     * @return sample count delta (≥ 0)
     */
    public abstract int getSampleDelta();

    /**
     * Return the first child {@link FilterChain} that needs to be enqueued
     * in the parent's scheduler, resetting the internal child-iterator.
     * Returns {@code null} if there are no children.
     *
     * <p>In {@link StreamMixer} this pulls from the "pending children" queue
     * ({@code n}, a {@link LinkedQueue}).
     *
     * obf: a() → FilterChain
     */
    public abstract FilterChain firstChild();

    /**
     * Mix (or decode) {@code length} samples starting at output-buffer
     * offset {@code offset}, accumulating into {@code buffer}.
     * Called by {@link #mix} when {@link #active} is {@code true}.
     *
     * <p>In {@link SoundDecoder} this is the full PCM-decode path
     * (reads from {@link SampleBuffer}, applies pitch/volume, accumulates
     * 32-bit integers into {@code buffer}).
     * In {@link StreamMixer} this fans out to all child chains.
     *
     * obf: b(int[], int, int)
     *
     * @param buffer accumulation buffer (32-bit PCM, one entry per sample)
     * @param offset index of first sample to write
     * @param length number of samples to mix
     */
    public abstract void mixInto(int[] buffer, int offset, int length);

    /**
     * Return the next child {@link FilterChain} from the internal
     * child-iterator (following a call to {@link #firstChild()}).
     * Returns {@code null} when the iterator is exhausted.
     *
     * obf: b() → FilterChain
     */
    public abstract FilterChain nextChild();

    // -----------------------------------------------------------------------
    // Concrete methods
    // -----------------------------------------------------------------------

    /**
     * Return the volume/weight of this chain in the range [0, 255].
     * Used by {@link AudioChannel} when computing the mix weight for child
     * chains: {@code weight = parentPriority * child.getVolume() >> 8}.
     *
     * <p>The base implementation returns the maximum value (255 = full
     * volume / no attenuation). {@link SoundDecoder} overrides this to
     * return a value derived from its pitch-step and loop state.
     *
     * obf: c()
     *
     * @return volume in [0, 255]
     */
    public int getVolume() {
        return 255;
    }

    /**
     * Dispatch a mix tick to this node.
     *
     * <p>If {@link #active} is {@code true}, the full PCM mix path is taken
     * ({@link #mixInto(int[], int, int)}). If {@link #active} is
     * {@code false}, only the skip path is taken ({@link #skipSamples(int)}),
     * discarding the output offset and length and using only the total
     * {@code length} to advance the read head.
     *
     * <p>This is the entry point called by {@link AudioChannel}'s inner mixing
     * loop (obf: {@code sa.a(int[],int)}).
     *
     * obf: a(int[], int, int)  [final]
     *
     * @param buffer accumulation buffer
     * @param offset first sample index to mix into
     * @param length number of samples in this tick
     */
    public final void mix(int[] buffer, int offset, int length) {
        if (this.active) {
            // Active path: decode/mix PCM samples into the accumulation buffer.
            this.mixInto(buffer, offset, length);
        } else {
            // Skip/drain path: advance read position without writing output.
            // Note: only `length` is passed — the buffer and offset are
            // irrelevant in skip mode.
            this.skipSamples(length);
        }
    }
}
