package client.audio;

/**
 * SoundDecoder — a concrete {@link FilterChain} (va) node that drives playback of a
 * single raw PCM sample buffer ({@link SampleBuffer}/vb) through the audio mixing graph.
 *
 * <p>Responsibilities:
 * <ul>
 *   <li>Tracks a fractional read-cursor ({@link #samplePos}) into the backing
 *       {@code byte[]} sample data at 8-bit fixed-point sub-sample resolution
 *       (low 8 bits = sub-sample fraction, high bits = integer sample index).</li>
 *   <li>Supports four loop modes keyed on {@link #loopCount}: infinite forward/back
 *       ({@code loopCount < 0}), finite-bounce ({@code loopCount > 0, pingPong}),
 *       finite-wrap, and one-shot ({@code loopCount == 0}).</li>
 *   <li>Interpolates between adjacent samples when {@code pitchStep} is not a whole
 *       multiple of 256 (i.e. resampling is needed for pitch).</li>
 *   <li>Supports both mono and stereo mixing paths, selected by
 *       {@link AudioChannel#isStereo} ({@code sa.i}): stereo writes interleaved
 *       left/right pairs; mono writes a single channel.</li>
 *   <li>Provides a smooth volume-glide system: when the target volume ({@link #targetVol})
 *       differs from the current amplitude ({@link #curAmp}), it computes per-sample
 *       amplitude and pan deltas and blends over {@link #glideRemaining} output
 *       frames.</li>
 * </ul>
 *
 * <p>Coordinate encoding: {@code samplePos} is a fixed-point Q8 value —
 * {@code samplePos >> 8} is the integer byte index into {@code SampleBuffer.data},
 * and {@code samplePos & 0xFF} is the linear-interpolation fraction
 * (0 = exact sample, 255 = almost next sample).
 *
 * <p>All mixing helpers follow the pattern:
 * <pre>
 *   sample = data[idx];                         // raw byte (signed)
 *   interpolated = (sample &lt;&lt; 8)
 *               + (data[idx+1] - sample) * frac; // 16-bit lerp result
 *   out[outIdx] += interpolated * gain &gt;&gt; 6;    // scale &amp; accumulate
 * </pre>
 * The {@code &gt;&gt; 6} keeps the accumulated mix from overflowing a 32-bit int when
 * up to 255 voices are summed.
 *
 * <p>Factory entry point: {@link #create(SampleBuffer, int, int)}.
 *
 * obf: sb
 */
public final class SoundDecoder extends FilterChain {

    // -------------------------------------------------------------------------
    // Instance fields
    // -------------------------------------------------------------------------

    /**
     * Loop end boundary (sample index, NOT Q8-encoded). Corresponds to
     * {@code SampleBuffer.loopEnd} (vb.k). The active playback window is
     * [loopStart, loopEnd).
     * obf: t
     */
    private int loopEnd;

    /**
     * Loop start boundary (sample index, NOT Q8-encoded). Corresponds to
     * {@code SampleBuffer.loopStart} (vb.h).
     * obf: x
     */
    private int loopStart;

    /**
     * Per-sample delta applied to {@link #curPanRight} each output frame during a
     * volume glide. +1, 0, or -1 depending on direction.
     * obf: o
     */
    private int panRightDelta;

    /**
     * Per-sample delta applied to {@link #curPanLeft} each output frame during a
     * volume glide. +1, 0, or -1 depending on direction.
     * obf: p
     */
    private int panLeftDelta;

    /**
     * Pitch step in Q8 fixed-point. Each output sample advances the read cursor
     * by this many 1/256ths of a source sample.  Computed from
     * {@code SampleBuffer.sampleRate * 256 * pitchPercent / (100 * AudioChannel.sampleRate)}.
     * Negative means play in reverse.
     * obf: n
     */
    private int pitchStep;

    /**
     * Target volume (raw, before pan split). Set by the constructor via
     * {@code volShift << 6}. {@code Integer.MIN_VALUE} is a sentinel that signals
     * "stop after the current glide finishes".
     * obf: r
     */
    private int targetVol;

    /**
     * Number of output frames remaining in the current volume glide. Zero means
     * the glide is complete and no blending is needed.
     * obf: y
     */
    private int glideRemaining;

    /**
     * Per-sample delta applied to {@link #curAmp} during a glide.
     * +1, 0, or -1 depending on glide direction.
     * obf: s
     */
    private int ampDelta;

    /**
     * Current read cursor into the sample data, Q8 fixed-point.
     * {@code samplePos >> 8} = integer byte index; {@code samplePos & 0xFF} = lerp fraction.
     * obf: v
     */
    private int samplePos;

    /**
     * Loop control:
     * <ul>
     *   <li>{@code < 0} — infinite loop</li>
     *   <li>{@code = 0} — one-shot (play once, then stop)</li>
     *   <li>{@code > 0} — finite loop (counts remaining bounces/wraps)</li>
     * </ul>
     * obf: l
     */
    private int loopCount;

    /**
     * Current right-channel gain (stereo only). Q-6 fixed-point:
     * {@code curPanRight >> 6} is the linear multiplier applied to each sample.
     * Glides from its current value toward the target computed by
     * {@link #computePanRight(int, int)}.
     * obf: m
     */
    private int curPanRight;

    /**
     * Current amplitude (mono) or left-channel gain (stereo). Q-6 fixed-point.
     * Glides toward the target computed by {@link #computePanLeft(int, int)}.
     * obf: w
     */
    private int curAmp;

    /**
     * Target left-channel gain, computed once at the start of each glide from
     * {@link #targetVol} and {@link #panPosition}.
     * obf: k
     */
    private int curPanLeft;

    /**
     * Pan position [0..16384]. Used by {@link #computePanLeft} / {@link #computePanRight}
     * to split the volume into left and right gains using a square-root pan law:
     * <pre>
     *   leftGain  = vol * sqrt((16384 - pan) / 16384)
     *   rightGain = vol * sqrt(pan / 16384)
     * </pre>
     * 8192 is center (equal power). Initialized to 8192 in the constructor.
     * obf: q
     */
    private int panPosition;

    /**
     * True if this sound should ping-pong (reverse direction) at loop boundaries
     * rather than wrap. Copied from {@code SampleBuffer.pingPong} (vb.j).
     * obf: u
     */
    private boolean pingPong;

    // -------------------------------------------------------------------------
    // Static factory
    // -------------------------------------------------------------------------

    /**
     * Creates a SoundDecoder for the given sample buffer, or returns {@code null}
     * if the buffer has no data.
     *
     * @param buf          the sample buffer to play
     * @param pitchPercent pitch as a percentage of nominal (100 = original pitch)
     * @param volShift     volume level; stored shifted left by 6 bits as the target
     *                     amplitude
     * @return a ready SoundDecoder, or {@code null} if {@code buf.data} is empty
     * obf: sb.a(vb, int, int)
     */
    public static final SoundDecoder create(SampleBuffer buf, int pitchPercent, int volShift) {
        if (buf.data == null || buf.data.length == 0) {
            return null;
        }
        // pitchStep = sampleRate * 256 * pitchPercent / (100 * AudioChannel.sampleRate)
        // The division normalises to the output sample rate so the cursor advances
        // exactly one source sample per output frame at 100% pitch.
        int pitchStep = (int)((long)buf.sampleRate * 256L * (long)pitchPercent
                               / (long)(100 * AudioChannel.sampleRate));
        return new SoundDecoder(buf, pitchStep, volShift << 6);
        // obf: sa.t  →  AudioChannel.sampleRate
    }

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /**
     * Private constructor — use {@link #create} instead.
     *
     * @param buf       backing sample buffer
     * @param step      pitch step (Q8)
     * @param targetVol initial (and target) amplitude
     * obf: sb(vb, int, int)
     */
    private SoundDecoder(SampleBuffer buf, int step, int targetVol) {
        this.h          = buf;                   // FilterChain.source  (obf: h)
        this.loopStart  = buf.loopStart;         // obf: vb.h
        this.loopEnd    = buf.loopEnd;           // obf: vb.k
        this.pingPong   = buf.pingPong;          // obf: vb.j
        this.pitchStep  = step;
        this.targetVol  = targetVol;
        this.panPosition = 8192;                 // center pan
        this.samplePos  = 0;
        snapToTarget();                          // initialise curAmp / curPanLeft / curPanRight
    }

    // -------------------------------------------------------------------------
    // FilterChain abstract method implementations
    // -------------------------------------------------------------------------

    /**
     * Returns the first child filter chain node (always {@code null} — this is a
     * leaf node backed directly by a SampleBuffer).
     * obf: va.a()
     */
    @Override
    public final FilterChain prevChain() {
        return null;
    }

    /**
     * Returns the next sibling filter chain node (always {@code null}).
     * obf: va.b()
     */
    @Override
    public final FilterChain nextChain() {
        return null;
    }

    /**
     * Returns 1 if this decoder is still active (has volume or a pending glide),
     * 0 if silent and idle.
     * obf: va.d()
     */
    @Override
    public final int isActive() {
        return (this.targetVol == 0 && this.glideRemaining == 0) ? 0 : 1;
    }

