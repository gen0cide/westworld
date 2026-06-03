package client.net;

import client.ui.Panel;        // qa — owns the 256-entry byte translation table `l`
import client.net.StreamBase;  // ib — provides the unsigned-byte mask helper a(int,int)
import client.net.Packet;      // b  — owns the shared static decode buffer `v`
import client.util.IntHolder;  // ka — owns the shared static read cursor `b`

/**
 * ChatCipher (obfuscated: {@code v}) — package {@code client.net}.
 *
 * <p>A tiny stateless byte-(de)cipher / character-translation helper used while
 * decoding incoming text (chat messages, private messages, names, etc.). Its one
 * real job is {@link #translate(byte[], int, int)}: it copies a slice of raw
 * network bytes through a 256-entry translation table ({@link Panel#TRANSLATION_TABLE}
 * — obf {@code qa.l}), which remaps the wire byte values into the client's internal
 * character set. The remaining static methods are small utilities that operate on
 * the shared inbound buffer ({@link Packet} / {@link IntHolder}) and a character
 * classification predicate equivalent to the oracle's {@code WordFilter.isLetterOrDigit}.
 *
 * <p>The class is also (ab)used as a trivial 4-field record holder: the
 * {@link #ChatCipher(String, String, String, int)} constructor only stashes an int
 * into {@link #value}. This double-duty (cipher helper + data holder) is typical of
 * the obfuscator collapsing unrelated small classes together.
 *
 * <p>Cross-references resolved during deobfuscation:
 * <ul>
 *   <li>{@code qa.l}  → {@link Panel#TRANSLATION_TABLE} — {@code static byte[]} loaded from a cache resource.</li>
 *   <li>{@code ib.a(x,255)} → {@link StreamBase#mask(int,int)} — just {@code x & 255} (unsigned-byte).</li>
 *   <li>{@code b.v}   → {@link Packet#SHARED_BUFFER} — shared static {@code byte[]} decode buffer.</li>
 *   <li>{@code ka.b}  → {@link IntHolder#cursor} — shared static read cursor into that buffer.</li>
 * </ul>
 *
 * <p>Obfuscation stripped: the {@code client.vh} opaque-predicate dead branches,
 * per-method profiling counters, the {@code i.a(e, "sig(...)")} exception-context
 * wrappers, and the anti-tamper sentinel-parameter guards (see notes on each method).
 */
public final class ChatCipher {

   // ---------------------------------------------------------------------------
   // Profiling counters (obfuscator instrumentation). Each method bumped one of
   // these on entry; the values are never read. Kept named, otherwise inert.
   // ---------------------------------------------------------------------------
   /** Invocation counter for {@link #translate(byte[], int, int)} (obf {@code c}). */
   public static int translateCallCount;
   /** Invocation counter for {@link #readNextBufferByte()} (obf {@code d}). */
   public static int readByteCallCount;
   /** Invocation counter for {@link #toString()} (obf {@code f}). */
   public static int toStringCallCount;
   /** Invocation counter for {@link #isLetterOrDigit(char)} (obf {@code b}). */
   public static int isLetterOrDigitCallCount;

   // ---------------------------------------------------------------------------
   // Inert scratch fields the obfuscator left attached to this class. None are
   // read by any method here; they are placeholders / shared spill slots.
   // ---------------------------------------------------------------------------
   /** Unused scratch table cleared by the anti-tamper guard in {@link #translate} (obf {@code a}). */
   public static int[] scratchA;
   /** Unused static scratch (obf {@code h}). */
   public static int unusedH;
   /** Unused static scratch array (obf {@code g}). */
   public static int[] unusedG;
   /** Unused static scratch array (obf {@code e}). */
   public static int[] unusedE;

   /** The single payload of the record-holder usage of this class (obf {@code i}). */
   public int value;

   /**
    * Obfuscated error-context signature strings, decoded at class init.
    * Decoded contents (in order): {@code "null"}, {@code "v.<init>("}, {@code "{...}"},
    * {@code "v.C("}, {@code "v.A("}, {@code "v.toString()"}, {@code "v.B("}.
    * Only used to build the (now-removed) exception-wrapper messages; retained so the
    * original layout is documented. (obf {@code z})
    */
   private static final String[] SIG = new String[]{
      decode(scramble("\"tv%")),
      decode(scramble(":/& /%u$a")),
      decode(scramble("7/4g<")),
      decode(scramble(":/Ya")),
      decode(scramble(":/[a")),
      decode(scramble(":/n&8ss'&d(")),
      decode(scramble(":/Xa"))
   };

   /**
    * Translates {@code length} raw bytes starting at {@code srcOffset} through the
    * client's 256-entry byte-translation table, returning a fresh {@code byte[length]}.
    *
    * <p>For each input byte {@code b}, the result is {@code TRANSLATION_TABLE[b & 0xFF]}
    * — i.e. the unsigned byte value is used as an index into {@link Panel#TRANSLATION_TABLE},
    * remapping wire byte values to the client's internal character codes. (obf {@code a(byte[],int,int,int)})
    *
    * @param src       source byte array (raw network bytes)
    * @param length    number of bytes to translate
    * @param srcOffset starting index in {@code src}
    *
    * <p>Removed: the {@code sentinel} 3rd parameter (original {@code var2}); the only
    * thing it did was an anti-tamper guard {@code if (sentinel != -98) scratchA = null;}
    * — a no-op clear of an unused field. Dropped along with the guard.
    */
   public static final byte[] translate(byte[] src, int length, int srcOffset) {
      // Anti-tamper no-op (clearing an unused scratch field) removed.

      byte[] out = new byte[length];
      for (int i = 0; i < length; i++) {
         // ib.a(x, 255) == (x & 0xFF): treat the signed source byte as an unsigned
         // index, then map it through the client's byte-translation table.
         // obf ib.a(int,int) is declared StreamBase.bitwiseAnd (was StreamBase.mask);
         // obf qa.l is declared Panel.fontGlyphData (was the non-existent Panel.TRANSLATION_TABLE).
         out[i] = Panel.fontGlyphData[StreamBase.bitwiseAnd(src[srcOffset + i], 255)];
      }
      return out;
   }

