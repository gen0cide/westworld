/*
 * Decompiled with CFR 0.152.
 */
import java.io.IOException;

class b {
    static int w;
    static int b;
    static int i;
    static int s;
    static int o;
    String B = "";
    private o z;
    static int t;
    private int g = 0;
    private int A = 0;
    int d = 0;
    static int D;
    static int r;
    static int l;
    static byte[] v;
    private int e = 0;
    static int C;
    static int a;
    static int n;
    static int x;
    static int u;
    private int m = 5000;
    static int y;
    private int E = 0;
    static int c;
    static int[] h;
    boolean k = false;
    static int F;
    ja f;
    static int p;
    private o j;
    static nb q;
    private static final String[] G;

    final void a(int n2, boolean bl) throws IOException {
        try {
            if (!bl) {
                this.a(true);
            }
            ++i;
            if (this.k) {
                this.f.w = 3;
                this.E = 0;
                this.k = false;
                throw new IOException(this.B);
            }
            ++this.e;
            if (n2 > this.e) {
                return;
            }
            if (-1 > ~this.E) {
                this.e = 0;
                this.a(this.f.F, 0, this.E, (byte)-67);
            }
            this.f.w = 3;
            this.E = 0;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[19] + n2 + ',' + bl + ')');
        }
    }

    final void a(int n2) throws IOException {
        try {
            this.b(21294);
            ++C;
            this.a(0, true);
            if (n2 != -6924) {
                this.j = null;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[15] + n2 + ')');
        }
    }

    int b(boolean bl) throws IOException {
        try {
            if (!bl) {
                this.g = 126;
            }
            ++p;
            return 0;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[18] + bl + ')');
        }
    }

    void a(byte[] byArray, int n2, int n3, byte by) throws IOException {
        try {
            ++F;
            if (by != -67) {
                this.a(81, 91);
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[7]);
            String string = byArray != null ? G[1] : G[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(by).append(')').toString());
        }
    }

    final int a(int n2, int n3) {
        try {
            if (n2 != 507) {
                this.B = null;
            }
            ++r;
            return 0xFF & -this.j.c(n2 + -635) + n3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[16] + n2 + ',' + n3 + ')');
        }
    }

    private final int a(byte[] byArray, int n2) {
        try {
            block17: {
                ++s;
                try {
                    block19: {
                        block18: {
                            block16: {
                                ++this.g;
                                if (0 < this.d) {
                                    block15: {
                                        if (~this.d > ~this.g) break block15;
                                        break block16;
                                    }
                                    this.k = true;
                                    this.B = G[14];
                                    this.d += this.d;
                                    return 0;
                                }
                            }
                            if (this.A == n2 && this.b((byte)-124) >= 2) {
                                this.A = this.b(true);
                                if (this.A >= 160) {
                                    this.A = (-160 + this.A) * 256 - -this.b(true);
                                }
                            }
                            if (~this.A >= -1 || this.b((byte)-124) < this.A) break block17;
                            if (160 <= this.A) break block18;
                            byArray[-1 + this.A] = (byte)this.b(true);
                            if (1 >= this.A) break block19;
                            this.a(byArray, (byte)126, -1 + this.A);
                            if (!client.vh) break block19;
                        }
                        this.a(byArray, (byte)64, this.A);
                    }
                    int n3 = this.A;
                    this.g = 0;
                    this.A = 0;
                    return n3;
                }
                catch (IOException iOException) {
                    this.k = true;
                    this.B = iOException.getMessage();
                }
            }
            return 0;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[13]);
            String string = byArray != null ? G[1] : G[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(')').toString());
        }
    }

    int b(byte by) throws IOException {
        try {
            ++D;
            int n2 = -6 % ((by - -64) / 56);
            return 0;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[17] + by + ')');
        }
    }

    private final void a(byte[] byArray, byte by, int n2) throws IOException {
        if (by < 51) {
            return;
        }
        try {
            ++l;
            this.a(byArray, n2, 0, -112);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[6]);
            String string = byArray != null ? G[1] : G[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(by).append(',').append(n2).append(')').toString());
        }
    }

    final boolean a(byte by) {
        try {
            block5: {
                block4: {
                    int n2 = -119 / ((-44 - by) / 53);
                    ++o;
                    if (0 < this.E) break block4;
                    break block5;
                }
                return true;
            }
            return false;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[5] + by + ')');
        }
    }

    final void b(int n2, int n3) {
        try {
            ++w;
            if (~this.E < ~(4 * this.m / 5)) {
                try {
                    this.a(0, true);
                }
                catch (IOException iOException) {
                    this.k = true;
                    this.B = iOException.getMessage();
                }
            }
            this.f.w = this.E - -2;
            if (n3 != 0) {
                return;
            }
            this.f.c(n2, 82);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[11] + n2 + ',' + n3 + ')');
        }
    }

    final int a(int n2, ja ja2) {
        try {
            ++b;
            ja2.w = n2;
            return this.a(ja2.F, n2 ^ 0);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[3]).append(n2).append(',');
            String string = ja2 != null ? G[1] : G[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    void a(boolean bl) {
        try {
            ++n;
            if (!bl) {
                this.a(116, (ja)null);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[20] + bl + ')');
        }
    }

    final void a(byte by, int[] nArray) {
        try {
            if (by >= -68) {
                this.d = -84;
            }
            this.j = new o(nArray);
            ++u;
            this.z = new o(nArray);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[12]).append(by).append(',');
            String string = nArray != null ? G[1] : G[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    void a(byte[] byArray, int n2, int n3, int n4) throws IOException {
        try {
            int n5 = -61 / ((-25 - n4) / 35);
            ++y;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[0]);
            String string = byArray != null ? G[1] : G[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void b(int n2) {
        try {
            block13: {
                int n3;
                block12: {
                    if (this.z != null) {
                        n3 = 0xFF & this.f.F[this.E + 2];
                        this.f.F[2 + this.E] = (byte)(this.z.c(-83) + n3);
                    }
                    ++x;
                    n3 = this.f.w + -this.E - 2;
                    if (-161 < ~n3) break block12;
                    this.f.F[this.E] = (byte)(160 + n3 / 256);
                    this.f.F[1 + this.E] = (byte)ib.a(n3, 255);
                    if (!client.vh) break block13;
                }
                this.f.F[this.E] = (byte)n3;
                --this.f.w;
                this.f.F[1 + this.E] = this.f.F[this.f.w];
            }
            if (-10001 <= ~this.m) {
                int n4;
                int n5 = n4 = this.f.F[this.E + 2] & 0xFF;
                ia.d[n5] = ia.d[n5] + 1;
                int n6 = n4;
                m.i[n6] = m.i[n6] + (-this.E + this.f.w);
            }
            if (n2 != 21294) {
                c = -78;
            }
            this.E = this.f.w;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[9] + n2 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final String a(int n2, byte by, String string) {
        boolean bl = client.vh;
        try {
            ++t;
            String string2 = "";
            if (by != -5) {
                b.a(null, -63, 17);
            }
            int n3 = 0;
            do {
                block23: {
                    block18: {
                        block20: {
                            block22: {
                                char c2;
                                block21: {
                                    block19: {
                                        if (n3 >= n2) return string2;
                                        if (~n3 <= ~string.length()) break block18;
                                        c2 = string.charAt(n3);
                                        if (-98 < ~c2 || c2 > 'z') break block19;
                                        string2 = string2 + c2;
                                        if (!bl) break block20;
                                    }
                                    if ('A' > c2 || -91 > ~c2) break block21;
                                    string2 = string2 + c2;
                                    if (!bl) break block20;
                                }
                                if (~c2 > -49 || ~c2 < -58) break block22;
                                string2 = string2 + c2;
                                if (!bl) break block20;
                            }
                            string2 = string2 + '_';
                        }
                        if (!bl) break block23;
                    }
                    string2 = string2 + " ";
                }
                ++n3;
            } while (!bl);
            return string2;
        }
        catch (RuntimeException runtimeException) {
            String string3;
            StringBuilder stringBuilder = new StringBuilder().append(G[10]).append(n2).append(',').append(by).append(',');
            if (string != null) {
                string3 = G[1];
                throw i.a(runtimeException, stringBuilder.append(string3).append(')').toString());
            }
            string3 = G[2];
            throw i.a(runtimeException, stringBuilder.append(string3).append(')').toString());
        }
    }

    static final void a(tb tb2, int n2, int n3) {
        try {
            if (null != q) {
                try {
                    q.a(0L, n2 ^ 0xFFFF9785);
                    q.a(24, -107, n3, tb2.F);
                }
                catch (Exception exception) {
                    // empty catch block
                }
            }
            if (n2 != 26628) {
                v = null;
            }
            ++a;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[8]);
            String string = tb2 != null ? G[1] : G[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(')').toString());
        }
    }

    protected b() {
        try {
            this.f = new ja(this.m);
            this.f.w = 3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[4]);
        }
    }

    static {
        G = new String[]{b.z(b.z("]4R\u0012")), b.z(b.z("D4:\u0014+")), b.z(b.z("QoxV")), b.z(b.z("]4E\u0012")), b.z(b.z("]4(S8Vn*\u0012\u007f")), b.z(b.z("]4G\u0012")), b.z(b.z("]4Y\u0012")), b.z(b.z("]4P\u0012")), b.z(b.z("]4@\u0012")), b.z(b.z("]4X\u0012")), b.z(b.z("]4^\u0012")), b.z(b.z("]4Z\u0012")), b.z(b.z("]4]\u0012")), b.z(b.z("]4[\u0012")), b.z(b.z("Ksy_{Po`")), b.z(b.z("]4D\u0012")), b.z(b.z("]4_\u0012")), b.z(b.z("]4V\u0012")), b.z(b.z("]4U\u0012")), b.z(b.z("]4F\u0012")), b.z(b.z("]4S\u0012"))};
        q = null;
        c = 0;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x56);
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
                        n4 = 63;
                        break;
                    }
                    case 1: {
                        n4 = 26;
                        break;
                    }
                    case 2: {
                        n4 = 20;
                        break;
                    }
                    case 3: {
                        n4 = 58;
                        break;
                    }
                    default: {
                        n4 = 86;
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

