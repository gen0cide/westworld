package client.audio;

import java.awt.Component;
import client.shell.LoaderThread;
import client.util.ArrayUtil;
import client.util.Timer;
import client.util.Utility;

/**
 * AudioChannel — abstract base class for an audio output voice/channel in the RSC client's
 * software audio engine (rev ~233-235, Microsoft J++ build).
 *
 * Each concrete AudioChannel owns:
 *   - a mix buffer ({@code sampleBuffer}, int[256] in mono or int[512] in stereo)
 *   - a {@link FilterChain} root node ({@code activeFilterChain}) for the currently-playing sound
 *   - two parallel arrays ({@code bucketHead[]} / {@code bucketTail[]}) that form 8 priority
 *     buckets used to schedule child {@link FilterChain} voices
 *   - timing state tracking the audio hardware cursor vs. what has been written so far
 *
 * The concrete subclass {@link SourceLinePlayer} (obf: {@code pb}) drives a Java Sound
 * {@code SourceDataLine}.  A fallback no-op instance is returned on error.
 *
 * The static factory {@link #create} is called by {@link AudioMixer} (obf: {@code eb}).
 * The static configurator {@link #configure} must be called once before any channel is created;
 * it sets the sample rate, stereo flag, and optional mixer-thread priority.
 *
 * Threading: the mixer thread (AudioMixer.run) calls {@link #tick()} at ~100 Hz; all
 * externally visible mutating methods are {@code synchronized}.
 *
 * Obfuscated class name: {@code sa}
 */
public class AudioChannel {

    // -------------------------------------------------------------------------
    // Static configuration (shared across all channels)
    // -------------------------------------------------------------------------

    /**
     * Sample rate in Hz (e.g. 11025, 22050, 44100).
     * Must be set via {@link #configure} before any channel is created.
     * obf: sa.t
     */
    public static int sampleRate;                          // obf: t

    /**
     * {@code true} if stereo output is requested (doubles all buffer sizes).
     * obf: sa.i
     */
    public static boolean stereo;                          // obf: i

    /**
     * Mixer-thread priority passed to {@link LoaderThread#startThread}.
     * 0 means: do not create a dedicated mixer thread (caller drives mixing).
     * obf: sa.o
     */
    private static int mixerThreadPriority;         // obf: o

    /**
     * The singleton {@link AudioMixer} (obf: {@code eb}) instance.
     * Non-null only when {@code mixerThreadPriority > 0} and at least one channel exists.
     * Protected by the channel's monitor when reading/writing {@code n.f[]}.
     * obf: sa.n
     */
    private static AudioMixer mixer;                // obf: n

    // -------------------------------------------------------------------------
    // Instance state
    // -------------------------------------------------------------------------

    /**
     * The mix buffer: 256 ints (mono) or 512 ints (stereo), filled each tick.
     * Exposed package-private so {@link SourceLinePlayer} can read it.
     * obf: sa.j
     */
    public int[] sampleBuffer;                             // obf: j

    /**
     * Whether this channel has been closed/released.
     * When {@code true} {@link #tick()} returns immediately.
     * obf: sa.b (field)
     */
    private boolean closed = false;                 // obf: b

    /**
     * Soft-cap on the total "mix depth" (sum of FilterChain.d() contributions) accumulated in
     * one fill pass before overflow is declared and mixing is aborted early.
     * Default 32; corresponds to the maximum simultaneous voice count the engine tolerates.
     * obf: sa.r
     */
    private int maxMixDepth = 32;                   // obf: r

    /**
     * The {@link FilterChain} (obf: {@code va}) that is currently the root of the active
     * sound graph.  Written by {@link #setFilterChain(FilterChain)}.
     * obf: sa.a (field)
     */
    private FilterChain activeFilterChain;          // obf: a

    /**
     * The "write-position offset" relative to {@code bufferSizeFrames}: how many frames ahead
     * of the hardware cursor we are trying to stay.
     * obf: sa.g
     */
    private int writeAheadOffset;                   // obf: g

    /**
     * Requested hardware buffer size in frames (samples per channel), passed to
     * {@link #openLine(int)} on construction and re-opened on resize.
     * Rounded up to the nearest 1024-frame boundary, capped at 16384.
     * obf: sa.q
     */
    public int requestedBufferFrames;              // obf: q

