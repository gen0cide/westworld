package client.util;

import java.net.Socket;

/**
 * StreamFactory — static factory used by the engine to obtain two kinds of
 * runtime objects:
 *
 * <ol>
 *   <li><b>EntityDef record data</b>: {@link #lookupEntityDefRecord} delegates to
 *       {@link client.data.EntityDef#extractArchiveEntry} to search an archive byte-array
 *       for a named record and return its raw bytes.</li>
 *   <li><b>Proxy-aware sockets</b>: {@link #createSocketFactory} allocates a
 *       {@link client.net.ProxySocketFactory} pre-configured with the given host
 *       and port, returning it as the abstract {@link client.net.SocketFactory}.</li>
 * </ol>
 *
 * The class also holds the eight RSC chat colour-code strings ({@link #CHAT_COLOR_CODES})
 * used elsewhere in the engine, and three static profiling counters that the
 * obfuscator injected at method entry (all dead in production).
 *
 * <p>Obfuscated class name: {@code na}.  Package placed in {@code client.util}
 * because the class is a pure utility; it has no state beyond dead profiling
 * counters and the color-code table.</p>
 */
public final class StreamFactory {

    // -------------------------------------------------------------------------
    // Dead profiling fields (injected by obfuscator; incremented at method entry)
    // -------------------------------------------------------------------------

    /**
     * Per-method profiling counter for {@link #lookupEntityDefRecord}.
     * Incremented at method entry; never read by game logic.
     * obf: na.c
     */
    public static int lookupRecordCallCount;   // obf: c

    /**
     * Per-method profiling counter for {@link #createSocketFactory}.
     * Incremented at method entry; never read by game logic.
     * obf: na.b
     */
    public static int createFactoryCallCount;  // obf: b

    /**
     * Global size counter written by the static initializer of
     * {@link client.net.SocketFactory} (via bytecode: {@code putstatic na.e}).
     * Used as an array dimension for several engine subsystems.
     * Incremented at method entry; never read by game logic here.
     * obf: na.e
     */
    public static int engineArraySize;         // obf: e

    // -------------------------------------------------------------------------
    // Dead anti-tamper field
    // -------------------------------------------------------------------------

    /**
     * Null-out array used by the anti-tamper guard in
     * {@link #lookupEntityDefRecord}: the guard writes {@code null} here to
     * obscure control flow.  Never read by real game logic.
     * obf: na.a
     * Note: field letter 'a' — not to be confused with method overloads named 'a'
     * (see NAMING.md obf-i caveat).
     */
    public static int[] deadGuardArray;        // obf: a

    // -------------------------------------------------------------------------
    // Chat colour-code table
    // -------------------------------------------------------------------------

    /**
     * Eight RSC in-game chat colour-code strings, indexed 0–7.
     * Decoded from XOR pool ({@code z(z("…"))}):
     * <pre>
     *   [0] "@whi@"  – white
     *   [1] "@cya@"  – cyan
     *   [2] "@cya@"  – cyan
     *   [3] "@whi@"  – white
     *   [4] "@yel@"  – yellow
     *   [5] "@cya@"  – cyan
     *   [6] "@whi@"  – white
     *   [7] "@whi@"  – white
     * </pre>
     * obf: na.d
     */
    public static final String[] CHAT_COLOR_CODES = new String[]{
        "@whi@",   // [0]  obf: z(z("q,l?/"))
        "@cya@",   // [1]  obf: z(z("q8}7/"))
        "@cya@",   // [2]  obf: z(z("q8}7/"))
        "@whi@",   // [3]  obf: z(z("q,l?/"))
        "@yel@",   // [4]  obf: z(z("q\"a:/"))
        "@cya@",   // [5]  obf: z(z("q8}7/"))
        "@whi@",   // [6]  obf: z(z("q,l?/"))
        "@whi@",   // [7]  obf: z(z("q,l?/"))
    };

    // -------------------------------------------------------------------------
    // Factory methods
    // -------------------------------------------------------------------------

