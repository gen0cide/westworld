package client.util;

/**
 * ArrayUtil — tiny static array helper class.
 *
 * Provides two hand-unrolled bulk-memory operations that the compiler could not
 * inline as System.arraycopy calls in the Microsoft J++ build:
 *
 *   fill(int[], offset, length)              — zero-fills a range of an int array
 *   copy(byte[], srcOff, byte[], dstOff, len) — overlapping-safe byte-array copy
 *
 * Both methods use an 8-element Duff's-device-style unroll: they advance by 8
 * elements per iteration for the bulk of the range, then mop up the trailing
 * remainder one element at a time.  The copy method additionally detects the
 * self-overlapping case (same array, destination inside the source range) and
 * copies backwards to avoid corruption — matching the semantics of memmove(3).
 *
 * These helpers are called throughout the engine wherever the client needs fast
 * bulk zeroing of pixel/model buffers or safe in-place byte-buffer shifts.
 *
 * Obfuscated class: ab
 * No obfuscation artifacts were present in this class (no opaque predicate, no
 * profiling counter, no exception wrapper, no XOR string pool).
 */
public final class ArrayUtil {

    // Private constructor: all methods are static; this class is never instantiated.
    private ArrayUtil() {}

    // -------------------------------------------------------------------------
    // fill — zero a contiguous range of an int array
    // -------------------------------------------------------------------------

    /**
     * Fills {@code length} elements of {@code arr} with 0, starting at index
     * {@code offset}.
     *
     * The inner loop is unrolled 8× so that the JIT/interpreter can pipeline
     * the stores.  Any trailing 1–7 elements are handled by a scalar cleanup
     * loop.
     *
     * Equivalent to: Arrays.fill(arr, offset, offset + length, 0)
     *
     * Obf: ab.a(int[], int, int)
     *
     * @param arr    the array to zero
     * @param offset first index to clear (inclusive)
     * @param length number of elements to clear
     */
    public static final void fill(int[] arr, int offset, int length) {
        // Compute the last index covered by the 8-element unrolled loop.
        // After the loop, 'end' is bumped back up to the true end so the
        // scalar tail loop covers the residual 0–7 elements.
        int end = offset + length - 7; // obf: var2

        // 8-element unrolled bulk zeroing
        while (offset < end) {
            arr[offset++] = 0;
            arr[offset++] = 0;
            arr[offset++] = 0;
            arr[offset++] = 0;
            arr[offset++] = 0;
            arr[offset++] = 0;
            arr[offset++] = 0;
            arr[offset++] = 0;
        }

        // Scalar tail: covers the remainder when length % 8 != 0.
        // Also handles the entire range when length < 8 (end starts ≤ offset,
        // so the unrolled loop body is skipped entirely and we land here).
        end += 7; // restore to true end index (exclusive)
        while (offset < end) {
            arr[offset++] = 0;
        }
    }

    // -------------------------------------------------------------------------
    // copy — memmove-style overlapping-safe byte array copy
    // -------------------------------------------------------------------------

    /**
     * Copies {@code length} bytes from {@code src[srcOffset..]} to
     * {@code dst[dstOffset..]}, handling the case where {@code src == dst} and
     * the destination window overlaps the source window (i.e. an in-place
     * shift).
     *
     * Copy direction:
     *   - Forward  (src != dst, OR non-overlapping): copies low-to-high, 8×
     *     unrolled then scalar tail.
     *   - Backward (src == dst AND dst overlaps src from above, i.e.
     *     dstOffset > srcOffset && dstOffset < srcOffset + length): copies
     *     high-to-low to prevent overwriting source bytes before they are read.
     *     Also 8× unrolled then scalar tail.
     *
     * Equivalent to: System.arraycopy(src, srcOffset, dst, dstOffset, length)
     *
     * Obf: ab.a(byte[], int, byte[], int, int)
     *
     * @param src       source byte array
     * @param srcOffset first source index (inclusive)
     * @param dst       destination byte array (may be the same object as src)
     * @param dstOffset first destination index (inclusive)
     * @param length    number of bytes to copy
     */
    public static final void copy(byte[] src, int srcOffset, byte[] dst, int dstOffset, int length) {
        // Overlapping self-copy: if src and dst are the same array object and
        // the destination start falls inside the source window, we must copy
        // backwards to avoid clobbering unread bytes.
        if (src == dst) {
            if (srcOffset == dstOffset) {
                // Trivial no-op: source and destination are identical.
                return;
            }

            if (dstOffset > srcOffset && dstOffset < srcOffset + length) {
                // Backward copy path (memmove overlapping case).
                // Repoint both cursors to the last byte of their respective
                // windows so we can decrement towards the start.
                srcOffset += --length; // srcOffset now points to last src byte; length is (original - 1)
                dstOffset += length;   // dstOffset now points to last dst byte

                // 'end' becomes the index of the first byte of the region
                // (i.e. original srcOffset before the += above), adjusted for
                // the 8× unroll.
                int end = srcOffset - length; // obf: var4  (== original srcOffset)
                end += 7;                      // sentinel for 8× unrolled loop

                // 8-element unrolled backward bulk copy
                while (srcOffset >= end) {
                    dst[dstOffset--] = src[srcOffset--];
                    dst[dstOffset--] = src[srcOffset--];
                    dst[dstOffset--] = src[srcOffset--];
                    dst[dstOffset--] = src[srcOffset--];
                    dst[dstOffset--] = src[srcOffset--];
                    dst[dstOffset--] = src[srcOffset--];
                    dst[dstOffset--] = src[srcOffset--];
                    dst[dstOffset--] = src[srcOffset--];
                }

                // Scalar backward tail
                end -= 7; // restore to true first index of the region
                while (srcOffset >= end) {
                    dst[dstOffset--] = src[srcOffset--];
                }
                return;
            }
        }

        // Forward copy path: either different arrays, or same array with
        // non-overlapping / safe-forward overlap.
        int end = length + srcOffset; // obf: var4  (exclusive end of source window)
        end -= 7;                      // sentinel for 8× unrolled loop

        // 8-element unrolled forward bulk copy
        while (srcOffset < end) {
            dst[dstOffset++] = src[srcOffset++];
            dst[dstOffset++] = src[srcOffset++];
            dst[dstOffset++] = src[srcOffset++];
            dst[dstOffset++] = src[srcOffset++];
            dst[dstOffset++] = src[srcOffset++];
            dst[dstOffset++] = src[srcOffset++];
            dst[dstOffset++] = src[srcOffset++];
            dst[dstOffset++] = src[srcOffset++];
        }

        // Scalar forward tail
        end += 7; // restore to true exclusive end
        while (srcOffset < end) {
            dst[dstOffset++] = src[srcOffset++];
        }
    }
}
