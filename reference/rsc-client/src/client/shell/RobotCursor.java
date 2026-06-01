package client.shell;

import java.awt.Component;
import java.awt.Point;
import java.awt.Robot;
import java.awt.image.BufferedImage;

/**
 * RobotCursor — AWT robot mouse mover and custom-cursor controller.
 *
 * <p>Wraps {@link java.awt.Robot} for programmatic mouse positioning and provides two cursor
 * management helpers used by the engine shell:
 *
 * <ul>
 *   <li><b>hideCursor / showCursor</b> ({@link #setCursorVisibility}) — installs a 1×1
 *       transparent {@link BufferedImage} as a custom cursor to hide the OS pointer over the game
 *       canvas, or restores the default OS cursor.  Tracks the currently-hidden component in
 *       {@link #cursorHiddenComponent} so that re-showing is always paired correctly.
 *   <li><b>setCustomCursor</b> ({@link #setCustomCursor}) — uploads an ARGB pixel array as a
 *       custom cursor image (e.g. for a crosshair or item cursor) with a caller-supplied hotspot.
 *       Passing {@code null} for the pixel array resets to the default OS cursor.
 *   <li><b>moveMouse</b> ({@link #moveMouse}) — delegates to {@link Robot#mouseMove} so that the
 *       game engine can warp the physical pointer (e.g. to re-center the mouse after a camera drag).
 * </ul>
 *
 * <p>This class was added for the rev-233+ Microsoft J++ build and has no direct equivalent in the
 * rev-204 oracle; cursor hiding in rev 204 is handled inline inside {@code GameShell}.  The
 * {@code Robot} instance requires the AWT event-permission grant that the signed applet provides.
 *
 * <p>Obfuscated class: {@code j} (default package, Microsoft J++ JAR rev ~233-235).
 *
 * <p>Obfuscation artifacts stripped:
 * <ul>
 *   <li>All {@code try/catch(RuntimeException)} wrappers that only rethrow — unwrapped to bare body.
 *   <li>No opaque-predicate, profiling counter, or XOR string pool was present in this class.
 * </ul>
 */
final class RobotCursor { // obf: j

    // -------------------------------------------------------------------------
    // Fields
    // -------------------------------------------------------------------------

    /**
     * AWT {@link Robot} used for programmatic mouse movement.
     * Initialised once in the constructor; requires the security permission granted to signed applets.
     * obf: a
     */
    private Robot robot; // obf: a

    /**
     * The {@link Component} over which the OS cursor is currently hidden, or {@code null} if the
     * cursor is currently visible.  Tracked so that {@link #setCursorVisibility} can restore the
     * default cursor on the correct component before switching to a new one.
     * obf: b
     */
    private Component cursorHiddenComponent; // obf: b

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /**
     * Constructs a new {@code RobotCursor}, creating the underlying {@link Robot} instance.
     *
     * @throws Exception if {@link Robot} cannot be created (e.g. security restrictions in
     *                   unsigned applets, or a headless environment).
     */
    public RobotCursor() throws Exception {
        // Robot() itself throws AWTException (checked), which propagates as-is.
        this.robot = new Robot();
    }

    // -------------------------------------------------------------------------
    // Public API
    // -------------------------------------------------------------------------

