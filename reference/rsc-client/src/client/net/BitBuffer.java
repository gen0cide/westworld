package client.net;

import java.math.BigInteger;

/**
 * BitBuffer — obf: {@code ja}
 *
 * <p>Extends {@link Buffer} (obf: {@code tb}) with bit-granularity reading support on top of the
 * raw byte array {@code data[]} (obf: {@code F}) inherited from the superclass.  An independent
 * bit-position cursor {@link #bitPos} tracks the current read head in <em>bits</em> (rather than
 * bytes); callers must bracket bit reads with {@link #initBitAccess()} / {@link #finishBitAccess()}
 * to keep the byte-level cursor ({@code Buffer.offset}, obf: {@code w}) in sync.
 *
 * <p>The class also stores the RSA public modulus used by the login handshake.  When the client
 * sends login credentials it calls {@code Buffer.rsaEncrypt()} (obf: {@code tb.a(BigInteger,int,
 * BigInteger)}) passing the exponent {@code 65537} and {@code K} as the modulus, producing a
 * {@code BigInteger.modPow} encrypted block.  The modulus is the server's 1024-bit public key,
 * embedded here as a hex string that is decoded at class-load time from the XOR string pool.
 *
 * <p>Packet uses this class as its bit-access buffer: {@code Packet} (obf: {@code b}) holds a
 * field of type {@code ja f} and calls these methods during world-state update parsing where run-
 * length compressed flags are packed at sub-byte granularity (e.g. player movement flags, NPC
 * update masks).
 *
 * <p>Revision: ~233–235, Microsoft J++ build.
 */
final class BitBuffer extends Buffer {

    // -------------------------------------------------------------------------
    // Static fields
    // -------------------------------------------------------------------------

    /**
     * RSA 1024-bit public modulus used for login-block encryption.
     *
     * <p>Decoded at class-load time from a two-pass XOR pool (inner {@code z(String)} produces the
     * raw char array; outer {@code z(char[])} XORs each character with the repeating key
     * {@code [48, 50, 27, 36, 70]}).  The resulting hex string is passed to the
     * {@link BigInteger#BigInteger(String, int)} constructor with radix 16.
     *
     * <p>Decoded value (1024-bit, hex):
     * {@code ca950472ae9765185bf290ff54a823b1d29b46dc3cf676203bb871efa278d9c4
     *        9e16defc53ff479305123454505082f4700b0da381047f51b872f9bbeea653f2
     *        1fd248a10ff5239b30234add35913cb6068d316edd418611334ae047fcd9acb7
     *        b0c13b30393a26204dc85183e0a95555c01bee800440e974bb9b441f464f4057}
     *
     * obf: {@code K}
     */
    static BigInteger RSA_MODULUS = new BigInteger(
        // z(z("SS\"...")) decoded → the hex string below (1024-bit RSA public key)
        "ca950472ae9765185bf290ff54a823b1d29b46dc3cf676203bb871efa278d9c4"
        + "9e16defc53ff479305123454505082f4700b0da381047f51b872f9bbeea653f2"
        + "1fd248a10ff5239b30234add35913cb6068d316edd418611334ae047fcd9acb7"
        + "b0c13b30393a26204dc85183e0a95555c01bee800440e974bb9b441f464f4057",
        16
    );

    // Dead profiling / unused counters — kept as obf traces, never read by game logic.
    /** obf: {@code I} — dead per-call profiling counter on {@link #finishBitAccess()} */
    static int PROF_FINISH;   // obf: I
    /** obf: {@code J} — dead per-call profiling counter on {@link #readBits(int)} */
    static int PROF_READ;     // obf: J
    /** obf: {@code G} — dead per-call profiling counter on {@link #initBitAccess()} */
    static int PROF_INIT;     // obf: G
    /** obf: {@code H} — dead per-call profiling counter on {@link #getBitPosition()} */
    static int PROF_GET;      // obf: H

    /** obf: {@code N} — 100-element int array; unused dead field, never written meaningfully */
    static int[] UNUSED_N = new int[100]; // obf: N
    /** obf: {@code L} — String array; unused dead field */
    static String[] UNUSED_L;             // obf: L

    // -------------------------------------------------------------------------
    // Instance fields
    // -------------------------------------------------------------------------

    /**
     * Current bit-level read cursor, measured in <em>bits</em> from the start of {@code data[]}.
     *
     * <p>Invariant: {@code M == offset * 8 + bitOffset}, where {@code bitOffset ∈ [0, 7]}.
     * {@link #initBitAccess()} initialises this from the byte-level cursor; {@link #finishBitAccess()}
     * converts it back.
     *
     * obf: {@code M}
     */
    private int bitPos; // obf: M

