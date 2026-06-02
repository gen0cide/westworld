package client;

import java.awt.Component;

import client.audio.AudioChannel;
import client.audio.SampleBuffer;
import client.audio.StreamMixer;
import client.data.NameHash;
import client.net.ChatCipher;
import client.net.ClientStream;
import client.scene.ImageLoader;
import client.shell.InputState;

/**
 * ClientSound — audio subsystem extracted from Mudclient (extract-delegate idiom).
 *
 * <p>Two methods, moved verbatim:
 * <ul>
 *   <li>{@link #playSound(int, String)} — decode a named {@code .pcm} sample out of the
 *       {@code Sounds} archive (m.Uh) and schedule it on the {@link StreamMixer} (m.hk).</li>
 *   <li>{@link #initSounds(int)} — load the {@code Sounds} archive into m.Uh, open an
 *       {@link AudioChannel} (m.ni) attached to the applet host component, and wire in a
 *       {@link StreamMixer} (m.hk).</li>
 * </ul>
 *
 * <p>Touches only m.Uh, m.ne, m.hk, m.ni; calls the static
 * {@code Mudclient.findStringInData(...)} and the inherited public {@code m.readDataFile(...)}.
 *
 * <p>NOTE (behaviour-preserving deviation): in the original {@code initSounds}, the
 * AudioChannel host fallback was {@code host = this;} — {@code this} being the Mudclient
 * (an {@link java.applet.Applet}, hence a {@link Component}). In this delegate {@code this}
 * is the ClientSound (not a Component), so the fallback becomes {@code host = m;} to keep the
 * same runtime host object. That single {@code this -> m} substitution is the only change from
 * verbatim motion.
 */
class ClientSound {
    final Mudclient m;

    ClientSound(Mudclient m) {
        this.m = m;
    }

    // -------------------------------------------------------------------------
    // playSound  — obf: void a(int,String)
    // -------------------------------------------------------------------------

    /** Play a named .pcm sound effect at full volume (256). Looks the sample up in the Uh sound
     *  archive by name+".pcm" (STRINGS[515]). Muted while sleeping (ne). */
    final void playSound(int param, String name) {
        if (m.hk == null) return;   // hk
        if (m.ne) return;                     // sleeping → no sfx
        if (param >= -43) return;                // anti-tamper guard

        int offset = NameHash.getFileOffset(name + Mudclient.STRINGS[515], (byte) 68, m.Uh);   // NameHash.a → byte offset
        int length = Mudclient.findStringInData(m.Uh, name + Mudclient.STRINGS[515], -125);     // a(byte[],String,int) → length
        if (length == 0) return;                 // not found (~len == -1 idiom)

        SampleBuffer sample = new SampleBuffer(8000, ChatCipher.translate(m.Uh, length, offset), 0, length); // SampleBuffer / ChatCipher.a
        m.hk.scheduleStage(sample, 100, 256);
    }

    // -------------------------------------------------------------------------
    // initSounds  — obf: void E(int)
    // -------------------------------------------------------------------------

    /** Bring up the audio engine: load the "Sounds" archive (Uh), open an AudioChannel (ni) at
     *  22050 Hz attached to the applet host component, and wire in a StreamMixer (hk).
     *
     *  FIX vs old: the AudioChannel.a host arg is ImageLoader.k (pa.k), not "Timer.k". */
    final void initSounds(int param) {
        if (param > -55) return;   // anti-tamper guard

        m.Uh = m.readDataFile(Mudclient.STRINGS[345], 90, 10, 66);   // "Sounds" archive
        try {
            AudioChannel.configure(22050, false, 1);   // AudioChannel static init

            Object host;
            if (InputState.gameFrame != null) {           // InputState.a (applet host) takes priority
                host = InputState.gameFrame;
            } else if (ClientStream.applet != null) {   // ClientStream.ctrlDown fallback
                host = ClientStream.applet;
            } else {
                host = m;   // delegate: original was `this` (the Mudclient Component) — see class note
            }

            m.ni = AudioChannel.create(ImageLoader.imageWidthCarrier, (Component) host, 0, 22050);   // FIX: ImageLoader.k, not Timer.k
            m.hk = new StreamMixer();          // StreamMixer
            m.ni.setFilterChain(m.hk);
        } catch (Throwable t) {
            System.out.println(Mudclient.STRINGS[344] + t);   // "Unable to init sounds: "
        }
    }
}
