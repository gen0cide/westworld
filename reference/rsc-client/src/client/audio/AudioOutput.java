package client.audio;

/**
 * Marker interface for the audio output backend used by the RSC client's audio subsystem.
 *
 * <p>The client supports two audio output implementations selected at runtime:
 * <ul>
 *   <li>{@code DirectSoundPlayer} (obf: {@code rb}) — Microsoft DirectSound path via
 *       {@code com.ms.directX.DirectSound}/{@code DSBufferDesc}/{@code DSCursors};
 *       used on the Microsoft J++ VM (Windows-only native path).  Declares
 *       {@code implements AudioOutput} in its class header.</li>
 *   <li>{@code SourceLinePlayer} (obf: {@code pb}) — standard {@code javax.sound.sampled}
 *       path; this class extends {@code AudioChannel} ({@code sa}) and does NOT implement
 *       {@code AudioOutput} directly — it is the pure-Java fallback.</li>
 * </ul>
 *
 * <p>The interface itself carries no methods; its sole purpose is to tag
 * {@code DirectSoundPlayer} so that the {@code AudioChannel} ({@code sa}) factory
 * ({@code AudioChannel.a(LoaderThread, Component, int, int)}) can distinguish the
 * DirectSound backend from the Java Sound backend via an {@code instanceof} check
 * (or simply by construction — the factory always allocates {@code SourceLinePlayer},
 * so the DirectSound path would be selected by a different code route).
 *
 * <p>In the oracle (mudclient 204), no equivalent interface exists because that build
 * predates the Microsoft J++ / DirectX integration.  The interface was introduced in
 * the rev-233+ J++ builds alongside the {@code com.ms.directX} package.
 *
 * <p>Obfuscated name: {@code ma}  (original package: default/unnamed)
 */
// obf: ma
public interface AudioOutput {
    // Empty marker interface — no methods.
    // DirectSoundPlayer (obf: rb) is the sole implementor in this revision.
}
