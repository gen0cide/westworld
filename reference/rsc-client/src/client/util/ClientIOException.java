package client.util;

import java.io.IOException;

/**
 * ClientIOException (obfuscated: {@code fa}) — the client's custom <em>checked</em>
 * exception type, a thin subclass of {@link IOException}.
 *
 * <p>This is the I/O sibling of {@code ClientRuntimeException} ({@code la}, the unchecked
 * wrapper). It carries a message string and is thrown from the networking layer to signal
 * connection/transport failures that callers are expected to handle. For example the proxy
 * socket factory ({@code ProxySocketFactory}/{@code gb}) catches this type while iterating
 * candidate proxies and remembers the last one so it can be re-thrown if every proxy fails,
 * while letting plain {@link IOException}s fall through silently.</p>
 *
 * <p>Beyond the exception behaviour, the J++ obfuscator hung an unrelated piece of state on
 * this class as a hiding place: an XOR-obfuscated string table ({@link #STRING_TABLE}) plus
 * some dead profiling/scratch fields. Decoding that table at class-load time yields ordinary
 * UI prompt strings (the single entry here is the bank "buy quantity" prompt). The
 * {@link #xorDecodeStage1}/{@link #xorDecodeStage2} pair is the obfuscator's standard
 * compile-time string-literal decoder, reproduced here verbatim so the literals stay opaque
 * in the class file but resolve to plain text at runtime.</p>
 */
public final class ClientIOException extends IOException {

   /**
    * Dead profiling/scratch counter emitted by the obfuscator; never read in real logic.
    * Retained only so the class layout matches the original.
    */
   public static int profilingCounter;

   /**
    * Obfuscated string table. Each entry is produced by decoding a compile-time XOR-encoded
    * literal via {@link #xorDecodeStage1} then {@link #xorDecodeStage2}. The sole decoded
    * entry is the bank prompt: <code>"Type the number of items to buy and press enter"</code>.
    */
   public static String[] STRING_TABLE = new String[]{
      xorDecodeStage2(xorDecodeStage1(
         "F%:f4/ZZg1(F23,Z]f9'\tf3jAk|+P2,8Ga|/@w."))
   };

   /** Client build/revision number (234) baked in by the obfuscator as an anti-tamper marker. */
   public static int BUILD_REVISION = 234;

   /** Unused scratch int array left in place by the obfuscator; never initialized or read. */
   public static int[] unusedScratchA;

   /** Unused scratch int array left in place by the obfuscator; never initialized or read. */
   public static int[] unusedScratchB;

   /**
    * Constructs the checked exception with the given detail message, delegating to
    * {@link IOException#IOException(String)}.
    */
   public ClientIOException(String message) {
      super(message);
   }

   /**
    * Obfuscator string-decode stage 1: only meaningful for 1-char literals, where it XORs the
    * lone character with {@code '4'} (0x34). For the long literals stored here the length is
    * {@code >= 2}, so this is effectively a pass-through that simply returns the char array.
    */
   private static char[] xorDecodeStage1(String encoded) {
      char[] chars = encoded.toCharArray();
      if (chars.length < 2) {
         chars[0] = (char)(chars[0] ^ '4'); // 0x34
      }
      return chars;
   }

   /**
    * Obfuscator string-decode stage 2: XORs each character against a fixed 5-element repeating
    * key keyed by {@code index % 5}, then interns the result. Reverses the compile-time
    * encoding to recover the original literal text.
    */
   private static String xorDecodeStage2(char[] chars) {
      for (int index = 0; index < chars.length; index++) {
         // Repeating 5-byte XOR key: positions 0..4 -> 18,92,74,122,52.
         byte key;
         switch (index % 5) {
            case 0:
               key = 18;
               break;
            case 1:
               key = 92;
               break;
            case 2:
               key = 74;
               break;
            case 3:
               key = 122;
               break;
            default:
               key = 52;
         }

         chars[index] = (char)(chars[index] ^ key);
      }

      return new String(chars).intern();
   }
}
