/*
 * Decompiled with CFR 0.152.
 */
final class lb {
    static int lb;
    static String[] ac;
    static int Jb;
    static int rb;
    static int z;
    static int ub;
    static int m;
    static int F;
    private int wb;
    private int Wb;
    private int Xb;
    private int[] H;
    private int[] J;
    static int V;
    static int d;
    private int nb;
    static int mb;
    static int w;
    private int[] Qb;
    private int ab;
    private n[] x;
    private int[] Vb;
    private int Zb;
    static int c;
    static int l;
    static int fb;
    private int xb;
    private int[][] i;
    private ua dc;
    private int[] jb;
    private int bc;
    private int[] Q;
    static int Y;
    private int o;
    private int e;
    private w[] y;
    private ca[] Ab;
    private int ib;
    private int u;
    static int U;
    private int[] pb;
    static int Sb;
    static int s;
    private int I;
    private int[] ob;
    int G;
    static int O;
    private int zb;
    private boolean f;
    private int[] yb;
    private int[] B;
    private int Kb;
    static int sb;
    static int k;
    int Mb;
    ca T;
    private int[] a;
    static int W;
    private int j;
    static int q;
    private byte[][] g;
    static int Yb;
    static int Rb;
    static int[] Tb;
    static int bb;
    private int[] Hb;
    private int[][] Ib;
    static int C;
    private int cc;
    private int[] qb;
    static int tb;
    private int[][] kb;
    private ca[] Z;
    static int E;
    private boolean K;
    static int Pb;
    static int hb;
    static int Db;
    private int[][] L;
    private int vb;
    int P;
    private int[][] ec;
    private int[] r;
    private boolean[] S;
    static int t;
    private int[] gb;
    private int[] Ob;
    private int b;
    static int M;
    static int Bb;
    private int[] Fb;
    private int Nb;
    private int cb;
    static int p;
    private boolean Ub;
    private int h;
    private int eb;
    private int A;
    static int Lb;
    private int Cb;
    int X;
    private int R;
    private int[] Eb;
    private int[] v;
    private long[] D;
    private int n;
    static int Gb;
    private int db;
    private static final String[] N;

