/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  java.applet.Applet
 *  java.applet.AppletContext
 */
import java.applet.Applet;
import java.applet.AppletContext;
import java.awt.Color;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics;
import java.awt.Image;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.awt.image.ImageObserver;
import java.io.IOException;
import java.lang.reflect.Method;
import java.net.URL;

public class e
extends Applet
implements Runnable,
MouseListener,
MouseMotionListener,
KeyListener {
    int K;
    static int y;
    private int n = 1;
    long[] F = new long[10];
    static int Nb;
    static byte[][] kb;
    String p = null;
    static int G;
    static int A;
    static int L;
    static int Y;
    static int c;
    static int Db;
    static int W;
    static int r;
    static int Rb;
    static int qb;
    static int Fb;
    static int cb;
    static String[] Mb;
    static int Pb;
    static int jb;
    Thread z = null;
    static int Lb;
    private int Ib = 20;
    static int pb;
    Font tb;
    static int Gb;
    private boolean hb = false;
    static int j;
    static int ab;
    static int ub;
    static int D;
    String B = Sb[44];
    static int rb;
    private int a = 384;
    boolean N = false;
    static int Hb;
    static int eb;
    static int[] nb;
    static int k;
    static int q;
    static int f;
    static int w;
    int Eb;
    static v i;
    static int t;
    static int h;
    int sb = 0;
    static int P;
    static int M;
    private int m = 512;
    private int S = 1000;
    static int g;
    static int mb;
    private int b = 0;
    static int fb;
    static int R;
    private int vb = 0;
    static int s;
    static int ob;
    static int d;
    private int V = 0;
    static int O;
    static int yb;
    static int o;
    static int[] wb;
    static int J;
    static int l;
    static int db;
    Font X;
    Font Jb;
    Graphics u;
    Image C;
    int I = 0;
    String e = "";
    private boolean Kb = false;
    int xb = 0;
    String x = "";
    boolean U = false;
    int Bb = 0;
    boolean gb = false;
    String Cb = "";
    boolean E = false;
    int Q = 1;
    boolean bb = false;
    boolean Z = false;
    int Qb = 0;
    String Ob = "";
    public static int Ab;
    public static boolean T;
    public static boolean H;
    public static int zb;
    public static int v;
    public static boolean ib;
    public static int lb;
    private static final String[] Sb;

    public final String getParameter(String string) {
        try {
            ++j;
            if (null != kb.a) {
                return null;
            }
            if (da.gb != null) {
                return da.gb.getParameter(string);
            }
            return super.getParameter(string);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[21]);
            String string2 = string != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(')').toString());
        }
    }

    private final void b(int n2) {
        block6: {
            try {
                block5: {
                    ++fb;
                    this.vb = -2;
                    System.out.println(Sb[28]);
                    this.a(false);
                    mb.a(11200, 1000L);
                    if (n2 != 100) {
                        this.e(27);
                    }
                    if (kb.a != null) break block5;
                    break block6;
                }
                kb.a.dispose();
                System.exit(0);
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, Sb[27] + n2 + ')');
            }
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    @Override
    public final void run() {
        boolean bl = client.vh;
        try {
            ++qb;
            try {
                int n2;
                int n3;
                block45: {
                    int n4;
                    block48: {
                        block49: {
                            block47: {
                                if (1 == this.n) {
                                    int n5;
                                    block44: {
                                        this.n = 2;
                                        while (!this.isDisplayable()) {
                                            n5 = this.vb;
                                            if (!bl) {
                                                if (n5 < 0) break;
                                                if (-1 > ~this.vb) {
                                                    --this.vb;
                                                    if (-1 == ~this.vb) {
                                                        this.b(100);
                                                        this.z = null;
                                                        return;
                                                    }
                                                }
                                                mb.a(11200, (long)this.Ib);
                                                if (!bl) continue;
                                            }
                                            break block44;
                                        }
                                        n5 = 0;
                                    }
                                    if (n5 > this.vb) {
                                        if (this.vb == -1) {
                                            this.b(100);
                                        }
                                        this.z = null;
                                        return;
                                    }
                                    if (this.b((byte)118)) {
                                        this.a((byte)-92);
                                        this.n = 0;
                                    } else {
                                        if (1 != ~this.vb) {
                                            this.b(100);
                                        }
                                        this.z = null;
                                        return;
                                    }
                                }
                                if (null == kb.a) break block47;
                                kb.a.addMouseListener(this);
                                kb.a.addMouseMotionListener(this);
                                kb.a.addKeyListener(this);
                                if (!bl) break block48;
                            }
                            if (null != da.gb) break block49;
                            this.addMouseListener(this);
                            this.addMouseMotionListener(this);
                            this.addKeyListener(this);
                            if (!bl) break block48;
                        }
                        da.gb.addMouseListener((MouseListener)this);
                        da.gb.addMouseMotionListener((MouseMotionListener)this);
                        da.gb.addKeyListener((KeyListener)this);
                    }
                    int n6 = 0;
                    int n7 = 256;
                    int n8 = 1;
                    int n9 = 0;
                    for (n4 = 0; n4 < 10; ++n4) {
                        this.F[n4] = p.a(0);
                        if (bl) return;
                        if (!bl) continue;
                    }
                    long l2 = p.a(0);
                    while (0 <= this.vb) {
                        int n10;
                        int n11;
                        block46: {
                            int n12;
                            int n13;
                            block51: {
                                block50: {
                                    n3 = -1;
                                    n2 = ~this.vb;
                                    if (bl) break block45;
                                    if (n3 > n2) {
                                        --this.vb;
                                        if (0 == this.vb) {
                                            this.b(100);
                                            this.z = null;
                                            return;
                                        }
                                    }
                                    n4 = n7;
                                    n7 = 300;
                                    int n14 = n8;
                                    n8 = 1;
                                    l2 = p.a(0);
                                    if ((this.F[n6] ^ 0xFFFFFFFFFFFFFFFFL) != -1L) break block50;
                                    n8 = n14;
                                    n7 = n4;
                                    if (!bl) break block51;
                                }
                                if ((l2 ^ 0xFFFFFFFFFFFFFFFFL) < (this.F[n6] ^ 0xFFFFFFFFFFFFFFFFL)) {
                                    n7 = (int)((long)(this.Ib * 2560) / (l2 + -this.F[n6]));
                                }
                            }
                            if (~n7 > -26) {
                                n7 = 25;
                            }
                            if (-257 > ~n7) {
                                n7 = 256;
                                n8 = (int)(-((-this.F[n6] + l2) / 10L) + (long)this.Ib);
                                if (n8 < this.Q) {
                                    n8 = this.Q;
                                }
                            }
                            mb.a(11200, (long)n8);
                            this.F[n6] = l2;
                            if (-2 > ~n8) {
                                for (n13 = 0; 10 > n13; ++n13) {
                                    long l3 = -1L - (this.F[n13] ^ 0xFFFFFFFFFFFFFFFFL);
                                    n12 = l3 == 0L ? 0 : (l3 < 0L ? -1 : 1);
                                    if (!bl) {
                                        if (n12 == 0) continue;
                                        int n15 = n13;
                                        this.F[n15] = this.F[n15] + (long)n8;
                                        if (!bl) continue;
                                    }
                                    break;
                                }
                            } else {
                                n6 = (1 + n6) % 10;
                                n12 = n13 = 0;
                            }
                            while (-257 < ~n9) {
                                this.e(119);
                                n9 += n7;
                                if (~this.S <= ~(++n13)) continue;
                                n9 = 0;
                                this.b += 6;
                                n11 = 25;
                                n10 = this.b;
                                if (!bl) {
                                    if (n11 >= n10) break;
                                    this.b = 0;
                                    this.U = true;
                                    break;
                                }
                                break block46;
                            }
                            --this.b;
                            n11 = n9;
                            n10 = 255;
                        }
                        n9 = n11 & n10;
                        this.b(false);
                        if (!bl) continue;
                    }
                    n3 = -1;
                    n2 = this.vb;
                }
                if (n3 == n2) {
                    this.b(100);
                }
                this.z = null;
                return;
            }
            catch (Exception exception) {
                mb.a(0x1FFFFF, exception, null);
                this.a(Sb[12], true);
                return;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[11]);
        }
    }

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    @Override
    public final synchronized void keyReleased(KeyEvent keyEvent) {
        try {
            this.a(keyEvent, (byte)-128);
            ++w;
            char c2 = keyEvent.getKeyChar();
            int n2 = keyEvent.getKeyCode();
            if (~((char)c2) != '\uffffffdf') {
                // empty if block
            }
            if (40 == n2) {
                // empty if block
            }
            if ('N' != (char)c2 && 'M' != (char)c2) {
                // empty if block
            }
            if (~n2 == -40) {
                this.E = false;
            }
            if ((char)c2 != 'n' && 'm' != (char)c2) {
                // empty if block
            }
            if ((char)c2 != '{') {
                // empty if block
            }
            if (n2 == 37) {
                this.Z = false;
            }
            if (38 == n2) {
                // empty if block
            }
            if (-126 != ~((char)c2)) return;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[58]);
            String string = keyEvent != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    synchronized void b(boolean bl) {
        try {
            if (bl) {
                return;
            }
            ++M;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[48] + bl + ')');
        }
    }

    synchronized void e(int n2) {
        try {
            if (n2 < 64) {
                return;
            }
            ++ob;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[49] + n2 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    public final void paint(Graphics graphics) {
        try {
            this.N = true;
            ++l;
            if (-3 != ~this.n || this.C == null) {
                if (0 != this.n) {
                    return;
                }
                this.a(-89);
                if (!client.vh) return;
            }
            this.a(this.B, this.V, 126);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[10]);
            if (graphics != null) {
                string = Sb[4];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = Sb[3];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(String string, int n2, int n3) {
        try {
            try {
                int n4;
                int n5;
                block23: {
                    block22: {
                        n5 = (-281 + this.m) / 2;
                        n4 = (-148 + this.a) / 2;
                        this.u.setColor(Color.black);
                        this.u.fillRect(0, 0, this.m, this.a);
                        if (!this.hb) {
                            this.u.drawImage(this.C, n5, n4, (ImageObserver)((Object)this));
                        }
                        n5 += 2;
                        this.V = n2;
                        n4 += 90;
                        this.B = string;
                        if (n3 <= 97) {
                            this.mouseReleased(null);
                        }
                        this.u.setColor(new Color(132, 132, 132));
                        if (this.hb) {
                            this.u.setColor(new Color(220, 0, 0));
                        }
                        this.u.drawRect(n5 + -2, n4 + -2, 280, 23);
                        this.u.fillRect(n5, n4, 277 * n2 / 100, 20);
                        this.u.setColor(new Color(198, 198, 198));
                        if (this.hb) {
                            this.u.setColor(new Color(255, 255, 255));
                        }
                        this.a(this.tb, string, 10 + n4, true, 138 + n5, this.u);
                        if (this.hb) break block22;
                        this.a(this.X, Sb[24], 30 + n4, true, n5 - -138, this.u);
                        this.a(this.X, Sb[23], n4 + 44, true, n5 - -138, this.u);
                        if (!client.vh) break block23;
                    }
                    this.u.setColor(new Color(132, 132, 152));
                    this.a(this.Jb, Sb[23], -20 + this.a, true, 138 + n5, this.u);
                }
                if (null != this.p) {
                    this.u.setColor(Color.white);
                    this.a(this.X, this.p, -120 + n4, true, n5 + 138, this.u);
                }
            }
            catch (Exception exception) {
                // empty catch block
            }
            ++r;
            return;
        }
        catch (RuntimeException runtimeException) {
            String string2;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[22]);
            if (string != null) {
                string2 = Sb[4];
                throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(n2).append(',').append(n3).append(')').toString());
            }
            string2 = Sb[3];
            throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(n2).append(',').append(n3).append(')').toString());
        }
    }

    @Override
    public final synchronized void mouseReleased(MouseEvent mouseEvent) {
        try {
            ++g;
            this.a(mouseEvent, (byte)-128);
            this.I = mouseEvent.getX() + -this.Eb;
            this.xb = mouseEvent.getY() + -this.K;
            this.Bb = 0;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[19]);
            String string = mouseEvent != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    final void a(int n2, byte by, String string) {
        try {
            ++Lb;
            try {
                int n3;
                int n4;
                block15: {
                    block14: {
                        n4 = (-281 + this.m) / 2;
                        n4 += 2;
                        n3 = (this.a - 148) / 2;
                        this.B = string;
                        this.V = n2;
                        n3 += 90;
                        int n5 = 277 * n2 / 100;
                        this.u.setColor(new Color(132, 132, 132));
                        if (this.hb) {
                            this.u.setColor(new Color(220, 0, 0));
                        }
                        this.u.fillRect(n4, n3, n5, 20);
                        this.u.setColor(Color.black);
                        this.u.fillRect(n5 + n4, n3, -n5 + 277, 20);
                        this.u.setColor(new Color(198, 198, 198));
                        if (this.hb) break block14;
                        break block15;
                    }
                    this.u.setColor(new Color(255, 255, 255));
                }
                this.a(this.tb, string, 10 + n3, true, 138 + n4, this.u);
            }
            catch (Exception exception) {
                // empty catch block
            }
            if (by > -96) {
                this.x = null;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[57]).append(n2).append(',').append(by).append(',');
            String string2 = string != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(')').toString());
        }
    }

    private final Image a(byte[] byArray, byte by) {
        try {
            if (by != -54) {
                this.Ob = null;
            }
            ++cb;
            return pa.a(79, (Component)((Object)this), byArray);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[68]);
            String string = byArray != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(by).append(')').toString());
        }
    }

    public static final void provideLoaderApplet(Applet applet) {
        try {
            da.gb = applet;
            ++t;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[54]);
            String string = applet != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    final void a(int n2) {
        try {
            if (n2 >= -54) {
                return;
            }
            ++Db;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[75] + n2 + ')');
        }
    }

    public final URL getDocumentBase() {
        try {
            ++jb;
            if (kb.a != null) {
                return null;
            }
            if (null != da.gb) {
                return da.gb.getDocumentBase();
            }
            return super.getDocumentBase();
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[51]);
        }
    }

    final void a(boolean bl, String string, int n2, String string2, int n3, byte by, int n4, int n5, int n6) {
        try {
            ++mb;
            try {
                System.out.println(Sb[16]);
                this.m = n5;
                this.a = n6;
                kb.a = new qb(this, 800, 600, string, bl, false);
                try {
                    kb.a.getClass().getMethod(Sb[17], Boolean.TYPE).invoke((Object)kb.a, Boolean.FALSE);
                }
                catch (Exception exception) {
                    // empty catch block
                }
                db.d = n4;
                this.n = 1;
                pa.b = pa.k = new c(n2, string2, 0, true);
                try {
                    if (by <= 20) {
                        return;
                    }
                    cb.a(new URL(Sb[15], Sb[14], n3, ""), this, -91);
                }
                catch (IOException iOException) {
                    mb.a(0x1FFFFF, iOException, null);
                }
                this.z = new Thread(this);
                this.z.start();
                this.z.setPriority(1);
            }
            catch (Exception exception) {
                mb.a(0x1FFFFF, exception, null);
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[18]).append(bl).append(',');
            String string3 = string != null ? Sb[4] : Sb[3];
            StringBuilder stringBuilder2 = stringBuilder.append(string3).append(',').append(n2).append(',');
            String string4 = string2 != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder2.append(string4).append(',').append(n3).append(',').append(by).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(')').toString());
        }
    }

    private final boolean b(byte by) {
        try {
            block16: {
                block15: {
                    byte[] byArray;
                    block14: {
                        block13: {
                            ++O;
                            byArray = this.a(Sb[32], 0, 3, 85);
                            if (byArray == null) break block13;
                            break block14;
                        }
                        return false;
                    }
                    if (by != 118) {
                        this.B = null;
                    }
                    byte[] byArray2 = na.a(Sb[30], 0, byArray, -120);
                    this.C = this.a(byArray2, (byte)-54);
                    if (!qa.a(this, Sb[29], 0, by + -118)) break block15;
                    break block16;
                }
                return false;
            }
            if (!qa.a(this, Sb[39], 1, 0)) {
                return false;
            }
            if (!qa.a(this, Sb[31], 2, 0)) {
                return false;
            }
            if (!qa.a(this, Sb[36], 3, 0)) {
                return false;
            }
            if (!qa.a(this, Sb[38], 4, by + -118)) {
                return false;
            }
            if (!qa.a(this, Sb[34], 5, 0)) {
                return false;
            }
            if (!qa.a(this, Sb[35], 6, by ^ 0x76)) {
                return false;
            }
            return qa.a(this, Sb[37], 7, by + -118);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[33] + by + ')');
        }
    }

    void a(boolean bl) {
        try {
            ++k;
            if (bl) {
                this.vb = 85;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[72] + bl + ')');
        }
    }

    public final AppletContext getAppletContext() {
        try {
            ++f;
            if (null != kb.a) {
                return null;
            }
            if (da.gb != null) {
                return da.gb.getAppletContext();
            }
            return super.getAppletContext();
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[41]);
        }
    }

    void a(int n2, int n3, int n4, int n5) {
        try {
            ++s;
            if (n3 < 87) {
                this.z = null;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[50] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    private final void a(String string, boolean bl) {
        try {
            block11: {
                ++P;
                if (this.Kb) {
                    return;
                }
                this.Kb = bl;
                System.out.println(Sb[63] + string);
                try {
                    if (null != da.gb) {
                        a.a(Sb[59], (byte)82, da.gb);
                        if (!client.vh) break block11;
                    }
                    a.a(Sb[59], (byte)-73, this);
                }
                catch (Throwable throwable) {
                    // empty catch block
                }
            }
            try {
                this.getAppletContext().showDocument(new URL(this.getCodeBase(), Sb[63] + string + Sb[62]), Sb[60]);
            }
            catch (Exception exception) {}
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[61]);
            String string2 = string != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(bl).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    @Override
    public final synchronized void mouseDragged(MouseEvent mouseEvent) {
        try {
            this.a(mouseEvent, (byte)-128);
            ++q;
            this.I = mouseEvent.getX() - this.Eb;
            this.xb = mouseEvent.getY() - this.K;
            if (mouseEvent.isMetaDown()) {
                this.Bb = 2;
                if (!client.vh) return;
            }
            this.Bb = 1;
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[76]);
            if (mouseEvent != null) {
                string = Sb[4];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = Sb[3];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    final void a(int n2, int n3, int n4, int n5, int n6) {
        try {
            ++G;
            try {
                System.out.println(Sb[69]);
                this.n = 1;
                this.m = n6;
                this.a = n2;
                nb.s = this.getCodeBase();
                db.d = n3;
                if (pa.k == null) {
                    c c2;
                    c c3 = c2;
                    c c4 = c2;
                    int n7 = n4;
                    String string = null;
                    int n8 = 0;
                    boolean bl = null != da.gb;
                    c3(n7, string, n8, bl);
                    pa.k = c4;
                    pa.b = c4;
                }
                if (n5 != 2) {
                    return;
                }
                if (null != da.gb) {
                    Method method;
                    Method method2 = c.y;
                    if (null != method2) {
                        try {
                            method2.invoke((Object)da.gb, Boolean.TRUE);
                        }
                        catch (Throwable throwable) {
                            // empty catch block
                        }
                    }
                    if (null != (method = c.u)) {
                        try {
                            method.invoke((Object)da.gb, Boolean.FALSE);
                        }
                        catch (Throwable throwable) {
                            // empty catch block
                        }
                    }
                }
                try {
                    cb.a(this.getCodeBase(), this, n5 + -110);
                }
                catch (IOException iOException) {
                    iOException.printStackTrace();
                }
                this.a(n5 + -1, this);
            }
            catch (Exception exception) {
                mb.a(n5 ^ 0x1FFFFD, exception, null);
                this.a(Sb[12], true);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[70] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    @Override
    public final void mouseEntered(MouseEvent mouseEvent) {
        try {
            this.a(mouseEvent, (byte)-128);
            ++Rb;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[20]);
            String string = mouseEvent != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    final boolean d(int n2) {
        try {
            Graphics graphics;
            block6: {
                block5: {
                    ++ab;
                    graphics = this.getGraphics();
                    if (graphics == null) break block5;
                    break block6;
                }
                return false;
            }
            if (n2 != 2) {
                this.sb = -7;
            }
            this.u = graphics.create();
            this.u.translate(this.Eb, this.K);
            this.u.setColor(Color.black);
            this.u.fillRect(0, 0, this.m, this.a);
            this.a(Sb[56], 0, n2 ^ 0x67);
            return true;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[55] + n2 + ')');
        }
    }

    private final void a(InputEvent inputEvent, byte by) {
        try {
            if (by > -127) {
                return;
            }
            ++d;
            int n2 = inputEvent.getModifiers();
            e e2 = this;
            boolean bl = 0 != (n2 & 2);
            e2.bb = bl;
            e e3 = this;
            boolean bl2 = (n2 & 1) != 0;
            e3.gb = bl2;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[74]);
            String string = inputEvent != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(by).append(')').toString());
        }
    }

    public final Dimension getSize() {
        try {
            ++db;
            if (kb.a != null) {
                return kb.a.getSize();
            }
            if (null != da.gb) {
                return da.gb.getSize();
            }
            return super.getSize();
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[67]);
        }
    }

    public final Graphics getGraphics() {
        try {
            ++y;
            if (kb.a != null) {
                return kb.a.getGraphics();
            }
            if (da.gb != null) {
                return da.gb.getGraphics();
            }
            return super.getGraphics();
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[8]);
        }
    }

    public final URL getCodeBase() {
        ++Y;
        if (null != kb.a) {
            return null;
        }
        if (da.gb != null) {
            return da.gb.getCodeBase();
        }
        return super.getCodeBase();
    }

    @Override
    public final void mouseClicked(MouseEvent mouseEvent) {
        try {
            ++h;
            this.a(mouseEvent, (byte)-128);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[7]);
            String string = mouseEvent != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    public final Image createImage(int n2, int n3) {
        try {
            ++Gb;
            if (null != kb.a) {
                return kb.a.createImage(n2, n3);
            }
            if (null != da.gb) {
                return da.gb.createImage(n2, n3);
            }
            return super.createImage(n2, n3);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[71] + n2 + ',' + n3 + ')');
        }
    }

    final void a(Font font, String string, int n2, boolean bl, int n3, Graphics graphics) {
        try {
            Object object;
            block16: {
                block15: {
                    ++D;
                    if (null != kb.a) break block15;
                    object = this;
                    if (!client.vh) break block16;
                }
                object = kb.a;
            }
            FontMetrics fontMetrics = ((Component)object).getFontMetrics(font);
            fontMetrics.stringWidth(string);
            if (!bl) {
                this.c(68);
            }
            graphics.setFont(font);
            graphics.drawString(string, n3 + -(fontMetrics.stringWidth(string) / 2), n2 - -(fontMetrics.getHeight() / 4));
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[2]);
            String string2 = font != null ? Sb[4] : Sb[3];
            StringBuilder stringBuilder2 = stringBuilder.append(string2).append(',');
            String string3 = string != null ? Sb[4] : Sb[3];
            StringBuilder stringBuilder3 = stringBuilder2.append(string3).append(',').append(n2).append(',').append(bl).append(',').append(n3).append(',');
            String string4 = graphics != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder3.append(string4).append(')').toString());
        }
    }

    public final void destroy() {
        block5: {
            try {
                this.vb = -1;
                ++Hb;
                mb.a(11200, 5000L);
                if (-1 != this.vb) break block5;
                System.out.println(Sb[0]);
                this.b(100);
                if (this.z != null) {
                    this.z.stop();
                    this.z = null;
                }
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, Sb[1]);
            }
        }
    }

    void a(byte by, int n2) {
        try {
            if (by <= 105) {
                this.a((InputEvent)null, (byte)83);
            }
            ++c;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[46] + by + ',' + n2 + ')');
        }
    }

    @Override
    public final synchronized void mouseMoved(MouseEvent mouseEvent) {
        try {
            this.a(mouseEvent, (byte)-128);
            ++yb;
            this.I = mouseEvent.getX() - this.Eb;
            this.xb = mouseEvent.getY() + -this.K;
            this.sb = 0;
            this.Bb = 0;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[45]);
            String string = mouseEvent != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    final void c(int n2) {
        block7: {
            boolean bl = client.vh;
            try {
                if (n2 != -28492) {
                    return;
                }
                for (int i2 = 0; i2 < 10; ++i2) {
                    this.F[i2] = 0L;
                    if (!bl) {
                        if (!bl) continue;
                        break;
                    }
                    break block7;
                }
                ++R;
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, Sb[53] + n2 + ')');
            }
        }
    }

    final byte[] a(String string, int n2, int n3, int n4) {
        try {
            ++J;
            if (n4 <= 53) {
                this.c(15);
            }
            try {
                return ib.a(-101, string, n2, n3);
            }
            catch (IOException iOException) {
                mb.a(0x1FFFFF, iOException, Sb[65] + n3);
                return null;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[66]);
            String string2 = string != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(')').toString());
        }
    }

    @Override
    public final void keyTyped(KeyEvent keyEvent) {
        try {
            ++Nb;
            this.a(keyEvent, (byte)-128);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[6]);
            String string = keyEvent != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    public final void start() {
        try {
            ++Fb;
            if (this.vb >= 0) {
                this.vb = 0;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[25]);
        }
    }

    @Override
    public final void mouseExited(MouseEvent mouseEvent) {
        try {
            ++eb;
            this.a(mouseEvent, (byte)-128);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[64]);
            String string = mouseEvent != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    @Override
    public final synchronized void mousePressed(MouseEvent mouseEvent) {
        try {
            block9: {
                block8: {
                    this.a(mouseEvent, (byte)-128);
                    ++Pb;
                    this.I = mouseEvent.getX() - this.Eb;
                    this.xb = mouseEvent.getY() + -this.K;
                    if (mouseEvent.isMetaDown()) break block8;
                    this.Bb = 1;
                    if (!client.vh) break block9;
                }
                this.Bb = 2;
            }
            this.Qb = this.Bb;
            this.sb = 0;
            this.a(this.I, 94, this.Bb, this.xb);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[9]);
            if (mouseEvent != null) {
                string = Sb[4];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = Sb[3];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    void a(byte by) {
        try {
            ++o;
            if (by != -92) {
                e.provideLoaderApplet(null);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[73] + by + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    @Override
    public final synchronized void keyPressed(KeyEvent keyEvent) {
        boolean bl = client.vh;
        try {
            int n2;
            int n3;
            int n22;
            block39: {
                ++L;
                this.a(keyEvent, (byte)-128);
                n22 = keyEvent.getKeyChar();
                int n5 = keyEvent.getKeyCode();
                this.a((byte)126, n22);
                if (-113 == ~n5) {
                    boolean bl2 = this.U = !this.U;
                }
                if ('\uffffffb1' == ~((char)n22) || 'M' != (char)n22) {
                    // empty if block
                }
                if (~((char)n22) == '\uffffffdf') {
                    // empty if block
                }
                if ('\uffffff84' != ~((char)n22)) {
                    // empty if block
                }
                this.sb = 0;
                if (~((char)n22) == '\uffffff91' || ~((char)n22) != '\uffffff92') {
                    // empty if block
                }
                if (-41 != ~n5) {
                    // empty if block
                }
                if (~n5 == -40) {
                    this.E = true;
                }
                if ((char)n22 == '}') {
                    // empty if block
                }
                if (~n5 == -39) {
                    // empty if block
                }
                if (~n5 == -38) {
                    this.Z = true;
                }
                boolean bl2 = false;
                for (int i2 = 0; i2 < i.f.length(); ++i2) {
                    n3 = ~n22;
                    n2 = ~i.f.charAt(i2);
                    if (!bl) {
                        if (n3 != n2) continue;
                        bl2 = true;
                        if (!bl) break;
                        if (!bl) continue;
                    }
                    break block39;
                }
                if (bl2 && ~this.e.length() > -21) {
                    this.e = this.e + (char)n22;
                }
                if (bl2 && 80 > this.x.length()) {
                    this.x = this.x + (char)n22;
                }
                n3 = 8;
                n2 = n22;
            }
            if (n3 == n2 && -1 > ~this.e.length()) {
                this.e = this.e.substring(0, this.e.length() - 1);
            }
            if (8 == n22 && ~this.x.length() < -1) {
                this.x = this.x.substring(0, -1 + this.x.length());
            }
            if (-11 != ~n22 && 13 != n22) {
                return;
            }
            this.Cb = this.e;
            this.Ob = this.x;
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[26]);
            if (keyEvent != null) {
                string = Sb[4];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = Sb[3];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    final void a(int n2, byte by) {
        try {
            ++rb;
            this.Ib = 1000 / n2;
            if (by <= 104) {
                this.Eb = 113;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[13] + n2 + ',' + by + ')');
        }
    }

    public final void stop() {
        try {
            ++pb;
            if (~this.vb <= -1) {
                this.vb = 4000 / this.Ib;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[52]);
        }
    }

    public final boolean isDisplayable() {
        try {
            ++W;
            if (null != kb.a) {
                boolean bl = kb.a.getPeer() != null;
                return bl;
            }
            if (null != da.gb) {
                boolean bl = null != da.gb.getPeer();
                return bl;
            }
            return super.getPeer() != null;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Sb[47]);
        }
    }

    public final void update(Graphics graphics) {
        try {
            ++ub;
            this.paint(graphics);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[40]);
            String string = graphics != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    void a(int n2, Runnable runnable) {
        try {
            ++A;
            Thread thread = new Thread(runnable);
            if (n2 != 1) {
                return;
            }
            thread.setDaemon(true);
            thread.start();
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Sb[5]).append(n2).append(',');
            String string = runnable != null ? Sb[4] : Sb[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    protected e() {
        this.tb = new Font(Sb[43], 0, 15);
        this.X = new Font(Sb[42], 1, 13);
        this.Jb = new Font(Sb[42], 0, 12);
    }

    static {
        Sb = new String[]{e.z(e.z("-#;m\u000bwm,{H}{8a\u001a}gd(\u000ewq+a\u0006\u007f##a\u0004t")), e.z(e.z("}-,m\u001blq'q@1")), e.z(e.z("}-\u0004M@")), e.z(e.z("vv$d")), e.z(e.z("c-f&\u0015")), e.z(e.z("}-\u001b ")), e.z(e.z("}-#m\u0011Lz8m\f0")), e.z(e.z("}-%g\u001dkf\u000bd\u0001{h-l@")), e.z(e.z("}-/m\u001c_q)x\u0000q`; A")), e.z(e.z("}-%g\u001dkf\u0018z\rkp-l@")), e.z(e.z("}-8i\u0001vw`")), e.z(e.z("}-:}\u00060*")), e.z(e.z("{q){\u0000")), e.z(e.z("}-\u0001M@")), e.z(e.z(")1\u007f&X63f9")), e.z(e.z("pw<x")), e.z(e.z("Kw)z\u001c}ghi\u0018ho!k\tlj'f")), e.z(e.z("kf<N\u0007{v;\\\u001ayu-z\u001byo\u0003m\u0011kF&i\ntf,")), e.z(e.z("}-\u0006M@")), e.z(e.z("}-%g\u001dkf\u001am\u0004}b;m\f0")), e.z(e.z("}-%g\u001dkf\rf\u001c}q-l@")), e.z(e.z("}-/m\u001cHb:i\u0005}w-z@")), e.z(e.z("}-\u000eM@")), e.z(e.z("\u00b1#z8X).z8Y-#\u0002i\u000f}{hD\u001c|")), e.z(e.z("[q-i\u001c}ghj\u00118I\tO\r@#e(\u001eqp!|Hot?&\u0002yd-pF{l%")), e.z(e.z("}-;|\tjw`!")), e.z(e.z("}-#m\u0011Hq-{\u001b}g`")), e.z(e.z("}-\u001aM@")), e.z(e.z("[o'{\u0001vdhx\u001awd:i\u0005")), e.z(e.z("p2yx")), e.z(e.z("tl/gFld)")), e.z(e.z("p2zx")), e.z(e.z("Rb/m\u00108o!j\u001ayq1")), e.z(e.z("}-\u0002M@")), e.z(e.z("p2~j")), e.z(e.z("p1xj")), e.z(e.z("p2{j")), e.z(e.z("p1|j")), e.z(e.z("p2|j")), e.z(e.z("p2zj")), e.z(e.z("}-=x\fyw- ")), e.z(e.z("}-/m\u001cYs8d\rl@'f\u001c}{< A")), e.z(e.z("Pf$~\rlj+i")), e.z(e.z("Lj%m\u001bJl%i\u0006")), e.z(e.z("Tl)l\u0001vd")), e.z(e.z("}-%g\u001dkf\u0005g\u001e}g`")), e.z(e.z("}-\tL@")), e.z(e.z("}-!{,qp8d\tab*d\r0*")), e.z(e.z("}-\u0002L@")), e.z(e.z("}-\u0005I@")), e.z(e.z("}-\u0019 ")), e.z(e.z("}-/m\u001c\\l+}\u0005}m<J\tkf`!")), e.z(e.z("}-;|\u0007h+a")), e.z(e.z("}-\u0018M@")), e.z(e.z("}-8z\u0007nj,m$wb,m\u001aYs8d\rl+")), e.z(e.z("}-\u0005M@")), e.z(e.z("Tl)l\u0001vdf&F")), e.z(e.z("}-\rM@")), e.z(e.z("}-#m\u0011Jf$m\tkf, ")), e.z(e.z("tl/o\r|l=|")), e.z(e.z("Gw'x")), e.z(e.z("}-\u0003M@")), e.z(e.z("6t;")), e.z(e.z("}q:g\u001aGd)e\rG")), e.z(e.z("}-%g\u001dkf\rp\u0001lf, ")), e.z(e.z("Mm)j\u0004}#<gHtl)lH{l&|\rvwhx\t{hh")), e.z(e.z("}-\u0000M@")), e.z(e.z("}-/m\u001cKj2m@1")), e.z(e.z("}-\u000fM@")), e.z(e.z("Kw)z\u001c}ghi\u0018ho-|")), e.z(e.z("}-\u0007M@")), e.z(e.z("}-+z\ryw-A\u0005yd- ")), e.z(e.z("}-\u001bI@")), e.z(e.z("}-\u0003K@")), e.z(e.z("}-\u001bM@")), e.z(e.z("}-\u0019M@")), e.z(e.z("}-%g\u001dkf\fz\t\u007fd-l@"))};
        kb = new byte[250][];
        nb = new int[512];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x68);
        }
        return cArray;
    }

    /*
     * Handled impossible loop by duplicating code
     * Enabled aggressive block sorting
     */
    private static String z(char[] cArray) {
        char[] cArray2;
        block9: {
            int n2;
            int n3;
            block8: {
                cArray2 = cArray;
                n3 = cArray.length;
                n2 = 0;
                if (!true) break block8;
                n3 = n3;
                if (n3 <= n2) break block9;
            }
            do {
                int n4;
                cArray2 = cArray2;
                int n5 = n2;
                char c2 = cArray2[n5];
                switch (n2 % 5) {
                    case 0: {
                        n4 = 24;
                        break;
                    }
                    case 1: {
                        n4 = 3;
                        break;
                    }
                    case 2: {
                        n4 = 72;
                        break;
                    }
                    case 3: {
                        n4 = 8;
                        break;
                    }
                    default: {
                        n4 = 104;
                    }
                }
                cArray2[n5] = (char)(c2 ^ n4);
                ++n2;
                n3 = n3;
            } while (n3 > n2);
        }
        return new String(cArray2).intern();
    }
}