    // XOR-encoded error-message string pool (decoded at class init).
    // Entries: [0]="ja.D(", [1]="ja.A(", [2]="ja.C(", [3]="ja.B("
    // obf: O
    private static final String[] ERROR_TAGS = new String[]{
        /* O[0] */ "ja.D(",  // obf: z(z("ZS5`n"))
        /* O[1] */ "ja.A(",  // obf: z(z("ZS5en"))
        /* O[2] */ "ja.C(",  // obf: z(z("ZS5gn"))
        /* O[3] */ "ja.B("   // obf: z(z("ZS5fn"))
    };

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /**
     * Allocates a new {@code BitBuffer} backed by a fresh byte array of {@code capacity} bytes.
     * Delegates to {@link Buffer#Buffer(int)} which allocates {@code data[]} and zeroes the byte
     * cursor.
     *
     * obf: {@code ja(int)}
     *
     * @param capacity number of bytes to allocate
     */
    BitBuffer(int capacity) {
        super(capacity);
    }

    // -------------------------------------------------------------------------
    // Bit-access lifecycle
    // -------------------------------------------------------------------------

    /**
     * Prepares the buffer for bit-level reading by aligning the bit cursor to the current byte
     * position.
     *
     * <p>Sets {@code bitPos = 8 * offset} so that subsequent {@link #readBits(int)} calls start
     * at exactly the byte boundary the byte-level cursor was sitting on.  Call this before any
     * sequence of {@link #readBits(int)} calls.
     *
     * obf: {@code i(int)} — the {@code int} parameter was an anti-tamper dummy; stripped.
     * Original guard: {@code if (var1 != -2231) this.k(-48);} (dead path).
     */
    final void initBitAccess() {
        // Convert the byte-level cursor into a bit-level cursor.
        bitPos = 8 * offset; // obf: this.M = 8 * this.w  (Buffer.offset, obf w)
        // ++PROF_INIT; // dead profiling counter stripped
    }

    /**
     * Finishes a bit-reading session by advancing the byte-level cursor past any bytes that were
     * consumed during bit reads, rounding up to the next whole byte.
     *
     * <p>The formula {@code offset = (bitPos + 7) / 8} is equivalent to a ceiling division: it
     * ensures that even a partial byte consumed by the last {@link #readBits(int)} call moves the
     * byte cursor to the next byte boundary.  Call this after the last {@link #readBits(int)} in a
     * sequence, then resume byte-level reads normally.
     *
     * obf: {@code j(int)} — the {@code int} parameter was an anti-tamper dummy; stripped.
     * Original guard: {@code if (var1 != 25505) this.f(12, -68);} (dead path — the f() call with
     * magic arg 12 is itself guarded by an anti-tamper check that would corrupt K; both are dead).
     */
    final void finishBitAccess() {
        // Ceil-divide bit cursor back to a byte offset.
        offset = (7 + bitPos) / 8; // obf: this.w = (7 + this.M) / 8  (Buffer.offset, obf w)
        // ++PROF_FINISH; // dead profiling counter stripped
    }

    // -------------------------------------------------------------------------
    // Core bit reader
    // -------------------------------------------------------------------------

