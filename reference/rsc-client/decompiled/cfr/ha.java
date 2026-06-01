/*
 * Decompiled with CFR 0.152.
 */
import java.awt.DisplayMode;
import java.awt.Frame;
import java.awt.GraphicsDevice;
import java.awt.GraphicsEnvironment;

final class ha {
    private DisplayMode a;
    private GraphicsDevice b;

    public final void exit() {
        if (null != this.a) {
            block6: {
                block5: {
                    this.b.setDisplayMode(this.a);
                    if (!this.b.getDisplayMode().equals(this.a)) break block5;
                    break block6;
                }
                throw new RuntimeException("");
            }
            this.a = null;
        }
        this.a(null, (byte)109);
    }

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     */
    private final void a(Frame frame, byte by) {
        this.b.setFullScreenWindow(frame);
        int n2 = -113 % ((by - -39) / 47);
    }

    public final int[] listmodes() {
        DisplayMode[] displayModeArray = this.b.getDisplayModes();
        int[] nArray = new int[displayModeArray.length << -1704625342];
        int n2 = 0;
        while (~displayModeArray.length < ~n2) {
            nArray[n2 << -186613982] = displayModeArray[n2].getWidth();
            nArray[(n2 << -970847806) - -1] = displayModeArray[n2].getHeight();
            nArray[(n2 << -921158878) - -2] = displayModeArray[n2].getBitDepth();
            nArray[(n2 << -2140589790) - -3] = displayModeArray[n2].getRefreshRate();
            ++n2;
        }
        return nArray;
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    public ha() throws Exception {
        GraphicsDevice[] graphicsDeviceArray;
        GraphicsEnvironment graphicsEnvironment = GraphicsEnvironment.getLocalGraphicsEnvironment();
        this.b = graphicsEnvironment.getDefaultScreenDevice();
        if (this.b.isFullScreenSupported()) return;
        GraphicsDevice[] graphicsDeviceArray2 = graphicsDeviceArray = graphicsEnvironment.getScreenDevices();
        int i2 = 0;
        while (graphicsDeviceArray2.length > i2) {
            GraphicsDevice graphicsDevice = graphicsDeviceArray2[i2];
            if (graphicsDevice != null && graphicsDevice.isFullScreenSupported()) {
                this.b = graphicsDevice;
                return;
            }
            ++i2;
        }
        throw new Exception();
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    public final void enter(Frame frame, int n2, int n3, int n4, int n5) {
        this.a = this.b.getDisplayMode();
        if (null == this.a) throw new NullPointerException();
        frame.setUndecorated(true);
        frame.enableInputMethods(false);
        this.a(frame, (byte)-93);
        if (n5 == 0) {
            int n6 = this.a.getRefreshRate();
            DisplayMode[] displayModeArray = this.b.getDisplayModes();
            boolean bl = false;
            for (int n7 = 0; n7 < displayModeArray.length; ++n7) {
                if (n2 != displayModeArray[n7].getWidth() || ~n3 != ~displayModeArray[n7].getHeight() || ~displayModeArray[n7].getBitDepth() != ~n4) continue;
                int n8 = displayModeArray[n7].getRefreshRate();
                if (bl && ~Math.abs(-n6 + n8) <= ~Math.abs(-n6 + n5)) continue;
                n5 = n8;
                bl = true;
            }
            if (!bl) {
                n5 = n6;
            }
        }
        this.b.setDisplayMode(new DisplayMode(n2, n3, n4, n5));
        return;
    }
}

