package client.nativeapi;

import com.ms.directX.DSBufferDesc;
import com.ms.directX.DSCursors;
import com.ms.directX.DirectSound;
import com.ms.directX.WaveFormatEx;

import client.audio.AudioOutput;

/**
 * DirectSoundPlayer (obfuscated {@code rb}) — the Win32/DirectX audio output backend.
 *
 * <p>This is the Microsoft J++ build's native audio sink. Where the pure-Java
 * revisions of the client route mixed PCM through {@code javax.sound.sampled}
 * (see the RSC 204 oracle {@code StreamAudioPlayer}), this build talks directly
 * to DirectSound via the {@code com.ms.directX} VM extension classes. It is the
 * concrete implementation of the {@link AudioOutput} (obfuscated {@code ma})
 * marker interface, so the audio mixer ({@code eb}/AudioMixer) can hold it
 * polymorphically as an output device regardless of platform.
 *
 * <p>In the jar that survives, this class is only a skeleton: the constructor
 * allocates the DirectX objects and the per-buffer bookkeeping arrays, but the
 * streaming/write/close methods that would normally lock the secondary buffer,
 * copy samples in, and advance the cursors are not present (they were either
 * stripped from this build or never emitted). The structure that does remain
 * tells us how the player was meant to work:
 *
 * <ul>
 *   <li>A single {@link DirectSound} device object (the COM {@code IDirectSound}
 *       handle) used to create secondary sound buffers.</li>
 *   <li>A {@link WaveFormatEx} describing the PCM wave format (sample rate, bit
 *       depth, channel count) of the buffers — the analogue of the oracle's
 *       {@code AudioFormat}.</li>
 *   <li>Two {@link DSBufferDesc} buffer descriptors and two {@link DSCursors}
 *       play/write cursor trackers — i.e. <b>double buffering</b>. The two-slot
 *       arrays let the player fill one secondary buffer while the other is being
 *       played, ping-ponging between them to stream continuously without gaps.</li>
 * </ul>
 *
 * <p>Because {@code com.ms.directX.*} is a proprietary Microsoft VM extension,
 * these imports will not resolve on a standard JDK; that is expected for this
 * Win32-only build.
 */
final class DirectSoundPlayer implements AudioOutput {

   /**
    * Number of DirectSound secondary buffers managed by this player — two, for
    * double-buffered (ping-pong) streaming: one buffer plays while the other is
    * refilled.
    */
   private static final int BUFFER_COUNT = 2;

   /**
    * Per-buffer DirectSound buffer descriptors ({@code DSBUFFERDESC}). Each
    * describes the capabilities/size/format of one secondary sound buffer that
    * would be created from the {@link DirectSound} device. (Obfuscated field
    * {@code b}.)
    */
   private DSBufferDesc[] bufferDescriptors = new DSBufferDesc[BUFFER_COUNT];

   /**
    * Per-buffer play/write cursor trackers. Each {@link DSCursors} holds the
    * current play and write cursor positions within the matching secondary
    * buffer, used to know how far playback has progressed and where it is safe
    * to write the next chunk of samples. (Obfuscated field {@code a}.)
    */
   private DSCursors[] cursors = new DSCursors[BUFFER_COUNT];

   /**
    * Initialises the DirectSound output device and allocates the double-buffer
    * bookkeeping. Throws {@link Exception} because creating the native DirectX
    * objects can fail when the device is unavailable.
    */
   public DirectSoundPlayer() throws Exception {
      // Create the DirectSound device handle (IDirectSound) used to allocate
      // secondary buffers, and the wave-format descriptor for those buffers.
      // The constructed objects are not retained as fields in this skeleton
      // build; instantiating them performs the native initialisation.
      new DirectSound();
      new WaveFormatEx();

      // Allocate one buffer descriptor per secondary buffer (double buffering).
      for (int bufferIndex = 0; bufferIndex < BUFFER_COUNT; bufferIndex++) {
         this.bufferDescriptors[bufferIndex] = new DSBufferDesc();
      }

      // Allocate one play/write cursor tracker per secondary buffer.
      for (int bufferIndex = 0; bufferIndex < BUFFER_COUNT; bufferIndex++) {
         this.cursors[bufferIndex] = new DSCursors();
      }
   }
}