    /**
     * Reads the next {@code numBits} bits from the bit stream and returns them as an unsigned
     * integer.
     *
     * <p>Bit ordering matches the server's encoding: the most-significant bits of a byte are
     * consumed first.  The algorithm mirrors {@code Utility.getBitMask()} from the rev-204 oracle
     * ({@code mudclient204/src/Utility.java}):
     *
     * <ol>
     *   <li>Determine the byte index ({@code byteIdx = bitPos >> 3}) and the number of bits
     *       remaining in that byte ({@code bitsLeft = 8 - (bitPos & 7)}).</li>
     *   <li>While the requested count exceeds the bits left in the current byte, consume the whole
     *       remaining fragment, shift it into the accumulator, subtract from the count, and advance
     *       to the next byte (8 bits left).</li>
     *   <li>When the request exactly fills the remaining bits, mask and add directly.  When fewer
     *       bits are needed than remain, right-shift to align the desired bits to the LSB and mask.</li>
     * </ol>
     *
     * <p>The bitmask table is {@code Utility.BITMASK} (obf: {@code mb.i}), where entry {@code n}
     * equals {@code (1 << n) - 1} for {@code n ∈ [0, 31]} and {@code -1} at index 32.
     *
     * obf: {@code f(int, int)} — first parameter was anti-tamper dummy; stripped.
     * Original dead guard: {@code if (var1 >= -67) K = null;} — a trap that would null the RSA
     * modulus; removed.  Second parameter {@code var2} is {@code numBits}.
     *
     * @param numBits number of bits to read (1–32)
     * @return the unsigned integer value of the next {@code numBits} bits
     */
    final int readBits(int numBits) {
        // obf: var3 = this.M >> 3; var4 = -(this.M & 7) + 8;
        int byteIdx = bitPos >> 3;           // byte index into data[]
        int bitsLeft = 8 - (bitPos & 7);    // bits remaining in the current byte (1–8)

        int result = 0;
        bitPos += numBits; // advance the bit cursor by the requested count

        // Consume whole bytes while more bits are requested than remain in the current byte.
        // Condition: ~bitsLeft > ~numBits  ⟺  numBits > bitsLeft  (complement reversal of order)
        while (numBits > bitsLeft) {
            // Take all bitsLeft bits from data[byteIdx], shift them to their final position, and OR in.
            result += (Utility.BITMASK[bitsLeft] & data[byteIdx++]) << (numBits - bitsLeft);
            numBits -= bitsLeft;
            bitsLeft = 8; // fresh byte: all 8 bits available
        }

        // Consume the remaining numBits from the current byte (numBits <= bitsLeft now).
        if (numBits == bitsLeft) {
            // Exactly fills the remaining bits in this byte — just mask.
            result += data[byteIdx] & Utility.BITMASK[bitsLeft];
        } else {
            // Fewer bits needed than remain — right-align the target bits and mask.
            // obf: var5 += this.F[var3] >> -var2 + var4 & mb.i[var2];
            //   = data[byteIdx] >> (bitsLeft - numBits) & BITMASK[numBits]
            result += (data[byteIdx] >> (bitsLeft - numBits)) & Utility.BITMASK[numBits];
        }

        return result;
    }

    // -------------------------------------------------------------------------
    // Accessor (used internally and from Packet)
    // -------------------------------------------------------------------------

    /**
     * Returns the current bit-level cursor position (in bits from the start of {@code data[]}).
     *
     * <p>Packet (obf: {@code b}) uses this to determine how many bits have been consumed when
     * switching back to byte-level I/O.
     *
     * obf: {@code k(int)} — parameter was anti-tamper dummy; stripped.
     * Original guard: {@code if (var1 != -31874) return 40;} (dead path).
     *
     * @return current value of {@link #bitPos}
     */
    final int getBitPosition() {
        return bitPos; // obf: return this.M
    }

    // -------------------------------------------------------------------------
    // XOR string-pool helpers (class-private; one copy per obfuscated class)
    // -------------------------------------------------------------------------
    // These two z() overloads implement the two-pass XOR decode used for all
    // per-class string constants.  They are NOT called at runtime outside of
    // static initialisation; we keep them for fidelity but they are effectively
    // dead after class loading.

    /**
     * Pass 1: converts the encoded literal String to a char array.  For strings longer than one
     * character the array is returned unchanged (the inner XOR layer is already baked into the
     * byte values of the source literal).  For a single-character string the sole char is XOR'd
     * with {@code 0x46} ('F').
     *
     * obf: {@code z(String)}
     */
    private static char[] z(String encoded) {
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char)(chars[0] ^ 0x46);
        }
        return chars;
    }

    /**
     * Pass 2: XOR-decodes a char array using a 5-byte repeating key
     * {@code [48, 50, 27, 36, 70]} and interns the result.
     *
     * <p>Key schedule by position {@code i % 5}:
     * <pre>
     *   0 → 48 ('0')
     *   1 → 50 ('2')
     *   2 → 27 (ESC)
     *   3 → 36 ('$')
     *   4 → 70 ('F')
     * </pre>
     *
     * obf: {@code z(char[])}
     */
    private static String z(char[] encoded) {
        int len = encoded.length;
        for (int i = 0; i < len; i++) {
            byte key;
            switch (i % 5) {
                case 0:  key = 48; break;  // '0'
                case 1:  key = 50; break;  // '2'
                case 2:  key = 27; break;  // ESC
                case 3:  key = 36; break;  // '$'
                default: key = 70; break;  // 'F'
            }
            encoded[i] = (char)(encoded[i] ^ key);
        }
        return new String(encoded).intern();
    }
}
