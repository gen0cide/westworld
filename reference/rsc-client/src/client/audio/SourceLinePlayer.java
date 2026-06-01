package client.audio;

import java.awt.Component;
import javax.sound.sampled.AudioFormat;
import javax.sound.sampled.AudioSystem;
import javax.sound.sampled.DataLine;
import javax.sound.sampled.LineUnavailableException;
import javax.sound.sampled.Mixer;
import javax.sound.sampled.SourceDataLine;
import client.data.BZip;
import client.world.GameCharacter;

/**
 * SourceLinePlayer — Java Sound (javax.sound.sampled) audio output backend.
 *
 * <p>Extends {@link AudioChannel} (obf: {@code sa}) to implement the actual
 * PCM write path using a {@link SourceDataLine}.  The {@link AudioMixer}
 * (obf: {@code eb}) calls {@link #writeSamples()} on this object once per
 * mixer tick; the object converts the mixer's 24-bit signed integer sample
 * buffer ({@code AudioChannel.sampleBuffer}) into 16-bit little-endian PCM bytes
 * and pushes them to the OS audio queue via {@code SourceDataLine.write()}.
 *
 * <p>At initialisation ({@link #initOutput(Component)}) the class scans all
 * available {@link Mixer.Info} entries and silently skips any whose name
 * contains {@code "soundmax"} — a compatibility workaround for the buggy
 * SoundMAX AC'97 driver shipping on many Pentium-4 era motherboards that
 * deadlocked or produced silence under Java Sound.  The rest of the init
 * allocates an {@link AudioFormat} that matches the engine's configured sample
 * rate ({@code AudioChannel.sampleRate}) and channel count, then pre-allocates
 * the byte write-buffer.
 *
 * <p>Buffer sizing uses power-of-two arithmetic: {@link #openLine(int)}
 * falls back to the next higher power of two if the OS refuses to open a line
 * at the exact requested buffer size.
 *
 * Obfuscated class name: {@code pb}
 * Package: {@code client.audio}
 */
public final class SourceLinePlayer extends AudioChannel {

    // -------------------------------------------------------------------------
    // Fields
    // -------------------------------------------------------------------------

    /**
     * The number of sample frames (not bytes) that fit in the open
     * {@link SourceDataLine}'s OS buffer.  Used by {@link #getBufferedFrames()}
     * to report how many frames of audio are still queued for playback.
     *
     * obf: {@code x}
     */
    private int lineBufferFrames;

    /**
     * Scratch byte array used when writing PCM data to the
     * {@link SourceDataLine}.  Pre-allocated in {@link #initOutput(Component)}
     * to avoid GC pressure in the hot write path.
     *
     * Size: 256 frames * 2 bytes/sample * (1 or 2 channels) =
     *       512 bytes (mono) or 1024 bytes (stereo).
     *
     * obf: {@code v}
     */
    private byte[] pcmWriteBuffer;

    /**
     * The {@link AudioFormat} negotiated during {@link #initOutput(Component)}.
     * 16-bit, signed, little-endian PCM at the engine's configured sample rate.
     *
     * obf: {@code y}
     */
    private AudioFormat audioFormat;

    /**
     * The active {@link SourceDataLine} through which PCM is streamed to the
     * OS mixer.  Null when no line is open (before {@link #openLine} succeeds
     * or after {@link #closeLine()} has been called).
     *
     * obf: {@code w}
     */
    private SourceDataLine sourceDataLine;

    /**
     * Decoded XOR string pool constant: {@code "soundmax"}.
     *
     * Used in {@link #initOutput(Component)} to skip SoundMAX AC'97 mixer
     * entries that are known to be problematic under Java Sound on the
     * Microsoft JVM.
     *
     * obf field: {@code z}  (XOR-encoded literal {@code "}Kqc"}
     * decoded by the two {@code z()} helpers with the mod-5 key table
     * {@code [14, 117, 106, 37, 21]})
     */
    private static final String SOUNDMAX_MIXER_FILTER = "soundmax"; // obf: z

    // -------------------------------------------------------------------------
    // AudioChannel overrides
    // -------------------------------------------------------------------------

    /**
     * Closes the {@link SourceDataLine} and clears the reference.
     * Called by {@link AudioChannel#detach()} when this voice is removed
     * from the {@link AudioMixer}.
     *
     * obf: {@code e()}
     */
    @Override
    public final void closeLine() { // obf: e
        if (sourceDataLine != null) {
            sourceDataLine.close();
            sourceDataLine = null;
        }
    }