    /**
     * Computes and returns the current output amplitude clamped to [0, 255].
     * Accounts for the fraction of the sample buffer already consumed when in
     * one-shot or infinite-loop mode.
     *
     * <p>Bit trick: {@code (n ^ (n >> 31)) + (n >>> 31)} is {@code Math.abs(n)}
     * without branching.
     * obf: va.c()
     */
    @Override
    public final int getAmplitude() {
        // curAmp uses a factor of 3/64 to scale to the [0,255] range
        int amp = this.curAmp * 3 >> 6;
        amp = (amp ^ (amp >> 31)) + (amp >>> 31);   // abs(amp)

        SampleBuffer buf = (SampleBuffer) this.h;
        if (this.loopCount == 0) {
            // One-shot: fade out as the cursor approaches the end of the buffer.
            amp -= amp * this.samplePos / (buf.data.length << 8);
        } else if (this.loopCount >= 0) {
            // Finite-loop remaining: scale by loopStart / bufferLength.
            amp -= amp * this.loopStart / buf.data.length;
        }
        // loopCount < 0 → infinite loop, no amplitude reduction
        return amp > 255 ? 255 : amp;
    }

    /**
     * Advances the read cursor by {@code frames} frames (no sample output is
     * produced — used to skip ahead in time, e.g. when this decoder is off-screen
     * or muted).
     *
     * <p>Also handles glide advancement and loop-boundary logic for the skipped
     * span.
     *
     * @param frames number of output frames to skip
     * obf: va.b(int)
     */
    @Override
    public final synchronized void skip(int frames) {
        // Advance any in-progress volume glide.
        if (this.glideRemaining > 0) {
            if (frames >= this.glideRemaining) {
                // Glide completes within this skip.
                if (this.targetVol == Integer.MIN_VALUE) {
                    // Sentinel: decoder was asked to stop.
                    this.targetVol  = 0;
                    this.curPanRight = 0;
                    this.curPanLeft  = 0;
                    this.curAmp      = 0;
                    this.notifyDone(-27331);    // obf: this.a(-27331)
                    // Don't update frames — fall through so the cursor advances
                    // by the originally requested amount.
                    frames = this.glideRemaining;
                }
                this.glideRemaining = 0;
                snapToTarget();
            } else {
                // Partial glide step.
                this.curAmp       += this.ampDelta      * frames;
                this.curPanLeft   += this.panLeftDelta  * frames;
                this.curPanRight  += this.panRightDelta * frames;
                this.glideRemaining -= frames;
            }
        }

        SampleBuffer buf = (SampleBuffer) this.h;
        int loopStartQ8 = this.loopStart << 8;
        int loopEndQ8   = this.loopEnd   << 8;
        int bufLenQ8    = buf.data.length << 8;
        int loopRangeQ8 = loopEndQ8 - loopStartQ8;
        if (loopRangeQ8 <= 0) {
            this.loopCount = 0;
        }

        // Clamp samplePos to valid range.
        if (this.samplePos < 0) {
            if (this.pitchStep <= 0) {
                cancelGlide();
                this.notifyDone(-27331);
                return;
            }
            this.samplePos = 0;
        }
        if (this.samplePos >= bufLenQ8) {
            if (this.pitchStep >= 0) {
                cancelGlide();
                this.notifyDone(-27331);
                return;
            }
            this.samplePos = bufLenQ8 - 1;
        }

        // Advance cursor by frames * pitchStep.
        this.samplePos += this.pitchStep * frames;

        if (this.loopCount < 0) {
            // --- Infinite loop ---
            if (!this.pingPong) {
                if (this.pitchStep < 0) {
                    if (this.samplePos >= loopStartQ8) return;
                    // Wrap: stay within [loopStart, loopEnd).
                    this.samplePos = loopEndQ8 - 1 - (loopEndQ8 - 1 - this.samplePos) % loopRangeQ8;
                } else {
                    if (this.samplePos < loopEndQ8) return;
                    this.samplePos = loopStartQ8 + (this.samplePos - loopStartQ8) % loopRangeQ8;
                }
            } else {
                // Ping-pong.
                if (this.pitchStep < 0) {
                    if (this.samplePos >= loopStartQ8) return;
                    this.samplePos = loopStartQ8 + loopStartQ8 - 1 - this.samplePos;
                    this.pitchStep = -this.pitchStep;
                }
                while (this.samplePos >= loopEndQ8) {
                    this.samplePos = loopEndQ8 + loopEndQ8 - 1 - this.samplePos;
                    this.pitchStep = -this.pitchStep;
                    if (this.samplePos >= loopStartQ8) return;
                    this.samplePos = loopStartQ8 + loopStartQ8 - 1 - this.samplePos;
                    this.pitchStep = -this.pitchStep;
                }
            }
        } else {
            if (this.loopCount > 0) {
                // --- Finite loop ---
                if (!this.pingPong) {
                    if (this.pitchStep < 0) {
                        if (this.samplePos >= loopStartQ8) return;
                        int bounces = (loopEndQ8 - 1 - this.samplePos) / loopRangeQ8;
                        if (bounces < this.loopCount) {
                            this.samplePos   += loopRangeQ8 * bounces;
                            this.loopCount   -= bounces;
                            return;
                        }
                        this.samplePos += loopRangeQ8 * this.loopCount;
                        this.loopCount  = 0;
                    } else {
                        if (this.samplePos < loopEndQ8) return;
                        int bounces = (this.samplePos - loopStartQ8) / loopRangeQ8;
                        if (bounces < this.loopCount) {
                            this.samplePos   -= loopRangeQ8 * bounces;
                            this.loopCount   -= bounces;
                            return;
                        }
                        this.samplePos -= loopRangeQ8 * this.loopCount;
                        this.loopCount  = 0;
                    }
                } else {
                    // Finite ping-pong.
                    if (this.pitchStep < 0) {
                        if (this.samplePos >= loopStartQ8) return;
                        this.samplePos = loopStartQ8 + loopStartQ8 - 1 - this.samplePos;
                        this.pitchStep = -this.pitchStep;
                        if (--this.loopCount == 0) {
                            // Fall through to one-shot boundary handling below.
                        }
                    }
                    if (this.loopCount != 0) {
                        do {
                            if (this.samplePos < loopEndQ8) return;
                            this.samplePos = loopEndQ8 + loopEndQ8 - 1 - this.samplePos;
                            this.pitchStep = -this.pitchStep;
                            if (--this.loopCount == 0) break;
                            if (this.samplePos >= loopStartQ8) return;
                            this.samplePos = loopStartQ8 + loopStartQ8 - 1 - this.samplePos;
                            this.pitchStep = -this.pitchStep;
                        } while (--this.loopCount != 0);
                    }
                }
            }

            // --- One-shot (or finite loop exhausted): check end-of-buffer ---
            if (this.pitchStep < 0) {
                if (this.samplePos < 0) {
                    this.samplePos = -1;
                    cancelGlide();
                    this.notifyDone(-27331);
                }
            } else {
                if (this.samplePos >= bufLenQ8) {
                    this.samplePos = bufLenQ8;
                    cancelGlide();
                    this.notifyDone(-27331);
                }
            }
        }
    }

