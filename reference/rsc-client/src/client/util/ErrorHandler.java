package client.util;

import client.net.ChatCipher;       // obf v
import client.scene.Surface;        // obf ua
import client.net.ClientStream;     // obf da
import client.ui.CharTable;         // obf ga
import client.world.GameCharacter;  // obf ta
import client.audio.AudioMixer;     // obf eb
import client.net.ProxySocketFactory;// obf gb

/**
 * ErrorHandler  (obfuscated class {@code i}, package {@code client.util}).
 *
 * <p>This is the central error/diagnostics utility for the obfuscated J++ RuneScape Classic
 * client (rev ~233-235). It serves three loosely-related purposes that the obfuscator bundled
 * into one class:</p>
 *
 * <ol>
 *   <li><b>Exception wrapping</b> — {@link #wrap(Throwable, String)} (obf {@code a(Throwable,String)})
 *       is the sink for the per-method {@code try { BODY } catch (RuntimeException e) { throw i.a(e, "sig(args)"); }}
 *       pattern that appears in nearly every method of the build. It produces / augments a
 *       {@link ClientRuntimeException} (obf {@code la}) carrying a human-readable "call signature"
 *       breadcrumb. Repeated wraps append further breadcrumbs to the same exception, building a
 *       textual call-stack of where the failure propagated.</li>
 *
 *   <li><b>Windows-1252 string encoding</b> — {@link #encodeCp1252(int, int, int, CharSequence, byte[])}
 *       (obf {@code B}, exposed as {@code a(int,int,int,CharSequence,byte,byte[])}) converts a
 *       {@code CharSequence} into Windows-1252 (CP1252) bytes written into a destination array.
 *       Used by {@code Buffer.putString} (obf {@code tb}) when serialising chat / text fields onto
 *       the wire.</li>
 *
 *   <li><b>ChatCipher registry</b> — {@link #chatCiphers()} (obf {@code C}, exposed as {@code a(int)})
 *       returns the fixed set of {@link ChatCipher} (obf {@code v}) instances scattered across the
 *       engine, so look-up code (e.g. {@code NameTable} / obf {@code ub}) can resolve a cipher by id.</li>
 * </ol>
 *
 * <p>It also holds two static tables: a CRC-64/ECMA-182 lookup table ({@link #CRC64_TABLE}) and a
 * custom Base64 alphabet ({@link #BASE64_ALPHABET}) consumed by {@code GameShell} (obf {@code e}).</p>
 *
 * <p>All obfuscation artifacts have been removed: the opaque {@code boolean bl = client.vh}
 * predicate and its dead {@code if (bl) ...} branches, the per-method profiling counters, the
 * anti-tamper dummy-parameter guards, and the self-wrapping try/catch blocks. The encrypted string
 * pool {@code z[]} has been decoded inline to its plaintext (method-signature breadcrumbs).</p>
 */
public final class ErrorHandler {

   // -----------------------------------------------------------------------------------------
   // Profiling counters (obf b, e, c, d). Each method bumped one of these on entry as part of the
   // obfuscation; they are otherwise unused. Retained as fields for structural fidelity only.
   // -----------------------------------------------------------------------------------------
   /** Obf {@code b}: invocation counter for {@link #encodeCp1252}. Unused diagnostic. */
   static int encodeCp1252CallCount;
   /** Obf {@code e}: invocation counter for {@link #toString()}. Unused diagnostic. */
   static int toStringCallCount;
   /** Obf {@code c}: invocation counter for {@link #wrap}. Unused diagnostic. */
   static int wrapCallCount;
   /** Obf {@code d}: invocation counter for {@link #chatCiphers}. Unused diagnostic. */
   static int chatCiphersCallCount;

   /** Obf {@code a}: per-instance scratch int set by the constructor. Purpose unclear; unread. */
   int instanceValue;

   /** Obf {@code g}: declared but never populated/read in this class. Unused scratch array. */
   static int[] unusedIntTable;

   /**
    * Obf {@code h}: CRC-64 (ECMA-182, polynomial {@code 0xC96C5795D7870F42}) lookup table,
    * one entry per possible byte value. Built in the static initialiser.
    */
   private static long[] CRC64_TABLE = new long[256];

   /**
    * Obf {@code f}: 64-character custom Base64 alphabet ({@code A-Z a-z 0-9 ! "}) followed by a
    * 32-byte trailing data block. The first 64 characters are used by {@code GameShell} (obf
    * {@code e}) to encode/decode Base64 (it scans this string with {@code charAt} to map a
    * character to its 6-bit index). Purpose of the trailing 32 bytes is unconfirmed.
    */
   static String BASE64_ALPHABET;

