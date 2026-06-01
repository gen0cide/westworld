package client.net;

import java.math.BigInteger;

// Import aliases for renamed classes (canonical names per NAMING.md)
// client.util.ErrorHandler  (obf: i)   — dual-purpose: catch-sink i.a(Throwable,String):la
//                                        AND the char-encoder i.a(int,int,int,CharSequence,byte,byte[]):int
//                                        that copies/encodes string bytes (handles Euro 8364, etc.).
//                                        NOTE: the real StringCodec is obf 'u' (client.net), NOT 'i'.
// client.ui.CharTable        (obf: ga)  — ga.a(int,int,int,byte[]) decodes string bytes
// client.util.Utility        (obf: mb)
// client.world.WorldEntity   (obf: w)  — provides CRC helper w.a(int,int,byte[],int)

/**
 * Buffer — general-purpose byte-stream read/write cursor used throughout the RSC network and data layer.
 *
 * <p>Wraps a raw {@code byte[]} with an integer cursor ({@link #offset}) and exposes big-endian
 * put/get methods for every width (byte, short, int3, int, long, variable-length int, null-terminated
 * string).  Also provides RSA {@code modPow} encryption for the login block.
 *
 * <p>Extends {@link StreamBase} ({@code ib}), which supplies the static URL-loader machinery shared
 * across the net/audio resource hierarchy.  {@link BitBuffer} ({@code ja}) subclasses this class and
 * adds bit-level reading for compressed data streams.
 *
 * <p>Oracle: {@code mudclient204/src/Buffer.java} — logic matches almost exactly; method names
 * transferred from there plus RevLang cross-references to LeadingBot / OpenRSC Payload235.
 *
 * <p>obf: {@code tb}
 */
public class Buffer extends StreamBase { // obf: tb extends ib

    // -----------------------------------------------------------------------
    // Dead profiling counters (obfuscation artifact — never read for logic)
    // Each method increments its own static counter; they are meaningless at
    // runtime.  Kept here for traceability; they could be deleted entirely.
    // -----------------------------------------------------------------------
    /** obf: B — profiling counter for {@link #getUnsignedShort()} */
    public static int _prof_getUnsignedShort;           // obf: B
    /** obf: v — profiling counter for {@link #readRawByte()} */
    public static int _prof_readRawByte;                // obf: v
    /** obf: r — profiling counter for {@link #getInt()} */
    public static int _prof_getInt;                     // obf: r
    /** obf: D — profiling counter for {@link #putBytes(int,int,int,byte[])} */
    public static int _prof_putBytes;                   // obf: D
    /** obf: x — profiling counter for {@link #putStringPrefixed(String,int)} */
    public static int _prof_putStringPrefixed;          // obf: x
    /** obf: g — profiling counter for {@link #getLong()} */
    public static int _prof_getLong;                    // obf: g
    /** obf: m — profiling counter for {@link #teaDecrypt(byte,int,int[],int)} */
    public static int _prof_teaDecrypt;                 // obf: m
    /** obf: o — profiling counter for {@link #getString()} */
    public static int _prof_getString;                  // obf: o
    /** (note: field "i" in obfuscated source shadows ErrorHandler class "i"; this is a counter) */
    public static int _prof_getBytes;                   // obf: i (counter, shadows class i/ErrorHandler)
    /** obf: A — profiling counter for {@link #putString(byte,String)} */
    public static int _prof_putString;                  // obf: A
    /** obf: q — profiling counter for {@link #putInt3(int,byte)} */
    public static int _prof_putInt3;                    // obf: q
    /** obf: n — profiling counter for {@link #getSmartUnsigned()} */
    public static int _prof_getSmartUnsigned;           // obf: n
    /** obf: E — profiling counter for {@link #putByte(int,int)} */
    public static int _prof_putByte;                    // obf: E
    /** obf: k — profiling counter for {@link #patchShortBack(int,int)} */
    public static int _prof_patchShortBack;             // obf: k
    /** obf: s — profiling counter for {@link #putInt(int,int)} */
    public static int _prof_putInt;                     // obf: s
    /** obf: C — profiling counter for {@link #putShort(int,int)} */
    public static int _prof_putShort;                   // obf: C
    /** obf: u — profiling counter for {@link #getSignedShort()} */
    public static int _prof_getSignedShort;             // obf: u
    /** obf: j — profiling counter for {@link #rsaEncrypt(BigInteger,int,BigInteger)} */
    public static int _prof_rsaEncrypt;                 // obf: j
    /** obf: z — profiling counter for {@link #toByteArray(int)} */
    public static int _prof_toByteArray;               // obf: z
    /** obf: p — profiling counter for {@link #getUnsignedByte()} */
    public static int _prof_getUnsignedByte;            // obf: p
    /** obf: h — profiling counter for {@link #putVariableLengthShort(int,byte)} */
    public static int _prof_putVariableLengthShort;     // obf: h
    /** obf: y — profiling counter for {@link #getSmartSigned()} */
    public static int _prof_getSmartSigned;             // obf: y
    /** obf: t — profiling counter for {@link #verifyCrc(int)} */
    public static int _prof_verifyCrc;                  // obf: t

