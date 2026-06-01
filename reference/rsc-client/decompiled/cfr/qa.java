/*
 * Decompiled with CFR 0.152.
 */
import java.awt.Font;
import java.awt.FontMetrics;

final class qa {
    private String[][] Ab;
    static int s;
    static int P;
    static int L;
    private int[] N;
    static int ub;
    int[] j;
    static int x;
    private int hb = 0;
    static int Gb;
    static int M;
    static int lb;
    static int r;
    static int nb;
    static int db;
    private String[] yb;
    static int mb;
    private String[][] p;
    private String[][] Fb;
    private int R;
    static int xb;
    private int gb = -1;
    static int[] K;
    private int i;
    private int tb;
    static int rb;
    private int[] O;
    static int H;
    private int G = 0;
    static int y;
    static int a;
    static int o;
    static int wb;
    static int u;
    private int[] kb;
    private ua w;
    private int eb = 0;
    private boolean[] D;
    static int ab;
    static byte[] l;
    static int z;
    private int Eb;
    private int[] ob;
    private int ib;
    static int h;
    static int X;
    static int Db;
    static int e;
    private int[] B;
    private int F;
    private int[] sb;
    private boolean[] Y;
    static int n;
    static int q;
    private boolean t = true;
    static int V;
    static int Bb;
    static int v;
    static int jb;
    static int b;
    private int f;
    private int zb = 0;
    private int E;
    private int bb = 0;
    private int Hb = 0;
    private int[] k;
    static int I;
    private int C;
    private int[] vb;
    static int Z;
    private boolean[] d;
    private boolean[] cb;
    static int T;
    private int J;
    static int Q;
    private boolean[] g;
    static int Cb;
    static int c;
    private int fb;
    private int[] U;
    int[] pb;
    static int S;
    private int[][] m;
    static int A;
    private int qb;
    private static final String[] W;

