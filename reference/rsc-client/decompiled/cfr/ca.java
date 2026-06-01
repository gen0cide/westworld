/*
 * Decompiled with CFR 0.152.
 */
import java.io.DataInputStream;
import java.io.IOException;
import java.io.InputStream;

final class ca {
    private int f;
    private int ec;
    private int[] Eb;
    static int q;
    int[] H;
    private int[] jc;
    private int xb;
    static byte[][] tb;
    private boolean c;
    static int W;
    private boolean v;
    int[] Hb;
    int[] Ob;
    private int Ib;
    int[] M;
    byte[] zb;
    boolean cb;
    int[] a;
    private int Bb;
    private int C;
    static int wb;
    private int[] Pb;
    private int[] Gb;
    private int sb;
    private int z;
    static int ac;
    private int[] Cb;
    int[] bb;
    static int Rb;
    int[] E;
    static int p;
    private int eb;
    int[] pb;
    private boolean b;
    private int P;
    boolean Kb;
    private int Sb;
    private int[] Lb;
    private int jb;
    static int d;
    private int r;
    private int g;
    int Yb;
    private int Tb;
    static int m;
    private int e;
    static int S;
    int[] qb;
    static int mb;
    static int ab;
    private int x;
    private int hb;
    static int R;
    private int K;
    int t;
    static int ub;
    private int Mb;
    private int[] w;
    private boolean Nb;
    private int T;
    static int h;
    private int[] n;
    private int[] Q;
    private int i;
    private int G;
    static int s;
    private int[] ic;
    static int D;
    int[] gb;
    static int N;
    private int yb;
    static int A;
    static int Xb;
    static int[] B;
    int[] cc;
    static int Ub;
    boolean dc;
    static int Wb;
    int[][] o;
    static int u;
    int hc;
    static int kb;
    private int U;
    int rb;
    int[] bc;
    private int j;
    static int I;
    static int nb;
    private int Vb;
    int Db;
    static int ib;
    static int Qb;
    byte[] Ab;
    private int gc;
    private int Y;
    boolean db;
    private int[] Zb;
    int[] V;
    static int L;
    static int O;
    private int[] fb;
    private int F;
    private int Fb;
    private int[] ob;
    static int y;
    static int Z;
    int Jb;
    private int X;
    int[] lb;
    static int J;
    private int[][] fc;
    static int l;
    int[] k;
    static int vb;
    private static final String[] kc;

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2) {
        boolean bl = client.vh;
        try {
            ++h;
            this.x = 999999;
            this.ec = 999999;
            this.gc = -999999;
            this.P = n2;
            this.j = -999999;
            this.e = -999999;
            this.sb = -999999;
            int n3 = 0;
            do {
                int n4;
                int n5;
                int n6;
                int n7;
                int n8;
                int n9;
                int n10;
                int n11;
                block37: {
                    if (~this.t >= ~n3) return;
                    int[] nArray = this.o[n3];
                    int n12 = this.lb[n3];
                    int n13 = nArray[0];
                    n10 = n11 = this.jc[n13];
                    n8 = n9 = this.ic[n13];
                    n6 = n7 = this.Gb[n13];
                    if (bl) return;
                    for (int i2 = 0; n12 > i2; ++i2) {
                        block41: {
                            block40: {
                                block39: {
                                    block38: {
                                        n13 = nArray[i2];
                                        n5 = ~n11;
                                        n4 = ~this.jc[n13];
                                        if (bl) break block37;
                                        if (n5 < n4) break block38;
                                        if (~n10 <= ~this.jc[n13]) break block39;
                                        n10 = this.jc[n13];
                                        if (!bl) break block39;
                                    }
                                    n11 = this.jc[n13];
                                }
                                if (this.ic[n13] >= n9) break block40;
                                n9 = this.ic[n13];
                                if (!bl) break block41;
                            }
                            if (this.ic[n13] > n8) {
                                n8 = this.ic[n13];
                            }
                        }
                        if (~this.Gb[n13] <= ~n7) {
                            if (~n6 <= ~this.Gb[n13]) continue;
                            n6 = this.Gb[n13];
                            if (!bl) continue;
                        }
                        n7 = this.Gb[n13];
                        if (!bl) continue;
                    }
                    if (!this.c) {
                        this.w[n3] = n7;
                        this.n[n3] = n6;
                        this.Q[n3] = n9;
                        this.Lb[n3] = n8;
                        this.Zb[n3] = n11;
                        this.Eb[n3] = n10;
                    }
                    n5 = -n7 + n6;
                    n4 = this.sb;
                }
                if (n5 > n4) {
                    this.sb = -n7 + n6;
                }
                if (n8 + -n9 > this.sb) {
                    this.sb = -n9 + n8;
                }
                if (this.j < n6) {
                    this.j = n6;
                }
                if (~this.gc > ~n10) {
                    this.gc = n10;
                }
                if (-n11 + n10 > this.sb) {
                    this.sb = -n11 + n10;
                }
                if (~n8 < ~this.e) {
                    this.e = n8;
                }
                if (~this.x < ~n7) {
                    this.x = n7;
                }
                if (~this.P < ~n9) {
                    this.P = n9;
                }
                if (this.ec > n11) {
                    this.ec = n11;
                }
                ++n3;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[19] + n2 + ')');
        }
    }

    final void g(int n2, int n3, int n4, int n5) {
        try {
            this.C = n2 & 0xFF;
            ++m;
            if (n3 != -999999) {
                this.a(115, -103, 21, -85, -116, -56);
            }
            this.F = n4 & 0xFF;
            this.X = n5 & 0xFF;
            this.b((byte)-117);
            this.Yb = 1;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[37] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    private final void a(int n2, ca[] caArray, boolean bl, int n3) {
        block30: {
            boolean bl2 = client.vh;
            try {
                int n4;
                int n5;
                block29: {
                    ++kb;
                    n5 = 0;
                    int n6 = 0;
                    n4 = 0;
                    while (~n3 < ~n4) {
                        n5 += caArray[n4].t;
                        n6 += caArray[n4].Db;
                        ++n4;
                        if (!bl2) {
                            if (!bl2) continue;
                            break;
                        }
                        break block29;
                    }
                    this.a(n5, n6, 88);
                }
                if (bl) {
                    this.fc = new int[n5][];
                }
                n4 = n2;
                block17: while (true) {
                    int n7 = n4;
                    block18: while (n7 < n3) {
                        ca ca2 = caArray[n4];
                        ca2.a((byte)-28);
                        this.Ib = ca2.Ib;
                        this.g = ca2.g;
                        this.Bb = ca2.Bb;
                        this.Mb = ca2.Mb;
                        this.Fb = ca2.Fb;
                        this.Jb = ca2.Jb;
                        if (bl2) break block30;
                        int n8 = 0;
                        while (~n8 > ~ca2.t) {
                            block34: {
                                block33: {
                                    int n9;
                                    block35: {
                                        block32: {
                                            block31: {
                                                int[] nArray = new int[ca2.lb[n8]];
                                                int[] nArray2 = ca2.o[n8];
                                                n7 = 0;
                                                if (bl2) continue block18;
                                                int n10 = n7;
                                                while (~n10 > ~ca2.lb[n8]) {
                                                    nArray[n10] = this.e(ca2.a[nArray2[n10]], ca2.bc[nArray2[n10]], ca2.ob[nArray2[n10]], -122);
                                                    ++n10;
                                                    if (!bl2) {
                                                        if (!bl2) continue;
                                                        break;
                                                    }
                                                    break block31;
                                                }
                                                n10 = this.a(ca2.lb[n8], nArray, ca2.V[n8], ca2.qb[n8], false);
                                                this.Hb[n10] = ca2.Hb[n8];
                                                this.M[n10] = ca2.M[n8];
                                                this.k[n10] = ca2.k[n8];
                                            }
                                            if (bl) break block32;
                                            break block33;
                                        }
                                        if (1 < n3) break block35;
                                        this.fc[n10] = new int[ca2.fc[n8].length];
                                        n9 = 0;
                                        while (~ca2.fc[n8].length < ~n9) {
                                            this.fc[n10][n9] = ca2.fc[n8][n9];
                                            ++n9;
                                            if (!bl2) {
                                                if (!bl2) continue;
                                                break;
                                            }
                                            break block34;
                                        }
                                        if (!bl2) break block33;
                                    }
                                    this.fc[n10] = new int[ca2.fc[n8].length - -1];
                                    this.fc[n10][0] = n4;
                                    for (n9 = 0; n9 < ca2.fc[n8].length; ++n9) {
                                        this.fc[n10][1 + n9] = ca2.fc[n8][n9];
                                        if (!bl2) {
                                            if (!bl2) continue;
                                            break;
                                        }
                                        break block34;
                                    }
                                }
                                ++n8;
                            }
                            if (!bl2) continue;
                        }
                        ++n4;
                        if (!bl2) continue block17;
                    }
                    break;
                }
                this.Yb = 1;
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(kc[35]).append(n2).append(',');
                String string = caArray != null ? kc[1] : kc[3];
                throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(bl).append(',').append(n3).append(')').toString());
            }
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int n3, int n4) {
        try {
            block19: {
                block18: {
                    if (!this.db) {
                        this.zb = new byte[n2];
                        this.E = new int[n2];
                    }
                    this.k = new int[n2];
                    this.ob = new int[n3];
                    this.Ab = new byte[n3];
                    this.a = new int[n3];
                    this.M = new int[n2];
                    this.bc = new int[n3];
                    this.qb = new int[n2];
                    this.gb = new int[n3];
                    this.V = new int[n2];
                    this.o = new int[n2][];
                    this.lb = new int[n2];
                    if (!this.b) {
                        this.Ob = new int[n3];
                        this.H = new int[n3];
                        this.cc = new int[n3];
                        this.bb = new int[n3];
                        this.pb = new int[n3];
                    }
                    ++nb;
                    this.Hb = new int[n2];
                    this.Y = 256;
                    this.X = 0;
                    this.U = 256;
                    this.Tb = 256;
                    this.T = 256;
                    if (!this.c) {
                        this.n = new int[n2];
                        this.Q = new int[n2];
                        this.Zb = new int[n2];
                        this.Lb = new int[n2];
                        this.Eb = new int[n2];
                        this.w = new int[n2];
                    }
                    if (!this.Nb || !this.c) {
                        this.Cb = new int[n2];
                        this.Pb = new int[n2];
                        this.fb = new int[n2];
                    }
                    this.G = 256;
                    this.t = 0;
                    this.F = 0;
                    this.Sb = 0;
                    this.yb = 256;
                    if (this.v) break block18;
                    this.Gb = new int[n3];
                    this.jc = new int[n3];
                    this.ic = new int[n3];
                    if (!client.vh) break block19;
                }
                this.Gb = this.a;
                this.jc = this.bc;
                this.ic = this.ob;
            }
            this.r = 0;
            this.K = n3;
            this.eb = 256;
            this.f = 256;
            this.xb = 0;
            this.jb = 0;
            this.Db = 0;
            this.i = 256;
            if (n4 <= 68) {
                this.lb = null;
            }
            this.z = n2;
            this.C = 0;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[29] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    final void a(int n2, int n3, int n4, int n5, int n6, int n7) {
        try {
            ++Ub;
            this.Jb = -(4 * n3) + 256;
            this.Mb = (64 - n2) * 16 + 128;
            if (n5 > -110) {
                this.Bb = -67;
            }
            if (this.Nb) {
                return;
            }
            this.g = n6;
            this.Fb = n7;
            this.Bb = n4;
            this.Ib = (int)Math.sqrt(n7 * n7 + (n4 * n4 + n6 * n6));
            this.e(-102);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[14] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ')');
        }
    }

    /*
     * Unable to fully structure code
     */
    final ca[] a(int var1_1, int var2_2, int var3_3, int var4_4, int var5_5, int var6_6, int var7_7, boolean var8_8, int var9_9) {
        var19_10 = client.vh;
        try {
            block29: {
                block28: {
                    block24: {
                        block23: {
                            ++ca.Rb;
                            this.a((byte)-28);
                            var10_11 = new int[var5_5];
                            var11_13 = new int[var5_5];
                            var12_14 = 0;
                            while (~var12_14 > ~var5_5) {
                                var10_11[var12_14] = 0;
                                var11_13[var12_14] = 0;
                                ++var12_14;
                                if (!var19_10) {
                                    if (!var19_10) continue;
                                    break;
                                }
                                break block23;
                            }
                            var12_14 = 0;
                        }
                        while (var12_14 < this.t) {
                            block25: {
                                var13_16 = 0;
                                var14_17 = 0;
                                var15_18 = this.lb[var12_14];
                                var16_19 = this.o[var12_14];
                                v1 = 0;
                                if (var19_10) break block24;
                                var17_21 = v1;
                                while (~var17_21 > ~var15_18) {
                                    var13_16 += this.a[var16_19[var17_21]];
                                    var14_17 += this.bc[var16_19[var17_21]];
                                    ++var17_21;
                                    if (!var19_10) {
                                        if (!var19_10) continue;
                                        break;
                                    }
                                    break block25;
                                }
                                v3 = var17_21 = var13_16 / (var15_18 * var7_7) + var14_17 / (var3_3 * var15_18) * var2_2;
                                var10_11[v3] = var10_11[v3] + var15_18;
                                v4 = var17_21;
                                var11_13[v4] = var11_13[v4] + 1;
                                ++var12_14;
                            }
                            if (!var19_10) continue;
                        }
                        v1 = var5_5;
                    }
                    var12_15 = new ca[v1];
                    var13_16 = 0;
                    while (~var13_16 > ~var5_5) {
                        block27: {
                            block26: {
                                v5 = var6_6;
                                v6 = var10_11[var13_16];
                                if (!var19_10) {
                                    if (v5 < v6) break block26;
                                    break block27;
                                }
                                ** GOTO lbl73
                            }
                            var10_11[var13_16] = var6_6;
                        }
                        var12_15[var13_16] = new ca(var10_11[var13_16], var11_13[var13_16], true, true, true, var8_8, true);
                        var12_15[var13_16].Mb = this.Mb;
                        var12_15[var13_16].Jb = this.Jb;
                        ++var13_16;
                        if (!var19_10) continue;
                    }
                    var13_16 = 0;
                    block16: do {
                        v5 = this.t;
                        v6 = var13_16;
lbl73:
                        // 2 sources

                        if (v5 <= v6) break;
                        var14_17 = 0;
                        var15_18 = 0;
                        var16_20 = this.lb[var13_16];
                        var17_23 = this.o[var13_16];
                        v8 = 0;
                        if (var19_10) break block28;
                        var18_24 = v8;
                        while (~var16_20 < ~var18_24) {
                            var14_17 += this.a[var17_23[var18_24]];
                            var15_18 += this.bc[var17_23[var18_24]];
                            ++var18_24;
                            if (var19_10) continue block16;
                            if (!var19_10) continue;
                            break;
                        }
                        var18_24 = var14_17 / (var16_20 * var7_7) + var2_2 * (var15_18 / (var3_3 * var16_20));
                        this.a(var17_23, var12_15[var18_24], var16_20, var13_16, 5916);
                        ++var13_16;
                    } while (!var19_10);
                    var13_16 = 0;
                    v8 = var14_17 = -50 / ((var4_4 - -33) / 60);
                }
                while (~var13_16 > ~var5_5) {
                    v10 = var12_15;
                    if (!var19_10) {
                        v10[var13_16].c((byte)71);
                        ++var13_16;
                        if (!var19_10) continue;
                        break;
                    }
                    break block29;
                }
                v10 = var12_15;
            }
            return v10;
        }
        catch (RuntimeException var10_12) {
            throw i.a(var10_12, ca.kc[23] + var1_1 + ',' + var2_2 + ',' + var3_3 + ',' + var4_4 + ',' + var5_5 + ',' + var6_6 + ',' + var7_7 + ',' + var8_8 + ',' + var9_9 + ')');
        }
    }

    final int a(int n2, int[] nArray, int n3, int n4, boolean bl) {
        try {
            block8: {
                block7: {
                    ++u;
                    if (bl) {
                        this.f(30, -84, 10, 23);
                    }
                    if (~this.t <= ~this.z) break block7;
                    break block8;
                }
                return -1;
            }
            this.lb[this.t] = n2;
            this.o[this.t] = nArray;
            this.V[this.t] = n3;
            this.qb[this.t] = n4;
            this.Yb = 1;
            return this.t++;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(kc[2]).append(n2).append(',');
            String string = nArray != null ? kc[1] : kc[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n3).append(',').append(n4).append(',').append(bl).append(')').toString());
        }
    }

    final void a(int n2, int n3, int n4, boolean bl) {
        try {
            this.xb += n4;
            this.Sb += n3;
            if (!bl) {
                this.jc = null;
            }
            this.r += n2;
            ++L;
            this.b((byte)-127);
            this.Yb = 1;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[31] + n2 + ',' + n3 + ',' + n4 + ',' + bl + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int n3, int n4, int n5, int n6, int n7, byte by) {
        boolean bl = client.vh;
        try {
            ++S;
            int n8 = 118 / ((-80 - by) / 46);
            int i2 = 0;
            do {
                if (i2 >= this.Db) return;
                if (bl) return;
                if (-1 != ~n4) {
                    int n9 = i2;
                    this.Gb[n9] = this.Gb[n9] + (this.ic[i2] * n4 >> 1095852296);
                }
                if (0 != n5) {
                    int n10 = i2;
                    this.jc[n10] = this.jc[n10] + (n5 * this.ic[i2] >> -706532120);
                }
                if (-1 != ~n6) {
                    int n11 = i2;
                    this.Gb[n11] = this.Gb[n11] + (n6 * this.jc[i2] >> -324256440);
                }
                if (-1 != ~n3) {
                    int n12 = i2;
                    this.ic[n12] = this.ic[n12] + (n3 * this.jc[i2] >> 423792200);
                }
                if (-1 != ~n2) {
                    int n13 = i2;
                    this.jc[n13] = this.jc[n13] + (n2 * this.Gb[i2] >> -1731094840);
                }
                if (-1 != ~n7) {
                    int n14 = i2;
                    this.ic[n14] = this.ic[n14] + (this.Gb[i2] * n7 >> 1779723400);
                }
                ++i2;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[16] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ',' + by + ')');
        }
    }

    final void b(int n2, int n3, int n4) {
        try {
            ++ub;
            this.t -= n4;
            if (n3 > -110) {
                ca.a((byte)-3, null);
            }
            if (~this.t > -1) {
                this.t = 0;
            }
            this.Db -= n2;
            if (this.Db < 0) {
                this.Db = 0;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[18] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    final void c(int n2, int n3, int n4, int n5) {
        try {
            this.xb = n2;
            this.r = n5;
            this.Sb = n4;
            if (n3 > -112) {
                this.a(-96, (int[])null, -8, 42, true);
            }
            ++q;
            this.b((byte)-114);
            this.Yb = 1;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[21] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, int n3, int n4, byte by, int n5, int n6, int n7, int n8, int n9) {
        boolean bl = client.vh;
        try {
            block34: {
                block33: {
                    ++ac;
                    this.d(7972);
                    if (~da.K > ~this.ec || m.j > this.gc || ~this.x < ~oa.b || aa.f > this.j || this.P > nb.y || this.e < aa.b) break block33;
                    this.dc = true;
                    if (!bl) break block34;
                }
                this.dc = false;
                return;
            }
            int n10 = 0;
            if (by > -105) {
                this.c((byte)-103);
            }
            int n11 = 0;
            int n12 = 0;
            int n13 = 0;
            int n14 = 0;
            if (n7 != 0) {
                n11 = pa.j[1024 + n7];
                n10 = pa.j[n7];
            }
            int n15 = 0;
            if (0 != n8) {
                n12 = pa.j[n8];
                n13 = pa.j[n8 - -1024];
            }
            if (n6 != 0) {
                n14 = pa.j[n6];
                n15 = pa.j[n6 - -1024];
            }
            int n16 = 0;
            do {
                int n17;
                int n18;
                int n19;
                block38: {
                    block37: {
                        block36: {
                            block35: {
                                int n20;
                                if (this.Db <= n16) return;
                                n19 = this.Gb[n16] + -n4;
                                n18 = -n2 + this.ic[n16];
                                if (bl) return;
                                if (n7 != 0) {
                                    n20 = n18 * n10 - -(n11 * n19) >> 1285023887;
                                    n18 = -(n19 * n10) + n18 * n11 >> 1814361327;
                                    n19 = n20;
                                }
                                n17 = this.jc[n16] - n5;
                                if (~n6 != -1) {
                                    n20 = n15 * n19 + n17 * n14 >> 1043773743;
                                    n17 = -(n19 * n14) + n15 * n17 >> -1543110961;
                                    n19 = n20;
                                }
                                if (~n8 != -1) {
                                    n20 = n18 * n13 + -(n12 * n17) >> 2015943439;
                                    n17 = n12 * n18 - -(n13 * n17) >> -446444369;
                                    n18 = n20;
                                }
                                if (n17 >= n9) break block35;
                                this.pb[n16] = n19 << n3;
                                if (!bl) break block36;
                            }
                            this.pb[n16] = (n19 << n3) / n17;
                        }
                        if (n17 >= n9) break block37;
                        this.Ob[n16] = n18 << n3;
                        if (!bl) break block38;
                    }
                    this.Ob[n16] = (n18 << n3) / n17;
                }
                this.cc[n16] = n19;
                this.H[n16] = n18;
                this.bb[n16] = n17;
                ++n16;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[12] + n2 + ',' + n3 + ',' + n4 + ',' + by + ',' + n5 + ',' + n6 + ',' + n7 + ',' + n8 + ',' + n9 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void b(byte by) {
        try {
            block23: {
                block24: {
                    block22: {
                        block21: {
                            block19: {
                                block20: {
                                    block18: {
                                        ++vb;
                                        if (by >= -103) {
                                            this.a(true, -115, true, true, false);
                                        }
                                        if (256 != this.Tb || this.G != 256 || ~this.Y != -257 || this.f != 256 || -257 != ~this.U || -257 != ~this.eb) break block18;
                                        if (-257 != ~this.yb) break block19;
                                        break block20;
                                    }
                                    this.jb = 4;
                                    return;
                                }
                                if (this.i == 256 && -257 == ~this.T) break block21;
                            }
                            this.jb = 3;
                            return;
                        }
                        if (0 != this.F || 0 != this.X || 0 != this.C) break block22;
                        if (this.r != 0) break block23;
                        break block24;
                    }
                    this.jb = 2;
                    return;
                }
                if (this.xb == 0 && -1 == ~this.Sb) {
                    this.jb = 0;
                    return;
                }
            }
            this.jb = 1;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[34] + by + ')');
        }
    }

    private final void a(byte by) {
        block9: {
            boolean bl = client.vh;
            try {
                this.d(7972);
                ++D;
                int n2 = 0;
                while (~this.Db < ~n2) {
                    this.a[n2] = this.Gb[n2];
                    this.ob[n2] = this.ic[n2];
                    this.bc[n2] = this.jc[n2];
                    ++n2;
                    if (!bl) {
                        if (!bl) continue;
                        break;
                    }
                    break block9;
                }
                if (by != -28) {
                    this.d(4);
                }
                this.C = 0;
                this.F = 0;
                this.U = 256;
                this.f = 256;
                this.xb = 0;
                this.X = 0;
                this.eb = 256;
                this.Y = 256;
                this.yb = 256;
                this.jb = 0;
                this.r = 0;
                this.i = 256;
                this.Tb = 256;
                this.T = 256;
                this.Sb = 0;
                this.G = 256;
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, kc[27] + by + ')');
            }
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, int n3, int n4, int n5, boolean bl, int n6, int n7) {
        boolean bl2 = client.vh;
        try {
            this.Mb = (-n6 + 64) * 16 + 128;
            int n8 = 76 / ((-8 - n7) / 49);
            this.Jb = 256 - n3 * 4;
            ++R;
            if (this.Nb) {
                return;
            }
            int n9 = 0;
            while (~this.t < ~n9) {
                block11: {
                    block10: {
                        if (bl2) return;
                        if (!bl) break block10;
                        this.Hb[n9] = this.Vb;
                        if (!bl2) break block11;
                    }
                    this.Hb[n9] = 0;
                }
                ++n9;
                if (!bl2) continue;
            }
            this.g = n5;
            this.Fb = n2;
            this.Bb = n4;
            this.Ib = (int)Math.sqrt(n4 * n4 + (n5 * n5 - -(n2 * n2)));
            this.e(-121);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[28] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + bl + ',' + n6 + ',' + n7 + ')');
        }
    }

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    private final void b(int n2, int n3, int n4, int n5) {
        boolean bl = client.vh;
        try {
            ++O;
            if (n3 != -27483) {
                this.e(-7, -82, -31, -24);
            }
            int n6 = 0;
            while (~n6 > ~this.Db) {
                try {
                    this.Gb[n6] = this.Gb[n6] * n2 >> -1363775576;
                    this.ic[n6] = this.ic[n6] * n5 >> -895455096;
                    this.jc[n6] = n4 * this.jc[n6] >> -242758104;
                    ++n6;
                    if (!bl && !bl) continue;
                    return;
                }
                catch (RuntimeException runtimeException) {
                    throw runtimeException;
                    return;
                }
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[22] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final int e(int n2, int n3, int n4, int n5) {
        boolean bl = client.vh;
        try {
            int n6;
            int n7;
            block11: {
                ++ib;
                int n8 = 0;
                while (~n8 > ~this.Db) {
                    n7 = ~this.a[n8];
                    n6 = ~n2;
                    if (!bl) {
                        if (n7 == n6 && n4 == this.ob[n8] && n3 == this.bc[n8]) {
                            return n8;
                        }
                        ++n8;
                        if (!bl) continue;
                    }
                    break block11;
                }
                n7 = 100;
                n6 = (-46 - n5) / 58;
            }
            int n9 = n7 / n6;
            if (this.Db < this.K) {
                this.a[this.Db] = n2;
                this.ob[this.Db] = n4;
                this.bc[this.Db] = n3;
                return this.Db++;
            }
            return -1;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[13] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int[] nArray, ca ca2, int n2, int n3, int n4) {
        boolean bl = client.vh;
        try {
            block13: {
                ++mb;
                int[] nArray2 = new int[n2];
                int n5 = 0;
                while (~n2 < ~n5) {
                    int n6 = nArray2[n5] = ca2.e(this.a[nArray[n5]], this.bc[nArray[n5]], this.ob[nArray[n5]], 107);
                    ca2.gb[n6] = this.gb[nArray[n5]];
                    ca2.Ab[n6] = this.Ab[nArray[n5]];
                    ++n5;
                    if (!bl) {
                        if (!bl) continue;
                    }
                    break block13;
                }
                if (n4 != 5916) {
                    this.yb = 77;
                }
                n5 = ca2.a(n2, nArray2, this.V[n3], this.qb[n3], false);
            }
            if (!ca2.db && !this.db) {
                ca2.E[n5] = this.E[n3];
            }
            ca2.Hb[n5] = this.Hb[n3];
            ca2.M[n5] = this.M[n3];
            ca2.k[n5] = this.k[n3];
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(kc[33]).append(nArray != null ? kc[1] : kc[3]).append(',');
            if (ca2 != null) {
                string = kc[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(')').toString());
            }
            string = kc[3];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(')').toString());
        }
    }

    final void c(int n2) {
        try {
            this.t = 0;
            if (n2 != 1) {
                this.Kb = true;
            }
            this.Db = 0;
            ++W;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[0] + n2 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int n3, int n4, int n5) {
        boolean bl = client.vh;
        try {
            ++ab;
            if (n2 >= -14) {
                this.Cb = null;
            }
            int n6 = 0;
            do {
                int n7;
                int n8;
                int n9;
                if (~this.Db >= ~n6) return;
                if (bl) return;
                if (n4 != 0) {
                    n9 = pa.a[n4 + 256];
                    n8 = pa.a[n4];
                    n7 = this.Gb[n6] * n9 + this.ic[n6] * n8 >> -1049305169;
                    this.ic[n6] = -(n8 * this.Gb[n6]) + this.ic[n6] * n9 >> -1441854257;
                    this.Gb[n6] = n7;
                }
                if (-1 != ~n3) {
                    n8 = pa.a[n3];
                    n9 = pa.a[256 + n3];
                    n7 = -(n8 * this.jc[n6]) + n9 * this.ic[n6] >> 1869232655;
                    this.jc[n6] = n8 * this.ic[n6] - -(n9 * this.jc[n6]) >> -1067157009;
                    this.ic[n6] = n7;
                }
                if (-1 != ~n5) {
                    n8 = pa.a[n5];
                    n9 = pa.a[256 + n5];
                    n7 = n8 * this.jc[n6] + this.Gb[n6] * n9 >> -1889761873;
                    this.jc[n6] = -(this.Gb[n6] * n8) + this.jc[n6] * n9 >> 89169167;
                    this.Gb[n6] = n7;
                }
                ++n6;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[17] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    /*
     * Unable to fully structure code
     */
    private final void e(int var1_1) {
        block31: {
            var11_2 = client.vh;
            try {
                block28: {
                    block27: {
                        ++ca.l;
                        if (this.Nb) {
                            return;
                        }
                        var2_3 = this.Mb * this.Ib >> 434666280;
                        for (var3_5 = 0; this.t > var3_5; ++var3_5) {
                            block34: {
                                v0 = this.Hb[var3_5];
                                if (var11_2) break block27;
                                if (v0 == this.Vb) continue;
                                break block34;
                                catch (RuntimeException v1) {
                                    throw v1;
                                }
                            }
                            this.Hb[var3_5] = (this.Cb[var3_5] * this.Bb + (this.fb[var3_5] * this.g - -(this.Fb * this.Pb[var3_5]))) / var2_3;
                            continue;
                        }
                        v0 = this.Db;
                    }
                    var3_6 = new int[v0];
                    var4_7 = new int[this.Db];
                    var5_8 = new int[this.Db];
                    var6_9 = new int[this.Db];
                    for (var7_10 = 0; this.Db > var7_10; ++var7_10) {
                        var3_6[var7_10] = 0;
                        var4_7[var7_10] = 0;
                        var5_8[var7_10] = 0;
                        var6_9[var7_10] = 0;
                        if (!var11_2) {
                            if (!var11_2) continue;
                            break;
                        }
                        break block28;
                    }
                    var7_10 = -16 / ((var1_1 - -55) / 32);
                }
                var8_11 = 0;
                while (var8_11 < this.t) {
                    block30: {
                        block29: {
                            v4 = ~this.Hb[var8_11];
                            v5 = ~this.Vb;
                            if (!var11_2) {
                                if (v4 != v5) break block29;
                            }
                            ** GOTO lbl78
                            for (var9_12 = 0; this.lb[var8_11] > var9_12; ++var9_12) {
                                var10_13 = this.o[var8_11][var9_12];
                                v7 = var10_13;
                                var3_6[v7] = var3_6[v7] + this.fb[var8_11];
                                v8 = var10_13;
                                var4_7[v8] = var4_7[v8] + this.Cb[var8_11];
                                v9 = var10_13;
                                var5_8[v9] = var5_8[v9] + this.Pb[var8_11];
                                v10 = var10_13;
                                var6_9[v10] = var6_9[v10] + 1;
                                if (!var11_2) {
                                    if (!var11_2) continue;
                                    break;
                                }
                                break block30;
                            }
                        }
                        ++var8_11;
                    }
                    if (!var11_2) continue;
                }
                var8_11 = 0;
                do {
                    block33: {
                        block32: {
                            block35: {
                                v4 = this.Db;
                                v5 = var8_11;
lbl78:
                                // 3 sources

                                if (v4 <= v5) break block31;
                                if (var11_2) break block31;
                                break block35;
                                catch (RuntimeException v12) {
                                    throw v12;
                                }
                            }
                            if (0 < var6_9[var8_11]) break block32;
                            break block33;
                            catch (RuntimeException v13) {
                                throw v13;
                            }
                        }
                        this.gb[var8_11] = (var5_8[var8_11] * this.Fb + (var3_6[var8_11] * this.g + var4_7[var8_11] * this.Bb)) / (var2_3 * var6_9[var8_11]);
                    }
                    ++var8_11;
                } while (!var11_2);
            }
            catch (RuntimeException var2_4) {
                throw i.a(var2_4, ca.kc[9] + var1_1 + ')');
            }
        }
    }

    final ca a(boolean bl, int n2, boolean bl2, boolean bl3, boolean bl4) {
        try {
            ++I;
            ca[] caArray = new ca[]{this};
            int n3 = 122 / ((-56 - n2) / 59);
            ca ca2 = new ca(caArray, 1, bl3, bl4, bl2, bl);
            ca2.hc = this.hc;
            return ca2;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[11] + bl + ',' + n2 + ',' + bl2 + ',' + bl3 + ',' + bl4 + ')');
        }
    }

    static final int a(byte by, String string) {
        boolean bl = client.vh;
        try {
            int n2;
            block9: {
                ++Xb;
                if (string.equalsIgnoreCase(kc[5])) {
                    return 0;
                }
                if (by != 91) {
                    B = null;
                }
                for (int i2 = 0; i2 < ia.b; ++i2) {
                    n2 = ub.c[i2].equalsIgnoreCase(string);
                    if (bl) break block9;
                    if (n2 == 0) continue;
                    return i2;
                }
                ub.c[ia.b++] = string;
                n2 = ia.b + -1;
            }
            return n2;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(kc[4]).append(by).append(',');
            String string2 = string != null ? kc[1] : kc[3];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(')').toString());
        }
    }

    final void a(ca ca2, int n2) {
        try {
            if (n2 != 6029) {
                this.jb = -128;
            }
            this.xb = ca2.xb;
            ++y;
            this.C = ca2.C;
            this.Sb = ca2.Sb;
            this.r = ca2.r;
            this.X = ca2.X;
            this.F = ca2.F;
            this.b((byte)-113);
            this.Yb = 1;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(kc[26]);
            String string = ca2 != null ? kc[1] : kc[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(')').toString());
        }
    }

    final ca b(int n2) {
        try {
            ++wb;
            ca[] caArray = new ca[1];
            if (n2 != -2) {
                this.b(117, -93, -34, -34);
            }
            caArray[0] = this;
            ca ca2 = new ca(caArray, 1);
            ca2.cb = this.cb;
            ca2.hc = this.hc;
            return ca2;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[10] + n2 + ')');
        }
    }

    private final void c(byte by) {
        try {
            if (by < 49) {
                this.a(40, 102, 104, 108, -20, -89);
            }
            this.bb = new int[this.Db];
            this.Ob = new int[this.Db];
            this.cc = new int[this.Db];
            this.H = new int[this.Db];
            this.pb = new int[this.Db];
            ++N;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[30] + by + ')');
        }
    }

    final void a(int n2, int n3, byte by) {
        try {
            if (by != -61) {
                return;
            }
            ++Wb;
            this.Ab[n2] = (byte)n3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[8] + n2 + ',' + n3 + ',' + by + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void d(int n2) {
        boolean bl = client.vh;
        try {
            block27: {
                block26: {
                    int n3;
                    if (n2 != 7972) {
                        this.a(120, 20, 57, true);
                    }
                    ++Qb;
                    if (2 == this.Yb) {
                        block25: {
                            this.Yb = 0;
                            n3 = 0;
                            while (~this.Db < ~n3) {
                                this.Gb[n3] = this.a[n3];
                                this.ic[n3] = this.ob[n3];
                                this.jc[n3] = this.bc[n3];
                                ++n3;
                                if (!bl) {
                                    if (!bl) continue;
                                }
                                break block25;
                            }
                            this.j = 9999999;
                            this.e = 9999999;
                            this.P = -9999999;
                            this.gc = 9999999;
                            this.ec = -9999999;
                            this.sb = 9999999;
                            this.x = -9999999;
                        }
                        if (!bl) return;
                    }
                    if (~this.Yb != -2) {
                        return;
                    }
                    this.Yb = 0;
                    n3 = 0;
                    while (~n3 > ~this.Db) {
                        this.Gb[n3] = this.a[n3];
                        this.ic[n3] = this.ob[n3];
                        this.jc[n3] = this.bc[n3];
                        ++n3;
                        if (!bl) {
                            if (!bl) continue;
                        }
                        break block26;
                    }
                    if (2 > this.jb) break block27;
                }
                this.a(-53, this.F, this.C, this.X);
            }
            if (this.jb >= 3) {
                this.b(this.yb, -27483, this.T, this.i);
            }
            if (4 <= this.jb) {
                this.a(this.U, this.f, this.Tb, this.G, this.Y, this.eb, (byte)-127);
            }
            if (1 <= this.jb) {
                this.d(n2 + -7972, this.xb, this.Sb, this.r);
            }
            this.a(999999);
            this.d((byte)14);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[25] + n2 + ')');
        }
    }

    final int b(boolean bl, int n2, int n3, int n4) {
        try {
            ++J;
            if (~this.K >= ~this.Db) {
                return -1;
            }
            this.a[this.Db] = n3;
            this.ob[this.Db] = n4;
            this.bc[this.Db] = n2;
            if (bl) {
                ca.a((byte)52, null);
            }
            return this.Db++;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[32] + bl + ',' + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void d(byte by) {
        boolean bl = client.vh;
        try {
            ++Z;
            if (this.Nb) {
                if (this.c) return;
            }
            if (by != 14) {
                return;
            }
            for (int i2 = 0; this.t > i2; ++i2) {
                int n2;
                int[] nArray = this.o[i2];
                int n6 = this.Gb[nArray[0]];
                int n7 = this.ic[nArray[0]];
                int n8 = this.jc[nArray[0]];
                int n9 = -n6 + this.Gb[nArray[1]];
                int n10 = this.ic[nArray[1]] + -n7;
                int n11 = -n8 + this.jc[nArray[1]];
                int n12 = -n6 + this.Gb[nArray[2]];
                int n13 = this.ic[nArray[2]] - n7;
                int n14 = -n8 + this.jc[nArray[2]];
                int n5 = n14 * n10 - n11 * n13;
                int n4 = -(n9 * n14) + n11 * n12;
                if (bl) return;
                int n3 = -(n12 * n10) + n9 * n13;
                do {
                    int n15;
                    int n16;
                    if (8192 >= n5) {
                        n16 = ~n4;
                        n15 = -8193;
                        if (!bl) {
                            if (n16 >= n15 && n3 <= 8192 && n5 >= -8192 && n4 >= -8192 && -8192 <= n3) break;
                        }
                    } else {
                        n4 >>= 1;
                        n3 >>= 1;
                        n16 = n5;
                        n15 = 1;
                    }
                    n5 = n16 >> n15;
                } while (!bl);
                if ((n2 = (int)(Math.sqrt(n4 * n4 + (n5 * n5 + n3 * n3)) * 256.0)) <= 0) {
                    n2 = 1;
                }
                this.fb[i2] = n5 * 65536 / n2;
                this.Cb[i2] = 65536 * n4 / n2;
                this.Pb[i2] = n3 * 65535 / n2;
                this.M[i2] = -1;
                if (!bl) continue;
            }
            this.e(by ^ 0xFFFFFFAB);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[36] + by + ')');
        }
    }

    final void f(int n2, int n3, int n4, int n5) {
        try {
            this.F = n5 + this.F & 0xFF;
            ++A;
            this.X = n4 + this.X & 0xFF;
            if (n3 != -31616) {
                return;
            }
            this.C = 0xFF & this.C - -n2;
            this.b((byte)-105);
            this.Yb = 1;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[7] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    private final void d(int n2, int n3, int n4, int n5) {
        boolean bl = client.vh;
        try {
            ++d;
            int n6 = n2;
            while (this.Db > n6) {
                try {
                    int n7 = n6;
                    this.Gb[n7] = this.Gb[n7] + n5;
                    int n8 = n6;
                    this.ic[n8] = this.ic[n8] + n3;
                    int n9 = n6++;
                    this.jc[n9] = this.jc[n9] + n4;
                    if (!bl && !bl) continue;
                    return;
                }
                catch (RuntimeException runtimeException) {
                    throw runtimeException;
                    return;
                }
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[15] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final int a(byte by, byte[] byArray) {
        boolean bl = client.vh;
        try {
            int n2;
            int n3;
            if (by != 76) {
                return 108;
            }
            ++s;
            while (~byArray[this.hb] == -11 || -14 == ~byArray[this.hb]) {
                ++this.hb;
                if (!bl) continue;
            }
            int n4 = pa.d[0xFF & byArray[this.hb++]];
            int n5 = pa.d[0xFF & byArray[this.hb++]];
            if (-123457 == ~(n3 = -131072 + n5 * 64 + (4096 * n4 + (n2 = pa.d[byArray[this.hb++] & 0xFF])))) return this.Vb;
            return n3;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(kc[20]).append(by).append(',');
            if (byArray != null) {
                string = kc[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = kc[3];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    final void a(boolean bl, int n2, int n3, int n4) {
        try {
            block6: {
                block5: {
                    if (bl) {
                        this.Yb = 71;
                    }
                    ++p;
                    if (this.Nb) break block5;
                    break block6;
                }
                return;
            }
            this.Fb = n4;
            this.Bb = n3;
            this.g = n2;
            this.Ib = (int)Math.sqrt(n4 * n4 + (n3 * n3 + n2 * n2));
            this.e(52);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[24] + bl + ',' + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    ca(int n2, int n3) {
        boolean bl = client.vh;
        this.Bb = 155;
        this.c = false;
        this.Yb = 1;
        this.b = false;
        this.Ib = 256;
        this.v = false;
        this.cb = false;
        this.Nb = false;
        this.sb = 12345678;
        this.g = 180;
        this.dc = true;
        this.Vb = 12345678;
        this.Kb = false;
        this.rb = -1;
        this.Mb = 512;
        this.Fb = 95;
        this.db = false;
        this.Jb = 32;
        this.hc = 0;
        try {
            this.a(n3, n2, 69);
            this.fc = new int[n3][1];
            for (int i2 = 0; i2 < n3; ++i2) {
                try {
                    this.fc[i2][0] = i2;
                    if (!bl && !bl) continue;
                    return;
                }
                catch (RuntimeException runtimeException) {
                    throw runtimeException;
                    return;
                }
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[6] + n2 + ',' + n3 + ')');
        }
    }

    ca(int n2, int n3, boolean bl, boolean bl2, boolean bl3, boolean bl4, boolean bl5) {
        this.Bb = 155;
        this.c = false;
        this.Yb = 1;
        this.b = false;
        this.Ib = 256;
        this.v = false;
        this.cb = false;
        this.Nb = false;
        this.sb = 12345678;
        this.g = 180;
        this.dc = true;
        this.Vb = 12345678;
        this.Kb = false;
        this.rb = -1;
        this.Mb = 512;
        this.Fb = 95;
        this.db = false;
        this.Jb = 32;
        this.hc = 0;
        try {
            this.c = bl2;
            this.b = bl5;
            this.db = bl4;
            this.v = bl;
            this.Nb = bl3;
            this.a(n3, n2, 69);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, kc[6] + n2 + ',' + n3 + ',' + bl + ',' + bl2 + ',' + bl3 + ',' + bl4 + ',' + bl5 + ')');
        }
    }

    /*
     * Unable to fully structure code
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    ca(byte[] var1_1, int var2_2, boolean var3_3) {
        var8_4 = client.vh;
        super();
        this.Bb = 155;
        this.c = false;
        this.Yb = 1;
        this.b = false;
        this.Ib = 256;
        this.v = false;
        this.cb = false;
        this.Nb = false;
        this.sb = 12345678;
        this.g = 180;
        this.dc = true;
        this.Vb = 12345678;
        this.Kb = false;
        this.rb = -1;
        this.Mb = 512;
        this.Fb = 95;
        this.db = false;
        this.Jb = 32;
        this.hc = 0;
        try {
            block45: {
                block44: {
                    block43: {
                        block42: {
                            block41: {
                                var4_5 = d.a(var2_2, (byte)7, var1_1);
                                var5_7 = d.a(var2_2 += 2, (byte)8, var1_1);
                                var2_2 += 2;
                                this.a(var5_7, var4_5, 115);
                                this.fc = new int[var5_7][1];
                                for (var6_8 = 0; var4_5 > var6_8; var2_2 += 2, ++var6_8) {
                                    this.a[var6_8] = w.a(var1_1, -1, var2_2);
                                    if (!var8_4) {
                                        if (!var8_4) continue;
                                    }
                                    break block41;
                                }
                                var6_8 = 0;
                            }
                            while (~var4_5 < ~var6_8) {
                                this.ob[var6_8] = w.a(var1_1, -1, var2_2);
                                var2_2 += 2;
                                ++var6_8;
                                if (!var8_4) {
                                    if (!var8_4) continue;
                                }
                                break block42;
                            }
                            var6_8 = 0;
                        }
                        while (var4_5 > var6_8) {
                            this.bc[var6_8] = w.a(var1_1, -1, var2_2);
                            var2_2 += 2;
                            ++var6_8;
                            if (!var8_4) {
                                if (!var8_4) continue;
                            }
                            break block43;
                        }
                        this.Db = var4_5;
                    }
                    var6_8 = 0;
                    while (~var5_7 < ~var6_8) {
                        this.lb[var6_8] = ib.a(255, var1_1[var2_2++]);
                        ++var6_8;
                        if (!var8_4) {
                            if (!var8_4) continue;
                        }
                        break block44;
                    }
                    var6_8 = 0;
                }
                while (~var5_7 < ~var6_8) {
                    this.V[var6_8] = w.a(var1_1, -1, var2_2);
                    v0 = this.V[var6_8];
                    v1 = 32767;
                    if (!var8_4) {
                        if (v0 == v1) {
                            this.V[var6_8] = this.Vb;
                        }
                        var2_2 += 2;
                        ++var6_8;
                        if (!var8_4) continue;
                    }
                    ** GOTO lbl77
                }
                var6_8 = 0;
                do {
                    v0 = ~var6_8;
                    v1 = ~var5_7;
lbl77:
                    // 2 sources

                    if (v0 <= v1) break;
                    this.qb[var6_8] = w.a(var1_1, -1, var2_2);
                    var2_2 += 2;
                    v2 = 32767;
                    v3 = this.qb[var6_8];
                    if (var8_4) ** GOTO lbl91
                    if (v2 == v3) {
                        this.qb[var6_8] = this.Vb;
                    }
                    ++var6_8;
                } while (!var8_4);
                var6_8 = 0;
                do {
                    block47: {
                        block46: {
                            v2 = ~var5_7;
                            v3 = ~var6_8;
lbl91:
                            // 2 sources

                            if (v2 >= v3) break;
                            v4 = var7_9 = 255 & var1_1[var2_2++];
                            if (var8_4) break block45;
                            if (v4 == 0) break block46;
                            this.Hb[var6_8] = this.Vb;
                            if (!var8_4) break block47;
                        }
                        this.Hb[var6_8] = 0;
                    }
                    ++var6_8;
                } while (!var8_4);
                v4 = 0;
            }
            var6_8 = v4;
            block25: while (true) {
                v5 = ~var6_8;
                v6 = ~var5_7;
                block26: while (v5 > v6) {
                    this.o[var6_8] = new int[this.lb[var6_8]];
                    if (var8_4 != false) return;
                    for (var7_9 = 0; this.lb[var6_8] > var7_9; ++var7_9) {
                        v5 = var4_5;
                        v6 = 256;
                        if (var8_4) continue block26;
                        if (v5 < v6) {
                            this.o[var6_8][var7_9] = ib.a(255, var1_1[var2_2++]);
                            if (!var8_4) continue;
                        }
                        this.o[var6_8][var7_9] = d.a(var2_2, (byte)102, var1_1);
                        var2_2 += 2;
                        if (!var8_4) continue;
                    }
                    ++var6_8;
                    if (!var8_4) continue block25;
                }
                break;
            }
            this.t = var5_7;
            this.Yb = 1;
            return;
        }
        catch (RuntimeException var4_6) {
            v7 = new StringBuilder().append(ca.kc[6]);
            if (var1_1 != null) {
                v8 = ca.kc[1];
                throw i.a(var4_6, v7.append(v8).append(',').append(var2_2).append(',').append(var3_3).append(')').toString());
            }
            v8 = ca.kc[3];
            throw i.a(var4_6, v7.append(v8).append(',').append(var2_2).append(',').append(var3_3).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    ca(String string) {
        boolean bl = client.vh;
        this.Bb = 155;
        this.c = false;
        this.Yb = 1;
        this.b = false;
        this.Ib = 256;
        this.v = false;
        this.cb = false;
        this.Nb = false;
        this.sb = 12345678;
        this.g = 180;
        this.dc = true;
        this.Vb = 12345678;
        this.Kb = false;
        this.rb = -1;
        this.Mb = 512;
        this.Fb = 95;
        this.db = false;
        this.Jb = 32;
        this.hc = 0;
        try {
            int n2;
            int n3;
            int n4;
            byte[] byArray;
            block28: {
                block27: {
                    int n5 = 0;
                    int n6 = 0;
                    byArray = null;
                    try {
                        DataInputStream dataInputStream;
                        block26: {
                            InputStream inputStream = nb.a(true, string);
                            dataInputStream = new DataInputStream(inputStream);
                            byArray = new byte[3];
                            n5 = 0;
                            this.hb = 0;
                            while (~n5 > -4) {
                                n5 += dataInputStream.read(byArray, n5, -n5 + 3);
                                if (!bl) {
                                    if (!bl) continue;
                                }
                                break block26;
                            }
                            n6 = this.a((byte)76, byArray);
                            n5 = 0;
                            this.hb = 0;
                        }
                        byArray = new byte[n6];
                        while (~n5 > ~n6) {
                            n5 += dataInputStream.read(byArray, n5, n6 - n5);
                            if (!bl) {
                                if (!bl) continue;
                            }
                            break block27;
                        }
                        dataInputStream.close();
                    }
                    catch (IOException iOException) {
                        this.t = 0;
                        this.Db = 0;
                        return;
                    }
                }
                int n7 = this.a((byte)76, byArray);
                n4 = this.a((byte)76, byArray);
                n3 = 0;
                this.a(n4, n7, 97);
                this.fc = new int[n4][];
                for (n2 = 0; n7 > n2; ++n2) {
                    int n8 = this.a((byte)76, byArray);
                    int n9 = this.a((byte)76, byArray);
                    int n10 = this.a((byte)76, byArray);
                    this.e(n8, n10, n9, 52);
                    if (!bl) {
                        if (!bl) continue;
                    }
                    break block28;
                }
                n2 = 0;
            }
            block14: while (~n4 < ~n2) {
                block31: {
                    block30: {
                        block29: {
                            int n11 = this.a((byte)76, byArray);
                            int n12 = this.a((byte)76, byArray);
                            int n13 = this.a((byte)76, byArray);
                            int n14 = this.a((byte)76, byArray);
                            this.Mb = this.a((byte)76, byArray);
                            this.Jb = this.a((byte)76, byArray);
                            n3 = this.a((byte)76, byArray);
                            int[] nArray = new int[n11];
                            if (bl) return;
                            for (int i2 = 0; i2 < n11; ++i2) {
                                nArray[i2] = this.a((byte)76, byArray);
                                if (bl) continue block14;
                                if (!bl) continue;
                            }
                            int[] nArray2 = new int[n14];
                            int n15 = 0;
                            while (~n15 > ~n14) {
                                nArray2[n15] = this.a((byte)76, byArray);
                                ++n15;
                                if (!bl) {
                                    if (!bl) continue;
                                }
                                break block29;
                            }
                            n15 = this.a(n11, nArray, n12, n13, false);
                            this.fc[n2] = nArray2;
                        }
                        if (0 != n3) break block30;
                        this.Hb[n15] = 0;
                        if (!bl) break block31;
                    }
                    this.Hb[n15] = this.Vb;
                }
                ++n2;
                if (!bl) continue;
            }
            this.Yb = 1;
            return;
        }
        catch (RuntimeException runtimeException) {
            String string2;
            StringBuilder stringBuilder = new StringBuilder().append(kc[6]);
            if (string != null) {
                string2 = kc[1];
                throw i.a(runtimeException, stringBuilder.append(string2).append(')').toString());
            }
            string2 = kc[3];
            throw i.a(runtimeException, stringBuilder.append(string2).append(')').toString());
        }
    }

    private ca(ca[] caArray, int n2, boolean bl, boolean bl2, boolean bl3, boolean bl4) {
        this.Bb = 155;
        this.c = false;
        this.Yb = 1;
        this.b = false;
        this.Ib = 256;
        this.v = false;
        this.cb = false;
        this.Nb = false;
        this.sb = 12345678;
        this.g = 180;
        this.dc = true;
        this.Vb = 12345678;
        this.Kb = false;
        this.rb = -1;
        this.Mb = 512;
        this.Fb = 95;
        this.db = false;
        this.Jb = 32;
        this.hc = 0;
        try {
            this.Nb = bl3;
            this.c = bl2;
            this.db = bl4;
            this.v = bl;
            this.a(0, caArray, false, n2);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(kc[6]);
            String string = caArray != null ? kc[1] : kc[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(bl).append(',').append(bl2).append(',').append(bl3).append(',').append(bl4).append(')').toString());
        }
    }

    private ca(ca[] caArray, int n2) {
        this.Bb = 155;
        this.c = false;
        this.Yb = 1;
        this.b = false;
        this.Ib = 256;
        this.v = false;
        this.cb = false;
        this.Nb = false;
        this.sb = 12345678;
        this.g = 180;
        this.dc = true;
        this.Vb = 12345678;
        this.Kb = false;
        this.rb = -1;
        this.Mb = 512;
        this.Fb = 95;
        this.db = false;
        this.Jb = 32;
        this.hc = 0;
        try {
            this.a(0, caArray, true, n2);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(kc[6]);
            String string = caArray != null ? kc[1] : kc[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(')').toString());
        }
    }

    static {
        kc = new String[]{ca.z(ca.z("x\u0016$p\r")), ca.z(ca.z("`Y$\u0013X")), ca.z(ca.z("x\u0016$m\r")), ca.z(ca.z("u\u0002fQ")), ca.z(ca.z("x\u0016$~d3")), ca.z(ca.z("u\u0016")), ca.z(ca.z("x\u0016$\u0001Lu\u001e~\u0003\r")), ca.z(ca.z("x\u0016$zd3")), ca.z(ca.z("x\u0016$ud3")), ca.z(ca.z("x\u0016${d3")), ca.z(ca.z("x\u0016$td3")), ca.z(ca.z("x\u0016$l\r")), ca.z(ca.z("x\u0016$h\r")), ca.z(ca.z("x\u0016$t\r")), ca.z(ca.z("x\u0016$z\r")), ca.z(ca.z("x\u0016$yd3")), ca.z(ca.z("x\u0016$x\r")), ca.z(ca.z("x\u0016$j\r")), ca.z(ca.z("x\u0016${\r")), ca.z(ca.z("x\u0016$~\r")), ca.z(ca.z("x\u0016$wd3")), ca.z(ca.z("x\u0016$q\r")), ca.z(ca.z("x\u0016$\u007fd3")), ca.z(ca.z("x\u0016$w\r")), ca.z(ca.z("x\u0016$\u007f\r")), ca.z(ca.z("x\u0016$k\r")), ca.z(ca.z("x\u0016$|d3")), ca.z(ca.z("x\u0016$o\r")), ca.z(ca.z("x\u0016$s\r")), ca.z(ca.z("x\u0016$i\r")), ca.z(ca.z("x\u0016$u\r")), ca.z(ca.z("x\u0016$r\r")), ca.z(ca.z("x\u0016$n\r")), ca.z(ca.z("x\u0016$y\r")), ca.z(ca.z("x\u0016$|\r")), ca.z(ca.z("x\u0016$vd3")), ca.z(ca.z("x\u0016$v\r")), ca.z(ca.z("x\u0016$xd3"))};
        tb = new byte[50][];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x25);
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
                        n4 = 27;
                        break;
                    }
                    case 1: {
                        n4 = 119;
                        break;
                    }
                    case 2: {
                        n4 = 10;
                        break;
                    }
                    case 3: {
                        n4 = 61;
                        break;
                    }
                    default: {
                        n4 = 37;
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

