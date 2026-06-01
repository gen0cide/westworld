/*
 * Decompiled with CFR 0.152.
 */
import java.awt.Component;
import java.awt.EventQueue;
import java.awt.Frame;
import java.awt.Point;
import java.awt.Toolkit;
import java.awt.datatransfer.Clipboard;
import java.awt.datatransfer.Transferable;
import java.io.DataInputStream;
import java.io.File;
import java.io.IOException;
import java.lang.reflect.Method;
import java.net.InetAddress;
import java.net.Socket;
import java.net.URL;

final class c
implements Runnable {
    private boolean e;
    private q t;
    private Thread r;
    private Object x;
    private static String m;
    private static String p;
    private boolean j;
    private Object i;
    EventQueue n;
    static Method u;
    private static int b;
    private g a;
    private boolean c;
    private wa w;
    private static String g;
    static Method y;
    d f;
    d s;
    private static String h;
    d v;
    static String q;
    private d[] l;
    private static volatile long d;
    private g o;
    static String k;
    private static final String[] z;

    final g a(String string, int n2, int n3) {
        if (n3 > -66) {
            return null;
        }
        return this.a(n2, (byte)81, string, false);
    }

    final g a(boolean bl, Runnable runnable, int n2) {
        if (!bl) {
            this.a(-34, 71, (byte)60, 103, null);
        }
        return this.a(2, 0, (byte)-21, n2, runnable);
    }

    private static final d a(int n2, String string, boolean bl, String string2) {
        String string3;
        block13: {
            block11: {
                block12: {
                    if (-34 == ~n2) break block11;
                    if (~n2 != -35) break block12;
                    string3 = z[0] + string + z[6] + string2 + z[7];
                    break block13;
                }
                string3 = z[0] + string + z[6] + string2 + z[4];
                break block13;
            }
            string3 = z[0] + string + z[6] + string2 + z[11];
        }
        if (bl) {
            return null;
        }
        String[] stringArray = new String[]{z[5], z[2], m, z[1], z[9], z[3], z[10], ""};
        for (int i2 = 0; i2 < stringArray.length; ++i2) {
            String string4 = stringArray[i2];
            if (0 < string4.length() && !new File(string4).exists()) {
                continue;
            }
            try {
                d d2 = new d(new File(string4, string3), z[8], 10000L);
                return d2;
            }
            catch (Exception exception) {
                // empty catch block
            }
        }
        return null;
    }

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    @Override
    public final void run() {
        while (true) {
            g g2;
            Object object = this;
            synchronized (object) {
                while (true) {
                    if (this.e) {
                        return;
                    }
                    if (null != this.a) {
                        g2 = this.a;
                        this.a = this.a.a;
                        if (null != this.a) break;
                        this.o = null;
                        break;
                    }
                    try {
                        this.wait();
                    }
                    catch (InterruptedException interruptedException) {}
                }
            }
            try {
                block82: {
                    block84: {
                        block86: {
                            block89: {
                                block90: {
                                    block92: {
                                        block93: {
                                            block94: {
                                                block96: {
                                                    block98: {
                                                        int n2;
                                                        block97: {
                                                            block95: {
                                                                block91: {
                                                                    block88: {
                                                                        block87: {
                                                                            block85: {
                                                                                block83: {
                                                                                    n2 = g2.g;
                                                                                    if (~n2 != -2) break block83;
                                                                                    if (p.a(0) < d) {
                                                                                        throw new IOException();
                                                                                    }
                                                                                    g2.d = new Socket(InetAddress.getByName((String)g2.f), g2.e);
                                                                                    break block82;
                                                                                }
                                                                                if (n2 == 22) break block84;
                                                                                if (2 != n2) break block85;
                                                                                Thread thread = new Thread((Runnable)g2.f);
                                                                                thread.setDaemon(true);
                                                                                thread.start();
                                                                                thread.setPriority(g2.e);
                                                                                g2.d = thread;
                                                                                break block82;
                                                                            }
                                                                            if (n2 == 4) break block86;
                                                                            if (8 != n2) break block87;
                                                                            Object[] objectArray = (Object[])g2.f;
                                                                            if (this.j && null == ((Class)objectArray[0]).getClassLoader()) {
                                                                                throw new SecurityException();
                                                                            }
                                                                            g2.d = ((Class)objectArray[0]).getDeclaredMethod((String)objectArray[1], (Class[])objectArray[2]);
                                                                            break block82;
                                                                        }
                                                                        if (-10 != ~n2) break block88;
                                                                        Object[] objectArray = (Object[])g2.f;
                                                                        if (this.j && ((Class)objectArray[0]).getClassLoader() == null) {
                                                                            throw new SecurityException();
                                                                        }
                                                                        g2.d = ((Class)objectArray[0]).getDeclaredField((String)objectArray[1]);
                                                                        break block82;
                                                                    }
                                                                    if (n2 == 18) break block89;
                                                                    if (19 == n2) break block90;
                                                                    if (!this.j) {
                                                                        throw new Exception("");
                                                                    }
                                                                    if (3 != n2) break block91;
                                                                    if (d > p.a(0)) {
                                                                        throw new IOException();
                                                                    }
                                                                    String string = (0xFF & g2.e >> -182496008) + "." + (0xFF & g2.e >> 954736400) + "." + ((0xFFC0 & g2.e) >> -58046680) + "." + (0xFF & g2.e);
                                                                    g2.d = InetAddress.getByName(string).getHostName();
                                                                    break block82;
                                                                }
                                                                if (n2 == 21) break block92;
                                                                if (5 == n2) break block93;
                                                                if (~n2 == -7) break block94;
                                                                if (n2 != 7) break block95;
                                                                if (!this.c) {
                                                                    Class.forName(z[15]).getMethod(z[13], new Class[0]).invoke(this.i, new Object[0]);
                                                                    break block82;
                                                                } else {
                                                                    this.w.a((Frame)g2.f, 0);
                                                                }
                                                                break block82;
                                                            }
                                                            if (-13 == ~n2) break block96;
                                                            if (-14 != ~n2) break block97;
                                                            d d2 = c.a(b, "", false, (String)g2.f);
                                                            g2.d = d2;
                                                            break block82;
                                                        }
                                                        if (this.j && n2 == 14) break block98;
                                                        if (!this.j || ~n2 != -16) {
                                                            if (!this.c && n2 == 17) {
                                                                Object[] objectArray = (Object[])g2.f;
                                                                Class.forName("j").getDeclaredMethod(z[20], Component.class, int[].class, Integer.TYPE, Integer.TYPE, Point.class).invoke(this.x, objectArray[0], objectArray[1], new Integer(g2.e), new Integer(g2.c), objectArray[2]);
                                                                break block82;
                                                            } else {
                                                                if (16 != n2) {
                                                                    throw new Exception("");
                                                                }
                                                                try {
                                                                    if (!g.startsWith(z[23])) {
                                                                        throw new Exception();
                                                                    }
                                                                    String string = (String)g2.f;
                                                                    if (!string.startsWith(z[18]) && !string.startsWith(z[16])) {
                                                                        throw new Exception();
                                                                    }
                                                                    String string2 = z[14];
                                                                    for (int i2 = 0; i2 < string.length(); ++i2) {
                                                                        if (-1 != string2.indexOf(string.charAt(i2))) continue;
                                                                        throw new Exception();
                                                                    }
                                                                    Runtime.getRuntime().exec(z[21] + string + "\"");
                                                                    g2.d = null;
                                                                }
                                                                catch (Exception exception) {
                                                                    g2.d = exception;
                                                                    throw exception;
                                                                }
                                                            }
                                                        } else {
                                                            boolean bl = ~g2.e != -1;
                                                            Component component = (Component)g2.f;
                                                            if (this.c) {
                                                                this.t.a(-4, component, bl);
                                                                break block82;
                                                            } else {
                                                                Class.forName("j").getDeclaredMethod(z[24], Component.class, Boolean.TYPE).invoke(this.x, component, new Boolean(bl));
                                                            }
                                                        }
                                                        break block82;
                                                    }
                                                    int n3 = g2.e;
                                                    int n4 = g2.c;
                                                    if (this.c) {
                                                        this.t.a(23529, n4, n3);
                                                        break block82;
                                                    } else {
                                                        Class.forName("j").getDeclaredMethod(z[19], Integer.TYPE, Integer.TYPE).invoke(this.x, new Integer(n3), new Integer(n4));
                                                    }
                                                    break block82;
                                                }
                                                d d3 = c.a(b, p, false, (String)g2.f);
                                                g2.d = d3;
                                                break block82;
                                            }
                                            Frame frame = new Frame(z[12]);
                                            g2.d = frame;
                                            frame.setResizable(false);
                                            if (this.c) {
                                                this.w.a(frame, g2.c >> -747878896, g2.c & 0xFFFF, g2.e & 0xFFFF, g2.e >>> 831913136, (byte)77);
                                                break block82;
                                            } else {
                                                Class.forName(z[15]).getMethod(z[17], Frame.class, Integer.TYPE, Integer.TYPE, Integer.TYPE, Integer.TYPE).invoke(this.i, frame, new Integer(g2.e >>> -1397573296), new Integer(0xFFFF & g2.e), new Integer(g2.c >> -1159913680), new Integer(g2.c & 0xFFFF));
                                            }
                                            break block82;
                                        }
                                        g2.d = this.c ? (Object)this.w.a((byte)-100) : Class.forName(z[15]).getMethod(z[22], new Class[0]).invoke(this.i, new Object[0]);
                                        break block82;
                                    }
                                    if (p.a(0) < d) {
                                        throw new IOException();
                                    }
                                    g2.d = InetAddress.getByName((String)g2.f).getAddress();
                                    break block82;
                                }
                                Transferable transferable = (Transferable)g2.f;
                                Clipboard clipboard = Toolkit.getDefaultToolkit().getSystemClipboard();
                                clipboard.setContents(transferable, null);
                                break block82;
                            }
                            Clipboard clipboard = Toolkit.getDefaultToolkit().getSystemClipboard();
                            g2.d = clipboard.getContents(null);
                            break block82;
                        }
                        if ((p.a(0) ^ 0xFFFFFFFFFFFFFFFFL) > (d ^ 0xFFFFFFFFFFFFFFFFL)) {
                            throw new IOException();
                        }
                        g2.d = new DataInputStream(((URL)g2.f).openStream());
                        break block82;
                    }
                    if ((d ^ 0xFFFFFFFFFFFFFFFFL) < (p.a(0) ^ 0xFFFFFFFFFFFFFFFFL)) {
                        throw new IOException();
                    }
                    try {
                        g2.d = na.a(4718, g2.e, (String)g2.f).a((byte)50);
                    }
                    catch (fa fa2) {
                        g2.d = fa2.getMessage();
                        throw fa2;
                    }
                }
                g2.b = 1;
            }
            catch (ThreadDeath threadDeath) {
                throw threadDeath;
            }
            catch (Throwable throwable) {
                g2.b = 2;
            }
            object = g2;
            synchronized (object) {
                g2.notify();
            }
        }
    }

    private final g a(int n2, byte by, String string, boolean bl) {
        if (by != 81) {
            this.a(3, (byte)-100, null, true);
        }
        c c2 = this;
        int n3 = bl ? 22 : 1;
        return c2.a(n3, 0, (byte)-21, n2, string);
    }

    final g a(byte by, URL uRL) {
        if (by != 74) {
            d = -110L;
        }
        return this.a(4, 0, (byte)-21, 0, uRL);
    }

    final g a(int n2, byte by) {
        int n3 = 9 / ((-58 - by) / 56);
        return this.a(3, 0, (byte)-21, n2, null);
    }

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     */
    private final g a(int n2, int n3, byte by, int n4, Object object) {
        g g2 = new g();
        g2.e = n4;
        g2.g = n2;
        g2.c = n3;
        g2.f = object;
        c c2 = this;
        synchronized (c2) {
            if (null == this.o) {
                this.o = this.a = g2;
            } else {
                this.o.a = g2;
                this.o = g2;
            }
            this.notify();
        }
        if (by != -21) {
            k = null;
        }
        return g2;
    }

    c(int n2, String string, int n3, boolean bl) throws Exception {
        block49: {
            block45: {
                block44: {
                    block43: {
                        block42: {
                            block41: {
                                block40: {
                                    this.j = false;
                                    this.a = null;
                                    this.f = null;
                                    this.s = null;
                                    this.e = false;
                                    this.c = false;
                                    this.v = null;
                                    this.o = null;
                                    k = z[42];
                                    b = n2;
                                    q = z[39];
                                    p = string;
                                    this.j = bl;
                                    try {
                                        q = System.getProperty(z[32]);
                                        k = System.getProperty(z[36]);
                                    }
                                    catch (Exception exception) {
                                        // empty catch block
                                    }
                                    if (~q.toLowerCase().indexOf(z[43]) != 0) break block40;
                                    break block41;
                                }
                                this.c = true;
                            }
                            try {
                                h = System.getProperty(z[29]);
                            }
                            catch (Exception exception) {
                                h = z[39];
                            }
                            g = h.toLowerCase();
                            try {
                                System.getProperty(z[34]).toLowerCase();
                            }
                            catch (Exception exception) {
                                // empty catch block
                            }
                            try {
                                System.getProperty(z[30]).toLowerCase();
                            }
                            catch (Exception exception) {
                                // empty catch block
                            }
                            try {
                                m = System.getProperty(z[31]);
                                if (m != null) {
                                    m = m + "/";
                                }
                            }
                            catch (Exception exception) {
                                // empty catch block
                            }
                            if (null == m) break block42;
                            break block43;
                        }
                        m = z[37];
                    }
                    try {
                        this.n = Toolkit.getDefaultToolkit().getSystemEventQueue();
                    }
                    catch (Throwable throwable) {
                        // empty catch block
                    }
                    if (!this.c) break block44;
                    break block45;
                }
                try {
                    u = Class.forName(z[40]).getDeclaredMethod(z[25], Boolean.TYPE);
                }
                catch (Exception exception) {
                    // empty catch block
                }
                try {
                    y = Class.forName(z[28]).getDeclaredMethod(z[38], Boolean.TYPE);
                }
                catch (Exception exception) {
                    // empty catch block
                }
            }
            r.a(b, (byte)101, p);
            if (this.j) {
                block47: {
                    this.s = new d(r.a(b, null, z[35], 0), z[8], 25L);
                    this.f = new d(r.a(2, z[33]), z[8], 314572800L);
                    this.v = new d(r.a(2, z[26]), z[8], 0x100000L);
                    this.l = new d[n3];
                    for (int i2 = 0; n3 > i2; ++i2) {
                        this.l[i2] = new d(r.a(2, z[41] + i2), z[8], 0x100000L);
                    }
                    if (this.c) {
                        try {
                            Class.forName(z[27]).newInstance();
                        }
                        catch (Throwable throwable) {
                            // empty catch block
                        }
                    }
                    try {
                        block46: {
                            if (!this.c) break block46;
                            this.w = new wa();
                            break block47;
                        }
                        this.i = Class.forName(z[15]).newInstance();
                    }
                    catch (Throwable throwable) {
                        // empty catch block
                    }
                }
                try {
                    block48: {
                        if (this.c) break block48;
                        this.x = Class.forName("j").newInstance();
                        break block49;
                    }
                    this.t = new q();
                }
                catch (Throwable throwable) {
                    // empty catch block
                }
            }
        }
        this.e = false;
        this.r = new Thread(this);
        this.r.setPriority(10);
        this.r.setDaemon(true);
        this.r.start();
    }

    static {
        z = new String[]{c.z(c.z("lHu\u0004:Y")), c.z(c.z("e\u0013=\u0016+hM}\u00161)")), c.z(c.z(")[a\u0002#eAwN")), c.z(c.z("e\u0013=")), c.z(c.z("(Ms\u0015")), c.z(c.z("e\u0013=\u00131eHq\t')")), c.z(c.z("YY`\u0004$c[w\u000f!cZ")), c.z(c.z("Y^{\u0011lbHf")), c.z(c.z("t^")), c.z(c.z("e\u0013=\u0016+hGfN")), c.z(c.z(")]\u007f\u0011m")), c.z(c.z("Y[qO&g]")), c.z(c.z("LHu\u0004:&og\r.&zq\u0013'cG")), c.z(c.z("cQ{\u0015")), c.z(c.z("gKq\u0005'`Nz\b(mE\u007f\u000f-vX`\u00126s_e\u0019;|hP\"\u0006CoU)\u000bLb^,\fIyC3\u0011R|D6\u001a_s\"Pp5\u001d'Wu>\u0010-G\u007f*\u00077JoY\n(Nh")), c.z(c.z("nH")), c.z(c.z("n]f\u00111<\u0006=")), c.z(c.z("cGf\u00040")), c.z(c.z("n]f\u0011x)\u0006")), c.z(c.z("kFd\u0004/i\\a\u0004")), c.z(c.z("uLf\u00027u]}\f!s[a\u000e0")), c.z(c.z("eDvAme\ta\u0015#t]2C($\t0")), c.z(c.z("j@a\u0015/iMw\u0012")), c.z(c.z("q@|")), c.z(c.z("uA}\u0016!s[a\u000e0")), c.z(c.z("uLf'-e\\a50g_w\u00131gEY\u0004;ul|\u0000 jLv")), c.z(c.z("kH{\u000f\u001d`@~\u0004\u001deHq\t'(@v\u0019p3\u001c")), c.z(c.z("tK")), c.z(c.z("lHd\u0000lg^fO\u0001iGf\u0000+hL`")), c.z(c.z("iZ<\u000f#kL")), c.z(c.z("iZ<\u0017'tZ{\u000e,")), c.z(c.z("sZw\u0013lnF\u007f\u0004")), c.z(c.z("lHd\u0000lpL|\u0005-t")), c.z(c.z("kH{\u000f\u001d`@~\u0004\u001deHq\t'(Ms\u0015p")), c.z(c.z("iZ<\u00000eA")), c.z(c.z("tH|\u0005-k\u0007v\u00006")), c.z(c.z("lHd\u0000lpL`\u0012+iG")), c.z(c.z("x\u0006")), c.z(c.z("uLf'-e\\a\";eEw3-i]")), c.z(c.z("SGy\u000f-qG")), c.z(c.z("lHd\u0000lg^fO\u0001iDb\u000e,cGf")), c.z(c.z("kH{\u000f\u001d`@~\u0004\u001deHq\t'(@v\u0019")), c.z(c.z("7\u0007#")), c.z(c.z("k@q\u0013-uFt\u0015"))};
        d = 0L;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x42);
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
                        n4 = 6;
                        break;
                    }
                    case 1: {
                        n4 = 41;
                        break;
                    }
                    case 2: {
                        n4 = 18;
                        break;
                    }
                    case 3: {
                        n4 = 97;
                        break;
                    }
                    default: {
                        n4 = 66;
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