    // -----------------------------------------------------------------------
    // Instance fields
    // -----------------------------------------------------------------------

    /**
     * The raw backing byte array.  Shared (not copied) when constructed from an existing array.
     * obf: F
     */
    public byte[] data;  // obf: F

    /**
     * Current read/write cursor position within {@link #data}.  Incremented by every put/get call.
     * obf: w
     */
    public int offset;   // obf: w

    /**
     * Scratch int[12] — appears to be dead/obfuscation padding (nulled inside putBytes without
     * any read path). Kept for binary-level fidelity.
     * obf: l
     */
    public static int[] _junkArray = new int[12]; // obf: l

    // -----------------------------------------------------------------------
    // XOR-encoded debug string pool (decoded at class-load time).
    // All entries are method-signature labels used only in catch-block error
    // messages; they carry no runtime logic.
    //
    // Decoded values (key table {87,76,30,114,122}, first pass len<2 XOR 'z'):
    //   G[ 0] = "tb.K("    G[ 1] = "tb.W("    G[ 2] = "tb.S("
    //   G[ 3] = "{...}"    G[ 4] = "null"      G[ 5] = "tb.AA("
    //   G[ 6] = "tb.T("    G[ 7] = "tb.M("     G[ 8] = "tb.H("
    //   G[ 9] = "tb.V("    G[10] = "tb.<init>(" G[11] = "tb.F("
    //   G[12] = "tb.I("    G[13] = "tb.P("     G[14] = "tb.G("
    //   G[15] = "tb.J("    G[16] = "tb.N("     G[17] = "tb.O("
    //   G[18] = "tb.E("    G[19] = "tb.BA("    G[20] = "tb.DA("
    //   G[21] = "tb.Q("    G[22] = "tb.L("     G[23] = "tb.R("
    //   G[24] = "tb.U("    G[25] = "tb.CA("
    // -----------------------------------------------------------------------
    private static final String[] ERROR_SIGS = new String[]{  // obf: G
        xorDecode(xorPass1("#.09R")),    // G[ 0] → "tb.K("
        xorDecode(xorPass1("#.0%R")),    // G[ 1] → "tb.W("
        xorDecode(xorPass1("#.0!R")),    // G[ 2] → "tb.S("
        xorDecode(xorPass1(",b0\\")), // G[ 3] → "{...}"
        xorDecode(xorPass1("99r")),   // G[ 4] → "null"
        xorDecode(xorPass1("#.03;")), // G[ 5] → "tb.AA("
        xorDecode(xorPass1("#.0&R")),    // G[ 6] → "tb.T("
        xorDecode(xorPass1("#.0?R")),    // G[ 7] → "tb.M("
        xorDecode(xorPass1("#.0:R")),    // G[ 8] → "tb.H("
        xorDecode(xorPass1("#.0$R")),    // G[ 9] → "tb.V("
        xorDecode(xorPass1("#.0N9%jLR")), // G[10] → "tb.<init>("
        xorDecode(xorPass1("#.04R")),    // G[11] → "tb.F("
        xorDecode(xorPass1("#.0;R")),    // G[12] → "tb.I("
        xorDecode(xorPass1("#.0\"R")),   // G[13] → "tb.P("
        xorDecode(xorPass1("#.05R")),    // G[14] → "tb.G("
        xorDecode(xorPass1("#.08R")),    // G[15] → "tb.J("
        xorDecode(xorPass1("#.0<R")),    // G[16] → "tb.N("
        xorDecode(xorPass1("#.0=R")),    // G[17] → "tb.O("
        xorDecode(xorPass1("#.07R")),    // G[18] → "tb.E("
        xorDecode(xorPass1("#.00;")),  // G[19] → "tb.BA("
        xorDecode(xorPass1("#.06;")),  // G[20] → "tb.DA("
        xorDecode(xorPass1("#.0#R")),    // G[21] → "tb.Q("
        xorDecode(xorPass1("#.0>R")),    // G[22] → "tb.L("
        xorDecode(xorPass1("#.0 R")),    // G[23] → "tb.R("
        xorDecode(xorPass1("#.0'R")),    // G[24] → "tb.U("
        xorDecode(xorPass1("#.01;")),  // G[25] → "tb.CA("
    };

