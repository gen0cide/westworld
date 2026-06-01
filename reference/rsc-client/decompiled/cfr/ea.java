/*
 * Decompiled with CFR 0.152.
 */
final class ea {
    private static ac a = new ac();

    private static final void b(ac ac2) {
        int n2 = 0;
        int n3 = 0;
        int n4 = 0;
        int n5 = 0;
        int n6 = 0;
        int n7 = 0;
        int n8 = 0;
        int n9 = 0;
        int n10 = 0;
        int n11 = 0;
        int n12 = 0;
        int n13 = 0;
        int n14 = 0;
        int n15 = 0;
        int n16 = 0;
        int n17 = 0;
        byte by = 0;
        byte by2 = 0;
        int n18 = 0;
        int[] nArray = null;
        int[] nArray2 = null;
        int[] nArray3 = null;
        ac2.f = 1;
        if (ua.Mb == null) {
            ua.Mb = new int[ac2.f * 100000];
        }
        boolean bl = true;
        while (bl) {
            int n19;
            int n20;
            int n21;
            byte by3 = ea.c(ac2);
            if (by3 == 23) {
                return;
            }
            by3 = ea.c(ac2);
            by3 = ea.c(ac2);
            by3 = ea.c(ac2);
            by3 = ea.c(ac2);
            by3 = ea.c(ac2);
            by3 = ea.c(ac2);
            by3 = ea.c(ac2);
            by3 = ea.c(ac2);
            by3 = ea.c(ac2);
            by3 = ea.d(ac2);
            if (by3 != 0) {
                // empty if block
            }
            ac2.E = 0;
            by3 = ea.c(ac2);
            ac2.E = ac2.E << 8 | by3 & 0xFF;
            by3 = ea.c(ac2);
            ac2.E = ac2.E << 8 | by3 & 0xFF;
            by3 = ea.c(ac2);
            ac2.E = ac2.E << 8 | by3 & 0xFF;
            for (n2 = 0; n2 < 16; ++n2) {
                by3 = ea.d(ac2);
                ac2.v[n2] = by3 == 1;
            }
            for (n2 = 0; n2 < 256; ++n2) {
                ac2.n[n2] = false;
            }
            for (n2 = 0; n2 < 16; ++n2) {
                if (!ac2.v[n2]) continue;
                for (n3 = 0; n3 < 16; ++n3) {
                    by3 = ea.d(ac2);
                    if (by3 != 1) continue;
                    ac2.n[n2 * 16 + n3] = true;
                }
            }
            ea.a(ac2);
            n5 = ac2.K + 2;
            n6 = ea.a(3, ac2);
            n7 = ea.a(15, ac2);
            for (n2 = 0; n2 < n7; ++n2) {
                n3 = 0;
                while ((by3 = ea.d(ac2)) != 0) {
                    ++n3;
                }
                ac2.s[n2] = (byte)n3;
            }
            byte[] byArray = new byte[6];
            for (n21 = 0; n21 < n6; n21 = (int)((byte)(n21 + 1))) {
                byArray[n21] = n21;
            }
            for (n2 = 0; n2 < n7; ++n2) {
                n20 = byArray[n21];
                for (n21 = ac2.s[n2]; n21 > 0; n21 = (int)((byte)(n21 - 1))) {
                    byArray[n21] = byArray[n21 - 1];
                }
                byArray[0] = n20;
                ac2.j[n2] = n20;
            }
            for (n4 = 0; n4 < n6; ++n4) {
                n15 = ea.a(5, ac2);
                for (n2 = 0; n2 < n5; ++n2) {
                    while ((by3 = ea.d(ac2)) != 0) {
                        by3 = ea.d(ac2);
                        if (by3 == 0) {
                            ++n15;
                            continue;
                        }
                        --n15;
                    }
                    ac2.B[n4][n2] = (byte)n15;
                }
            }
            for (n4 = 0; n4 < n6; ++n4) {
                int n22 = 32;
                byte by4 = 0;
                for (n2 = 0; n2 < n5; ++n2) {
                    if (ac2.B[n4][n2] > by4) {
                        by4 = ac2.B[n4][n2];
                    }
                    if (ac2.B[n4][n2] >= n22) continue;
                    n22 = ac2.B[n4][n2];
                }
                ea.a(ac2.u[n4], ac2.t[n4], ac2.J[n4], ac2.B[n4], n22, by4, n5);
                ac2.D[n4] = n22;
            }
            n8 = ac2.K + 1;
            n9 = -1;
            n10 = 0;
            for (n2 = 0; n2 <= 255; ++n2) {
                ac2.m[n2] = 0;
            }
            n21 = 4095;
            for (n19 = 15; n19 >= 0; --n19) {
                for (n20 = 15; n20 >= 0; --n20) {
                    ac2.A[n21] = (byte)(n19 * 16 + n20);
                    --n21;
                }
                ac2.r[n19] = n21 + 1;
            }
            n12 = 0;
            if (n10 == 0) {
                n10 = 50;
                by2 = ac2.j[++n9];
                n18 = ac2.D[by2];
                nArray = ac2.u[by2];
                nArray3 = ac2.J[by2];
                nArray2 = ac2.t[by2];
            }
            --n10;
            n16 = n18;
            n17 = ea.a(n16, ac2);
            while (n17 > nArray[n16]) {
                ++n16;
                by = ea.d(ac2);
                n17 = n17 << 1 | by;
            }
            n11 = nArray3[n17 - nArray2[n16]];
            while (n11 != n8) {
                int n23;
                if (n11 == 0 || n11 == 1) {
                    n13 = -1;
                    n14 = 1;
                    do {
                        if (n11 == 0) {
                            n13 += 1 * n14;
                        } else if (n11 == 1) {
                            n13 += 2 * n14;
                        }
                        n14 *= 2;
                        if (n10 == 0) {
                            n10 = 50;
                            by2 = ac2.j[++n9];
                            n18 = ac2.D[by2];
                            nArray = ac2.u[by2];
                            nArray3 = ac2.J[by2];
                            nArray2 = ac2.t[by2];
                        }
                        --n10;
                        n16 = n18;
                        n17 = ea.a(n16, ac2);
                        while (n17 > nArray[n16]) {
                            ++n16;
                            by = ea.d(ac2);
                            n17 = n17 << 1 | by;
                        }
                    } while ((n11 = nArray3[n17 - nArray2[n16]]) == 0 || n11 == 1);
                    by3 = ac2.d[ac2.A[ac2.r[0]] & 0xFF];
                    int n24 = by3 & 0xFF;
                    ac2.m[n24] = ac2.m[n24] + ++n13;
                    while (n13 > 0) {
                        ua.Mb[n12] = by3 & 0xFF;
                        ++n12;
                        --n13;
                    }
                    continue;
                }
                int n25 = n11 - 1;
                if (n25 < 16) {
                    n23 = ac2.r[0];
                    by3 = ac2.A[n23 + n25];
                    while (n25 > 3) {
                        int n26 = n23 + n25;
                        ac2.A[n26] = ac2.A[n26 - 1];
                        ac2.A[n26 - 1] = ac2.A[n26 - 2];
                        ac2.A[n26 - 2] = ac2.A[n26 - 3];
                        ac2.A[n26 - 3] = ac2.A[n26 - 4];
                        n25 -= 4;
                    }
                    while (n25 > 0) {
                        ac2.A[n23 + n25] = ac2.A[n23 + n25 - 1];
                        --n25;
                    }
                    ac2.A[n23] = by3;
                } else {
                    int n27 = n25 / 16;
                    int n28 = n25 % 16;
                    by3 = ac2.A[n23];
                    for (n23 = ac2.r[n27] + n28; n23 > ac2.r[n27]; --n23) {
                        ac2.A[n23] = ac2.A[n23 - 1];
                    }
                    int n29 = n27;
                    ac2.r[n29] = ac2.r[n29] + 1;
                    while (n27 > 0) {
                        int n30 = n27;
                        ac2.r[n30] = ac2.r[n30] - 1;
                        ac2.A[ac2.r[n27]] = ac2.A[ac2.r[n27 - 1] + 16 - 1];
                        --n27;
                    }
                    ac2.r[0] = ac2.r[0] - 1;
                    ac2.A[ac2.r[0]] = by3;
                    if (ac2.r[0] == 0) {
                        n21 = 4095;
                        for (n19 = 15; n19 >= 0; --n19) {
                            for (n20 = 15; n20 >= 0; --n20) {
                                ac2.A[n21] = ac2.A[ac2.r[n19] + n20];
                                --n21;
                            }
                            ac2.r[n19] = n21 + 1;
                        }
                    }
                }
                int n31 = ac2.d[by3 & 0xFF] & 0xFF;
                ac2.m[n31] = ac2.m[n31] + 1;
                ua.Mb[n12] = ac2.d[by3 & 0xFF] & 0xFF;
                ++n12;
                if (n10 == 0) {
                    n10 = 50;
                    by2 = ac2.j[++n9];
                    n18 = ac2.D[by2];
                    nArray = ac2.u[by2];
                    nArray3 = ac2.J[by2];
                    nArray2 = ac2.t[by2];
                }
                --n10;
                n16 = n18;
                n17 = ea.a(n16, ac2);
                while (n17 > nArray[n16]) {
                    ++n16;
                    by = ea.d(ac2);
                    n17 = n17 << 1 | by;
                }
                n11 = nArray3[n17 - nArray2[n16]];
            }
            ac2.F = 0;
            ac2.g = 0;
            ac2.w[0] = 0;
            for (n2 = 1; n2 <= 256; ++n2) {
                ac2.w[n2] = ac2.m[n2 - 1];
            }
            for (n2 = 1; n2 <= 256; ++n2) {
                int n32 = n2;
                ac2.w[n32] = ac2.w[n32] + ac2.w[n2 - 1];
            }
            for (n2 = 0; n2 < n12; ++n2) {
                by3 = (byte)(ua.Mb[n2] & 0xFF);
                int n33 = ac2.w[by3 & 0xFF];
                ua.Mb[n33] = ua.Mb[n33] | n2 << 8;
                int n34 = by3 & 0xFF;
                ac2.w[n34] = ac2.w[n34] + 1;
            }
            ac2.H = ua.Mb[ac2.E] >> 8;
            ac2.L = 0;
            ac2.H = ua.Mb[ac2.H];
            ac2.h = (byte)(ac2.H & 0xFF);
            ac2.H >>= 8;
            ++ac2.L;
            ac2.b = n12;
            ea.e(ac2);
            if (ac2.L == ac2.b + 1 && ac2.F == 0) {
                bl = true;
                continue;
            }
            bl = false;
        }
    }