   /**
    * Obf {@code z[]}: decoded pool of method-signature breadcrumb fragments used to build the
    * context strings passed to {@link #wrap}. Kept as named constants below for readability;
    * the original packed these into an XOR-obfuscated string array.
    */
   private static final String SIG_CONSTRUCTOR = "i.<init>(";  // z[0]
   private static final String ARG_NULL        = "null";       // z[1]: rendered for a null arg
   private static final String ARG_OBJECT      = "{...}";      // z[2]: placeholder for a non-null object arg
   private static final String SIG_CHAT_CIPHERS = "i.C(";      // z[3]: signature of chatCiphers()
   private static final String SIG_ENCODE_CP1252 = "i.B(";     // z[4]: signature of encodeCp1252()
   private static final String SIG_TO_STRING   = "i.toString()"; // z[5]

   /**
    * Deliberately always throws — used by the obfuscator as a tamper/trip wire. Logs through the
    * standard {@link #wrap} sink so the breadcrumb "i.toString()" is recorded.
    */
   @Override
   public final String toString() {
      throw wrap(new IllegalStateException(), SIG_TO_STRING);
   }

   /**
    * Obf {@code B} (exposed as {@code a(int,int,int,CharSequence,byte,byte[])}):
    * encode {@code length} characters of {@code text} (starting at {@code srcStart}) into
    * Windows-1252 bytes, writing them into {@code dest} starting at {@code destStart}.
    *
    * <p>Characters {@code 0x01-0x7F} (ASCII) and {@code 0xA0-0xFF} (Latin-1 high range) are stored
    * directly as their code-unit value. Characters in the {@code 0x80-0x9F} band are mapped to their
    * Windows-1252 byte via the switch below; any character not representable in CP1252 is stored as
    * {@code '?'} (0x3F). Returns the number of bytes written (always {@code length}).</p>
    *
    * <p>The anti-tamper dummy parameter {@code marker} (obf byte param 4) was checked against
    * {@code -78} and otherwise tripped {@link #wrap}; that dead guard has been dropped. Callers pass
    * arbitrary values (e.g. {@code -118}, {@code -112}) for it.</p>
    *
    * @param length    number of characters to encode (was param 0)
    * @param destStart destination offset in {@code dest} (was param 1)
    * @param srcStart  source offset in {@code text} (was param 2)
    * @param text      source characters (was param 3)
    * @param dest      destination byte array (was param 5)
    * @return number of bytes written
    */
   static final int encodeCp1252(int length, int destStart, int srcStart, CharSequence text, byte[] dest) {
      for (int i = 0; i < length; i++) {
         char ch = text.charAt(srcStart + i);
         byte encoded;

         // Pass-through ranges: ASCII (1..0x7F) and Latin-1 high (0xA0..0xFF).
         if ((ch > 0 && ch < '') || (ch >= ' ' && ch <= 'ÿ')) {
            encoded = (byte) ch;
         } else {
            // Windows-1252 special band (0x80-0x9F). Each constant is the CP1252 byte for that
            // Unicode code point (verified against the official CP1252 table).
            switch (ch) {
               case '€': encoded = (byte) -128; break; // U+20AC EURO            -> 0x80
               case '‚': encoded = (byte) -126; break; // U+201A SINGLE LOW QUOTE -> 0x82
               case 'ƒ': encoded = (byte) -125; break; // U+0192 f WITH HOOK      -> 0x83
               case '„': encoded = (byte) -124; break; // U+201E DOUBLE LOW QUOTE -> 0x84
               case '…': encoded = (byte) -123; break; // U+2026 HORIZONTAL ELLIPSIS -> 0x85
               case '†': encoded = (byte) -122; break; // U+2020 DAGGER          -> 0x86
               case '‡': encoded = (byte) -121; break; // U+2021 DOUBLE DAGGER   -> 0x87
               case 'ˆ': encoded = (byte) -120; break; // U+02C6 CIRCUMFLEX      -> 0x88
               case '‰': encoded = (byte) -119; break; // U+2030 PER MILLE       -> 0x89
               case 'Š': encoded = (byte) -118; break; // U+0160 S WITH CARON    -> 0x8A
               case '‹': encoded = (byte) -117; break; // U+2039 SINGLE L ANGLE  -> 0x8B
               case 'Œ': encoded = (byte) -116; break; // U+0152 OE LIGATURE     -> 0x8C
               case 'Ž': encoded = (byte) -114; break; // U+017D Z WITH CARON    -> 0x8E
               case '‘': encoded = (byte) -111; break; // U+2018 LEFT SINGLE QUOTE  -> 0x91
               case '’': encoded = (byte) -110; break; // U+2019 RIGHT SINGLE QUOTE -> 0x92
               case '“': encoded = (byte) -109; break; // U+201C LEFT DOUBLE QUOTE  -> 0x93
               case '”': encoded = (byte) -108; break; // U+201D RIGHT DOUBLE QUOTE -> 0x94
               case '•': encoded = (byte) -107; break; // U+2022 BULLET          -> 0x95
               case '–': encoded = (byte) -106; break; // U+2013 EN DASH         -> 0x96
               case '—': encoded = (byte) -105; break; // U+2014 EM DASH         -> 0x97
               case '˜': encoded = (byte) -104; break; // U+02DC SMALL TILDE     -> 0x98
               case '™': encoded = (byte) -103; break; // U+2122 TRADE MARK      -> 0x99
               case 'š': encoded = (byte) -102; break; // U+0161 s WITH CARON    -> 0x9A
               case '›': encoded = (byte) -101; break; // U+203A SINGLE R ANGLE  -> 0x9B
               case 'œ': encoded = (byte) -100; break; // U+0153 oe LIGATURE     -> 0x9C
               case 'ž': encoded = (byte) -98;  break; // U+017E z WITH CARON    -> 0x9E
               case 'Ÿ': encoded = (byte) -97;  break; // U+0178 Y WITH DIAERESIS-> 0x9F
               default:       encoded = (byte) '?';  break; // not representable in CP1252
            }
         }

         dest[destStart + i] = encoded;
      }

      return length;
   }