    // -----------------------------------------------------------------------
    // Constructors
    // -----------------------------------------------------------------------

    /**
     * Allocates a new buffer of the given capacity via {@code Utility.allocateBytes}.
     * The cursor starts at 0.
     *
     * @param capacity number of bytes to allocate; obf param: {@code var1}
     * obf: tb(int)
     */
    public Buffer(int capacity) {  // obf: tb(int var1)
        // mb.a(int, byte) = Utility.allocateBytes — returns a new byte[] of the requested size
        this.data = mb.a(capacity, (byte) -104); // obf: this.F = mb.a(var1, (byte)-104)
        this.offset = 0;
    }

    /**
     * Wraps an existing byte array without copying.  The cursor starts at 0.
     *
     * @param src byte array to wrap; obf param: {@code var1}
     * obf: tb(byte[])
     */
    public Buffer(byte[] src) {  // obf: tb(byte[] var1)
        this.data = src;
        this.offset = 0;
    }

    // -----------------------------------------------------------------------
    // Write methods
    // -----------------------------------------------------------------------

    /**
     * Writes a single byte to the buffer at the current cursor and advances it.
     * The extra {@code dummy} parameter is an obfuscation anti-tamper guard
     * (checked against 43; value never matters at legal call sites — removed).
     *
     * @param value the byte value (only low 8 bits are written); obf param: {@code var1}
     * obf: c(int, int) — guard: 121 / ((-5 - dummy) / 32) junk arithmetic; param dummy dropped
     */
    public final void putByte(int value) {  // obf: c(int var1, int var2)
        // Anti-tamper: int var3 = 121 / ((-5 - var2) / 32);  ← dead, removed
        this.data[this.offset++] = (byte) value;
    }

    /**
     * Writes a 16-bit big-endian unsigned short.
     * The extra {@code dummy} parameter is an obfuscation anti-tamper guard
     * (equality checked against 393; removed).
     *
     * @param value the short value (16 bits written big-endian); obf param: {@code var2}
     * obf: e(int, int) — guard param dropped
     */
    public final void putShort(int value) {  // obf: e(int var1, int var2)
        // Vineflower/CFR both confirm n3>>8 then n3 when n2==393:
        this.data[this.offset++] = (byte) (value >> 8); // high byte
        this.data[this.offset++] = (byte) value;        // low byte
    }

    /**
     * Writes a 24-bit big-endian integer (3 bytes).
     * The extra {@code dummy} byte is an anti-tamper guard (checked against -13); removed.
     *
     * @param value the int value (bits 23-0 written big-endian); obf param: {@code var1}
     * obf: a(int, byte) — guard param dropped
     */
    public final void putInt3(int value) {  // obf: a(int var1, byte var2)
        this.data[this.offset++] = (byte) (value >> 16); // bits 23-16
        this.data[this.offset++] = (byte) (value >> 8);  // bits 15-8
        this.data[this.offset++] = (byte) value;         // bits  7-0
    }

    /**
     * Writes a 32-bit big-endian signed integer (4 bytes).
     * The extra {@code dummy} parameter is an anti-tamper guard (checked against -422797528); removed.
     *
     * @param value the int value; obf param: {@code var2}
     * obf: b(int, int) — guard param dropped
     */
    public final void putInt(int value) {  // obf: b(int var1, int var2)
        this.data[this.offset++] = (byte) (value >> 24); // bits 31-24
        this.data[this.offset++] = (byte) (value >> 16); // bits 23-16
        this.data[this.offset++] = (byte) (value >> 8);  // bits 15-8
        this.data[this.offset++] = (byte) value;         // bits  7-0
    }

    /**
     * Writes a variable-length unsigned integer (1 or 2 bytes) — "smart" encoding.
     * Values 0–127 are written as a single byte; 0–32767 are written as a 2-byte big-endian
     * short with the high bit of the first byte set (value += 32768).
     *
     * <p>Throws {@link IllegalArgumentException} if {@code value} is negative or >= 32768.
     *
     * @param value unsigned value in range [0, 32767]; obf param: {@code var1}
     * obf: b(int, byte) — guard param dropped
     */
    public final void putVariableLengthShort(int value) {  // obf: b(int var1, byte var2)
        if (value >= 0 && value < 128) {
            // Single-byte path: fits in 7 bits
            putByte(value);    // obf: this.c(var1, 43)
        } else if (value >= 0 && value < 32768) {
            // Two-byte path: set high bit on first byte to signal 2-byte form
            putShort(value + 32768); // obf: this.e(393, 32768 - -var1)  → 32768 + value
        } else {
            throw new IllegalArgumentException();
        }
    }

