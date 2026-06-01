/*
 * Decompiled with CFR 0.152.
 */
import java.awt.Component;
import java.awt.Graphics;
import java.awt.Image;
import java.awt.image.ColorModel;
import java.awt.image.DirectColorModel;
import java.awt.image.ImageConsumer;
import java.awt.image.ImageObserver;
import java.awt.image.ImageProducer;

class ua
implements ImageProducer,
ImageObserver {
    static int o;
    private int[] G;
    static int W;
    private int[] tb;
    private int[] Sb;
    static int S;
    boolean xb;
    int[] qb;
    static int d;
    static int db;
    static int sb;
    static int K;
    static int j;
    static int zb;
    static int N;
    static int e;
    static int Q;
    boolean i;
    static int Z;
    static int q;
    static int m;
    private int[] Xb;
    static int Cb;
    static int L;
    static v E;
    private int[] t;
    int[] Eb;
    private int[] Tb;
    private int[] Wb;
    int[] rb;
    static int r;
    static int bb;
    static int n;
    static int b;
    private ColorModel nb;
    static int ib;
    int[] R;
    static int Vb;
    byte[][] gb;
    static int a;
    static int y;
    static int ab;
    int[] kb;
    static int U;
    static int pb;
    int u;
    static int v;
    static int yb;
    int[][] ob;
    private int A;
    static int V;
    static int w;
    private int hb;
    private int D;
    static int ub;
    static int s;
    int[][] Y;
    static int Ob;
    static int f;
    static int Jb;
    static int Pb;
    static int Ub;
    static int mb;
    static int Ib;
    static int O;
    static int H;
    int k;
    static int J;
    static int x;
    private boolean[] Qb;
    static String[] wb;
    static int T;
    static int jb;
    static int B;
    private int lb;
    private ImageConsumer fb;
    static int c;
    static int Nb;
    static int[] Bb;
    static int X;
    static int F;
    static int z;
    static int Db;
    static int C;
    private Image Gb;
    static int vb;
    static int l;
    static int Fb;
    static int eb;
    static int p;
    static int cb;
    static int Lb;
    private int[] M;
    static int I;
    static int[] Ab;
    private int Rb;
    static String[] Kb;
    static int P;
    private int[] Hb;
    static int g;
    static String[] h;
    static int[] Mb;
    private static final String[] Yb;

    final void e(int n2, int n3, int n4, int n5, int n6, int n7) {
        try {
            this.b(n3, n7, n2, n4, (byte)115);
            if (n5 != 27785) {
                return;
            }
            ++l;
            this.b(n3, n7, n2, -1 + n6 + n4, (byte)-117);
            this.b(n2, n4, n7, n6, 0);
            this.b(n3 + n2 - 1, n4, n7, n6, 0);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[7] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void b(int n2, int n3, int n4, int n5, int n6, int n7, int n8) {
        boolean bl = client.vh;
        try {
            ++eb;
            if (this.hb > n2) {
                n4 -= -n2 + this.hb;
                n2 = this.hb;
            }
            if (this.lb < n4 + n2) {
                n4 = -n2 + this.lb;
            }
            int n16 = (n5 & 0xFF8DC6) >> 1599703024;
            int n15 = n5 >> -272197656 & 0xFF;
            int n14 = n5 & 0xFF;
            int n13 = (n3 & 0xFFDB64) >> -411014704;
            int n12 = n3 >> 442466760 & 0xFF;
            int n11 = n3 & 0xFF;
            int n10 = -n4 + this.u;
            int n9 = 1;
            if (!(!this.i)) {
                n10 += this.u;
                n9 = 2;
                if (0 != (n7 & 1)) {
                    ++n7;
                    --n6;
                }
            }
            if (n8 != 19020) {
                this.a(-124, 53, -53, -76, (byte)-44);
            }
            int n17 = n2 - -(this.u * n7);
            int i2 = 0;
            do {
                block22: {
                    block21: {
                        block20: {
                            if (n6 <= i2) return;
                            if (bl) return;
                            if (~(i2 + n7) > ~this.A || n7 + i2 >= this.Rb) break block21;
                            int n18 = ((n15 * i2 + n12 * (-i2 + n6)) / n6 << -1085162904) + ((n13 * (-i2 + n6) + n16 * i2) / n6 << -1270717776) + (i2 * n14 - -(n11 * (-i2 + n6))) / n6;
                            int n19 = -n4;
                            while (-1 < ~n19) {
                                this.rb[n17++] = n18;
                                ++n19;
                                if (!bl) {
                                    if (!bl) continue;
                                }
                                break block20;
                            }
                            n17 += n10;
                        }
                        if (!bl) break block22;
                    }
                    n17 += this.u;
                }
                i2 += n9;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[10] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ',' + n8 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int[] nArray, int n2, int n3, int n4, int n5, int[] nArray2, byte by, int n6, int n7, int n8, int n9, int n10, int n11, int n12) {
        boolean bl = client.vh;
        try {
            ++O;
            try {
                int n13 = n8;
                int n14 = -123 % ((-11 - by) / 63);
                int n15 = -n7;
                do {
                    int n16;
                    int n17;
                    block14: {
                        if (~n15 <= -1) return;
                        int n18 = (n5 >> -204204944) * n11;
                        n5 += n6;
                        if (bl) return;
                        int n19 = -n10;
                        while (-1 < ~n19) {
                            block16: {
                                block15: {
                                    n4 = nArray[(n8 >> 80730192) - -n18];
                                    n8 += n3;
                                    n17 = 0;
                                    n16 = n4;
                                    if (bl) break block14;
                                    if (n17 == n16) break block15;
                                    nArray2[n12++] = n4;
                                    if (!bl) break block16;
                                }
                                ++n12;
                            }
                            ++n19;
                            if (!bl) continue;
                        }
                        n12 += n9;
                        n8 = n13;
                        n17 = n15;
                        n16 = n2;
                    }
                    n15 = n17 + n16;
                } while (!bl);
                return;
            }
            catch (Exception exception) {
                System.out.println(Yb[47]);
                return;
            }
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[48]).append(nArray != null ? Yb[1] : Yb[0]).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',');
            if (nArray2 != null) {
                string = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(by).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(')').toString());
            }
            string = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(by).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(')').toString());
        }
    }

    private final void a(int n2, int n3, int n4, int[] nArray, int n5, int n6, int n7, int[] nArray2, int n8, int n9, int n10, int n11) {
        boolean bl = client.vh;
        try {
            ++c;
            int n12 = 256 - n6;
            if (n8 > -54) {
                return;
            }
            int n13 = -n3;
            while (-1 < ~n13 && !bl) {
                int n14;
                int n15;
                block13: {
                    int n16 = -n11;
                    while (~n16 > -1) {
                        block15: {
                            block14: {
                                n5 = nArray[n2++];
                                n15 = 0;
                                n14 = n5;
                                if (bl) break block13;
                                if (n15 != n14) break block14;
                                ++n4;
                                if (!bl) break block15;
                            }
                            int n17 = nArray2[n4];
                            nArray2[n4++] = ib.a(0xFF0000, n6 * ib.a(n5, 65280) + n12 * ib.a(65280, n17)) + ib.a(n12 * ib.a(n17, 0xFF00FF) + ib.a(n5, 0xFF00FF) * n6, -16711936) >> -379053496;
                        }
                        ++n16;
                        if (!bl) continue;
                    }
                    n2 += n9;
                    n4 += n10;
                    n15 = n13;
                    n14 = n7;
                }
                n13 = n15 + n14;
                if (!bl) continue;
                break;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[91]).append(n2).append(',').append(n3).append(',').append(n4).append(',');
            String string = nArray != null ? Yb[1] : Yb[0];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',');
            String string2 = nArray2 != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(',').append(n11).append(')').toString());
        }
    }

    final void b(int n2, String string, int n3, int n4, int n5, int n6) {
        try {
            int n7 = 24 % ((n5 - -11) / 58);
            ++s;
            this.a(n2, n3, string, n4, -12200, n6, 0);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[69]).append(n2).append(',');
            String string2 = string != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final void a(int n2, int[] nArray, int n3, int[] nArray2, int n4, int n5, int n6, int n7) {
        boolean bl = client.vh;
        try {
            int n8;
            int n9;
            block13: {
                ++v;
                if (-1 >= ~n3) {
                    return;
                }
                n4 = nArray[(0xFFA7 & n2) >> 389407848];
                n2 += (n5 <<= 2);
                n8 = n9 = n3 / 16;
                while (-1 < ~n8) {
                    nArray2[n6++] = n4 - -ib.a(0x7F7F7F, nArray2[n6] >> 980406721);
                    nArray2[n6++] = n4 + ib.a(0x7F7F7F, nArray2[n6] >> -33478591);
                    nArray2[n6++] = n4 + (ib.a(nArray2[n6], 0xFEFEFF) >> 1438034561);
                    nArray2[n6++] = (ib.a(0xFEFEFF, nArray2[n6]) >> -1664277151) + n4;
                    n4 = nArray[0xFF & n2 >> 1869141800];
                    nArray2[n6++] = ib.a(nArray2[n6] >> -651215775, 0x7F7F7F) + n4;
                    nArray2[n6++] = ib.a(nArray2[n6] >> 1567416321, 0x7F7F7F) + n4;
                    nArray2[n6++] = (ib.a(0xFEFEFF, nArray2[n6]) >> -109945983) + n4;
                    nArray2[n6++] = ib.a(nArray2[n6] >> -1634216127, 0x7F7F7F) + n4;
                    n4 = nArray[(n2 += n5) >> 1972579688 & 0xFF];
                    nArray2[n6++] = (ib.a(nArray2[n6], 0xFEFEFE) >> 18481057) + n4;
                    nArray2[n6++] = (ib.a(0xFEFEFF, nArray2[n6]) >> 1645567265) + n4;
                    nArray2[n6++] = n4 - -ib.a(nArray2[n6] >> 363686529, 0x7F7F7F);
                    nArray2[n6++] = n4 + ib.a(0x7F7F7F, nArray2[n6] >> -417782847);
                    n4 = nArray[(0xFF16 & (n2 += n5)) >> -491054904];
                    nArray2[n6++] = (ib.a(nArray2[n6], 0xFEFEFF) >> -1655491807) + n4;
                    nArray2[n6++] = n4 + ib.a(nArray2[n6] >> 421283745, 0x7F7F7F);
                    nArray2[n6++] = n4 - -(ib.a(nArray2[n6], 0xFEFEFF) >> 1309685921);
                    nArray2[n6++] = n4 + (ib.a(0xFEFEFF, nArray2[n6]) >> 1995672417);
                    n4 = nArray[(n2 += n5) >> 1728371944 & 0xFF];
                    n2 += n5;
                    ++n8;
                    if (!bl) {
                        if (!bl) continue;
                    }
                    break block13;
                }
                n9 = -(n3 % 16);
            }
            n8 = n7;
            do {
                if (~n8 <= ~n9) return;
                nArray2[n6++] = ib.a(nArray2[n6] >> 1543799489, 0x7F7F7F) + n4;
                if (bl) return;
                if ((3 & n8) == 3) {
                    n4 = nArray[(n2 & 0xFF38) >> -300394456];
                    n2 += n5;
                    n2 += n5;
                }
                ++n8;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[63]).append(n2).append(',').append(nArray != null ? Yb[1] : Yb[0]).append(',').append(n3).append(',');
            if (nArray2 != null) {
                string = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(')').toString());
            }
            string = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(')').toString());
        }
    }

    @Override
    public final synchronized void addConsumer(ImageConsumer imageConsumer) {
        try {
            this.fb = imageConsumer;
            ++r;
            imageConsumer.setDimensions(this.u, this.k);
            imageConsumer.setProperties(null);
            imageConsumer.setColorModel(this.nb);
            imageConsumer.setHints(14);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[78]);
            String string = imageConsumer != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void b(int n2, int n3, int n4, int n5) {
        try {
            int n6;
            if (n2 != -1) {
                this.a(41, 58, 102, 22, (byte)102);
            }
            ++Lb;
            if (this.Qb[n3]) {
                n5 += this.Sb[n3];
                n4 += this.G[n3];
            }
            int n7 = n4 * this.u + n5;
            int n8 = 0;
            int n9 = this.R[n3];
            int n10 = this.kb[n3];
            int n11 = -n10 + this.u;
            int n12 = 0;
            if (~n4 > ~this.A) {
                n6 = this.A - n4;
                n9 -= n6;
                n4 = this.A;
                n7 += this.u * n6;
                n8 += n6 * n10;
            }
            if (~this.Rb >= ~(n4 - -n9)) {
                n9 -= 1 + (n9 + n4 - this.Rb);
            }
            if (~this.hb < ~n5) {
                n6 = -n5 + this.hb;
                n8 += n6;
                n11 += n6;
                n10 -= n6;
                n12 += n6;
                n5 = this.hb;
                n7 += n6;
            }
            if (n5 - -n10 >= this.lb) {
                n6 = n5 - -n10 + -this.lb - -1;
                n10 -= n6;
                n12 += n6;
                n11 += n6;
            }
            if (0 >= n10) return;
            if (n9 <= 0) {
                return;
            }
            n6 = 1;
            if (!(!this.i)) {
                n11 += this.u;
                if (~(1 & n4) != -1) {
                    n7 += this.u;
                    --n9;
                }
                n6 = 2;
                n12 += this.kb[n3];
            }
            if (this.ob[n3] != null) {
                this.a(n10, this.rb, n6, n9, 0, n8, (byte)123, n7, this.ob[n3], n11, n12);
                if (!client.vh) return;
            }
            this.a(n7, this.Y[n3], n8, n12, this.rb, n6, n9, true, this.gb[n3], n10, n11);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[17] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    @Override
    public final void requestTopDownLeftRightResend(ImageConsumer imageConsumer) {
        try {
            ++Ib;
            System.out.println(Yb[59]);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[58]);
            String string = imageConsumer != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, int n3, int n4, int n5, int n6, int n7, int n8) {
        boolean bl = client.vh;
        try {
            int n9;
            block15: {
                int n10 = n5;
                block7: while (true) {
                    int n11 = ~(n5 + n7);
                    block8: while (n11 < ~n10) {
                        n9 = n4;
                        if (bl) break block15;
                        int n12 = n9;
                        block9: while (true) {
                            int n13 = ~n12;
                            block10: while (n13 > ~(n4 + n2)) {
                                int n14 = 0;
                                int n15 = 0;
                                int n16 = 0;
                                int n17 = 0;
                                n11 = -n8 + n10;
                                if (bl) continue block8;
                                int n18 = n11;
                                block11: while (true) {
                                    int n19 = ~(n8 + n10);
                                    int n20 = ~n18;
                                    block12: while (n19 <= n20) {
                                        n13 = n18;
                                        if (bl) continue block10;
                                        if (n13 >= 0 && this.u > n18) {
                                            for (int i2 = n12 - n3; i2 <= n12 + n3; ++i2) {
                                                n19 = -1;
                                                n20 = ~i2;
                                                if (bl) continue block12;
                                                if (n19 < n20 || ~i2 <= ~this.k) continue;
                                                int n21 = this.rb[this.u * i2 + n18];
                                                n16 += 0xFF & n21;
                                                ++n17;
                                                n15 += (n21 & 0xFF81) >> 743340392;
                                                n14 += (n21 & 0xFF64A6) >> 483715504;
                                                if (!bl) continue;
                                            }
                                        }
                                        ++n18;
                                        if (!bl) continue block11;
                                    }
                                    break;
                                }
                                this.rb[n10 + n12 * this.u] = n16 / n17 + ((n14 / n17 << -148272656) + (n15 / n17 << 2002983304));
                                ++n12;
                                if (!bl) continue block9;
                            }
                            break;
                        }
                        ++n10;
                        if (!bl) continue block7;
                    }
                    break;
                }
                ++Vb;
                n9 = n6;
            }
            if (n9 == 0xFF7000) return;
            this.a(-18, (byte)79, -10, 106, -42, 27);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[20] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ',' + n8 + ')');
        }
    }

    final void a(int n2, int n3, int n4, int n5, int n6, byte by, int n7) {
        try {
            ++ub;
            try {
                int n8;
                int n9;
                int n10;
                int n11 = this.kb[n4];
                int n12 = this.R[n4];
                int n13 = 0;
                int n14 = 0;
                int n15 = (n11 << -1350099408) / n7;
                int n16 = (n12 << 653525744) / n6;
                if (this.Qb[n4]) {
                    block16: {
                        n10 = this.Eb[n4];
                        n9 = this.qb[n4];
                        if (0 != n10 && -1 != ~n9) break block16;
                        return;
                    }
                    if (~(this.Sb[n4] * n7 % n10) != -1) {
                        n13 = (-(this.Sb[n4] * n7 % n10) + n10 << -1137453552) / n7;
                    }
                    n2 += (-1 + (n7 * this.Sb[n4] - -n10)) / n10;
                    n15 = (n10 << -716906352) / n7;
                    n5 += (n9 + n6 * this.G[n4] - 1) / n9;
                    n16 = (n9 << 1305987728) / n6;
                    if (-1 != ~(this.G[n4] * n6 % n9)) {
                        n14 = (-(n6 * this.G[n4] % n9) + n9 << -1890544144) / n6;
                    }
                    n6 = n6 * (-(n14 >> 1020185680) + this.R[n4]) / n9;
                    n7 = (-(n13 >> -839014480) + this.kb[n4]) * n7 / n10;
                }
                if (by <= 102) {
                    this.Y = null;
                }
                n10 = n2 + n5 * this.u;
                n9 = this.u - n7;
                if (~n5 > ~this.A) {
                    n8 = this.A + -n5;
                    n5 = 0;
                    n6 -= n8;
                    n10 += n8 * this.u;
                    n14 += n8 * n16;
                }
                if (~n2 > ~this.hb) {
                    n8 = this.hb - n2;
                    n7 -= n8;
                    n9 += n8;
                    n13 += n8 * n15;
                    n10 += n8;
                    n2 = 0;
                }
                if (~(n5 - -n6) <= ~this.Rb) {
                    n6 -= n5 - -n6 - (this.Rb + -1);
                }
                if (~(n2 + n7) <= ~this.lb) {
                    n8 = n2 - -n7 + (-this.lb - -1);
                    n7 -= n8;
                    n9 += n8;
                }
                n8 = 1;
                if (!(!this.i)) {
                    n16 += n16;
                    if ((1 & n5) != 0) {
                        --n6;
                        n10 += this.u;
                    }
                    n8 = 2;
                    n9 += this.u;
                }
                this.a(n14, n15, n7, n13, n9, this.ob[n4], n10, this.rb, 0, n11, false, n16, n6, n3, n8);
            }
            catch (Exception exception) {
                System.out.println(Yb[16]);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[57] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + by + ',' + n7 + ')');
        }
    }

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    final void b(int n2) {
        boolean bl = client.vh;
        try {
            ++Z;
            int n3 = this.k * this.u;
            if (n2 != 0xF8F8F9) {
                wb = null;
            }
            for (int i2 = 0; n3 > i2; ++i2) {
                int n4 = 0xFFFFFF & this.rb[i2];
                try {
                    this.rb[i2] = ib.a(n4 >>> -93223452, 986895) + ((ib.a(n4, 0xF8F8F9) >>> 2097500387) + (ib.a(0xFEFEFF, n4) >>> 1336934849) - -ib.a(-2143338689, n4 >>> 1527263298));
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
            throw i.a(runtimeException, Yb[55] + n2 + ')');
        }
    }

    private final void a(int n2, int n3, int n4, int[] nArray, int n5, int n6, int n7, int[] nArray2, int n8, int n9, boolean bl) {
        block13: {
            boolean bl2 = client.vh;
            try {
                for (n4 = n2; n4 < 0; ++n4) {
                    this.rb[n9++] = nArray2[(n6 >> -1782373679) * n3 + (n7 >> -324278255)];
                    n7 += n5;
                    n6 += n8;
                    if (!bl2) {
                        if (!bl2) continue;
                        break;
                    }
                    break block13;
                }
                if (!bl) {
                    this.a(-59, -116, -115, true, 1, 118, 33, -46, -78, -30);
                }
                ++V;
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(Yb[68]).append(n2).append(',').append(n3).append(',').append(n4).append(',');
                String string = nArray != null ? Yb[1] : Yb[0];
                StringBuilder stringBuilder2 = stringBuilder.append(string).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',');
                String string2 = nArray2 != null ? Yb[1] : Yb[0];
                throw i.a(runtimeException2, stringBuilder2.append(string2).append(',').append(n8).append(',').append(n9).append(',').append(bl).append(')').toString());
            }
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, int n3, int n4, int n5) {
        try {
            ++db;
            if (this.hb > n3) return;
            if (~n2 > ~this.A) return;
            if (~n3 <= ~this.lb) return;
            if (~n2 <= ~this.Rb) return;
            if (n4 <= 44) {
                return;
            }
            this.rb[n3 - -(this.u * n2)] = n5;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[61] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void c(int n2, int n3, int n4, int n5, int n6, int n7, int n8) {
        boolean bl = client.vh;
        try {
            ++z;
            if (n6 < this.A) {
                n4 -= this.A - n6;
                n6 = this.A;
            }
            if (~n3 > ~this.hb) {
                n7 -= this.hb - n3;
                n3 = this.hb;
            }
            if (this.lb < n3 - -n7) {
                n7 = this.lb - n3;
            }
            if (this.Rb < n4 + n6) {
                n4 = -n6 + this.Rb;
            }
            int n14 = 256 - n2;
            int n13 = n2 * (n8 >> -2055680880 & 0xFF);
            int n12 = ((n8 & 0xFFC4) >> 1364192264) * n2;
            int n11 = n2 * (n8 & 0xFF);
            int n10 = this.u - n7;
            int n9 = 1;
            if (!(!this.i)) {
                if (-1 != ~(n6 & 1)) {
                    --n4;
                    ++n6;
                }
                n10 += this.u;
                n9 = 2;
            }
            int n15 = n3 - -(this.u * n6);
            int n16 = n5;
            block8: do {
                if (~n4 >= ~n16) return;
                if (bl) return;
                for (int i2 = -n7; 0 > i2; ++i2) {
                    int n17 = n14 * (this.rb[n15] & 0xFF);
                    int n18 = n14 * ((0xFF8835 & this.rb[n15]) >> 1661674448);
                    int n19 = n14 * ((0xFF7A & this.rb[n15]) >> 2108168104);
                    int n20 = (n17 + n11 >> -1220075704) + (n19 + n12 >> -855772152 << -628820632) + (n13 + n18 >> -540786712 << -681889584);
                    this.rb[n15++] = n20;
                    if (bl) continue block8;
                    if (!bl) continue;
                }
                n15 += n10;
                n16 += n9;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[62] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ',' + n8 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final int a(int n2, int n3, String string) {
        boolean bl = client.vh;
        try {
            int n4;
            ++Nb;
            int n5 = 0;
            if (n3 <= 67) {
                this.Rb = 74;
            }
            byte[] byArray = m.b[n2];
            for (int i2 = 0; i2 < string.length(); ++i2) {
                n4 = 64;
                if (bl) return n4;
                if (n4 == string.charAt(i2) && 4 + i2 < string.length() && string.charAt(i2 - -4) == '@') {
                    i2 += 4;
                    if (!bl) continue;
                }
                if (~string.charAt(i2) != '\uffffff81' || i2 + 4 >= string.length() || '~' != string.charAt(i2 + 4)) {
                    int n6 = string.charAt(i2);
                    if (~n6 > -1 || ~n.a.length >= ~n6) {
                        n6 = 32;
                    }
                    n5 += byArray[n.a[n6] + 7];
                    if (!bl) continue;
                }
                i2 += 4;
                if (!bl) continue;
            }
            n4 = n5;
            return n4;
        }
        catch (RuntimeException runtimeException) {
            String string2;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[11]).append(n2).append(',').append(n3).append(',');
            if (string != null) {
                string2 = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string2).append(')').toString());
            }
            string2 = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string2).append(')').toString());
        }
    }

    @Override
    public final boolean imageUpdate(Image image, int n2, int n3, int n4, int n5, int n6) {
        try {
            ++B;
            return true;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[84]);
            String string = image != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int n3, int n4, int[] nArray, int[] nArray2, int n5, int n6, int n7, int n8, int n9, byte by) {
        boolean bl = client.vh;
        try {
            ++yb;
            int n10 = n3;
            if (by != 102) {
                return;
            }
            do {
                block14: {
                    block13: {
                        if (-1 >= ~n10) return;
                        n9 = nArray[(n5 >> 930333777) * n7 + (n4 >> 98213361)];
                        n4 += n8;
                        if (bl) return;
                        if (-1 != ~n9) break block13;
                        ++n6;
                        if (!bl) break block14;
                    }
                    this.rb[n6++] = n9;
                }
                n5 += n2;
                ++n10;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[86]).append(n2).append(',').append(n3).append(',').append(n4).append(',').append(nArray != null ? Yb[1] : Yb[0]).append(',');
            if (nArray2 != null) {
                string = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(n9).append(',').append(by).append(')').toString());
            }
            string = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(n9).append(',').append(by).append(')').toString());
        }
    }

    private final void a(int n2, int n3, byte[] byArray, int n4, boolean bl, int n5, int n6, int n7, int n8, int[] nArray, int[] nArray2, int n9) {
        boolean bl2 = client.vh;
        try {
            ++U;
            int n10 = -n7 + 256;
            if (bl) {
                return;
            }
            int n11 = -n5;
            while (0 > n11 && !bl2) {
                int n12;
                int n13;
                block15: {
                    for (int i2 = -n8; i2 < 0; ++i2) {
                        int n14;
                        block16: {
                            n14 = byArray[n2++];
                            n13 = ~n14;
                            n12 = -1;
                            if (bl2) break block15;
                            if (n13 != n12) break block16;
                            ++n9;
                            if (!bl2) continue;
                        }
                        n14 = nArray2[0xFF & n14];
                        int n15 = nArray[n9];
                        nArray[n9++] = ib.a(n10 * ib.a(n15, 0xFF00FF) + ib.a(0xFF00FF, n14) * n7, -16711936) + ib.a(0xFF0000, n7 * ib.a(65280, n14) + n10 * ib.a(65280, n15)) >> 1273033224;
                        if (!bl2) continue;
                    }
                    n9 += n3;
                    n2 += n6;
                    n13 = n11;
                    n12 = n4;
                }
                n11 = n13 + n12;
                if (!bl2) continue;
                break;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[76]).append(n2).append(',').append(n3).append(',');
            String string = byArray != null ? Yb[1] : Yb[0];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',').append(n4).append(',').append(bl).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',');
            String string2 = nArray != null ? Yb[1] : Yb[0];
            StringBuilder stringBuilder3 = stringBuilder2.append(string2).append(',');
            String string3 = nArray2 != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder3.append(string3).append(',').append(n9).append(')').toString());
        }
    }

    private final int c(int n2, int n3) {
        try {
            ++o;
            if (~n3 == -1) {
                return m.b[n3][8] - 2;
            }
            if (n2 < 49) {
                this.a(-22, 77, 112, -35, -44, null, -45, null, -39, -33, false, 50, 61, 37, -7);
            }
            return m.b[n3][8] + -1;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[88] + n2 + ',' + n3 + ')');
        }
    }

    final void b(int n2, int n3, int n4, int n5, int n6, int n7) {
        boolean bl = client.vh;
        try {
            int n8;
            block10: {
                ++ib;
                this.kb[n6] = n7;
                this.R[n6] = n2;
                this.Qb[n6] = false;
                this.Sb[n6] = 0;
                this.G[n6] = 0;
                this.Eb[n6] = n7;
                this.qb[n6] = n2;
                int n9 = n2 * n7;
                int n10 = 0;
                this.ob[n6] = new int[n9];
                int n11 = n3;
                while (n3 + n7 > n11) {
                    block11: {
                        n8 = n4;
                        if (bl) break block10;
                        int n12 = n8;
                        while (~(n2 + n4) < ~n12) {
                            this.ob[n6][n10++] = this.rb[n11 + this.u * n12];
                            ++n12;
                            if (!bl) {
                                if (!bl) continue;
                                break;
                            }
                            break block11;
                        }
                        ++n11;
                    }
                    if (!bl) continue;
                }
                n8 = n5;
            }
            if (n8 != -27966) {
                this.a(73, -62, -30, (byte)-113, 44, -64, -91, 100, -79, null, 117, 11, 127, -109, null);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[18] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ')');
        }
    }

    void a(int n2, int n3, int n4, int n5, int n6, int n7, byte by, int n8) {
        try {
            this.f(n5, n6, n4, n7, 5924, n3);
            if (by != 29) {
                return;
            }
            ++T;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[56] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ',' + by + ',' + n8 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void c(int n2, int n3, int n4, int n5, int n6, int n7) {
        boolean bl = client.vh;
        try {
            ++X;
            int n12 = -n2 + 256;
            int n11 = n2 * (0xFF & n6 >> -25949072);
            int n10 = (0xFF & n6 >> -1057205208) * n2;
            int n9 = n2 * (n6 & 0xFF);
            if (n3 != -1057205208) {
                return;
            }
            int n8 = -n4 + n5;
            if (~n8 > -1) {
                n8 = 0;
            }
            int n13 = n5 - -n4;
            if (this.k <= n13) {
                n13 = -1 + this.k;
            }
            int n14 = 1;
            if (!(!this.i)) {
                if ((1 & n8) != 0) {
                    ++n8;
                }
                n14 = 2;
            }
            int n15 = n8;
            block9: do {
                if (n13 < n15) return;
                int n19 = n15 + -n5;
                int n18 = (int)Math.sqrt(-(n19 * n19) + n4 * n4);
                int n17 = n7 + -n18;
                if (bl) return;
                if (~n17 > -1) {
                    n17 = 0;
                }
                int n16 = n7 - -n18;
                if (this.u <= n16) {
                    n16 = this.u + -1;
                }
                int n20 = n17 + this.u * n15;
                for (int i2 = n17; n16 >= i2; ++i2) {
                    int n21 = (this.rb[n20] & 0xFF) * n12;
                    int n22 = n12 * ((0xFFD5 & this.rb[n20]) >> -1460723256);
                    int n23 = n12 * ((this.rb[n20] & 0xFF18B9) >> -342059728);
                    int n24 = (n21 + n9 >> 1268749032) + ((n10 - -n22 >> -404063160 << -187266712) + (n11 - -n23 >> 296955496 << 1872374416));
                    this.rb[n20++] = n24;
                    if (bl) continue block9;
                    if (!bl) continue;
                }
                n15 += n14;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[75] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int[] nArray, int[] nArray2, int n2, int n3, int n4, int n5, int n6, int n7, int n8, int n9, int n10, int n11, int n12, int n13, int n14, int n15, int n16) {
        boolean bl = client.vh;
        try {
            ++a;
            int n17 = n14 >> 1196549680 & 0xFF;
            int n18 = n14 >> -340653432 & 0xFF;
            int n19 = n14 & 0xFF;
            int n20 = n7 >> -922169040 & 0xFF;
            if (n5 != 1603920392) {
                this.Rb = 29;
            }
            int n21 = n7 >> 5163400 & 0xFF;
            int n22 = 0xFF & n7;
            try {
                int n23 = n10;
                int n24 = -n15;
                do {
                    int n25;
                    int n26;
                    int n27;
                    if (-1 >= ~n24) return;
                    int n28 = (n12 >> -1328879408) * n13;
                    int n29 = n4 >> -299524176;
                    int n30 = n2;
                    if (bl) return;
                    if (~n29 > ~this.hb) {
                        n27 = this.hb + -n29;
                        n30 -= n27;
                        n10 += n27 * n9;
                        n29 = this.hb;
                    }
                    if (this.lb <= n29 - -n30) {
                        n27 = -this.lb + n29 - -n30;
                        n30 -= n27;
                    }
                    n11 = 1 - n11;
                    if (0 != n11) {
                        for (n27 = n29; n29 + n30 > n27; n10 += n9, ++n27) {
                            n6 = nArray2[n28 + (n10 >> -1009344688)];
                            n26 = -1;
                            n25 = ~n6;
                            if (!bl) {
                                if (n26 == n25) continue;
                                int n31 = 0xFF & n6 >> 2008995472;
                                int n32 = 0xFF & n6 >> 1043792552;
                                int n33 = n6 & 0xFF;
                                if (n31 == n32 && n33 == n32) {
                                    nArray[n27 - -n16] = (n33 * n19 >> -1147526104) + ((n18 * n32 >> 1601552776 << -566695064) + (n31 * n17 >> 1603920392 << -538385168));
                                    if (!bl) continue;
                                }
                                if (-256 != ~n31 || ~n33 != ~n32) {
                                    nArray[n27 + n16] = n6;
                                    if (!bl) continue;
                                }
                                nArray[n27 - -n16] = (n22 * n33 >> -1345290488) + (n31 * n20 >> 1685589832 << -1431216016) - -(n32 * n21 >> 508305352 << -483710840);
                                if (!bl) continue;
                            }
                            break;
                        }
                    } else {
                        n12 += n8;
                        n10 = n23;
                        n16 += this.u;
                        n26 = n4;
                        n25 = n3;
                    }
                    n4 = n26 + n25;
                    ++n24;
                } while (!bl);
                return;
            }
            catch (Exception exception) {
                System.out.println(Yb[65]);
                return;
            }
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[66]).append(nArray != null ? Yb[1] : Yb[0]).append(',');
            if (nArray2 != null) {
                string = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(',').append(n15).append(',').append(n16).append(')').toString());
            }
            string = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(',').append(n15).append(',').append(n16).append(')').toString());
        }
    }

    private final void a(int n2, int[] nArray, int n3, int n4, int n5, int n6, int n7, int n8, byte[] byArray, int n9, byte by, int n10, int n11, int n12, int n13, int n14, int[] nArray2, int n15) {
        block18: {
            boolean bl = client.vh;
            try {
                ++J;
                int n16 = n15 >> 1698138864 & 0xFF;
                if (by <= 8) {
                    return;
                }
                int n17 = (n15 & 0xFF65) >> 1758159464;
                int n18 = n15 & 0xFF;
                int n19 = n11 >> 34872048 & 0xFF;
                int n20 = 0xFF & n11 >> 225029096;
                int n21 = 0xFF & n11;
                try {
                    int n22 = n10;
                    for (int i2 = -n2; 0 > i2; ++i2) {
                        int n23;
                        block19: {
                            block20: {
                                int n24;
                                int n25 = n3 * (n8 >> -1095936688);
                                int n26 = n12 >> -2034604816;
                                int n27 = n14;
                                if (bl) break block18;
                                if (~n26 > ~this.hb) {
                                    n24 = this.hb + -n26;
                                    n10 += n24 * n13;
                                    n27 -= n24;
                                    n26 = this.hb;
                                }
                                n9 = -n9 + 1;
                                if (n27 + n26 >= this.lb) {
                                    n24 = n26 + (n27 - this.lb);
                                    n27 -= n24;
                                }
                                if (n9 == 0) break block20;
                                for (n24 = n26; n24 < n27 + n26; ++n24) {
                                    block21: {
                                        int n28;
                                        int n29;
                                        int n30;
                                        block22: {
                                            block23: {
                                                n23 = n5 = 0xFF & byArray[(n10 >> -91857424) - -n25];
                                                if (bl) break block19;
                                                if (n23 == 0) break block21;
                                                n5 = nArray[n5];
                                                n30 = n5 & 0xFF;
                                                n29 = 0xFF & n5 >> 1941812904;
                                                n28 = 0xFF & n5 >> -309996368;
                                                if (n29 == n28 && ~n29 == ~n30) break block22;
                                                if (-256 != ~n28 || n29 != n30) break block23;
                                                nArray2[n7 + n24] = (n19 * n28 >> 1700051816 << -1733973648) + (n29 * n20 >> 269461512 << 971885320) + (n21 * n30 >> -1282776696);
                                                if (!bl) break block21;
                                            }
                                            nArray2[n24 + n7] = n5;
                                            if (!bl) break block21;
                                        }
                                        nArray2[n24 + n7] = (n18 * n30 >> -178220952) + (n28 * n16 >> -1022678840 << 2122310768) + (n17 * n29 >> 901102216 << -552193592);
                                    }
                                    n10 += n13;
                                    if (!bl) continue;
                                }
                            }
                            n8 += n4;
                            n12 += n6;
                            n10 = n22;
                            n23 = n7 + this.u;
                        }
                        n7 = n23;
                        if (!bl) continue;
                    }
                }
                catch (Exception exception) {
                    System.out.println(Yb[65]);
                }
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(Yb[85]).append(n2).append(',');
                String string = nArray != null ? Yb[1] : Yb[0];
                StringBuilder stringBuilder2 = stringBuilder.append(string).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',');
                String string2 = byArray != null ? Yb[1] : Yb[0];
                StringBuilder stringBuilder3 = stringBuilder2.append(string2).append(',').append(n9).append(',').append(by).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(',');
                String string3 = nArray2 != null ? Yb[1] : Yb[0];
                throw i.a(runtimeException2, stringBuilder3.append(string3).append(',').append(n15).append(')').toString());
            }
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(byte by, boolean bl, byte[] byArray, int n2, int n3, int n4, int n5) {
        try {
            int n6;
            ++P;
            int n7 = n2 - -byArray[5 + n4];
            if (by < 24) {
                return;
            }
            int n8 = n5 + -byArray[6 + n4];
            int n9 = byArray[n4 - -3];
            int n10 = byArray[4 + n4];
            int n11 = 16384 * byArray[n4] + byArray[n4 - -1] * 128 + byArray[n4 + 2];
            int n12 = n7 - -(n8 * this.u);
            int n13 = this.u + -n9;
            if (~this.A < ~n8) {
                n6 = this.A - n8;
                n11 += n6 * n9;
                n12 += this.u * n6;
                n10 -= n6;
                n8 = this.A;
            }
            int n14 = 0;
            if (~this.Rb >= ~(n8 + n10)) {
                n10 -= 1 + (n8 + n10) + -this.Rb;
            }
            if (~n7 > ~this.hb) {
                n6 = -n7 + this.hb;
                n14 += n6;
                n9 -= n6;
                n11 += n6;
                n7 = this.hb;
                n13 += n6;
                n12 += n6;
            }
            if (~(n9 + n7) <= ~this.lb) {
                n6 = -this.lb + (n9 + n7) + 1;
                n13 += n6;
                n14 += n6;
                n9 -= n6;
            }
            if (0 >= n9) return;
            if (-1 <= ~n10) {
                return;
            }
            if (bl) {
                this.a(byArray, n3, n9, n12, n10, n14, 1504725224, n13, this.rb, n11);
                if (!client.vh) return;
            }
            this.a(n3, this.rb, n12, (byte)37, n13, n10, n9, n11, byArray, n14);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[60]).append(by).append(',').append(bl).append(',');
            if (byArray != null) {
                string = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(')').toString());
            }
            string = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void f(int n2, int n3, int n4, int n5, int n6, int n7) {
        try {
            ++cb;
            try {
                int n8;
                int n9;
                int n10;
                int n15 = this.kb[n7];
                int n16 = this.R[n7];
                int n14 = 0;
                int n13 = 0;
                int n12 = (n15 << -311879024) / n5;
                int n11 = (n16 << -1693975408) / n4;
                if (this.Qb[n7]) {
                    n10 = this.Eb[n7];
                    n9 = this.qb[n7];
                    if (~n10 == -1) return;
                    if (n9 == 0) return;
                    if (-1 != ~(this.G[n7] * n4 % n9)) {
                        n13 = (n9 - n4 * this.G[n7] % n9 << 1500479664) / n4;
                    }
                    n12 = (n10 << 1015741296) / n5;
                    if (0 != this.Sb[n7] * n5 % n10) {
                        n14 = (n10 - this.Sb[n7] * n5 % n10 << -1148345008) / n5;
                    }
                    n2 += (-1 + (n5 * this.Sb[n7] - -n10)) / n10;
                    n11 = (n9 << -959198096) / n4;
                    n3 += (n9 + n4 * this.G[n7] - 1) / n9;
                    n4 = (this.R[n7] - (n13 >> 1117428208)) * n4 / n9;
                    n5 = n5 * (this.kb[n7] - (n14 >> -725293584)) / n10;
                }
                n10 = n2 + this.u * n3;
                if (~this.A < ~n3) {
                    n8 = this.A - n3;
                    n13 += n11 * n8;
                    n4 -= n8;
                    n10 += this.u * n8;
                    n3 = 0;
                }
                n9 = this.u + -n5;
                if (~this.Rb >= ~(n3 - -n4)) {
                    n4 -= -this.Rb + n3 - -n4 - -1;
                }
                if (n2 < this.hb) {
                    n8 = -n2 + this.hb;
                    n5 -= n8;
                    n9 += n8;
                    n10 += n8;
                    n2 = 0;
                    n14 += n12 * n8;
                }
                if (this.lb <= n2 + n5) {
                    n8 = 1 + (n2 + (n5 - this.lb));
                    n9 += n8;
                    n5 -= n8;
                }
                n8 = 1;
                if (!(!this.i)) {
                    if (0 != (n3 & 1)) {
                        --n4;
                        n10 += this.u;
                    }
                    n9 += this.u;
                    n8 = 2;
                    n11 += n11;
                }
                this.a(this.ob[n7], n8, n12, 0, n13, this.rb, (byte)78, n11, n4, n14, n9, n5, n15, n10);
            }
            catch (Exception exception) {
                System.out.println(Yb[16]);
            }
            if (n6 == 5924) return;
            this.u = -15;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[15] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ')');
        }
    }

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    final void b(int n2, int n3, int n4, int n5, byte by) {
        boolean bl = client.vh;
        try {
            block13: {
                block12: {
                    ++F;
                    if (this.A > n5) return;
                    if (n5 >= this.Rb) return;
                    if (~n4 > ~this.hb) {
                        n2 -= -n4 + this.hb;
                        n4 = this.hb;
                    }
                    if (~this.lb > ~(n4 - -n2)) break block12;
                    break block13;
                }
                n2 = -n4 + this.lb;
            }
            int n6 = -44 / ((by - 15) / 37);
            if (0 >= n2) return;
            int n7 = n4 - -(this.u * n5);
            for (int i2 = 0; n2 > i2; ++i2) {
                try {
                    this.rb[n7 - -i2] = n3;
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
            throw i.a(runtimeException, Yb[72] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + by + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, int n3, String string, int n4, int n5, byte by, int n6) {
        boolean bl = client.vh;
        try {
            int n7;
            int n8;
            block54: {
                try {
                    int n9;
                    if (~n2 < -1 && this.gb[n9 = -1 + (n2 + this.D)] != null) {
                        this.b(-1, n9, -this.R[n9] + n3, n4);
                        n4 += this.kb[n9] - -5;
                    }
                    int n10 = -5 % ((by - -91) / 33);
                    byte[] byArray = m.b[n6];
                    for (int i2 = 0; string.length() > i2; ++i2) {
                        int n11;
                        int n12;
                        block56: {
                            block59: {
                                block57: {
                                    block60: {
                                        block61: {
                                            block62: {
                                                block65: {
                                                    block66: {
                                                        block68: {
                                                            block69: {
                                                                block71: {
                                                                    block72: {
                                                                        block74: {
                                                                            block73: {
                                                                                block70: {
                                                                                    block67: {
                                                                                        block64: {
                                                                                            block63: {
                                                                                                block58: {
                                                                                                    n8 = -65;
                                                                                                    n7 = ~string.charAt(i2);
                                                                                                    if (bl) break block54;
                                                                                                    if (n8 != n7 || ~string.length() >= ~(i2 - -4) || string.charAt(i2 - -4) != '@') break block56;
                                                                                                    if (string.substring(i2 - -1, i2 - -4).equalsIgnoreCase(Yb[38])) break block57;
                                                                                                    if (!string.substring(i2 - -1, i2 - -4).equalsIgnoreCase(Yb[28])) break block58;
                                                                                                    n5 = 16748608;
                                                                                                    if (!bl) break block59;
                                                                                                }
                                                                                                if (string.substring(i2 - -1, 4 + i2).equalsIgnoreCase(Yb[37])) break block60;
                                                                                                if (string.substring(1 + i2, i2 + 4).equalsIgnoreCase(Yb[23])) break block61;
                                                                                                if (string.substring(i2 - -1, 4 + i2).equalsIgnoreCase(Yb[32])) break block62;
                                                                                                if (!string.substring(1 + i2, 4 + i2).equalsIgnoreCase(Yb[27])) break block63;
                                                                                                n5 = 65535;
                                                                                                if (!bl) break block59;
                                                                                            }
                                                                                            if (!string.substring(1 + i2, i2 + 4).equalsIgnoreCase(Yb[25])) break block64;
                                                                                            n5 = 0xFF00FF;
                                                                                            if (!bl) break block59;
                                                                                        }
                                                                                        if (string.substring(i2 + 1, 4 + i2).equalsIgnoreCase(Yb[33])) break block65;
                                                                                        if (string.substring(i2 + 1, i2 - -4).equalsIgnoreCase(Yb[35])) break block66;
                                                                                        if (!string.substring(1 + i2, i2 - -4).equalsIgnoreCase(Yb[31])) break block67;
                                                                                        n5 = 0xC00000;
                                                                                        if (!bl) break block59;
                                                                                    }
                                                                                    if (string.substring(1 + i2, 4 + i2).equalsIgnoreCase(Yb[41])) break block68;
                                                                                    if (string.substring(i2 - -1, 4 + i2).equalsIgnoreCase(Yb[39])) break block69;
                                                                                    if (!string.substring(1 + i2, 4 + i2).equalsIgnoreCase(Yb[24])) break block70;
                                                                                    n5 = 0xFFB000;
                                                                                    if (!bl) break block59;
                                                                                }
                                                                                if (string.substring(i2 + 1, 4 + i2).equalsIgnoreCase(Yb[40])) break block71;
                                                                                if (string.substring(i2 + 1, i2 - -4).equalsIgnoreCase(Yb[22])) break block72;
                                                                                if (!string.substring(i2 + 1, 4 + i2).equalsIgnoreCase(Yb[29])) break block73;
                                                                                n5 = 0xC0FF00;
                                                                                if (!bl) break block59;
                                                                            }
                                                                            if (!string.substring(1 + i2, 4 + i2).equalsIgnoreCase(Yb[34])) break block74;
                                                                            n5 = 0x80FF00;
                                                                            if (!bl) break block59;
                                                                        }
                                                                        if (!string.substring(i2 - -1, 4 + i2).equalsIgnoreCase(Yb[36])) break block59;
                                                                        n5 = 0x40FF00;
                                                                        if (!bl) break block59;
                                                                    }
                                                                    n5 = 0xFF3000;
                                                                    if (!bl) break block59;
                                                                }
                                                                n5 = 0xFF7000;
                                                                if (!bl) break block59;
                                                            }
                                                            n5 = (int)(1.6777215E7 * Math.random());
                                                            if (!bl) break block59;
                                                        }
                                                        n5 = 16748608;
                                                        if (!bl) break block59;
                                                    }
                                                    n5 = 0;
                                                    if (!bl) break block59;
                                                }
                                                n5 = 0xFFFFFF;
                                                if (!bl) break block59;
                                            }
                                            n5 = 255;
                                            if (!bl) break block59;
                                        }
                                        n5 = 65280;
                                        if (!bl) break block59;
                                    }
                                    n5 = 0xFFFF00;
                                    if (!bl) break block59;
                                }
                                n5 = 0xFF0000;
                            }
                            i2 += 4;
                            if (!bl) continue;
                        }
                        if ('\uffffff81' == ~string.charAt(i2) && ~string.length() < ~(4 + i2) && '~' == string.charAt(4 + i2)) {
                            n12 = string.charAt(1 + i2);
                            n11 = string.charAt(2 + i2);
                            char c2 = string.charAt(i2 + 3);
                            if (-49 >= ~n12 && ~n12 >= -58 && n11 >= 48 && 57 >= n11 && c2 >= '0' && c2 <= '9') {
                                n4 = Integer.parseInt(string.substring(i2 + 1, i2 - -4));
                            }
                            i2 += 4;
                            if (!bl) continue;
                        }
                        if ((n12 = string.charAt(i2)) == 160) {
                            n12 = 32;
                        }
                        if (~n12 > -1 || n.a.length <= n12) {
                            n12 = 32;
                        }
                        n11 = n.a[n12];
                        if (this.xb && !fb.k[n6] && 0 != n5) {
                            this.a((byte)53, fb.k[n6], byArray, 1 + n4, 0, n11, n3);
                        }
                        if (this.xb && !fb.k[n6] && ~n5 != -1) {
                            this.a((byte)101, fb.k[n6], byArray, n4, 0, n11, n3 - -1);
                        }
                        this.a((byte)73, fb.k[n6], byArray, n4, n5, n11, n3);
                        n4 += byArray[n11 + 7];
                        if (!bl) continue;
                        break;
                    }
                }
                catch (Exception exception) {
                    System.out.println(Yb[30] + exception);
                    exception.printStackTrace();
                }
                n8 = w;
                n7 = 1;
            }
            w = n8 + n7;
            return;
        }
        catch (RuntimeException runtimeException) {
            String string2;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[26]).append(n2).append(',').append(n3).append(',');
            if (string != null) {
                string2 = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(n4).append(',').append(n5).append(',').append(by).append(',').append(n6).append(')').toString());
            }
            string2 = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(n4).append(',').append(n5).append(',').append(by).append(',').append(n6).append(')').toString());
        }
    }

    @Override
    public final synchronized void removeConsumer(ImageConsumer imageConsumer) {
        try {
            if (this.fb == imageConsumer) {
                this.fb = null;
            }
            ++p;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[42]);
            String string = imageConsumer != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    private final synchronized void b(boolean bl) {
        try {
            ++N;
            if (null == this.fb) {
                return;
            }
            this.fb.setPixels(0, 0, this.u, this.k, this.nb, this.rb, 0, this.u);
            this.fb.imageComplete(2);
            if (!bl) {
                this.startProduction(null);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[21] + bl + ')');
        }
    }

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    final void b(int n2, int n3, int n4, int n5, int n6) {
        boolean bl = client.vh;
        try {
            block12: {
                block11: {
                    ++j;
                    if (~n2 > ~this.hb) return;
                    if (~this.lb >= ~n2) return;
                    if (~this.A < ~n3) {
                        n5 -= this.A + -n3;
                        n3 = this.A;
                    }
                    if (~this.Rb > ~(n3 + n5)) break block11;
                    break block12;
                }
                n5 = -n3 + this.Rb;
            }
            if (n5 <= n6) {
                return;
            }
            int n7 = n2 - -(this.u * n3);
            int n8 = 0;
            while (~n8 > ~n5) {
                try {
                    this.rb[n7 - -(this.u * n8)] = n4;
                    ++n8;
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
            throw i.a(runtimeException, Yb[43] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, int n3, int n4, int n5, int n6) {
        try {
            int n7;
            if (this.Qb[n2]) {
                n4 += this.Sb[n2];
                n6 += this.G[n2];
            }
            ++sb;
            int n8 = this.u * n6 + n4;
            int n9 = n3;
            int n10 = this.R[n2];
            int n11 = this.kb[n2];
            int n12 = this.u - n11;
            int n13 = 0;
            if (~this.A < ~n6) {
                n7 = -n6 + this.A;
                n6 = this.A;
                n10 -= n7;
                n9 += n7 * n11;
                n8 += n7 * this.u;
            }
            if (~this.Rb >= ~(n10 + n6)) {
                n10 -= 1 + (n10 + (n6 + -this.Rb));
            }
            if (n4 < this.hb) {
                n7 = this.hb + -n4;
                n12 += n7;
                n13 += n7;
                n8 += n7;
                n9 += n7;
                n4 = this.hb;
                n11 -= n7;
            }
            if (this.lb <= n4 + n11) {
                n7 = -this.lb + (n4 - (-n11 - 1));
                n12 += n7;
                n11 -= n7;
                n13 += n7;
            }
            if (~n11 >= -1) return;
            if (~n10 >= -1) return;
            n7 = 1;
            if (!(!this.i)) {
                n13 += this.kb[n2];
                n7 = 2;
                if ((1 & n6) != 0) {
                    --n10;
                    n8 += this.u;
                }
                n12 += this.u;
            }
            if (this.ob[n2] == null) {
                this.a(n9, n12, this.gb[n2], n7, false, n10, n13, n5, n11, this.rb, this.Y[n2], n8);
                if (!client.vh) return;
            }
            this.a(n9, n10, n8, this.ob[n2], 0, n5, n7, this.rb, -107, n13, n12, n11);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[19] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    final void a(Graphics graphics, int n2, int n3, int n4) {
        try {
            ++e;
            this.b(true);
            graphics.drawImage(this.Gb, n2, n4, this);
            if (n3 != 256) {
                Kb = null;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[49]);
            String string = graphics != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, byte by, int n3, int n4, int n5, int n6, int n7) {
        try {
            ++I;
            try {
                int n8;
                int n9;
                int n10;
                int n15 = this.kb[n2];
                int n16 = this.R[n2];
                int n14 = 0;
                int n13 = 0;
                int n12 = (n15 << 2133607216) / n5;
                int n11 = (n16 << -1116806288) / n3;
                if (this.Qb[n2]) {
                    n10 = this.Eb[n2];
                    n9 = this.qb[n2];
                    if (n10 == 0) return;
                    if (-1 == ~n9) return;
                    n11 = (n9 << 1837802000) / n3;
                    n6 += (n9 + n3 * this.G[n2] + -1) / n9;
                    n4 += (n10 + this.Sb[n2] * n5 + -1) / n10;
                    n12 = (n10 << -405443120) / n5;
                    if (-1 != ~(n5 * this.Sb[n2] % n10)) {
                        n14 = (-(this.Sb[n2] * n5 % n10) + n10 << 1444777936) / n5;
                    }
                    if (-1 != ~(this.G[n2] * n3 % n9)) {
                        n13 = (n9 + -(this.G[n2] * n3 % n9) << -1176347888) / n3;
                    }
                    n5 = n5 * (this.kb[n2] - (n14 >> 7993200)) / n10;
                    n3 = (-(n13 >> 826090000) + this.R[n2]) * n3 / n9;
                }
                n10 = n6 * this.u + n4;
                if (by > -121) {
                    return;
                }
                if (~n6 > ~this.A) {
                    n8 = -n6 + this.A;
                    n3 -= n8;
                    n6 = 0;
                    n10 += this.u * n8;
                    n13 += n11 * n8;
                }
                n9 = this.u + -n5;
                if (~this.hb < ~n4) {
                    n8 = -n4 + this.hb;
                    n4 = 0;
                    n14 += n8 * n12;
                    n10 += n8;
                    n5 -= n8;
                    n9 += n8;
                }
                if (~this.Rb >= ~(n6 + n3)) {
                    n3 -= 1 + n3 + (n6 + -this.Rb);
                }
                if (~this.lb >= ~(n4 + n5)) {
                    n8 = 1 + n4 + (n5 - this.lb);
                    n9 += n8;
                    n5 -= n8;
                }
                n8 = 1;
                if (!(!this.i)) {
                    n11 += n11;
                    n9 += this.u;
                    if (0 != (n6 & 1)) {
                        n10 += this.u;
                        --n3;
                    }
                    n8 = 2;
                }
                this.a(n8, n13, n5, (byte)-61, n11, n15, n12, n3, n10, this.ob[n2], 0, n14, n9, n7, this.rb);
                return;
            }
            catch (Exception exception) {
                System.out.println(Yb[16]);
                return;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[54] + n2 + ',' + by + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ')');
        }
    }

    /*
     * Unable to fully structure code
     */
    private final void a(int var1_1, int[] var2_2, int var3_3, byte var4_4, int var5_5, int var6_6, int var7_7, int var8_8, byte[] var9_9, int var10_10) {
        var14_11 = client.vh;
        try {
            block37: {
                ++ua.g;
                try {
                    var11_12 = -(var7_7 >> -1201305182);
                    var7_7 = -(var7_7 & 3);
                    var12_15 = -var6_6;
                    while (~var12_15 > -1) {
                        block42: {
                            v0 = var11_12;
                            if (var14_11) break block37;
                            var13_16 = v0;
                            while (~var13_16 > -1) {
                                block41: {
                                    block47: {
                                        block40: {
                                            block46: {
                                                block39: {
                                                    block45: {
                                                        block38: {
                                                            v1 = 0;
                                                            v2 = var9_9[var8_8++];
                                                            if (var14_11) ** GOTO lbl79
                                                            if (v1 != v2) {
                                                            }
                                                            ** GOTO lbl25
                                                            var2_2[var3_3++] = var1_1;
                                                            if (!var14_11) break block38;
lbl25:
                                                            // 2 sources

                                                            ++var3_3;
                                                        }
                                                        if (~var9_9[var8_8++] != -1) ** GOTO lbl38
                                                        ++var3_3;
                                                        if (!var14_11) break block39;
                                                        break block45;
                                                        catch (Exception v5) {
                                                            throw v5;
                                                        }
                                                    }
                                                    var2_2[var3_3++] = var1_1;
                                                }
                                                if (0 != var9_9[var8_8++]) ** GOTO lbl52
                                                ++var3_3;
                                                if (!var14_11) break block40;
                                                break block46;
                                                catch (Exception v7) {
                                                    throw v7;
                                                }
                                            }
                                            var2_2[var3_3++] = var1_1;
                                        }
                                        if (~var9_9[var8_8++] == -1) ** GOTO lbl66
                                        var2_2[var3_3++] = var1_1;
                                        if (!var14_11) break block41;
                                        break block47;
                                        catch (Exception v9) {
                                            throw v9;
                                        }
                                    }
                                    ++var3_3;
                                }
                                ++var13_16;
                                if (!var14_11) continue;
                            }
                            var13_16 = var7_7;
                            do {
                                block43: {
                                    block48: {
                                        v1 = ~var13_16;
                                        v2 = -1;
lbl79:
                                        // 3 sources

                                        if (v1 <= v2) break;
                                        v11 = var9_9[var8_8++];
                                        if (var14_11) break block42;
                                        break block48;
                                        catch (Exception v12) {
                                            throw v12;
                                        }
                                    }
                                    if (v11 == 0) {
                                    }
                                    ** GOTO lbl95
                                    ++var3_3;
                                    if (!var14_11) break block43;
lbl95:
                                    // 2 sources

                                    var2_2[var3_3++] = var1_1;
                                }
                                ++var13_16;
                            } while (!var14_11);
                            var8_8 += var10_10;
                            v11 = var3_3 + var5_5;
                        }
                        var3_3 = v11;
                        ++var12_15;
                        if (!var14_11) continue;
                        break;
                    }
                }
                catch (Exception var11_13) {
                    System.out.println(ua.Yb[14] + var11_13);
                    var11_13.printStackTrace();
                }
                v0 = 82 % ((-45 - var4_4) / 48);
            }
            var11_12 = v0;
        }
        catch (RuntimeException var11_14) {
            v15 = var11_14;
            v16 = new StringBuilder().append(ua.Yb[13]).append(var1_1).append(',');
            v17 = var2_2 != null ? ua.Yb[1] : ua.Yb[0];
            v19 = v16.append(v17).append(',').append(var3_3).append(',').append(var4_4).append(',').append(var5_5).append(',').append(var6_6).append(',').append(var7_7).append(',').append(var8_8).append(',');
            v20 = var9_9 != null ? ua.Yb[1] : ua.Yb[0];
            throw i.a(v15, v19.append(v20).append(',').append(var10_10).append(')').toString());
        }
    }

    /*
     * Unable to fully structure code
     */
    private final void a(int var1_1, int[] var2_2, int var3_3, int var4_4, int[] var5_5, int var6_6, int var7_7, boolean var8_8, byte[] var9_9, int var10_10, int var11_11) {
        var16_12 = client.vh;
        try {
            if (!var8_8) {
                this.i = true;
            }
            ++ua.Ub;
            var12_13 = -(var10_10 >> -1243049534);
            var10_10 = -(3 & var10_10);
            var13_15 = -var7_7;
            while (~var13_15 > -1 && !var16_12) {
                block39: {
                    for (var14_16 = var12_13; 0 > var14_16; ++var14_16) {
                        block44: {
                            block38: {
                                block43: {
                                    block37: {
                                        block42: {
                                            block36: {
                                                var15_17 = var9_9[var3_3++];
                                                v0 = ~var15_17;
                                                v1 = -1;
                                                if (var16_12) ** GOTO lbl77
                                                if (v0 == v1) {
                                                }
                                                ** GOTO lbl24
                                                ++var1_1;
                                                if (!var16_12) break block36;
lbl24:
                                                // 2 sources

                                                var5_5[var1_1++] = var2_2[ib.a(var15_17, 255)];
                                            }
                                            var15_17 = var9_9[var3_3++];
                                            if (var15_17 != 0) ** GOTO lbl38
                                            ++var1_1;
                                            if (!var16_12) break block37;
                                            break block42;
                                            catch (RuntimeException v4) {
                                                throw v4;
                                            }
                                        }
                                        var5_5[var1_1++] = var2_2[ib.a(var15_17, 255)];
                                    }
                                    var15_17 = var9_9[var3_3++];
                                    if (~var15_17 == -1) ** GOTO lbl53
                                    var5_5[var1_1++] = var2_2[ib.a(255, var15_17)];
                                    if (!var16_12) break block38;
                                    break block43;
                                    catch (RuntimeException v6) {
                                        throw v6;
                                    }
                                }
                                ++var1_1;
                            }
                            var15_17 = var9_9[var3_3++];
                            if (var15_17 == 0) ** GOTO lbl68
                            var5_5[var1_1++] = var2_2[ib.a(var15_17, 255)];
                            if (!var16_12) continue;
                            break block44;
                            catch (RuntimeException v8) {
                                throw v8;
                            }
                        }
                        ++var1_1;
                        continue;
                    }
                    var14_16 = var10_10;
                    do {
                        block40: {
                            v0 = 0;
                            v1 = var14_16;
lbl77:
                            // 2 sources

                            if (v0 <= v1) break;
                            var15_17 = var9_9[var3_3++];
                            v10 = 0;
                            v11 = var15_17;
                            if (var16_12) break block39;
                            if (v10 != v11) {
                            }
                            ** GOTO lbl91
                            var5_5[var1_1++] = var2_2[ib.a(var15_17, 255)];
                            if (!var16_12) break block40;
lbl91:
                            // 2 sources

                            ++var1_1;
                        }
                        ++var14_16;
                    } while (!var16_12);
                    var1_1 += var11_11;
                    var3_3 += var4_4;
                    v10 = var13_15;
                    v11 = var6_6;
                }
                var13_15 = v10 + v11;
                if (!var16_12) continue;
                break;
            }
        }
        catch (RuntimeException var12_14) {
            v14 = var12_14;
            v15 = new StringBuilder().append(ua.Yb[82]).append(var1_1).append(',');
            v16 = var2_2 != null ? ua.Yb[1] : ua.Yb[0];
            v18 = v15.append(v16).append(',').append(var3_3).append(',').append(var4_4).append(',');
            v19 = var5_5 != null ? ua.Yb[1] : ua.Yb[0];
            v21 = v18.append(v19).append(',').append(var6_6).append(',').append(var7_7).append(',').append(var8_8).append(',');
            v22 = var9_9 != null ? ua.Yb[1] : ua.Yb[0];
            throw i.a(v14, v21.append(v22).append(',').append(var10_10).append(',').append(var11_11).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void b(int n2, int n3) {
        boolean bl = client.vh;
        try {
            int n4;
            int[] nArray;
            block12: {
                ++b;
                if (null == this.gb[n2]) {
                    return;
                }
                int n5 = this.kb[n2] * this.R[n2];
                byte[] byArray = this.gb[n2];
                int[] nArray2 = this.Y[n2];
                nArray = new int[n5];
                int n6 = 0;
                while (~n6 > ~n5) {
                    int n7;
                    block14: {
                        block13: {
                            n4 = n7 = nArray2[0xFF & byArray[n6]];
                            if (bl) break block12;
                            if (n4 != 0) break block13;
                            n7 = 1;
                            if (!bl) break block14;
                        }
                        if (-16711936 == ~n7) {
                            n7 = 0;
                        }
                    }
                    nArray[n6] = n7;
                    ++n6;
                    if (!bl) continue;
                }
                n4 = n3;
            }
            if (n4 != -342059728) {
                this.Eb = null;
            }
            this.ob[n2] = nArray;
            this.gb[n2] = null;
            this.Y[n2] = null;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[77] + n2 + ',' + n3 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(int n2, int n3, int n4, int n5, int n6, int n7) {
        boolean bl;
        block120: {
            bl = client.vh;
            try {
                int n8;
                int n9;
                int n10;
                int[] nArray;
                int n11;
                int n12;
                int n13;
                int n14;
                int n15;
                block119: {
                    int n16;
                    int n17;
                    int n18;
                    int n19;
                    int n20;
                    int n21;
                    int n22;
                    int n23;
                    block137: {
                        int n24;
                        int n25;
                        int n26;
                        int n27;
                        int n28;
                        block136: {
                            int n29;
                            int n30;
                            int n31;
                            int n32;
                            int n33;
                            block118: {
                                int n34;
                                int n35;
                                int n36;
                                int n37;
                                block135: {
                                    block134: {
                                        int n38;
                                        int n39;
                                        int n40;
                                        block117: {
                                            block133: {
                                                int n41;
                                                int n42;
                                                int n43;
                                                block132: {
                                                    block116: {
                                                        int n44;
                                                        block131: {
                                                            block130: {
                                                                block129: {
                                                                    block128: {
                                                                        block127: {
                                                                            block126: {
                                                                                block125: {
                                                                                    block124: {
                                                                                        block123: {
                                                                                            block121: {
                                                                                                block122: {
                                                                                                    ++S;
                                                                                                    n15 = this.u;
                                                                                                    n23 = this.k;
                                                                                                    if (this.Hb == null) {
                                                                                                        this.Hb = new int[512];
                                                                                                        for (n43 = 0; n43 < 256; ++n43) {
                                                                                                            this.Hb[n43] = (int)(Math.sin((double)n43 * 0.02454369) * 32768.0);
                                                                                                            this.Hb[256 + n43] = (int)(Math.cos((double)n43 * 0.02454369) * 32768.0);
                                                                                                            if (!bl) {
                                                                                                                if (!bl) continue;
                                                                                                            }
                                                                                                            break;
                                                                                                        }
                                                                                                    } else {
                                                                                                        n43 = -this.Eb[n2] / 2;
                                                                                                    }
                                                                                                    n44 = -this.qb[n2] / 2;
                                                                                                    if (this.Qb[n2]) {
                                                                                                        n43 += this.Sb[n2];
                                                                                                        n44 += this.G[n2];
                                                                                                    }
                                                                                                    n28 = this.kb[n2] + n43;
                                                                                                    n27 = this.R[n2] + n44;
                                                                                                    n40 = n28;
                                                                                                    n37 = n44;
                                                                                                    n33 = n43;
                                                                                                    n32 = n27;
                                                                                                    int n45 = this.Hb[n7 &= 0xFF] * n6;
                                                                                                    int n46 = this.Hb[n7 + 256] * n6;
                                                                                                    n42 = n4 + (n46 * n43 + n44 * n45 >> 2041339190);
                                                                                                    n41 = n3 - -(-(n43 * n45) + n44 * n46 >> -1101614986);
                                                                                                    n36 = (n40 * n46 + n37 * n45 >> -1412824714) + n4;
                                                                                                    n35 = n3 + (n46 * n37 - n40 * n45 >> -1277097834);
                                                                                                    n26 = (n45 * n27 + n46 * n28 >> 1894119030) + n4;
                                                                                                    n25 = (-(n45 * n28) + n46 * n27 >> 195973654) + n3;
                                                                                                    n31 = (n33 * n46 + n45 * n32 >> -352508522) + n4;
                                                                                                    n24 = n3 - -(n32 * n46 - n33 * n45 >> -223287050);
                                                                                                    if (~n6 == -193 && ~(0x3F & client.ef) == ~(0x3F & n7)) break block121;
                                                                                                    if (~n6 != -129) break block122;
                                                                                                    client.ef = n7;
                                                                                                    if (!bl) break block123;
                                                                                                }
                                                                                                ++da.M;
                                                                                                if (!bl) break block123;
                                                                                            }
                                                                                            ++nb.g;
                                                                                        }
                                                                                        n14 = n41;
                                                                                        n13 = n41;
                                                                                        if (n14 <= n35) break block124;
                                                                                        n14 = n35;
                                                                                        if (!bl) break block125;
                                                                                    }
                                                                                    if (~n13 > ~n35) {
                                                                                        n13 = n35;
                                                                                    }
                                                                                }
                                                                                if (~n25 <= ~n14) break block126;
                                                                                n14 = n25;
                                                                                if (!bl) break block127;
                                                                            }
                                                                            if (n25 > n13) {
                                                                                n13 = n25;
                                                                            }
                                                                        }
                                                                        if (~n24 <= ~n14) break block128;
                                                                        n14 = n24;
                                                                        if (!bl) break block129;
                                                                    }
                                                                    if (~n13 > ~n24) {
                                                                        n13 = n24;
                                                                    }
                                                                }
                                                                if (~n14 > ~this.A) {
                                                                    n14 = this.A;
                                                                }
                                                                if (this.Xb == null || ~this.Xb.length != ~(n23 + 1)) {
                                                                    this.tb = new int[n23 + 1];
                                                                    this.M = new int[1 + n23];
                                                                    this.t = new int[n23 + 1];
                                                                    this.Tb = new int[1 + n23];
                                                                    this.Wb = new int[1 + n23];
                                                                    this.Xb = new int[n23 - -1];
                                                                }
                                                                if (~this.Rb > ~n13) {
                                                                    n13 = this.Rb;
                                                                }
                                                                n22 = n14;
                                                                while (~n13 <= ~n22) {
                                                                    this.Xb[n22] = 99999999;
                                                                    this.t[n22] = -99999999;
                                                                    ++n22;
                                                                    if (!bl) continue;
                                                                }
                                                                n21 = 0;
                                                                n20 = 0;
                                                                n34 = 0;
                                                                n12 = this.kb[n2];
                                                                n28 = -1 + n12;
                                                                int n47 = this.R[n2];
                                                                n40 = n12 + -1;
                                                                n43 = 0;
                                                                n44 = 0;
                                                                n37 = 0;
                                                                n33 = 0;
                                                                n27 = n47 + -1;
                                                                n32 = -1 + n47;
                                                                if (n41 <= n24) break block130;
                                                                n19 = n41;
                                                                n18 = n32 << 840930536;
                                                                n22 = n24;
                                                                n17 = n31 << -731169944;
                                                                if (!bl) break block131;
                                                            }
                                                            n19 = n24;
                                                            n22 = n41;
                                                            n18 = n44 << 813501160;
                                                            n17 = n42 << 1063907752;
                                                        }
                                                        if (~n24 != ~n41) {
                                                            n34 = (-n44 + n32 << 822762216) / (n24 + -n41);
                                                            n21 = (-n42 + n31 << -1960375352) / (n24 - n41);
                                                        }
                                                        if (-1 < ~n22) {
                                                            n17 -= n21 * n22;
                                                            n18 -= n22 * n34;
                                                            n22 = 0;
                                                        }
                                                        if (~n19 < ~(-1 + n23)) {
                                                            n19 = n23 + -1;
                                                        }
                                                        if (n5 != 842218000) {
                                                            return;
                                                        }
                                                        n11 = n22;
                                                        while (~n11 >= ~n19) {
                                                            this.Xb[n11] = this.t[n11] = n17;
                                                            n17 += n21;
                                                            nArray = this.M;
                                                            this.Tb[n11] = 0;
                                                            n10 = n11;
                                                            nArray[n10] = 0;
                                                            this.tb[n11] = this.Wb[n11] = n18;
                                                            n18 += n34;
                                                            ++n11;
                                                            if (!bl) {
                                                                if (!bl) continue;
                                                            }
                                                            break block116;
                                                        }
                                                        if (n41 != n35) {
                                                            n21 = (n36 + -n42 << -1951284024) / (-n41 + n35);
                                                            n20 = (-n43 + n40 << 1234137512) / (-n41 + n35);
                                                        }
                                                    }
                                                    if (n35 >= n41) break block132;
                                                    n16 = n40 << -2081746392;
                                                    n19 = n41;
                                                    n22 = n35;
                                                    n17 = n36 << 1385760008;
                                                    if (!bl) break block133;
                                                }
                                                n17 = n42 << 743454440;
                                                n22 = n41;
                                                n16 = n43 << 705115528;
                                                n19 = n35;
                                            }
                                            if (n23 + -1 < n19) {
                                                n19 = n23 - 1;
                                            }
                                            if (0 > n22) {
                                                n16 -= n20 * n22;
                                                n17 -= n21 * n22;
                                                n22 = 0;
                                            }
                                            for (n11 = n22; n19 >= n11; n17 += n21, n16 += n20, ++n11) {
                                                n39 = ~n17;
                                                n38 = ~this.Xb[n11];
                                                if (!bl) {
                                                    if (n39 > n38) {
                                                        this.Xb[n11] = n17;
                                                        this.M[n11] = n16;
                                                        this.tb[n11] = 0;
                                                    }
                                                    if (~this.t[n11] <= ~n17) continue;
                                                    this.t[n11] = n17;
                                                    this.Tb[n11] = n16;
                                                    this.Wb[n11] = 0;
                                                    if (!bl) continue;
                                                }
                                                break block117;
                                            }
                                            n39 = ~n35;
                                            n38 = ~n25;
                                        }
                                        if (n39 < n38) break block134;
                                        n16 = n40 << -1289051512;
                                        n18 = n37 << -1525813880;
                                        n22 = n35;
                                        n17 = n36 << -952809592;
                                        n19 = n25;
                                        if (!bl) break block135;
                                    }
                                    n16 = n28 << -794260248;
                                    n19 = n35;
                                    n18 = n27 << -606906680;
                                    n17 = n26 << -1631775192;
                                    n22 = n25;
                                }
                                if (n35 != n25) {
                                    n34 = (-n37 + n27 << -665843768) / (-n35 + n25);
                                    n21 = (-n36 + n26 << -740697688) / (n25 + -n35);
                                }
                                if (n19 > n23 + -1) {
                                    n19 = n23 + -1;
                                }
                                if (~n22 > -1) {
                                    n17 -= n21 * n22;
                                    n18 -= n34 * n22;
                                    n22 = 0;
                                }
                                n11 = n22;
                                while (~n19 <= ~n11) {
                                    n30 = ~n17;
                                    n29 = ~this.Xb[n11];
                                    if (!bl) {
                                        if (n30 > n29) {
                                            this.Xb[n11] = n17;
                                            this.M[n11] = n16;
                                            this.tb[n11] = n18;
                                        }
                                        if (n17 > this.t[n11]) {
                                            this.t[n11] = n17;
                                            this.Tb[n11] = n16;
                                            this.Wb[n11] = n18;
                                        }
                                        n17 += n21;
                                        n18 += n34;
                                        ++n11;
                                        if (!bl) continue;
                                    }
                                    break block118;
                                }
                                n30 = ~n25;
                                n29 = ~n24;
                            }
                            if (n30 != n29) {
                                n21 = (n31 + -n26 << -1317172024) / (n24 + -n25);
                                n20 = (n33 + -n28 << -236389848) / (-n25 + n24);
                            }
                            if (~n25 >= ~n24) break block136;
                            n16 = n33 << 45836488;
                            n19 = n25;
                            n22 = n24;
                            n18 = n32 << 1400553864;
                            n17 = n31 << 2055980168;
                            if (!bl) break block137;
                        }
                        n22 = n25;
                        n18 = n27 << 222482728;
                        n17 = n26 << -523743608;
                        n16 = n28 << -171150040;
                        n19 = n24;
                    }
                    if (n22 < 0) {
                        n17 -= n22 * n21;
                        n16 -= n22 * n20;
                        n22 = 0;
                    }
                    if (n23 + -1 < n19) {
                        n19 = -1 + n23;
                    }
                    for (n11 = n22; n19 >= n11; n17 += n21, n16 += n20, ++n11) {
                        n9 = ~this.Xb[n11];
                        n8 = ~n17;
                        if (!bl) {
                            if (n9 < n8) {
                                this.Xb[n11] = n17;
                                this.M[n11] = n16;
                                this.tb[n11] = n18;
                            }
                            if (~this.t[n11] <= ~n17) continue;
                            this.t[n11] = n17;
                            this.Tb[n11] = n16;
                            this.Wb[n11] = n18;
                            if (!bl) continue;
                        }
                        break block119;
                    }
                    n9 = n14;
                    n8 = n15;
                }
                n11 = n9 * n8;
                nArray = this.ob[n2];
                n10 = n14;
                while (~n13 < ~n10) {
                    block139: {
                        block140: {
                            int n48;
                            int n49;
                            int n50;
                            int n51;
                            int n52;
                            int n53;
                            block141: {
                                block138: {
                                    n53 = this.Xb[n10] >> -299165016;
                                    n52 = this.t[n10] >> -1707577976;
                                    if (bl) break block120;
                                    if (~(n52 + -n53) < -1) break block138;
                                    n11 += n15;
                                    if (!bl) break block139;
                                }
                                n51 = this.M[n10] << -122904183;
                                n50 = ((this.Tb[n10] << 1728443241) - n51) / (n52 - n53);
                                n49 = this.tb[n10] << 1699710857;
                                n48 = (-n49 + (this.Wb[n10] << -470306615)) / (-n53 + n52);
                                if (~this.lb > ~n52) {
                                    n52 = this.lb;
                                }
                                if (n53 < this.hb) {
                                    n49 += (-n53 + this.hb) * n48;
                                    n51 += (this.hb + -n53) * n50;
                                    n53 = this.hb;
                                }
                                if (!(!this.i) && ~(n10 & 1) != -1) break block140;
                                if (!this.Qb[n2]) break block141;
                                this.a(n48, -n52 + n53, n51, nArray, this.rb, n49, n53 + n11, n12, n50, 0, (byte)102);
                                if (!bl) break block140;
                            }
                            this.a(-n52 + n53, n12, 0, this.rb, n50, n49, n51, nArray, n48, n53 + n11, true);
                        }
                        n11 += n15;
                    }
                    ++n10;
                    if (!bl) continue;
                    break;
                }
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, Yb[79] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ')');
            }
        }
        if (e.Ab != 0) {
            client.vh = !bl;
        }
    }

    final void d(int n2, int n3, int n4, int n5, int n6, int n7) {
        boolean bl = client.vh;
        try {
            this.kb[n2] = n5;
            ++Jb;
            this.R[n2] = n3;
            if (n4 <= 108) {
                return;
            }
            this.Qb[n2] = false;
            this.Sb[n2] = 0;
            this.G[n2] = 0;
            this.Eb[n2] = n5;
            this.qb[n2] = n3;
            int n8 = n5 * n3;
            this.ob[n2] = new int[n8];
            int n9 = 0;
            int n10 = n6;
            while (~(n3 + n6) < ~n10 && !bl) {
                block8: {
                    for (int i2 = n7; i2 < n5 + n7; ++i2) {
                        this.ob[n2][n9++] = this.rb[this.u * n10 + i2];
                        if (!bl) {
                            if (!bl) continue;
                            break;
                        }
                        break block8;
                    }
                    ++n10;
                }
                if (!bl) continue;
                break;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[53] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int n3, int n4, int n5, int n6, int[] nArray, int n7, int[] nArray2, int n8, int n9, boolean bl, int n10, int n11, int n12, int n13) {
        boolean bl2 = client.vh;
        try {
            ++vb;
            int n14 = (n12 & 0xFF5F75) >> -846694704;
            int n15 = 0xFF & n12 >> -1400710808;
            int n16 = 0xFF & n12;
            try {
                int n17 = n5;
                if (bl) {
                    this.Tb = null;
                }
                int n18 = -n11;
                do {
                    int n19;
                    int n20;
                    block22: {
                        if (~n18 <= -1) return;
                        int n21 = (n2 >> -324011088) * n9;
                        if (bl2) return;
                        int n22 = -n4;
                        while (-1 < ~n22) {
                            block26: {
                                block23: {
                                    block25: {
                                        int n23;
                                        int n24;
                                        int n25;
                                        block24: {
                                            n8 = nArray[n21 + (n5 >> 1856265008)];
                                            n20 = ~n8;
                                            n19 = -1;
                                            if (bl2) break block22;
                                            if (n20 == n19) break block23;
                                            n25 = n8 >> -1835533520 & 0xFF;
                                            n24 = (0xFF60 & n8) >> 1180773672;
                                            n23 = 0xFF & n8;
                                            if (~n24 == ~n25 && ~n24 == ~n23) break block24;
                                            nArray2[n7++] = n8;
                                            if (!bl2) break block25;
                                        }
                                        nArray2[n7++] = (n25 * n14 >> 1799197256 << -273346448) + ((n15 * n24 >> -1796300920 << 1450486152) + (n23 * n16 >> 835792776));
                                    }
                                    if (!bl2) break block26;
                                }
                                ++n7;
                            }
                            n5 += n3;
                            ++n22;
                            if (!bl2) continue;
                        }
                        n2 += n10;
                        n7 += n6;
                        n5 = n17;
                        n20 = n18;
                        n19 = n13;
                    }
                    n18 = n20 + n19;
                } while (!bl2);
                return;
            }
            catch (Exception exception) {
                System.out.println(Yb[47]);
                return;
            }
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[52]).append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(nArray != null ? Yb[1] : Yb[0]).append(',').append(n7).append(',');
            if (nArray2 != null) {
                string = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n8).append(',').append(n9).append(',').append(bl).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(')').toString());
            }
            string = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n8).append(',').append(n9).append(',').append(bl).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(')').toString());
        }
    }

    final void d(int n2, int n3) {
        try {
            this.D = n3;
            int n4 = -54 / ((-63 - n2) / 55);
            ++bb;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[87] + n2 + ',' + n3 + ')');
        }
    }

    private final void a(int n2, int n3, int n4, int n5, String string, int n6, int n7) {
        try {
            if (n2 != 11815) {
                return;
            }
            this.a(n5, n7, string, n6 - this.a(n4, 92, string) / 2, n3, (byte)-124, n4);
            ++zb;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[9]).append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',');
            String string2 = string != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n6).append(',').append(n7).append(')').toString());
        }
    }

    private final void a(int n2, int n3, int n4, byte by, int n5, int n6, int n7, int n8, int n9, int[] nArray, int n10, int n11, int n12, int n13, int[] nArray2) {
        block14: {
            boolean bl = client.vh;
            try {
                ++d;
                int n14 = 256 - n13;
                try {
                    int n15 = n11;
                    if (by != -61) {
                        return;
                    }
                    int n16 = -n8;
                    while (~n16 > -1) {
                        int n17;
                        int n18;
                        block15: {
                            int n19 = n6 * (n3 >> -330929776);
                            n3 += n5;
                            if (bl) break block14;
                            int n20 = -n4;
                            while (-1 < ~n20) {
                                block17: {
                                    block16: {
                                        n10 = nArray[n19 + (n11 >> -483039408)];
                                        n11 += n7;
                                        n18 = ~n10;
                                        n17 = -1;
                                        if (bl) break block15;
                                        if (n18 != n17) break block16;
                                        ++n9;
                                        if (!bl) break block17;
                                    }
                                    int n21 = nArray2[n9];
                                    nArray2[n9++] = ib.a(ib.a(65280, n21) * n14 + ib.a(65280, n10) * n13, 0xFF0000) + ib.a(ib.a(n10, 0xFF00FF) * n13 + n14 * ib.a(0xFF00FF, n21), -16711936) >> -130221816;
                                }
                                ++n20;
                                if (!bl) continue;
                            }
                            n9 += n12;
                            n11 = n15;
                            n18 = n16;
                            n17 = n2;
                        }
                        n16 = n18 + n17;
                        if (!bl) continue;
                    }
                }
                catch (Exception exception) {
                    System.out.println(Yb[45]);
                }
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(Yb[46]).append(n2).append(',').append(n3).append(',').append(n4).append(',').append(by).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(n9).append(',');
                String string = nArray != null ? Yb[1] : Yb[0];
                StringBuilder stringBuilder2 = stringBuilder.append(string).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',');
                String string2 = nArray2 != null ? Yb[1] : Yb[0];
                throw i.a(runtimeException2, stringBuilder2.append(string2).append(')').toString());
            }
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int[] nArray, int n3, int n4, int n5, int n6, int n7, int[] nArray2, int n8, int n9, int n10, int n11, int n12, int n13, int n14, int n15) {
        boolean bl = client.vh;
        try {
            int n16;
            int n17;
            block23: {
                ++L;
                int n18 = (0xFF1712 & n15) >> -1742839120;
                int n19 = (n15 & 0xFFEE) >> -734006424;
                int n20 = 0xFF & n15;
                try {
                    int n21 = n6;
                    int n22 = -n8;
                    while (-1 < ~n22) {
                        int n23;
                        int n24;
                        int n25 = (n5 >> 1187405840) * n13;
                        int n26 = n12 >> -1195642352;
                        int n27 = n7;
                        n17 = n26;
                        n16 = this.hb;
                        if (bl) break block23;
                        if (n17 < n16) {
                            n24 = this.hb + -n26;
                            n26 = this.hb;
                            n27 -= n24;
                            n6 += n11 * n24;
                        }
                        n14 = -n14 + 1;
                        if (~this.lb >= ~(n26 - -n27)) {
                            n24 = n27 + n26 + -this.lb;
                            n27 -= n24;
                        }
                        if (0 != n14) {
                            for (n24 = n26; n24 < n27 + n26; n6 += n11, ++n24) {
                                n23 = n4 = nArray[n25 + (n6 >> -1996606992)];
                                if (!bl) {
                                    if (n23 == 0) continue;
                                    int n28 = n4 & 0xFF;
                                    int n29 = 0xFF & n4 >> -7992056;
                                    int n30 = (n4 & 0xFFE317) >> 1817608464;
                                    if (~n29 == ~n30 && n28 == n29) {
                                        nArray2[n10 + n24] = (n28 * n20 >> -1923074072) + (n19 * n29 >> -1585910200 << -1586887384) + (n18 * n30 >> 2107059944 << -1441257008);
                                        if (!bl) continue;
                                    }
                                    nArray2[n24 + n10] = n4;
                                    if (!bl) continue;
                                }
                                break;
                            }
                        } else {
                            n5 += n3;
                            n6 = n21;
                            n12 += n9;
                            n23 = n10 + this.u;
                        }
                        n10 = n23;
                        ++n22;
                        if (!bl) continue;
                        break;
                    }
                }
                catch (Exception exception) {
                    System.out.println(Yb[65]);
                }
                n17 = n2;
                n16 = 20;
            }
            if (n17 >= n16) return;
            this.t = null;
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[89]).append(n2).append(',').append(nArray != null ? Yb[1] : Yb[0]).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',');
            if (nArray2 != null) {
                string = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(',').append(n15).append(')').toString());
            }
            string = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(',').append(n15).append(')').toString());
        }
    }

    final int a(int n2, int n3) {
        try {
            block13: {
                block12: {
                    if (n2 != 508305352) {
                        this.G = null;
                    }
                    ++Pb;
                    if (n3 == 0) break block12;
                    break block13;
                }
                return 12;
            }
            if (n3 == 1) {
                return 14;
            }
            if (n3 == 2) {
                return 14;
            }
            if (-4 == ~n3) {
                return 15;
            }
            if (-5 == ~n3) {
                return 15;
            }
            if (~n3 == -6) {
                return 19;
            }
            if (6 == n3) {
                return 24;
            }
            if (7 == n3) {
                return 29;
            }
            return this.c(60, n3);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[74] + n2 + ',' + n3 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(byte[] byArray, int n2, int n3, int n4, int[] nArray, int n5, int n6, int n7, int n8, int n9, int[] nArray2, int n10, int n11, int n12, int n13, int n14, int n15) {
        boolean bl = client.vh;
        try {
            ++m;
            int n16 = (0xFF262A & n2) >> 1137456112;
            int n17 = (0xFF9D & n2) >> -1550693080;
            int n18 = 82 % ((n11 - -52) / 32);
            int n19 = n2 & 0xFF;
            try {
                int n20 = n12;
                int n21 = -n15;
                do {
                    int n22;
                    int n23;
                    block22: {
                        block23: {
                            int n24;
                            if (n21 >= 0) return;
                            int n25 = (n6 >> 160375440) * n5;
                            int n26 = n8 >> 698170864;
                            int n27 = n14;
                            if (bl) return;
                            if (~this.hb < ~n26) {
                                n24 = this.hb + -n26;
                                n26 = this.hb;
                                n27 -= n24;
                                n12 += n24 * n9;
                            }
                            if (~this.lb >= ~(n27 + n26)) {
                                n24 = n26 + (n27 - this.lb);
                                n27 -= n24;
                            }
                            n10 = 1 + -n10;
                            n6 += n7;
                            if (-1 == ~n10) break block23;
                            n24 = n26;
                            while (~(n27 + n26) < ~n24) {
                                block24: {
                                    int n28;
                                    int n29;
                                    int n30;
                                    block25: {
                                        n3 = byArray[(n12 >> -1390664752) - -n25] & 0xFF;
                                        n23 = -1;
                                        n22 = ~n3;
                                        if (bl) break block22;
                                        if (n23 == n22) break block24;
                                        n3 = nArray[n3];
                                        n30 = n3 & 0xFF;
                                        n29 = n3 >> 1006963688 & 0xFF;
                                        n28 = n3 >> -692748912 & 0xFF;
                                        if (~n29 == ~n28 && ~n29 == ~n30) break block25;
                                        nArray2[n13 + n24] = n3;
                                        if (!bl) break block24;
                                    }
                                    nArray2[n24 - -n13] = (n17 * n29 >> 436099816 << -1418519128) + (n16 * n28 >> -2076941880 << 547107024) - -(n19 * n30 >> -2023878008);
                                }
                                n12 += n9;
                                ++n24;
                                if (!bl) continue;
                            }
                        }
                        n8 += n4;
                        n23 = n13;
                        n22 = this.u;
                    }
                    n13 = n23 + n22;
                    n12 = n20;
                    ++n21;
                } while (!bl);
                return;
            }
            catch (Exception exception) {
                System.out.println(Yb[65]);
                return;
            }
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[67]).append(byArray != null ? Yb[1] : Yb[0]).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(nArray != null ? Yb[1] : Yb[0]).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(n9).append(',');
            if (nArray2 != null) {
                string = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(',').append(n15).append(')').toString());
            }
            string = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(',').append(n15).append(')').toString());
        }
    }

    private final void a(int n2, int n3, String string, int n4, int n5, int n6, int n7) {
        try {
            this.a(n7, n3, string, n2 - this.a(n6, 114, string), n4, (byte)123, n6);
            ++H;
            if (n5 != -12200) {
                this.b(75, -128, -127, 3, 49, -124);
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[5]).append(n2).append(',').append(n3).append(',');
            String string2 = string != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(byte[] byArray, int n2, int n3, int n4, int n5, int n6, int n7, int n8, int[] nArray, int n9) {
        boolean bl = client.vh;
        try {
            ++Fb;
            if (n7 != 1504725224) {
                this.rb = null;
            }
            int n10 = -n5;
            do {
                int n11;
                int n12;
                block14: {
                    if (-1 >= ~n10) return;
                    if (bl) return;
                    for (int i2 = -n3; i2 < 0; ++i2) {
                        int n13 = 0xFF & byArray[n9++];
                        n12 = -31;
                        n11 = ~n13;
                        if (!bl) {
                            if (n12 <= n11) {
                                ++n4;
                                if (!bl) continue;
                            }
                            if (-231 < ~n13) {
                                int n14 = nArray[n4];
                                nArray[n4++] = ib.a(-16711936, ib.a(0xFF00FF, n2) * n13 - -(ib.a(n14, 0xFF00FF) * (-n13 + 256))) - -ib.a((256 - n13) * ib.a(65280, n14) + n13 * ib.a(65280, n2), 0xFF0000) >> 1504725224;
                                if (!bl) continue;
                            }
                            nArray[n4++] = n2;
                            if (!bl) continue;
                        }
                        break block14;
                    }
                    n9 += n6;
                    n12 = n4;
                    n11 = n8;
                }
                n4 = n12 + n11;
                ++n10;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[64]).append(byArray != null ? Yb[1] : Yb[0]).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',');
            if (nArray != null) {
                string = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n9).append(')').toString());
            }
            string = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n9).append(')').toString());
        }
    }

    final void a(int n2, String string, int n3, int n4, int n5, int n6) {
        try {
            this.a(11815, n3, n5, n4, string, n2, n6);
            ++ab;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[90]).append(n2).append(',');
            String string2 = string != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(')').toString());
        }
    }

    final void a(byte by, byte[] byArray, int n2) {
        block26: {
            boolean bl = client.vh;
            try {
                int n3;
                int n4;
                int n5;
                int n6;
                int n7;
                int[] nArray;
                block21: {
                    ++Cb;
                    this.ob[n2] = new int[10200];
                    nArray = this.ob[n2];
                    this.kb[n2] = 255;
                    this.R[n2] = 40;
                    this.Sb[n2] = 0;
                    this.G[n2] = 0;
                    if (by != -118) {
                        this.a(82, -105, -7, 8, 9);
                    }
                    this.Eb[n2] = 255;
                    this.qb[n2] = 40;
                    this.Qb[n2] = false;
                    int n8 = 0;
                    n7 = 1;
                    n6 = 0;
                    while (n6 < 255) {
                        block22: {
                            n5 = 0xFF & byArray[n7++];
                            n4 = 0;
                            if (bl) break block21;
                            n3 = n4;
                            while (~n5 < ~n3) {
                                nArray[n6++] = n8;
                                ++n3;
                                if (!bl) {
                                    if (!bl) continue;
                                    break;
                                }
                                break block22;
                            }
                            n8 = -n8 + 0xFFFFFF;
                        }
                        if (!bl) continue;
                    }
                    n4 = 1;
                }
                n5 = n4;
                block14: while (true) {
                    int n9 = ~n5;
                    block15: while (n9 > -41 && !bl) {
                        n3 = 0;
                        while (~n3 > -256) {
                            block25: {
                                block23: {
                                    block24: {
                                        int n10 = byArray[n7++] & 0xFF;
                                        n9 = 0;
                                        if (bl) continue block15;
                                        int n11 = n9;
                                        while (~n11 > ~n10) {
                                            nArray[n6] = nArray[-255 + n6];
                                            ++n6;
                                            ++n3;
                                            ++n11;
                                            if (!bl) {
                                                if (!bl) continue;
                                                break;
                                            }
                                            break block23;
                                        }
                                        if (~n3 > -256) break block24;
                                        break block25;
                                    }
                                    nArray[n6] = 0xFFFFFF - nArray[n6 - 255];
                                    ++n6;
                                }
                                ++n3;
                            }
                            if (!bl) continue;
                        }
                        ++n5;
                        if (!bl) continue block14;
                    }
                    break block26;
                    break;
                }
                {
                    break block26;
                    break;
                }
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(Yb[4]).append(by).append(',');
                String string = byArray != null ? Yb[1] : Yb[0];
                throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(')').toString());
            }
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(boolean bl, int n2) {
        boolean bl2 = client.vh;
        try {
            int n3;
            int n4;
            int n5;
            int n6;
            int[] nArray;
            int[] nArray2;
            int[] nArray3;
            int n7;
            block23: {
                int[] nArray4;
                block22: {
                    ++Ob;
                    n7 = this.R[n2] * this.kb[n2];
                    nArray3 = this.ob[n2];
                    nArray2 = new int[32768];
                    if (bl) {
                        this.a(-63, 58, -7, -36, -99);
                    }
                    for (int i2 = 0; i2 < n7; ++i2) {
                        int n8 = nArray3[i2];
                        nArray4 = nArray2;
                        if (!bl2) {
                            int n9 = (0x1F & n8 >> 454314147) + ((n8 >> 400635145 & 0x7C00) + ((n8 & 0xF800) >> 303743686));
                            nArray4[n9] = nArray4[n9] + 1;
                            if (!bl2) continue;
                        }
                        break block22;
                    }
                    nArray4 = new int[256];
                }
                nArray = nArray4;
                nArray[0] = 0xFF00FF;
                int[] nArray5 = new int[256];
                int n10 = 0;
                block9: while (true) {
                    int n11 = n10;
                    int n12 = 32768;
                    block10: while (n11 < n12) {
                        n6 = nArray2[n10];
                        n5 = nArray5[255];
                        if (bl2) break block23;
                        if (n5 < n6) {
                            n4 = 1;
                            while (-257 < ~n4) {
                                n11 = nArray5[n4];
                                n12 = n6;
                                if (bl2) continue block10;
                                if (n11 < n12) {
                                    block24: {
                                        for (n3 = 255; n4 < n3; --n3) {
                                            nArray[n3] = nArray[-1 + n3];
                                            nArray5[n3] = nArray5[n3 - 1];
                                            if (!bl2) {
                                                if (!bl2) continue;
                                            }
                                            break block24;
                                        }
                                        nArray[n4] = 263172 + ((ib.a(31, n10) << 257025667) + ib.a(63488, n10 << -1695574842) + ib.a(0xF80000, n10 << 986275945));
                                        nArray5[n4] = n6;
                                    }
                                    if (!bl2) break;
                                }
                                ++n4;
                                if (!bl2) continue;
                            }
                        }
                        nArray2[n10] = -1;
                        ++n10;
                        if (!bl2) continue block9;
                    }
                    break;
                }
                n5 = n7;
            }
            byte[] byArray = new byte[n5];
            n6 = 0;
            block13: while (true) {
                int n13 = ~n7;
                int n14 = ~n6;
                block14: while (n13 < n14) {
                    n4 = nArray3[n6];
                    n3 = (n4 >> 1087745987 & 0x1F) + ((n4 & 0xF800) >> -1963155290) + ((0xF80000 & n4) >> 10645225);
                    int n15 = nArray2[n3];
                    if (bl2) return;
                    if (-1 == n15) {
                        int n16 = 999999999;
                        int n17 = 0xFF & n4 >> 100065008;
                        int n18 = 0xFF & n4 >> 862974792;
                        int n19 = n4 & 0xFF;
                        for (int i3 = 0; 256 > i3; ++i3) {
                            int n20;
                            int n21 = nArray[i3];
                            int n22 = (0xFF1E98 & n21) >> -112833712;
                            int n23 = n21 >> 1743095144 & 0xFF;
                            int n24 = 0xFF & n21;
                            n13 = n20 = (-n24 + n19) * (-n24 + n19) + ((n17 - n22) * (n17 - n22) - -((-n23 + n18) * (-n23 + n18)));
                            n14 = n16;
                            if (bl2) continue block14;
                            if (n13 >= n14) continue;
                            n15 = i3;
                            n16 = n20;
                            if (!bl2) continue;
                        }
                        nArray2[n3] = n15;
                    }
                    byArray[n6] = (byte)n15;
                    ++n6;
                    if (!bl2) continue block13;
                }
                break;
            }
            this.gb[n2] = byArray;
            this.Y[n2] = nArray;
            this.ob[n2] = null;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[44] + bl + ',' + n2 + ')');
        }
    }

    final void a(int n2, String string, int n3, int n4, int n5, int n6, boolean bl, int n7) {
        boolean bl2 = client.vh;
        try {
            ++Q;
            try {
                int n8;
                int n9;
                block17: {
                    int n10 = 0;
                    byte[] byArray = m.b[n5];
                    if (n4 < 44) {
                        return;
                    }
                    n9 = 0;
                    int n11 = 0;
                    for (int i2 = 0; string.length() > i2; ++i2) {
                        block21: {
                            block19: {
                                block20: {
                                    n8 = -65;
                                    if (bl2) break block17;
                                    if (n8 == ~string.charAt(i2) && ~string.length() < ~(4 + i2) && -65 == ~string.charAt(i2 - -4)) break block19;
                                    if (-127 == ~string.charAt(i2) && string.length() > 4 + i2 && ~string.charAt(4 + i2) == -127) break block20;
                                    int n12 = string.charAt(i2);
                                    if (-1 < ~n12 || ~n12 <= ~n.a.length) {
                                        n12 = 32;
                                    }
                                    n10 += byArray[7 + n.a[n12]];
                                    if (!bl2) break block21;
                                }
                                i2 += 4;
                                if (!bl2) break block21;
                            }
                            i2 += 4;
                        }
                        if (string.charAt(i2) == ' ') {
                            n11 = i2;
                        }
                        if (string.charAt(i2) == '%' && bl) {
                            n10 = 1000;
                            n11 = i2;
                        }
                        if (n10 <= n2) continue;
                        if (n11 <= n9) {
                            n11 = i2;
                        }
                        n10 = 0;
                        this.a(11815, n7, n5, 0, string.substring(n9, n11), n3, n6);
                        n9 = i2 = 1 + n11;
                        n6 += this.a(508305352, n5);
                        if (!bl2) continue;
                    }
                    n8 = n10;
                }
                if (n8 > 0) {
                    this.a(11815, n7, n5, 0, string.substring(n9), n3, n6);
                }
            }
            catch (Exception exception) {
                System.out.println(Yb[81] + exception);
                exception.printStackTrace();
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[80]).append(n2).append(',');
            String string2 = string != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(bl).append(',').append(n7).append(')').toString());
        }
    }

    final void a(int n2, int n3, int n4, boolean bl, int n5, int n6, int n7, int n8, int n9, int n10) {
        block30: {
            boolean bl2 = client.vh;
            try {
                ++y;
                try {
                    int n11;
                    int n12;
                    int n13;
                    int n14;
                    int n15;
                    int n16;
                    int n17;
                    int n18;
                    int n19;
                    block28: {
                        block29: {
                            block27: {
                                block26: {
                                    block25: {
                                        if (n4 == 0) {
                                            n4 = 0xFFFFFF;
                                        }
                                        if (-1 == ~n3) break block25;
                                        break block26;
                                    }
                                    n3 = 0xFFFFFF;
                                }
                                n19 = this.kb[n6];
                                int n20 = this.R[n6];
                                n18 = 0;
                                n17 = 0;
                                n16 = n5 << -822406960;
                                n15 = (n19 << 264332976) / n8;
                                n14 = (n20 << -1079220976) / n7;
                                n13 = -(n5 << -1507499888) / n7;
                                if (this.Qb[n6]) break block27;
                                break block28;
                            }
                            n12 = this.Eb[n6];
                            n11 = this.qb[n6];
                            if (n12 != 0 && ~n11 != -1) break block29;
                            return;
                        }
                        n15 = (n12 << -179525200) / n8;
                        n14 = (n11 << 842218000) / n7;
                        int n21 = this.Sb[n6];
                        if (bl) {
                            n21 = n12 + -this.kb[n6] - n21;
                        }
                        int n22 = this.G[n6];
                        n9 += (-1 + (n12 + n21 * n8)) / n12;
                        int n23 = (-1 + (n22 * n7 + n11)) / n11;
                        if (n21 * n8 % n12 != 0) {
                            n18 = (-(n8 * n21 % n12) + n12 << 306741872) / n8;
                        }
                        n2 += n23;
                        n16 += n23 * n13;
                        if (0 != n22 * n7 % n11) {
                            n17 = (n11 - n7 * n22 % n11 << -894050704) / n7;
                        }
                        n8 = (n15 + ((this.kb[n6] << -1406651696) - (n18 + 1))) / n15;
                        n7 = ((this.R[n6] << -1596145424) + -n17 - (-n14 + 1)) / n14;
                    }
                    n12 = this.u * n2;
                    n16 += n9 << 189764144;
                    if (n2 < this.A) {
                        n11 = this.A + -n2;
                        n16 += n13 * n11;
                        n7 -= n11;
                        n17 += n11 * n14;
                        n12 += this.u * n11;
                        n2 = this.A;
                    }
                    if (n2 - -n7 >= this.Rb) {
                        n7 -= 1 + (n2 + n7) - this.Rb;
                    }
                    n11 = n12 / this.u & n10;
                    if (!this.i) {
                        n11 = 2;
                    }
                    if (~n4 == -16777216) {
                        if (null != this.ob[n6]) {
                            if (bl) {
                                this.a(n10 ^ 0x4A, this.ob[n6], n14, 0, n17, (this.kb[n6] << 102617264) + (-n18 - 1), n8, this.rb, n7, n13, n12, -n15, n16, n19, n11, n3);
                                if (!bl2) break block30;
                            }
                            this.a(n10 + 89, this.ob[n6], n14, 0, n17, n18, n8, this.rb, n7, n13, n12, n15, n16, n19, n11, n3);
                            if (!bl2) break block30;
                        }
                        if (!bl) {
                            this.a(this.gb[n6], n3, 0, n13, this.Y[n6], n19, n17, n14, n16, n15, this.rb, n11, -110, n18, n12, n8, n7);
                            if (!bl2) break block30;
                        }
                        this.a(this.gb[n6], n3, 0, n13, this.Y[n6], n19, n17, n14, n16, -n15, this.rb, n11, n10 ^ 0xFFFFFF84, -1 + (-n18 + (this.kb[n6] << -1997207152)), n12, n8, n7);
                        if (!bl2) break block30;
                    }
                    if (this.ob[n6] == null) {
                        if (bl) {
                            this.a(n7, this.Y[n6], n19, n14, 0, n13, n12, n17, this.gb[n6], n11, (byte)76, -n18 + ((this.kb[n6] << -1651772048) + -1), n4, n16, -n15, n8, this.rb, n3);
                            if (!bl2) break block30;
                        }
                        this.a(n7, this.Y[n6], n19, n14, 0, n13, n12, n17, this.gb[n6], n11, (byte)78, n18, n4, n16, n15, n8, this.rb, n3);
                        if (!bl2) break block30;
                    }
                    if (bl) {
                        this.a(this.rb, this.ob[n6], n8, n13, n16, n10 + 1603920391, 0, n4, n14, -n15, -n18 + (this.kb[n6] << -1212875536) + -1, n11, n17, n19, n3, n7, n12);
                        if (!bl2) break block30;
                    }
                    this.a(this.rb, this.ob[n6], n8, n13, n16, 1603920392, 0, n4, n14, n15, n18, n11, n17, n19, n3, n7, n12);
                }
                catch (Exception exception) {
                    System.out.println(Yb[16]);
                }
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, Yb[71] + n2 + ',' + n3 + ',' + n4 + ',' + bl + ',' + n5 + ',' + n6 + ',' + n7 + ',' + n8 + ',' + n9 + ',' + n10 + ')');
            }
        }
    }

    @Override
    public final void startProduction(ImageConsumer imageConsumer) {
        try {
            this.addConsumer(imageConsumer);
            ++mb;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[12]);
            String string = imageConsumer != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    @Override
    public final synchronized boolean isConsumer(ImageConsumer imageConsumer) {
        try {
            ++q;
            return this.fb == imageConsumer;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[2]);
            String string = imageConsumer != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int[] nArray, int n3, int n4, int n5, int n6, byte by, int n7, int[] nArray2, int n8, int n9) {
        boolean bl = client.vh;
        try {
            ++x;
            if (by <= 122) {
                this.e(121, 54, -117, -34, 67, -103);
            }
            int n10 = -(n2 >> -1677003518);
            n2 = -(3 & n2);
            int n11 = -n4;
            do {
                int n12;
                int n13;
                block33: {
                    int n14;
                    block32: {
                        int n15;
                        if (~n11 <= -1) return;
                        if (bl) return;
                        for (n14 = n10; n14 < 0; ++n14) {
                            block39: {
                                block38: {
                                    block37: {
                                        block36: {
                                            block35: {
                                                block34: {
                                                    n15 = n5 = nArray2[n6++];
                                                    if (bl) break block32;
                                                    if (n15 != 0) break block34;
                                                    ++n7;
                                                    if (!bl) break block35;
                                                }
                                                nArray[n7++] = n5;
                                            }
                                            if (~(n5 = nArray2[n6++]) != -1) break block36;
                                            ++n7;
                                            if (!bl) break block37;
                                        }
                                        nArray[n7++] = n5;
                                    }
                                    if (0 == (n5 = nArray2[n6++])) break block38;
                                    nArray[n7++] = n5;
                                    if (!bl) break block39;
                                }
                                ++n7;
                            }
                            if (0 == (n5 = nArray2[n6++])) {
                                ++n7;
                                if (!bl) continue;
                            }
                            nArray[n7++] = n5;
                            if (!bl) continue;
                        }
                        n15 = n14 = n2;
                    }
                    while (-1 < ~n14) {
                        block41: {
                            block40: {
                                n5 = nArray2[n6++];
                                n13 = 0;
                                n12 = n5;
                                if (bl) break block33;
                                if (n13 != n12) break block40;
                                ++n7;
                                if (!bl) break block41;
                            }
                            nArray[n7++] = n5;
                        }
                        ++n14;
                        if (!bl) continue;
                    }
                    n6 += n9;
                    n7 += n8;
                    n13 = n11;
                    n12 = n3;
                }
                n11 = n13 + n12;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[8]).append(n2).append(',').append(nArray != null ? Yb[1] : Yb[0]).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(by).append(',').append(n7).append(',');
            if (nArray2 != null) {
                string = Yb[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n8).append(',').append(n9).append(')').toString());
            }
            string = Yb[0];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n8).append(',').append(n9).append(')').toString());
        }
    }

    final void a(int n2, int n3, int n4, int n5, byte by) {
        try {
            block14: {
                block13: {
                    block12: {
                        block11: {
                            if (this.k < n4) {
                                n4 = this.k;
                            }
                            if (0 > n5) {
                                n5 = 0;
                            }
                            if (-1 < ~n2) break block11;
                            break block12;
                        }
                        n2 = 0;
                    }
                    ++n;
                    if (~this.u > ~n3) break block13;
                    break block14;
                }
                n3 = this.u;
            }
            if (by <= 15) {
                C = 109;
            }
            this.A = n5;
            this.Rb = n4;
            this.lb = n3;
            this.hb = n2;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[6] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + by + ')');
        }
    }

    final void a(boolean bl) {
        block14: {
            boolean bl2 = client.vh;
            try {
                int n2;
                ++f;
                int n3 = this.k * this.u;
                boolean bl3 = bl;
                boolean bl4 = !this.i;
                if (bl3 == bl4) {
                    for (n2 = 0; n3 > n2; ++n2) {
                        this.rb[n2] = 0;
                        if (!bl2) {
                            if (!bl2) continue;
                            break;
                        }
                        break block14;
                    }
                    if (!bl2) break block14;
                }
                n2 = 0;
                int n4 = -this.k;
                while (~n4 > -1 && !bl2) {
                    block15: {
                        for (int i2 = -this.u; 0 > i2; ++i2) {
                            this.rb[n2++] = 0;
                            if (!bl2) {
                                if (!bl2) continue;
                                break;
                            }
                            break block15;
                        }
                        n2 += this.u;
                        n4 += 2;
                    }
                    if (!bl2) continue;
                    break;
                }
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, Yb[51] + bl + ')');
            }
        }
    }

    final void a(int n2, byte by, int n3, int n4, int n5, int n6) {
        boolean bl = client.vh;
        try {
            block22: {
                block21: {
                    block20: {
                        block19: {
                            block18: {
                                block17: {
                                    if (n2 < this.hb) {
                                        n6 -= this.hb + -n2;
                                        n2 = this.hb;
                                    }
                                    if (this.A > n4) break block17;
                                    break block18;
                                }
                                n5 -= -n4 + this.A;
                                n4 = this.A;
                            }
                            ++K;
                            if (~(n4 + n5) < ~this.Rb) break block19;
                            break block20;
                        }
                        n5 = -n4 + this.Rb;
                    }
                    if (~(n6 + n2) < ~this.lb) break block21;
                    break block22;
                }
                n6 = -n2 + this.lb;
            }
            int n7 = this.u + -n6;
            int n8 = 1;
            int n9 = -124 / ((-39 - by) / 59);
            if (this.i) {
                block24: {
                    block23: {
                        n7 += this.u;
                        if (0 != (n4 & 1)) break block23;
                        break block24;
                    }
                    --n5;
                    ++n4;
                }
                n8 = 2;
            }
            int n10 = n2 - -(this.u * n4);
            int n11 = -n5;
            while (n11 < 0 && !bl) {
                block25: {
                    int n12 = -n6;
                    while (-1 < ~n12) {
                        this.rb[n10++] = n3;
                        ++n12;
                        if (!bl) {
                            if (!bl) continue;
                            break;
                        }
                        break block25;
                    }
                    n10 += n7;
                    n11 += n8;
                }
                if (!bl) continue;
                break;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[3] + n2 + ',' + by + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    final void a(String string, int n2, int n3, int n4, boolean bl, int n5) {
        try {
            ++pb;
            this.a(0, n3, string, n2, n4, (byte)124, n5);
            if (bl) {
                this.a(-43, 36, -60, -88, 93, null, 114, null, -53, 59, true, 66, 34, 34, 70);
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(Yb[50]);
            String string2 = string != null ? Yb[1] : Yb[0];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(bl).append(',').append(n5).append(')').toString());
        }
    }

    final void a(int n2) {
        try {
            this.lb = this.u;
            if (n2 != -1) {
                this.Sb = null;
            }
            this.hb = 0;
            this.A = 0;
            this.Rb = this.k;
            ++W;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, Yb[83] + n2 + ')');
        }
    }

    /*
     * Exception decompiling
     */
    final void a(int var1_1, int var2_2, byte[] var3_3, int var4_4, byte[] var5_5) {
        /*
         * This method has failed to decompile.  When submitting a bug report, please provide this stack trace, and (if you hold appropriate legal rights) the relevant class file.
         * 
         * org.benf.cfr.reader.util.ConfusedCFRException: Tried to end blocks [26[DOLOOP]], but top level block is 31[FORLOOP]
         *     at org.benf.cfr.reader.bytecode.analysis.opgraph.Op04StructuredStatement.processEndingBlocks(Op04StructuredStatement.java:435)
         *     at org.benf.cfr.reader.bytecode.analysis.opgraph.Op04StructuredStatement.buildNestedBlocks(Op04StructuredStatement.java:484)
         *     at org.benf.cfr.reader.bytecode.analysis.opgraph.Op03SimpleStatement.createInitialStructuredBlock(Op03SimpleStatement.java:736)
         *     at org.benf.cfr.reader.bytecode.CodeAnalyser.getAnalysisInner(CodeAnalyser.java:850)
         *     at org.benf.cfr.reader.bytecode.CodeAnalyser.getAnalysisOrWrapFail(CodeAnalyser.java:278)
         *     at org.benf.cfr.reader.bytecode.CodeAnalyser.getAnalysis(CodeAnalyser.java:201)
         *     at org.benf.cfr.reader.entities.attributes.AttributeCode.analyse(AttributeCode.java:94)
         *     at org.benf.cfr.reader.entities.Method.analyse(Method.java:531)
         *     at org.benf.cfr.reader.entities.ClassFile.analyseMid(ClassFile.java:1055)
         *     at org.benf.cfr.reader.entities.ClassFile.analyseTop(ClassFile.java:942)
         *     at org.benf.cfr.reader.Driver.doJarVersionTypes(Driver.java:257)
         *     at org.benf.cfr.reader.Driver.doJar(Driver.java:139)
         *     at org.benf.cfr.reader.CfrDriverImpl.analyse(CfrDriverImpl.java:76)
         *     at org.benf.cfr.reader.Main.main(Main.java:54)
         */
        throw new IllegalStateException("Decompilation failed");
    }

    ua(int n2, int n3, int n4, Component component) {
        block10: {
            boolean bl = client.vh;
            this.xb = false;
            this.A = 0;
            this.i = false;
            this.hb = 0;
            this.lb = 0;
            this.Rb = 0;
            try {
                block11: {
                    this.Y = new int[n4][];
                    this.ob = new int[n4][];
                    this.Rb = n3;
                    this.gb = new byte[n4][];
                    this.Qb = new boolean[n4];
                    this.lb = n2;
                    this.Eb = new int[n4];
                    this.rb = new int[n2 * n3];
                    this.Sb = new int[n4];
                    this.G = new int[n4];
                    this.qb = new int[n4];
                    this.R = new int[n4];
                    this.kb = new int[n4];
                    this.k = n3;
                    this.u = n2;
                    if (~n2 >= -2) break block10;
                    if (n3 <= 1 || component == null) break block10;
                    this.nb = new DirectColorModel(32, 0xFF0000, 65280, 255);
                    int n5 = this.k * this.u;
                    int n6 = 0;
                    while (~n6 > ~n5) {
                        this.rb[n6] = 0;
                        ++n6;
                        if (!bl) {
                            if (!bl) continue;
                            break;
                        }
                        break block11;
                    }
                    this.Gb = component.createImage(this);
                    this.b(true);
                    component.prepareImage(this.Gb, component);
                    this.b(true);
                    component.prepareImage(this.Gb, component);
                    this.b(true);
                }
                component.prepareImage(this.Gb, component);
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(Yb[73]).append(n2).append(',').append(n3).append(',').append(n4).append(',');
                String string = component != null ? Yb[1] : Yb[0];
                throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
            }
        }
    }

    static {
        Yb = new String[]{ua.z(ua.z("QEIj")), ua.z(ua.z("D\u001e\u000b(t")), ua.z(ua.z("JQ\u000boz|_Ku|RUW.")), ua.z(ua.z("JQ\u000bJH\u0017")), ua.z(ua.z("JQ\u000bOH\u0017")), ua.z(ua.z("JQ\u000bIH\u0017")), ua.z(ua.z("JQ\u000bHH\u0017")), ua.z(ua.z("JQ\u000bS!")), ua.z(ua.z("JQ\u000bLK\u0017")), ua.z(ua.z("JQ\u000bA!")), ua.z(ua.z("JQ\u000b@!")), ua.z(ua.z("JQ\u000bM!")), ua.z(ua.z("JQ\u000bu}^BQV{PTPe}V_K.")), ua.z(ua.z("JQ\u000bVH\u0017")), ua.z(ua.z("O\\JreZDQc{\u0005\u0010")), ua.z(ua.z("JQ\u000bB!")), ua.z(ua.z("ZBWi{\u001fYK&zOBLrl\u001fSIoyOYKa)M_Pr`QU")), ua.z(ua.z("JQ\u000bW!")), ua.z(ua.z("JQ\u000bDK\u0017")), ua.z(ua.z("JQ\u000bR!")), ua.z(ua.z("JQ\u000bPH\u0017")), ua.z(ua.z("JQ\u000bEH\u0017")), ua.z(ua.z("PB\u0016")), ua.z(ua.z("XB@")), ua.z(ua.z("PB\u0014")), ua.z(ua.z("RQB")), ua.z(ua.z("JQ\u000b@K\u0017")), ua.z(ua.z("\\ID")), ua.z(ua.z("SB@")), ua.z(ua.z("XB\u0014")), ua.z(ua.z("[BDqzKBLhn\u0005\u0010")), ua.z(ua.z("[B@")), ua.z(ua.z("]\\P")), ua.z(ua.z("HXL")), ua.z(ua.z("XB\u0017")), ua.z(ua.z("]\\D")), ua.z(ua.z("XB\u0016")), ua.z(ua.z("FUI")), ua.z(ua.z("MUA")), ua.z(ua.z("MQK")), ua.z(ua.z("PB\u0017")), ua.z(ua.z("PBD")), ua.z(ua.z("JQ\u000btlR_ScJP^VsdZB\r")), ua.z(ua.z("JQ\u000bAK\u0017")), ua.z(ua.z("JQ\u000bBK\u0017")), ua.z(ua.z("ZBWi{\u001fYK&}MQKYz\\QIc")), ua.z(ua.z("JQ\u000bCH\u0017")), ua.z(ua.z("ZBWi{\u001fYK&yS_QYz\\QIc")), ua.z(ua.z("JQ\u000bBH\u0017")), ua.z(ua.z("JQ\u000bTH\u0017")), ua.z(ua.z("JQ\u000bDH\u0017")), ua.z(ua.z("JQ\u000bN!")), ua.z(ua.z("JQ\u000bCK\u0017")), ua.z(ua.z("JQ\u000bQ!")), ua.z(ua.z("JQ\u000bC!")), ua.z(ua.z("JQ\u000bP!")), ua.z(ua.z("JQ\u000bD!")), ua.z(ua.z("JQ\u000b@H\u0017")), ua.z(ua.z("JQ\u000btlNE@u}k_UBfH^icoKbLaaKb@ulQT\r")), ua.z(ua.z("ktiT")), ua.z(ua.z("JQ\u000bUH\u0017")), ua.z(ua.z("JQ\u000bEK\u0017")), ua.z(ua.z("JQ\u000bMK\u0017")), ua.z(ua.z("JQ\u000bU!")), ua.z(ua.z("JQ\u000bKH\u0017")), ua.z(ua.z("ZBWi{\u001fYK&}MQKuy^B@h}\u001fCUt`KU\u0005vePD\u0005tfJDLhl")), ua.z(ua.z("JQ\u000bGH\u0017")), ua.z(ua.z("JQ\u000bOK\u0017")), ua.z(ua.z("JQ\u000bK!")), ua.z(ua.z("JQ\u000bL!")), ua.z(ua.z("JQ\u000bT!")), ua.z(ua.z("JQ\u000bGK\u0017")), ua.z(ua.z("JQ\u000bJK\u0017")), ua.z(ua.z("JQ\u000b:`QYQ8!")), ua.z(ua.z("JQ\u000bWH\u0017")), ua.z(ua.z("JQ\u000bQH\u0017")), ua.z(ua.z("JQ\u000bJ!")), ua.z(ua.z("JQ\u000bV!")), ua.z(ua.z("JQ\u000bgm[sJhzJ]@t!")), ua.z(ua.z("JQ\u000bI!")), ua.z(ua.z("JQ\u000bNK\u0017")), ua.z(ua.z("\\UKr{Z@Dth\u0005\u0010")), ua.z(ua.z("JQ\u000bMH\u0017")), ua.z(ua.z("JQ\u000bLH\u0017")), ua.z(ua.z("JQ\u000bod^W@Sy[QQc!")), ua.z(ua.z("JQ\u000bKK\u0017")), ua.z(ua.z("JQ\u000bSH\u0017")), ua.z(ua.z("JQ\u000bO!")), ua.z(ua.z("JQ\u000bNH\u0017")), ua.z(ua.z("JQ\u000bAH\u0017")), ua.z(ua.z("JQ\u000bH!")), ua.z(ua.z("JQ\u000bRH\u0017"))};
        E = new v(ua.z(ua.z("sysC")), "", "", 0);
        C = 114;
        wb = new String[100];
        Kb = new String[]{ua.z(ua.z("z^Qc{\u001f^PkkZB\u0005io\u001fYQcdL\u0010Qi)MUHi\u007fZ\u0010Dhm\u001f@WczL\u0010@h}ZB"))};
        h = new String[200];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 9);
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
                        n4 = 48;
                        break;
                    }
                    case 2: {
                        n4 = 37;
                        break;
                    }
                    case 3: {
                        n4 = 6;
                        break;
                    }
                    default: {
                        n4 = 9;
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

