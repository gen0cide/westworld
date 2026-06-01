package client.data;

import client.util.ErrorHandler;

import java.io.EOFException;
import java.io.File;
import java.io.IOException;
import java.io.RandomAccessFile;

/**
 * CacheFile (obfuscated {@code d}) — a thin wrapper around a {@link RandomAccessFile} used as a
 * size-bounded, sequential on-disk cache file (the client's {@code .file_store} caches).
 *
 * <p>Role in the engine: the {@link client.shell.LoaderThread} (obf {@code c}) opens several of
 * these at startup — one small index/scratch store, one large (~300&nbsp;MB) media store, one
 * ~1&nbsp;MB store, plus a per-region array of ~1&nbsp;MB stores. Each is created in {@code "rw"}
 * mode with a maximum byte budget. Higher-level cache logic ({@link client.data.DataStore},
 * {@link client.data.ArchiveReader}, {@link client.data.CacheUpdater}) reads and writes records
 * through this class.</p>
 *
 * <p>The wrapper keeps its own running {@link #position} cursor in parallel with the underlying
 * file pointer so that writes can be cheaply checked against {@link #maxLength}: when a write
 * would push the file past its budget the cache is flagged corrupt (a sentinel byte {@code 1} is
 * written at the budget offset) and an {@link EOFException} is raised so the caller can rebuild
 * the cache instead of silently overflowing it.</p>
 *
 * <p>On construction the very first byte of the file is read and immediately re-written (when the
 * file is opened writable); this is a "touch" that forces the OS to materialize/flush the file's
 * first page and validates that the store is writable before the client starts relying on it.</p>
 *
 * <p>Deobfuscation notes: the original class is riddled with per-method profiling counters
 * (static ints {@code n,e,l,k,d,i,f,a}), an {@code ErrorHandler}-based try/catch wrapper around
 * every body that only appends a method-signature breadcrumb, and dead anti-tamper guards keyed on
 * dummy parameters (e.g. {@code if (var1 != 25291) ...}). All of that has been stripped here; only
 * the real I/O logic remains. The dummy guard parameters that were removed are called out on each
 * method below.</p>
 */
final class CacheFile {

   // --- Profiling counters (original static ints, retained for completeness; never read for logic) ---
   /** Profiling counter incremented on entry to {@link #close()} (obf {@code n}). */
   static int closeCallCount;
   /** Profiling counter incremented on entry to {@link #read} (obf {@code e}). */
   static int readCallCount;
   /** Unused profiling/state counter (obf {@code l}); initialized to 0 and never used meaningfully. */
   static int unusedCounter = 0;
   /** Profiling counter incremented on entry to {@link #seek} (obf {@code k}). */
   static int seekCallCount;
   /** Profiling counter incremented on entry to {@link #write} (obf {@code d}). */
   static int writeCallCount;
   /** Profiling counter incremented on entry to {@link #length()} (obf {@code i}). */
   static int lengthCallCount;
   /** Profiling counter incremented on entry to {@link #getUnsignedShort} (obf {@code f}). */
   static int getUnsignedShortCallCount;
   /** Unused/uninitialized shared int array (obf {@code g}); never populated in this class. */
   static int[] sharedScratch;
   /** Profiling counter incremented on entry to {@link #finalize()} (obf {@code a}). */
   static int finalizeCallCount;

   /** The underlying random-access file handle (obf {@code j}); {@code null} once closed. */
   private RandomAccessFile file;

   /**
    * Running byte cursor / logical length tracked in parallel with the file pointer (obf {@code c}).
    * Advanced by {@link #write} and {@link #read}, set absolutely by {@link #seek}.
    */
   private long position;

   /** Maximum number of bytes this cache file is allowed to grow to (obf {@code b}). */
   private long maxLength;

   /**
    * Unused shared static reference (obf {@code h}, type obf {@code e} = GameShell). Only ever
    * assigned {@code null} by a dead anti-tamper branch; kept for fidelity.
    */
   static client.shell.GameShell unusedShell;

   /**
    * Unrelated string pool reused from the obfuscator's shared constant table (obf {@code m}).
    * Decodes to bank-withdrawal UI prompts and has nothing to do with cache files; preserved
    * verbatim because the obfuscator parked these strings in this class.
    */
   static String[] sharedStrings = new String[]{
      deobfuscate(deobfuscate("(XoB`EF^zY{\\Q\f.^zTG^z^XCzYFy")), // "Please enter the number of items to withdraw"
      deobfuscate(deobfuscate("Z.A\nQ\r}Z\nkC"))                                                                             // "and press enter"
   };