    /**
     * Mixes {@code length} output frames into {@code mixBuf} starting at
     * {@code offset}, advancing the sample cursor and applying all loop/glide logic.
     *
     * @param mixBuf the shared integer accumulation buffer
     * @param offset first frame index to write
     * @param length total frames to fill (exclusive end = offset + length)
     * obf: va.b(int[], int, int)
     */
    @Override
    public final synchronized void mix(int[] mixBuf, int offset, int length) {
        if (this.targetVol == 0 && this.glideRemaining == 0) {
            // Silent: just advance the cursor without mixing.
            skip(length);
            return;
        }

        SampleBuffer buf = (SampleBuffer) this.h;
        int loopStartQ8  = this.loopStart << 8;
        int loopEndQ8    = this.loopEnd   << 8;
        int bufLenQ8     = buf.data.length << 8;
        int loopRangeQ8  = loopEndQ8 - loopStartQ8;
        if (loopRangeQ8 <= 0) {
            this.loopCount = 0;
        }

        int writePos = offset;           // current write position in mixBuf
        int endPos   = offset + length;  // exclusive end

        // Clamp samplePos into valid range.
        if (this.samplePos < 0) {
            if (this.pitchStep <= 0) {
                cancelGlide();
                this.notifyDone(-27331);
                return;
            }
            this.samplePos = 0;
        }
        if (this.samplePos >= bufLenQ8) {
            if (this.pitchStep >= 0) {
                cancelGlide();
                this.notifyDone(-27331);
                return;
            }
            this.samplePos = bufLenQ8 - 1;
        }

        if (this.loopCount < 0) {
            // ====== Infinite-loop mode ======
            if (this.pingPong) {
                if (this.pitchStep < 0) {
                    writePos = mixBackward(mixBuf, offset, loopStartQ8, endPos, buf.data[this.loopStart]);
                    if (this.samplePos >= loopStartQ8) return;
                    this.samplePos = loopStartQ8 + loopStartQ8 - 1 - this.samplePos;
                    this.pitchStep = -this.pitchStep;
                }
                while (true) {
                    writePos = mixForward(mixBuf, writePos, loopEndQ8, endPos, buf.data[this.loopEnd - 1]);
                    if (this.samplePos < loopEndQ8) return;
                    this.samplePos = loopEndQ8 + loopEndQ8 - 1 - this.samplePos;
                    this.pitchStep = -this.pitchStep;
                    writePos = mixBackward(mixBuf, writePos, loopStartQ8, endPos, buf.data[this.loopStart]);
                    if (this.samplePos >= loopStartQ8) return;
                    this.samplePos = loopStartQ8 + loopStartQ8 - 1 - this.samplePos;
                    this.pitchStep = -this.pitchStep;
                }
            } else if (this.pitchStep < 0) {
                // Reverse wrap.
                while (true) {
                    writePos = mixBackward(mixBuf, writePos, loopStartQ8, endPos, buf.data[this.loopEnd - 1]);
                    if (this.samplePos >= loopStartQ8) return;
                    this.samplePos = loopEndQ8 - 1 - (loopEndQ8 - 1 - this.samplePos) % loopRangeQ8;
                }
            } else {
                // Forward wrap.
                while (true) {
                    writePos = mixForward(mixBuf, writePos, loopEndQ8, endPos, buf.data[this.loopStart]);
                    if (this.samplePos < loopEndQ8) return;
                    this.samplePos = loopStartQ8 + (this.samplePos - loopStartQ8) % loopRangeQ8;
                }
            }
        } else {
            // ====== One-shot or finite-loop mode ======
            if (this.loopCount > 0) {
                // Finite-loop: play remaining bounces.
                finiteLoopMix:
                if (this.pingPong) {
                    if (this.pitchStep < 0) {
                        writePos = mixBackward(mixBuf, offset, loopStartQ8, endPos, buf.data[this.loopStart]);
                        if (this.samplePos >= loopStartQ8) return;
                        this.samplePos = loopStartQ8 + loopStartQ8 - 1 - this.samplePos;
                        this.pitchStep = -this.pitchStep;
                        if (--this.loopCount == 0) break finiteLoopMix;
                    }
                    do {
                        writePos = mixForward(mixBuf, writePos, loopEndQ8, endPos, buf.data[this.loopEnd - 1]);
                        if (this.samplePos < loopEndQ8) return;
                        this.samplePos = loopEndQ8 + loopEndQ8 - 1 - this.samplePos;
                        this.pitchStep = -this.pitchStep;
                        if (--this.loopCount == 0) break finiteLoopMix;
                        writePos = mixBackward(mixBuf, writePos, loopStartQ8, endPos, buf.data[this.loopStart]);
                        if (this.samplePos >= loopStartQ8) return;
                        this.samplePos = loopStartQ8 + loopStartQ8 - 1 - this.samplePos;
                        this.pitchStep = -this.pitchStep;
                    } while (--this.loopCount != 0);
                } else if (this.pitchStep < 0) {
                    // Finite reverse wrap.
                    while (true) {
                        writePos = mixBackward(mixBuf, writePos, loopStartQ8, endPos, buf.data[this.loopEnd - 1]);
                        if (this.samplePos >= loopStartQ8) return;
                        int bounces = (loopEndQ8 - 1 - this.samplePos) / loopRangeQ8;
                        if (bounces >= this.loopCount) {
                            this.samplePos += loopRangeQ8 * this.loopCount;
                            this.loopCount  = 0;
                            break;
                        }
                        this.samplePos  += loopRangeQ8 * bounces;
                        this.loopCount  -= bounces;
                    }
                } else {
                    // Finite forward wrap.
                    while (true) {
                        writePos = mixForward(mixBuf, writePos, loopEndQ8, endPos, buf.data[this.loopStart]);
                        if (this.samplePos < loopEndQ8) return;
                        int bounces = (this.samplePos - loopStartQ8) / loopRangeQ8;
                        if (bounces >= this.loopCount) {
                            this.samplePos -= loopRangeQ8 * this.loopCount;
                            this.loopCount  = 0;
                            break;
                        }
                        this.samplePos  -= loopRangeQ8 * bounces;
                        this.loopCount  -= bounces;
                    }
                }
            }

            // One-shot (or loop exhausted): play to end of buffer, then stop.
            if (this.pitchStep < 0) {
                mixBackward(mixBuf, writePos, 0, endPos, 0);
                if (this.samplePos < 0) {
                    this.samplePos = -1;
                    cancelGlide();
                    this.notifyDone(-27331);
                }
            } else {
                mixForward(mixBuf, writePos, bufLenQ8, endPos, 0);
                if (this.samplePos >= bufLenQ8) {
                    this.samplePos = bufLenQ8;
                    cancelGlide();
                    this.notifyDone(-27331);
                }
            }
        }
    }

    // -------------------------------------------------------------------------
    // Private volume-glide helpers
    // -------------------------------------------------------------------------

    /**
     * Snap the current amplitude and pan values to match the target immediately
     * (no glide). Called at construction, and when a glide completes.
     * obf: sb.e()
     */
    private final void snapToTarget() {
        this.curAmp      = this.targetVol;
        this.curPanLeft  = computePanLeft(this.targetVol, this.panPosition);
        this.curPanRight = computePanRight(this.targetVol, this.panPosition);
    }

    /**
     * Cancel any in-progress glide and snap to the current target. Used when
     * playback hits a boundary (loop end / one-shot end) while a glide is running.
     * obf: sb.g()
     */
    private final void cancelGlide() {
        if (this.glideRemaining != 0) {
            if (this.targetVol == Integer.MIN_VALUE) {
                this.targetVol = 0;   // sentinel → real zero
            }
            this.glideRemaining = 0;
            snapToTarget();
        }
    }

    /**
     * Compute the next glide step: determines direction and per-frame deltas for
     * {@link #curAmp}, {@link #curPanLeft}, {@link #curPanRight}, and
     * {@link #glideRemaining}.
     *
     * <p>Returns {@code true} if the decoder should be considered done (targetVol
     * was the stop sentinel and has now reached zero).  Returns {@code false}
     * otherwise.
     * obf: sb.f()
     */
    private final boolean computeNextGlide() {
        int vol   = this.targetVol;
        int panL, panR;
        if (vol == Integer.MIN_VALUE) {
            // Stop sentinel: glide toward silence.
            panL = 0;
            panR = 0;
            vol  = 0;
        } else {
            panL = computePanLeft(vol, this.panPosition);
            panR = computePanRight(vol, this.panPosition);
        }

        if (this.curAmp == vol && this.curPanLeft == panL && this.curPanRight == panR) {
            // Already at target.
            if (this.targetVol == Integer.MIN_VALUE) {
                // Stop sentinel reached: zero out and signal done.
                this.targetVol  = 0;
                this.curPanRight = 0;
                this.curPanLeft  = 0;
                this.curAmp      = 0;
                this.notifyDone(-27331);
                return true;
            } else {
                snapToTarget();
                return false;
            }
        }

        // Determine the glide duration: minimum distance across all three axes.
        if (this.curAmp < vol) {
            this.ampDelta     = 1;
            this.glideRemaining = vol - this.curAmp;
        } else if (this.curAmp > vol) {
            this.ampDelta     = -1;
            this.glideRemaining = this.curAmp - vol;
        } else {
            this.ampDelta     = 0;
            // glideRemaining left at 0; will be set by pan axes below
        }

        if (this.curPanLeft < panL) {
            this.panLeftDelta = 1;
            int dist = panL - this.curPanLeft;
            if (this.glideRemaining == 0 || this.glideRemaining > dist) this.glideRemaining = dist;
        } else if (this.curPanLeft > panL) {
            this.panLeftDelta = -1;
            int dist = this.curPanLeft - panL;
            if (this.glideRemaining == 0 || this.glideRemaining > dist) this.glideRemaining = dist;
        } else {
            this.panLeftDelta = 0;
        }

        if (this.curPanRight < panR) {
            this.panRightDelta = 1;
            int dist = panR - this.curPanRight;
            if (this.glideRemaining == 0 || this.glideRemaining > dist) this.glideRemaining = dist;
        } else if (this.curPanRight > panR) {
            this.panRightDelta = -1;
            int dist = this.curPanRight - panR;
            if (this.glideRemaining == 0 || this.glideRemaining > dist) this.glideRemaining = dist;
        } else {
            this.panRightDelta = 0;
        }

        return false;
    }

    // -------------------------------------------------------------------------
    // High-level mixing helpers (called by mix())
    // -------------------------------------------------------------------------

