/*
 * Decompiled with CFR 0.152.
 */
import java.io.DataInputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.net.Socket;
import java.net.URL;

final class jb
implements Runnable {
    static int m;
    private c b;
    static int i;
    private g f;
    static int d;
    static int j;
    private DataInputStream q;
    private tb h;
    static int e;
    private URL g;
    private g c;
    static int n;
    static int[] k;
    static int p;
    private g l;
    private int a;
    static int o;
    private static final String[] z;

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     */
    @Override
    public final void run() {
        boolean bl = client.vh;
        try {
            ++m;
            try {
                int n2;
                int n3;
                block18: {
                    while (~this.h.w > ~this.h.F.length) {
                        int n4 = this.q.read(this.h.F, this.h.w, this.h.F.length - this.h.w);
                        n3 = ~n4;
                        n2 = -1;
                        if (bl) break block18;
                        if (n3 > n2) break;
                        this.h.w += n4;
                        if (!bl) continue;
                        break;
                    }
                    n3 = this.h.w;
                    n2 = this.h.F.length;
                }
                if (n3 == n2) {
                    throw new Exception(z[0] + this.h.F.length + " " + this.g);
                }
                jb jb2 = this;
                synchronized (jb2) {
                    this.finalize();
                    this.a = 3;
                }
            }
            catch (Exception exception) {
                jb jb3 = this;
                synchronized (jb3) {
                    this.finalize();
                    ++this.a;
                }
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[1]);
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final void a(int[] nArray, int n2, int n3, int n4, int n5, int n6, int n7, int n8, int n9, int n10, int[] nArray2, boolean bl, int n11, int n12, int n13) {
        boolean bl2 = client.vh;
        try {
            int n14;
            int n15;
            block26: {
                block25: {
                    ++d;
                    if (~n8 >= -1) return;
                    n15 = 0;
                    n14 = 0;
                    if (~n4 != -1) {
                        n14 = n12 / n4 << 223739206;
                        n15 = n9 / n4 << -902154106;
                    }
                    n5 <<= 2;
                    if (bl) {
                        o = 21;
                    }
                    if (~n15 <= -1) break block25;
                    n15 = 0;
                    if (!bl2) break block26;
                }
                if (4032 < n15) {
                    n15 = 4032;
                }
            }
            int n16 = n8;
            do {
                int n17 = 0;
                int n18 = n16;
                block11: while (true) {
                    block28: {
                        block27: {
                            if (n17 >= n18) return;
                            n4 += n3;
                            n13 = n15;
                            n9 += n11;
                            n10 = n14;
                            n12 += n2;
                            if (bl2) return;
                            if (~n4 != -1) {
                                n15 = n9 / n4 << -213852602;
                                n14 = n12 / n4 << -474023130;
                            }
                            if (~n15 > -1) break block27;
                            if (4032 >= n15) break block28;
                            n15 = 4032;
                            if (!bl2) break block28;
                        }
                        n15 = 0;
                    }
                    int n19 = n14 - n10 >> -1841585212;
                    int n20 = -n13 + n15 >> 166246532;
                    int n21 = n6 >> -1249879148;
                    n13 += n6 & 0xC0000;
                    n6 += n5;
                    if (n16 >= 16) {
                        nArray[n7++] = ib.a(nArray[n7] >> -496952415, 0x7F7F7F) + (nArray2[ib.a(4032, n10) + (n13 >> 955305670)] >>> n21);
                        nArray[n7++] = (ib.a(nArray[n7], 0xFEFEFF) >> 393665345) + (nArray2[ib.a(4032, n10 += n19) - -((n13 += n20) >> 556791558)] >>> n21);
                        nArray[n7++] = (ib.a(0xFEFEFF, nArray[n7]) >> -2007060127) + (nArray2[ib.a(n10 += n19, 4032) + ((n13 += n20) >> -1069632730)] >>> n21);
                        nArray[n7++] = (ib.a(nArray[n7], 0xFEFEFF) >> -526841663) + (nArray2[((n13 += n20) >> -1891324570) + ib.a(4032, n10 += n19)] >>> n21);
                        n13 += n20;
                        n13 = (n6 & 0xC0000) + (0xFFF & n13);
                        n21 = n6 >> -580603052;
                        nArray[n7++] = (nArray2[ib.a(n10 += n19, 4032) + (n13 >> 1328606726)] >>> n21) + (ib.a(nArray[n7], 0xFEFEFE) >> 604787489);
                        nArray[n7++] = (nArray2[((n13 += n20) >> -830951482) + ib.a(4032, n10 += n19)] >>> n21) + (ib.a(nArray[n7], 0xFEFEFF) >> 310428257);
                        nArray[n7++] = (nArray2[ib.a(4032, n10 += n19) + ((n13 += n20) >> -1841159226)] >>> n21) - -(ib.a(nArray[n7], 0xFEFEFF) >> -1760233471);
                        nArray[n7++] = (nArray2[ib.a(4032, n10 += n19) + ((n13 += n20) >> 1454319654)] >>> n21) - -(ib.a(0xFEFEFF, nArray[n7]) >> 1605358369);
                        n13 += n20;
                        n13 = (0xC0000 & (n6 += n5)) + (0xFFF & n13);
                        n21 = n6 >> 1147218452;
                        nArray[n7++] = (nArray2[ib.a(4032, n10 += n19) - -(n13 >> 1983636742)] >>> n21) - -(ib.a(nArray[n7], 0xFEFEFE) >> 1637168449);
                        nArray[n7++] = (nArray2[ib.a(n10 += n19, 4032) - -((n13 += n20) >> 1901625030)] >>> n21) - -ib.a(nArray[n7] >> -256795167, 0x7F7F7F);
                        nArray[n7++] = (nArray2[ib.a(4032, n10 += n19) + ((n13 += n20) >> 1605754694)] >>> n21) - -(ib.a(nArray[n7], 0xFEFEFF) >> -216359295);
                        nArray[n7++] = (ib.a(0xFEFEFE, nArray[n7]) >> 791103809) + (nArray2[((n13 += n20) >> 371413222) + ib.a(n10 += n19, 4032)] >>> n21);
                        n13 += n20;
                        n13 = ((n6 += n5) & 0xC0000) + (n13 & 0xFFF);
                        n21 = n6 >> 711720340;
                        nArray[n7++] = (nArray2[ib.a(n10 += n19, 4032) - -(n13 >> -2780922)] >>> n21) - -(ib.a(0xFEFEFE, nArray[n7]) >> 962756193);
                        n6 += n5;
                        nArray[n7++] = (ib.a(nArray[n7], 0xFEFEFF) >> -838985215) + (nArray2[ib.a(n10 += n19, 4032) - -((n13 += n20) >> 1389805702)] >>> n21);
                        nArray[n7++] = (nArray2[ib.a(4032, n10 += n19) - -((n13 += n20) >> -1171869722)] >>> n21) - -ib.a(nArray[n7] >> 1072420929, 0x7F7F7F);
                        nArray[n7++] = (nArray2[((n13 += n20) >> 227032774) + ib.a(4032, n10 += n19)] >>> n21) + (ib.a(nArray[n7], 0xFEFEFF) >> -454180287);
                        if (!bl2) break;
                    }
                    for (int i2 = 0; n16 > i2; ++i2) {
                        nArray[n7++] = (nArray2[(n13 >> -208962138) + ib.a(n10, 4032)] >>> n21) - -(ib.a(0xFEFEFE, nArray[n7]) >> -1883934911);
                        n10 += n19;
                        n13 += n20;
                        n17 = 3;
                        n18 = i2 & 3;
                        if (bl2) continue block11;
                        if (n17 != n18) continue;
                        n21 = n6 >> -2030987340;
                        n13 = (n13 & 0xFFF) - -(0xC0000 & n6);
                        n6 += n5;
                        if (!bl2) continue;
                    }
                    break;
                }
                n16 -= 16;
            } while (!bl2);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(z[5]).append(nArray != null ? z[6] : z[7]).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(',');
            if (nArray2 != null) {
                string = z[6];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(bl).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(')').toString());
            }
            string = z[7];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(bl).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final synchronized boolean a(int n2) {
        try {
            ++e;
            if (this.a >= 2) {
                return true;
            }
            if (~this.a == -1) {
                if (null == this.c) {
                    this.c = this.b.a((byte)74, this.g);
                }
                if (0 == this.c.b) {
                    return false;
                }
                if (-2 != ~this.c.b) {
                    this.c = null;
                    ++this.a;
                    return false;
                }
            }
            if (-2 == ~this.a) {
                if (null == this.f) {
                    this.f = this.b.a(this.g.getHost(), 443, -68);
                }
                if (~this.f.b == -1) {
                    return false;
                }
                if (this.f.b != 1) {
                    this.f = null;
                    ++this.a;
                    return false;
                }
            }
            if (this.q == null) {
                try {
                    if (~this.a == -1) {
                        this.q = (DataInputStream)this.c.d;
                    }
                    if (~this.a == -2) {
                        Socket socket = (Socket)this.f.d;
                        socket.setSoTimeout(10000);
                        OutputStream outputStream = socket.getOutputStream();
                        outputStream.write(17);
                        outputStream.write(h.a(z[2] + this.g.getFile() + z[4], (byte)-104));
                        this.q = new DataInputStream(socket.getInputStream());
                    }
                    this.h.w = 0;
                }
                catch (IOException iOException) {
                    this.finalize();
                    ++this.a;
                }
            }
            if (this.l == null) {
                this.l = this.b.a(true, this, 5);
            }
            if (~this.l.b == -1) {
                return false;
            }
            if (~this.l.b != n2) {
                this.finalize();
                ++this.a;
            }
            return false;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[3] + n2 + ')');
        }
    }

    final tb a(byte by) {
        try {
            if (by > -110) {
                this.run();
            }
            ++i;
            if (3 == this.a) {
                return this.h;
            }
            return null;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[10] + by + ')');
        }
    }

    static final String a(boolean bl, String string, String string2, String string3) {
        String string4;
        block5: {
            boolean bl2 = client.vh;
            if (!bl) {
                jb.a(null, 78, -46, -87, -87, -58, -96, -121, 50, -80, null, false, -54, -83, 52);
            }
            ++j;
            int n2 = string3.indexOf(string2);
            while (~n2 != 0) {
                string4 = string3 = string3.substring(0, n2) + string + string3.substring(string2.length() + n2);
                if (!bl2) {
                    n2 = string4.indexOf(string2, n2 - -string.length());
                    if (!bl2) continue;
                }
                break block5;
            }
            string4 = string3;
        }
        return string4;
    }

    jb(c c2, URL uRL, int n2) {
        try {
            this.g = uRL;
            this.b = c2;
            this.h = new tb(n2);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[8]);
            String string = c2 != null ? z[6] : z[7];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',');
            String string2 = uRL != null ? z[6] : z[7];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(',').append(n2).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    protected final void finalize() {
        try {
            ++n;
            if (null != this.c) {
                if (null != this.c.d) {
                    try {
                        ((DataInputStream)this.c.d).close();
                    }
                    catch (Exception exception) {
                        // empty catch block
                    }
                }
                this.c = null;
            }
            if (null != this.f) {
                if (this.f.d != null) {
                    try {
                        ((Socket)this.f.d).close();
                    }
                    catch (Exception exception) {
                        // empty catch block
                    }
                }
                this.f = null;
            }
            if (null != this.q) {
                try {
                    this.q.close();
                }
                catch (Exception exception) {
                    // empty catch block
                }
                this.q = null;
            }
            this.l = null;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[9]);
        }
    }

    static {
        z = new String[]{jb.z(jb.z("o\u007f+$\u000f")), jb.z(jb.z("MZ4lZI\u00103")), jb.z(jb.z("my]Y}fz:")), jb.z(jb.z("MZ4\\\u0007")), jb.z(jb.z("-2")), jb.z(jb.z("MZ4Z\u0007")), jb.z(jb.z("\\\u001640R")), jb.z(jb.z("IMvr")), jb.z(jb.z("MZ4\"FIQn \u0007")), jb.z(jb.z("MZ4xFIYvwUB\u00103")), jb.z(jb.z("MZ4]\u0007"))};
        o = 0;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x2F);
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
                        n4 = 39;
                        break;
                    }
                    case 1: {
                        n4 = 56;
                        break;
                    }
                    case 2: {
                        n4 = 26;
                        break;
                    }
                    case 3: {
                        n4 = 30;
                        break;
                    }
                    default: {
                        n4 = 47;
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