   /**
    * Obf {@code i(String,int)} constructor. Stores {@code value} in {@link #instanceValue}.
    * (Instances of this class appear to be unused; the original wrapped the trivial body in the
    * standard self-reporting try/catch, which is dropped here.)
    *
    * @param label unused descriptive label (only referenced when building the error breadcrumb)
    * @param value scratch int stored in {@link #instanceValue}
    */
   ErrorHandler(String label, int value) {
      this.instanceValue = value;
   }

   /**
    * Obf {@code a(Throwable,String)}: the central exception-wrapping sink.
    *
    * <p>If {@code cause} is already a {@link ClientRuntimeException} (obf {@code la}), append the new
    * {@code context} breadcrumb to its existing message and return it (so a single exception
    * accumulates the chain of call signatures it bubbled through). Otherwise, wrap {@code cause} in a
    * fresh {@link ClientRuntimeException} carrying {@code context}.</p>
    *
    * @param cause   the throwable being reported
    * @param context human-readable "method(args)" breadcrumb
    * @return the (new or augmented) {@link ClientRuntimeException} to throw
    */
   static final ClientRuntimeException wrap(Throwable cause, String context) {
      ClientRuntimeException wrapped;
      if (cause instanceof ClientRuntimeException) {
         wrapped = (ClientRuntimeException) cause;
         wrapped.context = wrapped.context + ' ' + context; // obf field la.h: the breadcrumb message
      } else {
         wrapped = new ClientRuntimeException(cause, context);
      }
      return wrapped;
   }

   /**
    * Obf {@code C} (exposed as {@code a(int)}): registry of the engine's {@link ChatCipher}
    * (obf {@code v}) instances. Returns the fixed set so callers can resolve a cipher by its id
    * field. The anti-tamper dummy parameter (was checked against {@code -711} before tripping
    * {@link #wrap}) has been dropped; the sole caller {@code NameTable} passed {@code -711}.
    *
    * @return the seven well-known {@link ChatCipher} instances spread across the engine classes
    */
   static final ChatCipher[] chatCiphers() {
      return new ChatCipher[] {
         Surface.E,             // obf ua.E
         ClientStream.O,        // obf da.O
         CharTable.c,           // obf ga.c
         GameCharacter.f,       // obf ta.f
         ClientRuntimeException.b, // obf la.b
         AudioMixer.d,          // obf eb.d
         ProxySocketFactory.n   // obf gb.n
      };
   }

   /**
    * Static initialiser: builds the {@link #CRC64_TABLE} (CRC-64/ECMA-182) and decodes the
    * {@link #BASE64_ALPHABET}.
    */
   static {
      // CRC-64/ECMA-182 table generation. Reflected polynomial 0xC96C5795D7870F42.
      for (int index = 0; index < 256; index++) {
         long crc = index;
         for (int bit = 0; bit < 8; bit++) {
            if ((crc & 1L) == 1L) {
               crc = (crc >>> 1) ^ 0xC96C5795D7870F42L;
            } else {
               crc >>>= 1;
            }
         }
         CRC64_TABLE[index] = crc;
      }

      // Custom Base64 alphabet ("A-Za-z0-9!\"") plus a 32-byte trailing block.
      // (Original was stored XOR-obfuscated; decoded to its plaintext here.)
      BASE64_ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!\"";
   }
}