    /**
     * Writes a null-terminated ASCII string preceded by a null byte (double-null framing).
     * Format: {@code 0x00} | string bytes | {@code 0x00}.
     * Throws {@link IllegalArgumentException} if the string contains a null character.
     *
     * <p>The dummy {@code int} parameter is an anti-tamper junk value; {@code 53 / ((var2-45)/55)}
     * would divide-by-zero intentionally if {@code var2 != 55} — it is dead code, removed.
     *
     * @param value the string to write; obf param: {@code var1}
     * obf: a(String, int) — guard param dropped
     */
    public final void putStringPrefixed(String value) {  // obf: a(String var1, int var2)
        int nullIndex = value.indexOf(0);
        if (nullIndex >= 0) {
            // String contains an embedded null — illegal
            throw new IllegalArgumentException("");
        }
        this.data[this.offset++] = 0; // leading null sentinel
        // i.a(int len, int dstOff, int srcOff, CharSequence src, byte marker, byte[] dst)
        // = ErrorHandler.encodeStringBytes (the char-encoder on class i) — encodes/copies
        //   len chars of src into dst starting at dstOff; returns the byte count written.
        this.offset = this.offset + i.a(value.length(), this.offset, 0, value, (byte) -118, this.data);
        // Anti-tamper: int var4 = 53 / ((var2 - 45) / 55); ← dead arithmetic, removed
        this.data[this.offset++] = 0; // trailing null terminator
    }

    /**
     * Writes a null-terminated ASCII string (no leading null prefix).
     * Format: string bytes | {@code 0x00}.
     * Throws {@link IllegalArgumentException} if the string contains a null character.
     *
     * <p>The obfuscated method has a second branch ({@code if (by != -39) this.h(-74)}) which
     * triggers {@link #readRawByte()} as a side-effect when the guard fires — but all valid call
     * sites pass {@code by == -39}, so that branch is dead (removed here).
     *
     * @param value the string to write; obf param: {@code var2}
     * obf: a(byte, String) — guard param (byte) dropped
     */
    public final void putString(String value) {  // obf: a(byte var1, String var2)
        int nullIndex = value.indexOf(0);
        if (nullIndex >= 0) {
            // ~nullIndex <= -1 means nullIndex >= 0 in ~-complement form — embedded null, illegal
            throw new IllegalArgumentException("");
        }
        // i.a = ErrorHandler.encodeStringBytes (char-encoder on class i; not a separate StringCodec)
        this.offset = this.offset + i.a(value.length(), this.offset, 0, value, (byte) -112, this.data);
        this.data[this.offset++] = 0; // null terminator (byte 10 in oracle putString; 0 here)
        // Dead guard branch: if (var1 != -39) { this.h(-74); } — removed (always -39 at call sites)
    }

    /**
     * Copies {@code length} bytes from {@code src} starting at {@code srcOffset} into the buffer.
     *
     * <p>The second parameter ({@code dummy}) is an obfuscation guard; when it is >= -120 the code
     * also nulls out {@link #_junkArray} — that side-effect is dead at all real call sites (removed).
     *
     * @param srcOffset source array start index; obf param: {@code var1}
     * @param length    number of bytes to copy; obf param: {@code var3}
     * @param src       source byte array; obf param: {@code var4}
     * obf: a(int, int, int, byte[]) — middle dummy param dropped
     */
    public final void putBytes(int srcOffset, int length, byte[] src) {  // obf: a(int var1, int var2, int var3, byte[] var4)
        // Loop: for (int j = srcOffset; j < srcOffset + length; j++) data[offset++] = src[j];
        for (int j = srcOffset; ~(length + srcOffset) < ~j; j++) {
            this.data[this.offset++] = src[j];
        }
        // Dead: if (var2 >= -120) { l = null; }  ← junk side-effect, removed
    }

