package client.nativeapi;

import com.ms.awt.WComponentPeer;
import com.ms.dll.Callback;
import com.ms.dll.Root;
import com.ms.win32.User32;
import java.awt.Component;

/**
 * Win32MouseCallback — Windows WNDPROC subclass hook for cursor hide/show control.
 *
 * <p>This class installs itself as a Win32 window-procedure hook (via
 * {@code User32.SetWindowLong(hwnd, GWL_WNDPROC, this)}) on the AWT component window
 * that hosts the RSC game canvas. By intercepting {@code WM_SETCURSOR} messages it can
 * suppress the default Windows arrow cursor (returning {@code SetCursor(0)}), effectively
 * hiding the OS cursor while the game draws its own software cursor over the pixel buffer.
 * It also exposes a {@link #setCursorPosition} method so {@code LoaderThread} (class c)
 * can warp the real Windows cursor position via {@code User32.SetCursorPos}.
 *
 * <p>Lifetime: a single instance is created in {@code LoaderThread.run()} (class c, field t)
 * and kept alive for the session. {@link #installHook} is called whenever the target
 * AWT component changes or the cursor visibility flag needs updating.
 *
 * <p>Threading: {@code hookedHwnd} and {@code originalWndProc} are {@code volatile} and
 * the critical sections that swap the WNDPROC are {@code synchronized(this)} to prevent
 * races between the AWT event thread (which may call {@code installHook}) and the OS
 * callback thread (which calls {@link #callback}).
 *
 * <p>J++ / Microsoft VM specifics:
 * <ul>
 *   <li>{@code com.ms.dll.Callback} — Microsoft J++ base class that marshals a Java object
 *       as a native Win32 WNDPROC function pointer.</li>
 *   <li>{@code com.ms.dll.Root.alloc(Object)} — pins the Java object in the GC so the
 *       native side can hold the pointer without it being collected.</li>
 *   <li>{@code WComponentPeer.getTopHwnd()} — retrieves the Win32 HWND backing an AWT component.</li>
 * </ul>
 *
 * Obfuscated class: q
 */
final class Win32MouseCallback extends Callback {

    // -------------------------------------------------------------------------
    // Fields
    // -------------------------------------------------------------------------

    /**
     * HWND of the window currently being subclassed, or 0 if none, or -1 if invalidated.
     * Volatile because it is read/written from both the AWT thread and the OS callback thread.
     * obf: d
     */
    private volatile int hookedHwnd;

    /**
     * Original WNDPROC address saved when we subclassed the window (returned by
     * {@code SetWindowLong(hwnd, GWL_WNDPROC, this)}). Used to forward unhandled
     * messages via {@code CallWindowProc}.
     * Volatile for the same threading reason as {@link #hookedHwnd}.
     * obf: c
     */
    private volatile int originalWndProc;

    /**
     * HCURSOR handle for the standard arrow cursor ({@code IDC_ARROW = 0x7F00 = 32512}),
     * loaded once on first use via {@code User32.LoadCursor(0, IDC_ARROW)}.
     * When the game wants the cursor visible we restore this handle via {@code SetCursor};
     * when hidden we call {@code SetCursor(0)}.
     * obf: e
     */
    private int arrowCursorHandle;

    /**
     * One-shot initialisation flag. Set to {@code true} after {@link #arrowCursorHandle}
     * has been loaded and {@code Root.alloc(this)} has been called to pin the native pointer.
     * obf: a
     */
    private boolean initialized;

    /**
     * Whether the mouse cursor should be visible. {@code true} = show the OS arrow cursor;
     * {@code false} = hide it (pass {@code SetCursor(0)}). Initialised to {@code true} so
     * the cursor is visible before the hook is first installed.
     * obf: b
     */
    private volatile boolean cursorVisible = true;

    // -------------------------------------------------------------------------
    // Win32 constants
    // -------------------------------------------------------------------------

