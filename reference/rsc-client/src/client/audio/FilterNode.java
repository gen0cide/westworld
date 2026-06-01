package client.audio;

/**
 * Abstract audio filter node — base class for all sample-buffer nodes in the
 * RSC audio filter graph.
 *
 * <p>Sits at the junction between the filter-chain scheduler ({@link FilterChain}
 * / {@code va}) and concrete sample storage ({@link SampleBuffer} / {@code vb}).
 * A {@link FilterChain} carries a reference to a {@code FilterNode} via its
 * {@code h} field; the AudioChannel ({@code sa}) scheduler reads
 * {@code filterNode.samplesConsumed} to decide whether a chain has caught up
 * to the playback cursor, and increments it as the chain produces output.
 *
 * <p>Inheritance:
 * <pre>
 *   StreamBase  (ib)
 *     └─ FilterNode  (bb)   ← this class
 *          └─ SampleBuffer (vb)   — carries the raw byte[] sample data
 * </pre>
 *
 * The single field {@code samplesConsumed} is a running count (in samples)
 * of how many audio samples this node has contributed to the output stream
 * since the last scheduler reset.  The AudioChannel scheduler ({@code sa})
 * zeros it at the start of each mix quantum and adds to it as each linked
 * FilterChain is drained:
 * <pre>
 *   if (filterNode != null && filterNode.samplesConsumed > writeHead) skip;
 *   ...
 *   if (filterNode != null) filterNode.samplesConsumed += samplesProduced;
 * </pre>
 *
 * Obfuscated name: {@code bb} (default package, rev ~233–235 Microsoft J++ jar).
 */
// obf: bb
abstract class FilterNode extends StreamBase /* ib */ {

    /**
     * Running total of PCM samples that have been consumed / produced through
     * this node since the last scheduler reset.
     *
     * <p>The AudioChannel ({@code sa}) scheduler uses this to avoid scheduling
     * a chain twice within the same mix quantum: if
     * {@code samplesConsumed > writeHead} the chain is skipped.  At reset
     * (start of each mix block) the scheduler calls
     * {@code filterNode.samplesConsumed = 0} on every active node.
     *
     * obf: {@code g}
     */
    int samplesConsumed; // obf: g
}