    /**
     * Mix samples forward (pitchStep > 0) from the current cursor toward
     * {@code stopQ8}, writing into {@code mixBuf[writePos..endPos)}.
     *
     * <p>Delegates to one of the four low-level kernels depending on:
     * <ul>
     *   <li>Whether a glide is in progress ({@link #glideRemaining} > 0)</li>
     *   <li>Stereo vs. mono ({@code AudioChannel.isStereo})</li>
     *   <li>Whether the pitch is exactly 1.0x (step == 256, cursor byte-aligned)</li>
     * </ul>
     *
     * @param mixBuf   accumulation buffer
     * @param writePos start write index
     * @param stopQ8   Q8 cursor value at which to stop (exclusive)
     * @param endPos   hard stop (buffer length)
     * @param wrapSample the sample value to use for the last interpolation step
     *                   when the cursor passes the end of real data (wrap guard)
     * @return new writePos after mixing
     * obf: sb.a(int[], int, int, int, int)   [the private instance method]
     */
    private final int mixForward(int[] mixBuf, int writePos, int stopQ8, int endPos, int wrapSample) {
        // Process any pending glide frames first.
        while (this.glideRemaining > 0) {
            int glideEnd = writePos + this.glideRemaining;
            if (glideEnd > endPos) glideEnd = endPos;

            // Save glideRemaining (will be subtracted after mixing).
            this.glideRemaining += writePos;

            if (this.pitchStep == 256 && (this.samplePos & 0xFF) == 0) {
                // Exact 1:1 pitch — no interpolation needed, fast path.
                if (AudioChannel.isStereo) {
                    writePos = mixForwardStereoGlide(0,
                        ((SampleBuffer) this.h).data, mixBuf,
                        this.samplePos, writePos,
                        this.curPanLeft, this.curPanRight, this.panLeftDelta, this.panRightDelta,
                        0, glideEnd, endPos, this);
                } else {
                    writePos = mixForwardMonoGlide(
                        ((SampleBuffer) this.h).data, mixBuf,
                        this.samplePos, writePos,
                        this.curAmp, this.ampDelta,
                        0, glideEnd, endPos, this);
                }
            } else if (AudioChannel.isStereo) {
                writePos = mixForwardStereoGlideInterp(0, 0,
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curPanLeft, this.curPanRight, this.panLeftDelta, this.panRightDelta,
                    0, glideEnd, endPos, this, this.pitchStep, wrapSample);
            } else {
                writePos = mixForwardMonoGlideInterp(0, 0,
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curAmp, this.ampDelta,
                    0, glideEnd, endPos, this, this.pitchStep, wrapSample);
            }

            this.glideRemaining -= writePos;
            if (this.glideRemaining != 0) return writePos;
            if (this.computeNextGlide()) return endPos;
        }

        // No glide: steady-state mix.
        if (this.pitchStep == 256 && (this.samplePos & 0xFF) == 0) {
            if (AudioChannel.isStereo) {
                return mixForwardStereoFlat(0,
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curPanLeft, this.curPanRight, 0, stopQ8, endPos, this);
            } else {
                return mixForwardMonoFlat(
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curAmp, 0, stopQ8, endPos, this);
            }
        } else {
            if (AudioChannel.isStereo) {
                return mixForwardStereoFlatInterp(0, 0,
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curPanLeft, this.curPanRight, 0, stopQ8, endPos, this, this.pitchStep, wrapSample);
            } else {
                return mixForwardMonoFlatInterp(0, 0,
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curAmp, 0, stopQ8, endPos, this, this.pitchStep, wrapSample);
            }
        }
    }

    /**
     * Mix samples backward (pitchStep < 0) from the current cursor toward
     * {@code stopQ8}, writing into {@code mixBuf[writePos..endPos)}.
     *
     * Mirrors {@link #mixForward} but uses the reverse-reading kernels.
     *
     * obf: sb.b(int[], int, int, int, int)   [the private instance method]
     */
    private final int mixBackward(int[] mixBuf, int writePos, int stopQ8, int endPos, int wrapSample) {
        while (this.glideRemaining > 0) {
            int glideEnd = writePos + this.glideRemaining;
            if (glideEnd > endPos) glideEnd = endPos;

            this.glideRemaining += writePos;

            if (this.pitchStep == -256 && (this.samplePos & 0xFF) == 0) {
                if (AudioChannel.isStereo) {
                    writePos = mixBackwardStereoGlide(0,
                        ((SampleBuffer) this.h).data, mixBuf,
                        this.samplePos, writePos,
                        this.curPanLeft, this.curPanRight, this.panLeftDelta, this.panRightDelta,
                        0, glideEnd, endPos, this);
                } else {
                    writePos = mixBackwardMonoGlide(
                        ((SampleBuffer) this.h).data, mixBuf,
                        this.samplePos, writePos,
                        this.curAmp, this.ampDelta,
                        0, glideEnd, endPos, this);
                }
            } else if (AudioChannel.isStereo) {
                writePos = mixBackwardStereoGlideInterp(0, 0,
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curPanLeft, this.curPanRight, this.panLeftDelta, this.panRightDelta,
                    0, glideEnd, endPos, this, this.pitchStep, wrapSample);
            } else {
                writePos = mixBackwardMonoGlideInterp(0, 0,
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curAmp, this.ampDelta,
                    0, glideEnd, endPos, this, this.pitchStep, wrapSample);
            }

            this.glideRemaining -= writePos;
            if (this.glideRemaining != 0) return writePos;
            if (this.computeNextGlide()) return endPos;
        }

        if (this.pitchStep == -256 && (this.samplePos & 0xFF) == 0) {
            if (AudioChannel.isStereo) {
                return mixBackwardStereoFlat(0,
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curPanLeft, this.curPanRight, 0, stopQ8, endPos, this);
            } else {
                return mixBackwardMonoFlat(
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curAmp, 0, stopQ8, endPos, this);
            }
        } else {
            if (AudioChannel.isStereo) {
                return mixBackwardStereoFlatInterp(0, 0,
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curPanLeft, this.curPanRight, 0, stopQ8, endPos, this, this.pitchStep, wrapSample);
            } else {
                return mixBackwardMonoFlatInterp(0, 0,
                    ((SampleBuffer) this.h).data, mixBuf,
                    this.samplePos, writePos,
                    this.curAmp, 0, stopQ8, endPos, this, this.pitchStep, wrapSample);
            }
        }
    }

    // =========================================================================
    // Low-level mixing kernels
    // =========================================================================
    // Each kernel is a static helper to avoid object overhead in the tight mix loop.
    // Naming convention:
    //   mixForward  / mixBackward  — cursor direction
    //   Mono / Stereo              — channel count
    //   Flat / Glide               — constant vs. ramped amplitude
    //   Interp                     — linear interpolation between samples (for resampling)
    // All kernels update self.samplePos (v) and self.curAmp/curPanLeft/curPanRight
    // via the SoundDecoder reference, and return the new writePos.
    // =========================================================================

    // -------------------------------------------------------------------------
    // MONO FLAT FORWARD  (exact pitch, no glide)
    // "b(byte[], int[], int, int, int, int, int, int, SoundDecoder)"  [8-param variant]
    // -------------------------------------------------------------------------

    /**
     * Mono forward, flat amplitude, exact 1:1 pitch (no interpolation).
     * Reads source bytes forward; cursor advances by 1 byte per frame (step=256).
     * Writes one int per frame to {@code mixBuf}.
     *
     * @param src       raw PCM sample data
     * @param mixBuf    output accumulation buffer
     * @param srcPosQ8  Q8 read cursor (low bits must be zero for this path)
     * @param writePos  start write index
     * @param amp       amplitude (Q-6 fixed-point, pre-scaled by 4: amp<<2 done inside)
     * @param unused    padding parameter (always 0)
     * @param stopQ8    Q8 cursor value at which to stop
     * @param endWrite  hard stop write index
     * @param self      SoundDecoder to update (samplePos)
     * @return new writePos
     * obf: sb.b(byte[], int[], int, int, int, int, int, int, sb)   [9-param]
     */
    private static final int mixForwardMonoFlat(
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int amp, int unused,
            int stopQ8, int endWrite,
            SoundDecoder self) {
        int srcIdx = srcPosQ8 >> 8;
        // clean: var5 = var3 + (var7>>8) - (var2>>8); i.e. limit = writePos + (stopQ8>>8) - srcIdx
        int limit = writePos + (stopQ8 >> 8) - srcIdx;   // frames until stopQ8
        // Clamp: don't exceed endWrite frames.
        if (limit > endWrite) limit = endWrite;
        limit -= 3;   // unroll by 4

        // Amplitude is stored Q-2 in the caller (amp << 2 done externally),
        // we replicate by doing it here: ampScaled = amp << 2
        int ampScaled = amp << 2;

        // Unrolled 4x for performance.
        while (writePos < limit) {
            mixBuf[writePos++] += src[srcIdx++] * ampScaled;
            mixBuf[writePos++] += src[srcIdx++] * ampScaled;
            mixBuf[writePos++] += src[srcIdx++] * ampScaled;
            mixBuf[writePos++] += src[srcIdx++] * ampScaled;
        }
        limit += 3;
        while (writePos < limit) {
            mixBuf[writePos++] += src[srcIdx++] * ampScaled;
        }

        self.samplePos = srcIdx << 8;
        return writePos;
        // obf: sb.b(byte[], int[], int, int, int, int, int, int, sb) — 9-param version
    }

    // -------------------------------------------------------------------------
    // MONO FLAT BACKWARD  (exact pitch, no glide, reverse)
    // "a(byte[], int[], int, int, int, int, int, int, sb)"  [9-param]
    // -------------------------------------------------------------------------

    /**
     * Mono backward, flat amplitude, exact 1:1 pitch (reverse playback, step=-256).
     * Reads source bytes in reverse; cursor moves toward index 0.
     *
     * @param src       raw PCM sample data
     * @param mixBuf    output accumulation buffer
     * @param srcPosQ8  Q8 read cursor
     * @param writePos  start write index
     * @param amp       amplitude
     * @param unused    padding (always 0)
     * @param stopQ8    Q8 cursor stop value (exclusive lower bound)
     * @param endWrite  hard stop
     * @param self      SoundDecoder to update
     * @return new writePos
     * obf: sb.a(byte[], int[], int, int, int, int, int, int, sb)   [9-param]
     */
    private static final int mixBackwardMonoFlat(
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int amp, int unused,
            int stopQ8, int endWrite,
            SoundDecoder self) {
        int srcIdx   = srcPosQ8 >> 8;
        int stopIdx  = stopQ8  >> 8;
        // Number of frames until we hit stopQ8 (moving backward).
        int limit = writePos + srcIdx - (stopIdx - 1);
        if (limit > endWrite) limit = endWrite;
        limit -= 3;

        int ampScaled = amp << 2;

        while (writePos < limit) {
            mixBuf[writePos++] += src[srcIdx--] * ampScaled;
            mixBuf[writePos++] += src[srcIdx--] * ampScaled;
            mixBuf[writePos++] += src[srcIdx--] * ampScaled;
            mixBuf[writePos++] += src[srcIdx--] * ampScaled;
        }
        limit += 3;
        while (writePos < limit) {
            mixBuf[writePos++] += src[srcIdx--] * ampScaled;
        }

        self.samplePos = srcIdx << 8;
        return writePos;
        // obf: sb.a(byte[], int[], int, int, int, int, int, int, sb)
    }