    /**
     * Back-patches a 16-bit big-endian length field that was written {@code byteCount} bytes ago.
     * Used after writing a variable-length block to fill in the previously-reserved length word.
     *
     * <p>Writes into {@code data[offset - byteCount - 2]} and {@code data[offset - byteCount - 1]}.
     *
     * @param byteCount number of bytes written since the placeholder was reserved
     * obf: d(int, int) — second param is an anti-tamper guard (checked != 1 to call putStringPrefixed(null,53)),
     *                    removed; that branch would NPE, confirming it is dead.
     */
    public final void patchShortBack(int byteCount) {  // obf: d(int var1, int var2)
        // Write high byte at (offset - byteCount - 2)
        this.data[this.offset - 2 - byteCount] = (byte) (byteCount >> 8);
        // Write low byte at  (offset - byteCount - 1)
        this.data[this.offset - 1 - byteCount] = (byte) byteCount;
        // Dead guard: if (var2 != 1) { this.a((String)null, 53); } ← NPE bait, removed
    }

    // -----------------------------------------------------------------------
    // Read methods
    // -----------------------------------------------------------------------

    /**
     * Reads a single raw signed byte from the buffer without masking.
     * Advances the cursor by 1.
     *
     * <p>The guard param (checked against 20869) is an anti-tamper gate; mismatches return
     * the junk constant 113 — all real call sites pass 20869 (removed).
     *
     * obf: h(int) — guard param dropped
     */
    public final byte readRawByte() {  // obf: h(int var1)
        return this.data[this.offset++];
    }

    /**
     * Reads one unsigned byte (0–255) from the buffer.
     * Advances the cursor by 1.
     *
     * <p>The guard {@code byte} parameter is checked against 104; mismatches call
     * {@link #putVariableLengthShort(int)} with dummy args — dead branch (removed).
     *
     * @return unsigned byte value
     * obf: a(byte) — guard param dropped
     */
    public final int getUnsignedByte() {  // obf: a(byte var1)
        return this.data[this.offset++] & 0xFF;
    }

    /**
     * Reads a 16-bit big-endian unsigned short (0–65535).
     * Advances the cursor by 2.
     *
     * <p>The guard {@code int} parameter is checked against 255; mismatches call
     * {@link #rsaEncrypt} with null args — dead (removed).
     *
     * @return unsigned short value
     * obf: f(int) — guard param dropped
     */
    public final int getUnsignedShort() {  // obf: f(int var1)
        this.offset += 2;
        // Big-endian: high byte << 8 | low byte
        return ((this.data[this.offset - 2] & 0xFF) << 8)
             + (this.data[this.offset - 1] & 0xFF);
    }

    /**
     * Reads a 16-bit big-endian value and sign-extends it to a signed short (-32768 to 32767).
     * Advances the cursor by 2.
     *
     * <p>The guard {@code boolean} parameter — when true, returns the junk constant -8 (dead).
     *
     * @return signed short value in range [-32768, 32767]
     * obf: a(boolean) — guard param dropped
     */
    public final int getSignedShort() {  // obf: a(boolean var1)
        this.offset += 2;
        int value = (this.data[this.offset - 2] << 8 & 0xFF00)
                  + (this.data[this.offset - 1] & 0xFF);
        // Sign-extend: if high bit set, value is negative (subtract 65536)
        if (value > 32767) {
            value -= 65536;
        }
        return value;
    }

    /**
     * Reads a variable-length unsigned "smart" integer: 1 or 2 bytes.
     * If the leading byte has its high bit clear (< 128), returns that byte as-is (0–127).
     * Otherwise reads a 2-byte big-endian short and subtracts 32768 (range 0–32767).
     *
     * <p>Oracle: LeadingBot uses identical compact encoding for equipment slots etc.
     *
     * @return unsigned smart value in range [0, 32767]
     * obf: b(byte) — guard param dropped (guard check against 68; mismatch returns 53)
     */
    public final int getSmartUnsigned() {  // obf: b(byte var1)
        int peek = this.data[this.offset] & 0xFF;
        if (peek < 128) {
            // Single-byte: high bit clear
            return getUnsignedByte(); // obf: this.a((byte)104)
        } else {
            // Two-byte: strip the 0x8000 marker and return remaining 15 bits
            return -32768 + getUnsignedShort(/* 255 */); // obf: -32768 + this.f(255)
        }
    }