    /**
     * The max observed gap measured during the previous 2-second window, retained so that
     * {@code writeAheadOffset} can be smoothed: {@code writeAheadOffset = min(prevMaxGap, maxObservedGap)}.
     * Also used alongside {@code maxObservedGap} to detect true silence
     * (both zero → schedule quiet window).
     * obf: sa.c (field)
     */
    private int prevMaxGap;                         // obf: c

    /**
     * The actual hardware buffer size in frames as reported/negotiated with the audio line.
     * May be larger than {@code requestedBufferFrames}; capped at 16384.
     * obf: sa.k
     */
    public int bufferSizeFrames;                   // obf: k

    /**
     * Per-priority-bucket head pointers (8 buckets, indices 0-7).
     * Together with {@code bucketTail[]} they form a singly-linked list of {@link FilterChain}
     * nodes per scheduling bucket.  {@code bucketHead[b]} is the first node to process in
     * bucket {@code b}; {@code bucketTail[b]} is the last (used when appending).
     * obf: sa.s (field)
     */
    private FilterChain[] bucketTail = new FilterChain[8];  // obf: s

    /**
     * Timestamp (ms, from {@link Timer#getTime}) at which mixing should resume after a
     * voluntary silence window.  0 means "no pending silence".
     * obf: sa.e (field)
     */
    private long silenceUntilTime = 0L;             // obf: e

    /**
     * Smoothed "pressure" counter: how many frames worth of audio data are buffered in
     * the hardware line beyond what is playing right now.  Decremented by each fill; used
     * to pace writes and avoid overrun.
     * obf: sa.p (field)
     */
    private int bufferPressure = 0;                 // obf: p

    /**
     * Maximum observed gap between the hardware play cursor and the write cursor, over the
     * last 2-second measurement window.  Used to adapt write-ahead.
     * obf: sa.m (field)
     */
    private int maxObservedGap = 0;                 // obf: m

    /**
     * Per-priority-bucket tail pointers (8 buckets, indices 0-7).
     * {@code bucketHead[b]} is the first FilterChain in bucket {@code b}.
     * obf: sa.d (field)
     */
    private FilterChain[] bucketHead = new FilterChain[8];  // obf: d

    /**
     * Timestamp of the next periodic "gap measurement" window update (every ~2 s).
     * obf: sa.f (field)
     */
    private long nextGapUpdateTime = 0L;            // obf: f

    /**
     * Hardware write-cursor position (in frames) as of the last {@link #tick} call.
     * obf: sa.h (field)
     */
    private int lastWritePosition = 0;              // obf: h

    /**
     * {@code true} after a line resize or initial open; clears after the first tick in which
     * mixing actually produced frames, suppressing the gap-update logic for one cycle.
     * obf: sa.l (field)
     */
    private boolean justReopened = true;            // obf: l

    /**
     * Wall-clock timestamp of the last time we actually wrote audio data, used to detect
     * stalls and schedule silence windows.
     * obf: sa.u
     */
    private long lastWriteTime;                     // obf: u

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /**
     * Default constructor; initialises timing via {@link Timer#getTime}.
     * Normally only called indirectly through {@link #create}.
     */
    public AudioChannel() {
        this.lastWriteTime = Timer.getTime(0);
        // bucketTail, bucketHead already allocated as new FilterChain[8] above
        // all other numeric fields default-init to 0/false/true as documented
    }

    // -------------------------------------------------------------------------
    // Static API
    // -------------------------------------------------------------------------

    /**
     * One-time audio subsystem configuration.  Must be called before {@link #create}.
     *
     * @param rate     sample rate in Hz; must be in [8000, 48000]
     * @param isStereo {@code true} for stereo, {@code false} for mono
     * @param threadPriority mixer-thread priority (0 = caller-driven, no thread spawned)
     *
     * obf: sa.a(int, boolean, int)  [static]
     */
    public static final void configure(int rate, boolean isStereo, int threadPriority) {
        // Validate sample rate range
        if (rate < 8000 || rate > 48000) {
            throw new IllegalArgumentException();
        }
        sampleRate = rate;
        stereo = isStereo;
        mixerThreadPriority = threadPriority;
    }

