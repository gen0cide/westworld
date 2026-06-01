package client.nativeapi;

import com.ms.awt.WComponentPeer;
import com.ms.com.IUnknown;
import com.ms.directX.DDSurfaceDesc;
import com.ms.directX.DirectDraw;
import com.ms.directX.IEnumModesCallback;
import com.ms.win32.User32;
import java.awt.Frame;

/**
 * DirectDraw display-mode enumerator and exclusive fullscreen setup helper.
 *
 * <p>Wraps the Microsoft J++ {@code com.ms.directX.DirectDraw} COM object and implements
 * {@link IEnumModesCallback} so it can be passed directly to
 * {@link DirectDraw#enumDisplayModes} as the callback receiver. The class is used
 * exclusively on Windows (Microsoft JVM) when the client is running in its native
 * DirectDraw path; the AWT fallback lives in {@code client.nativeapi.DisplayModeSetter}
 * (obf: {@code ha}).
 *
 * <p>Usage pattern (from {@code client.shell.LoaderThread}, obf: {@code c}):
 * <ol>
 *   <li>Construct → initialises a default-adapter {@code DirectDraw} object.
 *   <li>Call {@link #listModes()} to enumerate supported display modes as a flat
 *       {@code int[]} of {@code {width, height, bpp, refreshRate}} quads.
 *   <li>Call {@link #enterFullscreen(Frame, int, int, int, int)} to switch the AWT
 *       {@code Frame} into exclusive fullscreen at the desired mode.
 *   <li>Call {@link #exitFullscreen(Frame)} to restore the original display mode and
 *       drop exclusive cooperative level.
 * </ol>
 *
 * <p>Two static fields ({@link #modeBuffer} / {@link #modeIndex}) are used as shared
 * state between the two-pass enumeration inside {@link #callbackEnumModes}; they are
 * always reset to {@code null}/0 before this object leaves any public method.
 *
 * <p>obf: {@code wa}
 */
final class DirectDrawModes implements IEnumModesCallback {

    // -------------------------------------------------------------------------
    // Static shared state used during two-pass mode enumeration
    // -------------------------------------------------------------------------

    /**
     * Flat array accumulating display-mode quads during the second enumeration pass.
     * {@code null} during the first (counting) pass; allocated to the exact size
     * before the second pass begins.
     *
     * <p>obf: {@code b}
     */
    private static int[] modeBuffer; // obf: b

    /**
     * Dual-purpose static counter:
     * <ul>
     *   <li>First pass (counting): incremented by 4 for each mode, so after the pass
     *       its value equals {@code numModes * 4} — the required array length.
     *   <li>Second pass (filling): used as a write cursor into {@link #modeBuffer}.
     * </ul>
     * Always reset to 0 when enumeration is complete.
     *
     * <p>obf: {@code a}
     */
    private static int modeIndex; // obf: a

    // -------------------------------------------------------------------------
    // Instance fields
    // -------------------------------------------------------------------------

    /**
     * The underlying DirectDraw COM object initialised against the default display
     * adapter ({@code initialize(null)}).
     *
     * <p>obf: {@code c}
     */
    private DirectDraw directDraw; // obf: c

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /**
     * Constructs a new {@code DirectDrawModes}, creating and initialising a
     * {@link DirectDraw} object against the primary display adapter.
     *
     * <p>obf: {@code wa()}
     */
    public DirectDrawModes() {
        // Allocate the DirectDraw COM wrapper and initialise it for the default
        // (primary) display adapter.  Passing null selects the primary adapter.
        this.directDraw = new DirectDraw();
        this.directDraw.initialize(null);
    }

    // -------------------------------------------------------------------------
    // IEnumModesCallback implementation
    // -------------------------------------------------------------------------