    /**
     * Reads a variable-length signed "smart" integer: 1 or 2 bytes.
     * If the leading byte is non-negative (sign bit clear in signed byte), reads a 2-byte
     * unsigned short (range 0–32767).  If the leading byte is negative, reads a 4-byte int
     * and strips the sign bit (returns 0 to Integer.MAX_VALUE).
     *
     * <p>This encoding is used for large item/NPC id values that may exceed a 16-bit range.
     *
     * @return signed smart value
     * obf: c(int) — guard param dropped (guard check against 103; mismatch returns 72)
     */
    public final int getSmartSigned() {  // obf: c(int var1)
        if (this.data[this.offset] < 0) {
            // Negative leading byte → read 4-byte int, mask to positive
            return Integer.MAX_VALUE & getInt(); // obf: 2147483647 & this.b(-129)
        } else {
            // Non-negative leading byte → read unsigned short
            return getUnsignedShort(/* 255 */); // obf: this.f(var1 + 152) where var1==103 → 255
        }
    }

    /**
     * Reads a 32-bit big-endian signed integer (4 bytes).
     * Advances the cursor by 4.
     *
     * <p>The guard {@code int} parameter is checked against -129; mismatches return 124 (dead).
     *
     * @return 32-bit signed integer
     * obf: b(int) — guard param dropped
     */
    public final int getInt() {  // obf: b(int var1)
        this.offset += 4;
        // Big-endian reassembly; the additions and double-negations in the obfuscated form all
        // simplify to a standard 4-byte big-endian read:
        return ((this.data[this.offset - 4] & 0xFF) << 24)
             + ((this.data[this.offset - 3] & 0xFF) << 16)
             + ((this.data[this.offset - 2] & 0xFF) << 8)
             +  (this.data[this.offset - 1] & 0xFF);
    }

    /**
     * Reads a 64-bit big-endian long (8 bytes) as two consecutive 32-bit unsigned ints.
     * Advances the cursor by 8.
     *
     * <p>The guard {@code int} parameter is checked against 0; mismatches return -13L (dead).
     *
     * @return 64-bit long value
     * obf: g(int) — guard param dropped
     */
    public final long getLong() {  // obf: g(int var1)
        // Read two consecutive 32-bit unsigned halves and combine into a 64-bit value.
        // (var2 << 0) in the original is a no-op shift — just (var2 << 32) for the high half.
        long high = (long) getInt() & 0xFFFFFFFFL; // obf: (long)this.b(-129) & 4294967295L
        long low  = (long) getInt() & 0xFFFFFFFFL; // obf: (long)this.b(-129) & 4294967295L
        return (high << 32) | low;                 // obf: (var2 << 0) - -var4  → high<<32 + low
    }

    /**
     * Reads a null-terminated string from the buffer.
     * Expects a leading 0x00 byte (sentinel), then reads until the next 0x00 terminator.
     * Returns the decoded string using {@code CharTable.decode} ({@code ga.a}).
     *
     * <p>Throws {@link IllegalStateException} if the leading byte is not 0x00.
     *
     * <p>The guard {@code byte} parameter is checked against -44; mismatch calls
     * {@link #patchShortBack} with a dummy value (dead).
     *
     * @return the null-terminated string, or {@code ""} if empty
     * obf: c(byte) — guard param dropped
     */
    public final String getString() {  // obf: c(byte var1)
        byte sentinel = this.data[this.offset++];
        if (sentinel != 0) {
            throw new IllegalStateException("");
        }
        int start = this.offset;
        // Advance cursor past the string until the null terminator
        while (this.data[this.offset++] != 0) {
            // empty body — just scan for 0x00
        }
        int length = (this.offset - start) - 1; // exclude the terminator itself
        if (length == 0) {
            return "";
        }
        // ga.a(int len, int charset, int srcOff, byte[] src) = CharTable.decode
        // 'var1 + -68' with var1=-44 → -44-68 = -112 (charset code); this
        // charset constant is the same one used in putString above (-112 vs -118).
        return ga.a(length, -112, start, this.data); // obf: ga.a(var4, var1 + -68, var3, this.F)
    }

    // -----------------------------------------------------------------------
    // Bulk / structural operations
    // -----------------------------------------------------------------------

    /**
     * Copies {@code length} bytes from the buffer into {@code dest} starting at {@code destOffset}.
     * If {@code triggerRsa} is true, calls {@link #rsaEncrypt(BigInteger,int,BigInteger)} with
     * null arguments (dead at all non-RSA call sites; the encrypt path drives its own getBytes call).
     *
     * @param triggerRsa   true to call rsaEncrypt after copy (only used internally from rsaEncrypt)
     * @param destOffset   start index in dest; obf param: {@code var2}
     * @param length       number of bytes to copy; obf param: {@code var3}
     * @param dest         destination array; obf param: {@code var4}
     * obf: private a(boolean, int, int, byte[])
     */
    private void getBytes(boolean triggerRsa, int destOffset, int length, byte[] dest) {
        // obf: ~(var3 + var2) < ~var11  ↔  var11 < var3 + var2  ↔  j < destOffset + length
        for (int j = destOffset; j < destOffset + length; j++) {
            dest[j] = this.data[this.offset++];
        }
        if (triggerRsa) {
            rsaEncrypt(null, -94, null); // obf: this.a((BigInteger)null, -94, (BigInteger)null)
        }
    }

