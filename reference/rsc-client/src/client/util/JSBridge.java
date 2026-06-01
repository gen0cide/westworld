package client.util;

import java.applet.Applet;
import netscape.javascript.JSObject;

/**
 * JSBridge — thin static bridge from the game client into the browser JavaScript environment.
 *
 * In the Microsoft J++ / Netscape plugin era the live Applet object can reach its host page's
 * JavaScript runtime through {@code netscape.javascript.JSObject.getWindow(applet)}.  This class
 * wraps that single entry-point so the rest of the client has a stable, named call site instead of
 * scattering raw JSObject calls throughout the codebase.
 *
 * Typical usage inside Mudclient (class "client"): report session events or errors to the page, e.g.
 *   {@code JSBridge.call("worldLogin", applet)}
 *
 * The original obfuscated class name was "a" (default package).  It contained exactly one method,
 * also named "a", plus a dead anti-tamper dummy parameter and a dead junk-modulo expression.
 *
 * obf: a
 */
final class JSBridge {

    // Private constructor — all members are static; this class is never instantiated.
    private JSBridge() {}

    /**
     * Invoke a named JavaScript function in the browser page that is hosting this applet.
     *
     * Equivalent to the JavaScript expression {@code window[methodName]()}.  No arguments are
     * passed to the JS function ({@code null} argument array).
     *
     * The return value is whatever the JSObject bridge deserialises from the JavaScript return
     * value (typically a {@code String}, {@code Double}, or {@code null}).
     *
     * @param methodName  name of the JavaScript function to call on the browser window object
     * @param applet      the live applet instance used to obtain the JSObject window handle
     * @return            the JavaScript return value, or {@code null}
     * @throws Throwable  any exception propagated from the JS engine or JSObject layer
     *
     * obf: a(String, byte, Applet)
     *
     * STRIPPED (anti-tamper): the original had a dummy {@code byte} parameter (obf name var1 / "by")
     *   whose sole use was the dead junk expression {@code int var3 = -38 % ((var1 - 14) / 53);}.
     *   The result was never read.  The parameter and expression are removed; only the two
     *   semantically meaningful parameters remain.
     *
     * STRIPPED (exception wrapper): Vineflower showed a try/catch that merely re-threw
     *   RuntimeException — the standard per-method obfuscation wrapper.  Unwrapped to bare body.
     */
    static Object call(String methodName, Applet applet) throws Throwable {
        // Obtain the JavaScript window object for the page hosting this applet, then invoke the
        // named function with no arguments.  This is the only real logic in the original class.
        return JSObject.getWindow(applet).call(methodName, null);
    }
}