    /**
     * Callback invoked by DirectDraw once per supported display mode during
     * {@link DirectDraw#enumDisplayModes}.  Operates in two distinct phases
     * controlled by whether {@link #modeBuffer} is {@code null}:
     *
     * <ul>
     *   <li><b>Counting pass</b> ({@code modeBuffer == null}): increments
     *       {@link #modeIndex} by 4 so that after the pass it holds
     *       {@code numModes * 4}.
     *   <li><b>Filling pass</b> ({@code modeBuffer != null}): writes
     *       {@code width}, {@code height}, {@code rgbBitCount}, and
     *       {@code refreshRate} into consecutive slots of {@link #modeBuffer}.
     * </ul>
     *
     * <p>The {@code userArg} parameter ({@link IUnknown}) is unused; DirectDraw
     * passes it through from the original {@code enumDisplayModes} call.
     *
     * <p>obf: {@code callbackEnumModes(DDSurfaceDesc, IUnknown)}
     */
    @Override
    public final void callbackEnumModes(DDSurfaceDesc modeDesc, IUnknown userArg) {
        if (modeBuffer == null) {
            // First pass: count each mode as 4 ints (width, height, bpp, refresh).
            modeIndex += 4;
            return;
        }

        // Second pass: store the four properties of this mode into the buffer.
        modeBuffer[modeIndex++] = modeDesc.width;
        modeBuffer[modeIndex++] = modeDesc.height;
        modeBuffer[modeIndex++] = modeDesc.rgbBitCount;
        modeBuffer[modeIndex++] = modeDesc.refreshRate;
    }

    // -------------------------------------------------------------------------
    // Public API
    // -------------------------------------------------------------------------

    /**
     * Enumerates all display modes supported by the DirectDraw primary adapter and
     * returns them as a flat {@code int[]} of consecutive
     * {@code {width, height, bitsPerPixel, refreshRate}} quads.
     *
     * <p>Uses a two-pass strategy so no extra data structures are needed:
     * <ol>
     *   <li>First {@code enumDisplayModes} call: callback runs in counting mode
     *       ({@link #modeBuffer}{@code == null}), leaving the total array length in
     *       {@link #modeIndex}.
     *   <li>Allocate {@link #modeBuffer} to that length and reset {@link #modeIndex}.
     *   <li>Second {@code enumDisplayModes} call: callback fills the buffer.
     * </ol>
     *
     * <p>The dummy {@code byte} parameter present in the obfuscated source
     * ({@code byte var1}) was an anti-tamper sentinel; its only use was the junk
     * expression {@code 55 % ((var1 - 22) / 53)}, which is computed but never
     * stored or returned.  It has been removed.
     *
     * @return flat int array of mode quads, length == {@code numModes * 4}.
     *
     * <p>obf: {@code a(byte)}
     */
    final int[] listModes() {
        // Pass 1: count how many modes DirectDraw reports.  modeBuffer is null so
        // the callback simply adds 4 to modeIndex for every mode it sees.
        directDraw.enumDisplayModes(0, null, null, this);

        // Allocate the exact buffer needed and reset the write cursor.
        modeBuffer = new int[modeIndex];
        modeIndex = 0;

        // Pass 2: fill the buffer with actual mode data.
        directDraw.enumDisplayModes(0, null, null, this);

        // Retrieve result and clear shared state so there are no stale references.
        int[] result = modeBuffer;
        modeBuffer = null;
        modeIndex = 0;

        return result;
    }

    /**
     * Restores the system to windowed (non-exclusive) DirectDraw cooperative level
     * and resets the display mode to its original resolution.
     *
     * <p>The {@code wasSetViaDirectDraw} parameter ({@code var2} in obf source) is
     * present in the bytecode but the non-zero branch calls
     * {@link #enterFullscreen(Frame, int, int, int, int)} with a frame of
     * {@code null} and a deliberately invalid anti-tamper byte (97 != 77), which
     * would have null-cleared {@code this.directDraw} before crashing.  That branch
     * is therefore unreachable dead code inserted by the obfuscator and is omitted
     * here; the real path is always {@link DirectDraw#restoreDisplayMode()} +
     * {@link DirectDraw#setCooperativeLevel(Frame, int)}.
     *
     * <p>{@code setCooperativeLevel(frame, 8)}: flag {@code 8} =
     * {@code DDSCL_NORMAL} — drops exclusive ownership, returns to normal windowed
     * cooperative mode.
     *
     * @param frame the AWT {@link Frame} whose cooperative level is being released.
     *
     * <p>obf: {@code a(Frame, int)}
     */
    final void exitFullscreen(Frame frame) {
        // Restore the original display mode that was active before enterFullscreen.
        directDraw.restoreDisplayMode();

        // Release exclusive cooperative level.
        // Flag 8 = DDSCL_NORMAL (windowed, non-exclusive).
        directDraw.setCooperativeLevel(frame, 8);
    }