    /**
     * Converts the 24-bit integer samples in {@link AudioChannel#sampleBuffer}
     * to 16-bit signed little-endian PCM and writes one 256-frame block to
     * the {@link SourceDataLine}.
     *
     * <p>The mixer accumulates samples as 24-bit signed integers (range
     * {@code -8388608} to {@code +8388607}).  This method clamps them to that
     * range, then extracts the upper 16 bits as a signed 16-bit value, storing
     * it in little-endian order (low byte first, then high byte) in
     * {@link #pcmWriteBuffer}.
     *
     * <p>Clamping uses the standard branchless trick:
     * <pre>
     *   if ((sample + 0x800000 &amp; 0xFF000000) != 0)    // overflowed?
     *       sample = 0x7FFFFF ^ (sample &gt;&gt; 31);       // clamp to ±max
     * </pre>
     * {@code sample >> 31} is {@code -1} (0xFFFFFFFF) for negatives and
     * {@code 0} for positives, so {@code 0x7FFFFF ^ -1 = -8388608} (min) and
     * {@code 0x7FFFFF ^ 0 = 8388607} (max).
     *
     * <p>The frame count is 256 in mono or 512 in stereo ({@code stereo} flag
     * from {@link AudioChannel#stereo}).
     *
     * obf: {@code c()}
     */
    @Override
    public final void writeSamples() { // obf: c
        // Number of sample frames (= samples in mono, sample-pairs in stereo).
        // In stereo mode, sampleBuffer holds interleaved L/R pairs so we double
        // the frame count to cover both channels.
        int frameCount = 256;
        if (stereo) { // obf: i (inherited static from AudioChannel)
            frameCount <<= 1; // 512 frames (256 L + 256 R)
        }

        for (int frame = 0; frame < frameCount; frame++) {
            // Raw 24-bit signed sample from the mixer accumulation buffer.
            int sample = sampleBuffer[frame]; // obf: this.j[var2]

            // Clamp to 24-bit signed range [-8388608, 8388607].
            // Trick: adding 0x800000 pushes the value so that any out-of-range
            // value will overflow into bit 24+, which is detected by the mask.
            if ((sample + 0x800000 & 0xFF000000) != 0) {
                // sample >> 31 == -1 for negative, 0 for positive overflow.
                sample = 0x7FFFFF ^ (sample >> 31);
            }

            // Extract the high 16 bits of the 24-bit value as a signed 16-bit
            // PCM sample, stored little-endian:
            //   byte[0] = bits 8–15  (low byte of the 16-bit word)
            //   byte[1] = bits 16–23 (high byte of the 16-bit word)
            pcmWriteBuffer[frame * 2]     = (byte)(sample >> 8);  // low byte
            pcmWriteBuffer[frame * 2 + 1] = (byte)(sample >> 16); // high byte
        }

        // Push the encoded block to the OS audio queue.
        // frameCount << 1 = total bytes (2 bytes per sample frame).
        sourceDataLine.write(pcmWriteBuffer, 0, frameCount << 1);
    }

    /**
     * Returns the number of sample frames currently buffered (queued for
     * playback but not yet consumed by the OS).
     *
     * <p>Formula: {@code lineBufferFrames - available_frames}, where
     * {@code available()} returns the number of bytes still free in the OS
     * buffer.  Dividing by 2 (mono) or 4 (stereo) converts bytes to frames.
     *
     * obf: {@code b()} (no-arg overload)
     */
    @Override
    public final int getBufferedFrames() { // obf: b
        // sourceDataLine.available() returns free bytes in the OS ring buffer.
        // Shift right by 1 (mono: 2 bytes/frame) or 2 (stereo: 4 bytes/frame)
        // converts bytes → frames.
        return lineBufferFrames - (sourceDataLine.available() >> (stereo ? 2 : 1));
    }

    /**
     * Opens the {@link SourceDataLine} with the given OS buffer size (in sample
     * frames).  If the OS rejects the exact buffer size, the method retries
     * with the next power of two above the requested size.  If even that
     * minimum (determined by {@link GameCharacter#countBits}) fails, the line
     * reference is cleared and the exception is re-thrown.
     *
     * <p>Buffer size in bytes = {@code bufferFrames << 1} (mono) or
     * {@code bufferFrames << 2} (stereo), because each frame is 2 bytes in
     * mono or 4 bytes in stereo (16-bit PCM).
     *
     * @param bufferFrames  requested OS ring-buffer size in sample frames
     * @throws LineUnavailableException if no suitable audio line can be opened
     *
     * obf: {@code b(int)} (int-arg overload)
     */
    @Override
    public final void openLine(int bufferFrames) throws LineUnavailableException { // obf: b(int)
        try {
            // Bytes per frame: 2 (mono 16-bit) or 4 (stereo 16-bit).
            // The DataLine.Info constructor takes a byte-count buffer size.
            int bufferBytes = bufferFrames << (stereo ? 2 : 1);
            DataLine.Info lineInfo = new DataLine.Info(SourceDataLine.class, audioFormat, bufferBytes);
            sourceDataLine = (SourceDataLine) AudioSystem.getLine(lineInfo);
            sourceDataLine.open();
            sourceDataLine.start();
            lineBufferFrames = bufferFrames;
        } catch (LineUnavailableException ex) {
            // If the OS rejected the buffer size, check whether it was a
            // power-of-two issue.  GameCharacter.countBits (obf: ta.a) counts
            // set bits (popcount); a value with exactly one set bit is a power
            // of two.  If bufferFrames already has only 1 bit set, we cannot
            // round up further — give up and re-throw.
            if (GameCharacter.countBits(bufferFrames, (byte) -59) != 1) { // obf: ta.a(var1,(byte)-59)
                // BZip.nextPowerOfTwo (obf: aa.a) rounds up to the next power
                // of two; retry with that size.
                openLine(BZip.nextPowerOfTwo(bufferFrames, false)); // obf: aa.a(var1,false)
                return;
            }
            // Already a power of two and the OS still refused — give up.
            sourceDataLine = null;
            throw ex;
        }
    }