    final void c(byte by, int n2) {
        try {
            int n3 = 60 / ((19 - by) / 44);
            this.pb[n2] = 0;
            ++r;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[47] + by + ',' + n2 + ')');
        }
    }

    final int a(int n2, int n3, boolean bl, int n4, int n5, int n6, int n7, int n8, boolean bl2) {
        try {
            ++h;
            this.U[this.eb] = 5;
            this.g[this.eb] = true;
            this.cb[this.eb] = bl;
            if (n7 != 14179) {
                return 2;
            }
            this.D[this.eb] = false;
            this.k[this.eb] = n5;
            this.Y[this.eb] = bl2;
            this.kb[this.eb] = n4;
            this.B[this.eb] = n6;
            this.ob[this.eb] = n8;
            this.O[this.eb] = n3;
            this.sb[this.eb] = n2;
            this.yb[this.eb] = "";
            return this.eb++;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[16] + n2 + ',' + n3 + ',' + bl + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ',' + n8 + ',' + bl2 + ')');
        }
    }

    private final int a(int n2, int n3, int n4, byte by) {
        try {
            ++y;
            if (by >= -70) {
                this.kb = null;
            }
            return o.a(ua.C * n4 / 114, 9570, ta.g * n2 / 176, n3 * aa.d / 114);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[5] + n2 + ',' + n3 + ',' + n4 + ',' + by + ')');
        }
    }

    final int d(int n2, int n3, int n4, int n5, int n6) {
        try {
            this.U[this.eb] = 10;
            ++x;
            this.g[this.eb] = true;
            this.D[this.eb] = false;
            int n7 = 118 / ((n5 - -22) / 48);
            this.kb[this.eb] = -(n3 / 2) + n2;
            this.B[this.eb] = -(n6 / 2) + n4;
            this.ob[this.eb] = n3;
            this.O[this.eb] = n6;
            return this.eb++;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[13] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(byte by) {
        boolean bl = client.vh;
        try {
            ++Gb;
            if (by > 0) {
                this.R = -121;
            }
            for (int i2 = 0; i2 < this.eb; ++i2) {
                if (bl) return;
                if (!this.g[i2]) continue;
                if (this.U[i2] != 0) {
                    if (1 == this.U[i2]) {
                        this.a(this.yb[i2], this.B[i2], (byte)117, this.k[i2], this.kb[i2] - this.w.a(this.k[i2], 108, this.yb[i2]) / 2, i2, 0);
                        if (!bl) continue;
                    }
                    if (this.U[i2] == 2) {
                        this.a(this.ob[i2], (byte)69, this.O[i2], this.kb[i2], this.B[i2]);
                        if (!bl) continue;
                    }
                    if (3 != this.U[i2]) {
                        if (4 != this.U[i2]) {
                            if (5 == this.U[i2] || this.U[i2] == 6) {
                                this.a(this.k[i2], this.yb[i2], this.ob[i2], this.O[i2], true, this.kb[i2], this.B[i2], i2);
                                if (!bl) continue;
                            }
                            if (this.U[i2] != 7) {
                                if (this.U[i2] == 8) {
                                    this.a(this.Ab[i2], i2, -121, this.k[i2], this.kb[i2], this.B[i2]);
                                    if (!bl) continue;
                                }
                                if (-10 == ~this.U[i2]) {
                                    this.a(this.O[i2], this.kb[i2], i2, this.Ab[i2], this.pb[i2], this.k[i2], this.B[i2], this.m[i2], 0, this.j[i2], this.ob[i2]);
                                    if (!bl) continue;
                                }
                                if (this.U[i2] != 11) {
                                    if (~this.U[i2] != -13) {
                                        if (14 != this.U[i2]) continue;
                                        this.a((byte)52, this.kb[i2], this.O[i2], this.B[i2], this.ob[i2], i2);
                                        if (!bl) continue;
                                    }
                                    this.b(-82, this.B[i2], this.kb[i2], this.k[i2]);
                                    if (!bl) continue;
                                }
                                this.a(this.B[i2], this.ob[i2], true, this.kb[i2], this.O[i2]);
                                if (!bl) continue;
                            }
                            this.a(this.kb[i2], this.k[i2], (byte)-73, this.Ab[i2], i2, this.B[i2]);
                            if (!bl) continue;
                        }
                        this.a(this.m[i2], this.j[i2], i2, this.kb[i2], this.B[i2], this.ob[i2], this.O[i2], this.pb[i2], this.Ab[i2], this.k[i2], false);
                        if (!bl) continue;
                    }
                    this.a(this.B[i2], 0, this.kb[i2], this.ob[i2]);
                    if (!bl) continue;
                }
                this.a(this.yb[i2], this.B[i2], (byte)-65, this.k[i2], this.kb[i2], i2, 0);
                if (!bl) continue;
            }
            this.zb = 0;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[11] + by + ')');
        }
    }

    final int a(boolean bl, byte by, int n2, int n3, String string, int n4) {
        try {
            this.U[this.eb] = 1;
            if (by >= -71) {
                this.a(-106, null, 31, 96, (String)null, null, 36);
            }
            ++A;
            this.g[this.eb] = true;
            this.D[this.eb] = false;
            this.k[this.eb] = n2;
            this.Y[this.eb] = bl;
            this.kb[this.eb] = n3;
            this.B[this.eb] = n4;
            this.yb[this.eb] = string;
            return this.eb++;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(W[28]).append(bl).append(',').append(by).append(',').append(n2).append(',').append(n3).append(',');
            String string2 = string != null ? W[2] : W[1];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n4).append(')').toString());
        }
    }

    final int a(int n2, int n3, int n4, int n5, int n6, int n7, int n8, boolean bl) {
        try {
            ++v;
            this.U[this.eb] = 4;
            this.g[this.eb] = true;
            this.D[this.eb] = false;
            this.kb[this.eb] = n4;
            this.B[this.eb] = n6;
            this.ob[this.eb] = n2;
            this.O[this.eb] = n3;
            this.Y[this.eb] = bl;
            this.k[this.eb] = n7;
            this.sb[this.eb] = n5;
            this.pb[this.eb] = 0;
            this.j[this.eb] = 0;
            this.Ab[this.eb] = new String[n5];
            if (n8 != 63) {
                this.a(39, 117, 98, -16);
            }
            this.m[this.eb] = new int[n5];
            this.Fb[this.eb] = new String[n5];
            this.p[this.eb] = new String[n5];
            return this.eb++;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[45] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ',' + n8 + ',' + bl + ')');
        }
    }

    final String a(int n2, int n3, int n4) {
        try {
            if (n3 >= -90) {
                this.d = null;
            }
            ++X;
            return this.p[n4][n2];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[36] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    final void e(int n2, int n3) {
        try {
            if (n3 != 14) {
                this.d(-43, 43, 111, -105, -59);
            }
            ++ab;
            this.j[n2] = 0;
            this.N[n2] = -1;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[33] + n2 + ',' + n3 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int[] nArray, int n2, int n3, int n4, int n5, int n6, int n7, int n8, String[] stringArray, int n9, boolean bl) {
        boolean bl2 = client.vh;
        try {
            int n10;
            int n11;
            int n12;
            int n13;
            block49: {
                block51: {
                    block50: {
                        ++o;
                        n13 = n7 / this.w.a(508305352, n9);
                        if (n2 > -n13 + n8) {
                            n2 = n8 - n13;
                        }
                        if (~n2 > -1) {
                            n2 = 0;
                        }
                        if (bl) {
                            this.a(null, -61, -120, -114, 65, -8);
                        }
                        this.j[n3] = n2;
                        if (n8 <= n13) break block49;
                        n12 = -12 + n6 + n4;
                        n11 = n13 * (n7 + -27) / n8;
                        if (6 > n11) {
                            n11 = 6;
                        }
                        n10 = (-n11 + -27 + n7) * n2 / (-n13 + n8);
                        if (-2 == ~this.Hb && ~this.bb <= ~n12 && ~(12 + n12) <= ~this.bb) {
                            if (this.hb > n5 && ~this.hb > ~(n5 - -12) && n2 > 0) {
                                --n2;
                            }
                            if (-12 + (n7 + n5) >= this.hb || this.hb >= n7 + n5 || ~n2 > ~(n8 - n13)) {
                                // empty if block
                            }
                            this.j[n3] = ++n2;
                        }
                        if (~this.Hb == -2 && (~n12 >= ~this.bb && ~(n12 + 12) <= ~this.bb || this.bb >= -12 + n12 && ~this.bb >= ~(24 + n12) && this.d[n3])) break block50;
                        this.d[n3] = false;
                        if (!bl2) break block51;
                    }
                    if (this.hb > 12 + n5 && ~(n5 - -n7 - 12) < ~this.hb) {
                        this.d[n3] = true;
                        int n14 = -(n11 / 2) + this.hb - n5 + -12;
                        n2 = n14 * n8 / (n7 + -24);
                        if (-n13 + n8 < n2) {
                            n2 = n8 + -n13;
                        }
                        if (0 > n2) {
                            n2 = 0;
                        }
                        this.j[n3] = n2;
                    }
                }
                n10 = n2 * (-n11 + (-27 + n7)) / (n8 + -n13);
                this.a(n11, n10, n4, n6, n7, n5, (byte)113);
            }
            n12 = -(this.w.a(508305352, n9) * n13) + n7;
            n11 = n5 + (5 * this.w.a(508305352, n9) / 6 + n12 / 2);
            n10 = n2;
            do {
                if (n8 <= n10) return;
                if (bl2) return;
                if (this.zb != 0 && 2 + n4 <= this.bb && ~this.bb >= ~(this.w.a(n9, 97, stringArray[n10]) + n4 - -2) && n11 >= -2 + this.hb && this.hb - 2 > -this.w.a(508305352, n9) + n11) {
                    this.D[n3] = true;
                    this.vb[n3] = d.a(this.zb << -259102544, n10);
                }
                this.a(n9, n3, true, nArray[n10], 2 + n4, n11, stringArray[n10]);
                if ((n11 += this.w.a(508305352, n9) + -ia.i) >= n5 - -n7) return;
                ++n10;
            } while (!bl2);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(W[37]).append(nArray != null ? W[2] : W[1]).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',');
            if (stringArray != null) {
                string = W[2];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n9).append(',').append(bl).append(')').toString());
            }
            string = W[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n9).append(',').append(bl).append(')').toString());
        }
    }

    private final void a(int n2, int n3, boolean bl, int n4, int n5, int n6, String string) {
        try {
            int n7;
            block9: {
                block8: {
                    ++Cb;
                    if (!bl) {
                        this.a(-3, 44, (byte)37, null, -78, 86);
                    }
                    if (this.Y[n3]) break block8;
                    n7 = 0;
                    if (!client.vh) break block9;
                }
                n7 = 0xFFFFFF;
            }
            this.w.a(n4, n6, string, n5, n7, (byte)-54, n2);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(W[15]).append(n2).append(',').append(n3).append(',').append(bl).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',');
            String string2 = string != null ? W[2] : W[1];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(')').toString());
        }
    }

    final void a(int n2, String string, int n3) {
        try {
            if (n3 != 27642) {
                return;
            }
            this.yb[n2] = string;
            ++mb;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(W[48]).append(n2).append(',');
            String string2 = string != null ? W[2] : W[1];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n3).append(')').toString());
        }
    }

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    private final void a(byte by, int n2, int n3, int n4, int n5, int n6) {
        boolean bl = client.vh;
        try {
            this.w.a(n2, (byte)100, 0xFFFFFF, n4, n3, n5);
            if (by != 52) {
                return;
            }
            ++Db;
            this.w.b(n5, this.tb, n2, n4, (byte)-116);
            this.w.b(n2, n4, this.tb, n3, 0);
            this.w.b(n5, this.J, n2, -1 + (n3 + n4), (byte)-124);
            this.w.b(n5 + n2 - 1, n4, this.J, n3, by ^ 0x34);
            if (this.vb[n6] != 1) return;
            int n7 = 0;
            while (~n3 < ~n7) {
                try {
                    this.w.b(1, 0, n7 + n2, n4 + n7, (byte)88);
                    this.w.b(1, 0, n5 + n2 + (-1 - n7), n7 + n4, (byte)106);
                    ++n7;
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
            throw i.a(runtimeException, W[39] + by + ',' + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void b(int n2, int n3, int n4, int n5, int n6) {
        boolean bl = client.vh;
        try {
            block43: {
                qa qa2;
                int n7;
                block45: {
                    block44: {
                        int n8;
                        this.hb = n3;
                        this.bb = n6;
                        if (-1 != ~n5) {
                            this.zb = n5;
                        }
                        this.Hb = n2;
                        ++xb;
                        if (-2 == ~n5) {
                            n7 = 0;
                            while (~this.eb < ~n7) {
                                n8 = this.g[n7];
                                if (!bl) {
                                    if (n8 != 0 && 10 == this.U[n7] && this.kb[n7] <= this.bb && this.hb >= this.B[n7] && this.ob[n7] + this.kb[n7] >= this.bb && this.hb <= this.O[n7] + this.B[n7]) {
                                        this.D[n7] = true;
                                    }
                                    if (this.g[n7] && this.U[n7] == 14 && this.bb >= this.kb[n7] && this.hb >= this.B[n7] && this.ob[n7] + this.kb[n7] >= this.bb && this.O[n7] + this.B[n7] >= this.hb) {
                                        this.vb[n7] = 1 - this.vb[n7];
                                    }
                                    ++n7;
                                    if (!bl) continue;
                                }
                                break;
                            }
                        } else {
                            n8 = n4;
                        }
                        if (n8 != -9989) {
                            this.a(-43, 104);
                        }
                        if (1 != n2) break block44;
                        ++this.G;
                        if (!bl) break block45;
                    }
                    this.G = 0;
                }
                if (-2 != ~n5 && ~this.G >= -21) {
                    return;
                }
                for (n7 = 0; this.eb > n7; ++n7) {
                    qa2 = this;
                    if (!bl) {
                        if (!qa2.g[n7] || this.U[n7] != 15 || ~this.kb[n7] < ~this.bb || this.hb < this.B[n7] || this.ob[n7] + this.kb[n7] < this.bb || ~(this.B[n7] + this.O[n7]) > ~this.hb) continue;
                        this.D[n7] = true;
                        if (!bl) continue;
                    }
                    break block43;
                }
                qa2 = this;
            }
            qa2.G -= 5;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[12] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    final int a(int n2, int n3, int n4, int n5, int n6) {
        try {
            ++P;
            this.U[this.eb] = 11;
            if (n5 != 26531) {
                return 59;
            }
            this.g[this.eb] = true;
            this.D[this.eb] = false;
            this.kb[this.eb] = -(n4 / 2) + n3;
            this.B[this.eb] = -(n2 / 2) + n6;
            this.ob[this.eb] = n4;
            this.O[this.eb] = n2;
            return this.eb++;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[0] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int n3, int n4, String[] stringArray, int n5, int n6, int n7, int[] nArray, int n8, int n9, int n10) {
        boolean bl = client.vh;
        try {
            int n11;
            int n12;
            int n13;
            int n14;
            int n15;
            block54: {
                block56: {
                    block55: {
                        block53: {
                            ++db;
                            if (n8 != 0) {
                                this.a(56, -19);
                            }
                            if (~(n15 = n2 / this.w.a(n8 + 508305352, n6)) > ~n5) break block53;
                            n9 = 0;
                            this.j[n4] = 0;
                            if (!bl) break block54;
                        }
                        n14 = -12 + n10 + n3;
                        n13 = n15 * (n2 + -27) / n5;
                        if (-7 < ~n13) {
                            n13 = 6;
                        }
                        n12 = (-n13 + (n2 + -27)) * n9 / (n5 - n15);
                        if (-2 == ~this.Hb && this.bb >= n14 && ~this.bb >= ~(12 + n14)) {
                            if (n7 < this.hb && ~(12 + n7) < ~this.hb && n9 > 0) {
                                --n9;
                            }
                            if (this.hb <= -12 + n2 + n7 || ~(n7 + n2) >= ~this.hb || ~(-n15 + n5) < ~n9) {
                                // empty if block
                            }
                            this.j[n4] = ++n9;
                        }
                        if (1 != this.Hb || (~n14 < ~this.bb || ~this.bb < ~(12 + n14)) && (~(-12 + n14) < ~this.bb || ~this.bb < ~(n14 + 24) || !this.d[n4])) break block55;
                        if (~(12 + n7) <= ~this.hb || n7 - (-n2 - -12) <= this.hb) break block56;
                        this.d[n4] = true;
                        n11 = -12 + this.hb + (-n7 - n13 / 2);
                        n9 = n5 * n11 / (n2 - 24);
                        if (-1 < ~n9) {
                            n9 = 0;
                        }
                        if (~n9 < ~(n5 - n15)) {
                            n9 = n5 + -n15;
                        }
                        this.j[n4] = n9;
                        if (!bl) break block56;
                    }
                    this.d[n4] = false;
                }
                n12 = (-27 + (n2 - n13)) * n9 / (-n15 + n5);
                this.a(n13, n12, n3, n10, n2, n7, (byte)-25);
            }
            this.N[n4] = -1;
            n14 = n2 + -(this.w.a(508305352, n6) * n15);
            n13 = 5 * this.w.a(508305352, n6) / 6 + (n7 - -(n14 / 2));
            n12 = n9;
            do {
                block59: {
                    block61: {
                        block60: {
                            block58: {
                                block57: {
                                    if (n12 >= n5) return;
                                    if (bl) return;
                                    if (!this.Y[n4]) break block57;
                                    n11 = 0xFFFFFF;
                                    if (!bl) break block58;
                                }
                                n11 = 0;
                            }
                            if (this.bb < 2 + n3 || this.bb > this.w.a(n6, 95, stringArray[n12]) + (2 + n3) || -2 + this.hb > n13 || -2 + this.hb <= -this.w.a(508305352, n6) + n13) break block59;
                            if (!this.Y[n4]) break block60;
                            n11 = 0x808080;
                            if (!bl) break block61;
                        }
                        n11 = 0xFFFFFF;
                    }
                    this.N[n4] = n12;
                    if (this.zb == 1) {
                        this.vb[n4] = n12;
                        this.D[n4] = true;
                    }
                }
                if (~this.vb[n4] == ~n12 && this.t) {
                    n11 = 0xFF0000;
                }
                this.w.a(nArray[n12], n13, stringArray[n12], n3 - -2, n11, (byte)-9, n6);
                if ((n13 += this.w.a(508305352, n6)) >= n2 + n7) {
                    if (!bl) return;
                }
                ++n12;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(W[35]).append(n2).append(',').append(n3).append(',').append(n4).append(',').append(stringArray != null ? W[2] : W[1]).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',');
            if (nArray != null) {
                string = W[2];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(')').toString());
            }
            string = W[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, String string, int n3, int n4, boolean bl, int n5, int n6, int n7) {
        boolean bl2 = client.vh;
        try {
            int n8;
            block32: {
                block31: {
                    ++wb;
                    if (!(!this.cb[n7])) {
                        n8 = string.length();
                        string = "";
                        int n9 = 0;
                        while (~n8 < ~n9) {
                            string = string + "X";
                            ++n9;
                            if (!bl2) {
                                if (!bl2) continue;
                            }
                            break block31;
                        }
                    }
                    if (~this.U[n7] != -6) break block31;
                    if (this.zb != 1 || n5 > this.bb || ~this.hb > ~(-(n4 / 2) + n6) || this.bb > n5 - -n3 || ~this.hb < ~(n4 / 2 + n6)) break block32;
                    this.gb = n7;
                    if (!bl2) break block32;
                }
                if (-7 == ~this.U[n7]) {
                    if (1 == this.zb && ~this.bb <= ~(-(n3 / 2) + n5) && ~(-(n4 / 2) + n6) >= ~this.hb && n3 / 2 + n5 >= this.bb && ~(n6 + n4 / 2) <= ~this.hb) {
                        this.gb = n7;
                    }
                    n5 -= this.w.a(n2, 76, string) / 2;
                }
            }
            if (this.gb == n7) {
                string = string + "*";
            }
            if (!bl) {
                this.a(28, (byte)94, -2, 23, 126);
            }
            n8 = this.w.a(508305352, n2) / 3 + n6;
            this.a(n2, n7, bl, 0, n5, n8, string);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string2;
            StringBuilder stringBuilder = new StringBuilder().append(W[3]).append(n2).append(',');
            if (string != null) {
                string2 = W[2];
                throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(n3).append(',').append(n4).append(',').append(bl).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(')').toString());
            }
            string2 = W[1];
            throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(n3).append(',').append(n4).append(',').append(bl).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(')').toString());
        }
    }

    final boolean a(byte by, int n2) {
        try {
            ++V;
            if (by > -95) {
                return true;
            }
            if (this.g[n2] && this.D[n2]) {
                this.D[n2] = false;
                return true;
            }
            return false;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[25] + by + ',' + n2 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final boolean a(e e2, String string, int n2, int n3) {
        boolean bl = client.vh;
        try {
            int n4;
            boolean bl2;
            block27: {
                int n5;
                int n6;
                block26: {
                    int n7;
                    ++Bb;
                    bl2 = false;
                    string = string.toLowerCase();
                    boolean bl3 = false;
                    if (string.startsWith(W[32])) {
                        string = string.substring(9);
                    }
                    if (string.startsWith("h")) {
                        string = string.substring(1);
                    }
                    if (string.startsWith("f")) {
                        string = string.substring(1);
                        bl2 = true;
                    }
                    if (string.startsWith("d")) {
                        string = string.substring(1);
                        bl3 = true;
                    }
                    if (string.endsWith(W[29])) {
                        string = string.substring(0, -3 + string.length());
                    }
                    n6 = 0;
                    if (string.endsWith("b")) {
                        n6 = 1;
                        string = string.substring(0, string.length() + -1);
                    }
                    if (string.endsWith("p")) {
                        string = string.substring(0, -1 + string.length());
                    }
                    n4 = Integer.parseInt(string);
                    Font font = new Font(W[31], n6, n4);
                    FontMetrics fontMetrics = e2.getFontMetrics(font);
                    String string2 = W[7];
                    b.c = 855;
                    n5 = 0;
                    while (-96 < ~n5) {
                        n7 = s.a(n2, font, n5, -95, e2, string2.charAt(n5), fontMetrics, bl3) ? 1 : 0;
                        if (!bl) {
                            if (n7 == 0) return false;
                            ++n5;
                            if (!bl) continue;
                        }
                        break block26;
                    }
                    m.b[n2] = new byte[b.c];
                    n7 = n5 = n3;
                }
                while (~n5 > ~b.c) {
                    m.b[n2][n5] = qb.k[n5];
                    ++n5;
                    if (!bl) {
                        if (!bl) continue;
                    }
                    break block27;
                }
                if (1 == n6 && fb.k[n2]) {
                    fb.k[n2] = false;
                    if (!qa.a(e2, "f" + n4 + "p", n2, 0)) return false;
                }
            }
            if (!bl2) return true;
            if (fb.k[n2]) return true;
            fb.k[n2] = false;
            if (!qa.a(e2, "d" + n4 + "p", n2, n3 ^ 0)) return false;
            return true;
        }
        catch (RuntimeException runtimeException) {
            String string2;
            StringBuilder stringBuilder = new StringBuilder().append(W[30]).append(e2 != null ? W[2] : W[1]).append(',');
            if (string != null) {
                string2 = W[2];
                throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(n2).append(',').append(n3).append(')').toString());
            }
            string2 = W[1];
            throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(n2).append(',').append(n3).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int n3, byte by, String[] stringArray, int n4, int n5) {
        boolean bl = client.vh;
        try {
            int n6;
            int n7;
            int n8;
            int n9;
            block29: {
                ++M;
                int n10 = 0;
                n9 = stringArray.length;
                if (by != -73) {
                    return;
                }
                for (n8 = 0; n9 > n8; ++n8) {
                    n10 += this.w.a(n3, 106, stringArray[n8]);
                    n7 = ~(n9 - 1);
                    n6 = ~n8;
                    if (!bl) {
                        if (n7 >= n6) continue;
                        n10 += this.w.a(n3, 92, W[20]);
                        if (!bl) continue;
                    }
                    break block29;
                }
                n8 = n2 + -(n10 / 2);
                n7 = this.w.a(by + 508305425, n3) / 3;
                n6 = n5;
            }
            int n11 = n7 + n6;
            int i2 = 0;
            do {
                int n12;
                block35: {
                    block36: {
                        block32: {
                            block34: {
                                block33: {
                                    block31: {
                                        block30: {
                                            if (n9 <= i2) return;
                                            if (bl) return;
                                            if (this.Y[n4]) break block30;
                                            n12 = 0;
                                            if (!bl) break block31;
                                        }
                                        n12 = 0xFFFFFF;
                                    }
                                    if (~n8 < ~this.bb || ~this.bb < ~(n8 + this.w.a(n3, 73, stringArray[i2])) || this.hb > n11 || ~this.hb >= ~(n11 + -this.w.a(508305352, n3))) break block32;
                                    if (!this.Y[n4]) break block33;
                                    n12 = 0x808080;
                                    if (!bl) break block34;
                                }
                                n12 = 0xFFFFFF;
                            }
                            if (this.zb == 1) {
                                this.vb[n4] = i2;
                                this.D[n4] = true;
                            }
                        }
                        if (~this.vb[n4] != ~i2) break block35;
                        if (!this.Y[n4]) break block36;
                        n12 = 0xFF0000;
                        if (!bl) break block35;
                    }
                    n12 = 0xC00000;
                }
                this.w.a(0, n11, stringArray[i2], n8, n12, (byte)-53, n3);
                n8 += this.w.a(n3, 127, stringArray[i2] + W[20]);
                ++i2;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(W[19]).append(n2).append(',').append(n3).append(',').append(by).append(',');
            if (stringArray != null) {
                string = W[2];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n4).append(',').append(n5).append(')').toString());
            }
            string = W[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n4).append(',').append(n5).append(')').toString());
        }
    }

    final void d(int n2, int n3) {
        try {
            this.gb = n2;
            ++T;
            if (n3 > -70) {
                this.b(33, -60, -29);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[42] + n2 + ',' + n3 + ')');
        }
    }

    private final void a(String string, int n2, byte by, int n3, int n4, int n5, int n6) {
        try {
            ++Z;
            int n7 = -109 / ((by - 14) / 62);
            int n8 = n2 - -(this.w.a(508305352, n3) / 3);
            this.a(n3, n5, true, n6, n4, n8, string);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(W[34]);
            String string2 = string != null ? W[2] : W[1];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n2).append(',').append(by).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(')').toString());
        }
    }

    final int c(int n2, int n3, int n4, int n5, int n6) {
        try {
            ++s;
            this.U[this.eb] = 2;
            if (n2 > -56) {
                this.a(127, 127);
            }
            this.g[this.eb] = true;
            this.D[this.eb] = false;
            this.kb[this.eb] = n5 - n3 / 2;
            this.B[this.eb] = n6 + -(n4 / 2);
            this.ob[this.eb] = n3;
            this.O[this.eb] = n4;
            return this.eb++;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[4] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    private final void a(int n2, byte n3, int n4, int n5, int n6) {
        boolean bl = client.vh;
        try {
            int n7;
            block11: {
                this.w.a(n5, n5 - -n2, n6 + n4, n6, (byte)30);
                ++b;
                this.w.b(n5, this.J, n2, this.tb, n4, n6, 19020);
                if (p.d) {
                    int n8 = -(n6 & 0x3F) + n5;
                    while (n8 < n2 + n5) {
                        block12: {
                            n7 = -(0x1F & n6) + n6;
                            if (bl) break block11;
                            for (int i2 = v564941; n6 - -n4 > i2; i2 += 128) {
                                this.w.a(6 - -u.g, 0, n8, 128, i2);
                                if (!bl) {
                                    if (!bl) continue;
                                    break;
                                }
                                break block12;
                            }
                            n8 += 128;
                        }
                        if (!bl) continue;
                    }
                }
                this.w.b(n2, this.tb, n5, n6, (byte)111);
                n7 = n3;
            }
            if (n7 != 69) {
                l = null;
            }
            this.w.b(n2 - 2, this.tb, 1 + n5, n6 - -1, (byte)-124);
            this.w.b(-4 + n2, this.F, 2 + n5, 2 + n6, (byte)-99);
            this.w.b(n5, n6, this.tb, n4, 0);
            this.w.b(1 + n5, n6 - -1, this.tb, -2 + n4, 0);
            this.w.b(n5 - -2, n6 + 2, this.F, -4 + n4, n3 ^ 0x45);
            this.w.b(n2, this.J, n5, n6 + (n4 - 1), (byte)100);
            this.w.b(n2 - 2, this.J, n5 - -1, n6 + (n4 - 2), (byte)103);
            this.w.b(-4 + n2, this.Eb, 2 + n5, -3 + (n6 - -n4), (byte)73);
            this.w.b(-1 + (n2 + n5), n6, this.J, n4, 0);
            this.w.b(n2 + n5 - 2, n6 + 1, this.J, n4 + -2, 0);
            this.w.b(n5 + (n2 - 3), n6 - -2, this.Eb, n4 - 4, 0);
            this.w.a(-1);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[8] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    private final void b(int n2, int n3, int n4, int n5) {
        try {
            this.w.b(-1, n5, n3, n4);
            int n6 = 71 / ((n2 - 67) / 32);
            ++I;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[44] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    final void c(int n2, int n3) {
        try {
            ++a;
            this.g[n2] = true;
            if (n3 < 114) {
                this.a(7, -8, false, -58, -39, 47, null);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[18] + n2 + ',' + n3 + ')');
        }
    }

    final int a(int n2, int n3, int n4, boolean bl, int n5, int n6, int n7, boolean bl2, int n8) {
        try {
            if (n2 != 0) {
                this.B = null;
            }
            this.U[this.eb] = 6;
            ++c;
            this.g[this.eb] = true;
            this.cb[this.eb] = bl2;
            this.D[this.eb] = false;
            this.k[this.eb] = n6;
            this.Y[this.eb] = bl;
            this.kb[this.eb] = n8;
            this.B[this.eb] = n5;
            this.ob[this.eb] = n4;
            this.O[this.eb] = n7;
            this.sb[this.eb] = n3;
            this.yb[this.eb] = "";
            return this.eb++;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[21] + n2 + ',' + n3 + ',' + n4 + ',' + bl + ',' + n5 + ',' + n6 + ',' + n7 + ',' + bl2 + ',' + n8 + ')');
        }
    }

    final int a(int n2, int n3, int n4, boolean bl, int n5, int n6, int n7, int n8) {
        try {
            this.U[this.eb] = 9;
            ++jb;
            this.g[this.eb] = true;
            this.D[this.eb] = false;
            this.k[this.eb] = n8;
            this.Y[this.eb] = bl;
            this.kb[this.eb] = n2;
            if (n5 < 40) {
                return 21;
            }
            this.B[this.eb] = n7;
            this.ob[this.eb] = n3;
            this.O[this.eb] = n4;
            this.sb[this.eb] = n6;
            this.Ab[this.eb] = new String[n6];
            this.m[this.eb] = new int[n6];
            this.Fb[this.eb] = new String[n6];
            this.p[this.eb] = new String[n6];
            this.pb[this.eb] = 0;
            this.j[this.eb] = 0;
            this.vb[this.eb] = -1;
            this.N[this.eb] = -1;
            return this.eb++;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[43] + n2 + ',' + n3 + ',' + n4 + ',' + bl + ',' + n5 + ',' + n6 + ',' + n7 + ',' + n8 + ')');
        }
    }

    private final void a(int n2, int n3, int n4, int n5) {
        try {
            ++Q;
            this.w.b(n5, n3, n4, n2, (byte)-119);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[38] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    final void b(byte by, int n2) {
        try {
            this.g[n2] = false;
            if (by <= 33) {
                this.R = -86;
            }
            ++rb;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[17] + by + ',' + n2 + ')');
        }
    }

    final String g(int n2, int n3) {
        try {
            ++nb;
            if (n3 != 4) {
                return null;
            }
            if (null == this.yb[n2]) {
                return W[1];
            }
            return this.yb[n2];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[41] + n2 + ',' + n3 + ')');
        }
    }

    final void a(int n2, String string, int n3, int n4, String string2, String string3, int n5) {
        block11: {
            try {
                block10: {
                    ++ub;
                    this.Ab[n5][n2] = string3;
                    int n6 = 78 % ((n3 - -14) / 54);
                    this.m[n5][n2] = n4;
                    this.Fb[n5][n2] = string2;
                    this.p[n5][n2] = string;
                    if (n2 - -1 > this.pb[n5]) break block10;
                    break block11;
                }
                this.pb[n5] = n2 + 1;
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(W[26]).append(n2).append(',');
                String string4 = string != null ? W[2] : W[1];
                StringBuilder stringBuilder2 = stringBuilder.append(string4).append(',').append(n3).append(',').append(n4).append(',');
                String string5 = string2 != null ? W[2] : W[1];
                StringBuilder stringBuilder3 = stringBuilder2.append(string5).append(',');
                String string6 = string3 != null ? W[2] : W[1];
                throw i.a(runtimeException2, stringBuilder3.append(string6).append(',').append(n5).append(')').toString());
            }
        }
    }

    final int f(int n2, int n3) {
        try {
            ++H;
            if (n2 != 14458) {
                this.t = true;
            }
            return this.vb[n3];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[27] + n2 + ',' + n3 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, int n3) {
        boolean bl = client.vh;
        try {
            int n4;
            int n5;
            if (n2 != -12) {
                this.d(-17, 7);
            }
            ++q;
            if (-1 == ~n3) return;
            if (0 == ~this.gb) return;
            if (null == this.yb[this.gb]) return;
            if (!this.g[this.gb]) return;
            int n6 = this.yb[this.gb].length();
            if (8 == n3 && 0 < n6) {
                this.yb[this.gb] = this.yb[this.gb].substring(0, -1 + n6);
            }
            if ((n3 == 10 || 13 == n3) && n6 > 0) {
                this.D[this.gb] = true;
            }
            String string = W[7];
            if (n6 < this.sb[this.gb]) {
                int n7 = 0;
                while (~string.length() < ~n7) {
                    n5 = ~string.charAt(n7);
                    n4 = ~n3;
                    if (!bl) {
                        if (n5 == n4) {
                            int n8 = this.gb;
                            this.yb[n8] = this.yb[n8] + (char)n3;
                        }
                        ++n7;
                        if (!bl) continue;
                    }
                    break;
                }
            } else {
                n5 = ~n3;
                n4 = -10;
            }
            if (n5 != n4) return;
            do {
                this.gb = (1 + this.gb) % this.eb;
                if (5 == this.U[this.gb]) return;
            } while (this.U[this.gb] != 6);
            if (bl) return;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[6] + n2 + ',' + n3 + ')');
        }
    }

    private final void a(int n2, int n3, boolean bl, int n4, int n5) {
        try {
            this.w.a(n4, (byte)-127, 0, n2, n5, n3);
            ++n;
            this.w.e(n4, n3, n2, 27785, n5, this.fb);
            if (!bl) {
                return;
            }
            this.w.e(1 + n4, n3 + -2, 1 + n2, 27785, -2 + n5, this.E);
            this.w.e(n4 - -2, -4 + n3, 2 + n2, 27785, -4 + n5, this.f);
            this.w.b(-1, u.g + 2, n2, n4);
            this.w.b(-1, 3 - -u.g, n2, -7 + (n3 + n4));
            this.w.b(-1, 4 + u.g, n2 - (-n5 - -7), n4);
            this.w.b(-1, 5 + u.g, -7 + n5 + n2, -7 + (n4 - -n3));
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[10] + n2 + ',' + n3 + ',' + bl + ',' + n4 + ',' + n5 + ')');
        }
    }

    private final void a(int n2, int n3, int n4, int n5, int n6, int n7, byte by) {
        try {
            ++L;
            int n8 = n4 - -n5 + -12;
            int n9 = -91 % ((58 - by) / 33);
            this.w.e(n8, 12, n7, 27785, n6, 0);
            this.w.b(-1, 0 + u.g, n7 + 1, 1 + n8);
            this.w.b(-1, u.g + 1, -12 + n6 + n7, 1 + n8);
            this.w.b(12, 0, n8, 13 + n7, (byte)-49);
            this.w.b(12, 0, n8, -13 + n7 - -n6, (byte)-119);
            this.w.b(1 + n8, this.R, 11, this.qb, -27 + n6, 14 + n7, 19020);
            this.w.a(n8 - -3, (byte)-105, this.i, 14 + n7 + n3, n2, 7);
            this.w.b(n8 - -2, n7 + n3 + 14, this.C, n2, 0);
            this.w.b(n8 + 10, 14 + n3 - -n7, this.ib, n2, 0);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[46] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ',' + by + ')');
        }
    }

    final int b(int n2, int n3) {
        try {
            ++lb;
            if (n3 != 17050) {
                this.E = 56;
            }
            int n4 = this.N[n2];
            return n4;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[23] + n2 + ',' + n3 + ')');
        }
    }

    final int c(int n2, int n3, int n4, int n5) {
        try {
            if (n5 >= -52) {
                this.d = null;
            }
            ++S;
            int n6 = this.w.kb[n2];
            int n7 = this.w.R[n2];
            this.U[this.eb] = 12;
            this.g[this.eb] = true;
            this.D[this.eb] = false;
            this.kb[this.eb] = n4 + -(n6 / 2);
            this.B[this.eb] = -(n7 / 2) + n3;
            this.ob[this.eb] = n6;
            this.O[this.eb] = n7;
            this.k[this.eb] = n2;
            return this.eb++;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[14] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    final void a(String string, boolean bl, int n2, String string2, String string3, byte by, int n3) {
        boolean bl2 = client.vh;
        try {
            block22: {
                block21: {
                    int n4;
                    block20: {
                        ++z;
                        int n5 = n3;
                        int n6 = this.pb[n5];
                        this.pb[n5] = n6 + 1;
                        n4 = n6;
                        if (~this.sb[n3] >= ~n4) break block20;
                        break block21;
                    }
                    int n7 = n3;
                    this.pb[n7] = this.pb[n7] - 1;
                    --n4;
                    for (int i2 = 0; n4 > i2; ++i2) {
                        this.Ab[n3][i2] = this.Ab[n3][1 + i2];
                        this.m[n3][i2] = this.m[n3][i2 - -1];
                        this.Fb[n3][i2] = this.Fb[n3][i2 + 1];
                        this.p[n3][i2] = this.p[n3][i2 + 1];
                        if (!bl2) {
                            if (!bl2) continue;
                            break;
                        }
                        break block22;
                    }
                }
                this.Ab[n3][n4] = string;
                this.m[n3][n4] = n2;
            }
            if (by > -39) {
                this.pb = null;
            }
            this.Fb[n3][n4] = string3;
            this.p[n3][n4] = string2;
            if (bl) {
                this.j[n3] = 999999;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(W[9]);
            String string4 = string != null ? W[2] : W[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string4).append(',').append(bl).append(',').append(n2).append(',');
            String string5 = string2 != null ? W[2] : W[1];
            StringBuilder stringBuilder3 = stringBuilder2.append(string5).append(',');
            String string6 = string3 != null ? W[2] : W[1];
            throw i.a(runtimeException2, stringBuilder3.append(string6).append(',').append(by).append(',').append(n3).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(String[] stringArray, int n2, int n3, int n4, int n5, int n6) {
        boolean bl = client.vh;
        try {
            ++u;
            int n7 = stringArray.length;
            int n8 = 80 % ((n3 - -55) / 61);
            int n9 = -((n7 - 1) * this.w.a(508305352, n4) / 2) + n6;
            int n10 = 0;
            do {
                int n11;
                int n12;
                block27: {
                    block28: {
                        block24: {
                            block26: {
                                block25: {
                                    block23: {
                                        block22: {
                                            if (~n10 <= ~n7) return;
                                            if (bl) return;
                                            if (!this.Y[n2]) break block22;
                                            n12 = 0xFFFFFF;
                                            if (!bl) break block23;
                                        }
                                        n12 = 0;
                                    }
                                    if (~(-((n11 = this.w.a(n4, 112, stringArray[n10])) / 2) + n5) < ~this.bb || this.bb > n5 - -(n11 / 2) || ~n9 > ~(-2 + this.hb) || ~(this.hb - 2) >= ~(-this.w.a(508305352, n4) + n9)) break block24;
                                    if (!this.Y[n2]) break block25;
                                    n12 = 0x808080;
                                    if (!bl) break block26;
                                }
                                n12 = 0xFFFFFF;
                            }
                            if (this.zb == 1) {
                                this.vb[n2] = n10;
                                this.D[n2] = true;
                            }
                        }
                        if (~this.vb[n2] != ~n10) break block27;
                        if (this.Y[n2]) break block28;
                        n12 = 0xC00000;
                        if (!bl) break block27;
                    }
                    n12 = 0xFF0000;
                }
                this.w.a(0, n9, stringArray[n10], -(n11 / 2) + n5, n12, (byte)-126, n4);
                n9 += this.w.a(508305352, n4);
                ++n10;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(W[24]);
            if (stringArray != null) {
                string = W[2];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(')').toString());
            }
            string = W[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(')').toString());
        }
    }

    final String b(int n2, int n3, int n4) {
        try {
            ++e;
            if (n3 != 19680) {
                this.a(38, -6, -46, -34, -108, -42, (byte)-17);
            }
            return this.Fb[n4][n2];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, W[40] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    qa(ua ua2, int n2) {
        try {
            this.Fb = new String[n2][];
            this.Y = new boolean[n2];
            this.B = new int[n2];
            this.cb = new boolean[n2];
            this.k = new int[n2];
            this.pb = new int[n2];
            this.ob = new int[n2];
            this.g = new boolean[n2];
            this.w = ua2;
            this.N = new int[n2];
            this.vb = new int[n2];
            this.U = new int[n2];
            this.m = new int[n2][];
            this.yb = new String[n2];
            this.d = new boolean[n2];
            this.j = new int[n2];
            this.Ab = new String[n2][];
            this.O = new int[n2];
            this.D = new boolean[n2];
            this.p = new String[n2][];
            this.kb = new int[n2];
            this.sb = new int[n2];
            this.R = this.a(176, 114, 114, (byte)-98);
            this.qb = this.a(62, 14, 14, (byte)-75);
            this.C = this.a(232, 208, 200, (byte)-88);
            this.i = this.a(184, 129, 96, (byte)-79);
            this.ib = this.a(115, 95, 53, (byte)-92);
            this.fb = this.a(171, 142, 117, (byte)-124);
            this.E = this.a(158, 122, 98, (byte)-117);
            this.f = this.a(136, 100, 86, (byte)-113);
            this.tb = this.a(179, 146, 135, (byte)-89);
            this.F = this.a(151, 112, 97, (byte)-122);
            this.Eb = this.a(136, 102, 88, (byte)-124);
            this.J = this.a(120, 93, 84, (byte)-92);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(W[22]);
            String string = ua2 != null ? W[2] : W[1];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(')').toString());
        }
    }

    static {
        W = new String[]{qa.z(qa.z("\u0004I$CY")), qa.z(qa.z("\u001b]fe")), qa.z(qa.z("\u000e\u0006$'\f")), qa.z(qa.z("\u0004I$B0]")), qa.z(qa.z("\u0004I$KY")), qa.z(qa.z("\u0004I$X0]")), qa.z(qa.z("\u0004I$AY")), qa.z(qa.z("4jIM43oB@;>dGG>%yXZ% ~]Q(/Ihj\u0015\u0010Nma\u0018\u001fCfd\u001f\u001aX{{\u0002\u0001]|~\t\fR:8CF\u001c??FM\u0011++\u00d2Q\rT/[]\u0001'VL^sqT\fN\u0012-IR\u000b\u00046'OZ\u0017VuQ")), qa.z(qa.z("\u0004I$OY")), qa.z(qa.z("\u0004I$C0]")), qa.z(qa.z("\u0004I$_Y")), qa.z(qa.z("\u0004I$A0]")), qa.z(qa.z("\u0004I$FY")), qa.z(qa.z("\u0004I$NY")), qa.z(qa.z("\u0004I$Y0]")), qa.z(qa.z("\u0004I$@0]")), qa.z(qa.z("\u0004I$EY")), qa.z(qa.z("\u0004I$L0]")), qa.z(qa.z("\u0004I$G0]")), qa.z(qa.z("\u0004I$M0]")), qa.z(qa.z("U\b")), qa.z(qa.z("\u0004I$\\Y")), qa.z(qa.z("\u0004I$5\u0018\u001bA~7Y")), qa.z(qa.z("\u0004I$JY")), qa.z(qa.z("\u0004I$O0]")), qa.z(qa.z("\u0004I$^Y")), qa.z(qa.z("\u0004I$XY")), qa.z(qa.z("\u0004I$J0]")), qa.z(qa.z("\u0004I$ZY")), qa.z(qa.z("[Bl")), qa.z(qa.z("\u0004I$N0]")), qa.z(qa.z("=Mf\u007f\u0014\u0001Aih")), qa.z(qa.z("\u001dMf\u007f\u0014\u0001Aih")), qa.z(qa.z("\u0004I$GY")), qa.z(qa.z("\u0004I$[Y")), qa.z(qa.z("\u0004I$MY")), qa.z(qa.z("\u0004I$@Y")), qa.z(qa.z("\u0004I$DY")), qa.z(qa.z("\u0004I$[0]")), qa.z(qa.z("\u0004I$BY")), qa.z(qa.z("\u0004I$]Y")), qa.z(qa.z("\u0004I$K0]")), qa.z(qa.z("\u0004I$F0]")), qa.z(qa.z("\u0004I$H0]")), qa.z(qa.z("\u0004I$E0]")), qa.z(qa.z("\u0004I$LY")), qa.z(qa.z("\u0004I$HY")), qa.z(qa.z("\u0004I$YY")), qa.z(qa.z("\u0004I$D0]"))};
        l = pa.a(-126);
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x71);
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
                        n4 = 40;
                        break;
                    }
                    case 2: {
                        n4 = 10;
                        break;
                    }
                    case 3: {
                        n4 = 9;
                        break;
                    }
                    default: {
                        n4 = 113;
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

