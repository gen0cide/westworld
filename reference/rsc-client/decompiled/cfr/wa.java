/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.ms.awt.WComponentPeer
 *  com.ms.com.IUnknown
 *  com.ms.directX.DDSurfaceDesc
 *  com.ms.directX.DirectDraw
 *  com.ms.directX.IEnumModesCallback
 *  com.ms.win32.User32
 */
import com.ms.awt.WComponentPeer;
import com.ms.com.IUnknown;
import com.ms.directX.DDSurfaceDesc;
import com.ms.directX.DirectDraw;
import com.ms.directX.IEnumModesCallback;
import com.ms.win32.User32;
import java.awt.Component;
import java.awt.Frame;

final class wa
implements IEnumModesCallback {
    private static int[] b;
    private DirectDraw c = new DirectDraw();
    private static int a;

    public final void callbackEnumModes(DDSurfaceDesc dDSurfaceDesc, IUnknown iUnknown) {
        if (null == b) {
            a += 4;
        } else {
            wa.b[wa.a++] = dDSurfaceDesc.width;
            wa.b[wa.a++] = dDSurfaceDesc.height;
            wa.b[wa.a++] = dDSurfaceDesc.rgbBitCount;
            wa.b[wa.a++] = dDSurfaceDesc.refreshRate;
        }
    }

    final int[] a(byte by) {
        this.c.enumDisplayModes(0, null, null, (IEnumModesCallback)this);
        b = new int[a];
        a = 0;
        this.c.enumDisplayModes(0, null, null, (IEnumModesCallback)this);
        int[] nArray = b;
        b = null;
        a = 0;
        int n2 = 55 % ((by - 22) / 53);
        return nArray;
    }

    final void a(Frame frame, int n2) {
        if (n2 != 0) {
            this.a(null, 68, -102, 21, 73, (byte)97);
        }
        this.c.restoreDisplayMode();
        this.c.setCooperativeLevel((Component)frame, 8);
    }

    final void a(Frame frame, int n2, int n3, int n4, int n5, byte by) {
        if (by != 77) {
            this.c = null;
        }
        frame.setVisible(true);
        WComponentPeer wComponentPeer = (WComponentPeer)frame.getPeer();
        int n6 = wComponentPeer.getHwnd();
        User32.SetWindowLong((int)n6, (int)-16, (int)Integer.MIN_VALUE);
        User32.SetWindowLong((int)n6, (int)-20, (int)8);
        this.c.setCooperativeLevel((Component)frame, 17);
        this.c.setDisplayMode(n5, n4, n2, n3, 0);
        frame.setBounds(0, 0, n5, n4);
        frame.toFront();
        frame.requestFocus();
    }

    public wa() {
        this.c.initialize(null);
    }
}