    private static final byte c(ac ac2) {
        return (byte)ea.a(8, ac2);
    }

    private static final byte d(ac ac2) {
        return (byte)ea.a(1, ac2);
    }

    private static final void a(ac ac2) {
        ac2.K = 0;
        for (int i2 = 0; i2 < 256; ++i2) {
            if (!ac2.n[i2]) continue;
            ac2.d[ac2.K] = (byte)i2;
            ++ac2.K;
        }
    }

    private static final void a(int[] nArray, int[] nArray2, int[] nArray3, byte[] byArray, int n2, int n3, int n4) {
        int n5;
        int n6 = 0;
        for (n5 = n2; n5 <= n3; ++n5) {
            for (int i2 = 0; i2 < n4; ++i2) {
                if (byArray[i2] != n5) continue;
                nArray3[n6] = i2;
                ++n6;
            }
        }
        for (n5 = 0; n5 < 23; ++n5) {
            nArray2[n5] = 0;
        }
        for (n5 = 0; n5 < n4; ++n5) {
            int n7 = byArray[n5] + 1;
            nArray2[n7] = nArray2[n7] + 1;
        }
        for (n5 = 1; n5 < 23; ++n5) {
            int n8 = n5;
            nArray2[n8] = nArray2[n8] + nArray2[n5 - 1];
        }
        for (n5 = 0; n5 < 23; ++n5) {
            nArray[n5] = 0;
        }
        int n9 = 0;
        for (n5 = n2; n5 <= n3; ++n5) {
            nArray[n5] = (n9 += nArray2[n5 + 1] - nArray2[n5]) - 1;
            n9 <<= 1;
        }
        for (n5 = n2 + 1; n5 <= n3; ++n5) {
            nArray2[n5] = (nArray[n5 - 1] + 1 << 1) - nArray2[n5];
        }
    }

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     */
    static final int a(byte[] byArray, int n2, byte[] byArray2, int n3, int n4) {
        ac ac2 = a;
        synchronized (ac2) {
            ea.a.q = byArray2;
            ea.a.a = n4;
            ea.a.i = byArray;
            ea.a.o = 0;
            ea.a.y = n2;
            ea.a.p = 0;
            ea.a.e = 0;
            ea.a.c = 0;
            ea.a.G = 0;
            ea.b(a);
            ea.a.q = null;
            ea.a.i = null;
            return n2 -= ea.a.y;
        }
    }