   /**
    * Reads and returns the next byte (as an unsigned 0–255 value) from the shared
    * static decode buffer {@link Packet#SHARED_BUFFER}, advancing the shared read
    * cursor {@link IntHolder#cursor} by one. (obf {@code a(int)})
    *
    * <p>Removed: the {@code sentinel} parameter (original {@code var0}) and its
    * anti-tamper guard {@code if (sentinel != -30504) readNextBufferByte(113);}
    * — a dead recursive self-call (obf {@code a(113)}) on a never-taken branch. Dropped.
    *
    * @return the next buffer byte, masked to 0–255
    */
   public static final int readNextBufferByte() {
      // obf b.v = Packet.legacyBytes (was Packet.SHARED_BUFFER); ka.b = IntHolder.bufferOffset (was IntHolder.cursor).
      int value = Packet.legacyBytes[IntHolder.bufferOffset] & 0xFF;
      IntHolder.bufferOffset++;
      return value;
   }

   /**
    * Always throws {@link IllegalStateException}: this class is never meant to be
    * stringified, so the obfuscator made {@code toString()} a tamper trap. (obf {@code toString()})
    */
   @Override
   public final String toString() {
      throw new IllegalStateException();
   }

   /**
    * Returns whether {@code c} is an ASCII alphanumeric character — a digit
    * {@code '0'..'9'}, an uppercase letter {@code 'A'..'Z'}, or a lowercase letter
    * {@code 'a'..'z'}. Equivalent to the oracle's {@code WordFilter.isLetterOrDigit}.
    * (obf {@code a(char,int)})
    *
    * <p>The obfuscated bytecode expressed the range tests with bitwise-complement
    * comparisons (e.g. {@code ~c >= -58} ⇔ {@code c <= 57}); they are restored here
    * to the natural inclusive ranges.
    *
    * <p>Removed: the {@code sentinel} 2nd parameter (original {@code param1}) and its
    * anti-tamper guard {@code if (sentinel <= 111) translate(null, 51, 127, 27);}
    * — a dead self-call. Dropped.
    *
    * @param c the character to classify
    * @return {@code true} iff {@code c} is {@code [0-9A-Za-z]}
    */
   public static final boolean isLetterOrDigit(char c) {
      if (c >= '0' && c <= '9') {   // ~c in [-58,-49]  →  digits
         return true;
      }
      if (c >= 'A' && c <= 'Z') {   // c>='A' && ~c>=-91 →  uppercase
         return true;
      }
      if (c < 'a') {                // ~c > -98  →  below 'a', not alphanumeric
         return false;
      }
      if (c > 'z') {                // ~c < -123 →  above 'z', not alphanumeric
         return false;
      }
      return true;                  // 'a'..'z' →  lowercase
   }

   /**
    * Record-holder constructor: the obfuscator reuses this class as a 4-field
    * tuple, but only the int is ever stored. The three string parameters are
    * accepted and ignored (they appeared only in the now-removed exception-context
    * message). (obf {@code v(String,String,String,int)})
    *
    * @param unusedA ignored string slot
    * @param unusedB ignored string slot
    * @param unusedC ignored string slot
    * @param value   the int payload, stored in {@link #value}
    */
   public ChatCipher(String unusedA, String unusedB, String unusedC, int value) {
      this.value = value;
   }

   /**
    * First stage of the two-stage string-constant deobfuscator. Returns the
    * character array of {@code s}; if the string has fewer than 2 chars, its single
    * char is XOR-ed with {@code 'A'} (0x41). (obf {@code z(String)})
    *
    * @param s the scrambled literal
    * @return its mutable char buffer, lightly pre-mangled for short strings
    */
   private static char[] scramble(String s) {
      char[] chars = s.toCharArray();
      if (chars.length < 2) {
         chars[0] = (char) (chars[0] ^ 'A');
      }
      return chars;
   }

   /**
    * Second stage of the string-constant deobfuscator: XORs each char by a
    * position-dependent key cycling through {@code {76, 1, 26, 73, 65}} (keyed by
    * {@code index % 5}), then interns the result. Decodes the {@link #SIG} table.
    * (obf {@code z(char[])})
    *
    * @param chars the char buffer from {@link #scramble(String)} (mutated in place)
    * @return the decoded, interned string
    */
   private static String decode(char[] chars) {
      for (int i = 0; i < chars.length; i++) {
         byte key;
         switch (i % 5) {
            case 0:  key = 76; break;
            case 1:  key = 1;  break;
            case 2:  key = 26; break;
            case 3:  key = 73; break;
            default: key = 65;
         }
         chars[i] = (char) (chars[i] ^ key);
      }
      return new String(chars).intern();
   }
}