    /**
     * Factory: create and register a new AudioChannel for the given slot index.
     *
     * <p>Instantiates a {@link SourceLinePlayer} (obf: {@code pb}), sizes its sample buffer,
     * opens the audio line, and registers it with the singleton {@link AudioMixer}.  If any
     * step fails (e.g. line unavailable) a silent no-op {@code AudioChannel} is returned
     * instead so the client degrades gracefully.
     *
     * @param loaderThread  the {@link LoaderThread} (obf: {@code c}) used to spawn the mixer
     *                      thread if one does not yet exist
     * @param component     AWT component used by {@link SourceLinePlayer} to find the best
     *                      output mixer (passed to {@link #initAudioFormat(Component)})
     * @param slotIndex     which slot in {@code AudioMixer.voices[]} to assign (0 or 1)
     * @param bufferFrames  requested hardware buffer size in frames; clamped to [256, 16384]
     * @return a fully initialised AudioChannel, or a silent stub on failure
     *
     * obf: sa.a(c, Component, int, int)  [static]
     */
    public static final AudioChannel create(LoaderThread loaderThread, Component component,
                                     int slotIndex, int bufferFrames) {
        // Guard: configure() must have been called first
        if (sampleRate == 0) {
            throw new IllegalStateException();
        }
        // Guard: slotIndex must be a valid mixer slot (0 or 1)
        if (slotIndex < 0 || slotIndex >= 2) {
            throw new IllegalArgumentException();
        }
        // Clamp minimum buffer size
        if (bufferFrames < 256) {
            bufferFrames = 256;
        }
        try {
            SourceLinePlayer player = new SourceLinePlayer();
            // Allocate the mix buffer: 256 ints mono, 512 ints stereo
            player.sampleBuffer = new int[256 * (stereo ? 2 : 1)];
            player.requestedBufferFrames = bufferFrames;
            // Let the subclass initialise its AudioFormat and find the best mixer device
            player.initAudioFormat(component);
            // Round up to the nearest 1024-frame boundary
            player.bufferSizeFrames = (bufferFrames & ~0x3FF) + 1024;
            if (player.bufferSizeFrames > 16384) {
                player.bufferSizeFrames = 16384;
            }
            // Open the hardware line
            player.openLine(player.bufferSizeFrames);

            // Lazily create the mixer thread if needed
            if (mixerThreadPriority > 0 && mixer == null) {
                mixer = new AudioMixer();
                mixer.loaderThread = loaderThread;      // obf: n.g = var0
                // loaderThread.startThread(true, mixer, mixerThreadPriority) — starts run()
                loaderThread.startThread(true, mixer, mixerThreadPriority); // obf: var0.a(true, n, o)
            }

            // Register in the mixer's slot array
            if (mixer != null) {
                if (mixer.voices[slotIndex] != null) {  // obf: n.f[var2]
                    throw new IllegalArgumentException(); // slot already occupied
                }
                mixer.voices[slotIndex] = player;       // obf: n.f[var2] = var4
            }
            return player;
        } catch (Throwable t) {
            // Degrade gracefully: return a silent stub channel
            return new AudioChannel();
        }
    }

    // -------------------------------------------------------------------------
    // Instance API — called from AudioMixer.run() and subclasses
    // -------------------------------------------------------------------------