    /**
     * Switches the given AWT {@link Frame} into exclusive DirectDraw fullscreen at
     * the specified display mode.
     *
     * <p>Steps performed:
     * <ol>
     *   <li>Makes the frame visible.
     *   <li>Retrieves the native Win32 HWND via {@link WComponentPeer#getHwnd()}.
     *   <li>Strips the window's decorations and forces it always-on-top by
     *       overwriting Win32 window styles with {@link User32#SetWindowLong}:
     *       <ul>
     *         <li>{@code GWL_STYLE (-16)} → {@code WS_POPUP (0x80000000)} —
     *             removes title bar, borders, caption, etc.
     *         <li>{@code GWL_EXSTYLE (-20)} → {@code WS_EX_TOPMOST (8)} —
     *             keeps the window above all others.
     *       </ul>
     *   <li>Sets DirectDraw cooperative level to {@code 17}
     *       ({@code DDSCL_EXCLUSIVE (1) | DDSCL_FULLSCREEN (16)}).
     *   <li>Calls {@link DirectDraw#setDisplayMode(int, int, int, int, int)} with
     *       the requested {@code (width, height, bpp, refreshRate, 0)}.
     *   <li>Resizes the AWT frame to cover the full screen and brings it to front.
     * </ol>
     *
     * <p>The obfuscated method accepted a dummy {@code byte var6} anti-tamper guard
     * ({@code if (var6 != 77) this.directDraw = null}); the caller in
     * {@code LoaderThread} always passes {@code (byte) 77} so the guard never
     * fires.  The parameter has been dropped.
     *
     * @param frame       the AWT {@link Frame} to make fullscreen.
     * @param bitsPerPixel colour depth in bits (e.g. 16 or 32).
     * @param refreshRate  desired refresh rate in Hz (e.g. 60).
     * @param height       desired display height in pixels (e.g. 480).
     * @param width        desired display width  in pixels (e.g. 640).
     *
     * <p>obf: {@code a(Frame, int, int, int, int, byte)}
     *        Note: obf param order was (frame, bpp, refreshRate, height, width, guard).
     */
    final void enterFullscreen(Frame frame, int bitsPerPixel, int refreshRate, int height, int width) {
        // Make the frame visible before manipulating its native window handle.
        frame.setVisible(true);

        // Retrieve the Win32 HWND through the Microsoft AWT peer.
        WComponentPeer peer = (WComponentPeer) frame.getPeer();
        int hwnd = peer.getHwnd();

        // Strip all window decorations by replacing GWL_STYLE with WS_POPUP.
        // Integer.MIN_VALUE == 0x80000000 == WS_POPUP.
        User32.SetWindowLong(hwnd, -16 /* GWL_STYLE */, Integer.MIN_VALUE /* WS_POPUP */);

        // Force always-on-top by setting GWL_EXSTYLE to WS_EX_TOPMOST (8).
        User32.SetWindowLong(hwnd, -20 /* GWL_EXSTYLE */, 8 /* WS_EX_TOPMOST */);

        // Acquire exclusive + fullscreen cooperative level.
        // 17 = DDSCL_EXCLUSIVE (1) | DDSCL_FULLSCREEN (16)
        directDraw.setCooperativeLevel(frame, 17);

        // Switch the hardware display mode.
        // API: setDisplayMode(width, height, bpp, refreshRate, flags)
        directDraw.setDisplayMode(width, height, bitsPerPixel, refreshRate, 0);

        // Resize the AWT frame to fill the entire screen and give it focus.
        frame.setBounds(0, 0, width, height);
        frame.toFront();
        frame.requestFocus();
    }
}
