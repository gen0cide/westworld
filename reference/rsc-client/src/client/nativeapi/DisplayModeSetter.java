package client.nativeapi;

import java.awt.DisplayMode;
import java.awt.Frame;
import java.awt.GraphicsDevice;
import java.awt.GraphicsEnvironment;

/**
 * DisplayModeSetter — AWT fullscreen display-mode switcher.
 *
 * Wraps a {@link GraphicsDevice} that supports exclusive fullscreen
 * ({@code GraphicsDevice.isFullScreenSupported()}).  The engine calls
 * {@link #enterFullscreen} to switch the game frame into exclusive
 * fullscreen at the requested resolution / bit-depth / refresh rate,
 * and {@link #exitFullscreen} to restore the desktop's original display
 * mode and release the exclusive window.
 *
 * Only used on platforms where the Microsoft JVM's DirectDraw path is
 * NOT available (i.e. non-Windows, or when DirectDraw init fails); on
 * Windows the {@code wa} (DirectDrawModes) class handles fullscreen
 * instead.  The two share the same public API surface expected by
 * {@code client} (Mudclient).
 *
 * obf: ha
 */
final class DisplayModeSetter {

    /**
     * The desktop's display mode captured at {@link #enterFullscreen} time;
     * used by {@link #exitFullscreen} to restore the original mode.
     * {@code null} when not in fullscreen.
     *
     * obf: a
     */
    private DisplayMode previousDisplayMode;

    /**
     * The AWT {@link GraphicsDevice} that supports exclusive fullscreen.
     * Resolved in the constructor: first tries the default screen device,
     * then falls back to scanning all screen devices.
     *
     * obf: b
     */
    private GraphicsDevice graphicsDevice;

    // -----------------------------------------------------------------------
    // Construction
    // -----------------------------------------------------------------------

    /**
     * Finds a fullscreen-capable {@link GraphicsDevice} on the local machine.
     *
     * Tries the default screen device first; if that device does not support
     * exclusive fullscreen ({@code isFullScreenSupported()} returns false),
     * scans all screen devices and picks the first one that does.  Throws
     * {@link Exception} if no fullscreen-capable device is found at all.
     *
     * obf: ha()
     *
     * @throws Exception if no screen device supports exclusive fullscreen.
     */
    public DisplayModeSetter() throws Exception {
        GraphicsEnvironment env = GraphicsEnvironment.getLocalGraphicsEnvironment();
        this.graphicsDevice = env.getDefaultScreenDevice();

        // Fast path: default screen already supports fullscreen.
        if (this.graphicsDevice.isFullScreenSupported()) {
            return;
        }

        // Slow path: scan all screen devices for one that supports fullscreen.
        GraphicsDevice[] devices = env.getScreenDevices();
        for (int i = 0; i < devices.length; i++) {
            GraphicsDevice dev = devices[i];
            if (dev != null && dev.isFullScreenSupported()) {
                this.graphicsDevice = dev;
                return;
            }
        }

        // No fullscreen-capable device found — caller must fall back.
        throw new Exception();
    }

    // -----------------------------------------------------------------------
    // Public API
    // -----------------------------------------------------------------------

    /**
     * Enters exclusive fullscreen at the specified resolution, bit depth, and
     * refresh rate.
     *
     * Steps performed:
     * <ol>
     *   <li>Saves the current desktop {@link DisplayMode} for later
     *       restoration by {@link #exitFullscreen}.</li>
     *   <li>Makes the frame undecorated and disables input methods
     *       (required by AWT before handing the window to the graphics
     *       device).</li>
     *   <li>Calls {@link GraphicsDevice#setFullScreenWindow} with the
     *       supplied frame (via the private helper).</li>
     *   <li>If {@code desiredRefreshRate} is 0, searches the device's
     *       available display modes for the one whose refresh rate is
     *       closest to the current desktop rate at the requested
     *       width/height/bitDepth.  Otherwise uses the caller-supplied
     *       rate directly.</li>
     *   <li>Calls {@link GraphicsDevice#setDisplayMode} with the
     *       resolved parameters.</li>
     * </ol>
     *
     * obf: enter
     *
     * @param frame            the AWT Frame to make fullscreen.
     * @param width            desired display width in pixels.
     * @param height           desired display height in pixels.
     * @param bitDepth         desired color bit depth (e.g. 16 or 32).
     * @param desiredRefreshRate desired refresh rate in Hz, or 0 to
     *                         auto-select the rate closest to the
     *                         current desktop rate.
     */
    public final void enterFullscreen(Frame frame, int width, int height,
                                      int bitDepth, int desiredRefreshRate) {
        // Save the desktop display mode so exitFullscreen() can restore it.
        this.previousDisplayMode = this.graphicsDevice.getDisplayMode();
        if (this.previousDisplayMode == null) {
            // Should never happen on a properly initialised GraphicsDevice,
            // but guard defensively — the original bytecode throws NPE here.
            throw new NullPointerException();
        }

        // AWT requires the frame to be undecorated before it can be made
        // the full-screen window.
        frame.setUndecorated(true);
        frame.enableInputMethods(false);

        // Hand the frame to the graphics device as the full-screen window.
        // The private helper strips the junk modulo from the obfuscator.
        this.setFullScreenWindow(frame);

        // --- Refresh-rate selection ----------------------------------------
        // If the caller passes 0 for the refresh rate, scan the device's
        // available display modes and pick the one whose refresh rate is
        // closest to the current desktop refresh rate.
        if (desiredRefreshRate == 0) {
            int currentRefreshRate = this.previousDisplayMode.getRefreshRate();
            DisplayMode[] availableModes = this.graphicsDevice.getDisplayModes();
            boolean foundMatch = false;

            for (int i = 0; i < availableModes.length; i++) {
                DisplayMode mode = availableModes[i];

                // Only consider modes that match the requested width, height,
                // and bit depth exactly.
                if (mode.getWidth()    != width    ) continue;
                if (mode.getHeight()   != height   ) continue;
                if (mode.getBitDepth() != bitDepth ) continue;

                int candidateRate = mode.getRefreshRate();

                // Pick this candidate if we haven't found anything yet, or if
                // it is closer to the current desktop rate than the best so far.
                // ~abs(x) > ~abs(y)  ↔  abs(x) < abs(y) — see bytecode.
                boolean candidateIsCloser =
                    Math.abs(candidateRate    - currentRefreshRate) <
                    Math.abs(desiredRefreshRate - currentRefreshRate);

                if (!foundMatch || candidateIsCloser) {
                    desiredRefreshRate = candidateRate;
                    foundMatch = true;
                }
            }

            // No matching mode found for the requested w/h/depth — fall back
            // to the current desktop refresh rate.
            if (!foundMatch) {
                desiredRefreshRate = currentRefreshRate;
            }
        }

        // Apply the display mode.
        this.graphicsDevice.setDisplayMode(
            new DisplayMode(width, height, bitDepth, desiredRefreshRate));
    }