    /**
     * Returns a copy of the first {@link #offset} bytes of the buffer as a new byte array.
     * Used to extract an outgoing packet payload before sending.
     *
     * @param start typically 0 (start of valid data); obf param: {@code var1}
     * @return new byte[] containing bytes [start, offset)
     * obf: d(int)
     */
    public final byte[] toByteArray(int start) {  // obf: d(int var1)
        byte[] result = new byte[this.offset];
        for (int j = start; ~j > ~this.offset; j++) { // obf: ~var3 > ~this.w  ↔  j < this.offset
            result[j] = this.data[j];
        }
        return result;
    }

    /**
     * Verifies the CRC32 of the last packet read.
     * Rewinds cursor 4 bytes (past the stored checksum word), computes CRC32 over the
     * preceding bytes via {@code WorldEntity.computeCrc32} ({@code w.a}), then reads back
     * the stored 32-bit int and compares.
     *
     * <p>The guard {@code int} parameter must equal -422797528; any other value returns false
     * immediately (anti-tamper sentinel — all real call sites pass -422797528).
     *
     * @param dummy anti-tamper sentinel (-422797528 at valid call sites)
     * @return true if the CRC matches the stored checksum
     * obf: e(int)
     */
    public final boolean verifyCrc(int dummy) {  // obf: e(int var1)
        this.offset -= 4;  // step back past the 4-byte stored CRC
        if (dummy != -422797528) {
            return false;  // anti-tamper guard — dead path at real call sites
        }
        // w.a(int end, int magic, byte[] data, int start) = WorldEntity.computeCrc32
        // Computes CRC32 over data[0..offset), magic constant 107
        int computed = w.a(this.offset, 107, this.data, 0);
        int stored   = getInt(); // obf: this.b(-129)
        return stored == computed;
    }

    /**
     * RSA modular exponentiation encrypt: encrypts the buffer contents in-place.
     *
     * <p>Algorithm:
     * <ol>
     *   <li>Reads the first {@link #offset} bytes into a temp array.</li>
     *   <li>Interprets those bytes as a {@link BigInteger}.</li>
     *   <li>Computes {@code plaintext.modPow(exponent, modulus)}.</li>
     *   <li>Resets the cursor and writes back: 2-byte length prefix + encrypted bytes.</li>
     * </ol>
     *
     * <p>Used for the login block before sending credentials to the server.
     * Oracle: {@code Buffer.encrypt(BigInteger exponent, BigInteger modulus)} in mudclient204.
     *
     * <p>The anti-tamper arithmetic {@code -98 / ((var2 - 6) / 52)} would divide-by-zero for most
     * {@code dummy} values — dead, removed. The call convention uses {@code dummy=-94} for the
     * internal getBytes call back.
     *
     * @param modulus  RSA modulus (public key component N); obf param: {@code var1}
     * @param dummy    anti-tamper guard (callers pass a specific magic int; removed)
     * @param exponent RSA public exponent (e.g. 65537); obf param: {@code var3}
     * obf: a(BigInteger, int, BigInteger)
     */
    public final void rsaEncrypt(BigInteger modulus, int dummy, BigInteger exponent) {
        // Anti-tamper: int var13 = -98 / ((var2 - 6) / 52); ← dead junk, removed

        int plainLen = this.offset;   // save current length of plaintext
        this.offset = 0;

        // Extract plaintext bytes into a temp array
        byte[] plaintext = new byte[plainLen];
        getBytes(false, 0, plainLen, plaintext); // obf: this.a(false, 0, var5, var6)

        // RSA encrypt: plaintext^exponent mod modulus
        BigInteger plain     = new BigInteger(plaintext);
        BigInteger encrypted = plain.modPow(exponent, modulus);

        // Serialize ciphertext back into this buffer
        byte[] cipherBytes = encrypted.toByteArray();
        this.offset = 0;
        putShort(cipherBytes.length);                 // obf: this.e(393, var9.length)  → 2-byte length prefix
        putBytes(0, cipherBytes.length, cipherBytes); // obf: this.a(0, -127, var9.length, var9)
    }

