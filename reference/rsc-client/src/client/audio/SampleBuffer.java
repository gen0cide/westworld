package client.audio;

/**
 * Raw PCM audio sample data carrier — a concrete {@link FilterNode} that holds
 * the decoded sound bytes consumed by the {@link SoundDecoder} filter chain.
 *
 * <p>This is the leaf node of the RSC audio filter graph.  A {@code SampleBuffer}
 * is created once per sound-effect play request (inside {@code Mudclient}) and
 * handed to {@link SoundDecoder#create(SampleBuffer, int, int)} which wraps it
 * in a {@code SoundDecoder} ({@code sb}) filter chain and schedules it through
 * the {@link AudioChannel} ({@code sa}) mixer.
 *
 * <p>The one confirmed construction site (from {@code Mudclient}, obf field
 * {@code hk} = the {@link AudioChannel}):
 * <pre>
 *   // obf: new vb(8000, v.a(this.Uh, var4, -98, var8), 0, var4);
 *   SampleBuffer buf = new SampleBuffer(8000, sampleBytes, 0, sampleLength);
 *   audioChannel.play(buf, 100 /*volume*&#47;, 256 /*pan*&#47;);
 * </pre>
 *
 * <p>The {@link SoundDecoder} reads {@link #sampleData} directly (e.g.
 * {@code ((vb)this.h).l[this.x]}) to mix samples into the output
 * {@code int[]} buffer.  {@link #loopStart} and {@link #loopEnd} bound the
 * loop region; {@link #pingPong} selects bounce vs. wrap looping;
 * {@link #baseSampleRate} is scaled against {@code AudioChannel.sampleRate}
 * ({@code sa.t}) to compute the playback pitch increment.
 *
 * <p>Inheritance:
 * <pre>
 *   StreamBase  (ib)
 *     └─ FilterNode  (bb)
 *          └─ SampleBuffer  (vb)   ← this class
 * </pre>
 *
 * Obfuscated class name: {@code vb} (default package, rev ~233–235 Microsoft J++ jar).
 */
// obf: vb
public final class SampleBuffer extends FilterNode /* bb */ {

    /**
     * Raw signed 8-bit PCM sample bytes.
     *
     * <p>The {@link SoundDecoder} mixer reads from this array using a fixed-point
     * position cursor {@code v} (stored as {@code position << 8} so the low byte
     * carries the sub-sample interpolation fraction).  Bytes are signed and are
     * multiplied by a gain/volume factor before being accumulated into the
     * {@code int[]} output mix buffer.
     *
     * obf: {@code l}
     */
    public byte[] sampleData; // obf: l

    /**
     * Base (nominal) sample rate of this sound in Hz (e.g. {@code 8000}).
     *
     * <p>Used by {@link SoundDecoder#create(SampleBuffer, int, int)} to compute
     * the fixed-point pitch step:
     * <pre>
     *   pitchStep = (long) baseSampleRate * 256L * requestedPitch
     *               / (long)(100 * AudioChannel.sampleRate);
     * </pre>
     * A value of 8 000 with {@code AudioChannel.sampleRate} = 8 000 and pitch
     * = 100 yields a step of 256, i.e. one sample per output frame (no
     * resampling).
     *
     * obf: {@code i}
     */
    public int baseSampleRate; // obf: i

    /**
     * Ping-pong loop flag.
     *
     * <p>When {@code true} the {@link SoundDecoder} reflects the playback
     * direction each time it reaches a loop boundary (ping-pong / bounce loop).
     * When {@code false} the decoder wraps around (forward-loop or
     * backward-loop depending on the direction of the pitch step).
     *
     * obf: {@code j}
     */
    public boolean pingPong; // obf: j

    /**
     * Loop end sample index (byte offset into {@link #sampleData}).
     *
     * <p>The {@link SoundDecoder} copies this into its {@code t} field and
     * uses it as the upper bound of the loop region.  Specifically,
     * {@code SoundDecoder.t << 8} is compared against the fixed-point
     * position cursor.
     *
     * <p>In the only observed construction site this is set equal to the length
     * of {@link #sampleData} (i.e. {@code var4} = {@code sampleData.length}),
     * meaning the loop region covers the entire sample.
     *
     * obf: {@code k}
     */
    public int loopEnd; // obf: k

    /**
     * Loop start sample index (byte offset into {@link #sampleData}).
     *
     * <p>The {@link SoundDecoder} copies this into its {@code x} field and
     * uses it as the lower bound of the loop region.  Specifically,
     * {@code SoundDecoder.x << 8} is compared against the fixed-point
     * position cursor.
     *
     * <p>A value of {@code 0} means the loop begins at the very first sample.
     *
     * obf: {@code h}
     */
    public int loopStart; // obf: h

    /**
     * Constructs a new sample buffer from raw decoded PCM data.
     *
     * <p>No obfuscation artifacts were present in this constructor (no opaque
     * predicate, no profiling counter, no exception wrapper — the class is
     * trivially small).
     *
     * @param baseSampleRate  nominal sample rate in Hz (e.g. {@code 8000})
     * @param sampleData      signed 8-bit PCM bytes; may be {@code null} or
     *                        empty (the {@link SoundDecoder} factory returns
     *                        {@code null} in that case and the sound is skipped)
     * @param loopStart       first sample index of the loop region (0 = start)
     * @param loopEnd         one-past-last sample index of the loop region
     *                        (typically {@code sampleData.length})
     *
     * obf constructor: {@code vb(int, byte[], int, int)}
     */
    public SampleBuffer(int baseSampleRate, byte[] sampleData, int loopStart, int loopEnd) {
        // obf: this.i = var1; this.l = var2; this.h = var3; this.k = var4;
        this.baseSampleRate = baseSampleRate;
        this.sampleData     = sampleData;
        this.loopStart      = loopStart;
        this.loopEnd        = loopEnd;
        // pingPong defaults to false (Java field default)
    }
}
