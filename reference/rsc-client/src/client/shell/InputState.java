package client.shell;

/**
 * InputState — static cross-class holder for the two kinds of shared global
 * state that do not belong to a single object: the AWT host window and the
 * two parallel per-slot data arrays that are initialised from the server data
 * stream and consulted at render/game-logic time.
 *
 * <h3>Role in the engine</h3>
 * <ul>
 *   <li>{@link #gameFrame} is set once (in {@code GameShell.a(…)}) when the
 *       standalone-application window is created. The rest of {@code GameShell}
 *       uses it as a null-sentinel: if it is non-null the client is running as
 *       a standalone application rather than an Applet, and certain AWT calls
 *       are routed through the frame rather than through the applet.</li>
 *   <li>{@link #pixelData} and {@link #slotFlags} are parallel arrays whose
 *       length equals the entity/slot count decoded from the initial server data
 *       stream (stored in {@code ProxySocketFactory.p}).  They are filled by
 *       the global bootstrap method in {@code SocketFactory} and read back
 *       during rendering ({@code pixelData}) and availability checks
 *       ({@code slotFlags}).  On slot reset, {@code pixelData[i]} is zeroed
 *       and {@code slotFlags[i]} is set to {@code 1}.</li>
 *   <li>{@link #nameDataBuffer} is a raw byte array fetched from the server
 *       (URL index 7 of the {@code SocketFactory} XOR string pool).  It is
 *       consumed sequentially by {@code ISAAC.a(byte)} which reads
 *       null-terminated strings out of it using the {@code DownloadWorker.p}
 *       cursor.</li>
 * </ul>
 *
 * <h3>Obfuscation note</h3>
 * The original class name {@code kb} carries no semantic hint; the rename to
 * {@code InputState} follows the description in the naming map
 * ("static input/framebuffer holder").  All four fields were single-letter
 * ({@code a}, {@code b}, {@code c}, {@code d}) with no methods — there are
 * no obfuscation artefacts to strip beyond the field names themselves.
 *
 * <p>Package: {@code client.shell}.
 * <p>Obf class: {@code kb}.
 */
final class InputState {

    /**
     * The AWT {@code Frame} that hosts the client when running as a standalone
     * application (not as an Applet).  Created by
     * {@code GameShell.a(boolean,String,int,String,int,byte,int,int,int)} and
     * disposed on shutdown.  Null while running inside a browser.
     *
     * <p>obf: {@code kb.a}
     *
     * @see GameFrame
     */
    static GameFrame gameFrame = null; // obf: a

    /**
     * Per-slot pixel/color data array, parallel to {@link #slotFlags}.
     * Allocated to {@code ProxySocketFactory.p} entries and populated from the
     * server data stream via {@code NameTable.a(byte)} (4-byte big-endian int
     * reads through {@code Packet.v}).  Entries are used during rendering as
     * integer color/pixel values passed to the interpolation helper
     * {@code ISAAC.a(int,int,int,int,boolean,int,int,int)}.  On slot reset
     * the entry is written to {@code 0}.
     *
     * <p>obf: {@code kb.b}
     */
    static int[] pixelData; // obf: b

    /**
     * Per-slot availability/key-state flag array, parallel to
     * {@link #pixelData}.  Allocated to {@code ProxySocketFactory.p} entries
     * and populated from the server data stream via
     * {@code ChatCipher.a(int)} (single unsigned-byte reads through
     * {@code Packet.v}).  The client checks {@code slotFlags[i] == 1} to
     * determine whether a slot is currently active/available.  On slot reset
     * the entry is written to {@code 1}.
     *
     * <p>obf: {@code kb.c}
     */
    static int[] slotFlags; // obf: c

    /**
     * Raw server name data buffer.  Fetched once at bootstrap from the server
     * URL at string-pool index 7 of {@code SocketFactory} (decoded from the
     * XOR pool via the two {@code z()} helpers).  Read sequentially by
     * {@code ISAAC.a(byte)} which extracts null-terminated ASCII strings out
     * of it, advancing the {@code DownloadWorker.p} position cursor after
     * each character.  Set to {@code null} after the connection is torn down.
     *
     * <p>obf: {@code kb.d}
     */
    static byte[] nameDataBuffer; // obf: d

    // No constructor: all fields are static; the class is never instantiated.
    // The final modifier on the class (present in the obfuscated bytecode)
    // is preserved to match the original.
}