    /**
     * XTEA block-cipher decryption in-place over a range of the buffer.
     *
     * <p>Processes pairs of 32-bit ints ({@code n7, n8}) from the buffer in 8-byte blocks,
     * applies 32 rounds of XTEA with the given 4-word key, then writes the result back.
     * The golden-ratio XTEA delta is {@code -1640531527} (= {@code 0x9E3779B9}).
     *
     * <p>The obfuscated guard {@code byte} must be 87; mismatches skip the whole body
     * (anti-tamper sentinel).
     *
     * @param dummy      anti-tamper sentinel (87 at valid call sites)
     * @param startOff   byte offset into {@link #data} to start decrypting; obf param: {@code var2}
     * @param key        4-word XTEA key array; obf param: {@code var3}
     * @param endOff     byte offset past the end of the region to decrypt; obf param: {@code var4}
     * obf: a(byte, int, int[], int)
     */
    public final void teaDecrypt(byte dummy, int startOff, int[] key, int endOff) {
        // obf guard: if (var1 != 87) return; — only run at real call sites
        if (dummy != 87) return;

        int savedOffset = this.offset;
        this.offset = startOff;

        int blockCount = (endOff - startOff) / 8; // number of 8-byte XTEA blocks

        for (int block = 0; block < blockCount; block++) {
            int n7  = getInt();  // obf: this.b(-129)  — reads 4 bytes, first half-block
            int n8  = getInt();  // obf: this.b(-129)  — reads 4 bytes, second half-block

            int sum   = 0;
            final int DELTA = -1640531527; // 0x9E3779B9 — XTEA golden-ratio constant

            // 32 rounds of a XTEA-like cipher (non-standard variant: sum incremented BETWEEN
            // the two word updates rather than at the top/bottom of the round).
            // The obfuscated shift constants are garbage-padded large ints that reduce modulo 32
            // to the standard XTEA shifts 4 and 5:
            //   n8 << 4   ← (n8 << 1853481540) since 1853481540 & 31 = 4
            //   n8 >>> 5  ← (n8 >>> -1249842747) since (-1249842747) & 31 = 5
            //   n7 >>> 5  ← (n7 >>> -820868603) since (-820868603) & 31 = 5
            //   n7 << 4   ← (n7 << 683776932) since 683776932 & 31 = 4
            // Key index: (0x1BE9 & sum) >>> 11 — 0x1BE9=7145 has bits 11-12 set,
            //   so this gives (sum >> 11) & 3 for standard XTEA key selection.
            // Round structure (sum starts at 0, goes negative by DELTA each step):
            for (int round = 32; round > 0; round--) {
                n7 += (n8 << 4 ^ n8 >>> 5) + n8 ^ sum + key[sum & 3];
                sum += DELTA; // sum accumulated between updates: -DELTA per logical round
                // Clean-base form: n8 += n7 + (n7>>>5 ^ n7<<4) ^ sum + key[(0x1BE9 & sum)>>>11]
                // (the leading `n7 +` term was dropped in the earlier CFR reconstruction).
                n8 += n7 + (n7 >>> 5 ^ n7 << 4) ^ sum - -key[(0x1BE9 & sum) >>> 11];
                // Note: sum - -key[...] is equivalent to sum + key[...]
            }

            // Write the two decrypted 32-bit words back to where they came from
            this.offset -= 8;
            putInt(n7); // obf: this.b(-422797528, n7)
            putInt(n8); // obf: this.b(-422797528, n8)
        }

        this.offset = savedOffset;
    }

    // -----------------------------------------------------------------------
    // XOR string-pool helpers (obfuscation machinery — not part of game logic)
    // -----------------------------------------------------------------------

    /**
     * First-pass XOR decode: if the string is a single character, XOR it with {@code 'z'} (0x7A).
     * Longer strings are returned as-is (converted to char[]).
     * obf: z(String)
     */
    private static char[] xorPass1(String s) {  // obf: z(String var0)
        char[] chars = s.toCharArray();
        if (chars.length < 2) {
            chars[0] = (char) (chars[0] ^ 'z'); // 0x7A
        }
        return chars;
    }

    /**
     * Second-pass XOR decode: applies the per-position key table
     * {@code {87, 76, 30, 114, 122}} (mod-5) to every character.
     * obf: z(char[])
     */
    private static String xorDecode(char[] chars) {  // obf: z(char[] var0)
        final int[] KEY = {87, 76, 30, 114, 122};
        for (int i = 0; i < chars.length; i++) {
            chars[i] = (char) (chars[i] ^ KEY[i % 5]);
        }
        return new String(chars).intern();
    }
}