    /*
     * Unable to fully structure code
     */
    private static final void e(ac var0) {
        var2_1 = var0.g;
        var3_2 = var0.F;
        var4_3 = var0.L;
        var5_4 = var0.h;
        var6_5 = ua.Mb;
        var7_6 = var0.H;
        var8_7 = var0.i;
        var9_8 = var0.o;
        var11_10 = var10_9 = var0.y;
        var12_11 = var0.b + 1;
        block0: while (true) {
            if (var3_2 <= 0) ** GOTO lbl26
            while (var10_9 != 0) {
                if (var3_2 != 1) {
                    var8_7[var9_8] = var2_1;
                    --var3_2;
                    ++var9_8;
                    --var10_9;
                    continue;
                }
                if (var10_9 == 0) {
                    var3_2 = 1;
                    break block0;
                }
                var8_7[var9_8] = var2_1;
                ++var9_8;
                --var10_9;
lbl26:
                // 2 sources

                while (true) {
                    if (var4_3 == var12_11) {
                        var3_2 = 0;
                        break block0;
                    }
                    var2_1 = (byte)var5_4;
                    var7_6 = var6_5[var7_6];
                    var1_12 = (byte)var7_6;
                    var7_6 >>= 8;
                    ++var4_3;
                    if (var1_12 != var5_4) {
                        var5_4 = var1_12;
                        if (var10_9 == 0) {
                            var3_2 = 1;
                            break block0;
                        }
                        var8_7[var9_8] = var2_1;
                        ++var9_8;
                        --var10_9;
                        continue;
                    }
                    if (var4_3 != var12_11) break;
                    if (var10_9 == 0) {
                        var3_2 = 1;
                        break block0;
                    }
                    var8_7[var9_8] = var2_1;
                    ++var9_8;
                    --var10_9;
                }
                var3_2 = 2;
                var7_6 = var6_5[var7_6];
                var1_12 = (byte)var7_6;
                var7_6 >>= 8;
                if (++var4_3 == var12_11) continue block0;
                if (var1_12 != var5_4) {
                    var5_4 = var1_12;
                    continue block0;
                }
                var3_2 = 3;
                var7_6 = var6_5[var7_6];
                var1_12 = (byte)var7_6;
                var7_6 >>= 8;
                if (++var4_3 == var12_11) continue block0;
                if (var1_12 != var5_4) {
                    var5_4 = var1_12;
                    continue block0;
                }
                var7_6 = var6_5[var7_6];
                var1_12 = (byte)var7_6;
                var7_6 >>= 8;
                ++var4_3;
                var3_2 = (var1_12 & 255) + 4;
                var7_6 = var6_5[var7_6];
                var5_4 = (byte)var7_6;
                var7_6 >>= 8;
                ++var4_3;
                continue block0;
            }
            break;
        }
        var13_13 = var0.G;
        var0.G += var11_10 - var10_9;
        if (var0.G < var13_13) {
            // empty if block
        }
        var0.g = var2_1;
        var0.F = var3_2;
        var0.L = var4_3;
        var0.h = var5_4;
        ua.Mb = var6_5;
        var0.H = var7_6;
        var0.i = var8_7;
        var0.o = var9_8;
        var0.y = var10_9;
    }

    private static final int a(int n2, ac ac2) {
        while (true) {
            if (ac2.p >= n2) {
                int n3 = ac2.e >> ac2.p - n2 & (1 << n2) - 1;
                ac2.p -= n2;
                return n3;
            }
            ac2.e = ac2.e << 8 | ac2.q[ac2.a] & 0xFF;
            ac2.p += 8;
            ++ac2.a;
            ++ac2.c;
            if (ac2.c != 0) continue;
        }
    }
}