    /**
     * Main periodic tick: mix pending audio data into the hardware line.
     *
     * <p>Called by {@link AudioMixer#run} (or the owning thread) roughly every 10 ms.
     * Responsibilities:
     * <ol>
     *   <li>Catch up if the wall clock has drifted more than 6 s ahead of our last-write
     *       timestamp (prevents spiral of death on hiccup).
     *   <li>Fill any gaps between the current write position and the hardware play cursor by
     *       calling {@link #fillBuffer(int[], int)} repeatedly in 256-frame chunks.
     *   <li>Detect silence and schedule a voluntary quiet window (2 s) to save CPU.
     *   <li>Grow the hardware buffer (by 1024 frames) if the write cursor is getting too close
     *       to the play cursor.
     * </ol>
     *
     * obf: sa.a()  [instance, synchronized]
     */
    public final synchronized void tick() {
        if (closed) {
            return;
        }
        long now = Timer.getTime(0);

        // --- Clock drift correction ---
        // If now > lastWriteTime + 6000 ms, pretend only 6 s has elapsed to avoid
        // a massive backlog of fills after a pause (GC, suspend, etc.).
        try {
            if (now > lastWriteTime + 6000L) {
                lastWriteTime = now - 6000L;
            }
            // If we are more than 5 s behind: fast-forward by filling silence
            while (now > lastWriteTime + 5000L) {
                // fillBuffer with silence (256 frames, no active chain)
                fillSilenceFrames(256);
                // Each 256-frame block represents 256000/sampleRate ms of audio
                lastWriteTime += (long)(256000 / sampleRate);
                now = Timer.getTime(0);
            }
        } catch (Exception e) {
            // Clock error: just snap lastWriteTime to now and continue
            lastWriteTime = now;
        }

        // sampleBuffer null → closed/stub channel, nothing to do
        if (sampleBuffer == null) {
            return;
        }

        try {
            // --- Handle deferred silence window ---
            if (silenceUntilTime != 0L) {
                if (now < silenceUntilTime) {
                    return; // still in the quiet window
                }
                // Silence window expired: reopen the line at the current buffer size
                openLine(bufferSizeFrames);
                silenceUntilTime = 0L;
                justReopened = true;
            }

            // --- Compute write position relative to play cursor ---
            // getWritePosition() returns the subclass's view of how many frames have been
            // committed to the hardware line so far.
            int writePos = getWritePosition();

            // Track maximum observed gap between play cursor and our last write position
            if (lastWritePosition - writePos > maxObservedGap) {
                maxObservedGap = lastWritePosition - writePos;
            }

            // Compute target write position: requestedBufferFrames + writeAheadOffset,
            // capped so that it does not wrap past 16384.
            int targetWritePos = requestedBufferFrames + writeAheadOffset;
            if (targetWritePos + 256 > 16384) {
                targetWritePos = 16128; // 16384 - 256
            }

            // --- Grow hardware buffer if write head is too close to play cursor ---
            if (targetWritePos + 256 > bufferSizeFrames) {
                bufferSizeFrames += 1024;
                if (bufferSizeFrames > 16384) {
                    bufferSizeFrames = 16384;
                }
                // Close and reopen the line at the new larger size
                closeAndFlush();
                openLine(bufferSizeFrames);
                writePos = 0;
                justReopened = true;
                // If still not enough room, clamp targetWritePos to fit
                if (targetWritePos + 256 > bufferSizeFrames) {
                    targetWritePos = bufferSizeFrames - 256;
                    writeAheadOffset = targetWritePos - requestedBufferFrames;
                }
            }

            // --- Fill audio data up to targetWritePos ---
            // Each iteration fills one 256-frame block.
            while (writePos < targetWritePos) {
                fillBuffer(sampleBuffer, 256);
                commitBuffer(); // subclass writes sampleBuffer to the hardware line
                writePos += 256;
            }

            // --- Adaptive gap measurement (every ~2 s) ---
            if (now > nextGapUpdateTime) {
                if (!justReopened) {
                    if (maxObservedGap == 0 && prevMaxGap == 0) {
                        // True silence: schedule a voluntary quiet window
                        closeAndFlush();
                        silenceUntilTime = now + 2000L;
                        return;
                    }
                    // Adapt write-ahead offset: smooth toward min of previous and current window gap
                    writeAheadOffset = Math.min(prevMaxGap, maxObservedGap);
                    prevMaxGap = maxObservedGap;
                } else {
                    justReopened = false;
                }
                maxObservedGap = 0;
                nextGapUpdateTime = now + 2000L;
            }
            lastWritePosition = writePos;

        } catch (Exception e) {
            // On any error: enter a quiet window for 2 s then retry
            closeAndFlush();
            silenceUntilTime = now + 2000L;
        }
    }

    /**
     * Replace the root {@link FilterChain} for this channel.
     * Called from the client game loop to start a new sound.
     *
     * obf: sa.a(va)  [instance, synchronized]
     */
    public final synchronized void setFilterChain(FilterChain chain) {
        this.activeFilterChain = chain;             // obf: this.a = var1
    }