    private final int a(int n2, boolean bl, int n3, int n4, int n5, int n6) {
        try {
            ++hb;
            if (bl) {
                this.a(-74, 87, -109, true);
            }
            if (n4 == n2) {
                return n6;
            }
            return (-n6 + n5) * (-n2 + n3) / (-n2 + n4) + n6;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[34] + n2 + ',' + bl + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ')');
        }
    }

    /*
     * Exception decompiling
     */
    private final void a(int var1_1, int var2_2, w[] var3_3, int var4_4) {
        /*
         * This method has failed to decompile.  When submitting a bug report, please provide this stack trace, and (if you hold appropriate legal rights) the relevant class file.
         * 
         * org.benf.cfr.reader.util.ConfusedCFRException: Tried to end blocks [7[DOLOOP]], but top level block is 9[DOLOOP]
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

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    final void a(int n2, int n3, int n4, int n5, int n6, int n7) {
        boolean bl = client.vh;
        try {
            block8: {
                if (~n5 == -1) {
                    if (n7 != 0 || n2 != 0) break block8;
                    n5 = 32;
                }
            }
            ++mb;
            for (int i2 = n4; i2 < this.ab; ++i2) {
                try {
                    this.Z[i2].a(n3, n6, n7, -115, n5, n2);
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
            throw i.a(runtimeException, N[32] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final boolean a(byte by, w w2, w w3) {
        boolean bl = client.vh;
        try {
            int[] nArray;
            int[] nArray2;
            int[] nArray3;
            int[] nArray4;
            block43: {
                int n2;
                int n3;
                int[] nArray5;
                ca ca2;
                block44: {
                    int n4;
                    int n5;
                    int n6;
                    int n7;
                    int n8;
                    int[] nArray6;
                    ca ca3;
                    block41: {
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
                        int n21;
                        int n22;
                        block40: {
                            ++z;
                            if (w3.e >= w2.m) return true;
                            if (by > -42) {
                                this.c(95, -1);
                            }
                            if (w2.e >= w3.m) {
                                return true;
                            }
                            if (~w3.h <= ~w2.j) {
                                return true;
                            }
                            if (~w2.h <= ~w3.j) {
                                return true;
                            }
                            if (w2.q <= w3.u) {
                                return true;
                            }
                            if (w3.q < w2.u) return false;
                            ca3 = w3.o;
                            ca2 = w2.o;
                            n22 = w3.i;
                            int n23 = w2.i;
                            nArray6 = ca3.o[n22];
                            nArray5 = ca2.o[n23];
                            n8 = ca3.lb[n22];
                            n7 = ca2.lb[n23];
                            n21 = ca2.cc[nArray5[0]];
                            n20 = ca2.H[nArray5[0]];
                            n19 = ca2.bb[nArray5[0]];
                            n18 = w2.r;
                            n17 = w2.l;
                            n16 = w2.k;
                            n15 = ca2.k[n23];
                            n14 = w2.s;
                            n13 = 0;
                            for (n12 = 0; n12 < n8; ++n12) {
                                n3 = nArray6[n12];
                                n11 = (-ca3.bb[n3] + n19) * n16 + ((n20 + -ca3.H[n3]) * n17 + n18 * (-ca3.cc[n3] + n21));
                                n10 = ~(-n15);
                                n9 = ~n11;
                                if (!bl) {
                                    if (!(n10 < n9 && n14 < 0 || ~n15 > ~n11 && 0 < n14)) continue;
                                    n13 = 1;
                                    if (!bl) break;
                                    if (!bl) continue;
                                }
                                break block40;
                            }
                            n10 = ~n13;
                            n9 = -1;
                        }
                        if (n10 == n9) return true;
                        n13 = 0;
                        n14 = w3.s;
                        n20 = ca3.H[nArray6[0]];
                        n21 = ca3.cc[nArray6[0]];
                        n15 = ca3.k[n22];
                        n19 = ca3.bb[nArray6[0]];
                        n17 = w3.l;
                        n16 = w3.k;
                        n18 = w3.r;
                        for (n12 = 0; n7 > n12; ++n12) {
                            n3 = nArray5[n12];
                            n11 = (-ca2.bb[n3] + n19) * n16 + ((n21 - ca2.cc[n3]) * n18 + (n20 - ca2.H[n3]) * n17);
                            n6 = -n15;
                            n5 = n11;
                            if (!bl) {
                                if (!(n6 > n5 && ~n14 < -1 || ~n11 < ~n15 && ~n14 > -1)) continue;
                                n13 = 1;
                                if (!bl) break;
                                if (!bl) continue;
                            }
                            break block41;
                        }
                        if (n13 == 0) return true;
                        n6 = ~n8;
                        n5 = -3;
                    }
                    if (n6 != n5) {
                        nArray4 = new int[n8];
                        nArray3 = new int[n8];
                        for (n2 = 0; n2 < n8; ++n2) {
                            n4 = nArray6[n2];
                            nArray4[n2] = ca3.pb[n4];
                            nArray3[n2] = ca3.Ob[n4];
                            if (!bl && !bl) continue;
                            break;
                        }
                    } else {
                        nArray4 = new int[4];
                        nArray3 = new int[4];
                        n3 = nArray6[1];
                        n2 = nArray6[0];
                        nArray4[0] = ca3.pb[n2] - 20;
                        nArray4[1] = ca3.pb[n3] + -20;
                        nArray4[2] = 20 + ca3.pb[n3];
                        nArray4[3] = ca3.pb[n2] + 20;
                        nArray3[0] = nArray3[3] = ca3.Ob[n2];
                        nArray3[1] = nArray3[2] = ca3.Ob[n3];
                    }
                    if (n7 == 2) break block44;
                    nArray2 = new int[n7];
                    nArray = new int[n7];
                    n2 = 0;
                    while (~n7 < ~n2) {
                        n4 = nArray5[n2];
                        nArray2[n2] = ca2.pb[n4];
                        nArray[n2] = ca2.Ob[n4];
                        ++n2;
                        if (!bl) {
                            if (!bl) continue;
                        }
                        break block43;
                    }
                    if (!bl) break block43;
                }
                nArray = new int[4];
                nArray2 = new int[4];
                n2 = nArray5[0];
                n3 = nArray5[1];
                nArray2[0] = -20 + ca2.pb[n2];
                nArray2[1] = -20 + ca2.pb[n3];
                nArray2[2] = ca2.pb[n3] + 20;
                nArray2[3] = ca2.pb[n2] - -20;
                nArray[0] = nArray[3] = ca2.Ob[n2];
                nArray[1] = nArray[2] = ca2.Ob[n3];
            }
            if (this.a(nArray2, nArray3, nArray4, nArray, 1)) return false;
            return true;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(N[2]).append(by).append(',').append(w2 != null ? N[0] : N[1]).append(',');
            if (w3 != null) {
                string = N[0];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = N[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void b(int n2, int n3) {
        boolean bl = client.vh;
        try {
            int n4;
            int n5;
            int n6;
            int n7;
            block25: {
                int n8;
                block24: {
                    int n9;
                    int n10;
                    int[] nArray;
                    ca ca2;
                    block23: {
                        block22: {
                            block21: {
                                block20: {
                                    block19: {
                                        block18: {
                                            int n11 = 47 / ((n2 - -36) / 37);
                                            ++s;
                                            w w2 = this.y[n3];
                                            ca2 = w2.o;
                                            int n12 = w2.i;
                                            nArray = ca2.o[n12];
                                            int n13 = 0;
                                            int n14 = 0;
                                            int n15 = 1;
                                            int n16 = ca2.cc[nArray[0]];
                                            int n17 = ca2.H[nArray[0]];
                                            int n18 = ca2.bb[nArray[0]];
                                            ca2.k[n12] = 1;
                                            ca2.M[n12] = 0;
                                            w2.l = n14;
                                            w2.r = n13;
                                            w2.k = n15;
                                            w2.s = n18 * n15 + (n16 * n13 + n17 * n14);
                                            n6 = n7 = ca2.bb[nArray[0]];
                                            n9 = n10 = ca2.pb[nArray[0]];
                                            if (~n10 < ~ca2.pb[nArray[1]]) break block18;
                                            n9 = ca2.pb[nArray[1]];
                                            if (!bl) break block19;
                                        }
                                        n10 = ca2.pb[nArray[1]];
                                    }
                                    n5 = ca2.Ob[nArray[1]];
                                    n4 = ca2.Ob[nArray[0]];
                                    n8 = ca2.bb[nArray[1]];
                                    if (~n6 > ~n8) break block20;
                                    if (n7 <= n8) break block21;
                                    n7 = n8;
                                    if (!bl) break block21;
                                }
                                n6 = n8;
                            }
                            if (n9 < (n8 = ca2.pb[nArray[1]])) break block22;
                            if (n10 <= n8) break block23;
                            n10 = n8;
                            if (!bl) break block23;
                        }
                        n9 = n8;
                    }
                    n8 = ca2.Ob[nArray[1]];
                    w2.m = n9 + 20;
                    w2.e = n10 - 20;
                    if (n4 >= n8) break block24;
                    n4 = n8;
                    if (!bl) break block25;
                }
                if (~n5 < ~n8) {
                    n5 = n8;
                }
            }
            w2.q = n6;
            w2.u = n7;
            w2.j = n4;
            w2.h = n5;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[29] + n2 + ',' + n3 + ')');
        }
    }

    final int[] a(byte by) {
        try {
            ++q;
            int n2 = 38 / ((35 - by) / 58);
            return this.qb;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[19] + by + ')');
        }
    }

    final void a(int n2, int n3, int n4, int n5, int n6, int n7, int n8, int n9) {
        try {
            int n10;
            int n11;
            int n12;
            block14: {
                int n13;
                int n14;
                int n15;
                block13: {
                    block12: {
                        block11: {
                            block10: {
                                block9: {
                                    n9 = 0x3FF & n9;
                                    n5 = 0x3FF & n5;
                                    n7 = 0x3FF & n7;
                                    ++U;
                                    this.b = 0x3FF & 1024 - n9;
                                    this.Kb = 0x3FF & -n5 + 1024;
                                    this.xb = 0x3FF & -n7 + 1024;
                                    n12 = 0;
                                    if (n6 != -12349) {
                                        return;
                                    }
                                    n11 = 0;
                                    n10 = n4;
                                    if (-1 != ~n5) break block9;
                                    break block10;
                                }
                                n15 = ba.cc[n5];
                                n14 = ba.cc[n5 - -1024];
                                n13 = n14 * n11 + -(n15 * n10) >> -583464977;
                                n10 = n15 * n11 + n10 * n14 >> -1062576145;
                                n11 = n13;
                            }
                            if (0 != n7) break block11;
                            break block12;
                        }
                        n15 = ba.cc[n7];
                        n14 = ba.cc[n7 - -1024];
                        n13 = n12 * n14 + n10 * n15 >> 1189002575;
                        n10 = n14 * n10 - n15 * n12 >> 1156686319;
                        n12 = n13;
                    }
                    if (n9 != 0) break block13;
                    break block14;
                }
                n14 = ba.cc[n9 + 1024];
                n15 = ba.cc[n9];
                n13 = n12 * n14 + n15 * n11 >> 1299787663;
                n11 = -(n15 * n12) + n11 * n14 >> -165343153;
                n12 = n13;
            }
            this.I = -n10 + n3;
            this.o = n8 + -n11;
            this.bc = n2 + -n12;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[14] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ',' + n8 + ',' + n9 + ')');
        }
    }

    final int a(int n2, boolean bl) {
        try {
            ++ub;
            if (n2 == 12345678) {
                return 0;
            }
            this.b(n2, bl);
            if (~n2 <= -1) {
                return this.kb[n2][0];
            }
            if (n2 < 0) {
                n2 = -(n2 + 1);
                int n3 = (n2 & 0x7D29) >> -819737878;
                int n4 = (0x3F4 & n2) >> 1842072837;
                int n5 = 0x1F & n2;
                return (n5 << -759569597) + ((n4 << -2070222453) + (n3 << -2036842829));
            }
            return 0;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[18] + n2 + ',' + bl + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final boolean a(int n2, boolean bl, int n3, byte by, int n4) {
        try {
            if (by != -71) {
                return false;
            }
            ++O;
            if (!(bl && ~n2 <= ~n3 || n3 < n2)) {
                if (~n2 >= ~n4) return bl;
                return true;
            }
            if (n2 < n4) return true;
            if (bl) return false;
            return true;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[37] + n2 + ',' + bl + ',' + n3 + ',' + by + ',' + n4 + ')');
        }
    }

    final void a(ca ca2, byte by) {
        try {
            if (by != 118) {
                this.a(false, (w)null, (w)null);
            }
            ++c;
            if (ca2 == null) {
                System.out.println(N[7]);
            }
            if (~this.u < ~this.ab) {
                this.jb[this.ab] = 0;
                this.Z[this.ab++] = ca2;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(N[8]);
            String string = ca2 != null ? N[0] : N[1];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(by).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final boolean a(boolean bl, w w2, w w3) {
        boolean bl2 = client.vh;
        try {
            int n2;
            int n3;
            block25: {
                int n4;
                int n5;
                int n6;
                int n7;
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
                int[] nArray;
                int[] nArray2;
                int n19;
                ca ca2;
                ca ca3;
                block24: {
                    ++C;
                    ca3 = w2.o;
                    ca2 = w3.o;
                    n19 = w2.i;
                    int n20 = w3.i;
                    nArray2 = ca3.o[n19];
                    nArray = ca2.o[n20];
                    int n21 = ca3.lb[n19];
                    n18 = ca2.lb[n20];
                    n17 = ca2.cc[nArray[0]];
                    n16 = ca2.H[nArray[0]];
                    n15 = ca2.bb[nArray[0]];
                    n14 = w3.r;
                    n13 = w3.l;
                    n12 = w3.k;
                    n11 = ca2.k[n20];
                    n10 = 0;
                    n9 = w3.s;
                    n8 = 0;
                    while (~n8 > ~n21) {
                        n7 = nArray2[n8];
                        n6 = n13 * (-ca3.H[n7] + n16) + (-ca3.cc[n7] + n17) * n14 - -((-ca3.bb[n7] + n15) * n12);
                        n5 = ~n6;
                        n4 = ~(-n11);
                        if (!bl2) {
                            if (n5 > n4 && n9 < 0 || n6 > n11 && -1 > ~n9) {
                                n10 = 1;
                                if (!bl2) break;
                            }
                            ++n8;
                            if (!bl2) continue;
                        }
                        break block24;
                    }
                    n5 = 0;
                    n4 = n10;
                }
                if (n5 == n4) return true;
                n17 = ca3.cc[nArray2[0]];
                n12 = w2.k;
                if (bl) {
                    this.y = null;
                }
                n11 = ca3.k[n19];
                n15 = ca3.bb[nArray2[0]];
                n16 = ca3.H[nArray2[0]];
                n10 = 0;
                n14 = w2.r;
                n13 = w2.l;
                n9 = w2.s;
                n8 = 0;
                while (~n18 < ~n8) {
                    n7 = nArray[n8];
                    n3 = n6 = n14 * (n17 - ca2.cc[n7]) - (-(n13 * (-ca2.H[n7] + n16)) + -((-ca2.bb[n7] + n15) * n12));
                    n2 = -n11;
                    if (!bl2) {
                        if (n3 < n2 && 0 < n9 || n11 < n6 && -1 < ~n9) {
                            n10 = 1;
                            if (!bl2) break;
                        }
                        ++n8;
                        if (!bl2) continue;
                    }
                    break block25;
                }
                n3 = -1;
                n2 = ~n10;
            }
            if (n3 != n2) return false;
            return true;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(N[41]).append(bl).append(',').append(w2 != null ? N[0] : N[1]).append(',');
            if (w3 != null) {
                string = N[0];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = N[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    private final void a(int n2, byte by) {
        boolean bl = client.vh;
        try {
            int n3;
            int n4;
            int n5;
            int[] nArray;
            block17: {
                int n6;
                int n7;
                block20: {
                    block19: {
                        ++tb;
                        if (-1 == ~this.Hb[n2]) break block19;
                        n7 = 128;
                        if (!bl) break block20;
                    }
                    n7 = 64;
                }
                nArray = this.kb[n2];
                n5 = 0;
                if (by != 118) {
                    return;
                }
                n4 = 0;
                block6: while (true) {
                    int n8 = ~n4;
                    int n9 = ~n7;
                    block7: while (n8 > n9) {
                        n6 = 0;
                        if (bl) break block17;
                        n3 = n6;
                        while (~n3 > ~n7) {
                            int n10;
                            block18: {
                                block21: {
                                    n10 = this.L[n2][this.g[n2][n3 + n4 * n7] & 0xFF];
                                    n8 = ~(n10 &= 0xF8F8FF);
                                    n9 = -1;
                                    if (bl) continue block7;
                                    if (n8 == n9) break block21;
                                    if (-16253184 != ~n10) break block18;
                                    this.S[n2] = true;
                                    n10 = 0;
                                    if (!bl) break block18;
                                }
                                n10 = 1;
                            }
                            nArray[n5++] = n10;
                            ++n3;
                            if (!bl) continue;
                        }
                        ++n4;
                        if (!bl) continue block6;
                    }
                    break;
                }
                n6 = n4 = 0;
            }
            while (~n5 < ~n4) {
                n3 = nArray[n4];
                try {
                    nArray[n5 - -n4] = ib.a(-(n3 >>> -2145788925) + n3, 0xF8F8FF);
                    nArray[n4 + 2 * n5] = ib.a(-(n3 >>> -709025758) + n3, 0xF8F8FF);
                    nArray[n4 + n5 * 3] = ib.a(-(n3 >>> -239044189) + n3 + -(n3 >>> -391782814), 0xF8F8FF);
                    ++n4;
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
            throw i.a(runtimeException, N[35] + n2 + ',' + by + ')');
        }
    }

    final int a(int n2, int n3, int n4, int n5, int n6, int n7, int n8, byte by) {
        try {
            ++t;
            this.gb[this.n] = n2;
            this.Fb[this.n] = n5;
            this.a[this.n] = n6;
            this.Ob[this.n] = n3;
            this.ob[this.n] = n7;
            this.Eb[this.n] = n8;
            this.Q[this.n] = 0;
            if (by != 109) {
                return -125;
            }
            int n9 = this.T.b(false, n3, n5, n6);
            int n10 = this.T.b(false, n3, n5, -n8 + n6);
            int[] nArray = new int[]{n9, n10};
            this.T.a(2, nArray, 0, 0, false);
            this.T.E[this.n] = n4;
            this.T.zb[this.n++] = 0;
            return this.n - 1;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[25] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ',' + n8 + ',' + by + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int n3, int n4, w[] wArray) {
        boolean bl = client.vh;
        try {
            int n5;
            block25: {
                if (n4 >= -50) {
                    this.a((byte)-98, 32);
                }
                for (n5 = 0; n5 <= n2; ++n5) {
                    wArray[n5].c = false;
                    wArray[n5].f = n5;
                    wArray[n5].p = -1;
                    if (!bl) {
                        if (!bl) continue;
                    }
                    break block25;
                }
                ++lb;
            }
            n5 = 0;
            block15: while (true) {
                if (wArray[n5].c) {
                    ++n5;
                    if (bl) return;
                    if (!bl) continue;
                }
                int n6 = ~n2;
                int n7 = ~n5;
                block16: while (true) {
                    if (n6 == n7) return;
                    w w2 = wArray[n5];
                    w2.c = true;
                    int n8 = n5;
                    int n9 = n5 - -n3;
                    if (n9 >= n2) {
                        n9 = n2 - 1;
                    }
                    int n10 = n9;
                    do {
                        if (~(1 + n8) < ~n10) continue block15;
                        w w3 = wArray[n10];
                        n6 = ~w2.e;
                        n7 = ~w3.m;
                        if (bl) continue block16;
                        if (n6 > n7 && ~w2.m < ~w3.e && w3.j > w2.h && ~w2.j < ~w3.h && ~w3.p != ~w2.f && !this.a((byte)-84, w3, w2) && this.a(false, w3, w2)) {
                            this.a(n8, wArray, n10, (byte)34);
                            n8 = this.e;
                            if (wArray[n10] != w3) {
                                ++n10;
                            }
                            w3.p = w2.f;
                        }
                        --n10;
                    } while (!bl);
                    break;
                }
            }
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(N[20]).append(n2).append(',').append(n3).append(',').append(n4).append(',');
            if (wArray != null) {
                string = N[0];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = N[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    final void d(int n2, int n3) {
        boolean bl = client.vh;
        try {
            int n4;
            int n5;
            int n6;
            int n7;
            int[] nArray;
            block14: {
                ++d;
                if (null == this.kb[n3]) {
                    return;
                }
                nArray = this.kb[n3];
                n7 = 0;
                while (n7 < 64) {
                    block15: {
                        n6 = 4032 + n7;
                        n5 = nArray[n6];
                        n4 = 0;
                        if (bl) break block14;
                        int n8 = n4;
                        while (~n8 > -64) {
                            nArray[n6] = nArray[-64 + n6];
                            n6 -= 64;
                            ++n8;
                            if (!bl) {
                                if (!bl) continue;
                                break;
                            }
                            break block15;
                        }
                        this.kb[n3][n6] = n5;
                        ++n7;
                    }
                    if (!bl) continue;
                }
                n4 = n2;
            }
            if (n4 != 25013) {
                this.b = 60;
            }
            n7 = 4096;
            n6 = 0;
            while (~n7 < ~n6) {
                n5 = nArray[n6];
                try {
                    nArray[n7 - -n6] = ib.a(n5 - (n5 >>> 113007331), 0xF8F8FF);
                    nArray[n6 + n7 * 2] = ib.a(0xF8F8FF, n5 - (n5 >>> -391329182));
                    nArray[n6 + n7 * 3] = ib.a(0xF8F8FF, -(n5 >>> -702449277) + n5 - (n5 >>> 520093090));
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
            throw i.a(runtimeException, N[13] + n2 + ',' + n3 + ')');
        }
    }

    final int b(int n2) {
        try {
            ++rb;
            if (n2 != 0) {
                this.S = null;
            }
            return this.cc;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[9] + n2 + ')');
        }
    }

    final void a(byte by, int n2) {
        try {
            if (by != 67) {
                this.cb = 31;
            }
            this.n -= n2;
            ++E;
            this.T.b(2 * n2, -113, n2);
            if (-1 < ~this.n) {
                this.n = 0;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[6] + by + ',' + n2 + ')');
        }
    }

    final void b(int n2, int n3, int n4) {
        try {
            if (n2 <= 15) {
                this.kb = null;
            }
            this.Q[n3] = n4;
            ++Bb;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[17] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    private final boolean a(int n2, w[] wArray, int n3, byte by) {
        boolean bl = client.vh;
        try {
            int n4 = 109 % ((-9 - by) / 37);
            ++bb;
            while (true) {
                block19: {
                    block18: {
                        int n5;
                        block15: {
                            int n6;
                            w w2 = wArray[n2];
                            int n7 = n6 = n2 - -1;
                            while (n3 >= n6) {
                                w w3 = wArray[n6];
                                n7 = this.a((byte)-114, w2, w3) ? 1 : 0;
                                if (bl) continue;
                                if (n7 == 0) break;
                                wArray[n2] = w3;
                                n2 = n6;
                                wArray[n6] = w2;
                                if (~n2 == ~n3) {
                                    this.eb = n2 - 1;
                                    this.e = n2;
                                    return true;
                                }
                                ++n6;
                                if (!bl) continue;
                            }
                            w w4 = wArray[n3];
                            int n8 = -1 + n3;
                            while (~n2 >= ~n8) {
                                block17: {
                                    block16: {
                                        w w5 = wArray[n8];
                                        n5 = this.a((byte)-46, w5, w4) ? 1 : 0;
                                        if (bl) break block15;
                                        if (n5 == 0) break;
                                        wArray[n3] = w5;
                                        wArray[n8] = w4;
                                        n3 = n8--;
                                        if (~n2 == ~n3) break block16;
                                        break block17;
                                    }
                                    this.eb = n3;
                                    this.e = n3 + 1;
                                    return true;
                                }
                                if (!bl) continue;
                            }
                            n5 = ~n3;
                        }
                        if (n5 >= ~(n2 - -1)) break block18;
                        break block19;
                    }
                    this.eb = n3;
                    this.e = n2;
                    return false;
                }
                if (!this.a(n2 + 1, wArray, n3, (byte)70)) {
                    this.e = n2;
                    return false;
                }
                n3 = this.eb;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(N[40]).append(n2).append(',');
            String string = wArray != null ? N[0] : N[1];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n3).append(',').append(by).append(')').toString());
        }
    }

    final void a(int n2, int n3, int n4, int n5) {
        try {
            this.L = new int[n5][];
            this.g = new byte[n5][];
            ++sb;
            this.kb = new int[n5][];
            this.i = new int[n4][];
            this.S = new boolean[n5];
            k.e = n2;
            this.cb = n5;
            this.Hb = new int[n5];
            this.ec = new int[n3][];
            this.D = new long[n5];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[12] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    private final void a(int n2, int n3, int n4, boolean bl) {
        try {
            block33: {
                block32: {
                    block31: {
                        block30: {
                            int n5;
                            int n6;
                            int n7;
                            int n8;
                            block29: {
                                int n9;
                                block28: {
                                    block27: {
                                        int n10;
                                        block26: {
                                            ++M;
                                            n9 = 1024 + -this.Kb & 0x3FF;
                                            n8 = 0x3FF & 1024 + -this.xb;
                                            n10 = 1024 + -this.b & 0x3FF;
                                            if (!bl) {
                                                this.Zb = 25;
                                            }
                                            if (~n10 != -1) break block26;
                                            break block27;
                                        }
                                        n7 = ba.cc[n10];
                                        n6 = ba.cc[1024 + n10];
                                        n5 = n6 * n3 + n7 * n4 >> -2133812657;
                                        n4 = -(n7 * n3) + n4 * n6 >> 69817711;
                                        n3 = n5;
                                    }
                                    if (0 != n9) break block28;
                                    break block29;
                                }
                                n6 = ba.cc[1024 + n9];
                                n7 = ba.cc[n9];
                                n5 = -(n7 * n2) + n4 * n6 >> 618592719;
                                n2 = n6 * n2 + n7 * n4 >> 425007119;
                                n4 = n5;
                            }
                            if (n8 != 0) {
                                n7 = ba.cc[n8];
                                n6 = ba.cc[1024 + n8];
                                n5 = n7 * n2 + n3 * n6 >> 383771055;
                                n2 = n6 * n2 - n7 * n3 >> 1372459087;
                                n3 = n5;
                            }
                            if (da.K < n2) {
                                da.K = n2;
                            }
                            if (~n4 > ~aa.b) break block30;
                            break block31;
                        }
                        aa.b = n4;
                    }
                    if (~oa.b > ~n3) break block32;
                    break block33;
                }
                oa.b = n3;
            }
            if (~nb.y > ~n4) {
                nb.y = n4;
            }
            if (n2 < m.j) {
                m.j = n2;
            }
            if (n3 < aa.f) {
                aa.f = n3;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[16] + n2 + ',' + n3 + ',' + n4 + ',' + bl + ')');
        }
    }

    final void c(int n2, int n3) {
        try {
            ++Gb;
            this.T.zb[n3] = 1;
            if (n2 != 32768) {
                this.Cb = 32;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[28] + n2 + ',' + n3 + ')');
        }
    }

    /*
     * Unable to fully structure code
     */
    final void c(int var1_1) {
        block150: {
            var22_2 = client.vh;
            try {
                block149: {
                    block170: {
                        block138: {
                            block137: {
                                ++lb.Y;
                                this.f = this.dc.i;
                                var7_3 = this.A * this.Mb >> this.R;
                                aa.b = 0;
                                nb.y = 0;
                                m.j = 0;
                                da.K = 0;
                                var8_4 = this.Mb * this.wb >> this.R;
                                oa.b = 0;
                                aa.f = 0;
                                this.a(this.Mb, -var7_3, -var8_4, true);
                                this.a(this.Mb, -var7_3, var8_4, true);
                                this.a(this.Mb, var7_3, -var8_4, true);
                                this.a(this.Mb, var7_3, var8_4, true);
                                this.a(0, -this.A, -this.wb, true);
                                this.a(0, -this.A, this.wb, true);
                                this.a(0, this.A, -this.wb, true);
                                this.a(0, this.A, this.wb, true);
                                nb.y += this.o;
                                da.K += this.I;
                                aa.b += this.o;
                                aa.f += this.bc;
                                m.j += this.I;
                                oa.b += this.bc;
                                this.Z[this.ab] = this.T;
                                this.T.Yb = 2;
                                var3_5 = 0;
                                while (~var3_5 > ~this.ab) {
                                    this.Z[var3_5].a(this.o, this.R, this.bc, (byte)-122, this.I, this.xb, this.b, this.Kb, this.nb);
                                    ++var3_5;
                                    if (!var22_2) {
                                        if (!var22_2) continue;
                                        break;
                                    }
                                    break block137;
                                }
                                this.Z[this.ab].a(this.o, this.R, this.bc, (byte)-114, this.I, this.xb, this.b, this.Kb, this.nb);
                                this.zb = 0;
                            }
                            var9_6 = 0;
                            block83: while (true) {
                                v1 = ~this.ab;
                                block84: while (v1 < ~var9_6) {
                                    block139: {
                                        var2_9 = this.Z[var9_6];
                                        v2 = (int)var2_9.dc;
                                        if (var22_2) break block138;
                                        if (v2 == 0) break block139;
                                        var3_5 = 0;
                                        while (~var2_9.t < ~var3_5) {
                                            block147: {
                                                block169: {
                                                    block168: {
                                                        block144: {
                                                            block141: {
                                                                block140: {
                                                                    var10_14 = var2_9.lb[var3_5];
                                                                    var11_16 = var2_9.o[var3_5];
                                                                    var5_12 = 0;
                                                                    v1 = 0;
                                                                    if (var22_2) continue block84;
                                                                    var12_20 = v1;
                                                                    while (~var12_20 > ~var10_14) {
                                                                        var4_11 = var2_9.bb[var11_16[var12_20]];
                                                                        v4 = this.nb;
                                                                        v5 = var4_11;
                                                                        if (var22_2) break block140;
                                                                        if (v4 < v5) {
                                                                        }
                                                                        ** GOTO lbl83
                                                                        if (~var4_11 > ~this.Mb) {
                                                                        }
                                                                        ** GOTO lbl83
                                                                        var5_12 = 1;
                                                                        if (!var22_2) break;
lbl83:
                                                                        // 3 sources

                                                                        ++var12_20;
                                                                        if (!var22_2) continue;
                                                                        break;
                                                                    }
                                                                    v4 = ~var5_12;
                                                                    v5 = -1;
                                                                }
                                                                if (v4 == v5) break block147;
                                                                var5_12 = 0;
                                                                for (var12_20 = 0; var12_20 < var10_14; ++var12_20) {
                                                                    block143: {
                                                                        block142: {
                                                                            var4_11 = var2_9.pb[var11_16[var12_20]];
                                                                            v9 = ~var4_11;
                                                                            v10 = ~(-this.A);
                                                                            if (var22_2) break block141;
                                                                            if (v9 < v10) break block142;
                                                                            break block143;
                                                                        }
                                                                        var5_12 |= 1;
                                                                    }
                                                                    if (~var4_11 > ~this.A) {
                                                                        var5_12 |= 2;
                                                                    }
                                                                    if (-4 != ~var5_12) continue;
                                                                    if (!var22_2) break;
                                                                    continue;
                                                                    catch (RuntimeException v12) {
                                                                        throw v12;
                                                                    }
                                                                }
                                                                v9 = -4;
                                                                v10 = ~var5_12;
                                                            }
                                                            if (v9 != v10) break block147;
                                                            var5_12 = 0;
                                                            var12_20 = 0;
                                                            while (~var12_20 > ~var10_14) {
                                                                block146: {
                                                                    block145: {
                                                                        var4_11 = var2_9.Ob[var11_16[var12_20]];
                                                                        v14 = -this.wb;
                                                                        v15 = var4_11;
                                                                        if (var22_2) break block144;
                                                                        if (v14 < v15) break block145;
                                                                        break block146;
                                                                    }
                                                                    var5_12 |= 1;
                                                                }
                                                                if (~var4_11 > ~this.wb) {
                                                                    var5_12 |= 2;
                                                                }
                                                                if (3 == var5_12) break;
                                                                ++var12_20;
                                                                if (!var22_2) continue;
                                                                break;
                                                            }
                                                            v14 = 3;
                                                            v15 = var5_12;
                                                        }
                                                        if (v14 != v15 && !var22_2) break block147;
                                                        var12_21 = this.y[this.zb];
                                                        var12_21.o = var2_9;
                                                        var12_21.i = var3_5;
                                                        this.a(this.zb, -21875);
                                                        if (-1 >= ~var12_21.s) break block168;
                                                        var13_24 = var2_9.V[var3_5];
                                                        if (!var22_2) break block169;
                                                    }
                                                    var13_24 = var2_9.qb[var3_5];
                                                }
                                                if (~var13_24 != -12345679) {
                                                    block148: {
                                                        var6_13 = 0;
                                                        for (var14_25 = 0; var10_14 > var14_25; ++var14_25) {
                                                            var6_13 += var2_9.bb[var11_16[var14_25]];
                                                            if (!var22_2) {
                                                                if (!var22_2) continue;
                                                                break;
                                                            }
                                                            break block148;
                                                        }
                                                        var12_21.t = var2_9.hc + var6_13 / var10_14;
                                                        ++this.zb;
                                                    }
                                                    var12_21.b = var13_24;
                                                }
                                            }
                                            ++var3_5;
                                            if (!var22_2) continue;
                                        }
                                    }
                                    ++var9_6;
                                    if (!var22_2) continue block83;
                                }
                                break;
                            }
                            v2 = var1_1;
                        }
                        if (v2 > -99) {
                            this.H = null;
                        }
                        var2_9 = this.T;
                        if (!var2_9.dc) break block170;
                        for (var3_5 = 0; var2_9.t > var3_5; ++var3_5) {
                            block173: {
                                block172: {
                                    block171: {
                                        var9_7 = var2_9.o[var3_5];
                                        var10_14 = var9_7[0];
                                        var11_17 = var2_9.pb[var10_14];
                                        var12_22 = var2_9.Ob[var10_14];
                                        var13_24 = var2_9.bb[var10_14];
                                        v21 = this.nb;
                                        v22 = var13_24;
                                        if (var22_2) break block149;
                                        if (v21 >= v22) continue;
                                        if (~this.X >= ~var13_24) continue;
                                        var14_25 = (this.ob[var3_5] << this.R) / var13_24;
                                        var15_26 = (this.Eb[var3_5] << this.R) / var13_24;
                                        if (this.A < -(var14_25 / 2) + var11_17) continue;
                                        if (~(var11_17 + var14_25 / 2) <= ~(-this.A)) ** GOTO lbl228
                                        break block171;
                                        catch (RuntimeException v25) {
                                            throw v25;
                                        }
                                    }
                                    if (!var22_2) continue;
                                    break block172;
                                    catch (RuntimeException v26) {
                                        throw v26;
                                    }
                                }
                                if (var12_22 + -var15_26 > this.wb) continue;
                                break block173;
                                catch (RuntimeException v27) {
                                    throw v27;
                                }
                            }
                            if (~(-this.wb) < ~var12_22) continue;
                            var16_27 = this.y[this.zb];
                            var16_27.i = var3_5;
                            var16_27.o = var2_9;
                            this.b(-103, this.zb);
                            var16_27.t = (var2_9.bb[var9_7[1]] + var13_24) / 2;
                            ++this.zb;
                            if (!var22_2) continue;
                        }
                    }
                    v21 = 0;
                    v22 = this.zb;
                }
                if (v21 == v22) {
                    return;
                }
                this.a(0, -1, this.y, -1 + this.zb);
                this.a(this.zb, 100, -53, this.y);
                var9_8 = 0;
                while (~var9_8 > ~this.zb) {
                    block165: {
                        block167: {
                            block166: {
                                block184: {
                                    block183: {
                                        block151: {
                                            block161: {
                                                block153: {
                                                    block174: {
                                                        block152: {
                                                            var10_15 = this.y[var9_8];
                                                            var3_5 = var10_15.i;
                                                            var2_9 = var10_15.o;
                                                            if (var22_2) break block150;
                                                            if (var2_9 == this.T) break block151;
                                                            var14_25 = 0;
                                                            var16_28 = 0;
                                                            var17_29 = var2_9.lb[var3_5];
                                                            if (12345678 != var2_9.Hb[var3_5]) break block152;
                                                            break block153;
                                                        }
                                                        if (0 <= var10_15.s) break block174;
                                                        var16_28 = var2_9.Jb + -var2_9.Hb[var3_5];
                                                        if (!var22_2) break block153;
                                                    }
                                                    var16_28 = var2_9.Jb - -var2_9.Hb[var3_5];
                                                }
                                                var18_30 = var2_9.o[var3_5];
                                                var19_32 = 0;
                                                while (~var19_32 > ~var17_29) {
                                                    block157: {
                                                        block179: {
                                                            block178: {
                                                                block160: {
                                                                    block159: {
                                                                        block177: {
                                                                            block158: {
                                                                                block156: {
                                                                                    block176: {
                                                                                        block155: {
                                                                                            block175: {
                                                                                                block154: {
                                                                                                    var6_13 = var18_30[var19_32];
                                                                                                    this.Qb[var19_32] = var2_9.cc[var6_13];
                                                                                                    this.Vb[var19_32] = var2_9.H[var6_13];
                                                                                                    this.J[var19_32] = var2_9.bb[var6_13];
                                                                                                    v31 = ~var2_9.Hb[var3_5];
                                                                                                    v32 = -12345679;
                                                                                                    if (!var22_2) {
                                                                                                        if (v31 == v32) break block154;
                                                                                                        break block155;
                                                                                                    }
                                                                                                    ** GOTO lbl368
                                                                                                }
                                                                                                if (~var10_15.s <= -1) break block175;
                                                                                                var16_28 = -var2_9.gb[var6_13] + (var2_9.Jb - -var2_9.Ab[var6_13]);
                                                                                                if (!var22_2) break block155;
                                                                                            }
                                                                                            var16_28 = var2_9.Ab[var6_13] + (var2_9.Jb - -var2_9.gb[var6_13]);
                                                                                        }
                                                                                        if (var2_9.bb[var6_13] < this.nb) ** GOTO lbl323
                                                                                        this.yb[var14_25] = var2_9.pb[var6_13];
                                                                                        this.B[var14_25] = var2_9.Ob[var6_13];
                                                                                        this.r[var14_25] = var16_28;
                                                                                        if (var2_9.bb[var6_13] <= this.G) break block156;
                                                                                        break block176;
                                                                                        catch (RuntimeException v34) {
                                                                                            throw v34;
                                                                                        }
                                                                                    }
                                                                                    v35 = var14_25;
                                                                                    this.r[v35] = this.r[v35] + (-this.G + var2_9.bb[var6_13]) / this.P;
                                                                                }
                                                                                ++var14_25;
                                                                                if (!var22_2) break block157;
lbl323:
                                                                                // 2 sources

                                                                                if (-1 == ~var19_32) break block158;
                                                                                var15_26 = var18_30[var19_32 + -1];
                                                                                if (!var22_2) break block177;
                                                                            }
                                                                            var15_26 = var18_30[var17_29 - 1];
                                                                        }
                                                                        if (~this.nb >= ~var2_9.bb[var15_26]) break block159;
                                                                        break block160;
                                                                    }
                                                                    var13_24 = var2_9.bb[var6_13] - var2_9.bb[var15_26];
                                                                    var12_23 = var2_9.H[var6_13] - (var2_9.bb[var6_13] + -this.nb) * (-var2_9.H[var15_26] + var2_9.H[var6_13]) / var13_24;
                                                                    var11_18 = -((-var2_9.cc[var15_26] + var2_9.cc[var6_13]) * (var2_9.bb[var6_13] + -this.nb) / var13_24) + var2_9.cc[var6_13];
                                                                    this.yb[var14_25] = (var11_18 << this.R) / this.nb;
                                                                    this.B[var14_25] = (var12_23 << this.R) / this.nb;
                                                                    this.r[var14_25] = var16_28;
                                                                    ++var14_25;
                                                                }
                                                                if (~var19_32 != ~(-1 + var17_29)) break block178;
                                                                var15_26 = var18_30[0];
                                                                if (!var22_2) break block179;
                                                            }
                                                            var15_26 = var18_30[var19_32 - -1];
                                                        }
                                                        if (~this.nb >= ~var2_9.bb[var15_26]) {
                                                            var13_24 = -var2_9.bb[var15_26] + var2_9.bb[var6_13];
                                                            var12_23 = var2_9.H[var6_13] + -((-this.nb + var2_9.bb[var6_13]) * (var2_9.H[var6_13] - var2_9.H[var15_26]) / var13_24);
                                                            var11_18 = -((var2_9.cc[var6_13] + -var2_9.cc[var15_26]) * (-this.nb + var2_9.bb[var6_13]) / var13_24) + var2_9.cc[var6_13];
                                                            this.yb[var14_25] = (var11_18 << this.R) / this.nb;
                                                            this.B[var14_25] = (var12_23 << this.R) / this.nb;
                                                            this.r[var14_25] = var16_28;
                                                            ++var14_25;
                                                        }
                                                    }
                                                    ++var19_32;
                                                    if (!var22_2) continue;
                                                }
                                                var19_32 = 0;
                                                do {
                                                    block164: {
                                                        block182: {
                                                            block163: {
                                                                block162: {
                                                                    block181: {
                                                                        block180: {
                                                                            v31 = var19_32;
                                                                            v32 = var17_29;
lbl368:
                                                                            // 3 sources

                                                                            if (v31 >= v32) break;
                                                                            v39 = 0;
                                                                            v40 = this.r[var19_32];
                                                                            if (var22_2) break block161;
                                                                            break block180;
                                                                            catch (RuntimeException v41) {
                                                                                throw v41;
                                                                            }
                                                                        }
                                                                        if (v39 <= v40) {
                                                                        }
                                                                        ** GOTO lbl390
                                                                        if (this.r[var19_32] <= 255) break block162;
                                                                        this.r[var19_32] = 255;
                                                                        if (!var22_2) break block162;
                                                                        break block181;
                                                                        catch (RuntimeException v43) {
                                                                            throw v43;
                                                                        }
                                                                    }
                                                                    this.r[var19_32] = 0;
                                                                }
                                                                if (-1 >= ~var10_15.b) break block163;
                                                                break block164;
                                                            }
                                                            if (this.Hb[var10_15.b] == 1) ** GOTO lbl411
                                                            v46 = var19_32;
                                                            this.r[v46] = this.r[v46] << 6;
                                                            if (!var22_2) break block164;
                                                            break block182;
                                                            catch (RuntimeException v47) {
                                                                throw v47;
                                                            }
                                                        }
                                                        v48 = var19_32;
                                                        this.r[v48] = this.r[v48] << 9;
                                                    }
                                                    ++var19_32;
                                                } while (!var22_2);
                                                this.a(0, var3_5, this.B, 0, 0, var2_9, this.yb, this.r, 0, 5960, var14_25);
                                                v39 = ~this.Cb;
                                                v40 = ~this.Xb;
                                            }
                                            if (v39 < v40) {
                                                this.a(this.Vb, var2_9, 1, var17_29, var10_15.b, this.J, this.Qb, 0, 0);
                                            }
                                            if (!var22_2) break block165;
                                        }
                                        var11_19 = var2_9.o[var3_5];
                                        var12_23 = var11_19[0];
                                        var13_24 = var2_9.pb[var12_23];
                                        var14_25 = var2_9.Ob[var12_23];
                                        var15_26 = var2_9.bb[var12_23];
                                        var16_28 = (this.ob[var3_5] << this.R) / var15_26;
                                        var17_29 = (this.Eb[var3_5] << this.R) / var15_26;
                                        var18_31 = var14_25 - var2_9.Ob[var11_19[1]];
                                        var19_32 = var18_31 * (var2_9.pb[var11_19[1]] - var13_24) / var17_29;
                                        var19_32 = var2_9.pb[var11_19[1]] - var13_24;
                                        var20_33 = var13_24 - var16_28 / 2;
                                        var21_34 = this.Nb - (-var14_25 - -var17_29);
                                        this.dc.a((256 << this.R) / var15_26, this.gb[var3_5], var17_29, var20_33 - -this.Zb, var21_34, var16_28, (byte)29, var19_32);
                                        if (!this.K || ~this.cc <= ~this.db) break block165;
                                        var20_33 += (this.Q[var3_5] << this.R) / var15_26;
                                        if (var21_34 > this.Wb) break block165;
                                        if (~this.Wb < ~(var21_34 + var17_29)) break block165;
                                        break block183;
                                        catch (RuntimeException v51) {
                                            throw v51;
                                        }
                                    }
                                    if (~this.j > ~var20_33) break block165;
                                    break block184;
                                    catch (RuntimeException v52) {
                                        throw v52;
                                    }
                                }
                                if (~(var20_33 - -var16_28) <= ~this.j) break block166;
                                break block165;
                                catch (RuntimeException v53) {
                                    throw v53;
                                }
                            }
                            if (var2_9.db) break block165;
                            if (0 == var2_9.zb[var3_5]) break block167;
                            break block165;
                            catch (RuntimeException v55) {
                                throw v55;
                            }
                        }
                        this.Ab[this.cc] = var2_9;
                        this.qb[this.cc] = var3_5;
                        ++this.cc;
                    }
                    ++var9_8;
                    if (!var22_2) continue;
                }
                this.K = false;
            }
            catch (RuntimeException var2_10) {
                throw i.a(var2_10, lb.N[24] + var1_1 + ')');
            }
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void b(int n2, boolean bl) {
        boolean bl2 = client.vh;
        try {
            block26: {
                int n3;
                int n4;
                ++W;
                if (!bl) {
                    this.K = false;
                }
                if (n2 < 0) {
                    return;
                }
                ++k.e;
                if (this.kb[n2] != null) return;
                if (this.Hb[n2] != 0) {
                    block25: {
                        int n5 = 0;
                        while (~n5 > ~this.ec.length) {
                            if (this.ec[n5] == null) {
                                this.ec[n5] = new int[65536];
                                this.kb[n2] = this.ec[n5];
                                this.a(n2, (byte)118);
                                return;
                            }
                            ++n5;
                            if (!bl2) continue;
                        }
                        long l2 = 0x40000000L;
                        n4 = 0;
                        n3 = 0;
                        while (~n3 > ~this.cb) {
                            if (!bl2) {
                                if (n2 != n3 && ~this.Hb[n3] == -2 && null != this.kb[n3] && this.D[n3] < l2) {
                                    l2 = this.D[n3];
                                    n4 = n3;
                                }
                                ++n3;
                                if (!bl2) continue;
                            }
                            break block25;
                        }
                        this.kb[n2] = this.kb[n4];
                        this.kb[n4] = null;
                        this.a(n2, (byte)118);
                    }
                    if (!bl2) return;
                }
                int n6 = 0;
                while (~n6 > ~this.i.length) {
                    if (null == this.i[n6]) {
                        this.i[n6] = new int[16384];
                        this.kb[n2] = this.i[n6];
                        this.a(n2, (byte)118);
                        return;
                    }
                    ++n6;
                    if (!bl2) continue;
                }
                long l3 = 0x40000000L;
                n4 = 0;
                n3 = 0;
                while (~this.cb < ~n3) {
                    if (!bl2) {
                        if (n2 != n3 && ~this.Hb[n3] == -1 && null != this.kb[n3] && (l3 ^ 0xFFFFFFFFFFFFFFFFL) < (this.D[n3] ^ 0xFFFFFFFFFFFFFFFFL)) {
                            l3 = this.D[n3];
                            n4 = n3;
                        }
                        ++n3;
                        if (!bl2) continue;
                    }
                    break block26;
                }
                this.kb[n2] = this.kb[n4];
                this.kb[n4] = null;
            }
            this.a(n2, (byte)118);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[21] + n2 + ',' + bl + ')');
        }
    }

    /*
     * Unable to fully structure code
     */
    private final void a(int[] var1_1, ca var2_2, int var3_3, int var4_4, int var5_5, int[] var6_6, int[] var7_7, int var8_8, int var9_9) {
        var40_10 = client.vh;
        try {
            block109: {
                block116: {
                    block161: {
                        block137: {
                            block136: {
                                block133: {
                                    block155: {
                                        block160: {
                                            block156: {
                                                block108: {
                                                    block107: {
                                                        block157: {
                                                            block106: {
                                                                block105: {
                                                                    ++lb.F;
                                                                    if (-2 == var5_5) {
                                                                        return;
                                                                    }
                                                                    if (0 > var5_5) break block155;
                                                                    if (var5_5 >= this.cb) {
                                                                        var5_5 = 0;
                                                                    }
                                                                    this.b(var5_5, true);
                                                                    var10_11 = var7_7[0];
                                                                    var11_13 = var1_1[0];
                                                                    var12_14 = var6_6[0];
                                                                    var13_15 = var10_11 - var7_7[1];
                                                                    var14_19 = -var1_1[1] + var11_13;
                                                                    var15_20 = var12_14 + -var6_6[1];
                                                                    var16_21 = var7_7[--var4_4] + -var10_11;
                                                                    var17_22 = -var11_13 + var1_1[var4_4];
                                                                    var18_23 = var6_6[var4_4] - var12_14;
                                                                    if (~this.Hb[var5_5] != -2) break block156;
                                                                    var19_24 = -(var17_22 * var10_11) + var11_13 * var16_21 << -1007118196;
                                                                    var20_25 = var17_22 * var12_14 + -(var18_23 * var11_13) << 4 + -this.R + 5 - -7;
                                                                    var21_26 = var18_23 * var10_11 + -(var16_21 * var12_14) << 7 + -this.R + 5;
                                                                    var22_27 = var11_13 * var13_15 - var10_11 * var14_19 << 291418604;
                                                                    var23_28 = var14_19 * var12_14 - var15_20 * var11_13 << -this.R + 5 - -11;
                                                                    var24_29 = -(var12_14 * var13_15) + var10_11 * var15_20 << 7 + (5 - this.R);
                                                                    var25_30 = var16_21 * var14_19 + -(var17_22 * var13_15) << 1405568421;
                                                                    var26_31 = var15_20 * var17_22 - var14_19 * var18_23 << 4 + (-this.R + 5);
                                                                    var27_32 = var13_15 * var18_23 + -(var15_20 * var16_21) >> -5 + this.R;
                                                                    var28_33 = var20_25 >> 1240089476;
                                                                    var29_34 = var23_28 >> -557406012;
                                                                    var30_35 = var26_31 >> -384718332;
                                                                    var31_36 = -this.Nb + this.Xb;
                                                                    var32_37 = this.vb;
                                                                    var33_38 = var32_37 * this.Xb + this.Zb;
                                                                    var22_27 += var31_36 * var24_29;
                                                                    var34_39 = 1;
                                                                    var25_30 += var27_32 * var31_36;
                                                                    var19_24 += var21_26 * var31_36;
                                                                    if (this.f) break block105;
                                                                    break block106;
                                                                }
                                                                if (-2 == ~(this.Xb & 1)) {
                                                                    var22_27 += var24_29;
                                                                    var33_38 += var32_37;
                                                                    var19_24 += var21_26;
                                                                    var25_30 += var27_32;
                                                                    ++this.Xb;
                                                                }
                                                                var27_32 <<= 1;
                                                                var24_29 <<= 1;
                                                                var34_39 = 2;
                                                                var21_26 <<= 1;
                                                                var32_37 <<= 1;
                                                            }
                                                            if (var2_2.Kb) ** GOTO lbl156
                                                            if (this.S[var5_5]) break block107;
                                                            break block157;
                                                            catch (RuntimeException v1) {
                                                                throw v1;
                                                            }
                                                        }
                                                        v2 = true;
                                                        break block108;
                                                    }
                                                    v2 = false;
                                                }
                                                if (v2) ** GOTO lbl116
                                                var9_9 = this.Xb;
                                                while (~this.Cb < ~var9_9) {
                                                    block158: {
                                                        block112: {
                                                            block111: {
                                                                block110: {
                                                                    var35_40 = this.x[var9_9];
                                                                    var8_8 = var35_40.d >> 401408072;
                                                                    var36_41 = var35_40.k >> 627719496;
                                                                    var37_42 = -var8_8 + var36_41;
                                                                    v4 = -1;
                                                                    v5 = ~var37_42;
                                                                    if (var40_10) break block109;
                                                                    if (v4 > v5) break block110;
                                                                    var22_27 += var24_29;
                                                                    var19_24 += var21_26;
                                                                    var33_38 += var32_37;
                                                                    var25_30 += var27_32;
                                                                    if (!var40_10) break block158;
                                                                }
                                                                var38_43 = var35_40.e;
                                                                var39_44 = (-var38_43 + var35_40.l) / var37_42;
                                                                if (-this.A > var8_8) break block111;
                                                                break block112;
                                                            }
                                                            var38_43 += (-var8_8 + -this.A) * var39_44;
                                                            var8_8 = -this.A;
                                                            var37_42 = var36_41 - var8_8;
                                                        }
                                                        if (~this.A > ~var36_41) {
                                                            var36_41 = this.A;
                                                            var37_42 = -var8_8 + var36_41;
                                                        }
                                                        wb.a(var23_28, 10, 0, 0, this.pb, var25_30 + var8_8 * var30_35, var38_43, var8_8 * var28_33 + var19_24, var22_27 - -(var8_8 * var29_34), var8_8 + var33_38, var26_31, var39_44, 0, var20_25, this.kb[var5_5], var37_42);
                                                        var33_38 += var32_37;
                                                        var22_27 += var24_29;
                                                        var25_30 += var27_32;
                                                        var19_24 += var21_26;
                                                    }
                                                    var9_9 += var34_39;
                                                    if (!var40_10) continue;
lbl116:
                                                    // 4 sources

                                                    for (var9_9 = this.Xb; var9_9 < this.Cb; var9_9 += var34_39) {
                                                        block115: {
                                                            block114: {
                                                                block113: {
                                                                    var35_40 = this.x[var9_9];
                                                                    var8_8 = var35_40.d >> -1701677208;
                                                                    var36_41 = var35_40.k >> -278760920;
                                                                    var37_42 = var36_41 + -var8_8;
                                                                    v4 = 0;
                                                                    v5 = var37_42;
                                                                    if (var40_10) break block109;
                                                                    if (v4 < v5) break block113;
                                                                    var25_30 += var27_32;
                                                                    var19_24 += var21_26;
                                                                    var33_38 += var32_37;
                                                                    var22_27 += var24_29;
                                                                    if (!var40_10) continue;
                                                                }
                                                                var38_43 = var35_40.e;
                                                                var39_44 = (var35_40.l + -var38_43) / var37_42;
                                                                if (var8_8 < -this.A) {
                                                                    var38_43 += (-this.A - var8_8) * var39_44;
                                                                    var8_8 = -this.A;
                                                                    var37_42 = -var8_8 + var36_41;
                                                                }
                                                                if (this.A < var36_41) break block114;
                                                                break block115;
                                                            }
                                                            var36_41 = this.A;
                                                            var37_42 = var36_41 + -var8_8;
                                                        }
                                                        gb.a(var22_27 - -(var29_34 * var8_8), var20_25, (byte)50, var25_30 + var8_8 * var30_35, var38_43, var39_44 << 2091149602, this.kb[var5_5], var8_8 + var33_38, var8_8 * var28_33 + var19_24, var26_31, 0, 0, this.pb, var23_28, var37_42);
                                                        var19_24 += var21_26;
                                                        var25_30 += var27_32;
                                                        var33_38 += var32_37;
                                                        var22_27 += var24_29;
                                                        if (!var40_10) continue;
lbl156:
                                                        // 2 sources

                                                        var9_9 = this.Xb;
                                                        while (~this.Cb < ~var9_9) {
                                                            block159: {
                                                                block120: {
                                                                    block119: {
                                                                        block118: {
                                                                            block117: {
                                                                                var35_40 = this.x[var9_9];
                                                                                var8_8 = var35_40.d >> 1738077768;
                                                                                var36_41 = var35_40.k >> 2088632040;
                                                                                var37_42 = -var8_8 + var36_41;
                                                                                v4 = var37_42;
                                                                                if (var40_10) break block116;
                                                                                if (v4 <= 0) break block117;
                                                                                break block118;
                                                                            }
                                                                            var33_38 += var32_37;
                                                                            var25_30 += var27_32;
                                                                            var19_24 += var21_26;
                                                                            var22_27 += var24_29;
                                                                            if (!var40_10) break block159;
                                                                        }
                                                                        var38_43 = var35_40.e;
                                                                        var39_44 = (-var38_43 + var35_40.l) / var37_42;
                                                                        if (~var8_8 > ~(-this.A)) {
                                                                            var38_43 += var39_44 * (-var8_8 + -this.A);
                                                                            var8_8 = -this.A;
                                                                            var37_42 = var36_41 - var8_8;
                                                                        }
                                                                        if (~this.A > ~var36_41) break block119;
                                                                        break block120;
                                                                    }
                                                                    var36_41 = this.A;
                                                                    var37_42 = -var8_8 + var36_41;
                                                                }
                                                                cb.a(var33_38 + var8_8, var22_27 + var8_8 * var29_34, var19_24 + var8_8 * var28_33, 0, var38_43, var23_28, 0, var25_30 - -(var8_8 * var30_35), var20_25, var39_44 << -1940787422, this.kb[var5_5], var37_42, var26_31, this.pb, (byte)119);
                                                                var33_38 += var32_37;
                                                                var22_27 += var24_29;
                                                                var19_24 += var21_26;
                                                                var25_30 += var27_32;
                                                            }
                                                            var9_9 += var34_39;
                                                            if (!var40_10) continue;
                                                        }
                                                        break block70;
                                                    }
                                                }
                                                if (!var40_10) break block160;
                                            }
                                            var19_24 = var16_21 * var11_13 + -(var10_11 * var17_22) << -207508021;
                                            var20_25 = var12_14 * var17_22 + -(var18_23 * var11_13) << 4 + (6 + (-this.R + 5));
                                            var21_26 = var18_23 * var10_11 + -(var16_21 * var12_14) << 11 - this.R;
                                            var22_27 = var11_13 * var13_15 + -(var14_19 * var10_11) << -2039965749;
                                            var23_28 = var12_14 * var14_19 + -(var11_13 * var15_20) << 4 + -this.R + 11;
                                            var24_29 = var15_20 * var10_11 - var12_14 * var13_15 << 11 + -this.R;
                                            var25_30 = -(var17_22 * var13_15) + var16_21 * var14_19 << -1812048155;
                                            var26_31 = var17_22 * var15_20 - var14_19 * var18_23 << 4 + (-this.R + 5);
                                            var27_32 = var18_23 * var13_15 + -(var16_21 * var15_20) >> -5 + this.R;
                                            var28_33 = var20_25 >> 1432901060;
                                            var29_34 = var23_28 >> -1527794172;
                                            var30_35 = var26_31 >> 1007458052;
                                            var31_36 = this.Xb + -this.Nb;
                                            var32_37 = this.vb;
                                            var33_38 = var32_37 * this.Xb + this.Zb;
                                            var22_27 += var31_36 * var24_29;
                                            var34_39 = 1;
                                            var19_24 += var31_36 * var21_26;
                                            var25_30 += var27_32 * var31_36;
                                            v12 = this.f == false;
                                            if (!v12) {
                                                block122: {
                                                    block121: {
                                                        if (1 == (1 & this.Xb)) break block121;
                                                        break block122;
                                                    }
                                                    var22_27 += var24_29;
                                                    var25_30 += var27_32;
                                                    var19_24 += var21_26;
                                                    ++this.Xb;
                                                    var33_38 += var32_37;
                                                }
                                                var21_26 <<= 1;
                                                var32_37 <<= 1;
                                                var27_32 <<= 1;
                                                var34_39 = 2;
                                                var24_29 <<= 1;
                                            }
                                            if (!var2_2.Kb) ** GOTO lbl289
                                            for (var9_9 = this.Xb; this.Cb > var9_9; var9_9 += var34_39) {
                                                block127: {
                                                    block126: {
                                                        block125: {
                                                            block124: {
                                                                block123: {
                                                                    var35_40 = this.x[var9_9];
                                                                    var8_8 = var35_40.d >> -1299582008;
                                                                    var36_41 = var35_40.k >> -1270224664;
                                                                    var37_42 = var36_41 + -var8_8;
                                                                    v4 = 0;
                                                                    v5 = var37_42;
                                                                    if (var40_10) break block109;
                                                                    if (v4 < v5) break block123;
                                                                    var22_27 += var24_29;
                                                                    var25_30 += var27_32;
                                                                    var33_38 += var32_37;
                                                                    var19_24 += var21_26;
                                                                    if (!var40_10) continue;
                                                                }
                                                                var38_43 = var35_40.e;
                                                                var39_44 = (-var38_43 + var35_40.l) / var37_42;
                                                                if (~(-this.A) < ~var8_8) break block124;
                                                                break block125;
                                                            }
                                                            var38_43 += var39_44 * (-this.A + -var8_8);
                                                            var8_8 = -this.A;
                                                            var37_42 = var36_41 - var8_8;
                                                        }
                                                        if (~this.A > ~var36_41) break block126;
                                                        break block127;
                                                    }
                                                    var36_41 = this.A;
                                                    var37_42 = var36_41 - var8_8;
                                                }
                                                jb.a(this.pb, var23_28, var26_31, var8_8 * var30_35 + var25_30, var39_44, var38_43, var8_8 + var33_38, var37_42, var28_33 * var8_8 + var19_24, 0, this.kb[var5_5], false, var20_25, var8_8 * var29_34 + var22_27, 0);
                                                var33_38 += var32_37;
                                                var25_30 += var27_32;
                                                var19_24 += var21_26;
                                                var22_27 += var24_29;
                                                if (!var40_10) continue;
lbl289:
                                                // 2 sources

                                                if (this.S[var5_5]) ** GOTO lbl325
                                                for (var9_9 = this.Xb; this.Cb > var9_9; var9_9 += var34_39) {
                                                    block129: {
                                                        block128: {
                                                            var35_40 = this.x[var9_9];
                                                            var8_8 = var35_40.d >> 1370485064;
                                                            var36_41 = var35_40.k >> 2028733736;
                                                            var37_42 = var36_41 - var8_8;
                                                            v4 = 0;
                                                            v5 = var37_42;
                                                            if (var40_10) break block109;
                                                            if (v4 >= v5) break block128;
                                                            break block129;
                                                        }
                                                        var33_38 += var32_37;
                                                        var25_30 += var27_32;
                                                        var22_27 += var24_29;
                                                        var19_24 += var21_26;
                                                        if (!var40_10) continue;
                                                    }
                                                    var38_43 = var35_40.e;
                                                    var39_44 = (var35_40.l + -var38_43) / var37_42;
                                                    if (~var8_8 > ~(-this.A)) {
                                                        var38_43 += (-this.A + -var8_8) * var39_44;
                                                        var8_8 = -this.A;
                                                        var37_42 = -var8_8 + var36_41;
                                                    }
                                                    if (this.A < var36_41) {
                                                        var36_41 = this.A;
                                                        var37_42 = var36_41 + -var8_8;
                                                    }
                                                    p.a(var39_44, 1121159302, var23_28, var8_8 * var29_34 + var22_27, var20_25, this.kb[var5_5], var38_43, 0, var19_24 + var28_33 * var8_8, 0, this.pb, var33_38 + var8_8, var25_30 - -(var8_8 * var30_35), var26_31, var37_42);
                                                    var33_38 += var32_37;
                                                    var22_27 += var24_29;
                                                    var19_24 += var21_26;
                                                    var25_30 += var27_32;
                                                    if (!var40_10) continue;
lbl325:
                                                    // 4 sources

                                                    for (var9_9 = this.Xb; this.Cb > var9_9; var9_9 += var34_39) {
                                                        block132: {
                                                            block131: {
                                                                block130: {
                                                                    var35_40 = this.x[var9_9];
                                                                    var8_8 = var35_40.d >> 784040232;
                                                                    var36_41 = var35_40.k >> 306925512;
                                                                    var37_42 = -var8_8 + var36_41;
                                                                    v4 = 0;
                                                                    v5 = var37_42;
                                                                    if (var40_10) break block109;
                                                                    if (v4 < v5) break block130;
                                                                    var25_30 += var27_32;
                                                                    var33_38 += var32_37;
                                                                    var19_24 += var21_26;
                                                                    var22_27 += var24_29;
                                                                    if (!var40_10) continue;
                                                                }
                                                                var38_43 = var35_40.e;
                                                                var39_44 = (-var38_43 + var35_40.l) / var37_42;
                                                                if (var8_8 < -this.A) break block131;
                                                                break block132;
                                                            }
                                                            var38_43 += var39_44 * (-this.A + -var8_8);
                                                            var8_8 = -this.A;
                                                            var37_42 = -var8_8 + var36_41;
                                                        }
                                                        if (var36_41 > this.A) {
                                                            var36_41 = this.A;
                                                            var37_42 = -var8_8 + var36_41;
                                                        }
                                                        cb.a(var37_42, var30_35 * var8_8 + var25_30, 0, (byte)25, 0, var20_25, var26_31, var39_44, this.kb[var5_5], this.pb, var8_8 + var33_38, var8_8 * var28_33 + var19_24, 0, var23_28, var38_43, var29_34 * var8_8 + var22_27);
                                                        var25_30 += var27_32;
                                                        var22_27 += var24_29;
                                                        var33_38 += var32_37;
                                                        var19_24 += var21_26;
                                                        if (!var40_10) continue;
                                                    }
                                                }
                                            }
                                        }
                                        if (!var40_10) break block161;
                                    }
                                    var10_11 = 0;
                                    while (var10_11 < this.ib) {
                                        block135: {
                                            block134: {
                                                v21 = this.v[var10_11];
                                                v22 = var5_5;
                                                if (var40_10) break block133;
                                                if (v21 == v22) {
                                                }
                                                ** GOTO lbl382
                                                this.H = this.Ib[var10_11];
                                                if (!var40_10) break;
lbl382:
                                                // 2 sources

                                                if (var10_11 != -1 + this.ib) break block134;
                                                var11_13 = (int)(Math.random() * (double)this.ib);
                                                this.v[var11_13] = var5_5;
                                                var5_5 = -1 + -var5_5;
                                                var12_14 = ((32025 & var5_5) >> 1739349898) * 8;
                                                var13_15 = 8 * ((1019 & var5_5) >> -999926363);
                                                var14_19 = (31 & var5_5) * 8;
                                                for (var15_20 = 0; 256 > var15_20; ++var15_20) {
                                                    var16_21 = var15_20 * var15_20;
                                                    var17_22 = var12_14 * var16_21 / 65536;
                                                    var18_23 = var16_21 * var13_15 / 65536;
                                                    var19_24 = var14_19 * var16_21 / 65536;
                                                    this.Ib[var11_13][255 + -var15_20] = var19_24 + (var18_23 << -1447485240) + (var17_22 << 1470483152);
                                                    if (!var40_10) {
                                                        if (!var40_10) continue;
                                                        break;
                                                    }
                                                    break block135;
                                                }
                                                this.H = this.Ib[var11_13];
                                            }
                                            ++var10_11;
                                        }
                                        if (!var40_10) continue;
                                    }
                                    var10_11 = this.vb;
                                    v21 = this.Xb * var10_11;
                                    v22 = this.Zb;
                                }
                                var11_13 = v21 + v22;
                                var12_14 = 1;
                                if (this.f) break block136;
                                break block137;
                            }
                            if (-2 == ~(this.Xb & 1)) {
                                ++this.Xb;
                                var11_13 += var10_11;
                            }
                            var10_11 <<= 1;
                            var12_14 = 2;
                        }
                        if (!var2_2.cb) ** GOTO lbl467
                        var9_9 = this.Xb;
                        while (~var9_9 > ~this.Cb) {
                            block162: {
                                block141: {
                                    block140: {
                                        block139: {
                                            block138: {
                                                var13_16 = this.x[var9_9];
                                                var8_8 = var13_16.d >> -2057089432;
                                                var14_19 = var13_16.k >> -1070009624;
                                                var15_20 = var14_19 + -var8_8;
                                                v4 = ~var15_20;
                                                v5 = -1;
                                                if (var40_10) break block109;
                                                if (v4 >= v5) break block138;
                                                break block139;
                                            }
                                            var11_13 += var10_11;
                                            if (!var40_10) break block162;
                                        }
                                        var16_21 = var13_16.e;
                                        var17_22 = (var13_16.l + -var16_21) / var15_20;
                                        if (var8_8 < -this.A) {
                                            var16_21 += var17_22 * (-this.A + -var8_8);
                                            var8_8 = -this.A;
                                            var15_20 = var14_19 + -var8_8;
                                        }
                                        if (~var14_19 < ~this.A) break block140;
                                        break block141;
                                    }
                                    var14_19 = this.A;
                                    var15_20 = -var8_8 + var14_19;
                                }
                                ua.a(var16_21, this.H, -var15_20, this.pb, 0, var17_22, var8_8 + var11_13, var3_3 + -1);
                                var11_13 += var10_11;
                            }
                            var9_9 += var12_14;
                            if (!var40_10) continue;
lbl467:
                            // 2 sources

                            if (this.Ub) ** GOTO lbl508
                            for (var9_9 = this.Xb; this.Cb > var9_9; var9_9 += var12_14) {
                                block147: {
                                    block146: {
                                        block145: {
                                            block144: {
                                                block143: {
                                                    block142: {
                                                        var13_17 = this.x[var9_9];
                                                        var8_8 = var13_17.d >> 1753884616;
                                                        var14_19 = var13_17.k >> 1907773288;
                                                        var15_20 = var14_19 + -var8_8;
                                                        v4 = var15_20;
                                                        if (var40_10) break block116;
                                                        if (v4 <= 0) break block142;
                                                        break block143;
                                                    }
                                                    var11_13 += var10_11;
                                                    if (!var40_10) continue;
                                                }
                                                var16_21 = var13_17.e;
                                                var17_22 = (var13_17.l - var16_21) / var15_20;
                                                if (var8_8 < -this.A) break block144;
                                                break block145;
                                            }
                                            var16_21 += (-var8_8 + -this.A) * var17_22;
                                            var8_8 = -this.A;
                                            var15_20 = -var8_8 + var14_19;
                                        }
                                        if (~var14_19 < ~this.A) break block146;
                                        break block147;
                                    }
                                    var14_19 = this.A;
                                    var15_20 = -var8_8 + var14_19;
                                }
                                t.a(0, var17_22, -var15_20, this.pb, this.H, var16_21, var11_13 + var8_8, 418609192);
                                var11_13 += var10_11;
                                if (!var40_10) continue;
lbl508:
                                // 4 sources

                                for (var9_9 = this.Xb; var9_9 < this.Cb; var9_9 += var12_14) {
                                    block153: {
                                        block152: {
                                            block151: {
                                                block150: {
                                                    block149: {
                                                        block148: {
                                                            var13_18 = this.x[var9_9];
                                                            var8_8 = var13_18.d >> 228480648;
                                                            var14_19 = var13_18.k >> 2012464840;
                                                            var15_20 = var14_19 + -var8_8;
                                                            v4 = -1;
                                                            v5 = ~var15_20;
                                                            if (var40_10) break block109;
                                                            if (v4 <= v5) break block148;
                                                            break block149;
                                                        }
                                                        var11_13 += var10_11;
                                                        if (!var40_10) continue;
                                                    }
                                                    var16_21 = var13_18.e;
                                                    var17_22 = (var13_18.l - var16_21) / var15_20;
                                                    if (var8_8 < -this.A) break block150;
                                                    break block151;
                                                }
                                                var16_21 += (-this.A + -var8_8) * var17_22;
                                                var8_8 = -this.A;
                                                var15_20 = var14_19 + -var8_8;
                                            }
                                            if (~var14_19 < ~this.A) break block152;
                                            break block153;
                                        }
                                        var14_19 = this.A;
                                        var15_20 = var14_19 + -var8_8;
                                    }
                                    ia.a(var17_22, 0, this.H, var16_21, var8_8 + var11_13, this.pb, -var15_20, (byte)82);
                                    var11_13 += var10_11;
                                    if (!var40_10) continue;
                                }
                            }
                        }
                    }
                    v4 = var3_3;
                }
                v5 = 1;
            }
            if (v4 != v5) {
                this.a((byte)-48, (w)null, (w)null);
            }
        }
        catch (RuntimeException var10_12) {
            v36 = var10_12;
            v37 = new StringBuilder().append(lb.N[27]);
            v38 = var1_1 != null ? lb.N[0] : lb.N[1];
            v40 = v37.append(v38).append(',');
            v41 = var2_2 != null ? lb.N[0] : lb.N[1];
            v43 = v40.append(v41).append(',').append(var3_3).append(',').append(var4_4).append(',').append(var5_5).append(',');
            v44 = var6_6 != null ? lb.N[0] : lb.N[1];
            v46 = v43.append(v44).append(',');
            v47 = var7_7 != null ? lb.N[0] : lb.N[1];
            throw i.a(v36, v46.append(v47).append(',').append(var8_8).append(',').append(var9_9).append(')').toString());
        }
    }

    /*
     * Unable to fully structure code
     */
    private final void a(int var1_1, int var2_2, int[] var3_3, int var4_4, int var5_5, ca var6_6, int[] var7_7, int[] var8_8, int var9_9, int var10_10, int var11_11) {
        block333: {
            var51_12 = client.vh;
            try {
                block335: {
                    block334: {
                        block357: {
                            block356: {
                                block277: {
                                    block318: {
                                        block317: {
                                            block316: {
                                                block315: {
                                                    block314: {
                                                        block313: {
                                                            block312: {
                                                                block311: {
                                                                    block310: {
                                                                        block353: {
                                                                            block352: {
                                                                                block309: {
                                                                                    block308: {
                                                                                        block351: {
                                                                                            block350: {
                                                                                                block347: {
                                                                                                    block349: {
                                                                                                        block348: {
                                                                                                            block307: {
                                                                                                                block346: {
                                                                                                                    block345: {
                                                                                                                        block306: {
                                                                                                                            block278: {
                                                                                                                                block294: {
                                                                                                                                    block293: {
                                                                                                                                        block284: {
                                                                                                                                            block292: {
                                                                                                                                                block291: {
                                                                                                                                                    block290: {
                                                                                                                                                        block289: {
                                                                                                                                                            block282: {
                                                                                                                                                                block288: {
                                                                                                                                                                    block287: {
                                                                                                                                                                        block286: {
                                                                                                                                                                            block285: {
                                                                                                                                                                                block283: {
                                                                                                                                                                                    block279: {
                                                                                                                                                                                        block266: {
                                                                                                                                                                                            block265: {
                                                                                                                                                                                                block264: {
                                                                                                                                                                                                    block260: {
                                                                                                                                                                                                        block263: {
                                                                                                                                                                                                            block262: {
                                                                                                                                                                                                                block261: {
                                                                                                                                                                                                                    block341: {
                                                                                                                                                                                                                        block340: {
                                                                                                                                                                                                                            block259: {
                                                                                                                                                                                                                                block258: {
                                                                                                                                                                                                                                    block257: {
                                                                                                                                                                                                                                        block339: {
                                                                                                                                                                                                                                            block338: {
                                                                                                                                                                                                                                                block255: {
                                                                                                                                                                                                                                                    block256: {
                                                                                                                                                                                                                                                        block337: {
                                                                                                                                                                                                                                                            block336: {
                                                                                                                                                                                                                                                                block254: {
                                                                                                                                                                                                                                                                    ++lb.Yb;
                                                                                                                                                                                                                                                                    if (var11_11 != 3) ** GOTO lbl252
                                                                                                                                                                                                                                                                    var12_13 = this.Nb + var3_3[0];
                                                                                                                                                                                                                                                                    var13_17 = var3_3[1] - -this.Nb;
                                                                                                                                                                                                                                                                    var14_18 = this.Nb + var3_3[2];
                                                                                                                                                                                                                                                                    var15_19 = var7_7[0];
                                                                                                                                                                                                                                                                    var16_20 = var7_7[1];
                                                                                                                                                                                                                                                                    var17_21 = var7_7[2];
                                                                                                                                                                                                                                                                    var18_22 = var8_8[0];
                                                                                                                                                                                                                                                                    var19_23 = var8_8[1];
                                                                                                                                                                                                                                                                    var20_29 = var8_8[2];
                                                                                                                                                                                                                                                                    var21_30 = this.wb + (this.Nb + -1);
                                                                                                                                                                                                                                                                    var22_33 = 0;
                                                                                                                                                                                                                                                                    var23_34 = 0;
                                                                                                                                                                                                                                                                    var24_35 = 0;
                                                                                                                                                                                                                                                                    var25_36 = 0;
                                                                                                                                                                                                                                                                    var26_37 = 12345678;
                                                                                                                                                                                                                                                                    var27_38 = -12345678;
                                                                                                                                                                                                                                                                    if (var12_13 != var14_18) break block254;
                                                                                                                                                                                                                                                                    break block255;
                                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                                if (~var14_18 < ~var12_13) break block336;
                                                                                                                                                                                                                                                                var24_35 = var20_29 << 1007665448;
                                                                                                                                                                                                                                                                var26_37 = var14_18;
                                                                                                                                                                                                                                                                var27_38 = var12_13;
                                                                                                                                                                                                                                                                var22_33 = var17_21 << -50385688;
                                                                                                                                                                                                                                                                if (!var51_12) break block337;
                                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                                            var26_37 = var12_13;
                                                                                                                                                                                                                                                            var27_38 = var14_18;
                                                                                                                                                                                                                                                            var22_33 = var15_19 << -913509176;
                                                                                                                                                                                                                                                            var24_35 = var18_22 << -1865947224;
                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                        var25_36 = (-var18_22 + var20_29 << 1715444712) / (-var12_13 + var14_18);
                                                                                                                                                                                                                                                        var23_34 = (var17_21 - var15_19 << -1115363608) / (var14_18 + -var12_13);
                                                                                                                                                                                                                                                        if (var26_37 < 0) {
                                                                                                                                                                                                                                                            var22_33 -= var26_37 * var23_34;
                                                                                                                                                                                                                                                            var24_35 -= var25_36 * var26_37;
                                                                                                                                                                                                                                                            var26_37 = 0;
                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                        if (var27_38 > var21_30) break block256;
                                                                                                                                                                                                                                                        break block255;
                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                    var27_38 = var21_30;
                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                var28_39 = 0;
                                                                                                                                                                                                                                                var29_40 = 0;
                                                                                                                                                                                                                                                var30_41 = 0;
                                                                                                                                                                                                                                                var31_42 = 0;
                                                                                                                                                                                                                                                var32_43 = 12345678;
                                                                                                                                                                                                                                                var33_44 = -12345678;
                                                                                                                                                                                                                                                if (var12_13 == var13_17) break block258;
                                                                                                                                                                                                                                                var29_40 = (var16_20 + -var15_19 << 1472732104) / (-var12_13 + var13_17);
                                                                                                                                                                                                                                                var31_42 = (-var18_22 + var19_23 << -939950616) / (-var12_13 + var13_17);
                                                                                                                                                                                                                                                if (var13_17 <= var12_13) break block338;
                                                                                                                                                                                                                                                var30_41 = var18_22 << 2040011112;
                                                                                                                                                                                                                                                var33_44 = var13_17;
                                                                                                                                                                                                                                                var28_39 = var15_19 << -1833655000;
                                                                                                                                                                                                                                                var32_43 = var12_13;
                                                                                                                                                                                                                                                if (!var51_12) break block339;
                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                            var30_41 = var19_23 << 977843752;
                                                                                                                                                                                                                                            var32_43 = var13_17;
                                                                                                                                                                                                                                            var28_39 = var16_20 << -120706904;
                                                                                                                                                                                                                                            var33_44 = var12_13;
                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                        if (var33_44 > var21_30) {
                                                                                                                                                                                                                                            var33_44 = var21_30;
                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                        if (~var32_43 > -1) break block257;
                                                                                                                                                                                                                                        break block258;
                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                    var30_41 -= var31_42 * var32_43;
                                                                                                                                                                                                                                    var28_39 -= var32_43 * var29_40;
                                                                                                                                                                                                                                    var32_43 = 0;
                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                var34_45 = 0;
                                                                                                                                                                                                                                var35_46 = 0;
                                                                                                                                                                                                                                var36_47 = 0;
                                                                                                                                                                                                                                var37_48 = 0;
                                                                                                                                                                                                                                var38_49 = 12345678;
                                                                                                                                                                                                                                var39_50 = -12345678;
                                                                                                                                                                                                                                if (~var13_17 != ~var14_18) break block259;
                                                                                                                                                                                                                                break block260;
                                                                                                                                                                                                                            }
                                                                                                                                                                                                                            if (var14_18 <= var13_17) break block340;
                                                                                                                                                                                                                            var34_45 = var16_20 << -851716920;
                                                                                                                                                                                                                            var38_49 = var13_17;
                                                                                                                                                                                                                            var36_47 = var19_23 << -1528838520;
                                                                                                                                                                                                                            var39_50 = var14_18;
                                                                                                                                                                                                                            if (!var51_12) break block341;
                                                                                                                                                                                                                        }
                                                                                                                                                                                                                        var38_49 = var14_18;
                                                                                                                                                                                                                        var36_47 = var20_29 << 8909512;
                                                                                                                                                                                                                        var39_50 = var13_17;
                                                                                                                                                                                                                        var34_45 = var17_21 << -1140950328;
                                                                                                                                                                                                                    }
                                                                                                                                                                                                                    var37_48 = (var20_29 - var19_23 << 850108424) / (var14_18 + -var13_17);
                                                                                                                                                                                                                    var35_46 = (var17_21 + -var16_20 << -853884728) / (var14_18 - var13_17);
                                                                                                                                                                                                                    if (-1 < ~var38_49) break block261;
                                                                                                                                                                                                                    break block262;
                                                                                                                                                                                                                }
                                                                                                                                                                                                                var36_47 -= var38_49 * var37_48;
                                                                                                                                                                                                                var34_45 -= var35_46 * var38_49;
                                                                                                                                                                                                                var38_49 = 0;
                                                                                                                                                                                                            }
                                                                                                                                                                                                            if (var21_30 < var39_50) break block263;
                                                                                                                                                                                                            break block260;
                                                                                                                                                                                                        }
                                                                                                                                                                                                        var39_50 = var21_30;
                                                                                                                                                                                                    }
                                                                                                                                                                                                    this.Xb = var26_37;
                                                                                                                                                                                                    if (~var32_43 > ~this.Xb) break block264;
                                                                                                                                                                                                    break block265;
                                                                                                                                                                                                }
                                                                                                                                                                                                this.Xb = var32_43;
                                                                                                                                                                                            }
                                                                                                                                                                                            if (~var38_49 > ~this.Xb) {
                                                                                                                                                                                                this.Xb = var38_49;
                                                                                                                                                                                            }
                                                                                                                                                                                            this.Cb = var27_38;
                                                                                                                                                                                            if (var33_44 > this.Cb) {
                                                                                                                                                                                                this.Cb = var33_44;
                                                                                                                                                                                            }
                                                                                                                                                                                            if (~var39_50 < ~this.Cb) {
                                                                                                                                                                                                this.Cb = var39_50;
                                                                                                                                                                                            }
                                                                                                                                                                                            var40_51 = 0;
                                                                                                                                                                                            for (var4_4 = this.Xb; this.Cb > var4_4; ++var4_4) {
                                                                                                                                                                                                block272: {
                                                                                                                                                                                                    block276: {
                                                                                                                                                                                                        block275: {
                                                                                                                                                                                                            block274: {
                                                                                                                                                                                                                block273: {
                                                                                                                                                                                                                    block344: {
                                                                                                                                                                                                                        block269: {
                                                                                                                                                                                                                            block271: {
                                                                                                                                                                                                                                block270: {
                                                                                                                                                                                                                                    block343: {
                                                                                                                                                                                                                                        block342: {
                                                                                                                                                                                                                                            block268: {
                                                                                                                                                                                                                                                block267: {
                                                                                                                                                                                                                                                    v10 = ~var26_37;
                                                                                                                                                                                                                                                    v11 = ~var4_4;
                                                                                                                                                                                                                                                    if (var51_12) break block266;
                                                                                                                                                                                                                                                    if (v10 < v11) break block267;
                                                                                                                                                                                                                                                    if (~var27_38 < ~var4_4) break block268;
                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                var1_1 = 655360;
                                                                                                                                                                                                                                                var5_5 = -655360;
                                                                                                                                                                                                                                                if (!var51_12) break block342;
                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                            var1_1 = var5_5 = var22_33;
                                                                                                                                                                                                                                            var9_9 = var40_51 = var24_35;
                                                                                                                                                                                                                                            var22_33 += var23_34;
                                                                                                                                                                                                                                            var24_35 += var25_36;
                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                        if (var32_43 > var4_4) break block269;
                                                                                                                                                                                                                                        if (var4_4 >= var33_44) break block269;
                                                                                                                                                                                                                                        break block343;
                                                                                                                                                                                                                                        catch (RuntimeException v14) {
                                                                                                                                                                                                                                            throw v14;
                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                    if (~var28_39 < ~var5_5) break block270;
                                                                                                                                                                                                                                    break block271;
                                                                                                                                                                                                                                    catch (RuntimeException v15) {
                                                                                                                                                                                                                                        throw v15;
                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                var5_5 = var28_39;
                                                                                                                                                                                                                                var40_51 = var30_41;
                                                                                                                                                                                                                            }
                                                                                                                                                                                                                            if (var28_39 < var1_1) {
                                                                                                                                                                                                                                var1_1 = var28_39;
                                                                                                                                                                                                                                var9_9 = var30_41;
                                                                                                                                                                                                                            }
                                                                                                                                                                                                                            var30_41 += var31_42;
                                                                                                                                                                                                                            var28_39 += var29_40;
                                                                                                                                                                                                                        }
                                                                                                                                                                                                                        if (var4_4 < var38_49) break block272;
                                                                                                                                                                                                                        if (var39_50 <= var4_4) break block272;
                                                                                                                                                                                                                        break block344;
                                                                                                                                                                                                                        catch (RuntimeException v17) {
                                                                                                                                                                                                                            throw v17;
                                                                                                                                                                                                                        }
                                                                                                                                                                                                                    }
                                                                                                                                                                                                                    if (~var5_5 > ~var34_45) break block273;
                                                                                                                                                                                                                    break block274;
                                                                                                                                                                                                                    catch (RuntimeException v18) {
                                                                                                                                                                                                                        throw v18;
                                                                                                                                                                                                                    }
                                                                                                                                                                                                                }
                                                                                                                                                                                                                var40_51 = var36_47;
                                                                                                                                                                                                                var5_5 = var34_45;
                                                                                                                                                                                                            }
                                                                                                                                                                                                            if (var34_45 < var1_1) break block275;
                                                                                                                                                                                                            break block276;
                                                                                                                                                                                                        }
                                                                                                                                                                                                        var1_1 = var34_45;
                                                                                                                                                                                                        var9_9 = var36_47;
                                                                                                                                                                                                    }
                                                                                                                                                                                                    var36_47 += var37_48;
                                                                                                                                                                                                    var34_45 += var35_46;
                                                                                                                                                                                                }
                                                                                                                                                                                                var41_52 = this.x[var4_4];
                                                                                                                                                                                                var41_52.e = var9_9;
                                                                                                                                                                                                var41_52.l = var40_51;
                                                                                                                                                                                                var41_52.d = var1_1;
                                                                                                                                                                                                var41_52.k = var5_5;
                                                                                                                                                                                                if (!var51_12) continue;
                                                                                                                                                                                            }
                                                                                                                                                                                            v10 = ~(-this.wb + this.Nb);
                                                                                                                                                                                            v11 = ~this.Xb;
                                                                                                                                                                                        }
                                                                                                                                                                                        if (v10 < v11) {
                                                                                                                                                                                            this.Xb = this.Nb + -this.wb;
                                                                                                                                                                                        }
                                                                                                                                                                                        if (!var51_12) break block277;
lbl252:
                                                                                                                                                                                        // 2 sources

                                                                                                                                                                                        if (var11_11 == 4) break block278;
                                                                                                                                                                                        this.Xb = var3_3[0] = var3_3[0] + this.Nb;
                                                                                                                                                                                        this.Cb = var3_3[0];
                                                                                                                                                                                        var4_4 = 1;
                                                                                                                                                                                        while (~var4_4 > ~var11_11) {
                                                                                                                                                                                            block281: {
                                                                                                                                                                                                block280: {
                                                                                                                                                                                                    v22 = var4_4;
                                                                                                                                                                                                    v23 = var3_3[v22] + this.Nb;
                                                                                                                                                                                                    var3_3[v22] = v23;
                                                                                                                                                                                                    v24 = var12_13 = v23;
                                                                                                                                                                                                    v25 = this.Xb;
                                                                                                                                                                                                    if (var51_12) break block279;
                                                                                                                                                                                                    if (v24 >= v25) {
                                                                                                                                                                                                    }
                                                                                                                                                                                                    ** GOTO lbl281
                                                                                                                                                                                                    if (this.Cb < var12_13) break block280;
                                                                                                                                                                                                    break block281;
                                                                                                                                                                                                }
                                                                                                                                                                                                this.Cb = var12_13;
                                                                                                                                                                                                if (!var51_12) break block281;
lbl281:
                                                                                                                                                                                                // 2 sources

                                                                                                                                                                                                this.Xb = var12_13;
                                                                                                                                                                                            }
                                                                                                                                                                                            ++var4_4;
                                                                                                                                                                                            if (!var51_12) continue;
                                                                                                                                                                                        }
                                                                                                                                                                                        v24 = ~(this.Nb + this.wb);
                                                                                                                                                                                        v25 = ~this.Cb;
                                                                                                                                                                                    }
                                                                                                                                                                                    if (v24 >= v25) {
                                                                                                                                                                                        this.Cb = -1 + this.Nb - -this.wb;
                                                                                                                                                                                    }
                                                                                                                                                                                    if (this.Nb + -this.wb > this.Xb) {
                                                                                                                                                                                        this.Xb = -this.wb + this.Nb;
                                                                                                                                                                                    }
                                                                                                                                                                                    if (this.Xb >= this.Cb) {
                                                                                                                                                                                        return;
                                                                                                                                                                                    }
                                                                                                                                                                                    for (var4_4 = this.Xb; var4_4 < this.Cb; ++var4_4) {
                                                                                                                                                                                        var12_14 = this.x[var4_4];
                                                                                                                                                                                        var12_14.k = -655360;
                                                                                                                                                                                        var12_14.d = 655360;
                                                                                                                                                                                        if (!var51_12) continue;
                                                                                                                                                                                    }
                                                                                                                                                                                    var12_13 = var11_11 + -1;
                                                                                                                                                                                    var13_17 = var3_3[0];
                                                                                                                                                                                    var14_18 = var3_3[var12_13];
                                                                                                                                                                                    if (var13_17 < var14_18) break block282;
                                                                                                                                                                                    if (var14_18 < var13_17) break block283;
                                                                                                                                                                                    break block284;
                                                                                                                                                                                    catch (RuntimeException v30) {
                                                                                                                                                                                        throw v30;
                                                                                                                                                                                    }
                                                                                                                                                                                }
                                                                                                                                                                                var15_19 = var7_7[var12_13] << 171341224;
                                                                                                                                                                                var16_20 = (-var7_7[var12_13] + var7_7[0] << -289186712) / (-var14_18 + var13_17);
                                                                                                                                                                                var17_21 = var8_8[var12_13] << -1081894616;
                                                                                                                                                                                var18_22 = (var8_8[0] - var8_8[var12_13] << -1660313240) / (var13_17 + -var14_18);
                                                                                                                                                                                if (~this.Cb > ~var13_17) break block285;
                                                                                                                                                                                break block286;
                                                                                                                                                                            }
                                                                                                                                                                            var13_17 = this.Cb;
                                                                                                                                                                        }
                                                                                                                                                                        if (-1 < ~var14_18) break block287;
                                                                                                                                                                        break block288;
                                                                                                                                                                    }
                                                                                                                                                                    var17_21 -= var18_22 * var14_18;
                                                                                                                                                                    var15_19 -= var16_20 * var14_18;
                                                                                                                                                                    var14_18 = 0;
                                                                                                                                                                }
                                                                                                                                                                for (var4_4 = var14_18; var4_4 <= var13_17; ++var4_4) {
                                                                                                                                                                    var19_24 = this.x[var4_4];
                                                                                                                                                                    var19_24.d = var19_24.k = var15_19;
                                                                                                                                                                    var19_24.e = var19_24.l = var17_21;
                                                                                                                                                                    var15_19 += var16_20;
                                                                                                                                                                    var17_21 += var18_22;
                                                                                                                                                                    if (!var51_12) {
                                                                                                                                                                        if (!var51_12) continue;
                                                                                                                                                                        break;
                                                                                                                                                                    }
                                                                                                                                                                    break block284;
                                                                                                                                                                }
                                                                                                                                                                if (!var51_12) break block284;
                                                                                                                                                            }
                                                                                                                                                            var15_19 = var7_7[0] << -1640390552;
                                                                                                                                                            var16_20 = (var7_7[var12_13] + -var7_7[0] << 1369174728) / (var14_18 + -var13_17);
                                                                                                                                                            var17_21 = var8_8[0] << -1786427608;
                                                                                                                                                            var18_22 = (-var8_8[0] + var8_8[var12_13] << 740046088) / (var14_18 - var13_17);
                                                                                                                                                            if (0 > var13_17) break block289;
                                                                                                                                                            break block290;
                                                                                                                                                        }
                                                                                                                                                        var15_19 -= var13_17 * var16_20;
                                                                                                                                                        var17_21 -= var13_17 * var18_22;
                                                                                                                                                        var13_17 = 0;
                                                                                                                                                    }
                                                                                                                                                    if (var14_18 > this.Cb) break block291;
                                                                                                                                                    break block292;
                                                                                                                                                }
                                                                                                                                                var14_18 = this.Cb;
                                                                                                                                            }
                                                                                                                                            for (var4_4 = var13_17; var4_4 <= var14_18; ++var4_4) {
                                                                                                                                                var19_25 = this.x[var4_4];
                                                                                                                                                var19_25.e = var19_25.l = var17_21;
                                                                                                                                                var19_25.d = var19_25.k = var15_19;
                                                                                                                                                var15_19 += var16_20;
                                                                                                                                                var17_21 += var18_22;
                                                                                                                                                if (!var51_12) {
                                                                                                                                                    if (!var51_12) continue;
                                                                                                                                                    break;
                                                                                                                                                }
                                                                                                                                                break block293;
                                                                                                                                            }
                                                                                                                                        }
                                                                                                                                        var4_4 = 0;
                                                                                                                                    }
                                                                                                                                    block157: while (true) {
                                                                                                                                        v38 = ~var12_13;
                                                                                                                                        v39 = ~var4_4;
                                                                                                                                        block158: while (v38 < v39) {
                                                                                                                                            block297: {
                                                                                                                                                block303: {
                                                                                                                                                    block302: {
                                                                                                                                                        block301: {
                                                                                                                                                            block300: {
                                                                                                                                                                block295: {
                                                                                                                                                                    block299: {
                                                                                                                                                                        block298: {
                                                                                                                                                                            block296: {
                                                                                                                                                                                var13_17 = var3_3[var4_4];
                                                                                                                                                                                var15_19 = var4_4 - -1;
                                                                                                                                                                                var14_18 = var3_3[var15_19];
                                                                                                                                                                                v40 = ~var13_17;
                                                                                                                                                                                v41 = ~var14_18;
                                                                                                                                                                                if (var51_12) break block294;
                                                                                                                                                                                if (v40 > v41) break block295;
                                                                                                                                                                                if (var13_17 > var14_18) break block296;
                                                                                                                                                                                break block297;
                                                                                                                                                                            }
                                                                                                                                                                            var16_20 = var7_7[var15_19] << 625101064;
                                                                                                                                                                            var17_21 = (var7_7[var4_4] - var7_7[var15_19] << -584988376) / (-var14_18 + var13_17);
                                                                                                                                                                            var18_22 = var8_8[var15_19] << 203973320;
                                                                                                                                                                            var19_26 = (-var8_8[var15_19] + var8_8[var4_4] << -696921336) / (-var14_18 + var13_17);
                                                                                                                                                                            if (~var14_18 > -1) break block298;
                                                                                                                                                                            break block299;
                                                                                                                                                                        }
                                                                                                                                                                        var16_20 -= var17_21 * var14_18;
                                                                                                                                                                        var18_22 -= var14_18 * var19_26;
                                                                                                                                                                        var14_18 = 0;
                                                                                                                                                                    }
                                                                                                                                                                    if (~this.Cb > ~var13_17) {
                                                                                                                                                                        var13_17 = this.Cb;
                                                                                                                                                                    }
                                                                                                                                                                    for (var20_29 = var14_18; var13_17 >= var20_29; ++var20_29) {
                                                                                                                                                                        var21_31 = this.x[var20_29];
                                                                                                                                                                        v38 = var16_20;
                                                                                                                                                                        v39 = var21_31.d;
                                                                                                                                                                        if (var51_12) continue block158;
                                                                                                                                                                        if (v38 < v39) {
                                                                                                                                                                            var21_31.e = var18_22;
                                                                                                                                                                            var21_31.d = var16_20;
                                                                                                                                                                        }
                                                                                                                                                                        if (var16_20 > var21_31.k) {
                                                                                                                                                                            var21_31.l = var18_22;
                                                                                                                                                                            var21_31.k = var16_20;
                                                                                                                                                                        }
                                                                                                                                                                        var18_22 += var19_26;
                                                                                                                                                                        var16_20 += var17_21;
                                                                                                                                                                        if (!var51_12) continue;
                                                                                                                                                                    }
                                                                                                                                                                    if (!var51_12) break block297;
                                                                                                                                                                }
                                                                                                                                                                var16_20 = var7_7[var4_4] << -104335448;
                                                                                                                                                                var17_21 = (var7_7[var15_19] - var7_7[var4_4] << 383743496) / (-var13_17 + var14_18);
                                                                                                                                                                var18_22 = var8_8[var4_4] << 313160584;
                                                                                                                                                                var19_27 = (var8_8[var15_19] + -var8_8[var4_4] << -923644504) / (-var13_17 + var14_18);
                                                                                                                                                                if (var14_18 > this.Cb) break block300;
                                                                                                                                                                break block301;
                                                                                                                                                            }
                                                                                                                                                            var14_18 = this.Cb;
                                                                                                                                                        }
                                                                                                                                                        if (-1 < ~var13_17) break block302;
                                                                                                                                                        break block303;
                                                                                                                                                    }
                                                                                                                                                    var16_20 -= var13_17 * var17_21;
                                                                                                                                                    var18_22 -= var13_17 * var19_27;
                                                                                                                                                    var13_17 = 0;
                                                                                                                                                }
                                                                                                                                                var20_29 = var13_17;
                                                                                                                                                while (~var20_29 >= ~var14_18) {
                                                                                                                                                    block305: {
                                                                                                                                                        block304: {
                                                                                                                                                            var21_31 = this.x[var20_29];
                                                                                                                                                            v38 = var16_20;
                                                                                                                                                            v39 = var21_31.k;
                                                                                                                                                            if (var51_12) continue block158;
                                                                                                                                                            if (v38 > v39) {
                                                                                                                                                                var21_31.k = var16_20;
                                                                                                                                                                var21_31.l = var18_22;
                                                                                                                                                            }
                                                                                                                                                            if (var16_20 < var21_31.d) break block304;
                                                                                                                                                            break block305;
                                                                                                                                                        }
                                                                                                                                                        var21_31.d = var16_20;
                                                                                                                                                        var21_31.e = var18_22;
                                                                                                                                                    }
                                                                                                                                                    var18_22 += var19_27;
                                                                                                                                                    var16_20 += var17_21;
                                                                                                                                                    ++var20_29;
                                                                                                                                                    if (!var51_12) continue;
                                                                                                                                                }
                                                                                                                                            }
                                                                                                                                            ++var4_4;
                                                                                                                                            if (!var51_12) continue block157;
                                                                                                                                        }
                                                                                                                                        break;
                                                                                                                                    }
                                                                                                                                    v40 = -this.wb + this.Nb;
                                                                                                                                    v41 = this.Xb;
                                                                                                                                }
                                                                                                                                if (v40 > v41) {
                                                                                                                                    this.Xb = -this.wb + this.Nb;
                                                                                                                                }
                                                                                                                                if (!var51_12) break block277;
                                                                                                                            }
                                                                                                                            var12_13 = var3_3[0] - -this.Nb;
                                                                                                                            var13_17 = this.Nb + var3_3[1];
                                                                                                                            var14_18 = this.Nb + var3_3[2];
                                                                                                                            var15_19 = this.Nb + var3_3[3];
                                                                                                                            var16_20 = var7_7[0];
                                                                                                                            var17_21 = var7_7[1];
                                                                                                                            var18_22 = var7_7[2];
                                                                                                                            var19_28 = var7_7[3];
                                                                                                                            var20_29 = var8_8[0];
                                                                                                                            var21_32 = var8_8[1];
                                                                                                                            var22_33 = var8_8[2];
                                                                                                                            var23_34 = var8_8[3];
                                                                                                                            var24_35 = this.wb + this.Nb + -1;
                                                                                                                            var25_36 = 0;
                                                                                                                            var26_37 = 0;
                                                                                                                            var27_38 = 0;
                                                                                                                            var28_39 = 0;
                                                                                                                            var29_40 = 12345678;
                                                                                                                            var30_41 = -12345678;
                                                                                                                            if (~var12_13 != ~var15_19) break block306;
                                                                                                                            break block307;
                                                                                                                        }
                                                                                                                        var26_37 = (-var16_20 + var19_28 << 1731726120) / (var15_19 + -var12_13);
                                                                                                                        var28_39 = (var23_34 + -var20_29 << -1470863032) / (var15_19 + -var12_13);
                                                                                                                        if (var15_19 > var12_13) break block345;
                                                                                                                        var29_40 = var15_19;
                                                                                                                        var25_36 = var19_28 << -125722552;
                                                                                                                        var27_38 = var23_34 << -2097471832;
                                                                                                                        var30_41 = var12_13;
                                                                                                                        if (!var51_12) break block346;
                                                                                                                    }
                                                                                                                    var30_41 = var15_19;
                                                                                                                    var25_36 = var16_20 << -1160973656;
                                                                                                                    var29_40 = var12_13;
                                                                                                                    var27_38 = var20_29 << 84620424;
                                                                                                                }
                                                                                                                if (~var29_40 > -1) {
                                                                                                                    var27_38 -= var28_39 * var29_40;
                                                                                                                    var25_36 -= var29_40 * var26_37;
                                                                                                                    var29_40 = 0;
                                                                                                                }
                                                                                                                if (var24_35 < var30_41) {
                                                                                                                    var30_41 = var24_35;
                                                                                                                }
                                                                                                            }
                                                                                                            var31_42 = 0;
                                                                                                            var32_43 = 0;
                                                                                                            var33_44 = 0;
                                                                                                            var34_45 = 0;
                                                                                                            var35_46 = 12345678;
                                                                                                            var36_47 = -12345678;
                                                                                                            if (~var13_17 == ~var12_13) break block347;
                                                                                                            var34_45 = (var21_32 - var20_29 << 1153959368) / (var13_17 - var12_13);
                                                                                                            if (~var12_13 > ~var13_17) break block348;
                                                                                                            var35_46 = var13_17;
                                                                                                            var33_44 = var21_32 << -981790296;
                                                                                                            var36_47 = var12_13;
                                                                                                            var31_42 = var17_21 << 1316747656;
                                                                                                            if (!var51_12) break block349;
                                                                                                        }
                                                                                                        var35_46 = var12_13;
                                                                                                        var36_47 = var13_17;
                                                                                                        var31_42 = var16_20 << -1538475832;
                                                                                                        var33_44 = var20_29 << 632320808;
                                                                                                    }
                                                                                                    var32_43 = (-var16_20 + var17_21 << 1393772040) / (-var12_13 + var13_17);
                                                                                                    if (var24_35 < var36_47) {
                                                                                                        var36_47 = var24_35;
                                                                                                    }
                                                                                                    if (var35_46 < 0) {
                                                                                                        var31_42 -= var35_46 * var32_43;
                                                                                                        var33_44 -= var34_45 * var35_46;
                                                                                                        var35_46 = 0;
                                                                                                    }
                                                                                                }
                                                                                                var37_48 = 0;
                                                                                                var38_49 = 0;
                                                                                                var39_50 = 0;
                                                                                                var40_51 = 0;
                                                                                                var41_53 = 12345678;
                                                                                                var42_54 = -12345678;
                                                                                                if (var14_18 == var13_17) break block309;
                                                                                                var40_51 = (-var21_32 + var22_33 << -1549613272) / (-var13_17 + var14_18);
                                                                                                if (~var13_17 > ~var14_18) break block350;
                                                                                                var41_53 = var14_18;
                                                                                                var39_50 = var22_33 << 83408072;
                                                                                                var37_48 = var18_22 << -1865167864;
                                                                                                var42_54 = var13_17;
                                                                                                if (!var51_12) break block351;
                                                                                            }
                                                                                            var41_53 = var13_17;
                                                                                            var39_50 = var21_32 << 802252776;
                                                                                            var42_54 = var14_18;
                                                                                            var37_48 = var17_21 << 2124457512;
                                                                                        }
                                                                                        var38_49 = (-var17_21 + var18_22 << 889486952) / (var14_18 - var13_17);
                                                                                        if (-1 < ~var41_53) {
                                                                                            var39_50 -= var41_53 * var40_51;
                                                                                            var37_48 -= var38_49 * var41_53;
                                                                                            var41_53 = 0;
                                                                                        }
                                                                                        if (~var42_54 < ~var24_35) break block308;
                                                                                        break block309;
                                                                                    }
                                                                                    var42_54 = var24_35;
                                                                                }
                                                                                var43_55 = 0;
                                                                                var44_56 = 0;
                                                                                var45_57 = 0;
                                                                                var46_58 = 0;
                                                                                var47_59 = 12345678;
                                                                                var48_60 = -12345678;
                                                                                if (~var14_18 == ~var15_19) break block313;
                                                                                var46_58 = (-var22_33 + var23_34 << -391748568) / (-var14_18 + var15_19);
                                                                                if (~var15_19 < ~var14_18) break block352;
                                                                                var48_60 = var14_18;
                                                                                var45_57 = var23_34 << -425130424;
                                                                                var43_55 = var19_28 << 430633672;
                                                                                var47_59 = var15_19;
                                                                                if (!var51_12) break block353;
                                                                            }
                                                                            var45_57 = var22_33 << -1164816632;
                                                                            var48_60 = var15_19;
                                                                            var47_59 = var14_18;
                                                                            var43_55 = var18_22 << 1895169448;
                                                                        }
                                                                        var44_56 = (-var18_22 + var19_28 << -2111397368) / (var15_19 + -var14_18);
                                                                        if (-1 < ~var47_59) break block310;
                                                                        break block311;
                                                                    }
                                                                    var43_55 -= var47_59 * var44_56;
                                                                    var45_57 -= var47_59 * var46_58;
                                                                    var47_59 = 0;
                                                                }
                                                                if (~var48_60 < ~var24_35) break block312;
                                                                break block313;
                                                            }
                                                            var48_60 = var24_35;
                                                        }
                                                        this.Xb = var29_40;
                                                        if (this.Xb > var35_46) {
                                                            this.Xb = var35_46;
                                                        }
                                                        if (~this.Xb < ~var41_53) {
                                                            this.Xb = var41_53;
                                                        }
                                                        this.Cb = var30_41;
                                                        if (~var47_59 > ~this.Xb) {
                                                            this.Xb = var47_59;
                                                        }
                                                        if (var36_47 > this.Cb) {
                                                            this.Cb = var36_47;
                                                        }
                                                        if (var42_54 > this.Cb) break block314;
                                                        break block315;
                                                    }
                                                    this.Cb = var42_54;
                                                }
                                                if (~var48_60 < ~this.Cb) break block316;
                                                break block317;
                                            }
                                            this.Cb = var48_60;
                                        }
                                        var49_61 = 0;
                                        for (var4_4 = this.Xb; this.Cb > var4_4; ++var4_4) {
                                            block328: {
                                                block332: {
                                                    block331: {
                                                        block330: {
                                                            block329: {
                                                                block324: {
                                                                    block327: {
                                                                        block326: {
                                                                            block325: {
                                                                                block321: {
                                                                                    block323: {
                                                                                        block322: {
                                                                                            block355: {
                                                                                                block354: {
                                                                                                    block320: {
                                                                                                        block319: {
                                                                                                            v62 = ~var29_40;
                                                                                                            v63 = ~var4_4;
                                                                                                            if (var51_12) break block318;
                                                                                                            if (v62 < v63) break block319;
                                                                                                            if (var30_41 > var4_4) break block320;
                                                                                                        }
                                                                                                        var5_5 = -655360;
                                                                                                        var1_1 = 655360;
                                                                                                        if (!var51_12) break block354;
                                                                                                    }
                                                                                                    var9_9 = var49_61 = var27_38;
                                                                                                    var1_1 = var5_5 = var25_36;
                                                                                                    var27_38 += var28_39;
                                                                                                    var25_36 += var26_37;
                                                                                                }
                                                                                                if (var35_46 > var4_4) break block321;
                                                                                                if (~var4_4 <= ~var36_47) break block321;
                                                                                                break block355;
                                                                                                catch (RuntimeException v66) {
                                                                                                    throw v66;
                                                                                                }
                                                                                            }
                                                                                            if (~var1_1 < ~var31_42) break block322;
                                                                                            break block323;
                                                                                            catch (RuntimeException v67) {
                                                                                                throw v67;
                                                                                            }
                                                                                        }
                                                                                        var9_9 = var33_44;
                                                                                        var1_1 = var31_42;
                                                                                    }
                                                                                    if (var5_5 < var31_42) {
                                                                                        var49_61 = var33_44;
                                                                                        var5_5 = var31_42;
                                                                                    }
                                                                                    var31_42 += var32_43;
                                                                                    var33_44 += var34_45;
                                                                                }
                                                                                if (var4_4 < var41_53) break block324;
                                                                                if (~var42_54 < ~var4_4) break block325;
                                                                                break block324;
                                                                                catch (RuntimeException v69) {
                                                                                    throw v69;
                                                                                }
                                                                            }
                                                                            if (var37_48 > var5_5) break block326;
                                                                            break block327;
                                                                        }
                                                                        var5_5 = var37_48;
                                                                        var49_61 = var39_50;
                                                                    }
                                                                    if (~var1_1 < ~var37_48) {
                                                                        var1_1 = var37_48;
                                                                        var9_9 = var39_50;
                                                                    }
                                                                    var37_48 += var38_49;
                                                                    var39_50 += var40_51;
                                                                }
                                                                if (var47_59 > var4_4) break block328;
                                                                if (~var4_4 > ~var48_60) break block329;
                                                                break block328;
                                                                catch (RuntimeException v72) {
                                                                    throw v72;
                                                                }
                                                            }
                                                            if (var43_55 <= var5_5) break block330;
                                                            var49_61 = var45_57;
                                                            var5_5 = var43_55;
                                                        }
                                                        if (~var1_1 < ~var43_55) break block331;
                                                        break block332;
                                                    }
                                                    var1_1 = var43_55;
                                                    var9_9 = var45_57;
                                                }
                                                var45_57 += var46_58;
                                                var43_55 += var44_56;
                                            }
                                            var50_62 = this.x[var4_4];
                                            var50_62.e = var9_9;
                                            var50_62.d = var1_1;
                                            var50_62.k = var5_5;
                                            var50_62.l = var49_61;
                                            if (!var51_12) continue;
                                        }
                                        v62 = -this.wb + this.Nb;
                                        v63 = this.Xb;
                                    }
                                    if (v62 > v63) {
                                        this.Xb = this.Nb + -this.wb;
                                    }
                                }
                                if (var10_10 != 5960) {
                                    return;
                                }
                                if (!this.K) break block333;
                                if (~this.db >= ~this.cc) break block333;
                                if (~this.Xb < ~this.Wb) break block333;
                                break block356;
                                catch (RuntimeException v76) {
                                    throw v76;
                                }
                            }
                            if (~this.Wb <= ~this.Cb) break block333;
                            var12_15 = this.x[this.Wb];
                            if (this.j < var12_15.d >> -1650257144) break block333;
                            if (this.j > var12_15.k >> -1682750680) break block333;
                            break block357;
                            catch (RuntimeException v78) {
                                throw v78;
                            }
                        }
                        if (~var12_15.d >= ~var12_15.k) break block334;
                        break block333;
                        catch (RuntimeException v79) {
                            throw v79;
                        }
                    }
                    if (var6_6.db) break block333;
                    if (-1 == ~var6_6.zb[var2_2]) break block335;
                    break block333;
                    catch (RuntimeException v81) {
                        throw v81;
                    }
                }
                this.Ab[this.cc] = var6_6;
                this.qb[this.cc] = var2_2;
                ++this.cc;
            }
            catch (RuntimeException var12_16) {
                v83 = var12_16;
                v84 = new StringBuilder().append(lb.N[26]).append(var1_1).append(',').append(var2_2).append(',');
                v85 = var3_3 != null ? lb.N[0] : lb.N[1];
                v87 = v84.append(v85).append(',').append(var4_4).append(',').append(var5_5).append(',');
                v88 = var6_6 != null ? lb.N[0] : lb.N[1];
                v90 = v87.append(v88).append(',');
                v91 = var7_7 != null ? lb.N[0] : lb.N[1];
                v93 = v90.append(v91).append(',');
                v94 = var8_8 != null ? lb.N[0] : lb.N[1];
                throw i.a(v83, v93.append(v94).append(',').append(var9_9).append(',').append(var10_10).append(',').append(var11_11).append(')').toString());
            }
        }
    }

    final void a(ca ca2, int n2) {
        block12: {
            boolean bl = client.vh;
            try {
                if (n2 != -1) {
                    this.jb = null;
                }
                int n3 = 0;
                while (n3 < this.ab) {
                    block14: {
                        block13: {
                            if (bl) break block12;
                            if (this.Z[n3] != ca2) break block13;
                            --this.ab;
                            int n4 = n3;
                            while (~this.ab < ~n4) {
                                this.Z[n4] = this.Z[n4 - -1];
                                this.jb[n4] = this.jb[1 + n4];
                                ++n4;
                                if (!bl) {
                                    if (!bl) continue;
                                    break;
                                }
                                break block14;
                            }
                        }
                        ++n3;
                    }
                    if (!bl) continue;
                }
                ++m;
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(N[38]);
                String string = ca2 != null ? N[0] : N[1];
                throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(')').toString());
            }
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, int n3) {
        boolean bl = client.vh;
        try {
            int n4;
            int n5;
            int n6;
            int n7;
            int n8;
            int n9;
            int n10;
            int n11;
            int n12;
            int n13;
            int[] nArray;
            ca ca2;
            block34: {
                int n14;
                block33: {
                    ++l;
                    w w2 = this.y[n2];
                    ca2 = w2.o;
                    int n15 = w2.i;
                    nArray = ca2.o[n15];
                    n13 = ca2.lb[n15];
                    n14 = ca2.M[n15];
                    n12 = ca2.cc[nArray[0]];
                    n11 = ca2.H[nArray[0]];
                    n10 = ca2.bb[nArray[0]];
                    int n16 = -n12 + ca2.cc[nArray[1]];
                    int n17 = -n11 + ca2.H[nArray[1]];
                    int n18 = -n10 + ca2.bb[nArray[1]];
                    int n19 = -n12 + ca2.cc[nArray[2]];
                    int n20 = ca2.H[nArray[2]] + -n11;
                    int n21 = -n10 + ca2.bb[nArray[2]];
                    n9 = -(n18 * n20) + n21 * n17;
                    n8 = -(n16 * n21) + n19 * n18;
                    n7 = -(n19 * n17) + n16 * n20;
                    if (-1 == n14) break block33;
                    n7 >>= n14;
                    n9 >>= n14;
                    n8 >>= n14;
                    if (!bl) break block34;
                }
                n14 = 0;
                do {
                    int n22;
                    int n23;
                    if (-25001 <= ~n9) {
                        n23 = 25000;
                        n22 = n8;
                        if (!bl) {
                            if (n23 >= n22 && -25001 <= ~n7 && -25000 <= n9 && ~n8 <= 24999 && -25000 <= n7) break;
                        }
                    } else {
                        n9 >>= 1;
                        n8 >>= 1;
                        ++n14;
                        n23 = n7;
                        n22 = 1;
                    }
                    n7 = n23 >> n22;
                } while (!bl);
                ca2.M[n15] = n14;
                ca2.k[n15] = (int)((double)this.h * Math.sqrt(n7 * n7 + n8 * n8 + n9 * n9));
            }
            w2.k = n7;
            if (n3 != -21875) {
                this.Eb = null;
            }
            w2.r = n9;
            w2.l = n8;
            w2.s = n9 * n12 - (-(n11 * n8) + -(n7 * n10));
            int n24 = n6 = ca2.bb[nArray[0]];
            int n25 = n5 = ca2.pb[nArray[0]];
            int n26 = n4 = ca2.Ob[nArray[0]];
            for (int i2 = 1; n13 > i2; ++i2) {
                int n27;
                block38: {
                    block37: {
                        block36: {
                            block35: {
                                n27 = ca2.bb[nArray[i2]];
                                if (bl) return;
                                if (n27 > n24) break block35;
                                if (~n6 >= ~n27) break block36;
                                n6 = n27;
                                if (!bl) break block36;
                            }
                            n24 = n27;
                        }
                        if ((n27 = ca2.pb[nArray[i2]]) <= n25) break block37;
                        n25 = n27;
                        if (!bl) break block38;
                    }
                    if (~n5 < ~n27) {
                        n5 = n27;
                    }
                }
                if ((n27 = ca2.Ob[nArray[i2]]) > n26) {
                    n26 = n27;
                    if (!bl) continue;
                }
                if (n4 <= n27) continue;
                n4 = n27;
                if (!bl) continue;
            }
            w2.e = n5;
            w2.m = n25;
            w2.j = n26;
            w2.q = n24;
            w2.h = n4;
            w2.u = n6;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[30] + n2 + ',' + n3 + ')');
        }
    }

    static final void a(boolean bl, byte[] byArray) {
        try {
            if (!bl) {
                return;
            }
            ++w;
            if (null == u.d) {
                return;
            }
            u.d.setPixels(0, 0, k.o, da.bb, m.d, byArray, 0, k.o);
            u.d.imageComplete(3);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(N[3]).append(bl).append(',');
            String string = byArray != null ? N[0] : N[1];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    final void a(int n2, byte by, int[] nArray, int n3, byte[] byArray) {
        try {
            ++Rb;
            this.g[n2] = byArray;
            if (by <= 29) {
                this.b(-108);
            }
            this.L[n2] = nArray;
            this.Hb[n2] = n3;
            this.D[n2] = 0L;
            this.S[n2] = false;
            this.kb[n2] = null;
            this.b(n2, true);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(N[11]).append(n2).append(',').append(by).append(',');
            String string = nArray != null ? N[0] : N[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',').append(n3).append(',');
            String string2 = byArray != null ? N[0] : N[1];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(')').toString());
        }
    }

    final void a(int n2, boolean bl, int n3, int n4, int n5, int n6, int n7) {
        block9: {
            boolean bl2 = client.vh;
            try {
                this.R = n6;
                this.Zb = n7;
                this.vb = n3;
                this.Nb = n5;
                this.x = new n[n2 + n5];
                this.wb = n2;
                ++fb;
                this.A = n4;
                for (int i2 = 0; i2 < n5 + n2; ++i2) {
                    this.x[i2] = new n();
                    if (!bl2) {
                        if (!bl2) continue;
                        break;
                    }
                    break block9;
                }
                if (!bl) {
                    this.f = false;
                }
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, N[33] + n2 + ',' + bl + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ')');
            }
        }
    }

    final void a(int n2, int n3, int n4) {
        try {
            this.K = true;
            this.j = -this.Zb + n3;
            this.Wb = n4;
            this.cc = n2;
            ++p;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[36] + n2 + ',' + n3 + ',' + n4 + ')');
        }
    }

    final ca[] b(byte by) {
        try {
            if (by < 95) {
                return null;
            }
            ++Jb;
            return this.Ab;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[15] + by + ')');
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
            ++Db;
            if (~n4 == -1 && n3 == 0 && n2 == 0) {
                n4 = 32;
            }
            if (!bl) {
                this.c(-89);
            }
            int n5 = 0;
            do {
                if (~this.ab >= ~n5) return;
                this.Z[n5].a(false, n4, n3, n2);
                ++n5;
                if (bl2) return;
            } while (!bl2);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[22] + n2 + ',' + n3 + ',' + bl + ',' + n4 + ')');
        }
    }

    final void a(boolean bl) {
        block7: {
            boolean bl2 = client.vh;
            try {
                ++k;
                this.a(-118);
                if (bl) {
                    this.Xb = -11;
                }
                int n2 = 0;
                while (~n2 > ~this.ab) {
                    this.Z[n2] = null;
                    ++n2;
                    if (!bl2) {
                        if (!bl2) continue;
                        break;
                    }
                    break block7;
                }
                this.ab = 0;
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, N[23] + bl + ')');
            }
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final boolean a(byte by, boolean bl, int n2, int n3, int n4, int n5) {
        try {
            if (by >= -86) {
                this.jb = null;
            }
            ++Lb;
            if (!(bl && ~n5 <= ~n4 || n4 < n5)) {
                if (n4 < n3) return true;
                if (n2 < n5) return true;
                if (~n3 >= ~n2) return bl;
                return true;
            }
            if (n4 > n3) return true;
            if (n2 > n5) {
                return true;
            }
            if (n2 > n3) return true;
            if (bl) return false;
            return true;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[5] + by + ',' + bl + ',' + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    private final void a(int n2) {
        try {
            ++Sb;
            if (n2 > -115) {
                return;
            }
            this.n = 0;
            this.T.c(1);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, N[39] + n2 + ')');
        }
    }

    /*
     * Exception decompiling
     */
    private final boolean a(int[] var1_1, int[] var2_2, int[] var3_3, int[] var4_4, int var5_5) {
        /*
         * This method has failed to decompile.  When submitting a bug report, please provide this stack trace, and (if you hold appropriate legal rights) the relevant class file.
         * 
         * org.benf.cfr.reader.util.ConfusedCFRException: Tried to end blocks [102[UNCONDITIONALDOLOOP]], but top level block is 40[TRYBLOCK]
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

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    lb(ua ua2, int n2, int n3, int n4) {
        boolean bl = client.vh;
        this.nb = 5;
        this.Vb = new int[40];
        this.J = new int[40];
        this.ib = 50;
        this.f = false;
        this.G = 10;
        this.B = new int[40];
        this.cc = 0;
        this.wb = 192;
        this.Zb = 256;
        this.P = 20;
        this.Qb = new int[40];
        this.Ib = new int[this.ib][256];
        this.Nb = 256;
        this.Mb = 1000;
        this.vb = 512;
        this.r = new int[40];
        this.K = false;
        this.A = 256;
        this.cb = 0;
        this.h = 4;
        this.Ub = false;
        this.v = new int[this.ib];
        this.n = 0;
        this.yb = new int[40];
        this.X = 1000;
        this.db = 100;
        this.R = 8;
        this.Ab = new ca[this.db];
        this.qb = new int[this.db];
        try {
            int n5;
            block16: {
                block15: {
                    this.ab = 0;
                    this.pb = ua2.rb;
                    this.dc = ua2;
                    this.A = ua2.u / 2;
                    this.u = n2;
                    this.wb = ua2.k / 2;
                    this.Z = new ca[this.u];
                    this.zb = 0;
                    this.jb = new int[this.u];
                    this.y = new w[n3];
                    n5 = 0;
                    while (~n3 < ~n5) {
                        this.y[n5] = new w();
                        ++n5;
                        if (!bl) {
                            if (!bl) continue;
                        }
                        break block15;
                    }
                    this.n = 0;
                    this.T = new ca(2 * n4, n4);
                    this.ob = new int[n4];
                    this.Eb = new int[n4];
                }
                if (db.i == null) {
                    db.i = new byte[17691];
                }
                this.Fb = new int[n4];
                this.o = 0;
                this.b = 0;
                this.Ob = new int[n4];
                this.xb = 0;
                this.I = 0;
                this.bc = 0;
                this.Kb = 0;
                this.Q = new int[n4];
                this.gb = new int[n4];
                this.a = new int[n4];
                for (n5 = 0; n5 < 256; ++n5) {
                    e.nb[n5] = (int)(Math.sin(0.02454369 * (double)n5) * 32768.0);
                    e.nb[256 + n5] = (int)(Math.cos((double)n5 * 0.02454369) * 32768.0);
                    if (!bl) {
                        if (!bl) continue;
                    }
                    break block16;
                }
                n5 = 0;
            }
            do {
                if (1024 <= n5) return;
                ba.cc[n5] = (int)(Math.sin((double)n5 * 0.00613592315) * 32768.0);
                ba.cc[1024 + n5] = (int)(Math.cos((double)n5 * 0.00613592315) * 32768.0);
                ++n5;
                if (bl) return;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(N[31]);
            if (ua2 != null) {
                string = N[0];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(')').toString());
            }
            string = N[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(')').toString());
        }
    }

    static {
        N = new String[]{lb.z(lb.z("\u0007e\blW")), lb.z(lb.z("\u0012>J.")), lb.z(lb.z("\u0010)\b\u0004\u0002")), lb.z(lb.z("\u0010)\b\u000ekT")), lb.z(lb.z("\u0010)\b\u0003kT")), lb.z(lb.z("\u0010)\b\n\u0002")), lb.z(lb.z("\u0010)\b\u0003\u0002")), lb.z(lb.z("+*T,C\u0012,\u00066X\u0015.Bb^\u0013kG&N\\%S.F\\$D(O\u001f?\u0007")), lb.z(lb.z("\u0010)\b\fkT")), lb.z(lb.z("\u0010)\b\f\u0002")), lb.z(lb.z("\u0010)\b\u0000\u0002")), lb.z(lb.z("\u0010)\b\u0007\u0002")), lb.z(lb.z("\u0010)\b\u0017\u0002")), lb.z(lb.z("\u0010)\b\rkT")), lb.z(lb.z("\u0010)\b\u0007kT")), lb.z(lb.z("\u0010)\b\u0005\u0002")), lb.z(lb.z("\u0010)\b\bkT")), lb.z(lb.z("\u0010)\b\u0014\u0002")), lb.z(lb.z("\u0010)\b\u000fkT")), lb.z(lb.z("\u0010)\b\u0001\u0002")), lb.z(lb.z("\u0010)\b\u000b\u0002")), lb.z(lb.z("\u0010)\b\u0016\u0002")), lb.z(lb.z("\u0010)\b\u000e\u0002")), lb.z(lb.z("\u0010)\b\u0005kT")), lb.z(lb.z("\u0010)\b\u0012\u0002")), lb.z(lb.z("\u0010)\b\nkT")), lb.z(lb.z("\u0010)\b\u0010\u0002")), lb.z(lb.z("\u0010)\b\r\u0002")), lb.z(lb.z("\u0010)\b\u0011\u0002")), lb.z(lb.z("\u0010)\b\u000bkT")), lb.z(lb.z("\u0010)\b\u0006\u0002")), lb.z(lb.z("\u0010)\b~C\u0012\"R|\u0002")), lb.z(lb.z("\u0010)\b\tkT")), lb.z(lb.z("\u0010)\b\t\u0002")), lb.z(lb.z("\u0010)\b\u000f\u0002")), lb.z(lb.z("\u0010)\b\u0000kT")), lb.z(lb.z("\u0010)\b\b\u0002")), lb.z(lb.z("\u0010)\b\u0001kT")), lb.z(lb.z("\u0010)\b\u0015\u0002")), lb.z(lb.z("\u0010)\b\u0013\u0002")), lb.z(lb.z("\u0010)\b\u0004kT")), lb.z(lb.z("\u0010)\b\u0006kT"))};
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x2A);
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
                        n4 = 124;
                        break;
                    }
                    case 1: {
                        n4 = 75;
                        break;
                    }
                    case 2: {
                        n4 = 38;
                        break;
                    }
                    case 3: {
                        n4 = 66;
                        break;
                    }
                    default: {
                        n4 = 42;
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

