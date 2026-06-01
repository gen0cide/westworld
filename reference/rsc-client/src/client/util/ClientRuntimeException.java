package client.util;

import client.net.ChatCipher;

/**
 * ClientRuntimeException — the client's internal unchecked-exception carrier (obfuscated name: {@code la}).
 *
 * <p>This is the unchecked counterpart to {@link ClientIOException} ({@code fa}) and is the single
 * {@link RuntimeException} type produced by the engine's per-method error-handling obfuscation. Almost every
 * method in this Microsoft&nbsp;J++ build is wrapped in:
 * <pre>
 *     try { ...body... }
 *     catch (RuntimeException e) { throw ErrorHandler.a(e, "methodSig(args)"); }
 * </pre>
 * {@link ErrorHandler#wrap(Throwable, String)} ({@code i.a}) is what creates / extends instances of this class:
 * <ul>
 *   <li>If the caught {@code Throwable} is already a {@code ClientRuntimeException}, the new method-signature
 *       string is <b>appended</b> to {@link #context} (separated by a space), so {@link #context} accumulates a
 *       breadcrumb call-trail as the exception unwinds the obfuscated try/catch chain.</li>
 *   <li>Otherwise a fresh instance is allocated, storing the original throwable in {@link #cause} and the first
 *       method-signature string in {@link #context}.</li>
 * </ul>
 *
 * <p>The class itself carries no behaviour beyond holding those two fields — it does not override
 * {@code getMessage()} or chain {@code cause} into the standard {@link Throwable} machinery; the engine reads the
 * {@link #cause}/{@link #context} fields directly when reporting fatal errors.
 *
 * <p><b>Obfuscation note:</b> the static fields below ({@link #LOCAL_CIPHER}, {@link #intScratch},
 * {@link #intCounter}, {@link #byteScratch}, {@link #byteRowScratch}, {@link #stringScratch}) and the two {@code z}
 * string-decryption helpers are dead J++ boilerplate that the obfuscator stamps onto exception classes uniformly.
 * They are never read by the exception logic; {@link #LOCAL_CIPHER} is merely a lazily-built {@link ChatCipher}
 * whose constructor strings decrypt to {@code "LOCAL"}/{@code "local"}. They are preserved here (renamed) so this
 * deobfuscation stays faithful to the original class layout.
 */
public final class ClientRuntimeException extends RuntimeException {

   /** The original throwable that triggered the error (obf: {@code e}); the wrapped cause. */
   public Throwable cause;

   /**
    * Accumulated context breadcrumb (obf: {@code h}) — a space-separated chain of {@code "methodSig(args)"}
    * strings, one appended each time {@link ErrorHandler#wrap} re-wraps this exception while it unwinds.
    */
   public String context;

   /**
    * Dead obfuscator boilerplate (obf: {@code b}): a {@link ChatCipher} built from strings that decrypt to
    * {@code "LOCAL"}/{@code "local"}. Never used by the exception logic; present only because the J++ obfuscator
    * stamps an identical static field onto these utility/exception classes.
    */
   public static ChatCipher LOCAL_CIPHER = new ChatCipher(
       decryptString(decryptCharArray("[\fDi<")),  // decrypts to "LOCAL"
       "",
       decryptString(decryptCharArray("{,dI")), // decrypts to "local"
       4);

   /** Dead obfuscator scratch array (obf: {@code a}); never populated or read. */
   public static int[] intScratch;

   /** Dead obfuscator scratch counter (obf: {@code d}); never incremented or read. */
   public static int intCounter;

   /**
    * Shared 520-byte sector I/O scratch buffer (obf: {@code c}).
    *
    * <p>NOT dead: this is the one real field the obfuscator parked on this exception class.
    * {@link client.data.ArchiveReader} ({@code ob}) is the sole user — it reads and writes
    * every 520-byte cache sector header/payload through this buffer (obf references: {@code la.c}).
    * Kept {@code public} so the cross-package reader can reach it (in the original default-package
    * build it was package-private).
    */
   public static byte[] byteScratch = new byte[520];

   /** Dead obfuscator scratch buffer (obf: {@code g}), 12 rows; never used. */
   public static byte[][] byteRowScratch = new byte[12][];

   /** Dead obfuscator scratch string table (obf: {@code f}); never populated or read. */
   public static String[] stringScratch;

   /**
    * Wraps an original throwable plus an initial context string.
    * (Original wrapped its two field assignments in the standard rethrow-only try/catch, removed here.)
    *
    * @param cause   the throwable being wrapped (stored in {@link #cause})
    * @param context the initial method-signature breadcrumb (stored in {@link #context})
    */
   public ClientRuntimeException(Throwable cause, String context) {
      this.cause = cause;
      this.context = context;
   }

   /**
    * J++ string-decryption stage 1 (obf: {@code z(String)}): converts a packed string literal to a char array,
    * XOR-flipping the single character of length-1 strings with {@code 'p'} (0x70). Dead support code for the
    * unused {@link #LOCAL_CIPHER}.
    */
   private static char[] decryptCharArray(String packed) {
      char[] chars = packed.toCharArray();
      if (chars.length < 2) {
         chars[0] = (char)(chars[0] ^ 'p'); // 0x70
      }
      return chars;
   }

   /**
    * J++ string-decryption stage 2 (obf: {@code z(char[])}): XOR-decrypts each char against a fixed 5-element key
    * cycle {23, 67, 7, 40, 112} keyed by index mod 5, then interns the result. Dead support code for the unused
    * {@link #LOCAL_CIPHER}; produces strings such as {@code "LOCAL"}/{@code "local"}.
    */
   private static String decryptString(char[] chars) {
      for (int i = 0; i < chars.length; i++) {
         byte key;
         switch (i % 5) {
            case 0:
               key = 23;
               break;
            case 1:
               key = 67;
               break;
            case 2:
               key = 7;
               break;
            case 3:
               key = 40;
               break;
            default:
               key = 112;
         }
         chars[i] = (char)(chars[i] ^ key);
      }
      return new String(chars).intern();
   }
}