    /**
     * Close and release this channel, unregistering from the mixer.
     *
     * <ul>
     *   <li>Removes this channel from {@link AudioMixer#channels} (if present).
     *   <li>If the mixer no longer has any live channels, signals the mixer thread to stop
     *       and busy-waits (up to ~10 ms per iteration) for it to exit.
     *   <li>Calls {@link #closeAndFlush()} to release the hardware line.
     *   <li>Nulls {@code sampleBuffer} and sets {@code closed = true}.
     * </ul>
     *
     * obf: sa.d()  [instance, synchronized]
     */
    public final synchronized void release() {
        if (mixer != null) {
            boolean allSlotsEmpty = true;
            for (int slotIndex = 0; slotIndex < 2; slotIndex++) {
                if (mixer.voices[slotIndex] == this) {  // obf: n.f[var2]
                    mixer.voices[slotIndex] = null;
                }
                if (mixer.voices[slotIndex] != null) {
                    allSlotsEmpty = false;
                }
            }
            if (allSlotsEmpty) {
                // Signal the mixer thread to stop
                mixer.stopRequested = true;             // obf: n.a = true
                // Busy-wait for the mixer thread to acknowledge exit (i.e. mixer.isRunning → false)
                // mb.a(11200, 50L) → Utility.sleepWithProfile(11200 [opcode], 50L [ms])
                while (mixer.isRunning) {                // obf: n.i
                    Utility.sleepWithProfile(11200, 50L);
                }
                mixer = null;
            }
        }
        closeAndFlush();
        sampleBuffer = null;
        closed = true;
    }

    // -------------------------------------------------------------------------
    // Abstract / overridable hooks (implemented by SourceLinePlayer)
    // -------------------------------------------------------------------------

    /**
     * Return the current hardware write-cursor position in frames.
     * The base implementation returns {@code bufferSizeFrames} (stub/no-op).
     *
     * obf: sa.b() [instance, throws Exception]
     */
    public int getWritePosition() throws Exception {
        return this.bufferSizeFrames;              // obf: return this.k
    }

    /**
     * Open (or reopen) the hardware audio line with the given buffer size in frames.
     * No-op in the base class; overridden by {@link SourceLinePlayer}.
     *
     * @param bufSizeFrames hardware ring-buffer size in frames
     * obf: sa.b(int) [instance, throws Exception]
     */
    public void openLine(int bufSizeFrames) throws Exception {
    }

    /**
     * Commit the current contents of {@code sampleBuffer} to the hardware line.
     * No-op in the base class; overridden by {@link SourceLinePlayer}.
     *
     * obf: sa.c() [instance, throws Exception]
     */
    public void commitBuffer() throws Exception {
    }

    /**
     * Close the hardware line and release its resources without fully releasing this channel.
     * Called before reopening with a different buffer size or entering a silence window.
     * No-op in the base class; overridden by {@link SourceLinePlayer}.
     *
     * obf: sa.e() [instance]
     */
    public void closeAndFlush() {
    }

    /**
     * Initialise the audio format for the hardware line (called once during {@link #create}).
     * No-op in the base class; overridden by {@link SourceLinePlayer}.
     *
     * @param component AWT component used to locate the preferred output device
     * obf: sa.a(Component) [instance, throws Exception]
     */
    public void initAudioFormat(Component component) throws Exception {
    }

    // -------------------------------------------------------------------------
    // Private mixing helpers
    // -------------------------------------------------------------------------

    /**
     * Fill {@code count} frames of silence into the hardware line, advancing the timing state
     * but not actually writing real audio.  Used during clock drift catch-up in {@link #tick}.
     *
     * Decrements {@code bufferPressure} and notifies the active {@link FilterChain} (if any)
     * of the skipped frames via {@link FilterChain#skipFrames(int)}.
     *
     * obf: sa.a(int) [private]
     */
    private final void fillSilenceFrames(int frameCount) {
        bufferPressure -= frameCount;
        if (bufferPressure < 0) {
            bufferPressure = 0;
        }
        if (activeFilterChain != null) {
            activeFilterChain.skipFrames(frameCount);   // obf: this.a.b(int)
        }
    }

