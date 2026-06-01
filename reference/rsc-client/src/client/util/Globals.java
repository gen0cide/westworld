package client.util;

import java.applet.Applet;

/**
 * Globals — engine-wide static singleton holder (obf: {@code l}).
 *
 * <p>This is the simplest class in the client: a plain container for three
 * static fields that need to be reachable from many otherwise-unrelated
 * subsystems without creating circular dependencies between the large top-level
 * classes ({@code Mudclient}, {@code GameShell}, {@code SocketFactory}, etc.).
 *
 * <p>Field roles (inferred from every call-site in the decompiled corpus):
 * <ul>
 *   <li>{@link #applet} — the live {@link Applet} reference, set once by the
 *       bootstrap loader ({@code c} / LoaderThread) and read by
 *       {@code Utility.reportError} ({@code mb}) to form an error-report URL
 *       via {@code Applet.getCodeBase()}.  Checked for {@code null} before
 *       use.</li>
 *   <li>{@link #paramNames} — the HTML/PARAM name strings passed into the
 *       applet; sized and populated inside {@code SocketFactory.a(…)} ({@code m})
 *       alongside several other per-subsystem string arrays at startup
 *       ({@code putstatic l.a} / {@code getstatic l.a} at m.java bytecode
 *       offsets 0x875 / 0x8af).  Used by {@code Mudclient} when reading
 *       player-list data.</li>
 *   <li>{@link #strings} — a 100-element scratch/intern pool for decoded
 *       in-game strings; written and shifted by {@code Mudclient} in several
 *       chat/overhead-text routines ({@code getstatic l.c} at client.java
 *       bytecode offsets 0x280, 0x28b, 0x2b8, 0x0f3, 0x057, 0x187, 0x1ee,
 *       0x242, 0x227, 0x29b; shift at 0x… in the "remove from list" loop at
 *       client.java line 40378).</li>
 * </ul>
 *
 * <p>No methods, no constructor logic, no obfuscation artifacts — this class
 * had nothing to strip.
 */
public final class Globals {

    /**
     * The running {@link Applet} instance, used to obtain the code-base URL
     * for error reporting ({@code Utility.reportError} / {@code mb.a}).
     * Null until the bootstrap loader assigns it; all consumers guard with
     * {@code if (Globals.applet == null) return;}.
     *
     * obf: {@code l.b}
     */
    public static Applet applet; // obf: b

    /**
     * HTML PARAM names (or XOR-decoded string constants) allocated once at
     * startup by {@code SocketFactory} ({@code m}) in the same block that
     * sizes all other per-subsystem name arrays.  The array length mirrors
     * {@code Surface.Db} (the total number of interned string slots for this
     * revision).  Read by {@code Mudclient} during player-list rendering to
     * build display-name strings.
     *
     * obf: {@code l.a}
     */
    public static String[] paramNames; // obf: a

    /**
     * A 100-slot ring/scratch pool for decoded overhead chat strings and
     * other short-lived in-game text.  Entries are shifted downward when the
     * list shrinks (see the {@code Mudclient} "remove from list" loop that
     * does {@code l.c[i] = l.c[i+1]}).  Indexed by {@code LinkedQueue.g}
     * ({@code db.g}) which tracks the current fill count.
     *
     * obf: {@code l.c}
     */
    public static String[] strings = new String[100]; // obf: c
}