    // -------------------------------------------------------------------------
    // MONO GLIDE FORWARD  (exact pitch, ramped amplitude)
    // -------------------------------------------------------------------------

    /**
     * Mono forward, ramping amplitude glide, exact 1:1 pitch.
     * {@code curAmp} steps by {@code ampDelta} per frame.
     *
     * @param src           raw PCM data
     * @param mixBuf        output buffer
     * @param srcPosQ8      Q8 read cursor
     * @param writePos      start write index
     * @param curAmp        current amplitude (Q-6, pre-shifted ×4)
     * @param ampDelta      per-frame amplitude change (Q-6, pre-shifted ×4)
     * @param unused        padding
     * @param stopWrite     stop write index
     * @param endWrite      hard stop
     * @param self          SoundDecoder to update
     * @return new writePos
     * obf: sb.b(byte[], int[], int, int, int, int, int, int, int, sb)  [10-param]
     *
     * Note: curAmp and ampDelta are stored ×4 (left-shifted by 2) in the original
     * to avoid per-sample divisions.  The result is right-shifted back by 2 at the
     * end when stored into self.curAmp.
     */
    private static final int mixForwardMonoGlide(
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int curAmp, int ampDelta,
            int unused, int stopWrite, int endWrite,
            SoundDecoder self) {
        int srcIdx = srcPosQ8 >> 8;
        // clean: var6 = var3 + (var8>>8) - var2  →  limit = writePos + (endWrite>>8) - srcIdx
        int limit = writePos + (endWrite >> 8) - srcIdx;   // frames until stop
        if (limit > stopWrite) limit = stopWrite;   // clamp

        // Mono path ramps the amplitude per-sample (curAmp tracked by the local
        // ampScaled below); the pan channels are not interpolated in the loop, so
        // they are advanced in bulk here.  clean: var9.k/var9.m += delta * (var6 - var3).
        self.curPanLeft   += self.panLeftDelta  * (limit - writePos);
        self.curPanRight  += self.panRightDelta * (limit - writePos);
        limit -= 3;

        int ampScaled      = curAmp << 2;
        int ampDeltaScaled = ampDelta << 2;

        while (writePos < limit) {
            mixBuf[writePos++] += src[srcIdx++] * ampScaled;
            ampScaled          += ampDeltaScaled;
            mixBuf[writePos++] += src[srcIdx++] * ampScaled;
            ampScaled          += ampDeltaScaled;
            mixBuf[writePos++] += src[srcIdx++] * ampScaled;
            ampScaled          += ampDeltaScaled;
            mixBuf[writePos++] += src[srcIdx++] * ampScaled;
            ampScaled          += ampDeltaScaled;
        }
        limit += 3;
        while (writePos < limit) {
            mixBuf[writePos++] += src[srcIdx++] * ampScaled;
            ampScaled          += ampDeltaScaled;
        }

        self.curAmp    = ampScaled >> 2;
        self.samplePos = srcIdx << 8;
        return writePos;
        // obf: sb.b(byte[], int[], int, int, int, int, int, int, int, sb) — 10-param
    }

    // -------------------------------------------------------------------------
    // MONO GLIDE BACKWARD  (exact pitch, ramped amplitude, reverse)
    // -------------------------------------------------------------------------

    /**
     * Mono backward, ramping amplitude glide, exact 1:1 reverse pitch.
     *
     * obf: sb.b(byte[], int[], int, int, int, int, int, int, sb)   [9-param, different from MonoFlat]
     *
     * (Vineflower disambiguates by parameter count; this is the 9-param backward glide.)
     */
    private static final int mixBackwardMonoGlide(
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int curAmp, int ampDelta,
            int unused, int stopWrite, int endWrite,
            SoundDecoder self) {
        int srcIdx  = srcPosQ8 >> 8;
        int stopIdx = endWrite >> 8;
        int limit   = writePos + srcIdx - (stopIdx - 1);
        if (limit > stopWrite) limit = stopWrite;

        // Mono ramps amplitude per-sample (via ampScaled); only the pan channels
        // are advanced in bulk.  clean: var9.k/var9.m += delta * (var6 - var3).
        self.curPanLeft   += self.panLeftDelta  * (limit - writePos);
        self.curPanRight  += self.panRightDelta * (limit - writePos);
        limit -= 3;

        int ampScaled      = curAmp << 2;
        int ampDeltaScaled = ampDelta << 2;

        while (writePos < limit) {
            mixBuf[writePos++] += src[srcIdx--] * ampScaled;
            ampScaled          += ampDeltaScaled;
            mixBuf[writePos++] += src[srcIdx--] * ampScaled;
            ampScaled          += ampDeltaScaled;
            mixBuf[writePos++] += src[srcIdx--] * ampScaled;
            ampScaled          += ampDeltaScaled;
            mixBuf[writePos++] += src[srcIdx--] * ampScaled;
            ampScaled          += ampDeltaScaled;
        }
        limit += 3;
        while (writePos < limit) {
            mixBuf[writePos++] += src[srcIdx--] * ampScaled;
            ampScaled          += ampDeltaScaled;
        }

        self.curAmp    = ampScaled >> 2;
        self.samplePos = srcIdx << 8;
        return writePos;
        // obf: backward variant of b(byte[], int[], ..., sb) — 9-param
    }

    // -------------------------------------------------------------------------
    // MONO GLIDE FORWARD INTERPOLATED  (arbitrary pitch, ramped amplitude)
    // -------------------------------------------------------------------------

    /**
     * Mono forward, ramping amplitude glide, arbitrary pitch with linear interpolation.
     *
     * <p>Interpolation: for a Q8 cursor position {@code pos}:
     * <pre>
     *   idx   = pos >> 8
     *   frac  = pos & 0xFF
     *   s0    = data[idx]          (integer sample, before boundary)
     *   s1    = data[idx+1]        (next sample)
     *   interp = (s0 << 8) + (s1 - s0) * frac   // 16-bit lerped value
     *   out   += interp * gain >> 6
     * </pre>
     * When the cursor is past the end of real data, {@code wrapSample} is used
     * as the "next" sample value (e.g. the first sample of the next loop).
     *
     * @param unused0   padding params (always 0, consumed by obf overload resolution)
     * @param unused1   padding
     * @param src       raw PCM data
     * @param mixBuf    accumulation buffer
     * @param srcPosQ8  Q8 read cursor
     * @param writePos  start write index
     * @param curAmp    current amplitude (Q-6)
     * @param ampDelta  per-frame amplitude delta (Q-6), 0 for flat
     * @param unused2   padding
     * @param stopWrite stop write index
     * @param endWrite  hard stop
     * @param self      SoundDecoder to update
     * @param step      pitchStep (Q8)
     * @param wrapSample value used as next-sample past data boundary
     * @return new writePos
     * obf: sb.a(int, int, byte[], int[], int, int, int, int, int, int, int, sb, int, int)  [14-param, mono forward]
     */
    private static final int mixForwardMonoGlideInterp(
            int unused0, int unused1,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int curAmp, int ampDelta,
            int unused2, int stopWrite, int endWrite,
            SoundDecoder self,
            int step, int wrapSample) {
        // Advance the (non-interpolated) pan channels in bulk: subtract the initial
        // writePos contribution now, add the final writePos contribution at the end,
        // giving curPan{Left,Right} += delta * framesWritten.  clean (var11.k/.m).
        self.curPanLeft  -= self.panLeftDelta  * writePos;
        self.curPanRight -= self.panRightDelta * writePos;

        // Phase 1: while cursor is within real data.
        // Compute how many frames until we cross the data boundary.
        int framesInRange;
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            // clean: var5 + (var10 - var4 + var12 - 257) / var12  (no extra -256)
            framesInRange = writePos + (endWrite - srcPosQ8 + step - 257) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        while (writePos < framesInRange) {
            int idx  = srcPosQ8 >> 8;
            int s0   = src[idx];
            // Linear interpolation: (s0 << 8) + (s1 - s0) * frac
            int samp = (s0 << 8) + (src[idx + 1] - s0) * (srcPosQ8 & 0xFF);
            mixBuf[writePos++] += samp * curAmp >> 6;
            curAmp   += ampDelta;
            srcPosQ8 += step;
        }

        // Phase 2: cursor is near or past the end of real data — use wrapSample.
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = writePos + (endWrite - srcPosQ8 + step - 1) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        int ws = wrapSample;
        while (writePos < framesInRange) {
            int s0   = src[srcPosQ8 >> 8];
            // Lerp between last real sample and wrapSample.
            int samp = (s0 << 8) + (ws - s0) * (srcPosQ8 & 0xFF);
            mixBuf[writePos++] += samp * curAmp >> 6;
            curAmp   += ampDelta;
            srcPosQ8 += step;
        }

        self.curPanLeft  += self.panLeftDelta  * writePos;
        self.curPanRight += self.panRightDelta * writePos;
        self.curAmp     = curAmp;
        self.samplePos  = srcPosQ8;
        return writePos;
        // obf: sb.b(int, int, byte[], int[], int, int, int, int, int, int, int, sb, int, int) — 14-param mono forward GLIDE (clean L1172)
    }

    // -------------------------------------------------------------------------
    // MONO FLAT FORWARD INTERPOLATED  (arbitrary pitch, constant amplitude)
    // -------------------------------------------------------------------------