    /**
     * Core audio fill: zero the output array, then walk the 8-bucket priority queue,
     * mixing each active {@link FilterChain} into {@code out[]}.
     *
     * <p>The scheduler uses a 256-bit active-bucket bitmask ({@code activeBucketMask}) and a
     * base priority {@code priorityLevel} (0-7) to iterate over all buckets from highest to
     * lowest priority.  Within each bucket the chain's {@link FilterNode} (obf: {@code bb})
     * determines whether this bucket's chain is "expensive" (i.e. whether its computed cost
     * exceeds the current pressure threshold); if so the chain is skipped for this tick and
     * re-inserted with the pressure bit set.
     *
     * <p>After the pass, all remaining chain nodes in each bucket are unlinked and dropped
     * (they represent voices that have finished or been evicted).
     *
     * @param out        the mix buffer to fill (length = {@code frameCount * (stereo?2:1)})
     * @param frameCount number of PCM frames to produce (always 256 in practice)
     *
     * obf: sa.a(int[], int) [private]
     */
    private final void fillBuffer(int[] out, int frameCount) {
        // Determine the actual number of int slots to zero
        int slotCount = frameCount;
        if (stereo) {
            slotCount = frameCount << 1;   // stereo: 512 ints
        }

        // Zero the mix buffer
        ArrayUtil.fill(out, 0, slotCount);           // obf: ab.a(var1, 0, var3)

        bufferPressure -= frameCount;

        // If we have an active chain and pressure has gone non-positive: time to mix
        if (activeFilterChain != null && bufferPressure <= 0) {
            // Restore some pressure headroom: sampleRate/16 frames worth
            bufferPressure += sampleRate >> 4;       // obf: sa.t >> 4

            // Reset the "visited" flag on the root chain and all its children
            resetVisitedFlags(activeFilterChain);    // obf: sa.b(this.a)

            // Insert the root chain into its priority bucket
            insertIntoBucket(activeFilterChain, activeFilterChain.getDefaultPriority()); // obf: this.a(this.a, this.a.c())

            // --- Priority scheduler ---
            // activeBucketMask: one bit per slot within a 32-bit group, 8 groups total.
            // mixDepthAccum accumulates d() contributions; when >= maxMixDepth, abort.
            int mixDepthAccum = 0;
            int activeBucketMask = 255;         // bits 0-7 set = all 8 buckets active
            int priorityLevel = 7;              // scan from highest (7) to lowest (0)

            outerLoop:
            while (true) {
                // Determine shift and pressure threshold for this priority level
                int bitShift;
                int pressureThreshold;
                if (activeBucketMask == 0) {
                    break; // no more active buckets
                }
                if (priorityLevel >= 0) {
                    bitShift = priorityLevel;
                    pressureThreshold = 0;
                } else {
                    // priorityLevel < 0 encodes (shift, pressureThreshold) via:
                    //   bitShift = priorityLevel & 3
                    //   pressureThreshold = -(priorityLevel >> 2)
                    bitShift = priorityLevel & 3;
                    pressureThreshold = -(priorityLevel >> 2);
                }

                // Collect all active bits in this group of 4 buckets
                // 0x11111111 = interleaved nibble mask (every 4th bit)
                int nibbleMask = activeBucketMask >>> bitShift & 0x11111111; // obf: 286331153

                while (true) {
                    if (nibbleMask == 0) {
                        break;
                    }
                    if ((nibbleMask & 1) != 0) {
                        // This bucket slot has active chains — process them
                        activeBucketMask &= ~(1 << bitShift);

                        FilterChain prev = null;
                        FilterChain node = bucketHead[bitShift];

                        chainWalk:
                        while (node != null) {
                            FilterNode filterNode = node.filterNode;   // obf: node.h

                            if (filterNode != null && filterNode.cost > pressureThreshold) {
                                // Chain is "expensive" for this pressure level — keep it
                                activeBucketMask |= (1 << bitShift);
                                prev = node;
                                node = node.nextInBucket;               // obf: node.j
                                continue;
                            }

                            // Mark chain as visited/active
                            node.active = true;                         // obf: node.g = true
                            int depthContrib = node.getMixDepth();      // obf: node.d()
                            mixDepthAccum += depthContrib;
                            if (filterNode != null) {
                                filterNode.cost += depthContrib;        // obf: filterNode.g += depthContrib
                            }

                            if (mixDepthAccum >= maxMixDepth) {
                                break outerLoop; // too many simultaneous voices
                            }

                            // Mix this chain's children into the output
                            FilterChain firstChild = node.firstChild(); // obf: node.b()
                            if (firstChild != null) {
                                int parentPriority = node.priority;     // obf: node.i
                                FilterChain child = firstChild;
                                while (child != null) {
                                    // Priority of child = parentPriority * child.getDefaultPriority() >> 8
                                    insertIntoBucket(child, parentPriority * child.getDefaultPriority() >> 8);
                                    child = node.nextChild();           // obf: node.a()
                                }
                            }

                            // Unlink this node from the bucket
                            FilterChain nextNode = node.nextInBucket;   // obf: node.j
                            node.nextInBucket = null;
                            if (prev == null) {
                                bucketHead[bitShift] = nextNode;
                            } else {
                                prev.nextInBucket = nextNode;
                            }
                            if (nextNode == null) {
                                bucketTail[bitShift] = prev;
                            }
                            node = nextNode;
                        }
                    }

                    // Advance to the next nibble position (4 buckets apart)
                    bitShift += 4;
                    pressureThreshold++;
                    nibbleMask >>>= 4;
                }

                priorityLevel--;
            }

            // Clean up: unlink all remaining chain nodes in all 8 buckets
            for (int b = 0; b < 8; b++) {
                FilterChain node = bucketHead[b];
                bucketTail[b] = null;
                bucketHead[b] = null;
                while (node != null) {
                    FilterChain next = node.nextInBucket;
                    node.nextInBucket = null;
                    node = next;
                }
            }
        }

        // Clamp bufferPressure to non-negative
        if (bufferPressure < 0) {
            bufferPressure = 0;
        }

        // Let the active chain mix into 'out' (if present)
        if (activeFilterChain != null) {
            activeFilterChain.mixInto(out, 0, frameCount); // obf: this.a.b(int[], 0, int)
        }

        // Snapshot wall-clock time of this write
        lastWriteTime = Timer.getTime(0);             // obf: p.a(0)
    }