   /**
    * Obfuscated method-signature breadcrumbs (obf {@code z}) that the original code appended to
    * wrapped exceptions. Decoded values, in order:
    * {@code {0:"{...}", 1:"d.D(", 2:"null", 3:"d.B(", 4:"d.F(", 5:"d.<init>(", 6:"d.finalize()",
    * 7:"d.G(", 8:"d.A(", 9:"d.E(", 10:"d.C("}}. No longer referenced after the try/catch wrappers
    * were stripped; retained so the static initializer remains faithful.
    */
   private static final String[] SIGNATURES = new String[]{
      deobfuscate(deobfuscate("P L")),
      deobfuscate(deobfuscate(":&")),
      deobfuscate(deobfuscate("Ab")),
      deobfuscate(deobfuscate("<&")),
      deobfuscate(deobfuscate("8&")),
      deobfuscate(deobfuscate("Bg_@@&")),
      deobfuscate(deobfuscate("g_XtTP")),
      deobfuscate(deobfuscate("9&")),
      deobfuscate(deobfuscate("?&")),
      deobfuscate(deobfuscate(";&")),
      deobfuscate(deobfuscate("=&"))
   };

   /**
    * Opens a size-bounded cache file (obf constructor {@code d(File, String, long)}).
    *
    * @param target    the on-disk cache file
    * @param mode      {@link RandomAccessFile} access mode, e.g. {@code "rw"} or {@code "r"}
    * @param maxBytes  maximum byte budget; {@code -1} is treated as "unbounded" ({@link Long#MAX_VALUE})
    */
   CacheFile(File target, String mode, long maxBytes) throws IOException {
      // A budget of -1 means "no limit"; promote it to the maximum possible length.
      if (maxBytes == -1L) {
         maxBytes = Long.MAX_VALUE;
      }

      // If the existing file is already larger than its budget, it is stale/corrupt — discard it.
      // (Original used inverted-compare junk math: ~maxBytes > ~file.length()  <=>  maxBytes < file.length().)
      if (maxBytes < target.length()) {
         target.delete();
      }

      this.file = new RandomAccessFile(target, mode);
      this.maxLength = maxBytes;
      this.position = 0L;

      // "Touch" the first byte: read it and (for writable stores) write it straight back, forcing
      // the file's first page to materialize and confirming the store is writable.
      int firstByte = this.file.read();
      if (firstByte != -1 && !mode.equals("r")) {
         this.file.seek(0L);
         this.file.write(firstByte);
      }

      // Leave the file pointer at the start, ready for the first record access.
      this.file.seek(0L);
   }

   /**
    * Closes the underlying file handle and releases it (obf {@code a(int)} ; private).
    * Removed dummy guard param {@code var1} (original skipped a dead self-call unless it was 25291).
    */
   private void close() throws IOException {
      if (this.file != null) {
         this.file.close();
         this.file = null;
      }
   }

   /**
    * Returns the current physical length of the underlying file in bytes (obf {@code a(byte)}).
    * Removed dummy guard param {@code var1} (original clobbered {@link #position} to -52 unless 47).
    */
   long length() throws IOException {
      return this.file.length();
   }

   /**
    * Finalizer (obf {@code finalize()}): closes the file if it was left open. The original printed
    * an empty line as a side effect before closing; preserved for fidelity.
    */
   @Override
   protected void finalize() throws Throwable {
      if (this.file != null) {
         System.out.println("");
         this.close();
      }
   }

   /**
    * Bitwise-OR helper (obf static {@code a(int, int)}). Combines two int flags/values; used by
    * callers to fold record header bits together.
    */
   static int or(int a, int b) {
      return a | b;
   }