    /**
     * Mono forward, constant amplitude, arbitrary pitch with linear interpolation.
     * Steady-state (no-glide) counterpart of {@link #mixForwardMonoGlideInterp}:
     * the amplitude {@code amp} is NOT ramped and no curAmp/pan writeback occurs
     * (only {@code samplePos} is updated).  clean: sb.a(...13-param) at L907.
     *
     * obf: sb.a(int, int, byte[], int[], int, int, int, int, int, sb, int, int) — 13-param mono forward FLAT
     */
    private static final int mixForwardMonoFlatInterp(
            int unused0, int unused1,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int amp, int unused2,
            int stopWrite, int endWrite,
            SoundDecoder self,
            int step, int wrapSample) {
        // Phase 1: cursor within real data.
        int framesInRange;
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = writePos + (endWrite - srcPosQ8 + step - 257) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        while (writePos < framesInRange) {
            int idx  = srcPosQ8 >> 8;
            int s0   = src[idx];
            int samp = (s0 << 8) + (src[idx + 1] - s0) * (srcPosQ8 & 0xFF);
            mixBuf[writePos++] += samp * amp >> 6;
            srcPosQ8 += step;
        }

        // Phase 2: wrap boundary — use wrapSample.
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = writePos + (endWrite - srcPosQ8 + step - 1) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        int ws = wrapSample;
        while (writePos < framesInRange) {
            int s0   = src[srcPosQ8 >> 8];
            int samp = (s0 << 8) + (ws - s0) * (srcPosQ8 & 0xFF);
            mixBuf[writePos++] += samp * amp >> 6;
            srcPosQ8 += step;
        }

        self.samplePos = srcPosQ8;
        return writePos;
        // obf: sb.a(int, int, byte[], int[], int, int, int, int, int, int, int, sb, int, int) — clean L907
    }

    // -------------------------------------------------------------------------
    // MONO GLIDE BACKWARD INTERPOLATED  (arbitrary pitch, ramped amplitude, reverse)
    // -------------------------------------------------------------------------

    /**
     * Mono backward, ramping amplitude glide, arbitrary pitch with linear interpolation.
     *
     * <p>Reads one sample behind the cursor: {@code idx = (pos >> 8) - 1},
     * then interpolates with {@code data[idx]} at fraction {@code pos & 0xFF}.
     * This gives the sample pair "just behind" the cursor for backward playback.
     *
     * <p>Glide version (clean L429, sb.c): ramps {@code curAmp} by {@code ampDelta}
     * per frame and bulk-advances the pan channels + writes curAmp back.
     *
     * obf: sb.c(int, int, byte[], int[], int, int, int, int, int, int, int, sb, int, int)  [14-param, mono back GLIDE]
     */
    private static final int mixBackwardMonoGlideInterp(
            int unused0, int unused1,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int curAmp, int ampDelta,
            int unused2, int stopWrite, int endWrite,
            SoundDecoder self,
            int step, int wrapSample) {
        // Bulk-advance the (non-interpolated) pan channels.  clean: var11.k/.m.
        self.curPanLeft  -= self.panLeftDelta  * writePos;
        self.curPanRight -= self.panRightDelta * writePos;

        // Phase 1: cursor still within real data (srcPosQ8 > 256).
        int framesInRange;
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = writePos + (endWrite + 256 - srcPosQ8 + step) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        while (writePos < framesInRange) {
            int idx = srcPosQ8 >> 8;
            int s0  = src[idx - 1];
            // Backward interpolation: s0 is "at" the cursor-1 sample.
            int samp = (s0 << 8) + (src[idx] - s0) * (srcPosQ8 & 0xFF);
            mixBuf[writePos++] += samp * curAmp >> 6;
            curAmp   += ampDelta;
            srcPosQ8 += step;  // step is negative for backward
        }

        // Phase 2: cursor is at or near the wrap boundary — use wrapSample.
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = writePos + (endWrite - srcPosQ8 + step) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        int ws = wrapSample;
        while (writePos < framesInRange) {
            int s0   = src[srcPosQ8 >> 8];
            // clean L453: ((var0<<8) + (data[idx]-var0)*frac), var0 = wrapSample
            int samp = (ws << 8) + (s0 - ws) * (srcPosQ8 & 0xFF);
            mixBuf[writePos++] += samp * curAmp >> 6;
            curAmp      += ampDelta;
            srcPosQ8    += step;
        }

        self.curPanLeft  += self.panLeftDelta  * writePos;
        self.curPanRight += self.panRightDelta * writePos;
        self.curAmp     = curAmp;
        self.samplePos  = srcPosQ8;
        return writePos;
        // obf: sb.c(int, int, byte[], int[], ..., sb, int, int) — 14-param mono backward GLIDE (clean L429)
    }

    // -------------------------------------------------------------------------
    // MONO FLAT BACKWARD INTERPOLATED  (arbitrary pitch, constant amplitude, reverse)
    // -------------------------------------------------------------------------

    /**
     * Mono backward, constant amplitude, arbitrary pitch with linear interpolation.
     * Steady-state (no-glide) counterpart of {@link #mixBackwardMonoGlideInterp}:
     * amplitude is constant, no curAmp/pan writeback (only samplePos).
     * clean: sb.b(...13-param) at L842.
     *
     * obf: sb.b(int, int, byte[], int[], int, int, int, int, int, sb, int, int) — 13-param mono back FLAT
     */
    private static final int mixBackwardMonoFlatInterp(
            int unused0, int unused1,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int amp, int unused2,
            int stopWrite, int endWrite,
            SoundDecoder self,
            int step, int wrapSample) {
        // Phase 1: cursor still within real data.
        int framesInRange;
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = writePos + (endWrite + 256 - srcPosQ8 + step) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        while (writePos < framesInRange) {
            int idx = srcPosQ8 >> 8;
            int s0  = src[idx - 1];
            int samp = (s0 << 8) + (src[idx] - s0) * (srcPosQ8 & 0xFF);
            mixBuf[writePos++] += samp * amp >> 6;
            srcPosQ8 += step;
        }

        // Phase 2: wrap boundary — use wrapSample.
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = writePos + (endWrite - srcPosQ8 + step) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        int ws = wrapSample;
        while (writePos < framesInRange) {
            int s0   = src[srcPosQ8 >> 8];
            int samp = (ws << 8) + (s0 - ws) * (srcPosQ8 & 0xFF);
            mixBuf[writePos++] += samp * amp >> 6;
            srcPosQ8 += step;
        }

        self.samplePos = srcPosQ8;
        return writePos;
        // obf: sb.b(int, int, byte[], int[], ..., sb, int, int) — clean L842
    }

    // -------------------------------------------------------------------------
    // STEREO FLAT FORWARD  (exact pitch, no glide)
    // -------------------------------------------------------------------------

    /**
     * Stereo forward, flat amplitude, exact 1:1 pitch.
     * Writes two ints per frame: {@code mixBuf[2i]=left, mixBuf[2i+1]=right}.
     *
     * <p>FLAT (steady-state) version: pans are constant, no curAmp/pan writeback
     * (only samplePos).  clean: sb.a(...11-param) at L348.
     *
     * @param unused    padding
     * @param src       raw PCM data
     * @param mixBuf    interleaved stereo accumulation buffer
     * @param srcPosQ8  Q8 read cursor
     * @param writePos  start write-pair index (mono frames)
     * @param panLeft   left gain (Q-6)
     * @param panRight  right gain (Q-6)
     * @param unused2   padding (clean var7, overwritten as the limit)
     * @param stopWrite stop write-pair index (clamp)
     * @param endWrite  source-stop Q8 position
     * @param self      SoundDecoder to update
     * @return new writePos (mono frame count)
     * obf: sb.a(int, byte[], int[], int, int, int, int, int, int, int, sb)  [11-param stereo forward FLAT]
     */
    private static final int mixForwardStereoFlat(
            int unused,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int panLeft, int panRight,
            int unused2, int stopWrite, int endWrite,
            SoundDecoder self) {
        int srcIdx = srcPosQ8 >> 8;
        // clean: var7 = var4 + (var9>>8) - var3  →  limit = writePos + (endWrite>>8) - srcIdx
        int limit = writePos + (endWrite >> 8) - srcIdx;
        if (limit > stopWrite) limit = stopWrite;

        int outL = writePos << 1;
        int outR = limit  << 1;
        outR -= 6;   // unroll by 4 pairs

        int pL = panLeft  << 2;
        int pR = panRight << 2;

        while (outL < outR) {
            byte s;
            s = src[srcIdx++]; mixBuf[outL++] += s * pL; mixBuf[outL++] += s * pR;
            s = src[srcIdx++]; mixBuf[outL++] += s * pL; mixBuf[outL++] += s * pR;
            s = src[srcIdx++]; mixBuf[outL++] += s * pL; mixBuf[outL++] += s * pR;
            s = src[srcIdx++]; mixBuf[outL++] += s * pL; mixBuf[outL++] += s * pR;
        }
        outR += 6;
        while (outL < outR) {
            byte s = src[srcIdx++];
            mixBuf[outL++] += s * pL;
            mixBuf[outL++] += s * pR;
        }

        // FLAT: no curAmp/curPanLeft/curPanRight writeback (clean L348 only sets v).
        self.samplePos   = srcIdx << 8;
        return outL >> 1;
        // obf: sb.a(int, byte[], int[], int, int, int, int, int, int, int, sb) — 11-param FLAT (clean L348)
    }