    /** Win32 {@code GWL_WNDPROC} index used with {@code SetWindowLong}/{@code GetWindowLong}. */
    private static final int GWL_WNDPROC = -4;

    /**
     * {@code WM_SETCURSOR} (0x0020 = 32). Sent by Windows whenever the cursor moves
     * into a window and Windows needs to decide which cursor shape to display.
     * The LOWORD of lParam contains the hit-test code.
     */
    private static final int WM_SETCURSOR = 0x0020; // 32

    /**
     * {@code HTCLIENT} (1). Hit-test code meaning the cursor is in the client area.
     * We only intercept {@code WM_SETCURSOR} when LOWORD(lParam) == HTCLIENT so we do
     * not suppress the resize/border cursors on window edges.
     */
    private static final int HTCLIENT = 1;

    /**
     * {@code IDC_ARROW} (0x7F00 = 32512). The OEM resource identifier for the standard
     * Windows arrow cursor; passed to {@code LoadCursor(null, IDC_ARROW)}.
     */
    private static final int IDC_ARROW = 32512; // 0x7F00

    /**
     * Application-defined or J++-internal message ID (0x18AA0 = 101024) sent to the
     * hooked window via {@code SendMessage} to force an immediate {@code WM_SETCURSOR}
     * re-evaluation after the hook is installed or removed. This is not a standard
     * {@code WM_*} value; it is likely a private application or VM message used as a
     * trigger signal.
     */
    private static final int MSG_FORCE_CURSOR_UPDATE = 101024; // 0x18AA0

    /**
     * Message ID (1) received in the callback that triggers hook cleanup. In the standard
     * Win32 API, message 1 is {@code WM_CREATE}; in the J++ WNDPROC callback chain this
     * may instead be a synthetic "detach / window destroyed" notification. When received
     * on the hooked window we clear {@link #hookedHwnd} and reset {@link #cursorVisible}
     * before forwarding to the original proc.
     *
     * <p><b>Uncertainty:</b> the precise semantic of this message ID in the J++ Callback
     * framework is not fully documented. It behaves as a teardown notification.
     */
    private static final int MSG_HOOK_CLEANUP = 1;

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    /** Default constructor; no initialisation beyond field defaults. obf: (implicit) */
    Win32MouseCallback() {
        // hookedHwnd = 0, originalWndProc = 0, initialized = false, cursorVisible = true
    }

    // -------------------------------------------------------------------------
    // Public/package API
    // -------------------------------------------------------------------------

    /**
     * Warps the Windows cursor to the given screen coordinates via
     * {@code User32.SetCursorPos(x, y)}.
     *
     * <p>Called by {@code LoaderThread} (class c) when it needs to programmatically
     * reposition the mouse pointer (e.g. to keep the OS cursor centred while the game
     * tracks relative movement).
     *
     * <p>Anti-tamper note: the obfuscated code checked {@code param1 != 23529} and, if
     * mismatched, called {@code this.callback(-56, 122, 69, -57)} as a side-effect before
     * continuing (it did NOT return early). This is dead anti-tamper / obfuscation noise;
     * the magic-value check and the bogus callback invocation have been removed. The
     * caller ({@code LoaderThread}) always passes 23529 anyway.
     *
     * @param _magic  dropped anti-tamper guard param (original value: 23529); ignored.
     *                obf: var1 / param1
     * @param screenY screen Y coordinate to move the cursor to. obf: var2 / param2
     * @param screenX screen X coordinate to move the cursor to. obf: var3 / param3
     *
     * obf: a(int,int,int)
     */
    final void setCursorPosition(int _magic, int screenY, int screenX) {
        User32.SetCursorPos(screenX, screenY);
    }