    /**
     * Initialises the audio subsystem: scans available system mixers (skipping
     * any SoundMAX entry to avoid driver bugs), then creates the
     * {@link AudioFormat} and pre-allocates the PCM write buffer.
     *
     * <p>The {@link AudioFormat} is: PCM signed, 16-bit, little-endian,
     * at the engine sample rate ({@link AudioChannel#sampleRate}),
     * with 1 channel (mono) or 2 channels (stereo) depending on
     * {@link AudioChannel#stereo}.
     *
     * <p>The {@code Component} argument is accepted for API consistency with
     * the parent class ({@code AudioChannel.initOutput(Component)}) but is not
     * used in this implementation — the Java Sound path does not require an AWT
     * peer.
     *
     * @param parentComponent  AWT component (unused; present for API symmetry)
     *
     * obf: {@code a(Component)}
     */
    @Override
    public final void initOutput(Component parentComponent) { // obf: a(Component)
        // Enumerate all system mixer endpoints and note any SoundMAX device.
        // SoundMAX AC'97 drivers were known to cause hangs or silence under
        // early Java Sound implementations on Windows XP-era hardware, so the
        // client just avoids selecting them.  The loop body is intentionally
        // empty after the check — this is only a detection/logging step with no
        // branching action in the shipped code (the if-body was dead at
        // compile time).
        Mixer.Info[] mixerInfoList = AudioSystem.getMixerInfo();
        if (mixerInfoList != null) {
            for (Mixer.Info mixerInfo : mixerInfoList) {
                if (mixerInfo != null) {
                    String mixerName = mixerInfo.getName();
                    // If the mixer name contains "soundmax" (case-insensitive),
                    // skip it — SoundMAX driver workaround.
                    // (The if-body in the obfuscated source was empty; the
                    // detection logic was likely once used to avoid selecting
                    // that mixer but became a no-op in this revision.)
                    if (mixerName != null && mixerName.toLowerCase().indexOf(SOUNDMAX_MIXER_FILTER) >= 0) {
                        // SoundMAX mixer detected — no action taken in this revision.
                    }
                }
            }
        }

        // Build the AudioFormat for 16-bit signed PCM, little-endian.
        // sampleRate (obf: t) and stereo (obf: i) are inherited static
        // fields set by AudioChannel.configure().
        audioFormat = new AudioFormat(
                (float) sampleRate,       // obf: t  — engine sample rate (e.g. 8000 Hz)
                16,                       // bits per sample
                stereo ? 2 : 1,           // channels: 1 = mono, 2 = stereo
                true,                     // signed PCM
                false                     // little-endian (not big-endian)
        );

        // Pre-allocate the PCM write buffer.
        // 256 frames * 2 bytes/sample * channels:
        //   mono:   256 << 1 =  512 bytes
        //   stereo: 256 << 2 = 1024 bytes
        pcmWriteBuffer = new byte[256 << (stereo ? 2 : 1)];
    }

    // -------------------------------------------------------------------------
    // XOR string-pool decoders  (obfuscation infrastructure — kept for
    // reference; not called at runtime after static-field initialisation)
    // -------------------------------------------------------------------------

    /**
     * First stage of the two-step XOR string-pool decoder.
     * Converts a literal obfuscated {@link String} to a {@code char[]}.
     * For strings shorter than 2 characters the single character is XOR'd
     * with {@code 21}; longer strings are returned as plain char arrays
     * (the real XOR work happens in the second stage).
     *
     * obf: {@code z(String)}
     *
     * @param encoded  obfuscated string literal embedded in the class file
     * @return         intermediate char array for the second decode stage
     */
    private static char[] decodeStep1(String encoded) { // obf: z(String)
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 21);
        }
        return chars;
    }

    /**
     * Second stage of the two-step XOR string-pool decoder.
     * XORs each character with a rotating 5-byte key table
     * {@code [14, 117, 106, 37, 21]} (index {@code i % 5}) and interns
     * the result.
     *
     * <p>Applying both stages to the literal
     * {@code "}Kqc"} yields {@code "soundmax"}.
     *
     * obf: {@code z(char[])}
     *
     * @param chars  intermediate char array from {@link #decodeStep1}
     * @return       decoded, interned plain-text string
     */
    private static String decodeStep2(char[] chars) { // obf: z(char[])
        // Rotating 5-byte XOR key used by every class's string pool.
        final int[] KEY = {14, 117, 106, 37, 21};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char) (chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