    /**
     * Looks up a named record in a JAG archive byte-array and returns its
     * raw (possibly decompressed) bytes.
     *
     * <p>This is a thin wrapper around
     * {@link client.data.EntityDef#extractArchiveEntry(byte[], int, int, String, byte[])}
     * that supplies {@code -127} as the (unused) {@code destOffset} argument.
     * {@code extractArchiveEntry} always looks up the entry by name hash
     * ({@code hash = hash*61 + (ch-' ')}); there is no by-index mode here.
     * The {@code recordId} value is passed through as EntityDef's
     * {@code allocHint} (extra-allocation hint).</p>
     *
     * <p>Obfuscated signature: {@code static final byte[] a(String, int, byte[], int)}.
     * The fourth parameter ({@code var3 / dummyParam}) is an anti-tamper value;
     * the guard {@code if (dummyParam > -117) deadGuardArray = null} is always
     * taken but only writes null to a dead field — stripped here.</p>
     *
     * @param recordName  the upper-cased archive entry name to find
     * @param recordId    extra-allocation hint passed through as EntityDef's {@code allocHint}
     * @param archiveData raw JAG archive bytes to search
     * @return raw record bytes, or {@code null} if the name is not found
     *
     * obf: na.a(String, int, byte[], int):byte[]
     * Error string: "na.A(" (z[1]), args repr null/non-null via "{...}"(z[2])/"null"(z[0])
     */
    public static final byte[] lookupEntityDefRecord(
            String recordName,
            int    recordId,
            byte[] archiveData) {
        // Profiling counter (dead): ++lookupRecordCallCount;
        // Anti-tamper guard stripped: if (dummyParam > -117) { deadGuardArray = null; }
        // obf: t.a(var2, -127, var1, var0, null) — t = EntityDef
        // -127 is the (unused) destOffset arg; recordId maps to EntityDef's allocHint param.
        return client.data.EntityDef.extractArchiveEntry(archiveData, -127, recordId, recordName, null);
    }

    /**
     * Allocates a new {@link client.net.ProxySocketFactory} configured for the
     * given host and port, and returns it as the abstract
     * {@link client.net.SocketFactory}.
     *
     * <p>The obfuscated version accepted an extra first parameter
     * ({@code magicKey}) that was compared against the anti-tamper constant
     * {@code 4718} (0x126E); the factory was returned only when the key matched.
     * In practice every call site passes exactly 4718, so the guard is dead
     * anti-tamper and is stripped here.</p>
     *
     * <p>Obfuscated signature: {@code static final m a(int, int, String)}.
     * Maps to: {@code a(magicKey=4718, port, host)}.</p>
     *
     * @param port  TCP port for the outgoing connection
     * @param host  hostname / IP address for the outgoing connection
     * @return a new {@link client.net.ProxySocketFactory} with the given host+port,
     *         cast to {@link client.net.SocketFactory}
     *
     * obf: na.a(int, int, String):m
     * Error string: "na.B(" (z[3])
     */
    public static final client.net.SocketFactory createSocketFactory(int port, String host) {
        // Profiling counter (dead): ++createFactoryCallCount;
        // Anti-tamper guard stripped: if (magicKey != 4718) return null;
        client.net.ProxySocketFactory factory = new client.net.ProxySocketFactory();
        factory.host = host;   // gb.h  – the hostname field inherited from SocketFactory
        factory.port = port;   // gb.f  – the port field inherited from SocketFactory
        return factory;
    }

    // -------------------------------------------------------------------------
    // XOR string-pool helpers (dead after class loading; kept for traceability)
    // -------------------------------------------------------------------------

    /**
     * First decode pass: if the encoded string has fewer than 2 characters,
     * XOR the sole character with {@code 'o'} (0x6F = 111); otherwise return
     * the char array unchanged.
     * obf: na.z(String):char[]
     */
    private static char[] z(String encoded) {
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char)(chars[0] ^ 'o');  // single-char key: 0x6F
        }
        return chars;
    }

    /**
     * Second decode pass: XOR each character with a position-dependent key from
     * the five-element table {@code [49, 91, 4, 86, 111]} (indices 0–4, repeating
     * mod 5), then intern the resulting string.
     * obf: na.z(char[]):String
     */
    private static String z(char[] chars) {
        for (int i = 0; i < chars.length; i++) {
            final int key;
            switch (i % 5) {
                case 0:  key = 49;  break;   // '1'
                case 1:  key = 91;  break;   // '['
                case 2:  key = 4;   break;
                case 3:  key = 86;  break;   // 'V'
                default: key = 111; break;   // 'o'
            }
            chars[i] = (char)(chars[i] ^ key);
        }
        return new String(chars).intern();
    }
}