   /**
    * Appends {@code length} bytes from {@code data} (starting at {@code dataOffset}) to the cache,
    * enforcing the size budget (obf {@code b(byte[], int, int, int)}).
    *
    * <p>If the write would push {@link #position} past {@link #maxLength}, the store is flagged
    * corrupt by writing a sentinel byte {@code 1} at the budget offset and an {@link EOFException}
    * is thrown so the caller can rebuild the cache.</p>
    *
    * @param data       source buffer
    * @param length     number of bytes to write
    * @param dataOffset offset within {@code data} of the first byte to write
    *
    * Removed dummy guard param (original 3rd arg {@code var3}: a dead self-call unless != 1).
    * Note: original parameter order was {@code (data, length, guard, dataOffset)}; the guard has
    * been dropped, leaving {@code (data, length, dataOffset)}.
    */
   void write(byte[] data, int length, int dataOffset) throws IOException {
      // Budget check (original: ~maxLength > ~(position + length)  <=>  maxLength < position + length).
      if (this.maxLength < this.position + length) {
         this.file.seek(this.maxLength);
         this.file.write(1); // sentinel: mark the store as corrupt/overflowed
         throw new EOFException();
      }

      this.file.write(data, dataOffset, length);
      this.position += length;
   }

   /**
    * Reads a big-endian unsigned 16-bit value from {@code buffer} at {@code offset}
    * (obf static {@code a(int, byte, byte[])}; equivalent to oracle {@code Utility.getUnsignedShort}).
    *
    * <p>Original placed {@code buffer} last and computed {@code << 922410888}; since shifts are
    * taken mod 32, {@code 922410888 % 32 == 8}, i.e. a plain high-byte shift.</p>
    *
    * @return {@code (buffer[offset] << 8) | buffer[offset + 1]}, unsigned
    *
    * Removed dummy guard param (original {@code var1}: nulled {@link #unusedShell} when {@code < 4}).
    */
   static int getUnsignedShort(byte[] buffer, int offset) {
      int high = (buffer[offset] & 0xFF) << 8;
      int low = buffer[offset + 1] & 0xFF;
      return high + low;
   }

   /**
    * Seeks the file (and the logical cursor) to an absolute byte {@code offset}
    * (obf {@code a(int, long)} ; Vineflower failed to decompile this — reconstructed from CFR + bytecode).
    *
    * Removed dummy guard param (original 1st arg {@code var1}: early {@code return} unless 0).
    */
   void seek(long offset) throws IOException {
      this.file.seek(offset);
      this.position = offset;
   }

   /**
    * Reads up to {@code length} bytes into {@code buffer} at {@code bufferOffset}
    * (obf {@code a(byte[], int, int, int)}). Advances the logical {@link #position} by the number
    * of bytes actually read, unless nothing was read (EOF), in which case the cursor is left intact.
    *
    * @return the number of bytes read, or a non-positive value at EOF
    *
    * Removed dummy guard param (original 4th arg {@code var4}, effectively -1): the clean-base
    * check is {@code if (~bytesRead < var4) position += bytesRead}; with {@code var4 == -1} that is
    * {@code ~bytesRead < -1}  <=>  {@code bytesRead > 0}, i.e. advance only when bytes were read.
    * Note: original parameter order was {@code (buffer, length, bufferOffset, guard)}; dropping the
    * guard leaves {@code (buffer, length, bufferOffset)}.
    */
   int read(byte[] buffer, int length, int bufferOffset) throws IOException {
      int bytesRead = this.file.read(buffer, bufferOffset, length);
      if (bytesRead > 0) {
         this.position += bytesRead;
      }
      return bytesRead;
   }

   /**
    * Conditionally pre-mangles a 1-char encoded string before the main XOR pass (obf {@code z(String)}).
    * For single-character payloads the leading char is XOR-flipped with {@code '1'} (0x31); longer
    * strings are returned unchanged. Part of the static string-deobfuscation routine.
    */
   private static char[] deobfuscate(String encoded) {
      char[] chars = encoded.toCharArray();
      if (chars.length < 2) {
         chars[0] = (char) (chars[0] ^ '1');
      }
      return chars;
   }

   /**
    * XOR-decodes an obfuscated string (obf {@code z(char[])}). Each character is XOR'd with a
    * repeating 5-byte key keyed by index modulo 5: {@code [120, 52, 126, 14, 49]}. The result is
    * interned. Used to rebuild {@link #SIGNATURES} and {@link #sharedStrings}.
    */
   private static String deobfuscate(char[] chars) {
      for (int i = 0; i < chars.length; i++) {
         char c = chars[i];
         int key;
         switch (i % 5) {
            case 0:
               key = 120;
               break;
            case 1:
               key = 52;
               break;
            case 2:
               key = 126;
               break;
            case 3:
               key = 14;
               break;
            default:
               key = 49;
         }
         chars[i] = (char) (c ^ key);
      }
      return new String(chars).intern();
   }
}
