package client.util;

/**
 * IntHolder — shared static scratchpad used by several decoders and net helpers.
 *
 * This is a plain data holder with no methods of its own.  Three static fields
 * are assigned and read by unrelated classes that need a lightweight, allocation-
 * free place to park transient values:
 *
 *   bufferOffset  — shared read cursor into Packet.rawBytes (b.v).  Multiple
 *                   static helpers in ChatCipher (v), EntityDef (t), NameTable (ub),
 *                   and SocketFactory (m) advance this cursor as they consume bytes
 *                   from the shared packet buffer (byte/short/int reads step it by
 *                   1/2/4 respectively).  It is zeroed to 0 by the SocketFactory
 *                   bootstrap before each decode pass.
 *
 *   scratchInts   — int[] scratch array, allocated once to gb.p (ProxySocketFactory
 *                   capacity) elements by the SocketFactory bootstrap.  Used during
 *                   cipher/decode passes in SocketFactory.a(byte[],byte,boolean) to
 *                   accumulate per-slot decoded values, then compared against −1 XOR
 *                   (i.e. checked for the sentinel ~0) to drive the decode loop.
 *
 *   clientTag     — optional String override used by Utility.reportError to build
 *                   the "clienterror.ws" URL.  When non-null it substitutes for the
 *                   default numeric client-version token in the error-report query
 *                   string.  Null at startup and typically left null; may be set by
 *                   the applet parameter or server handshake to tag crash reports
 *                   with a build identifier.
 *
 * The class is declared final (no subclassing) and carries only static state —
 * it is never instantiated.  The three fields are intentionally public/package so
 * that every class in the default (obfuscated) package can access them directly.
 *
 * Obfuscated name: ka
 * Package: client.util  (default package in obfuscated jar)
 */
final class IntHolder {

    /**
     * Shared read cursor into Packet.rawBytes (b.v).
     *
     * Incremented atomically (no synchronisation — single-threaded decode paths)
     * by 1 (byte read in ChatCipher.readByte), 2 (short read in EntityDef.readShort),
     * or 4 (int read in NameTable.readInt / SocketFactory).  Reset to 0 by the
     * SocketFactory bootstrap before starting a new decode pass.
     *
     * Obf: ka.b
     */
    static int bufferOffset; // obf: b

    /**
     * Scratch int array used during cipher/decode passes in SocketFactory.
     *
     * Allocated to ProxySocketFactory.capacity (gb.p) elements by the SocketFactory
     * class-init bootstrap.  Each slot receives a decoded cipher value; slots are
     * then tested against the sentinel −1 (stored XOR-encoded as ~0) to decide
     * whether a string-pool entry has been decoded.
     *
     * Obf: ka.c
     */
    static int[] scratchInts; // obf: c

    /**
     * Optional client build-tag String for error reporting.
     *
     * When non-null, Utility.reportError substitutes this value for the default
     * numeric version token in the "clienterror.ws" query string, allowing crash
     * reports from a tagged build to be distinguished server-side.  Null at
     * startup; set externally (applet param or server) if desired.
     *
     * Obf: ka.a
     */
    static String clientTag; // obf: a
}
