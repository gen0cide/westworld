package client.net;

/**
 * TextEncoder — obf: h
 *
 * Converts a Java {@link CharSequence} (Unicode) into a raw byte array using
 * the Windows-1252 (CP1252) character encoding.  This is the only encoding
 * the RSC rev-233/235 client speaks: ASCII and Latin-1 characters (U+0001–
 * U+007F and U+00A0–U+00FF) pass through as-is via a direct cast to byte;
 * the 27 "special" code-points in the CP1252 0x80–0x9F extension block (the
 * range that ISO-8859-1 leaves undefined) are mapped to their corresponding
 * Windows-1252 byte values; and every other Unicode character is replaced
 * with '?' (0x3F).
 *
 * The single public entry-point is {@link #encode(CharSequence, byte)},
 * called by {@link client.net.StringCodec} (obf: u) to serialise chat
 * strings into packet buffers, and by {@link client.net.DownloadWorker}
 * (obf: jb) to encode HTTP request lines written to a raw socket.
 *
 * The five static fields {@code callCounter}, {@code scratchIntArray2},
 * {@code scratchIntArray1}, {@code registrySize}, and {@code scratchStrings}
 * are NOT used inside this class: they are scratch slots whose storage is
 * allocated by the loader ({@link client.net.SocketFactory}, obf: m) and
 * consumed by other classes.  They are preserved here because the JVM field
 * layout must match the original class file.
 *
 * Obfuscation applied to this class (all stripped):
 *   - Opaque predicate: {@code boolean bl = client.vh;} (always false),
 *     gating dead {@code if (bl) break;} / {@code while (!bl)} paths.
 *   - Profiling counter: {@code ++h.d;} incremented on every encode call.
 *   - Exception wrapper: {@code catch (RuntimeException e) { throw i.a(e, "…"); }}
 *     unwrapped to bare body.
 *   - Anti-tamper junk parameter: {@code byte _ignored} is divided into a
 *     constant ({@code -91 / ((_ignored + 26) / 53)}) whose result is stored
 *     but never read.  The division is designed so the caller must pass a
 *     value where {@code (_ignored + 26) / 53 != 0} (i.e. not in [-26, 26]);
 *     known callers use 30 and -104.
 *   - XOR string pool {@code z[]}: three encoded literals decoded at class-init
 *     time; used only in the exception-wrapper error message.
 */
final class TextEncoder { // obf: h

    // -------------------------------------------------------------------------
    // Scratch fields — allocated here, used by other classes via SocketFactory
    // (obf: m).  Do NOT reorder; field layout must match the original .class.
    // -------------------------------------------------------------------------

    /**
     * Scratch integer-array slot #2, sized by the loader.
     * obf: h.c (int[])
     * Used as temporary storage by SocketFactory (obf: m) during init.
     */
    static int[] scratchIntArray2; // obf: c

    /**
     * Profiling / call-count field incremented on every {@link #encode} call.
     * Dead profiling counter; never read by game logic.
     * obf: h.d (int)
     */
    static int encodeCallCount; // obf: d

    /**
     * Scratch integer-array slot #1, sized by the loader.
     * obf: h.b (int[])
     * Used as temporary storage by SocketFactory (obf: m) during init.
     */
    static int[] scratchIntArray1; // obf: b

    /**
     * Registry / size value set by the loader; used to size arrays in several
     * sibling classes (StringCodec/obf:u, ChatCipher/obf:v, etc.).
     * obf: h.a (int)
     */
    static int registrySize; // obf: a

    /**
     * Scratch String-array slot, sized by the loader ({@code t.g} elements).
     * obf: h.e (String[])
     * Used as temporary storage by SocketFactory (obf: m) during init.
     */
    static String[] scratchStrings; // obf: e

    // -------------------------------------------------------------------------
    // XOR-encoded string pool (decoded at class-init time)
    // Keys: first z pass uses per-char key 'H' (0x48) for length-1 strings;
    //       second z pass uses mod-5 byte table {5, 42, 10, 117, 72}.
    //
    //   z[0] decoded = "{...}"   — non-null argument label in error message
    //   z[1] decoded = "null"    — null argument label in error message
    //   z[2] decoded = "h.A("   — method-signature prefix in error message
    //
    // Only used inside the (now-stripped) exception-wrapper catch block.
    // -------------------------------------------------------------------------
    private static final String[] OBFSTRINGS; // obf: z
    static {
        // z[0] = z(z("~$[5"))  -> "{...}"
        // z[1] = z(z("k_f"))   -> "null"
        // z[2] = z(z("mK]"))   -> "h.A("
        OBFSTRINGS = new String[]{
            decodeXorChars(decodeXorString("~$[5")),  // "{...}"
            decodeXorChars(decodeXorString("k_f")),  // "null"
            decodeXorChars(decodeXorString("mK]"))   // "h.A("
        };
    }