    /**
     * Installs (or updates) this callback as the Win32 WNDPROC hook on the given AWT
     * component, setting the cursor-visibility state to {@code wantCursorVisible}.
     *
     * <p>Algorithm:
     * <ol>
     *   <li>Resolve the Win32 HWND backing {@code component} via {@code WComponentPeer.getTopHwnd()}.</li>
     *   <li>Early-exit if the HWND is unchanged AND the requested cursor-visibility flag
     *       already matches the current state.</li>
     *   <li>If {@code hookType != GWL_WNDPROC} (i.e. caller is signalling "no hook"),
     *       invalidate the stored HWND so the next call forces a re-hook.</li>
     *   <li>On first call: load the arrow cursor handle and pin {@code this} in the GC
     *       via {@code Root.alloc(this)} so the native WNDPROC pointer remains valid.</li>
     *   <li>If the HWND has changed:
     *     <ul>
     *       <li>If there was a previous hook, show the cursor on the old window and
     *           restore its original WNDPROC (thread-safe under {@code synchronized}).</li>
     *       <li>Subclass the new HWND, storing the old WNDPROC in {@link #originalWndProc}
     *           (thread-safe under {@code synchronized}).</li>
     *     </ul>
     *   </li>
     *   <li>Update {@link #cursorVisible} and send {@link #MSG_FORCE_CURSOR_UPDATE} to the
     *       window to trigger an immediate {@code WM_SETCURSOR} re-evaluation.</li>
     * </ol>
     *
     * @param hookType        expected to be {@link #GWL_WNDPROC} (-4); any other value
     *                        forces an HWND reset. obf: param1 / n2
     * @param component       the AWT component whose Win32 peer HWND to hook.
     *                        obf: param2 / component
     * @param wantCursorVisible {@code true} to show the OS cursor; {@code false} to hide it.
     *                        obf: param3 / bl
     *
     * obf: a(int,Component,boolean)
     */
    final void installHook(int hookType, Component component, boolean wantCursorVisible) {
        WComponentPeer peer = (WComponentPeer) component.getPeer();
        int hwnd = peer.getTopHwnd();

        // Early-exit: if the target window hasn't changed AND cursor state is already right.
        if (hwnd == hookedHwnd) {
            // (!cursorVisible != wantCursorVisible) means states already match -> return.
            // Equivalent to: if (cursorVisible == wantCursorVisible) return;
            if (cursorVisible == wantCursorVisible) {
                return;
            }
        }

        // If hookType is anything other than GWL_WNDPROC, invalidate the stored HWND.
        // This forces a full re-hook on the next call that provides GWL_WNDPROC.
        if (hookType != GWL_WNDPROC) {
            hookedHwnd = -1;
        }

        // One-time init: load arrow cursor and pin this object against GC.
        if (!initialized) {
            arrowCursorHandle = User32.LoadCursor(0, IDC_ARROW);
            Root.alloc(this); // prevents GC of this Callback while native code holds the ptr
            initialized = true;
        }

        // If the target HWND has changed, swap the WNDPROC hook.
        if (hwnd != hookedHwnd) {
            if (hookedHwnd != 0) {
                // Detach from the previous window: first ensure its cursor is visible,
                // then restore the original WNDPROC under the lock.
                cursorVisible = true;
                // Notify the new window to update its cursor state during transition.
                User32.SendMessage(hwnd, MSG_FORCE_CURSOR_UPDATE, 0, 0);
                synchronized (this) {
                    // Restore the old window's original WNDPROC.
                    User32.SetWindowLong(hookedHwnd, GWL_WNDPROC, originalWndProc);
                }
            }
            // Install self as WNDPROC on the new window; save original proc for forwarding.
            synchronized (this) {
                hookedHwnd = hwnd;
                // SetWindowLong(hwnd, GWL_WNDPROC, this) — J++ overload accepts an Object
                // (Callback) and marshals it to a native function pointer.
                originalWndProc = User32.SetWindowLong(hookedHwnd, GWL_WNDPROC, (Object) this);
            }
        }

        // Apply the new cursor-visibility setting.
        cursorVisible = wantCursorVisible;
        // Force Windows to re-evaluate the cursor shape immediately.
        User32.SendMessage(hwnd, MSG_FORCE_CURSOR_UPDATE, 0, 0);
    }