    /**
     * Exits exclusive fullscreen and restores the desktop display mode.
     *
     * If a saved display mode exists (i.e. {@link #enterFullscreen} was
     * previously called), restores it via {@link GraphicsDevice#setDisplayMode}
     * and verifies the switch succeeded (throws {@link RuntimeException} if
     * the mode after the call does not equal the target).  Then clears the
     * saved mode and releases the full-screen window by calling
     * {@link GraphicsDevice#setFullScreenWindow} with {@code null}.
     *
     * obf: exit
     */
    public final void exitFullscreen() {
        // Only restore if we actually entered fullscreen.
        if (this.previousDisplayMode != null) {
            this.graphicsDevice.setDisplayMode(this.previousDisplayMode);

            // Verify the mode change actually took effect.
            if (!this.graphicsDevice.getDisplayMode().equals(this.previousDisplayMode)) {
                // The original bytecode throws new RuntimeException("") here.
                throw new RuntimeException("Failed to restore display mode");
            }

            this.previousDisplayMode = null;
        }

        // Release the exclusive fullscreen window (pass null to unset it).
        this.setFullScreenWindow(null);
    }

    /**
     * Returns a flat {@code int[]} describing every display mode supported by
     * the device.  Each group of four consecutive elements encodes one mode:
     * <pre>
     *   [i*4 + 0]  width       (pixels)
     *   [i*4 + 1]  height      (pixels)
     *   [i*4 + 2]  bit depth
     *   [i*4 + 3]  refresh rate (Hz)
     * </pre>
     * The array length is therefore {@code numModes * 4}.
     *
     * obf: listmodes
     *
     * @return flat array of display mode parameters, 4 ints per mode.
     */
    public final int[] listAvailableModes() {
        DisplayMode[] modes = this.graphicsDevice.getDisplayModes();
        // Allocate 4 slots per mode.  The original uses "<< 2" (× 4).
        int[] result = new int[modes.length << 2];

        for (int i = 0; i < modes.length; i++) {
            result[i * 4    ] = modes[i].getWidth();
            result[i * 4 + 1] = modes[i].getHeight();
            result[i * 4 + 2] = modes[i].getBitDepth();
            result[i * 4 + 3] = modes[i].getRefreshRate();
        }

        return result;
    }

    // -----------------------------------------------------------------------
    // Private helpers
    // -----------------------------------------------------------------------

    /**
     * Thin wrapper around {@link GraphicsDevice#setFullScreenWindow}.
     *
     * The original obfuscated method carried a dummy {@code byte} parameter
     * and a dead junk-modulo expression ({@code -113 % ((by - -39) / 47)})
     * that is a pure anti-tamper artifact — the result is never used.  Both
     * have been removed here.
     *
     * Called with the game {@link Frame} to enter fullscreen, or {@code null}
     * to release exclusive fullscreen mode.
     *
     * obf: a(Frame, byte) — dummy byte param dropped (anti-tamper artifact)
     *
     * @param frame the frame to make fullscreen, or {@code null} to release.
     */
    private void setFullScreenWindow(Frame frame) {
        this.graphicsDevice.setFullScreenWindow(frame);
        // Dead junk: int junk = -113 % ((by + 39) / 47);  ← removed (anti-tamper)
    }
}