    /**
     * Shows or hides the OS cursor over a given AWT component.
     *
     * <p>When {@code hide} is {@code true} the component parameter is ignored and the cursor is
     * revealed on whichever component it was previously hidden over (i.e. this acts as "show
     * cursor").  When {@code hide} is {@code false} the cursor is hidden over {@code component} by
     * installing a 1×1 transparent custom cursor.
     *
     * <p>If the target component is already the one with the hidden cursor, the call is a no-op.
     *
     * <p>The transparent cursor is created via
     * {@code createCustomCursor(new BufferedImage(1,1,TYPE_INT_ARGB), new Point(0,0), null)}.
     * A 1×1 {@link BufferedImage} of type {@code 2} ({@code TYPE_INT_ARGB}) is fully transparent
     * by default, which effectively hides the pointer.
     *
     * <p>Parameter semantics (note: the obfuscated parameter names are reversed from what you might
     * expect — {@code var2=true} means <em>show</em> the cursor, not hide it):
     * <pre>
     *   setCursorVisibility(canvas, false)  → hide cursor over canvas
     *   setCursorVisibility(null,   true)   → show cursor (component param ignored)
     * </pre>
     *
     * obf: showcursor(Component var1, boolean var2)
     *
     * @param component the AWT component to hide the cursor over; ignored when {@code show} is {@code true}.
     * @param show      {@code true} to restore the OS cursor; {@code false} to hide it over {@code component}.
     * @throws NullPointerException if {@code show} is {@code false} and {@code component} is {@code null}.
     */
    public final void setCursorVisibility(Component component, boolean show) { // obf: showcursor
        // When show=true, we want to reveal the cursor: treat the target as null so any
        // existing hidden-component has its default cursor restored below.
        if (show) {
            component = null;
        } else {
            // Fail fast — hiding the cursor on null makes no sense.
            if (component == null) {
                throw new NullPointerException();
            }
        }

        // No change needed — the cursor is already in the desired state for this component.
        if (component == this.cursorHiddenComponent) {
            return;
        }

        // Restore the default OS cursor on the component that was previously hidden.
        if (this.cursorHiddenComponent != null) {
            this.cursorHiddenComponent.setCursor(null); // null → default system cursor
            this.cursorHiddenComponent = null;
        }

        // Install a 1×1 fully-transparent custom cursor to hide the pointer.
        if (component != null) {
            // BufferedImage type 2 = TYPE_INT_ARGB; a fresh 1×1 image has all pixels 0x00000000
            // (fully transparent), which renders as an invisible cursor.
            // Hotspot (0,0): the invisible "click point" is at the top-left — doesn't matter
            // since no pixel is visible.
            component.setCursor(
                component.getToolkit().createCustomCursor(
                    new BufferedImage(1, 1, BufferedImage.TYPE_INT_ARGB), // 1×1 transparent
                    new Point(0, 0),  // hotspot
                    null              // cursor name (not needed for a blank cursor)
                )
            );
            this.cursorHiddenComponent = component;
        }
    }

    /**
     * Installs an ARGB pixel array as a custom cursor on a component, or resets to the default
     * OS cursor.
     *
     * <p>The pixel array is written into a {@link BufferedImage} of type {@code TYPE_INT_ARGB}
     * (type id {@code 2}) via {@link BufferedImage#setRGB} with scanline stride equal to
     * {@code width}, then passed to {@link java.awt.Toolkit#createCustomCursor} with the supplied
     * hotspot {@link Point}.
     *
     * <p>Typical use: the engine passes a crosshair or item sprite's pixel data so the game
     * pointer reflects the currently selected item or action.
     *
     * obf: setcustomcursor(Component var1, int[] var2, int var3, int var4, Point var5)
     *
     * @param component the component on which to set the cursor.
     * @param pixels    ARGB pixel data (row-major, {@code width} pixels per row); pass {@code null}
     *                  to reset to the default OS cursor.
     * @param width     cursor image width in pixels.
     * @param height    cursor image height in pixels.
     * @param hotspot   the cursor hotspot (the pixel within the image that acts as the pointer tip).
     */
    public final void setCustomCursor(Component component, int[] pixels, int width, int height, Point hotspot) { // obf: setcustomcursor
        if (pixels != null) {
            // Build a BufferedImage from the caller-supplied ARGB pixel array.
            // TYPE_INT_ARGB = 2; scanline stride = width (no padding).
            BufferedImage cursorImage = new BufferedImage(width, height, BufferedImage.TYPE_INT_ARGB);
            cursorImage.setRGB(
                0, 0,       // destination origin
                width,      // scanline width
                height,     // scanline count
                pixels,     // source ARGB pixels
                0,          // source array offset
                width       // source scanline stride
            );
            component.setCursor(
                component.getToolkit().createCustomCursor(cursorImage, hotspot, null)
            );
        } else {
            // null pixel array → restore the default OS cursor.
            component.setCursor(null);
        }
    }

    /**
     * Moves the physical OS mouse pointer to the given screen coordinates using {@link Robot}.
     *
     * <p>Used by the engine to warp the pointer — e.g. re-centering after a mouse-drag camera
     * rotation so the next delta reads from the centre of the window.
     *
     * obf: movemouse(int var1, int var2)
     *
     * @param screenX target X coordinate in screen (not component) space.
     * @param screenY target Y coordinate in screen (not component) space.
     */
    public final void moveMouse(int screenX, int screenY) { // obf: movemouse
        this.robot.mouseMove(screenX, screenY);
    }
}