    // -------------------------------------------------------------------------
    // STEREO FLAT BACKWARD  (exact pitch, no glide, reverse)
    // -------------------------------------------------------------------------

    /**
     * Stereo backward, flat amplitude, exact 1:1 pitch, reverse playback.
     *
     * <p>FLAT (steady-state) version: pans constant, no curAmp/pan writeback
     * (only samplePos).  clean: sb.b(...11-param) at L651.
     *
     * obf: sb.b(int, byte[], int[], int, int, int, int, int, int, int, sb)  [11-param stereo back FLAT]
     */
    private static final int mixBackwardStereoFlat(
            int unused,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int panLeft, int panRight,
            int unused2, int stopWrite, int endWrite,
            SoundDecoder self) {
        int srcIdx = srcPosQ8 >> 8;
        int stopIdx = endWrite >> 8;
        // clean: var7 = var4 + var3 - (var9 - 1)  →  limit = writePos + srcIdx - ((endWrite>>8)-1)
        int limit  = writePos + srcIdx - (stopIdx - 1);
        if (limit > stopWrite) limit = stopWrite;

        int outL = writePos << 1;
        int outR = limit   << 1;
        outR -= 6;

        int pL = panLeft  << 2;
        int pR = panRight << 2;

        while (outL < outR) {
            byte s;
            s = src[srcIdx--]; mixBuf[outL++] += s * pL; mixBuf[outL++] += s * pR;
            s = src[srcIdx--]; mixBuf[outL++] += s * pL; mixBuf[outL++] += s * pR;
            s = src[srcIdx--]; mixBuf[outL++] += s * pL; mixBuf[outL++] += s * pR;
            s = src[srcIdx--]; mixBuf[outL++] += s * pL; mixBuf[outL++] += s * pR;
        }
        outR += 6;
        while (outL < outR) {
            byte s = src[srcIdx--];
            mixBuf[outL++] += s * pL;
            mixBuf[outL++] += s * pR;
        }

        // FLAT: no curAmp/curPanLeft/curPanRight writeback (clean L651 only sets v).
        self.samplePos   = srcIdx << 8;
        return outL >> 1;
        // obf: sb.b(int, byte[], ..., sb) — 11-param backward FLAT stereo (clean L651)
    }

    // -------------------------------------------------------------------------
    // STEREO GLIDE FORWARD  (exact pitch, ramped L/R pans)
    // -------------------------------------------------------------------------

    /**
     * Stereo forward, ramping L/R amplitude glide, exact 1:1 pitch.
     * Both {@code panLeft} and {@code panRight} increment by their respective
     * delta each frame.
     *
     * obf: sb.a(int, byte[], int[], int, int, int, int, int, int, int, int, sb)  [12-param forward stereo glide]
     */
    private static final int mixForwardStereoGlide(
            int unused,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int panLeft, int panRight,
            int panLeftDelta, int panRightDelta,
            int unused2, int stopWrite, int endWrite,
            SoundDecoder self) {
        int srcIdx = srcPosQ8 >> 8;
        int limit  = writePos + (endWrite >> 8) - srcIdx;
        if (limit > stopWrite) limit = stopWrite;

        // Accumulate glide position change.
        self.curAmp += self.ampDelta * (limit - writePos);

        int outL = writePos << 1;
        int outR = limit   << 1;
        outR -= 6;

        int pL  = panLeft       << 2;
        int pR  = panRight      << 2;
        int dpL = panLeftDelta  << 2;
        int dpR = panRightDelta << 2;

        while (outL < outR) {
            byte s;
            s = src[srcIdx++]; mixBuf[outL++] += s * pL; pL += dpL; mixBuf[outL++] += s * pR; pR += dpR;
            s = src[srcIdx++]; mixBuf[outL++] += s * pL; pL += dpL; mixBuf[outL++] += s * pR; pR += dpR;
            s = src[srcIdx++]; mixBuf[outL++] += s * pL; pL += dpL; mixBuf[outL++] += s * pR; pR += dpR;
            s = src[srcIdx++]; mixBuf[outL++] += s * pL; pL += dpL; mixBuf[outL++] += s * pR; pR += dpR;
        }
        outR += 6;
        while (outL < outR) {
            byte s = src[srcIdx++];
            mixBuf[outL++] += s * pL; pL += dpL;
            mixBuf[outL++] += s * pR; pR += dpR;
        }

        self.curPanLeft  = pL >> 2;
        self.curPanRight = pR >> 2;
        self.samplePos   = srcIdx << 8;
        return outL >> 1;
        // obf: a(int, byte[], ..., sb) — 12-param forward stereo glide
    }

    // -------------------------------------------------------------------------
    // STEREO GLIDE BACKWARD  (exact pitch, ramped L/R pans, reverse)
    // -------------------------------------------------------------------------

    /**
     * Stereo backward, ramping L/R amplitude glide, exact 1:1 pitch, reverse.
     *
     * obf: sb.b(int, byte[], int[], int, int, int, int, int, int, int, int, sb)  [12-param backward stereo glide]
     */
    private static final int mixBackwardStereoGlide(
            int unused,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int panLeft, int panRight,
            int panLeftDelta, int panRightDelta,
            int unused2, int stopWrite, int endWrite,
            SoundDecoder self) {
        int srcIdx  = srcPosQ8 >> 8;
        int stopIdx = endWrite >> 8;
        int limit   = writePos + srcIdx - (stopIdx - 1);
        if (limit > stopWrite) limit = stopWrite;

        self.curAmp += self.ampDelta * (limit - writePos);

        int outL = writePos << 1;
        int outR = limit   << 1;
        outR -= 6;

        int pL  = panLeft       << 2;
        int pR  = panRight      << 2;
        int dpL = panLeftDelta  << 2;
        int dpR = panRightDelta << 2;

        while (outL < outR) {
            byte s;
            s = src[srcIdx--]; mixBuf[outL++] += s * pL; pL += dpL; mixBuf[outL++] += s * pR; pR += dpR;
            s = src[srcIdx--]; mixBuf[outL++] += s * pL; pL += dpL; mixBuf[outL++] += s * pR; pR += dpR;
            s = src[srcIdx--]; mixBuf[outL++] += s * pL; pL += dpL; mixBuf[outL++] += s * pR; pR += dpR;
            s = src[srcIdx--]; mixBuf[outL++] += s * pL; pL += dpL; mixBuf[outL++] += s * pR; pR += dpR;
        }
        outR += 6;
        while (outL < outR) {
            byte s = src[srcIdx--];
            mixBuf[outL++] += s * pL; pL += dpL;
            mixBuf[outL++] += s * pR; pR += dpR;
        }

        self.curPanLeft  = pL >> 2;
        self.curPanRight = pR >> 2;
        self.samplePos   = srcIdx << 8;
        return outL >> 1;
        // obf: b(int, byte[], ..., sb) — 12-param backward stereo glide
    }

    // -------------------------------------------------------------------------
    // STEREO GLIDE FORWARD INTERPOLATED  (arbitrary pitch, ramped L/R, forward)
    // -------------------------------------------------------------------------

    /**
     * Stereo forward, ramping L/R amplitude glide, arbitrary pitch with linear
     * interpolation.  This is the most general forward kernel.
     *
     * <p>Writes two ints per frame (left/right interleaved).  Both gains step
     * independently each frame via their respective deltas.
     *
     * obf: sb.a(int, int, byte[], int[], int, int, int, int, int, int, int, int, int, sb, int, int)  [16-param forward stereo glide interp]
     */
    private static final int mixForwardStereoGlideInterp(
            int unused0, int unused1,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int panLeft, int panRight,
            int panLeftDelta, int panRightDelta,
            int unused2, int stopWrite, int endWrite,
            SoundDecoder self,
            int step, int wrapSample) {
        // Remove already-completed frames from running pan totals.
        self.curAmp       -= self.ampDelta      * writePos;   // (recalculated below)
        // Phase 1: cursor within real data.
        int framesInRange;
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = writePos + (endWrite - srcPosQ8 + step - 257) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        int outL = writePos << 1;
        int outEnd = framesInRange << 1;

        while (outL < outEnd) {
            int idx  = srcPosQ8 >> 8;
            int s0   = src[idx];
            int samp = (s0 << 8) + (src[idx + 1] - s0) * (srcPosQ8 & 0xFF);
            mixBuf[outL++] += samp * panLeft  >> 6;
            panLeft        += panLeftDelta;
            mixBuf[outL++] += samp * panRight >> 6;
            panRight       += panRightDelta;
            srcPosQ8       += step;
        }

        // Phase 2: wrap boundary — use wrapSample.
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = (outL >> 1) + (endWrite - srcPosQ8 + step - 1) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }
        outEnd = framesInRange << 1;

        int ws = wrapSample;
        while (outL < outEnd) {
            int s0   = src[srcPosQ8 >> 8];
            int samp = (s0 << 8) + (ws - s0) * (srcPosQ8 & 0xFF);
            mixBuf[outL++] += samp * panLeft  >> 6;
            panLeft        += panLeftDelta;
            mixBuf[outL++] += samp * panRight >> 6;
            panRight       += panRightDelta;
            srcPosQ8       += step;
        }