    // -------------------------------------------------------------------------
    // Public API
    // -------------------------------------------------------------------------

    /**
     * Encode a Unicode string to a Windows-1252 byte array.
     *
     * Each character is mapped as follows:
     * <ol>
     *   <li>U+0001–U+007F (printable ASCII): direct byte cast (pass-through).
     *   <li>U+00A0–U+00FF (Latin-1 supplement): direct byte cast (pass-through).
     *   <li>The 27 Unicode code-points that correspond to the Windows-1252
     *       extension bytes 0x80–0x9F (the range ISO-8859-1 leaves undefined):
     *       mapped to the appropriate CP1252 byte via an explicit switch table.
     *   <li>Anything else: replaced by '?' (0x3F).
     * </ol>
     *
     * The output array is always the same length as the input {@code CharSequence}.
     * No BOM, no multi-byte sequences: this is strictly a one-char-per-byte
     * substitution encoding.
     *
     * @param text     the Unicode text to encode; must not be null
     * @param _ignored anti-tamper junk parameter (see class doc); callers pass
     *                 30 ({@code StringCodec}) or -104 ({@code DownloadWorker}).
     *                 The value must satisfy {@code (_ignored + 26) / 53 != 0}
     *                 (integer division) to avoid an ArithmeticException in the
     *                 now-dead anti-tamper guard.
     * @return a newly-allocated {@code byte[]} of the same length as {@code text}
     *
     * obf: h.a(CharSequence, byte) : byte[]
     */
    static final byte[] encode(CharSequence text, byte _ignored) { // obf: a
        // Dead anti-tamper guard — result is computed but never used:
        //   int _junk = -91 / ((_ignored - -26) / 53);
        // The division was designed to always succeed for the callers' values.
        // Stripped per reverse-engineering conventions.

        // Dead profiling counter — kept as a field increment to preserve semantics,
        // but the counter itself is never read by game logic:
        ++encodeCallCount; // obf: ++h.d

        int len = text.length();
        byte[] out = new byte[len];

        for (int i = 0; i < len; i++) {
            char c = text.charAt(i);

            // Fast path: ASCII (U+0001–U+007F) — direct byte cast.
            // The original bytecode checks (c > 0 && c < 128) for this arm.
            if (c > 0 && c < 128) {
                out[i] = (byte) c;
                continue;
            }

            // Fast path: Latin-1 supplement (U+00A0–U+00FF) — direct byte cast.
            // Bytecode condition: (~c >= -161 && c <= 0xFF)
            //   ~c >= -161  ⟺  -(c+1) >= -161  ⟺  c <= 160  ← wait, that's ≤ 0xA0
            // More precisely: -161 >= ~c  ⟺  ~c <= -161  ⟺  c >= 160 (0xA0).
            // Combined with c <= 0xFF: matches the Latin-1 supplement range.
            if (c >= 0xA0 && c <= 0xFF) {
                out[i] = (byte) c;
                continue;
            }

            // CP1252 extension block: the 27 Unicode code-points that occupy the
            // 0x80–0x9F byte range in Windows-1252.  ISO-8859-1 leaves this range
            // undefined; Windows-1252 assigns typographic characters here.
            // The original bytecode implements these as a long chain of if/goto
            // comparisons (some using XOR with -1 instead of != for obfuscation).
            switch (c) {
                // 0x80 ─ Euro sign
                case '€': out[i] = (byte) 0x80; break; // € -> -128

                // 0x82 ─ Single low-9 quotation mark
                case '‚': out[i] = (byte) 0x82; break; // ‚ -> -126

                // 0x83 ─ Latin small letter f with hook (florin sign)
                case 'ƒ': out[i] = (byte) 0x83; break; // ƒ -> -125

                // 0x84 ─ Double low-9 quotation mark
                // Bytecode: ~c == -8223  ⟺  c == 8222 == 0x201E
                case '„': out[i] = (byte) 0x84; break; // „ -> -124

                // 0x85 ─ Horizontal ellipsis
                case '…': out[i] = (byte) 0x85; break; // … -> -123

                // 0x86 ─ Dagger
                // Bytecode: ~c == -8225  ⟺  c == 8224 == 0x2020
                case '†': out[i] = (byte) 0x86; break; // † -> -122

                // 0x87 ─ Double dagger
                // Bytecode: -8226 == ~c  ⟺  c == 8225 == 0x2021
                case '‡': out[i] = (byte) 0x87; break; // ‡ -> -121

                // 0x88 ─ Modifier letter circumflex accent
                case 'ˆ': out[i] = (byte) 0x88; break; // ˆ -> -120

                // 0x89 ─ Per mille sign
                // Bytecode: -8241 != ~c  (negated); this is the fall-through arm
                // that sets -119 when ~c == -8241, i.e. c == 8240 == 0x2030
                case '‰': out[i] = (byte) 0x89; break; // ‰ -> -119

                // 0x8A ─ Latin capital letter S with caron
                case 'Š': out[i] = (byte) 0x8A; break; // Š -> -118

                // 0x8B ─ Single left-pointing angle quotation mark
                case '‹': out[i] = (byte) 0x8B; break; // ‹ -> -117

                // 0x8C ─ Latin capital ligature OE
                case 'Œ': out[i] = (byte) 0x8C; break; // Œ -> -116

                // 0x8E ─ Latin capital letter Z with caron
                // (0x8D has no CP1252 assignment; 0x8F is also unassigned.)
                case 'Ž': out[i] = (byte) 0x8E; break; // Ž -> -114

                // 0x91 ─ Left single quotation mark
                case '‘': out[i] = (byte) 0x91; break; // ' -> -111

                // 0x92 ─ Right single quotation mark
                // Bytecode: -8218 == ~c  ⟺  c == 8217 == 0x2019
                case '’': out[i] = (byte) 0x92; break; // ' -> -110

                // 0x93 ─ Left double quotation mark
                case '“': out[i] = (byte) 0x93; break; // " -> -109

                // 0x94 ─ Right double quotation mark
                case '”': out[i] = (byte) 0x94; break; // " -> -108

                // 0x95 ─ Bullet
                case '•': out[i] = (byte) 0x95; break; // • -> -107

                // 0x96 ─ En dash
                case '–': out[i] = (byte) 0x96; break; // – -> -106

                // 0x97 ─ Em dash
                case '—': out[i] = (byte) 0x97; break; // — -> -105

                // 0x98 ─ Small tilde
                // Bytecode: -733 == ~c  ⟺  c == 732 == 0x02DC
                case '˜': out[i] = (byte) 0x98; break; // ˜ -> -104

                // 0x99 ─ Trade mark sign
                // Bytecode: ~c == -8483  ⟺  c == 8482 == 0x2122
                case '™': out[i] = (byte) 0x99; break; // ™ -> -103

                // 0x9A ─ Latin small letter s with caron
                // Bytecode: -354 == ~c  ⟺  c == 353 == 0x0161
                case 'š': out[i] = (byte) 0x9A; break; // š -> -102

                // 0x9B ─ Single right-pointing angle quotation mark
                case '›': out[i] = (byte) 0x9B; break; // › -> -101

                // 0x9C ─ Latin small ligature oe
                // Bytecode: ~c == -340  ⟺  c == 339 == 0x0153
                case 'œ': out[i] = (byte) 0x9C; break; // œ -> -100

                // 0x9E ─ Latin small letter z with caron
                // (0x9D has no CP1252 assignment.)
                case 'ž': out[i] = (byte) 0x9E; break; // ž -> -98

                // 0x9F ─ Latin capital letter Y with diaeresis
                // Bytecode: ~c != -377 branch falls to '?' (63); arm at -377 sets -97
                //   ~c == -377  ⟺  c == 376 == 0x0178
                case 'Ÿ': out[i] = (byte) 0x9F; break; // Ÿ -> -97

                // Default: character has no Windows-1252 representation.
                // Replaced with '?' (0x3F = 63).
                default: out[i] = (byte) '?'; break;
            }
        }

        return out;
    }

