/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  java.applet.Applet
 *  netscape.javascript.JSObject
 */
import java.applet.Applet;
import netscape.javascript.JSObject;

final class a {
    static final Object a(String string, byte by, Applet applet) throws Throwable {
        int n2 = -38 % ((by - 14) / 53);
        return JSObject.getWindow((Applet)applet).call(string, null);
    }
}

