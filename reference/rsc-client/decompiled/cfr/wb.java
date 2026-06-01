/*
 * Decompiled with CFR 0.152.
 */
final class wb {
    static int B;
    static int C;
    static int z;
    static int H;
    static int d;
    static int a;
    private int I;
    static int o;
    private ba t;
    static int K;
    static long w;
    private int n;
    static int l;
    static int u;
    static int x;
    static int c;
    static aa p;
    static int k;
    static int F;
    private int D;
    static int f;
    static int j;
    static int v;
    static int y;
    static int M;
    static int s;
    private int i;
    private String m;
    static int E;
    static int r;
    static int e;
    static int J;
    static int L;
    private t[] b;
    static int[] q;
    static int G;
    static int h;
    static int g;
    private static final String[] A;

    final int a(int n2) {
        try {
            ++j;
            if (n2 != -21224) {
                this.b(false, 0);
            }
            return this.I;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[5] + n2 + ')');
        }
    }

    final int a(boolean bl, int n2) {
        try {
            if (!bl) {
                this.b((byte)30, 75);
            }
            ++h;
            return this.b[n2].e;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[24] + bl + ',' + n2 + ')');
        }
    }

    final String b(byte by, int n2) {
        try {
            ++E;
            if (by <= 13) {
                return null;
            }
            return this.b[n2].o;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[25] + by + ',' + n2 + ')');
        }
    }

    final void b(int n2, int n3) {
        block11: {
            boolean bl = client.vh;
            try {
                block10: {
                    block9: {
                        ++F;
                        if (-1 >= ~n3) {
                            if (~n3 <= ~this.n) break block9;
                            break block10;
                        }
                    }
                    return;
                }
                t t2 = this.b[n3];
                int n4 = n3;
                int n5 = -66 % ((n2 - 37) / 52);
                while (~(this.n + -1) < ~n4) {
                    this.b[n4] = this.b[1 + n4];
                    ++n4;
                    if (!bl) {
                        if (!bl) continue;
                        break;
                    }
                    break block11;
                }
                this.b[--this.n] = t2;
                this.a(true);
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, A[17] + n2 + ',' + n3 + ')');
            }
        }
    }

    final int c(int n2) {
        try {
            ++f;
            if (n2 != -27153) {
                this.a(false);
            }
            return this.n;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[21] + n2 + ')');
        }
    }

    final int b(boolean bl, int n2) {
        try {
            ++l;
            if (!bl) {
                this.b(-33, (byte)91);
            }
            return this.b[n2].l;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[19] + bl + ',' + n2 + ')');
        }
    }

    final void a(String string, String string2, String string3, int n2, String string4, byte by) {
        try {
            int n3 = -26 % ((by - 15) / 33);
            this.a(0, string, 0, 0, string2, 0, null, n2, 0, 125, string3, string4);
            ++B;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(A[2]);
            String string5 = string != null ? A[0] : A[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string5).append(',');
            String string6 = string2 != null ? A[0] : A[1];
            StringBuilder stringBuilder3 = stringBuilder2.append(string6).append(',');
            String string7 = string3 != null ? A[0] : A[1];
            StringBuilder stringBuilder4 = stringBuilder3.append(string7).append(',').append(n2).append(',');
            String string8 = string4 != null ? A[0] : A[1];
            throw i.a(runtimeException2, stringBuilder4.append(string8).append(',').append(by).append(')').toString());
        }
    }

    final void a(int n2, int n3, boolean bl, String string, String string2) {
        try {
            this.a(0, string, 0, 0, string2, n2, null, n3, 0, 125, null, null);
            ++u;
            if (bl) {
                this.a(61);
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(A[3]).append(n2).append(',').append(n3).append(',').append(bl).append(',');
            String string3 = string != null ? A[0] : A[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string3).append(',');
            String string4 = string2 != null ? A[0] : A[1];
            throw i.a(runtimeException2, stringBuilder2.append(string4).append(')').toString());
        }
    }

    final String c(int n2, int n3) {
        try {
            ++k;
            if (n3 != -4126) {
                return null;
            }
            return this.b[n2].b;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[23] + n2 + ',' + n3 + ')');
        }
    }

    final int a(int n2, int n3) {
        try {
            ++C;
            if (n2 >= -14) {
                return -114;
            }
            return this.b[n3].d;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[26] + n2 + ',' + n3 + ')');
        }
    }

    final int a(int n2, int n3, int n4, byte by, int n5) {
        try {
            if (by != -12) {
                this.i = -77;
            }
            ++H;
            return this.a(n4, n5, n2, n3, -66, true);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[8] + n2 + ',' + n3 + ',' + n4 + ',' + by + ',' + n5 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final int a(int n2, int n3, int n4, int n5, int n6, boolean bl) {
        boolean bl2 = client.vh;
        try {
            int n7;
            ++a;
            if (-1 == ~this.D || this.I == 0) return -1;
            if (!bl) {
            } else {
                this.t.c(160, n5, this.I, 0, n4, this.D, 0xD0D0D0);
            }
            int n10 = 1 + this.t.a(508305352, this.i);
            int n9 = -3 + (n10 + n4);
            int n8 = -1;
            if (null != this.m) {
                if (n5 < n3 && n2 > n9 + (3 + -n10) && ~n2 > ~(n9 - -3) && n3 < n5 + this.D) {
                    if (!bl) {
                        return -2;
                    }
                    n8 = -2;
                }
                if (bl) {
                    this.t.a(this.m, 2 + n5, n9, 65535, false, this.i);
                }
                n9 += n10;
            }
            if (n6 >= -1) {
                this.m = null;
            }
            for (int i2 = 0; i2 < this.n; n9 += n10, ++i2) {
                int n11 = 0xFFFFFF;
                n7 = ~n3;
                if (bl2) return n7;
                if (n7 < ~n5 && ~(-n10 + (3 + n9)) > ~n2 && ~(3 + n9) < ~n2 && ~n3 > ~(n5 + this.D)) {
                    n11 = 0xFFFF00;
                    if (!bl) {
                        return i2;
                    }
                    n8 = i2;
                }
                if (!bl) continue;
                this.t.a(this.b[i2].p + " " + this.b[i2].o, n5 - -2, n9, n11, false, this.i);
                if (!bl2) continue;
            }
            n7 = n8;
            return n7;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[29] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + bl + ')');
        }
    }

    /*
     * Enabled force condition propagation
     * Lifted jumps to return sites
     */
    final void a(byte by) {
        boolean bl = client.vh;
        try {
            int n2;
            Object[] objectArray;
            block13: {
                ++o;
                if (this.n == 0) {
                    return;
                }
                int[] nArray = new int[this.n];
                objectArray = new Object[this.n];
                n2 = 0;
                while (~n2 > ~this.n) {
                    t t2 = this.b[n2];
                    nArray[n2] = t2.d;
                    objectArray[n2] = t2;
                    ++n2;
                    if (!bl) {
                        if (!bl) continue;
                        break;
                    }
                    break block13;
                }
                cb.a((byte)-70, objectArray, nArray);
                n2 = 0;
            }
            if (by != 16) {
                this.I = -103;
            }
            while (~this.n < ~n2) {
                try {
                    this.b[n2] = (t)objectArray[n2];
                    ++n2;
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
            throw i.a(runtimeException, A[22] + by + ')');
        }
    }

    final int a(int n2, byte by) {
        try {
            ++d;
            if (by != 22) {
                this.D = 4;
            }
            return this.b[n2].j;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[16] + n2 + ',' + by + ')');
        }
    }

    final void a(int n2, String string, String string2, int n3) {
        try {
            if (n3 != 30192) {
                this.a(true, 125);
            }
            ++e;
            this.a(0, string2, 0, 0, string, 0, null, n2, 0, n3 ^ 0x758D, null, null);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(A[28]).append(n2).append(',');
            String string3 = string != null ? A[0] : A[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string3).append(',');
            String string4 = string2 != null ? A[0] : A[1];
            throw i.a(runtimeException2, stringBuilder2.append(string4).append(',').append(n3).append(')').toString());
        }
    }

    final String b(int n2, byte by) {
        try {
            if (by != 53) {
                return null;
            }
            ++J;
            return this.b[n2].p;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[13] + n2 + ',' + by + ')');
        }
    }

    wb(ba ba2, int n2) {
        this(ba2, n2, null);
    }

    final void a(int n2, int n3, int n4, int n5, int n6, int n7, String string, String string2) {
        try {
            if (n6 <= 44) {
                return;
            }
            ++G;
            this.a(n3, string2, n4, 0, string, n5, null, n2, n7, 125, null, null);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(A[10]).append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',');
            String string3 = string != null ? A[0] : A[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string3).append(',');
            String string4 = string2 != null ? A[0] : A[1];
            throw i.a(runtimeException2, stringBuilder2.append(string4).append(')').toString());
        }
    }

    final int a(byte by, int n2) {
        try {
            ++y;
            if (by != 97) {
                return 2;
            }
            return this.b[n2].m;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[9] + by + ',' + n2 + ')');
        }
    }

    final void a(int n2, String string, int n3, int n4, int n5, int n6, int n7, String string2, int n8) {
        try {
            ++r;
            this.a(n2, string, n7, n4, string2, n8, null, n6, n5, 126, null, null);
            int n9 = -66 / ((n3 - -42) / 41);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(A[4]).append(n2).append(',');
            String string3 = string != null ? A[0] : A[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string3).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',');
            String string4 = string2 != null ? A[0] : A[1];
            throw i.a(runtimeException2, stringBuilder2.append(string4).append(',').append(n8).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(boolean bl) {
        boolean bl2 = client.vh;
        try {
            int n2;
            block14: {
                block13: {
                    ++c;
                    n2 = this.t.a(508305352, this.i) - -1;
                    if (null != this.m) break block13;
                    this.I = 0;
                    this.D = 0;
                    if (!bl2) break block14;
                }
                this.I = n2;
                this.D = 5 + this.t.a(this.i, 76, this.m);
            }
            if (!bl) {
                this.b(true, 124);
            }
            int n3 = 0;
            do {
                if (this.n <= n3) return;
                this.I += n2;
                int n4 = 5 + this.t.a(this.i, 105, this.b[n3].p + " " + this.b[n3].o);
                if (bl2) return;
                if (n4 > this.D) {
                    this.D = n4;
                }
                ++n3;
            } while (!bl2);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[27] + bl + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final void a(int n2, int n3, int n4, int n5, int[] nArray, int n6, int n7, int n8, int n9, int n10, int n11, int n12, int n13, int n14, int[] nArray2, int n15) {
        boolean bl = client.vh;
        try {
            int n16;
            int n17;
            block55: {
                block54: {
                    ++g;
                    if (n15 <= 0) return;
                    if (n3 != 10) {
                        wb.a(-30, -28, 22, 0, null, -78, 109, 44, -120, 67, 27, 2, 107, -113, null, 56);
                    }
                    n17 = 0;
                    n16 = 0;
                    n12 <<= 2;
                    if (~n6 != -1) {
                        n16 = n9 / n6 << 1419282055;
                        n17 = n8 / n6 << 1665286983;
                    }
                    if (n17 >= 0) break block54;
                    n17 = 0;
                    if (!bl) break block55;
                }
                if (16256 < n17) {
                    n17 = 16256;
                }
            }
            int n18 = n15;
            do {
                block58: {
                    int n19;
                    int n20;
                    int n21;
                    block53: {
                        if (n18 <= 0) return;
                        n8 += n14;
                        n5 = n16;
                        n6 += n11;
                        n4 = n17;
                        int n22 = n9;
                        int n23 = n2;
                        block20: while (true) {
                            block57: {
                                block56: {
                                    n9 = n22 + n23;
                                    if (bl) return;
                                    if (n6 != 0) {
                                        n17 = n8 / n6 << -47707001;
                                        n16 = n9 / n6 << 856475559;
                                    }
                                    if (n17 < 0) break block56;
                                    if (-16257 <= ~n17) break block57;
                                    n17 = 16256;
                                    if (!bl) break block57;
                                }
                                n17 = 0;
                            }
                            n21 = -n4 + n17 >> -716279868;
                            n20 = -n5 + n16 >> 310733700;
                            n19 = n7 >> -128834185;
                            n4 += 0x600000 & n7;
                            n7 += n12;
                            if (n18 >= 16) break block53;
                            for (int i2 = 0; n18 > i2; ++i2) {
                                n22 = -1;
                                n13 = nArray2[(n5 & 0x3F80) + (n4 >> -557206009)] >>> n19;
                                n23 = ~n13;
                                if (bl) continue block20;
                                if (n22 != n23) {
                                    nArray[n10] = n13;
                                }
                                ++n10;
                                n4 += n21;
                                n5 += n20;
                                if (-4 != ~(i2 & 3)) continue;
                                n4 = (n7 & 0x600000) + (0x3FFF & n4);
                                n19 = n7 >> 108767799;
                                n7 += n12;
                                if (!bl) continue;
                            }
                            break;
                        }
                        if (!bl) break block58;
                    }
                    if (-1 != ~(n13 = nArray2[(n5 & 0x3F80) + (n4 >> -1762053913)] >>> n19)) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                    n13 = nArray2[((n4 += n21) >> 142914567) + (0x3F80 & (n5 += n20))] >>> n19;
                    if (n13 != 0) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                    n13 = nArray2[((n4 += n21) >> 702349191) + (0x3F80 & (n5 += n20))] >>> n19;
                    if (n13 != 0) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                    n13 = nArray2[((n4 += n21) >> 84789639) + ((n5 += n20) & 0x3F80)] >>> n19;
                    if (0 != n13) {
                        nArray[n10] = n13;
                    }
                    n4 += n21;
                    ++n10;
                    n19 = n7 >> 84224663;
                    n4 = (n7 & 0x600000) + (0x3FFF & n4);
                    n7 += n12;
                    n13 = nArray2[((n5 += n20) & 0x3F80) + (n4 >> -1461444889)] >>> n19;
                    if (-1 != ~n13) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                    n13 = nArray2[(0x3F80 & (n5 += n20)) - -((n4 += n21) >> -497158425)] >>> n19;
                    if (~n13 != -1) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                    n13 = nArray2[((n4 += n21) >> -319394553) + (0x3F80 & (n5 += n20))] >>> n19;
                    if (n13 != 0) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                    n13 = nArray2[((n4 += n21) >> 1667252231) + ((n5 += n20) & 0x3F80)] >>> n19;
                    if (-1 != ~n13) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                    n4 += n21;
                    n19 = n7 >> -348227017;
                    n4 = (n4 & 0x3FFF) - -(0x600000 & n7);
                    n13 = nArray2[(n4 >> 2085548263) + ((n5 += n20) & 0x3F80)] >>> n19;
                    if (~n13 != -1) {
                        nArray[n10] = n13;
                    }
                    n7 += n12;
                    ++n10;
                    n13 = nArray2[((n4 += n21) >> -1433259769) + ((n5 += n20) & 0x3F80)] >>> n19;
                    if (-1 != ~n13) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                    n13 = nArray2[(0x3F80 & (n5 += n20)) - -((n4 += n21) >> -728567193)] >>> n19;
                    if (n13 != 0) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                    n13 = nArray2[((n4 += n21) >> -532021593) + ((n5 += n20) & 0x3F80)] >>> n19;
                    if (n13 != 0) {
                        nArray[n10] = n13;
                    }
                    n4 += n21;
                    ++n10;
                    n19 = n7 >> 96919607;
                    n13 = nArray2[((n5 += n20) & 0x3F80) + ((n4 = (n4 & 0x3FFF) - -(n7 & 0x600000)) >> 834075207)] >>> n19;
                    if (~n13 != -1) {
                        nArray[n10] = n13;
                    }
                    n7 += n12;
                    ++n10;
                    n13 = nArray2[((n4 += n21) >> 2085930247) + ((n5 += n20) & 0x3F80)] >>> n19;
                    if (-1 != ~n13) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                    n13 = nArray2[((n4 += n21) >> -1394491673) + ((n5 += n20) & 0x3F80)] >>> n19;
                    if (n13 != 0) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                    n13 = nArray2[(0x3F80 & (n5 += n20)) - -((n4 += n21) >> -1327954841)] >>> n19;
                    if (-1 != ~n13) {
                        nArray[n10] = n13;
                    }
                    ++n10;
                }
                n18 -= 16;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(A[12]).append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(nArray != null ? A[0] : A[1]).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(',');
            if (nArray2 != null) {
                string = A[0];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n15).append(')').toString());
            }
            string = A[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n15).append(')').toString());
        }
    }

    final void a(int n2, String string, int n3, String string2, int n4, int n5) {
        try {
            if (n5 != 3296) {
                w = -93L;
            }
            ++v;
            this.a(n4, string2, 0, 0, string, n2, null, n3, 0, n5 + -3170, null, null);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(A[15]).append(n2).append(',');
            String string3 = string != null ? A[0] : A[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string3).append(',').append(n3).append(',');
            String string4 = string2 != null ? A[0] : A[1];
            throw i.a(runtimeException2, stringBuilder2.append(string4).append(',').append(n4).append(',').append(n5).append(')').toString());
        }
    }

    final void d(int n2) {
        try {
            ++K;
            this.n = n2;
            this.a(true);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[11] + n2 + ')');
        }
    }

    final int b(int n2, int n3, int n4, byte by, int n5) {
        try {
            if (by != -40) {
                this.a((byte)-62);
            }
            ++M;
            return this.a(n5, n2, n4, n3, -3, false);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[18] + n2 + ',' + n3 + ',' + n4 + ',' + by + ',' + n5 + ')');
        }
    }

    final int a(int n2, boolean bl) {
        try {
            ++x;
            if (bl) {
                this.i = 119;
            }
            return this.b[n2].i;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[20] + n2 + ',' + bl + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2, String string, int n3, int n4, String string2, int n5, String string3, int n6, int n7, int n8, String string4, String string5) {
        boolean bl = client.vh;
        try {
            int n9;
            int n10;
            if (~this.n == ~this.b.length) {
                t[] tArray = this.b;
                this.b = new t[10 + this.n];
                for (int i2 = 0; this.b.length > i2; ++i2) {
                    n10 = this.n;
                    n9 = i2;
                    if (!bl) {
                        if (n10 > n9) {
                            this.b[i2] = tArray[i2];
                            if (!bl) continue;
                        }
                        this.b[i2] = new t();
                        if (!bl) continue;
                    }
                    break;
                }
            } else {
                n10 = n8;
                n9 = 124;
            }
            if (n10 <= n9) {
                return;
            }
            ++L;
            this.b[this.n++].a(string, n4, n5, n2, n7, string3, 100, n6, string2, string4, n3, string5);
            this.a(true);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string6;
            StringBuilder stringBuilder = new StringBuilder().append(A[7]).append(n2).append(',').append(string != null ? A[0] : A[1]).append(',').append(n3).append(',').append(n4).append(',').append(string2 != null ? A[0] : A[1]).append(',').append(n5).append(',').append(string3 != null ? A[0] : A[1]).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(string4 != null ? A[0] : A[1]).append(',');
            if (string5 != null) {
                string6 = A[0];
                throw i.a(runtimeException, stringBuilder.append(string6).append(')').toString());
            }
            string6 = A[1];
            throw i.a(runtimeException, stringBuilder.append(string6).append(')').toString());
        }
    }

    final int b(int n2) {
        try {
            if (n2 != 16256) {
                this.a((byte)-39);
            }
            ++s;
            return this.D;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, A[6] + n2 + ')');
        }
    }

    final void a(int n2, byte by, int n3, String string, String string2, int n4, int n5) {
        try {
            ++z;
            if (by != 22) {
                this.a(33, false);
            }
            this.a(n5, string, n4, 0, string2, n2, null, n3, 0, 127, null, null);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(A[30]).append(n2).append(',').append(by).append(',').append(n3).append(',');
            String string3 = string != null ? A[0] : A[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string3).append(',');
            String string4 = string2 != null ? A[0] : A[1];
            throw i.a(runtimeException2, stringBuilder2.append(string4).append(',').append(n4).append(',').append(n5).append(')').toString());
        }
    }

    wb(ba ba2, int n2, String string) {
        block10: {
            boolean bl = client.vh;
            this.I = 0;
            this.n = 0;
            this.D = 0;
            try {
                this.i = n2;
                this.b = new t[10];
                this.t = ba2;
                this.m = string;
                int n3 = 0;
                while (~n3 > -11) {
                    this.b[n3] = new t();
                    ++n3;
                    if (!bl) {
                        if (!bl) continue;
                        break;
                    }
                    break block10;
                }
                this.a(true);
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(A[14]);
                String string2 = ba2 != null ? A[0] : A[1];
                StringBuilder stringBuilder2 = stringBuilder.append(string2).append(',').append(n2).append(',');
                String string3 = string != null ? A[0] : A[1];
                throw i.a(runtimeException2, stringBuilder2.append(string3).append(')').toString());
            }
        }
    }

    static {
        A = new String[]{wb.z(wb.z("\u0013l\u000b\u0016:")), wb.z(wb.z("\u00067IT")), wb.z(wb.z("\u001f \u000b}o")), wb.z(wb.z("\u001f \u000b|\u0006@")), wb.z(wb.z("\u001f \u000bqo")), wb.z(wb.z("\u001f \u000blo")), wb.z(wb.z("\u001f \u000bz\u0006@")), wb.z(wb.z("\u001f \u000bvo")), wb.z(wb.z("\u001f \u000bmo")), wb.z(wb.z("\u001f \u000bso")), wb.z(wb.z("\u001f \u000bzo")), wb.z(wb.z("\u001f \u000bho")), wb.z(wb.z("\u001f \u000bko")), wb.z(wb.z("\u001f \u000byo")), wb.z(wb.z("\u001f \u000b\u0004.\u0006+Q\u0006o")), wb.z(wb.z("\u001f \u000bro")), wb.z(wb.z("\u001f \u000bpo")), wb.z(wb.z("\u001f \u000b\u007fo")), wb.z(wb.z("\u001f \u000b|o")), wb.z(wb.z("\u001f \u000bio")), wb.z(wb.z("\u001f \u000bto")), wb.z(wb.z("\u001f \u000b~o")), wb.z(wb.z("\u001f \u000by\u0006@")), wb.z(wb.z("\u001f \u000b{\u0006@")), wb.z(wb.z("\u001f \u000b{o")), wb.z(wb.z("\u001f \u000bwo")), wb.z(wb.z("\u001f \u000buo")), wb.z(wb.z("\u001f \u000b}\u0006@")), wb.z(wb.z("\u001f \u000bno")), wb.z(wb.z("\u001f \u000bjo")), wb.z(wb.z("\u001f \u000boo"))};
        p = new aa(new byte[]{22, 22, 22, 22, 22, 22, 21, 22, 22, 20, 22, 22, 22, 21, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 3, 8, 22, 16, 22, 16, 17, 7, 13, 13, 13, 16, 7, 10, 6, 16, 10, 11, 12, 12, 12, 12, 13, 13, 14, 14, 11, 14, 19, 15, 17, 8, 11, 9, 10, 10, 10, 10, 11, 10, 9, 7, 12, 11, 10, 10, 9, 10, 10, 12, 10, 9, 8, 12, 12, 9, 14, 8, 12, 17, 16, 17, 22, 13, 21, 4, 7, 6, 5, 3, 6, 6, 5, 4, 10, 7, 5, 6, 4, 4, 6, 10, 5, 4, 4, 5, 7, 6, 10, 6, 10, 22, 19, 22, 14, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 21, 22, 21, 22, 22, 22, 21, 22, 22});
        q = new int[256];
        int n2 = 0;
        while (~n2 > -257) {
            int n3 = n2;
            int n4 = 0;
            while (true) {
                block5: {
                    block4: {
                        if (n4 >= 8) break;
                        if (~(1 & n3) != -2) break block4;
                        n3 = n3 >>> -1257587711 ^ 0xEDB88320;
                        break block5;
                    }
                    n3 >>>= 1;
                }
                ++n4;
            }
            wb.q[n2] = n3;
            ++n2;
        }
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x47);
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
                        n4 = 104;
                        break;
                    }
                    case 1: {
                        n4 = 66;
                        break;
                    }
                    case 2: {
                        n4 = 37;
                        break;
                    }
                    case 3: {
                        n4 = 56;
                        break;
                    }
                    default: {
                        n4 = 71;
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