    // -------------------------------------------------------------------------
    // XOR string-pool helpers (class-private; decode the OBFSTRINGS array)
    // -------------------------------------------------------------------------

    /**
     * First decoding pass: converts an encoded {@code String} to a {@code char[]}
     * without XOR transformation (only single-character strings would be XOR'd
     * with 'H' = 0x48, but none of the three pool entries are length-1).
     *
     * obf: z(String) : char[]
     */
    private static char[] decodeXorString(String encoded) { // obf: z(String)
        char[] chars = encoded.toCharArray();
        if (chars.length < 2) {
            // Single-character string: XOR the sole character with 'H' (0x48).
            chars[0] = (char) (chars[0] ^ 'H');
        }
        return chars;
    }

    /**
     * Second decoding pass: XORs each character of the intermediate {@code char[]}
     * with a position-dependent key from the 5-element table
     * {@code {5, 42, 10, 117, 72}} (positions 0–4 mod 5), then interns the result.
     *
     * obf: z(char[]) : String
     */
    private static String decodeXorChars(char[] chars) { // obf: z(char[])
        // Mod-5 XOR key table, identical across all three encoded strings:
        //   pos % 5 == 0 -> key  5
        //   pos % 5 == 1 -> key 42
        //   pos % 5 == 2 -> key 10
        //   pos % 5 == 3 -> key 117
        //   pos % 5 == 4 -> key 72
        final int[] KEY_TABLE = {5, 42, 10, 117, 72};
        for (int pos = 0; pos < chars.length; pos++) {
            chars[pos] = (char) (chars[pos] ^ KEY_TABLE[pos % 5]);
        }
        return new String(chars).intern();
    }
}
