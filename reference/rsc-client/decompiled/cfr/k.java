/*
 * Decompiled with CFR 0.152.
 */
import java.io.IOException;

final class k {
    private boolean H;
    byte[] Q;
    private lb c;
    private boolean nb;
    static int lb;
    static int z;
    static int ib;
    private int[][] ab;
    private int[] w;
    private byte[][] L;
    static int cb;
    boolean Z;
    static int t;
    private int[][] s;
    private byte[][] R;
    static int u;
    int[][] bb;
    static int N;
    static int j;
    private byte[][] f;
    private byte[][] P;
    static int C;
    private byte[][] mb;
    private ca kb;
    static int h;
    static int n;
    private ca[] F;
    static int jb;
    static int a;
    static int l;
    private ua U;
    static int Y;
    static int d;
    static int M;
    static int r;
    static int p;
    byte[] m;
    private int[][] B;
    static int J;
    int x;
    static int K;
    static int T;
    ca[][] g;
    int[] E;
    static String[] G;
    byte[] gb;
    static int k;
    private byte[][] A;
    static int o;
    static int fb;
    static int S;
    static int X;
    static long e;
    static int b;
    byte[] I;
    static int O;
    static int W;
    static int D;
    static int V;
    static int y;
    static int i;
    private byte[][] eb;
    ca[][] db;
    static int hb;
    static int v;
    int[] q;
    private static final String[] ob;

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int n3, int n4, int n5, int n6, int n7) {
        boolean bl = client.vh;
        try {
            ++X;
            if (n5 != 2) {
                this.Q = null;
            }
            ca ca2 = this.F[n6 - -(n4 * 8)];
            int i2 = 0;
            do {
                if (ca2.Db <= i2) return;
                if (bl) return;
                if (~(128 * n7) == ~ca2.a[i2] && ~ca2.bc[i2] == ~(n2 * 128)) {
                    ca2.a(i2, n3, (byte)-61);
                    return;
                }
                ++i2;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[8] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2) {
        boolean bl = client.vh;
        try {
            ++O;
            int n3 = n2;
            do {
                int n4 = 96;
                int n5 = n3;
                block11: while (true) {
                    if (n4 <= n5) return;
                    if (bl) return;
                    for (int i2 = 0; i2 < 96; ++i2) {
                        n4 = ~this.b(0, n3, n2 ^ 4, i2);
                        n5 = -251;
                        if (bl) continue block11;
                        if (n4 != n5) continue;
                        if (47 != n3 || 250 == this.b(0, n3 - -1, n2 + 4, i2) || -3 == ~this.b(0, 1 + n3, 4, i2)) {
                            if (~i2 == -48 && 250 != this.b(0, n3, 4, i2 - -1) && ~this.b(0, n3, n2 ^ 4, 1 + i2) != -3) {
                                this.e(9, i2, 107, n3);
                                if (!bl) continue;
                            }
                            this.e(2, i2, 110, n3);
                            if (!bl) continue;
                        }
                        this.e(9, i2, n2 + 111, n3);
                        if (!bl) continue;
                    }
                    break;
                }
                ++n3;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[32] + n2 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, int n3, int n4, int n5) {
        boolean bl = client.vh;
        try {
            int n6;
            int n7;
            int n8;
            block32: {
                block31: {
                    ++h;
                    if (n5 != 4081) {
                        this.a(-98, 25, -8, -1, (byte)-83, 45);
                    }
                    if (n3 < 0) return;
                    if (n4 < 0) return;
                    if (95 <= n3) return;
                    if (95 <= n4) return;
                    if (1 != mb.a[n2]) {
                        if (~mb.a[n2] != -3) return;
                    }
                    if ((n8 = this.b(n3, n4, -79)) != 0 && ~n8 != -5) break block31;
                    n7 = ub.g[n2];
                    n6 = f.f[n2];
                    if (!bl) break block32;
                }
                n7 = f.f[n2];
                n6 = ub.g[n2];
            }
            int n9 = n3;
            block19: while (true) {
                int n10 = ~n9;
                int n11 = ~(n3 - -n6);
                block20: while (n10 > n11) {
                    if (bl) return;
                    for (int i2 = n4; n7 + n4 > i2; ++i2) {
                        n10 = -2;
                        n11 = ~mb.a[n2];
                        if (bl) continue block20;
                        if (n10 != n11) {
                            if (n8 == 0) {
                                this.bb[n9][i2] = ib.a(this.bb[n9][i2], 65533);
                                if (n9 <= 0) continue;
                                this.c(i2, n5 + 61454, -1 + n9, 8);
                                if (!bl) continue;
                            }
                            if (~n8 != -3) {
                                if (~n8 != -5) {
                                    if (-7 != ~n8) continue;
                                    this.bb[n9][i2] = ib.a(this.bb[n9][i2], 65534);
                                    if (~i2 >= -1) continue;
                                    this.c(-1 + i2, n5 ^ 0xF00E, n9, 4);
                                    if (!bl) continue;
                                }
                                this.bb[n9][i2] = ib.a(this.bb[n9][i2], 65527);
                                if (n9 >= 95) continue;
                                this.c(i2, n5 + 61454, 1 + n9, 2);
                                if (!bl) continue;
                            }
                            this.bb[n9][i2] = ib.a(this.bb[n9][i2], 65531);
                            if (~i2 <= -96) continue;
                            this.c(i2 - -1, n5 + 61454, n9, 1);
                            if (!bl) continue;
                        }
                        this.bb[n9][i2] = ib.a(this.bb[n9][i2], 65471);
                        if (!bl) continue;
                    }
                    ++n9;
                    if (!bl) continue block19;
                }
                break;
            }
            this.c(n6, n7, -82, n3, n4);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[43] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void e(int n2, int n3, int n4, int n5) {
        boolean bl = client.vh;
        try {
            block15: {
                int n6;
                block13: {
                    block14: {
                        ++v;
                        if (~n5 > -1) return;
                        if (-97 >= ~n5) return;
                        if (~n3 > -1) return;
                        if (96 <= n3) {
                            return;
                        }
                        n6 = 0;
                        if (n5 >= 48 && n3 < 48) break block13;
                        if (~n5 <= -49 || ~n3 > -49) break block14;
                        n6 = 2;
                        n3 -= 48;
                        if (!bl) break block15;
                    }
                    if (~n5 > -49 || n3 < 48) break block15;
                    n5 -= 48;
                    n3 -= 48;
                    n6 = 3;
                    if (!bl) break block15;
                }
                n6 = 1;
                n5 -= 48;
            }
            int n7 = -76 % ((n4 - 53) / 53);
            this.R[n6][n3 + 48 * n5] = (byte)n2;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[37] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, int n3, boolean bl, int n4) {
        boolean bl2 = client.vh;
        try {
            int n5;
            int n6;
            int n7;
            block31: {
                block30: {
                    if (bl) {
                        return;
                    }
                    ++u;
                    if (~n2 > -1) return;
                    if (-1 < ~n4) return;
                    if (n2 >= 95) return;
                    if (~n4 <= -96) {
                        return;
                    }
                    if (~mb.a[n3] != -2) {
                        if (mb.a[n3] != 2) return;
                    }
                    if (0 != (n7 = this.b(n2, n4, -107)) && ~n7 != -5) break block30;
                    n6 = f.f[n3];
                    n5 = ub.g[n3];
                    if (!bl2) break block31;
                }
                n5 = f.f[n3];
                n6 = ub.g[n3];
            }
            int n8 = n2;
            block17: while (true) {
                int n9 = n8;
                int n10 = n6 + n2;
                block18: while (n9 < n10) {
                    if (bl2) return;
                    for (int i2 = n4; n4 + n5 > i2; ++i2) {
                        n9 = -2;
                        n10 = ~mb.a[n3];
                        if (bl2) continue block18;
                        if (n9 == n10) {
                            this.bb[n8][i2] = d.a(this.bb[n8][i2], 64);
                            if (!bl2) continue;
                        }
                        if (n7 != 0) {
                            if (-3 == ~n7) {
                                this.bb[n8][i2] = d.a(this.bb[n8][i2], 4);
                                if (-96 >= ~i2) continue;
                                this.a(1, 1 + i2, (byte)-112, n8);
                                if (!bl2) continue;
                            }
                            if (4 != n7) {
                                if (6 != n7) continue;
                                this.bb[n8][i2] = d.a(this.bb[n8][i2], 1);
                                if (~i2 >= -1) continue;
                                this.a(4, -1 + i2, (byte)-112, n8);
                                if (!bl2) continue;
                            }
                            this.bb[n8][i2] = d.a(this.bb[n8][i2], 8);
                            if (~n8 <= -96) continue;
                            this.a(2, i2, (byte)-56, n8 + 1);
                            if (!bl2) continue;
                        }
                        this.bb[n8][i2] = d.a(this.bb[n8][i2], 2);
                        if (0 >= n8) continue;
                        this.a(8, i2, (byte)-109, n8 - 1);
                        if (!bl2) continue;
                    }
                    ++n8;
                    if (!bl2) continue block17;
                }
                break;
            }
            this.c(n6, n5, 94, n2, n4);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[0] + n2 + ',' + n3 + ',' + bl + ',' + n4 + ')');
        }
    }

    static final byte[] a(int n2, boolean bl, byte[] byArray) {
        try {
            int n3;
            int n4;
            ++S;
            if (n2 != 128) {
                o = 104;
            }
            if (~(n4 = ((byArray[1] & 0xFF) << 1420085512) + ((0xFF0000 & byArray[0] << -421308816) - -(0xFF & byArray[2]))) == ~(n3 = ((0xFF & byArray[4]) << -1631867672) + (0xFF0000 & byArray[3] << 440042672) + (0xFF & byArray[5]))) {
                byte[] byArray2 = new byte[byArray.length - 6];
                ab.a(byArray, 6, byArray2, 0, byArray2.length);
                return byArray2;
            }
            if (bl) {
                da.a(ob[18], 0, 0);
            }
            byte[] byArray3 = new byte[n4];
            ea.a(byArray3, n4, byArray, n3, 6);
            return byArray3;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(ob[19]).append(n2).append(',').append(bl).append(',');
            String string = byArray != null ? ob[1] : ob[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void b(int n2) {
        boolean bl = client.vh;
        try {
            if (n2 != -10185) {
                return;
            }
            if (this.nb) {
                this.c.a(false);
            }
            ++N;
            int n3 = 0;
            while (~n3 > -65) {
                block13: {
                    int n4;
                    block12: {
                        this.F[n3] = null;
                        if (bl) return;
                        n4 = 0;
                        while (~n4 > -5) {
                            this.g[n4][n3] = null;
                            ++n4;
                            if (!bl) {
                                if (!bl) continue;
                            }
                            break block12;
                        }
                        n4 = 0;
                    }
                    while (-5 < ~n4) {
                        this.db[n4][n3] = null;
                        ++n4;
                        if (!bl) {
                            if (!bl) continue;
                        }
                        break block13;
                    }
                    ++n3;
                }
                if (!bl) continue;
            }
            System.gc();
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[12] + n2 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int n3, boolean bl, int n4, int n5) {
        boolean bl2 = client.vh;
        try {
            int n6;
            int n7;
            block292: {
                block291: {
                    int n8;
                    int n9;
                    int n10;
                    int n11;
                    int n12;
                    int n13;
                    int n14;
                    int n15;
                    int n16;
                    int n17;
                    int n18;
                    int n19;
                    int n20;
                    block290: {
                        int n21;
                        block289: {
                            block288: {
                                int n22;
                                block287: {
                                    int n23;
                                    int n24;
                                    block279: {
                                        int n25;
                                        block285: {
                                            block293: {
                                                block284: {
                                                    block283: {
                                                        int n26;
                                                        int n27;
                                                        ca ca2;
                                                        block282: {
                                                            int n28;
                                                            block281: {
                                                                ++i;
                                                                int n29 = (24 + n2) / 48;
                                                                int n30 = (24 + n5) / 48;
                                                                this.b(n4, 0, n29 - 1, 0, -1 + n30);
                                                                this.b(n4, 1, n29, 0, n30 - 1);
                                                                if (n3 < 66) {
                                                                    return;
                                                                }
                                                                this.b(n4, 2, -1 + n29, 0, n30);
                                                                this.b(n4, 3, n29, 0, n30);
                                                                this.a(0);
                                                                if (this.kb == null) {
                                                                    this.kb = new ca(18688, 18688, true, true, false, false, true);
                                                                }
                                                                if (!bl) break block293;
                                                                this.U.a(true);
                                                                n24 = 0;
                                                                while (96 > n24) {
                                                                    block280: {
                                                                        n23 = 0;
                                                                        if (bl2) break block279;
                                                                        for (n7 = v588636; n7 < 96; ++n7) {
                                                                            this.bb[n24][n7] = 0;
                                                                            if (!bl2) {
                                                                                if (!bl2) continue;
                                                                            }
                                                                            break block280;
                                                                        }
                                                                        ++n24;
                                                                    }
                                                                    if (!bl2) continue;
                                                                }
                                                                ca2 = this.kb;
                                                                ca2.c(1);
                                                                n7 = 0;
                                                                block144: while (true) {
                                                                    int n32 = n7;
                                                                    n32 = 96;
                                                                    block145: while (n31 < n32) {
                                                                        n28 = 0;
                                                                        if (bl2) break block281;
                                                                        for (n6 = v588038; 96 > n6; ++n6) {
                                                                            n20 = -this.g(2, n6, n7);
                                                                            int n32 = ~this.b(n4, n7, 4, n6);
                                                                            n32 = -1;
                                                                            if (bl2) continue block145;
                                                                            if (n31 < n32 && 4 == da.N[-1 + this.b(n4, n7, 4, n6)]) {
                                                                                n20 = 0;
                                                                            }
                                                                            if (this.b(n4, -1 + n7, 4, n6) > 0 && -5 == ~da.N[this.b(n4, n7 - 1, 4, n6) + -1]) {
                                                                                n20 = 0;
                                                                            }
                                                                            if (this.b(n4, n7, 4, n6 + -1) > 0 && -5 == ~da.N[this.b(n4, n7, 4, n6 + -1) + -1]) {
                                                                                n20 = 0;
                                                                            }
                                                                            if (this.b(n4, n7 - 1, 4, n6 - 1) > 0 && 4 == da.N[this.b(n4, -1 + n7, 4, n6 - 1) - 1]) {
                                                                                n20 = 0;
                                                                            }
                                                                            n19 = ca2.e(n7 * 128, 128 * n6, n20, 107);
                                                                            n18 = (int)(Math.random() * 10.0) - 5;
                                                                            ca2.a(n19, n18, (byte)-61);
                                                                            if (!bl2) continue;
                                                                        }
                                                                        ++n7;
                                                                        if (!bl2) continue block144;
                                                                    }
                                                                    break;
                                                                }
                                                                n28 = 0;
                                                            }
                                                            n7 = n28;
                                                            block147: while (true) {
                                                                int n34 = ~n7;
                                                                n34 = -96;
                                                                block148: while (n33 > n34) {
                                                                    n27 = 0;
                                                                    if (bl2) break block282;
                                                                    for (n6 = v588396; n6 < 95; ++n6) {
                                                                        block303: {
                                                                            block305: {
                                                                                int[] nArray;
                                                                                int[] nArray2;
                                                                                block304: {
                                                                                    int n35;
                                                                                    block294: {
                                                                                        block296: {
                                                                                            block302: {
                                                                                                int n36;
                                                                                                block301: {
                                                                                                    block300: {
                                                                                                        block295: {
                                                                                                            block299: {
                                                                                                                block298: {
                                                                                                                    block297: {
                                                                                                                        n20 = this.a((byte)104, n7, n6);
                                                                                                                        n18 = n19 = this.w[n20];
                                                                                                                        n17 = n19;
                                                                                                                        int n34 = -2;
                                                                                                                        n34 = ~n4;
                                                                                                                        if (bl2) continue block148;
                                                                                                                        if (n33 == n34 || ~n4 == -3) {
                                                                                                                            n18 = 12345678;
                                                                                                                            n19 = 12345678;
                                                                                                                            n17 = 12345678;
                                                                                                                        }
                                                                                                                        n16 = 0;
                                                                                                                        if (this.b(n4, n7, 4, n6) <= 0) break block294;
                                                                                                                        n35 = this.b(n4, n7, 4, n6);
                                                                                                                        n20 = da.N[n35 + -1];
                                                                                                                        n36 = this.d(n4, n6, 15282, n7);
                                                                                                                        n19 = n18 = qa.K[-1 + n35];
                                                                                                                        if (~n20 == -5) {
                                                                                                                            n19 = 1;
                                                                                                                            n18 = 1;
                                                                                                                            if (~n35 == -13) {
                                                                                                                                n19 = 31;
                                                                                                                                n18 = 31;
                                                                                                                            }
                                                                                                                        }
                                                                                                                        if (~n20 != -6) break block295;
                                                                                                                        if (0 >= this.c(n7, n6, -49) || -24001 >= ~this.c(n7, n6, -49)) break block296;
                                                                                                                        if (~this.d(-8509, n7 + -1, n17, n4, n6) == -12345679 || -12345679 == ~this.d(-8509, n7, n17, n4, -1 + n6)) break block297;
                                                                                                                        n16 = 0;
                                                                                                                        n19 = this.d(-8509, n7 + -1, n17, n4, n6);
                                                                                                                        if (!bl2) break block296;
                                                                                                                    }
                                                                                                                    if (this.d(-8509, 1 + n7, n17, n4, n6) == 12345678 || 12345678 == this.d(-8509, n7, n17, n4, 1 + n6)) break block298;
                                                                                                                    n18 = this.d(-8509, n7 - -1, n17, n4, n6);
                                                                                                                    n16 = 0;
                                                                                                                    if (!bl2) break block296;
                                                                                                                }
                                                                                                                if (~this.d(-8509, 1 + n7, n17, n4, n6) != -12345679 && ~this.d(-8509, n7, n17, n4, -1 + n6) != -12345679) break block299;
                                                                                                                if (-12345679 == ~this.d(-8509, n7 - 1, n17, n4, n6) || this.d(-8509, n7, n17, n4, n6 - -1) == 12345678) break block296;
                                                                                                                n16 = 1;
                                                                                                                n19 = this.d(-8509, n7 - 1, n17, n4, n6);
                                                                                                                if (!bl2) break block296;
                                                                                                            }
                                                                                                            n18 = this.d(-8509, n7 + 1, n17, n4, n6);
                                                                                                            n16 = 1;
                                                                                                            if (!bl2) break block296;
                                                                                                        }
                                                                                                        if (2 == n20 && (0 >= this.c(n7, n6, -49) || -24001 >= ~this.c(n7, n6, -49))) break block296;
                                                                                                        if (n36 == this.d(n4, n6, 15282, -1 + n7) || this.d(n4, -1 + n6, 15282, n7) == n36) break block300;
                                                                                                        n19 = n17;
                                                                                                        n16 = 0;
                                                                                                        if (!bl2) break block296;
                                                                                                    }
                                                                                                    if (n36 == this.d(n4, n6, 15282, n7 + 1) || ~n36 == ~this.d(n4, n6 - -1, 15282, n7)) break block301;
                                                                                                    n16 = 0;
                                                                                                    n18 = n17;
                                                                                                    if (!bl2) break block296;
                                                                                                }
                                                                                                if (n36 != this.d(n4, n6, 15282, 1 + n7) && ~n36 != ~this.d(n4, -1 + n6, 15282, n7)) break block302;
                                                                                                if (n36 == this.d(n4, n6, 15282, -1 + n7) || n36 == this.d(n4, 1 + n6, 15282, n7)) break block296;
                                                                                                n19 = n17;
                                                                                                n16 = 1;
                                                                                                if (!bl2) break block296;
                                                                                            }
                                                                                            n18 = n17;
                                                                                            n16 = 1;
                                                                                        }
                                                                                        if (~ac.l[n35 - 1] != -1) {
                                                                                            this.bb[n7][n6] = d.a(this.bb[n7][n6], 64);
                                                                                        }
                                                                                        if (-3 == ~da.N[n35 + -1]) {
                                                                                            this.bb[n7][n6] = d.a(this.bb[n7][n6], 128);
                                                                                        }
                                                                                    }
                                                                                    this.a(n16, (byte)-122, n18, n7, n6, n19);
                                                                                    n35 = this.g(2, 1 + n6, n7 + 1) + (-this.g(2, n6, n7 - -1) - -this.g(2, n6 + 1, n7)) - this.g(2, n6, n7);
                                                                                    if (~n18 == ~n19 && -1 == ~n35) break block303;
                                                                                    nArray2 = new int[3];
                                                                                    nArray = new int[3];
                                                                                    if (-1 != ~n16) break block304;
                                                                                    if (n19 != 12345678) {
                                                                                        nArray2[1] = n7 * 96 + n6;
                                                                                        nArray2[0] = 96 + n6 + 96 * n7;
                                                                                        nArray2[2] = 1 + (n6 + 96 * n7);
                                                                                        n15 = ca2.a(3, nArray2, 12345678, n19, false);
                                                                                        this.q[n15] = n7;
                                                                                        this.E[n15] = n6;
                                                                                        ca2.E[n15] = n15 + 200000;
                                                                                    }
                                                                                    if (12345678 == n18) break block305;
                                                                                    nArray[2] = n6 + (96 * n7 - -96);
                                                                                    nArray[1] = 97 + (96 * n7 + n6);
                                                                                    nArray[0] = 1 + n7 * 96 + n6;
                                                                                    n15 = ca2.a(3, nArray, 12345678, n18, false);
                                                                                    this.q[n15] = n7;
                                                                                    this.E[n15] = n6;
                                                                                    ca2.E[n15] = n15 + 200000;
                                                                                    if (!bl2) break block305;
                                                                                }
                                                                                if (12345678 != n19) {
                                                                                    nArray2[2] = n6 - -(96 * n7);
                                                                                    nArray2[1] = 96 + (96 * n7 + (n6 - -1));
                                                                                    nArray2[0] = 1 + (96 * n7 + n6);
                                                                                    n15 = ca2.a(3, nArray2, 12345678, n19, false);
                                                                                    this.q[n15] = n7;
                                                                                    this.E[n15] = n6;
                                                                                    ca2.E[n15] = 200000 - -n15;
                                                                                }
                                                                                if (~n18 != -12345679) {
                                                                                    nArray[1] = n6 - -(96 * n7);
                                                                                    nArray[2] = n6 - (-(n7 * 96) - 97);
                                                                                    nArray[0] = 96 * n7 + n6 - -96;
                                                                                    n15 = ca2.a(3, nArray, 12345678, n18, false);
                                                                                    this.q[n15] = n7;
                                                                                    this.E[n15] = n6;
                                                                                    ca2.E[n15] = n15 + 200000;
                                                                                }
                                                                            }
                                                                            if (!bl2) continue;
                                                                        }
                                                                        if (~n19 == -12345679) continue;
                                                                        int[] nArray = new int[4];
                                                                        nArray[3] = n6 - (-(n7 * 96) + -96) - -1;
                                                                        nArray[2] = 1 + n7 * 96 + n6;
                                                                        nArray[0] = n6 - (-(n7 * 96) + -96);
                                                                        nArray[1] = n6 - -(n7 * 96);
                                                                        int n37 = ca2.a(4, nArray, 12345678, n19, false);
                                                                        this.q[n37] = n7;
                                                                        this.E[n37] = n6;
                                                                        ca2.E[n37] = n37 + 200000;
                                                                        if (!bl2) continue;
                                                                    }
                                                                    ++n7;
                                                                    if (!bl2) continue block147;
                                                                }
                                                                break;
                                                            }
                                                            n27 = 1;
                                                        }
                                                        n7 = n27;
                                                        block150: while (true) {
                                                            int n38 = 95;
                                                            block151: while (n38 > n7) {
                                                                n26 = 1;
                                                                if (bl2) break block283;
                                                                for (n6 = v588594; n6 < 95; ++n6) {
                                                                    n38 = this.b(n4, n7, 4, n6);
                                                                    if (bl2) continue block151;
                                                                    if (n38 > 0 && da.N[this.b(n4, n7, 4, n6) + -1] == 4) {
                                                                        n20 = qa.K[-1 + this.b(n4, n7, 4, n6)];
                                                                        n19 = ca2.e(n7 * 128, 128 * n6, -this.g(2, n6, n7), 13);
                                                                        n18 = ca2.e(128 * (n7 + 1), n6 * 128, -this.g(2, n6, 1 + n7), 107);
                                                                        n17 = ca2.e((1 + n7) * 128, (n6 - -1) * 128, -this.g(2, n6 + 1, n7 - -1), -116);
                                                                        n16 = ca2.e(n7 * 128, 128 + 128 * n6, -this.g(2, 1 + n6, n7), -124);
                                                                        int[] nArray = new int[]{n19, n18, n17, n16};
                                                                        int n39 = ca2.a(4, nArray, n20, 12345678, false);
                                                                        this.q[n39] = n7;
                                                                        this.E[n39] = n6;
                                                                        ca2.E[n39] = n39 + 200000;
                                                                        this.a(0, (byte)-121, n20, n7, n6, n20);
                                                                        if (!bl2) continue;
                                                                    }
                                                                    if (this.b(n4, n7, 4, n6) != 0 && da.N[this.b(n4, n7, 4, n6) - 1] == 3) continue;
                                                                    if (0 < this.b(n4, n7, 4, n6 - -1) && da.N[-1 + this.b(n4, n7, 4, 1 + n6)] == 4) {
                                                                        n20 = qa.K[-1 + this.b(n4, n7, 4, n6 + 1)];
                                                                        n19 = ca2.e(128 * n7, 128 * n6, -this.g(2, n6, n7), -124);
                                                                        n18 = ca2.e(128 * (n7 + 1), 128 * n6, -this.g(2, n6, 1 + n7), -118);
                                                                        n17 = ca2.e(128 + n7 * 128, 128 * (n6 - -1), -this.g(2, n6 + 1, 1 + n7), -124);
                                                                        n16 = ca2.e(128 * n7, 128 * n6 + 128, -this.g(2, 1 + n6, n7), -116);
                                                                        int[] nArray = new int[]{n19, n18, n17, n16};
                                                                        int n40 = ca2.a(4, nArray, n20, 12345678, false);
                                                                        this.q[n40] = n7;
                                                                        this.E[n40] = n6;
                                                                        ca2.E[n40] = n40 + 200000;
                                                                        this.a(0, (byte)34, n20, n7, n6, n20);
                                                                    }
                                                                    if (-1 > ~this.b(n4, n7, 4, n6 - 1) && ~da.N[this.b(n4, n7, 4, n6 - 1) - 1] == -5) {
                                                                        n20 = qa.K[this.b(n4, n7, 4, -1 + n6) + -1];
                                                                        n19 = ca2.e(n7 * 128, n6 * 128, -this.g(2, n6, n7), -122);
                                                                        n18 = ca2.e(128 * (1 + n7), n6 * 128, -this.g(2, n6, n7 + 1), 123);
                                                                        n17 = ca2.e(128 * (1 + n7), 128 * (n6 - -1), -this.g(2, n6 - -1, n7 + 1), -104);
                                                                        n16 = ca2.e(128 * n7, 128 + 128 * n6, -this.g(2, n6 - -1, n7), -127);
                                                                        int[] nArray = new int[]{n19, n18, n17, n16};
                                                                        int n41 = ca2.a(4, nArray, n20, 12345678, false);
                                                                        this.q[n41] = n7;
                                                                        this.E[n41] = n6;
                                                                        ca2.E[n41] = 200000 + n41;
                                                                        this.a(0, (byte)17, n20, n7, n6, n20);
                                                                    }
                                                                    if (this.b(n4, n7 - -1, 4, n6) > 0 && da.N[-1 + this.b(n4, n7 + 1, 4, n6)] == 4) {
                                                                        n20 = qa.K[this.b(n4, 1 + n7, 4, n6) + -1];
                                                                        n19 = ca2.e(128 * n7, n6 * 128, -this.g(2, n6, n7), -113);
                                                                        n18 = ca2.e(128 + n7 * 128, n6 * 128, -this.g(2, n6, 1 + n7), 89);
                                                                        n17 = ca2.e(128 * (1 + n7), 128 * n6 - -128, -this.g(2, 1 + n6, 1 + n7), 124);
                                                                        n16 = ca2.e(n7 * 128, (1 + n6) * 128, -this.g(2, n6 - -1, n7), -112);
                                                                        int[] nArray = new int[]{n19, n18, n17, n16};
                                                                        int n42 = ca2.a(4, nArray, n20, 12345678, false);
                                                                        this.q[n42] = n7;
                                                                        this.E[n42] = n6;
                                                                        ca2.E[n42] = n42 + 200000;
                                                                        this.a(0, (byte)-124, n20, n7, n6, n20);
                                                                    }
                                                                    if (-1 <= ~this.b(n4, -1 + n7, 4, n6) || da.N[this.b(n4, -1 + n7, 4, n6) - 1] != 4) continue;
                                                                    n20 = qa.K[this.b(n4, n7 + -1, 4, n6) - 1];
                                                                    n19 = ca2.e(n7 * 128, 128 * n6, -this.g(2, n6, n7), -123);
                                                                    n18 = ca2.e((n7 - -1) * 128, 128 * n6, -this.g(2, n6, 1 + n7), 106);
                                                                    n17 = ca2.e(128 + 128 * n7, n6 * 128 + 128, -this.g(2, 1 + n6, 1 + n7), 56);
                                                                    n16 = ca2.e(n7 * 128, 128 * (1 + n6), -this.g(2, n6 - -1, n7), 119);
                                                                    int[] nArray = new int[]{n19, n18, n17, n16};
                                                                    int n43 = ca2.a(4, nArray, n20, 12345678, false);
                                                                    this.q[n43] = n7;
                                                                    this.E[n43] = n6;
                                                                    ca2.E[n43] = n43 + 200000;
                                                                    this.a(0, (byte)-127, n20, n7, n6, n20);
                                                                    if (!bl2) continue;
                                                                }
                                                                ++n7;
                                                                if (!bl2) continue block150;
                                                            }
                                                            break;
                                                        }
                                                        ca2.a(-50, 40, -10, -50, true, 48, 105);
                                                        this.F = this.kb.a(0, 8, 1536, 112, 64, 233, 1536, false, 0);
                                                        n26 = n7 = 0;
                                                    }
                                                    while (~n7 > -65) {
                                                        this.c.a(this.F[n7], (byte)118);
                                                        ++n7;
                                                        if (!bl2) {
                                                            if (!bl2) continue;
                                                        }
                                                        break block284;
                                                    }
                                                    n7 = 0;
                                                }
                                                while (~n7 > -97) {
                                                    block286: {
                                                        n25 = 0;
                                                        if (bl2) break block285;
                                                        n6 = n25;
                                                        while (~n6 > -97) {
                                                            this.ab[n7][n6] = this.g(2, n6, n7);
                                                            ++n6;
                                                            if (!bl2) {
                                                                if (!bl2) continue;
                                                            }
                                                            break block286;
                                                        }
                                                        ++n7;
                                                    }
                                                    if (!bl2) continue;
                                                }
                                            }
                                            this.kb.c(1);
                                            n25 = 0x606060;
                                        }
                                        n24 = n25;
                                        n23 = 0;
                                    }
                                    n7 = n23;
                                    block156: while (true) {
                                        int n45 = n7;
                                        n45 = 95;
                                        block157: while (n44 < n45) {
                                            n22 = 0;
                                            if (bl2) break block287;
                                            n6 = n22;
                                            while (~n6 > -96) {
                                                n20 = this.a(n7, (byte)-124, n6);
                                                int n45 = 0;
                                                n45 = n20;
                                                if (bl2) continue block157;
                                                if (n44 < n45 && (lb.Tb[n20 + -1] == 0 || this.H)) {
                                                    this.a(-1 + n20, this.kb, 1 + n7, n6, n7, -14584, n6);
                                                    if (bl && u.a[-1 + n20] != 0) {
                                                        this.bb[n7][n6] = d.a(this.bb[n7][n6], 1);
                                                        if (n6 > 0) {
                                                            this.a(4, -1 + n6, (byte)-125, n7);
                                                        }
                                                    }
                                                    if (bl) {
                                                        this.U.b(3, n24, n7 * 3, n6 * 3, (byte)-109);
                                                    }
                                                }
                                                if (0 < (n20 = this.e(95, n7, n6)) && (~lb.Tb[-1 + n20] == -1 || this.H)) {
                                                    this.a(n20 + -1, this.kb, n7, n6, n7, -14584, 1 + n6);
                                                    if (bl && u.a[n20 + -1] != 0) {
                                                        this.bb[n7][n6] = d.a(this.bb[n7][n6], 2);
                                                        if (-1 > ~n7) {
                                                            this.a(8, n6, (byte)-72, n7 + -1);
                                                        }
                                                    }
                                                    if (bl) {
                                                        this.U.b(n7 * 3, 3 * n6, n24, 3, 0);
                                                    }
                                                }
                                                if (0 < (n20 = this.c(n7, n6, -49)) && ~n20 > -12001 && (lb.Tb[n20 + -1] == 0 || this.H)) {
                                                    this.a(-1 + n20, this.kb, n7 - -1, n6, n7, -14584, 1 + n6);
                                                    if (bl && 0 != u.a[n20 - 1]) {
                                                        this.bb[n7][n6] = d.a(this.bb[n7][n6], 32);
                                                    }
                                                    if (bl) {
                                                        this.U.a(3 * n6, n7 * 3, 82, n24);
                                                        this.U.a(1 + 3 * n6, 1 + n7 * 3, 69, n24);
                                                        this.U.a(2 + 3 * n6, n7 * 3 - -2, 65, n24);
                                                    }
                                                }
                                                if (12000 < n20 && n20 < 24000 && (~lb.Tb[-12001 + n20] == -1 || this.H)) {
                                                    this.a(-12001 + n20, this.kb, n7, n6, n7 - -1, -14584, 1 + n6);
                                                    if (bl && ~u.a[n20 + -12001] != -1) {
                                                        this.bb[n7][n6] = d.a(this.bb[n7][n6], 16);
                                                    }
                                                    if (bl) {
                                                        this.U.a(3 * n6, 2 + 3 * n7, 116, n24);
                                                        this.U.a(n6 * 3 + 1, n7 * 3 + 1, 99, n24);
                                                        this.U.a(2 + 3 * n6, n7 * 3, 90, n24);
                                                    }
                                                }
                                                ++n6;
                                                if (!bl2) continue;
                                            }
                                            ++n7;
                                            if (!bl2) continue block156;
                                        }
                                        break;
                                    }
                                    n22 = bl;
                                }
                                if (n22 != 0) {
                                    this.U.b(285, 0, 0, -27966, -1 + this.x, 285);
                                }
                                this.kb.a(-50, 60, -10, -50, false, 24, 122);
                                this.g[n4] = this.kb.a(0, 8, 1536, -120, 64, 338, 1536, true, 0);
                                for (n7 = 0; n7 < 64; ++n7) {
                                    this.c.a(this.g[n4][n7], (byte)118);
                                    if (!bl2) {
                                        if (!bl2) continue;
                                    }
                                    break block288;
                                }
                                n7 = 0;
                            }
                            block160: while (true) {
                                int n46 = 95;
                                block161: while (n46 > n7) {
                                    n21 = 0;
                                    if (bl2) break block289;
                                    for (n6 = v588922; n6 < 95; ++n6) {
                                        n46 = n20 = this.a(n7, (byte)-111, n6);
                                        if (bl2) continue block161;
                                        if (n46 > 0) {
                                            this.a(-1 + n20, n7 + 1, n6, n6, (byte)-50, n7);
                                        }
                                        if (0 < (n20 = this.e(61, n7, n6))) {
                                            this.a(n20 + -1, n7, n6, n6 - -1, (byte)-65, n7);
                                        }
                                        if ((n20 = this.c(n7, n6, -49)) > 0 && n20 < 12000) {
                                            this.a(-1 + n20, n7 - -1, n6, n6 + 1, (byte)-118, n7);
                                        }
                                        if (-12001 <= ~n20 || n20 >= 24000) continue;
                                        this.a(n20 + -12001, n7, n6, n6 + 1, (byte)82, n7 - -1);
                                        if (!bl2) continue;
                                    }
                                    ++n7;
                                    if (!bl2) continue block160;
                                }
                                break;
                            }
                            n21 = 1;
                        }
                        n7 = n21;
                        block163: while (true) {
                            int n48 = n7;
                            n48 = 95;
                            block164: while (n47 < n48) {
                                n14 = 1;
                                if (bl2) break block290;
                                n6 = n14;
                                while (~n6 > -96) {
                                    block306: {
                                        int n49;
                                        block313: {
                                            block312: {
                                                int n50;
                                                int n51;
                                                block311: {
                                                    block310: {
                                                        block309: {
                                                            block308: {
                                                                block307: {
                                                                    n20 = this.d(n6, n7, 115);
                                                                    int n48 = ~n20;
                                                                    n48 = -1;
                                                                    if (bl2) continue block164;
                                                                    if (n47 >= n48) break block306;
                                                                    n19 = n7;
                                                                    n18 = n6;
                                                                    n17 = n7 + 1;
                                                                    n16 = n6;
                                                                    n51 = 1 + n7;
                                                                    n50 = n6 - -1;
                                                                    n49 = n7;
                                                                    n15 = 1 + n6;
                                                                    n13 = 0;
                                                                    n12 = this.ab[n19][n18];
                                                                    n11 = this.ab[n17][n16];
                                                                    n10 = this.ab[n51][n50];
                                                                    if (-80001 > ~n11) {
                                                                        n11 -= 80000;
                                                                    }
                                                                    if (80000 < n12) {
                                                                        n12 -= 80000;
                                                                    }
                                                                    n9 = this.ab[n49][n15];
                                                                    if (-80001 > ~n10) {
                                                                        n10 -= 80000;
                                                                    }
                                                                    if (80000 < n9) {
                                                                        n9 -= 80000;
                                                                    }
                                                                    if (n12 > n13) {
                                                                        n13 = n12;
                                                                    }
                                                                    if (~n11 < ~n13) {
                                                                        n13 = n11;
                                                                    }
                                                                    if (~n10 < ~n13) {
                                                                        n13 = n10;
                                                                    }
                                                                    if (~n9 < ~n13) {
                                                                        n13 = n9;
                                                                    }
                                                                    if (-80001 >= ~n13) {
                                                                        n13 -= 80000;
                                                                    }
                                                                    if (~n12 <= -80001) break block307;
                                                                    this.ab[n19][n18] = n13;
                                                                    if (!bl2) break block308;
                                                                }
                                                                int[] nArray = this.ab[n19];
                                                                int n52 = n18;
                                                                nArray[n52] = nArray[n52] - 80000;
                                                            }
                                                            if (80000 <= n11) break block309;
                                                            this.ab[n17][n16] = n13;
                                                            if (!bl2) break block310;
                                                        }
                                                        int[] nArray = this.ab[n17];
                                                        int n53 = n16;
                                                        nArray[n53] = nArray[n53] - 80000;
                                                    }
                                                    if (-80001 >= ~n10) break block311;
                                                    this.ab[n51][n50] = n13;
                                                    if (!bl2) break block312;
                                                }
                                                int[] nArray = this.ab[n51];
                                                int n54 = n50;
                                                nArray[n54] = nArray[n54] - 80000;
                                            }
                                            if (n9 >= 80000) break block313;
                                            this.ab[n49][n15] = n13;
                                            if (!bl2) break block306;
                                        }
                                        int[] nArray = this.ab[n49];
                                        int n55 = n15;
                                        nArray[n55] = nArray[n55] - 80000;
                                    }
                                    ++n6;
                                    if (!bl2) continue;
                                }
                                ++n7;
                                if (!bl2) continue block163;
                            }
                            break;
                        }
                        this.kb.c(1);
                        n14 = 1;
                    }
                    n7 = n14;
                    block166: while (true) {
                        int n56 = 95;
                        block167: while (n56 > n7) {
                            n8 = 1;
                            if (bl2) break block291;
                            for (n6 = v589805; 95 > n6; ++n6) {
                                int[] nArray;
                                int n57;
                                int n58;
                                int n59;
                                int n60;
                                int n61;
                                int n62;
                                block314: {
                                    int n63;
                                    block315: {
                                        block316: {
                                            block317: {
                                                block318: {
                                                    block320: {
                                                        int[] nArray3;
                                                        int[] nArray4;
                                                        block319: {
                                                            n56 = n20 = this.d(n6, n7, 126);
                                                            if (bl2) continue block167;
                                                            if (n56 <= 0) continue;
                                                            n19 = n7;
                                                            n18 = n6;
                                                            n17 = n7 + 1;
                                                            n16 = n6;
                                                            int n64 = n7 - -1;
                                                            int n65 = 1 + n6;
                                                            int n66 = n7;
                                                            n15 = n6 + 1;
                                                            n13 = 128 * n7;
                                                            n12 = n6 * 128;
                                                            n11 = 128 + n13;
                                                            n10 = 128 + n12;
                                                            n9 = n13;
                                                            n62 = n12;
                                                            n61 = n11;
                                                            n60 = n10;
                                                            n63 = this.ab[n19][n18];
                                                            n59 = this.ab[n17][n16];
                                                            n58 = this.ab[n64][n65];
                                                            n57 = this.ab[n66][n15];
                                                            int n67 = i.g[-1 + n20];
                                                            if (this.a(false, n19, n18) && ~n63 > -80001) {
                                                                this.ab[n19][n18] = n63 += n67 + 80000;
                                                            }
                                                            if (this.a(false, n17, n16) && ~n59 > -80001) {
                                                                this.ab[n17][n16] = n59 += n67 - -80000;
                                                            }
                                                            if (this.a(false, n64, n65) && -80001 < ~n58) {
                                                                this.ab[n64][n65] = n58 += 80000 + n67;
                                                            }
                                                            if (-80001 >= ~n59) {
                                                                n59 -= 80000;
                                                            }
                                                            if (80000 <= n58) {
                                                                n58 -= 80000;
                                                            }
                                                            if (this.a(false, n66, n15) && ~n57 > -80001) {
                                                                this.ab[n66][n15] = n57 += n67 + 80000;
                                                            }
                                                            if (n63 >= 80000) {
                                                                n63 -= 80000;
                                                            }
                                                            if (n57 >= 80000) {
                                                                n57 -= 80000;
                                                            }
                                                            int n68 = 16;
                                                            if (!this.a(-1 + n19, 26431, n18)) {
                                                                n13 -= n68;
                                                            }
                                                            if (!this.a(n19 - -1, 26431, n18)) {
                                                                n13 += n68;
                                                            }
                                                            if (!this.a(n19, 26431, n18 + -1)) {
                                                                n12 -= n68;
                                                            }
                                                            if (!this.a(n19, 26431, 1 + n18)) {
                                                                n12 += n68;
                                                            }
                                                            if (!this.a(-1 + n17, 26431, n16)) {
                                                                n11 -= n68;
                                                            }
                                                            if (!this.a(n17 + 1, 26431, n16)) {
                                                                n11 += n68;
                                                            }
                                                            if (!this.a(n17, 26431, -1 + n16)) {
                                                                n62 -= n68;
                                                            }
                                                            if (!this.a(n17, 26431, n16 - -1)) {
                                                                n62 += n68;
                                                            }
                                                            if (!this.a(n64 + -1, 26431, n65)) {
                                                                n61 -= n68;
                                                            }
                                                            if (!this.a(1 + n64, 26431, n65)) {
                                                                n61 += n68;
                                                            }
                                                            if (!this.a(n64, 26431, n65 + -1)) {
                                                                n10 -= n68;
                                                            }
                                                            if (!this.a(n64, 26431, 1 + n65)) {
                                                                n10 += n68;
                                                            }
                                                            if (!this.a(n66 - 1, 26431, n15)) {
                                                                n9 -= n68;
                                                            }
                                                            if (!this.a(n66 + 1, 26431, n15)) {
                                                                n9 += n68;
                                                            }
                                                            if (!this.a(n66, 26431, n15 - 1)) {
                                                                n60 -= n68;
                                                            }
                                                            if (!this.a(n66, 26431, n15 - -1)) {
                                                                n60 += n68;
                                                            }
                                                            n59 = -n59;
                                                            n20 = d.g[-1 + n20];
                                                            n57 = -n57;
                                                            n58 = -n58;
                                                            n63 = -n63;
                                                            if (~this.c(n7, n6, -49) < -12001 && this.c(n7, n6, -49) < 24000 && 0 == this.d(n6 - 1, -1 + n7, 120)) break block314;
                                                            if (12000 < this.c(n7, n6, -49) && ~this.c(n7, n6, -49) > -24001 && -1 == ~this.d(n6 - -1, 1 + n7, 115)) {
                                                                nArray = new int[]{this.kb.e(n13, n12, n63, -128), this.kb.e(n11, n62, n59, -122), this.kb.e(n9, n60, n57, 12)};
                                                                this.kb.a(3, nArray, n20, 12345678, false);
                                                                if (!bl2) continue;
                                                            }
                                                            if (~this.c(n7, n6, -49) < -1 && this.c(n7, n6, -49) < 12000 && -1 == ~this.d(n6 - 1, n7 + 1, 111)) break block315;
                                                            if (0 < this.c(n7, n6, -49) && ~this.c(n7, n6, -49) > -12001 && ~this.d(1 + n6, n7 - 1, 103) == -1) break block316;
                                                            if (~n63 == ~n59 && ~n57 == ~n58) break block317;
                                                            if (n63 == n57 && n58 == n59) break block318;
                                                            int n69 = 1;
                                                            if (this.d(-1 + n6, n7 + -1, 117) > 0) {
                                                                n69 = 0;
                                                            }
                                                            if (-1 > ~this.d(n6 - -1, n7 + 1, 110)) {
                                                                n69 = 0;
                                                            }
                                                            if (-1 != ~n69) break block319;
                                                            nArray4 = new int[]{this.kb.e(n11, n62, n59, -114), this.kb.e(n61, n10, n58, 101), this.kb.e(n13, n12, n63, -126)};
                                                            this.kb.a(3, nArray4, n20, 12345678, false);
                                                            nArray3 = new int[]{this.kb.e(n9, n60, n57, -107), this.kb.e(n13, n12, n63, 63), this.kb.e(n61, n10, n58, 44)};
                                                            this.kb.a(3, nArray3, n20, 12345678, false);
                                                            if (!bl2) break block320;
                                                        }
                                                        nArray4 = new int[]{this.kb.e(n13, n12, n63, -112), this.kb.e(n11, n62, n59, -118), this.kb.e(n9, n60, n57, 103)};
                                                        this.kb.a(3, nArray4, n20, 12345678, false);
                                                        nArray3 = new int[]{this.kb.e(n61, n10, n58, -128), this.kb.e(n9, n60, n57, -119), this.kb.e(n11, n62, n59, 52)};
                                                        this.kb.a(3, nArray3, n20, 12345678, false);
                                                    }
                                                    if (!bl2) continue;
                                                }
                                                nArray = new int[]{this.kb.e(n9, n60, n57, -104), this.kb.e(n13, n12, n63, 23), this.kb.e(n11, n62, n59, 91), this.kb.e(n61, n10, n58, 13)};
                                                this.kb.a(4, nArray, n20, 12345678, false);
                                                if (!bl2) continue;
                                            }
                                            nArray = new int[]{this.kb.e(n13, n12, n63, 78), this.kb.e(n11, n62, n59, 46), this.kb.e(n61, n10, n58, -113), this.kb.e(n9, n60, n57, -125)};
                                            this.kb.a(4, nArray, n20, 12345678, false);
                                            if (!bl2) continue;
                                        }
                                        nArray = new int[]{this.kb.e(n11, n62, n59, 121), this.kb.e(n61, n10, n58, 39), this.kb.e(n13, n12, n63, 73)};
                                        this.kb.a(3, nArray, n20, 12345678, false);
                                        if (!bl2) continue;
                                    }
                                    nArray = new int[]{this.kb.e(n9, n60, n57, -107), this.kb.e(n13, n12, n63, -122), this.kb.e(n61, n10, n58, 35)};
                                    this.kb.a(3, nArray, n20, 12345678, false);
                                    if (!bl2) continue;
                                }
                                nArray = new int[]{this.kb.e(n61, n10, n58, -120), this.kb.e(n9, n60, n57, -116), this.kb.e(n11, n62, n59, 117)};
                                this.kb.a(3, nArray, n20, 12345678, false);
                                if (!bl2) continue;
                            }
                            ++n7;
                            if (!bl2) continue block166;
                        }
                        break;
                    }
                    this.kb.a(-50, 50, -10, -50, true, 50, -98);
                    this.db[n4] = this.kb.a(0, 8, 1536, -112, 64, 169, 1536, true, 0);
                    n8 = n7 = 0;
                }
                while (-65 < ~n7) {
                    this.c.a(this.db[n4][n7], (byte)118);
                    ++n7;
                    if (!bl2) {
                        if (!bl2) continue;
                    }
                    break block292;
                }
                if (this.db[n4][0] == null) {
                    throw new RuntimeException(ob[40]);
                }
            }
            n7 = 0;
            do {
                int n71 = -97;
                n71 = ~n7;
                block171: while (true) {
                    if (n70 >= n71) return;
                    if (bl2) return;
                    for (n6 = 0; 96 > n6; ++n6) {
                        int n71 = ~this.ab[n7][n6];
                        n71 = -80001;
                        if (bl2) continue block171;
                        if (n70 > n71) continue;
                        int[] nArray = this.ab[n7];
                        int n72 = n6;
                        nArray[n72] = nArray[n72] - 80000;
                        if (!bl2) continue;
                    }
                    break;
                }
                ++n7;
            } while (!bl2);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[39] + n2 + ',' + n3 + ',' + bl + ',' + n4 + ',' + n5 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final int d(int n2, int n3, int n4) {
        boolean bl = client.vh;
        try {
            int n5;
            block17: {
                block18: {
                    block16: {
                        ++a;
                        if (0 > n3 || 96 <= n3 || n2 < 0 || 96 <= n2) {
                            return 0;
                        }
                        n5 = 0;
                        if (n4 < 99) {
                            this.E = null;
                        }
                        if (~n3 > -49 || n2 >= 48) break block16;
                        n5 = 1;
                        n3 -= 48;
                        if (!bl) break block17;
                    }
                    if (n3 >= 48 || n2 < 48) break block18;
                    n5 = 2;
                    n2 -= 48;
                    if (!bl) break block17;
                }
                if (48 <= n3 && n2 >= 48) {
                    n5 = 3;
                    n3 -= 48;
                    n2 -= 48;
                }
            }
            return this.A[n5][n2 + n3 * 48];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[17] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void b(int n2, byte by, int n3, int n4) {
        try {
            ++lb;
            int n8 = n3 / 12;
            int n7 = n4 / 12;
            int n6 = (n3 + -1) / 12;
            int n5 = (n4 - 1) / 12;
            this.a(n4, n2, n7, 2, n8, n3);
            if (~n6 != ~n8) {
                this.a(n4, n2, n7, 2, n6, n3);
            }
            if (n5 != n7) {
                this.a(n4, n2, n5, 2, n8, n3);
            }
            if (n6 != n8 && ~n5 != ~n7) {
                this.a(n4, n2, n5, 2, n6, n3);
            }
            if (by > 23) return;
            this.c(122, -121, -56, -127, -62);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[24] + n2 + ',' + by + ',' + n3 + ',' + n4 + ')');
        }
    }

    private final int d(int n2, int n3, int n4, int n5) {
        try {
            int n6;
            ++hb;
            if (n4 != 15282) {
                this.m = null;
            }
            if ((n6 = this.b(n2, n5, n4 + -15278, n3)) == 0) {
                return -1;
            }
            int n7 = da.N[n6 - 1];
            if (2 != n7) {
                return 0;
            }
            return 1;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[42] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    private final void c(int n2, int n3, int n4, int n5) {
        try {
            ++T;
            this.bb[n4][n2] = ib.a(this.bb[n4][n2], n3 - n5);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[44] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    private final void a(int n2, int n3, byte by, int n4) {
        try {
            ++J;
            this.bb[n4][n3] = d.a(this.bb[n4][n3], n2);
            if (by >= -47) {
                this.eb = null;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[38] + n2 + ',' + n3 + ',' + by + ',' + n4 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(ca[] caArray, byte n2) {
        boolean bl = client.vh;
        try {
            int n3;
            block31: {
                ++j;
                int n4 = 0;
                block17: while (true) {
                    int n5 = n4;
                    int n6 = 94;
                    block18: while (n5 < n6) {
                        n3 = 0;
                        if (bl) break block31;
                        int n7 = n3;
                        block19: while (true) {
                            int n8 = ~n7;
                            block20: while (n8 > -95) {
                                block32: {
                                    int n9;
                                    int n10;
                                    int n11;
                                    block34: {
                                        block33: {
                                            n5 = 48000;
                                            n6 = this.c(n4, n7, -49);
                                            if (bl) continue block18;
                                            if (n5 >= n6 || 60000 <= this.c(n4, n7, -49)) break block32;
                                            n11 = this.c(n4, n7, -49) + -48001;
                                            int n12 = this.b(n4, n7, -91);
                                            if (0 == n12 || n12 == 4) break block33;
                                            n10 = ub.g[n11];
                                            n9 = f.f[n11];
                                            if (!bl) break block34;
                                        }
                                        n9 = ub.g[n11];
                                        n10 = f.f[n11];
                                    }
                                    this.a(n4, n11, false, n7);
                                    ca ca2 = caArray[fb.f[n11]].a(false, -120, false, false, true);
                                    int n13 = 128 * (n10 + n4 + n4) / 2;
                                    int n14 = (n9 + n7 + n7) * 128 / 2;
                                    ca2.a(n13, n14, -this.f(n13, n14, 74), true);
                                    ca2.g(0, -999999, 0, this.b(n4, n7, -78) * 32);
                                    this.c.a(ca2, (byte)118);
                                    ca2.a(48, 48, -10, n2 ^ 9, -50, -50);
                                    if (~n10 >= -2 && n9 <= 1) break block32;
                                    int n15 = n4;
                                    block21: while (true) {
                                        int n16 = n4 + n10;
                                        int n17 = n15;
                                        block22: while (n16 > n17) {
                                            n8 = n7;
                                            if (bl) continue block20;
                                            for (int i2 = v593288; n9 + n7 > i2; ++i2) {
                                                block37: {
                                                    int n18;
                                                    block35: {
                                                        block36: {
                                                            n16 = n4;
                                                            n17 = n15;
                                                            if (bl) continue block22;
                                                            if (n16 >= n17 && ~i2 >= ~n7 || n11 != this.c(n15, i2, n2 + 64) + -48001) continue;
                                                            n14 = i2;
                                                            n13 = n15;
                                                            n18 = 0;
                                                            if (-49 >= ~n13 && n14 < 48) break block35;
                                                            if (48 > n13 && -49 >= ~n14) break block36;
                                                            if (48 > n13 || n14 < 48) break block37;
                                                            n18 = 3;
                                                            n14 -= 48;
                                                            n13 -= 48;
                                                            if (!bl) break block37;
                                                        }
                                                        n18 = 2;
                                                        n14 -= 48;
                                                        if (!bl) break block37;
                                                    }
                                                    n13 -= 48;
                                                    n18 = 1;
                                                }
                                                this.s[n18][n13 * 48 + n14] = 0;
                                                if (!bl) continue;
                                            }
                                            ++n15;
                                            if (!bl) continue block21;
                                        }
                                        break;
                                    }
                                }
                                ++n7;
                                if (!bl) continue block19;
                            }
                            break;
                        }
                        ++n4;
                        if (!bl) continue block17;
                    }
                    break;
                }
                n3 = n2;
            }
            if (n3 == -113) return;
            this.b(-116, 16, 84);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(ob[5]);
            if (caArray != null) {
                string = ob[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(')').toString());
            }
            string = ob[2];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final int g(int n2, int n3, int n4) {
        boolean bl = client.vh;
        try {
            ++y;
            if (n4 < 0 || 96 <= n4 || 0 > n3 || -97 >= ~n3) return 0;
            int n5 = 0;
            if (n2 != 2) {
                return 79;
            }
            if (~n4 <= -49 && n3 < 48) {
                n5 = 1;
                n4 -= 48;
                if (!bl) return (0xFF & this.L[n5][48 * n4 - -n3]) * 3;
            }
            if (n4 < 48 && -49 >= ~n3) {
                n3 -= 48;
                n5 = 2;
                if (!bl) return (0xFF & this.L[n5][48 * n4 - -n3]) * 3;
            }
            if (-49 < ~n4 || n3 < 48) return (0xFF & this.L[n5][48 * n4 - -n3]) * 3;
            n5 = 3;
            n3 -= 48;
            n4 -= 48;
            return (0xFF & this.L[n5][48 * n4 - -n3]) * 3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[7] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(boolean bl, int n2, int n3, int n4, int n5) {
        boolean bl2 = client.vh;
        try {
            block19: {
                block21: {
                    block23: {
                        block22: {
                            block20: {
                                ++r;
                                if (0 > n4) return;
                                if (0 > n3) return;
                                if (-96 >= ~n4) return;
                                if (95 <= n3) {
                                    return;
                                }
                                if (1 != u.a[n5]) break block19;
                                if (n2 != 0) break block20;
                                this.bb[n4][n3] = ib.a(this.bb[n4][n3], 65534);
                                if (n3 <= 0) break block21;
                                this.c(-1 + n3, 65535, n4, 4);
                                if (!bl2) break block21;
                            }
                            if (-2 != ~n2) break block22;
                            this.bb[n4][n3] = ib.a(this.bb[n4][n3], 65533);
                            if (~n4 >= -1) break block21;
                            this.c(n3, 65535, -1 + n4, 8);
                            if (!bl2) break block21;
                        }
                        if (2 != n2) break block23;
                        this.bb[n4][n3] = ib.a(this.bb[n4][n3], 65519);
                        if (!bl2) break block21;
                    }
                    if (n2 == 3) {
                        this.bb[n4][n3] = ib.a(this.bb[n4][n3], 65503);
                    }
                }
                this.c(1, 1, -59, n4, n3);
            }
            if (bl) return;
            this.U = null;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[10] + bl + ',' + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    final void a(int n2, byte by, int n3, int n4) {
        block8: {
            try {
                int n5;
                int n6;
                block7: {
                    this.b(by ^ 0x2791);
                    ++jb;
                    n6 = (24 + n2) / 48;
                    if (by != -90) {
                        this.c(58, -126, -4);
                    }
                    this.a(n2, 122, true, n4, n3);
                    n5 = (24 + n3) / 48;
                    if (n4 == 0) break block7;
                    break block8;
                }
                this.a(n2, 112, false, 1, n3);
                this.a(n2, by ^ 0xFFFFFFE3, false, 2, n3);
                this.b(n4, 0, n6 - 1, 0, -1 + n5);
                this.b(n4, 1, n6, by + 90, n5 + -1);
                this.b(n4, 2, n6 + -1, 0, n5);
                this.b(n4, 3, n6, by + 90, n5);
                this.a(0);
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, ob[41] + n2 + ',' + by + ',' + n3 + ',' + n4 + ')');
            }
        }
    }

    private final void a(int n2, int n3, int n4, int n5, byte by, int n6) {
        try {
            int n7;
            block8: {
                block7: {
                    ++fb;
                    n7 = ib.d[n2];
                    if (~this.ab[n6][n4] > -80001) break block7;
                    break block8;
                }
                int[] nArray = this.ab[n6];
                int n8 = n4;
                nArray[n8] = nArray[n8] + (n7 + 80000);
            }
            if (80000 > this.ab[n3][n5]) {
                int[] nArray = this.ab[n3];
                int n9 = n5;
                nArray[n9] = nArray[n9] + (n7 + 80000);
            }
            int n10 = -39 / ((2 - by) / 51);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[35] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + by + ',' + n6 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final int a(int[] nArray, int n2, byte n3, int n4, int[] nArray2, int n5, int n6, int n7, int n8, boolean bl) {
        boolean bl2 = client.vh;
        try {
            int n9;
            int n10;
            int n11;
            boolean bl3;
            int n12;
            int n13;
            int n14;
            block109: {
                int n15;
                int n16;
                block107: {
                    n16 = 0;
                    while (-97 < ~n16) {
                        block108: {
                            n15 = 0;
                            if (bl2) break block107;
                            for (n14 = v597972; n14 < 96; ++n14) {
                                this.B[n16][n14] = 0;
                                if (!bl2) {
                                    if (!bl2) continue;
                                }
                                break block108;
                            }
                            ++n16;
                        }
                        if (!bl2) continue;
                    }
                    ++cb;
                    n16 = 0;
                    n15 = 0;
                }
                n14 = n15;
                n13 = n5;
                n12 = n6;
                this.B[n5][n6] = 99;
                nArray[n16] = n5;
                nArray2[n16++] = n6;
                int n17 = nArray.length;
                bl3 = false;
                while (~n16 != ~n14) {
                    n13 = nArray[n14];
                    n12 = nArray2[n14];
                    n14 = (1 + n14) % n17;
                    if (~n2 >= ~n13 && ~n7 <= ~n13 && n8 <= n12 && ~n12 >= ~n4) {
                        bl3 = true;
                        if (!bl2) break;
                    }
                    if (bl) {
                        if (-1 > ~n13 && n2 <= n13 + -1 && ~(n13 + -1) >= ~n7 && n8 <= n12 && n4 >= n12 && (this.bb[n13 - 1][n12] & 8) == 0) {
                            bl3 = true;
                            if (!bl2) break;
                        }
                        if (~n13 > -96 && ~n2 >= ~(1 + n13) && n13 - -1 <= n7 && ~n8 >= ~n12 && ~n12 >= ~n4 && -1 == ~(2 & this.bb[n13 + 1][n12])) {
                            bl3 = true;
                            if (!bl2) break;
                        }
                        if (0 < n12 && ~n13 <= ~n2 && n7 >= n13 && ~n8 >= ~(n12 + -1) && ~(-1 + n12) >= ~n4 && ~(4 & this.bb[n13][n12 + -1]) == -1) {
                            bl3 = true;
                            if (!bl2) break;
                        }
                        if (n12 < 95 && ~n13 <= ~n2 && n13 <= n7 && ~(n12 - -1) <= ~n8 && ~(n12 + 1) >= ~n4 && -1 == ~(1 & this.bb[n13][n12 - -1])) {
                            bl3 = true;
                            if (!bl2) break;
                        }
                    }
                    if (0 < n13 && ~this.B[-1 + n13][n12] == -1 && (this.bb[-1 + n13][n12] & 0x78) == 0) {
                        nArray[n16] = -1 + n13;
                        nArray2[n16] = n12;
                        this.B[-1 + n13][n12] = 2;
                        n16 = (n16 + 1) % n17;
                    }
                    if (95 > n13 && this.B[1 + n13][n12] == 0 && (this.bb[1 + n13][n12] & 0x72) == 0) {
                        nArray[n16] = 1 + n13;
                        nArray2[n16] = n12;
                        this.B[n13 + 1][n12] = 8;
                        n16 = (1 + n16) % n17;
                    }
                    if (n12 > 0 && 0 == this.B[n13][-1 + n12] && -1 == ~(0x74 & this.bb[n13][n12 + -1])) {
                        nArray[n16] = n13;
                        nArray2[n16] = n12 + -1;
                        this.B[n13][-1 + n12] = 1;
                        n16 = (n16 + 1) % n17;
                    }
                    if (~n12 > -96 && this.B[n13][1 + n12] == 0 && -1 == ~(0x71 & this.bb[n13][1 + n12])) {
                        nArray[n16] = n13;
                        nArray2[n16] = n12 - -1;
                        this.B[n13][n12 - -1] = 4;
                        n16 = (n16 - -1) % n17;
                    }
                    if (~n13 < -1 && ~n12 < -1 && (0x74 & this.bb[n13][-1 + n12]) == 0 && -1 == ~(0x78 & this.bb[n13 - 1][n12]) && 0 == (0x7C & this.bb[n13 - 1][n12 + -1]) && 0 == this.B[-1 + n13][-1 + n12]) {
                        nArray[n16] = -1 + n13;
                        nArray2[n16] = n12 - 1;
                        this.B[-1 + n13][n12 + -1] = 3;
                        n16 = (1 + n16) % n17;
                    }
                    if (~n13 > -96 && ~n12 < -1 && 0 == (this.bb[n13][n12 + -1] & 0x74) && 0 == (this.bb[1 + n13][n12] & 0x72) && ~(this.bb[n13 - -1][-1 + n12] & 0x76) == -1 && 0 == this.B[1 + n13][n12 + -1]) {
                        nArray[n16] = 1 + n13;
                        nArray2[n16] = -1 + n12;
                        this.B[n13 - -1][-1 + n12] = 9;
                        n16 = (1 + n16) % n17;
                    }
                    if (~n13 < -1 && ~n12 > -96 && -1 == ~(this.bb[n13][1 + n12] & 0x71) && ~(this.bb[n13 + -1][n12] & 0x78) == -1 && ~(this.bb[n13 - 1][1 + n12] & 0x79) == -1 && -1 == ~this.B[n13 + -1][1 + n12]) {
                        nArray[n16] = -1 + n13;
                        nArray2[n16] = 1 + n12;
                        n16 = (1 + n16) % n17;
                        this.B[-1 + n13][n12 - -1] = 6;
                    }
                    if (n13 >= 95) continue;
                    n11 = -96;
                    n10 = ~n12;
                    if (!bl2) {
                        if (n11 >= n10 || 0 != (0x71 & this.bb[n13][1 + n12]) || 0 != (this.bb[n13 + 1][n12] & 0x72) || 0 != (0x73 & this.bb[n13 - -1][1 + n12]) || this.B[n13 - -1][1 + n12] != 0) continue;
                        nArray[n16] = 1 + n13;
                        nArray2[n16] = 1 + n12;
                        this.B[1 + n13][1 + n12] = 12;
                        n16 = (n16 - -1) % n17;
                        if (!bl2) continue;
                    }
                    break block109;
                }
                n11 = n3;
                n10 = -48;
            }
            if (n11 > n10) {
                return -42;
            }
            if (!bl3) return -1;
            n14 = 0;
            nArray[n14] = n13;
            nArray2[n14++] = n12;
            int n18 = n9 = this.B[n13][n12];
            do {
                block115: {
                    block114: {
                        block113: {
                            block112: {
                                int n19;
                                int n20;
                                block111: {
                                    block110: {
                                        if (~n5 != ~n13) break block110;
                                        n20 = ~n6;
                                        n19 = ~n12;
                                        if (bl2) break block111;
                                        if (n20 == n19) return n14;
                                    }
                                    n20 = n9;
                                    n19 = n18;
                                }
                                if (n20 != n19) {
                                    n9 = n18;
                                    nArray[n14] = n13;
                                    nArray2[n14++] = n12;
                                }
                                if ((n18 & 1) != 0) break block112;
                                if (-1 == ~(4 & n18)) break block113;
                                --n12;
                                if (!bl2) break block113;
                            }
                            ++n12;
                        }
                        if (~(2 & n18) != -1) break block114;
                        if (-1 == ~(n18 & 8)) break block115;
                        --n13;
                        if (!bl2) break block115;
                    }
                    ++n13;
                }
                n18 = this.B[n13][n12];
            } while (!bl2);
            return n14;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(ob[20]).append(nArray != null ? ob[1] : ob[2]).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',');
            if (nArray2 != null) {
                string = ob[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(bl).append(')').toString());
            }
            string = ob[2];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(bl).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final int f(int n2, int n3, int n4) {
        try {
            int n5;
            int n6;
            int n7;
            ++k;
            int n11 = n2 >> -1924199641;
            int n10 = n3 >> -1127688185;
            int n12 = 79 / ((n4 - 30) / 35);
            int n9 = 0x7F & n2;
            int n8 = 0x7F & n3;
            if (~n11 > -1) return 0;
            if (0 > n10) return 0;
            if (95 <= n11) return 0;
            if (95 <= n10) {
                return 0;
            }
            if (n9 <= 128 + -n8) {
                n7 = this.g(2, n10, n11);
                n6 = this.g(2, n10, 1 + n11) + -n7;
                n5 = -n7 + this.g(2, 1 + n10, n11);
                if (!client.vh) return n5 * n8 / 128 + (n7 - -(n6 * n9 / 128));
            }
            n7 = this.g(2, n10 - -1, 1 + n11);
            n6 = -n7 + this.g(2, n10 + 1, n11);
            n5 = -n7 + this.g(2, n10, 1 + n11);
            n9 = -n9 + 128;
            n8 = 128 + -n8;
            return n5 * n8 / 128 + (n7 - -(n6 * n9 / 128));
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[34] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void c(int n2, int n3, int n4, int n5, int n6) {
        boolean bl = client.vh;
        try {
            int n7 = 116 % ((n4 - 15) / 39);
            ++D;
            if (1 > n5) return;
            if (-2 < ~n6) return;
            if (96 <= n2 + n5) return;
            if (96 <= n3 + n6) return;
            int n8 = n5;
            do {
                int n9 = n8;
                block12: while (true) {
                    if (n9 > n2 + n5) return;
                    if (bl) return;
                    int n10 = n6;
                    while (~n10 >= ~(n6 - -n3)) {
                        block17: {
                            block16: {
                                n9 = 0x63 & this.b((byte)-38, n10, n8);
                                if (bl) continue block12;
                                if (n9 != 0 || ~(0x59 & this.b((byte)-38, n10, -1 + n8)) != -1 || (this.b((byte)-38, n10 - 1, n8) & 0x56) != 0 || ~(this.b((byte)-38, -1 + n10, n8 - 1) & 0x6C) != -1) break block16;
                                this.b(0, (byte)118, n8, n10);
                                if (!bl) break block17;
                            }
                            this.b(35, (byte)50, n8, n10);
                        }
                        ++n10;
                        if (!bl) continue;
                    }
                    break;
                }
                ++n8;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[6] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final int a(byte by, int n2, int n3) {
        boolean bl = client.vh;
        try {
            ++n;
            if (n2 < 0 || 96 <= n2 || -1 < ~n3 || ~n3 <= -97) return 0;
            int n4 = 0;
            if (by != 104) {
                return -59;
            }
            if (~n2 <= -49 && 48 > n3) {
                n2 -= 48;
                n4 = 1;
                if (!bl) return this.eb[n4][n3 + 48 * n2] & 0xFF;
            }
            if (-49 < ~n2 && ~n3 <= -49) {
                n4 = 2;
                n3 -= 48;
                if (!bl) return this.eb[n4][n3 + 48 * n2] & 0xFF;
            }
            if (-49 < ~n2 || 48 > n3) return this.eb[n4][n3 + 48 * n2] & 0xFF;
            n2 -= 48;
            n4 = 3;
            n3 -= 48;
            return this.eb[n4][n3 + 48 * n2] & 0xFF;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[22] + by + ',' + n2 + ',' + n3 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final boolean a(int n2, int n3, int n4) {
        try {
            if (n3 != 26431) {
                return false;
            }
            ++t;
            if (~this.d(n4, n2, 119) < -1) return true;
            if (~this.d(n4, -1 + n2, 110) < -1) return true;
            if (0 < this.d(n4 + -1, n2 + -1, 109)) return true;
            if (-1 > ~this.d(-1 + n4, n2, n3 + -26318)) return true;
            return false;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[9] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    private final int d(int n2, int n3, int n4, int n5, int n6) {
        try {
            int n7;
            block6: {
                block5: {
                    if (n2 != -8509) {
                        return 58;
                    }
                    ++p;
                    n7 = this.b(n5, n3, n2 + 8513, n6);
                    if (-1 == ~n7) break block5;
                    break block6;
                }
                return n4;
            }
            return qa.K[-1 + n7];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[16] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final int c(int n2, int n3, int n4) {
        boolean bl = client.vh;
        try {
            int n5;
            block15: {
                block13: {
                    block14: {
                        ++z;
                        if (0 > n2 || n2 >= 96 || -1 < ~n3 || ~n3 <= -97) {
                            return 0;
                        }
                        n5 = 0;
                        if (~n2 <= n4 && ~n3 > -49) break block13;
                        if (-49 < ~n2 && ~n3 <= -49) break block14;
                        if (~n2 > -49 || ~n3 > -49) break block15;
                        n3 -= 48;
                        n5 = 3;
                        n2 -= 48;
                        if (!bl) break block15;
                    }
                    n5 = 2;
                    n3 -= 48;
                    if (!bl) break block15;
                }
                n2 -= 48;
                n5 = 1;
            }
            return this.s[n5][n2 * 48 + n3];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[4] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    /*
     * Unable to fully structure code
     */
    private final void b(int var1_1, int var2_2, int var3_3, int var4_4, int var5_5) {
        block91: {
            var13_6 = client.vh;
            try {
                if (var4_4 != 0) {
                    return;
                }
                ++k.l;
                var6_7 = "m" + var1_1 + var3_3 / 10 + var3_3 % 10 + var5_5 / 10 + var5_5 % 10;
                try {
                    block98: {
                        block97: {
                            block96: {
                                block95: {
                                    block94: {
                                        block93: {
                                            block92: {
                                                block102: {
                                                    block111: {
                                                        block90: {
                                                            block89: {
                                                                block88: {
                                                                    block87: {
                                                                        block104: {
                                                                            block103: {
                                                                                block86: {
                                                                                    block85: {
                                                                                        if (null == this.Q) break block102;
                                                                                        var7_9 = na.a(var6_7 + k.ob[26], 0, this.Q, -126);
                                                                                        if (null == var7_9 && this.I != null) {
                                                                                            var7_9 = na.a(var6_7 + k.ob[26], 0, this.I, -125);
                                                                                        }
                                                                                        if (null == var7_9 || -1 <= ~var7_9.length) ** GOTO lbl-1000
                                                                                        var8_11 = 0;
                                                                                        var9_13 = 0;
                                                                                        var10_14 = 0;
                                                                                        block10: while (2304 > var10_14) {
                                                                                            var11_15 = var7_9[var8_11++] & 255;
                                                                                            v0 = -129;
                                                                                            v1 = ~var11_15;
                                                                                            if (!var13_6) {
                                                                                                if (v0 < v1) {
                                                                                                    var9_13 = var11_15;
                                                                                                    this.L[var2_2][var10_14++] = (byte)var11_15;
                                                                                                }
                                                                                                if (var11_15 >= 128) {
                                                                                                    var12_16 = 0;
                                                                                                    while (~(var11_15 - 128) < ~var12_16) {
                                                                                                        this.L[var2_2][var10_14++] = (byte)var9_13;
                                                                                                        ++var12_16;
                                                                                                        if (var13_6) continue block10;
                                                                                                        if (!var13_6) continue;
                                                                                                    }
                                                                                                }
                                                                                                if (!var13_6) continue;
                                                                                            }
                                                                                            ** GOTO lbl38
                                                                                        }
                                                                                        var9_13 = 64;
                                                                                        var10_14 = 0;
                                                                                        block12: do {
                                                                                            v0 = ~var10_14;
                                                                                            v1 = -49;
lbl38:
                                                                                            // 2 sources

                                                                                            if (v0 <= v1) break;
                                                                                            v2 = 0;
                                                                                            if (var13_6) break block85;
                                                                                            for (var11_15 = v601407; var11_15 < 48; ++var11_15) {
                                                                                                var9_13 = 127 & var9_13 + this.L[var2_2][var11_15 * 48 - -var10_14];
                                                                                                this.L[var2_2][var10_14 + var11_15 * 48] = (byte)(var9_13 * 2);
                                                                                                if (var13_6) continue block12;
                                                                                                if (!var13_6) continue;
                                                                                            }
                                                                                            ++var10_14;
                                                                                        } while (!var13_6);
                                                                                        var9_13 = 0;
                                                                                        v2 = var10_14 = 0;
                                                                                    }
                                                                                    block14: while (var10_14 < 2304) {
                                                                                        var11_15 = 255 & var7_9[var8_11++];
                                                                                        v3 = ~var11_15;
                                                                                        v4 = -129;
                                                                                        if (!var13_6) {
                                                                                            if (v3 > v4) {
                                                                                                var9_13 = var11_15;
                                                                                                this.eb[var2_2][var10_14++] = (byte)var11_15;
                                                                                            }
                                                                                            if (var11_15 >= 128) {
                                                                                                var12_16 = 0;
                                                                                                while (~var12_16 > ~(-128 + var11_15)) {
                                                                                                    this.eb[var2_2][var10_14++] = (byte)var9_13;
                                                                                                    ++var12_16;
                                                                                                    if (var13_6) continue block14;
                                                                                                    if (!var13_6) continue;
                                                                                                }
                                                                                            }
                                                                                            if (!var13_6) continue;
                                                                                        }
                                                                                        ** GOTO lbl73
                                                                                    }
                                                                                    var9_13 = 35;
                                                                                    var10_14 = 0;
                                                                                    block16: do {
                                                                                        v3 = -49;
                                                                                        v4 = ~var10_14;
lbl73:
                                                                                        // 2 sources

                                                                                        if (v3 >= v4) break;
                                                                                        v5 = 0;
                                                                                        if (var13_6) break block86;
                                                                                        var11_15 = v5;
                                                                                        while (-49 < ~var11_15) {
                                                                                            var9_13 = 127 & var9_13 + this.eb[var2_2][var10_14 + 48 * var11_15];
                                                                                            this.eb[var2_2][var10_14 + 48 * var11_15] = (byte)(2 * var9_13);
                                                                                            ++var11_15;
                                                                                            if (var13_6) continue block16;
                                                                                            if (!var13_6) continue;
                                                                                        }
                                                                                        ++var10_14;
                                                                                    } while (!var13_6);
                                                                                    if (var13_6) lbl-1000:
                                                                                    // 2 sources

                                                                                    {
                                                                                        var8_11 = 0;
                                                                                        while (~var8_11 > -2305) {
                                                                                            this.L[var2_2][var8_11] = 0;
                                                                                            this.eb[var2_2][var8_11] = 0;
                                                                                            ++var8_11;
                                                                                            if (!var13_6) {
                                                                                                if (!var13_6) continue;
                                                                                            }
                                                                                            break;
                                                                                        }
                                                                                    } else {
                                                                                        var7_9 = na.a(var6_7 + k.ob[29], 0, this.gb, -125);
                                                                                    }
                                                                                    if (var7_9 == null && null != this.m) {
                                                                                        var7_9 = na.a(var6_7 + k.ob[29], 0, this.m, var4_4 ^ -125);
                                                                                    }
                                                                                    if (var7_9 == null) break block103;
                                                                                    v5 = -1;
                                                                                }
                                                                                if (v5 != ~var7_9.length) break block104;
                                                                            }
                                                                            throw new IOException();
                                                                        }
                                                                        var8_11 = 0;
                                                                        for (var9_13 = 0; 2304 > var9_13; ++var9_13) {
                                                                            this.f[var2_2][var9_13] = var7_9[var8_11++];
                                                                            if (!var13_6) {
                                                                                if (!var13_6) continue;
                                                                            }
                                                                            break block87;
                                                                        }
                                                                        var9_13 = 0;
                                                                    }
                                                                    while (var9_13 < 2304) {
                                                                        this.P[var2_2][var9_13] = var7_9[var8_11++];
                                                                        ++var9_13;
                                                                        if (!var13_6) {
                                                                            if (!var13_6) continue;
                                                                        }
                                                                        break block88;
                                                                    }
                                                                    var9_13 = 0;
                                                                }
                                                                while (~var9_13 > -2305) {
                                                                    this.s[var2_2][var9_13] = ib.a(255, var7_9[var8_11++]);
                                                                    ++var9_13;
                                                                    if (!var13_6) {
                                                                        if (!var13_6) continue;
                                                                    }
                                                                    break block89;
                                                                }
                                                                var9_13 = 0;
                                                            }
                                                            while (~var9_13 > -2305) {
                                                                var10_14 = 255 & var7_9[var8_11++];
                                                                v6 = ~var10_14;
                                                                v7 = -1;
                                                                if (!var13_6) {
                                                                    if (v6 < v7) {
                                                                        this.s[var2_2][var9_13] = var10_14 + 12000;
                                                                    }
                                                                    ++var9_13;
                                                                    if (!var13_6) continue;
                                                                }
                                                                ** GOTO lbl142
                                                            }
                                                            var9_13 = 0;
                                                            block23: while (true) {
                                                                block106: {
                                                                    block105: {
                                                                        v6 = 2304;
                                                                        v7 = var9_13;
lbl142:
                                                                        // 2 sources

                                                                        if (v6 <= v7) break;
                                                                        var10_14 = var7_9[var8_11++] & 255;
                                                                        v8 = 128;
                                                                        v9 = var10_14;
                                                                        if (var13_6) ** GOTO lbl165
                                                                        if (v8 <= v9) break block105;
                                                                        this.A[var2_2][var9_13++] = (byte)var10_14;
                                                                        if (!var13_6) break block106;
                                                                    }
                                                                    var11_15 = 0;
                                                                    while (~(-128 + var10_14) < ~var11_15) {
                                                                        this.A[var2_2][var9_13++] = 0;
                                                                        ++var11_15;
                                                                        if (var13_6) continue block23;
                                                                        if (!var13_6) continue;
                                                                    }
                                                                }
                                                                if (var13_6) break;
                                                            }
                                                            var9_13 = 0;
                                                            var10_14 = 0;
                                                            block25: while (true) {
                                                                block108: {
                                                                    block107: {
                                                                        v8 = var10_14;
                                                                        v9 = 2304;
lbl165:
                                                                        // 2 sources

                                                                        if (v8 >= v9) break;
                                                                        var11_15 = var7_9[var8_11++] & 255;
                                                                        v10 = ~var11_15;
                                                                        v11 = -129;
                                                                        if (var13_6) ** GOTO lbl186
                                                                        if (v10 <= v11) break block107;
                                                                        this.R[var2_2][var10_14++] = (byte)var11_15;
                                                                        var9_13 = var11_15;
                                                                        if (!var13_6) break block108;
                                                                    }
                                                                    for (var12_16 = 0; -128 + var11_15 > var12_16; ++var12_16) {
                                                                        this.R[var2_2][var10_14++] = (byte)var9_13;
                                                                        if (var13_6) continue block25;
                                                                        if (!var13_6) continue;
                                                                    }
                                                                }
                                                                if (var13_6) break;
                                                            }
                                                            var10_14 = 0;
                                                            block27: while (true) {
                                                                block110: {
                                                                    block109: {
                                                                        v10 = 2304;
                                                                        v11 = var10_14;
lbl186:
                                                                        // 2 sources

                                                                        if (v10 <= v11) break;
                                                                        var11_15 = 255 & var7_9[var8_11++];
                                                                        v12 = ~var11_15;
                                                                        v13 = -129;
                                                                        if (var13_6) break block90;
                                                                        if (v12 <= v13) break block109;
                                                                        this.mb[var2_2][var10_14++] = (byte)var11_15;
                                                                        if (!var13_6) break block110;
                                                                    }
                                                                    for (var12_16 = 0; var12_16 < -128 + var11_15; ++var12_16) {
                                                                        this.mb[var2_2][var10_14++] = 0;
                                                                        if (var13_6) continue block27;
                                                                        if (!var13_6) continue;
                                                                    }
                                                                }
                                                                if (var13_6) break;
                                                            }
                                                            if ((var7_9 = na.a(var6_7 + k.ob[30], 0, this.gb, -127)) == null) break block111;
                                                            v12 = 0;
                                                            v13 = var7_9.length;
                                                        }
                                                        if (v12 >= v13) break block111;
                                                        var8_11 = 0;
                                                        var10_14 = 0;
                                                        while (~var10_14 > -2305) {
                                                            block113: {
                                                                block112: {
                                                                    var11_15 = var7_9[var8_11++] & 255;
                                                                    if (var13_6) break block91;
                                                                    if (-129 >= ~var11_15) break block112;
                                                                    this.s[var2_2][var10_14++] = 48000 + var11_15;
                                                                    if (!var13_6) break block113;
                                                                }
                                                                var10_14 += -128 + var11_15;
                                                            }
                                                            if (!var13_6) continue;
                                                        }
                                                    }
                                                    if (!var13_6) break block91;
                                                }
                                                var7_9 = new byte[20736];
                                                ta.a(k.ob[28] + var6_7 + k.ob[31], var4_4 ^ -19675, var7_9, 20736);
                                                var8_11 = 0;
                                                var9_13 = 0;
                                                for (var10_14 = 0; 2304 > var10_14; ++var10_14) {
                                                    var8_11 = 255 & var8_11 - -var7_9[var9_13++];
                                                    this.L[var2_2][var10_14] = (byte)var8_11;
                                                    if (!var13_6) {
                                                        if (!var13_6) continue;
                                                    }
                                                    break block92;
                                                }
                                                var8_11 = 0;
                                            }
                                            var10_14 = 0;
                                            while (-2305 < ~var10_14) {
                                                var8_11 = 255 & var8_11 + var7_9[var9_13++];
                                                this.eb[var2_2][var10_14] = (byte)var8_11;
                                                ++var10_14;
                                                if (!var13_6) {
                                                    if (!var13_6) continue;
                                                }
                                                break block93;
                                            }
                                            var10_14 = 0;
                                        }
                                        while (var10_14 < 2304) {
                                            this.f[var2_2][var10_14] = var7_9[var9_13++];
                                            ++var10_14;
                                            if (!var13_6) {
                                                if (!var13_6) continue;
                                            }
                                            break block94;
                                        }
                                        var10_14 = 0;
                                    }
                                    while (var10_14 < 2304) {
                                        this.P[var2_2][var10_14] = var7_9[var9_13++];
                                        ++var10_14;
                                        if (!var13_6) {
                                            if (!var13_6) continue;
                                        }
                                        break block95;
                                    }
                                    var10_14 = 0;
                                }
                                while (2304 > var10_14) {
                                    this.s[var2_2][var10_14] = ib.a(255, var7_9[1 + var9_13]) + ib.a(var7_9[var9_13], 255) * 256;
                                    var9_13 += 2;
                                    ++var10_14;
                                    if (!var13_6) {
                                        if (!var13_6) continue;
                                    }
                                    break block96;
                                }
                                var10_14 = 0;
                            }
                            while (2304 > var10_14) {
                                this.A[var2_2][var10_14] = var7_9[var9_13++];
                                ++var10_14;
                                if (!var13_6) {
                                    if (!var13_6) continue;
                                }
                                break block97;
                            }
                            var10_14 = 0;
                        }
                        while (~var10_14 > -2305) {
                            this.R[var2_2][var10_14] = var7_9[var9_13++];
                            ++var10_14;
                            if (!var13_6) {
                                if (!var13_6) continue;
                            }
                            break block98;
                        }
                        var10_14 = 0;
                    }
                    while (2304 > var10_14) {
                        this.mb[var2_2][var10_14] = var7_9[var9_13++];
                        ++var10_14;
                        if (!var13_6) {
                            if (!var13_6) continue;
                        }
                        break block91;
                    }
                    break block91;
                }
                catch (IOException var7_10) {
                    var8_12 = 0;
                    while (~var8_12 > -2305) {
                        block101: {
                            block100: {
                                block99: {
                                    block114: {
                                        this.L[var2_2][var8_12] = 0;
                                        this.eb[var2_2][var8_12] = 0;
                                        this.f[var2_2][var8_12] = 0;
                                        this.P[var2_2][var8_12] = 0;
                                        this.s[var2_2][var8_12] = 0;
                                        this.A[var2_2][var8_12] = 0;
                                        this.R[var2_2][var8_12] = 0;
                                        if (var13_6) break block91;
                                        if (~var1_1 != -1) break block99;
                                        break block114;
                                        catch (IOException v14) {
                                            throw v14;
                                        }
                                    }
                                    this.R[var2_2][var8_12] = -6;
                                }
                                if (var1_1 == 3) break block100;
                                break block101;
                            }
                            this.R[var2_2][var8_12] = 8;
                        }
                        this.mb[var2_2][var8_12] = 0;
                        ++var8_12;
                        if (!var13_6) continue;
                    }
                    break block91;
                }
                {
                    break;
                }
            }
            catch (RuntimeException var6_8) {
                throw i.a(var6_8, k.ob[27] + var1_1 + ',' + var2_2 + ',' + var3_3 + ',' + var4_4 + ',' + var5_5 + ')');
            }
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final int b(int n2, int n3, int n4) {
        boolean bl = client.vh;
        try {
            ++ib;
            if (0 > n2 || ~n2 <= -97 || -1 < ~n3 || -97 >= ~n3) return 0;
            if (n4 > -68) {
                return -68;
            }
            int n5 = 0;
            if (~n2 > -49 || n3 >= 48) {
                if (-49 >= ~n2 || ~n3 > -49) {
                    if (48 > n2 || n3 < 48) return this.mb[n5][n3 + n2 * 48];
                    n2 -= 48;
                    n3 -= 48;
                    n5 = 3;
                    if (!bl) return this.mb[n5][n3 + n2 * 48];
                }
                n3 -= 48;
                n5 = 2;
                if (!bl) return this.mb[n5][n3 + n2 * 48];
            }
            n2 -= 48;
            n5 = 1;
            return this.mb[n5][n3 + n2 * 48];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[11] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, byte by, int n3, int n4, int n5, int n6) {
        try {
            ++Y;
            int n7 = n4 * 3;
            int n8 = n5 * 3;
            int n9 = this.c.a(n6, true);
            int n10 = -57 % ((by - -64) / 57);
            n9 = n9 >> 1085422433 & 0x7F7F7F;
            int n11 = this.c.a(n3, true);
            n11 = (0xFEFEFF & n11) >> 43967169;
            if (n2 == 0) {
                this.U.b(3, n9, n7, n8, (byte)109);
                this.U.b(2, n9, n7, 1 + n8, (byte)-65);
                this.U.b(1, n9, n7, n8 - -2, (byte)99);
                this.U.b(1, n11, 2 + n7, n8 - -1, (byte)73);
                this.U.b(2, n11, n7 + 1, n8 + 2, (byte)113);
                if (!client.vh) return;
            }
            if (1 != n2) return;
            this.U.b(3, n11, n7, n8, (byte)55);
            this.U.b(2, n11, 1 + n7, 1 + n8, (byte)62);
            this.U.b(1, n11, n7 + 2, n8 - -2, (byte)56);
            this.U.b(1, n9, n7, n8 + 1, (byte)70);
            this.U.b(2, n9, n7, 2 + n8, (byte)-85);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[36] + n2 + ',' + by + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final int b(byte by, int n2, int n3) {
        try {
            if (by != -38) {
                this.B = null;
            }
            ++d;
            if (0 <= n3 && ~n2 <= -1 && 96 > n3 && ~n2 > -97) {
                return this.bb[n3][n2];
            }
            return 0;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[33] + by + ',' + n2 + ',' + n3 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final int b(int n2, int n3, int n4, int n5) {
        boolean bl = client.vh;
        try {
            int n6;
            block20: {
                block18: {
                    block19: {
                        block16: {
                            block17: {
                                block15: {
                                    ++K;
                                    if (n3 < 0 || ~n3 <= -97 || n5 < 0 || -97 >= ~n5) break block15;
                                    n6 = 0;
                                    if (-49 < ~n3) break block16;
                                    break block17;
                                }
                                return 0;
                            }
                            if (n5 < 48) break block18;
                        }
                        if (48 <= n3 || 48 > n5) break block19;
                        n5 -= 48;
                        n6 = 2;
                        if (!bl) break block20;
                    }
                    if (-49 < ~n3 || n5 < 48) break block20;
                    n5 -= 48;
                    n3 -= 48;
                    n6 = 3;
                    if (!bl) break block20;
                }
                n3 -= 48;
                n6 = 1;
            }
            if (n4 != 4) {
                return -4;
            }
            return 0xFF & this.R[n6][48 * n3 + n5];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[23] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final int a(int n2, byte by, int n3) {
        boolean bl = client.vh;
        try {
            int n4;
            block17: {
                block18: {
                    block15: {
                        block16: {
                            block14: {
                                ++M;
                                if (0 > n2 || n2 >= 96 || n3 < 0 || n3 >= 96) break block14;
                                int n5 = -67 / ((by - -48) / 60);
                                n4 = 0;
                                if (-49 < ~n2) break block15;
                                break block16;
                            }
                            return 0;
                        }
                        if (48 <= n3) break block15;
                        n2 -= 48;
                        n4 = 1;
                        if (!bl) break block17;
                    }
                    if (-49 >= ~n2 || ~n3 > -49) break block18;
                    n3 -= 48;
                    n4 = 2;
                    if (!bl) break block17;
                }
                if (~n2 <= -49 && -49 >= ~n3) {
                    n3 -= 48;
                    n2 -= 48;
                    n4 = 3;
                }
            }
            return 0xFF & this.P[n4][n2 * 48 - -n3];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[13] + n2 + ',' + by + ',' + n3 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, ca ca2, int n3, int n4, int n5, int n6, int n7) {
        try {
            ++C;
            this.b(40, (byte)50, n5, n4);
            this.b(40, (byte)109, n3, n7);
            int n8 = ib.d[n2];
            int n9 = v.a[n2];
            if (n6 != -14584) {
                this.a((byte)-62, 104, -113);
            }
            int n10 = client.Jk[n2];
            int n11 = 128 * n5;
            int n12 = 128 * n4;
            int n13 = 128 * n3;
            int n14 = n7 * 128;
            int n15 = ca2.e(n11, n12, -this.ab[n5][n4], -111);
            int n16 = ca2.e(n11, n12, -this.ab[n5][n4] + -n8, -115);
            int n17 = ca2.e(n13, n14, -n8 + -this.ab[n3][n7], -125);
            int n18 = ca2.e(n13, n14, -this.ab[n3][n7], n6 ^ 0xFFFFC757);
            int[] nArray = new int[]{n15, n16, n17, n18};
            int n19 = ca2.a(4, nArray, n9, n10, false);
            if (-6 == ~lb.Tb[n2]) {
                ca2.E[n19] = 30000 + n2;
                if (!client.vh) return;
            }
            ca2.E[n19] = 0;
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(ob[15]).append(n2).append(',');
            if (ca2 != null) {
                string = ob[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(')').toString());
            }
            string = ob[2];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, int n3, int n4, int n5, int n6) {
        boolean bl = client.vh;
        try {
            block21: {
                block23: {
                    block22: {
                        block20: {
                            ++W;
                            if (-1 < ~n5) return;
                            if (~n2 > -1) return;
                            if (-96 >= ~n5) return;
                            if (n2 >= 95) return;
                            if (n6 != 11715) {
                                this.s = null;
                            }
                            if (u.a[n3] != 1) {
                                return;
                            }
                            if (-1 != ~n4) break block20;
                            this.bb[n5][n2] = d.a(this.bb[n5][n2], 1);
                            if (0 >= n2) break block21;
                            this.a(4, -1 + n2, (byte)-96, n5);
                            if (!bl) break block21;
                        }
                        if (1 != n4) break block22;
                        this.bb[n5][n2] = d.a(this.bb[n5][n2], 2);
                        if (0 >= n5) break block21;
                        this.a(8, n2, (byte)-89, -1 + n5);
                        if (!bl) break block21;
                    }
                    if (-3 == ~n4) break block23;
                    if (~n4 != -4) break block21;
                    this.bb[n5][n2] = d.a(this.bb[n5][n2], 32);
                    if (!bl) break block21;
                }
                this.bb[n5][n2] = d.a(this.bb[n5][n2], 16);
            }
            this.c(1, 1, 62, n5, n2);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[25] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final boolean a(boolean bl, int n2, int n3) {
        try {
            ++V;
            if (~this.d(n3, n2, 114) < -1 && this.d(n3, n2 - 1, 122) > 0 && -1 > ~this.d(-1 + n3, -1 + n2, 117) && ~this.d(-1 + n3, n2, 122) < -1) {
                return true;
            }
            return bl;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[14] + bl + ',' + n2 + ',' + n3 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final int e(int n2, int n3, int n4) {
        boolean bl = client.vh;
        try {
            ++b;
            int n5 = -120 / ((n2 - 15) / 43);
            if (n3 < 0) return 0;
            if (~n3 <= -97) return 0;
            if (-1 < ~n4) return 0;
            if (n4 >= 96) return 0;
            int n6 = 0;
            if (n3 < 48 || -49 >= ~n4) {
                if (-49 >= ~n3 || ~n4 > -49) {
                    if (n3 < 48) return 0xFF & this.f[n6][n3 * 48 + n4];
                    if (48 > n4) {
                        return 0xFF & this.f[n6][n3 * 48 + n4];
                    }
                    n4 -= 48;
                    n3 -= 48;
                    n6 = 3;
                    if (!bl) return 0xFF & this.f[n6][n3 * 48 + n4];
                }
                n6 = 2;
                n4 -= 48;
                if (!bl) return 0xFF & this.f[n6][n3 * 48 + n4];
            }
            n3 -= 48;
            n6 = 1;
            return 0xFF & this.f[n6][n3 * 48 + n4];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, ob[21] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    k(lb lb2, ua ua2) {
        boolean bl = client.vh;
        this.H = false;
        this.ab = new int[96][96];
        this.mb = new byte[4][2304];
        this.P = new byte[4][2304];
        this.R = new byte[4][2304];
        this.w = new int[256];
        this.B = new int[96][96];
        this.f = new byte[4][2304];
        this.E = new int[18432];
        this.nb = true;
        this.x = 750;
        this.g = new ca[4][64];
        this.L = new byte[4][2304];
        this.F = new ca[64];
        this.s = new int[4][2304];
        this.A = new byte[4][2304];
        this.Z = false;
        this.bb = new int[96][96];
        this.db = new ca[4][64];
        this.eb = new byte[4][2304];
        this.q = new int[18432];
        try {
            int n2;
            block21: {
                block20: {
                    block19: {
                        this.U = ua2;
                        this.c = lb2;
                        for (n2 = 0; n2 < 64; ++n2) {
                            this.w[n2] = da.a(255 + -(n2 * 4), (byte)-66, -(4 * n2) + 255, 255 - (int)((double)n2 * 1.75));
                            if (!bl) {
                                if (!bl) continue;
                            }
                            break block19;
                        }
                        n2 = 0;
                    }
                    while (~n2 > -65) {
                        this.w[64 + n2] = da.a(0, (byte)-66, 3 * n2, 144);
                        ++n2;
                        if (!bl) {
                            if (!bl) continue;
                        }
                        break block20;
                    }
                    n2 = 0;
                }
                while (n2 < 64) {
                    this.w[128 + n2] = da.a(0, (byte)-66, 192 - (int)((double)n2 * 1.5), -((int)((double)n2 * 1.5)) + 144);
                    ++n2;
                    if (!bl) {
                        if (!bl) continue;
                    }
                    break block21;
                }
                n2 = 0;
            }
            do {
                if (64 <= n2) return;
                this.w[192 + n2] = da.a(0, (byte)-66, 96 - (int)((double)n2 * 1.5), (int)((double)n2 * 1.5) + 48);
                ++n2;
                if (bl) return;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(ob[3]).append(lb2 != null ? ob[1] : ob[2]).append(',');
            if (ua2 != null) {
                string = ob[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = ob[2];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    static {
        ob = new String[]{k.z(k.z("\u001e\f\u0017r3")), k.z(k.z("\u000e\fz\u001df")), k.z(k.z("\u001bW8_")), k.z(k.z("\u001e\fhZu\u001cVj\u001b")), k.z(k.z("\u001e\f\u001cr3")), k.z(k.z("\u001e\f\u0012r3")), k.z(k.z("\u001e\f\u0016\u001b")), k.z(k.z("\u001e\f\u0007\u001b")), k.z(k.z("\u001e\f\u0015\u001b")), k.z(k.z("\u001e\f\u0010r3")), k.z(k.z("\u001e\f\u001dr3")), k.z(k.z("\u001e\f\u0016r3")), k.z(k.z("\u001e\f\u0013\u001b")), k.z(k.z("\u001e\f\u0006\u001b")), k.z(k.z("\u001e\f\u0011r3")), k.z(k.z("\u001e\f\u0012\u001b")), k.z(k.z("\u001e\f\u0019\u001b")), k.z(k.z("\u001e\f\u0004\u001b")), k.z(k.z(" L$Rx\u001eK:T;")), k.z(k.z("\u001e\f\u001f\u001b")), k.z(k.z("\u001e\f\u0005\u001b")), k.z(k.z("\u001e\f\u0018r3")), k.z(k.z("\u001e\f\u0002\u001b")), k.z(k.z("\u001e\f\u001e\u001b")), k.z(k.z("\u001e\f\u0001\u001b")), k.z(k.z("\u001e\f\u0003\u001b")), k.z(k.z("[J1Z")), k.z(k.z("\u001e\f\u001c\u001b")), k.z(k.z("[\f{Tz\u0018G0Ro\u0014\r9Rk\u0006\r")), k.z(k.z("[F5G")), k.z(k.z("[N;P")), k.z(k.z("[H9")), k.z(k.z("\u001e\f\u001a\u001b")), k.z(k.z("\u001e\f\u001er3")), k.z(k.z("\u001e\f\u0013r3")), k.z(k.z("\u001e\f\u0015r3")), k.z(k.z("\u001e\f\u0017\u001b")), k.z(k.z("\u001e\f\u001fr3")), k.z(k.z("\u001e\f\u001b\u001b")), k.z(k.z("\u001e\f\u001d\u001b")), k.z(k.z("\u001bW8_;\u0007M;U:")), k.z(k.z("\u001e\f\u0018\u001b")), k.z(k.z("\u001e\f\u0000\u001b")), k.z(k.z("\u001e\f\u0010\u001b")), k.z(k.z("\u001e\f\u0011\u001b"))};
        G = new String[100];
        e = 0L;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x1B);
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
                        n4 = 117;
                        break;
                    }
                    case 1: {
                        n4 = 34;
                        break;
                    }
                    case 2: {
                        n4 = 84;
                        break;
                    }
                    case 3: {
                        n4 = 51;
                        break;
                    }
                    default: {
                        n4 = 27;
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