    // -------------------------------------------------------------------------
    // com.ms.dll.Callback implementation
    // -------------------------------------------------------------------------

    /**
     * Win32 WNDPROC callback — invoked by the J++ native runtime whenever a Win32
     * window message is dispatched to the hooked window.
     *
     * <p>This method is {@code synchronized} to prevent concurrent access between the
     * OS callback thread and the AWT thread that may be calling {@link #installHook}.
     *
     * <p>Message handling:
     * <ul>
     *   <li>{@code hwnd != hookedHwnd}: we are not responsible for this window (can happen
     *       transiently during HWND swaps). Retrieve the real WNDPROC via
     *       {@code GetWindowLong} and delegate to it directly.</li>
     *   <li>{@link #WM_SETCURSOR} (0x20) with LOWORD(lParam) == {@link #HTCLIENT} (1):
     *       intercept cursor shape — show arrow or hide (SetCursor(0)) based on
     *       {@link #cursorVisible}. Return 0 (message handled).</li>
     *   <li>{@link #MSG_FORCE_CURSOR_UPDATE} (101024): synthetic trigger message; apply
     *       same cursor hide/show logic. Return 0.</li>
     *   <li>{@link #MSG_HOOK_CLEANUP} (1): hook teardown notification — clear
     *       {@link #hookedHwnd} and reset {@link #cursorVisible} to true, then forward
     *       to the original proc.</li>
     *   <li>All other messages: forward to {@link #originalWndProc} via
     *       {@code CallWindowProc}.</li>
     * </ul>
     *
     * @param hwnd   Win32 HWND of the window receiving the message.   obf: param1 / n2
     * @param msg    Win32 message identifier (WM_*).                   obf: param2 / n3
     * @param wParam WPARAM (message-specific).                         obf: param3 / n4
     * @param lParam LPARAM (message-specific; for WM_SETCURSOR, LOWORD = hit-test code).
     *               obf: param4 / n5
     * @return message result; 0 for handled cursor messages, or the result of
     *         {@code CallWindowProc} for all others.
     *
     * obf: callback(int,int,int,int)
     */
    @Override
    final synchronized int callback(int hwnd, int msg, int wParam, int lParam) {
        // If the message is for a window other than the one we hooked, look up the real
        // WNDPROC from the window itself and delegate — this handles transient races.
        if (hookedHwnd != hwnd) {
            int realProc = User32.GetWindowLong(hwnd, GWL_WNDPROC);
            return User32.CallWindowProc(realProc, hwnd, msg, wParam, lParam);
        }

        // WM_SETCURSOR (0x0020 = 32): intercept only when the cursor is in the client area.
        if (msg == WM_SETCURSOR) {
            int hitTestCode = lParam & 0xFFFF; // LOWORD(lParam) = hit-test result
            if (hitTestCode == HTCLIENT) {
                // Hide or show the OS cursor based on current game state.
                User32.SetCursor(cursorVisible ? arrowCursorHandle : 0);
                return 0; // message handled; suppress default cursor processing
            }
        }

        // MSG_FORCE_CURSOR_UPDATE (101024): internal trigger to refresh cursor state.
        if (msg == MSG_FORCE_CURSOR_UPDATE) {
            User32.SetCursor(cursorVisible ? arrowCursorHandle : 0);
            return 0;
        }

        // MSG_HOOK_CLEANUP (1): teardown / window detach notification from J++ runtime.
        // Clear state and reset to "cursor visible" before forwarding.
        if (msg == MSG_HOOK_CLEANUP) {
            hookedHwnd = 0;
            cursorVisible = true;
            // Fall through to CallWindowProc below.
        }

        // All other messages (including MSG_HOOK_CLEANUP after state reset):
        // forward to the original WNDPROC.
        return User32.CallWindowProc(originalWndProc, hwnd, msg, wParam, lParam);
    }
}