        outL >>= 1;   // convert back to mono frame count
        self.curAmp       += self.ampDelta      * outL;
        self.curPanLeft  = panLeft;
        self.curPanRight = panRight;
        self.samplePos   = srcPosQ8;
        return outL;
        // obf: sb.a(int, int, byte[], ..., sb, int, int) — 16-param forward stereo interp GLIDE (clean L505)
    }

    // -------------------------------------------------------------------------
    // STEREO FLAT FORWARD INTERPOLATED  (arbitrary pitch, constant L/R pans)
    // -------------------------------------------------------------------------

    /**
     * Stereo forward, constant L/R pans, arbitrary pitch with linear interpolation.
     * Steady-state (no-glide) counterpart of {@link #mixForwardStereoGlideInterp}:
     * pans are constant, no curAmp/pan writeback (only samplePos).
     * clean: sb.a(...14-param) at L244.
     *
     * obf: sb.a(int, int, byte[], int[], int, int, int, int, int, int, int, sb, int, int) — 14-param fwd stereo FLAT
     */
    private static final int mixForwardStereoFlatInterp(
            int unused0, int unused1,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int panLeft, int panRight,
            int unused2, int stopWrite, int endWrite,
            SoundDecoder self,
            int step, int wrapSample) {
        // Phase 1: cursor within real data.
        int framesInRange;
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = writePos + (endWrite - srcPosQ8 + step - 257) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        int outL   = writePos << 1;
        int outEnd = framesInRange << 1;

        while (outL < outEnd) {
            int idx  = srcPosQ8 >> 8;
            int s0   = src[idx];
            int samp = (s0 << 8) + (src[idx + 1] - s0) * (srcPosQ8 & 0xFF);
            mixBuf[outL++] += samp * panLeft  >> 6;
            mixBuf[outL++] += samp * panRight >> 6;
            srcPosQ8       += step;
        }

        // Phase 2: wrap boundary — use wrapSample.
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = (outL >> 1) + (endWrite - srcPosQ8 + step - 1) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }
        outEnd = framesInRange << 1;

        int ws = wrapSample;
        while (outL < outEnd) {
            int s0   = src[srcPosQ8 >> 8];
            int samp = (s0 << 8) + (ws - s0) * (srcPosQ8 & 0xFF);
            mixBuf[outL++] += samp * panLeft  >> 6;
            mixBuf[outL++] += samp * panRight >> 6;
            srcPosQ8       += step;
        }

        self.samplePos = srcPosQ8;
        return outL >> 1;
        // obf: sb.a(int, int, byte[], ..., sb, int, int) — clean L244
    }

    // -------------------------------------------------------------------------
    // STEREO GLIDE BACKWARD INTERPOLATED  (arbitrary pitch, ramped L/R, reverse)
    // -------------------------------------------------------------------------

    /**
     * Stereo backward, ramping L/R amplitude glide, arbitrary pitch with linear
     * interpolation.  Most general backward kernel.
     *
     * <p>Uses {@code src[idx-1]} and {@code src[idx]} for the backward lerp pair.
     *
     * obf: sb.b(int, int, byte[], int[], int, int, int, int, int, int, int, int, int, sb, int, int)  [16-param backward stereo glide interp]
     */
    private static final int mixBackwardStereoGlideInterp(
            int unused0, int unused1,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int panLeft, int panRight,
            int panLeftDelta, int panRightDelta,
            int unused2, int stopWrite, int endWrite,
            SoundDecoder self,
            int step, int wrapSample) {
        self.curAmp -= self.ampDelta * writePos;

        // Phase 1: within real data (cursor high enough for idx-1 to be valid).
        int framesInRange;
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = writePos + (endWrite + 256 - srcPosQ8 + step) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        int outL   = writePos << 1;
        int outEnd = framesInRange << 1;

        while (outL < outEnd) {
            int idx  = srcPosQ8 >> 8;
            int s0   = src[idx - 1];
            int samp = (s0 << 8) + (src[idx] - s0) * (srcPosQ8 & 0xFF);
            mixBuf[outL++] += samp * panLeft  >> 6;
            panLeft        += panLeftDelta;
            mixBuf[outL++] += samp * panRight >> 6;
            panRight       += panRightDelta;
            srcPosQ8       += step;   // step < 0 for backward
        }

        // Phase 2: wrap boundary — use wrapSample.
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = (outL >> 1) + (endWrite - srcPosQ8 + step) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }
        outEnd = framesInRange << 1;

        int ws = wrapSample;
        while (outL < outEnd) {
            int s0   = src[srcPosQ8 >> 8];
            int samp = (ws << 8) + (s0 - ws) * (srcPosQ8 & 0xFF);
            mixBuf[outL++] += samp * panLeft  >> 6;
            panLeft        += panLeftDelta;
            mixBuf[outL++] += samp * panRight >> 6;
            panRight       += panRightDelta;
            srcPosQ8       += step;
        }

        outL >>= 1;
        self.curAmp      += self.ampDelta      * outL;
        self.curPanLeft  = panLeft;
        self.curPanRight = panRight;
        self.samplePos   = srcPosQ8;
        return outL;
        // obf: sb.b(int, int, byte[], ..., sb, int, int) — 16-param backward stereo interp GLIDE (clean L718)
    }

    // -------------------------------------------------------------------------
    // STEREO FLAT BACKWARD INTERPOLATED  (arbitrary pitch, constant L/R, reverse)
    // -------------------------------------------------------------------------

    /**
     * Stereo backward, constant L/R pans, arbitrary pitch with linear interpolation.
     * Steady-state (no-glide) counterpart of {@link #mixBackwardStereoGlideInterp}:
     * pans are constant, no curAmp/pan writeback (only samplePos).
     * clean: sb.d(...14-param) at L464.
     *
     * obf: sb.d(int, int, byte[], int[], int, int, int, int, int, int, int, sb, int, int) — 14-param back stereo FLAT
     */
    private static final int mixBackwardStereoFlatInterp(
            int unused0, int unused1,
            byte[] src, int[] mixBuf,
            int srcPosQ8, int writePos,
            int panLeft, int panRight,
            int unused2, int stopWrite, int endWrite,
            SoundDecoder self,
            int step, int wrapSample) {
        // Phase 1: cursor within real data.
        int framesInRange;
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = writePos + (endWrite + 256 - srcPosQ8 + step) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }

        int outL   = writePos << 1;
        int outEnd = framesInRange << 1;

        while (outL < outEnd) {
            int idx  = srcPosQ8 >> 8;
            int s0   = src[idx - 1];
            int samp = (s0 << 8) + (src[idx] - s0) * (srcPosQ8 & 0xFF);
            mixBuf[outL++] += samp * panLeft  >> 6;
            mixBuf[outL++] += samp * panRight >> 6;
            srcPosQ8       += step;
        }

        // Phase 2: wrap boundary — use wrapSample.
        if (step == 0) {
            framesInRange = stopWrite;
        } else {
            framesInRange = (outL >> 1) + (endWrite - srcPosQ8 + step) / step;
            if (framesInRange > stopWrite) framesInRange = stopWrite;
        }
        outEnd = framesInRange << 1;

        int ws = wrapSample;
        while (outL < outEnd) {
            int s0   = src[srcPosQ8 >> 8];
            int samp = (ws << 8) + (s0 - ws) * (srcPosQ8 & 0xFF);
            mixBuf[outL++] += samp * panLeft  >> 6;
            mixBuf[outL++] += samp * panRight >> 6;
            srcPosQ8       += step;
        }

        self.samplePos = srcPosQ8;
        return outL >> 1;
        // obf: sb.d(int, int, byte[], ..., sb, int, int) — clean L464
    }

    // =========================================================================
    // Pan-law helpers
    // =========================================================================

    /**
     * Compute left-channel gain from a raw volume and pan position using a
     * square-root equal-power pan law.
     *
     * <pre>
     *   gain = vol * sqrt((16384 - pan) / 16384)
     *        = vol * sqrt((16384 - pan) * 1/2^13)
     *        ≈ vol * sqrt((16384 - pan) * 1.2207031e-4)
     * </pre>
     *
     * When {@code pan < 0}, returns {@code vol} unchanged (linear, no attenuation),
     * which is the mono path.
     *
     * @param vol raw amplitude
     * @param pan pan position [0..16384] (0=hard left, 8192=center, 16384=hard right)
     * @return left gain (same scale as vol)
     * obf: sb.b(int, int)   [2-param static, calls Math.sqrt]
     */
    private static final int computePanLeft(int vol, int pan) {
        if (pan < 0) {
            return vol;   // mono / unpanned: full amplitude
        }
        // Equal-power law: left = vol * sqrt((16384 - pan) / 16384)
        // 1/16384 = 1.2207031e-4  (1/2^13)
        return (int)((double) vol * Math.sqrt((double)(16384 - pan) * 1.2207031E-4f) + 0.5);
        // obf: sb.b(int, int) — static 2-param
    }

    /**
     * Compute right-channel gain from a raw volume and pan position using a
     * square-root equal-power pan law.
     *
     * <pre>
     *   gain = vol * sqrt(pan / 16384)
     *        ≈ vol * sqrt(pan * 1.2207031e-4)
     * </pre>
     *
     * When {@code pan < 0}, returns {@code -vol} (inverted, used for phase-flip).
     *
     * @param vol raw amplitude
     * @param pan pan position [0..16384]
     * @return right gain (same scale as vol)
     * obf: sb.c(int, int)   [2-param static, calls Math.sqrt]
     */
    private static final int computePanRight(int vol, int pan) {
        if (pan < 0) {
            return -vol;   // phase flip for mono hard-right / special case
        }
        return (int)((double) vol * Math.sqrt((double) pan * 1.2207031E-4f) + 0.5);
        // obf: sb.c(int, int) — static 2-param
    }
}