    /**
     * Insert a {@link FilterChain} node into the appropriate priority bucket.
     *
     * <p>The bucket index is {@code priority >> 5} (i.e. bits [12:5] select the slot in [0,7]).
     * If the bucket is empty the node becomes both head and tail; otherwise it is appended to
     * the tail.
     *
     * @param chain    the FilterChain to schedule
     * @param priority the scheduling priority (lower = higher priority; 0 = highest)
     *
     * obf: sa.a(va, int) [private]
     */
    private final void insertIntoBucket(FilterChain chain, int priority) {
        int bucketIndex = priority >> 5;            // obf: var3 = var2 >> 5
        FilterChain currentTail = bucketTail[bucketIndex]; // obf: this.s[var3]

        if (currentTail == null) {
            // Bucket was empty: chain becomes the head
            bucketHead[bucketIndex] = chain;        // obf: this.d[var3] = var1
        } else {
            // Append to current tail
            currentTail.nextInBucket = chain;       // obf: var4.j = var1
        }
        bucketTail[bucketIndex] = chain;            // obf: this.s[var3] = var1
        chain.priority = priority;                  // obf: var1.i = var2
    }

    /**
     * Recursively clear the {@code active} flag and reset the {@link FilterNode} cost counter
     * on a chain and all its children.  Called at the start of each fill to prepare the
     * priority scheduler.
     *
     * @param chain root of the chain graph to reset
     * obf: sa.b(va) [private static]
     */
    private static final void resetVisitedFlags(FilterChain chain) {
        chain.active = false;                       // obf: var0.g = false
        if (chain.filterNode != null) {
            chain.filterNode.cost = 0;              // obf: var0.h.g = 0
        }
        // Recurse into children
        FilterChain child = chain.firstChild();     // obf: var0.b()
        while (child != null) {
            resetVisitedFlags(child);
            child = chain.nextChild();              // obf: var0.a()
        }
    }
}
