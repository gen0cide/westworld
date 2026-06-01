/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  java.applet.Applet
 */
import java.applet.Applet;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.Socket;
import java.net.URL;

final class da
extends b
implements Runnable {
    static int[][] J;
    private byte[] Z = new byte[1];
    static fb db;
    private int W = 0;
    private boolean fb = false;
    static int S;
    private int G = 0;
    static int[] T;
    private OutputStream Q;
    static Applet gb;
    static int R;
    static int P;
    static int eb;
    private InputStream U;
    static int M;
    private Socket ab;
    private boolean X = true;
    static int V;
    static int cb;
    static int L;
    static int I;
    static int H;
    static v O;
    private byte[] Y;
    static int[] N;
    static int bb;
    static int K;
    private static final String[] hb;

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    @Override
    public final void run() {
        boolean bl = client.vh;
        try {
            ++R;
            do {
                int n2;
                int n3;
                if (this.X) return;
                boolean bl2 = true;
                if (!bl) {
                    // empty if block
                }
                if (!bl2) return;
                da da2 = this;
                synchronized (da2) {
                    block26: {
                        block25: {
                            if (this.G == this.W) {
                                try {
                                    this.wait();
                                }
                                catch (InterruptedException interruptedException) {
                                    // empty catch block
                                }
                            }
                            if (this.X) return;
                            if (this.W <= this.G) break block25;
                            n3 = 5000 - this.W;
                            if (!bl) break block26;
                        }
                        n3 = this.G + -this.W;
                    }
                    n2 = this.W;
                }
                if (-1 <= ~n3) {
                    continue;
                }
                try {
                    this.Q.write(this.Y, n2, n3);
                }
                catch (IOException iOException) {
                    this.k = true;
                    this.B = hb[16] + iOException;
                }
                this.W = (this.W - -n3) % 5000;
                try {
                    if (this.W != this.G) continue;
                    this.Q.flush();
                }
                catch (IOException iOException) {
                    this.k = true;
                    this.B = hb[16] + iOException;
                }
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, hb[15]);
        }
    }

    static final void a(String string, int n2, int n3) {
        try {
            if (n3 != 0) {
                da.a((String)null, -126, -28);
            }
            ++I;
            d.h.a(nb.q, (byte)-101, string + o.l + hb[4] + n2 + "%");
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(hb[5]);
            String string2 = string != null ? hb[3] : hb[2];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n2).append(',').append(n3).append(')').toString());
        }
    }

    static final byte[] a(URL uRL, boolean bl, boolean bl2) throws IOException {
        boolean bl3 = client.vh;
        try {
            tb tb2;
            block17: {
                block16: {
                    ++L;
                    jb jb2 = new jb(pa.k, uRL, 2000000);
                    if (bl) {
                        da.a("", 0, 0);
                    }
                    while (!jb2.a(-2)) {
                        mb.a(11200, 50L);
                        if (!bl3) continue;
                    }
                    tb2 = jb2.a((byte)-120);
                    if (tb2 == null) break block16;
                    break block17;
                }
                throw new IOException(hb[7]);
            }
            if (!bl2) {
                da.a((String)null, -15, -97);
            }
            if (bl) {
                da.a("", 100, 0);
            }
            return tb2.d(0);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(hb[6]);
            String string = uRL != null ? hb[3] : hb[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(bl).append(',').append(bl2).append(')').toString());
        }
    }

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     * WARNING - Removed back jump from a try to a catch block - possible behaviour change.
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    @Override
    final void a(byte[] byArray, int n2, int n3, byte by) throws IOException {
        boolean bl = client.vh;
        try {
            if (by != -67) {
                return;
            }
        }
        catch (IOException iOException) {
            throw iOException;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(hb[9]);
            if (byArray != null) {
                string = hb[3];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(by).append(')').toString());
            }
            string = hb[2];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(by).append(')').toString());
        }
        {
            ++S;
            if (true != !this.fb) return;
            if (null == this.Y) {
                this.Y = new byte[5000];
            }
            da da2 = this;
            synchronized (da2) {
                da da3;
                for (int i2 = 0; i2 < n3; ++i2) {
                    this.Y[this.G] = byArray[n2 + i2];
                    this.G = (this.G - -1) % 5000;
                    da3 = this;
                    if (bl) return;
                    if (da3.G != (4900 + this.W) % 5000) continue;
                    throw new IOException(hb[8]);
                }
                this.notify();
                da3 = da2;
                // ** MonitorExit[v3] (shouldn't be in output)
                return;
            }
        }
    }

    @Override
    final int b(byte by) throws IOException {
        ++H;
        if (this.fb) {
            return 0;
        }
        try {
            int n2 = -127 % ((by - -64) / 56);
            return this.U.available();
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, hb[10] + by + ')');
        }
    }

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     */
    @Override
    final void a(boolean bl) {
        try {
            ++V;
            super.a(true);
            this.fb = bl;
            try {
                if (this.U != null) {
                    this.U.close();
                }
                if (null != this.Q) {
                    this.Q.close();
                }
                if (null != this.ab) {
                    this.ab.close();
                }
            }
            catch (IOException iOException) {
                System.out.println(hb[1]);
            }
            this.X = true;
            da da2 = this;
            synchronized (da2) {
                this.notify();
            }
            this.Y = null;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, hb[0] + bl + ')');
        }
    }

    static final int a(int n2, byte by, int n3, int n4) {
        try {
            if (by != -66) {
                K = 35;
            }
            ++cb;
            return -(n4 / 8 * 32) + -1 + -(n3 / 8 * 1024) - n2 / 8;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, hb[17] + n2 + ',' + by + ',' + n3 + ',' + n4 + ')');
        }
    }

    @Override
    final int b(boolean bl) throws IOException {
        try {
            block5: {
                block4: {
                    ++P;
                    if (!this.fb != bl) break block4;
                    break block5;
                }
                return 0;
            }
            this.a(this.Z, 1, 0, 123);
            return 0xFF & this.Z[0];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, hb[14] + bl + ')');
        }
    }

    @Override
    final void a(byte[] byArray, int n2, int n3, int n4) throws IOException {
        boolean bl = client.vh;
        ++eb;
        if (this.fb) {
            return;
        }
        try {
            int n5 = -81 / ((-25 - n4) / 35);
            int n6 = 0;
            for (int i2 = 0; i2 < n2 && !bl; i2 += n6) {
                n6 = this.U.read(byArray, n3 + i2, -i2 + n2);
                if (~n6 < -1) continue;
                throw new IOException(hb[12]);
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(hb[13]);
            String string = byArray != null ? hb[3] : hb[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(')').toString());
        }
    }

    da(Socket socket, e e2) throws IOException {
        try {
            this.ab = socket;
            this.U = socket.getInputStream();
            this.Q = socket.getOutputStream();
            this.X = false;
            e2.a(1, this);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(hb[11]);
            String string = socket != null ? hb[3] : hb[2];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',');
            String string2 = e2 != null ? hb[3] : hb[2];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(')').toString());
        }
    }

    static {
        hb = new String[]{da.z(da.z("0CF?\u001a")), da.z(da.z("\u0011P\u001a\u0017@tA\u0004\u0017A=L\u000fXA P\r\u0019_")), da.z(da.z(":W\u0004\u0014")), da.z(da.z("/\fFVO")), da.z(da.z("t\u000fH")), da.z(da.z("0CF=\u001a")), da.z(da.z("0CF0\u001a")), da.z(da.z("\u0017M\u001d\u0014V:\u0005\u001cXV;U\u0006\u0014]5FH\u001e[8G")), da.z(da.z("6W\u000e\u001eW&\u0002\u0007\u000eW&D\u0004\u0017E")), da.z(da.z("0CF<\u001a")), da.z(da.z("0CF:\u001a")), da.z(da.z("0CFD[:K\u001cF\u001a")), da.z(da.z("\u0011m.")), da.z(da.z("0CF>\u001a")), da.z(da.z("0CF9\u001a")), da.z(da.z("0CF\nG:\nA")), da.z(da.z("\u0000U\u001a\u0011F1PR")), da.z(da.z("0CF;\u001a"))};
        M = 0;
        db = new fb();
        O = new v(da.z(da.z("\u0003v:;")), da.z(da.z(";D\u000e\u0011Q1")), da.z(da.z("\u000bP\u000b")), 1);
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x32);
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
                        n4 = 84;
                        break;
                    }
                    case 1: {
                        n4 = 34;
                        break;
                    }
                    case 2: {
                        n4 = 104;
                        break;
                    }
                    case 3: {
                        n4 = 120;
                        break;
                    }
                    default: {
                        n4 = 50;
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

