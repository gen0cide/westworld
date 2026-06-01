/*
 * Decompiled with CFR 0.152.
 */
import java.io.IOException;
import java.net.URL;

final class cb {
    static int a;
    static int g;
    static String[] c;
    static int f;
    static int b;
    static String[] e;
    static int d;
    private static final String[] z;

    static final void a(URL uRL, e e2, int n2) throws IOException {
        boolean bl = client.vh;
        try {
            int n3;
            tb tb2;
            block18: {
                ++f;
                d.h = e2;
                ib.c = uRL;
                URL uRL2 = new URL(ib.c, z[3] + Long.toHexString(p.a(0)));
                o.l = z[5];
                byte[] byArray = da.a(uRL2, true, true);
                tb2 = new tb(byArray);
                n3 = 0;
                while (-13 < ~n3) {
                    tb.l[n3] = tb2.b(-129);
                    ++n3;
                    if (!bl) {
                        if (!bl) continue;
                        break;
                    }
                    break block18;
                }
                tb2.b(-129);
            }
            if (!tb2.e(-422797528)) {
                throw new IOException(z[6]);
            }
            try {
                n3 = 81 % ((0 - n2) / 54);
                if (pa.k.f != null) {
                    s.a = new nb(pa.k.f, 5200, 0);
                    n.h = new nb(pa.k.v, 6000, 0);
                    m.e = new ob(0, s.a, n.h, 1000000);
                    pa.k.f = null;
                    pa.k.v = null;
                }
            }
            catch (IOException iOException) {
                s.a = null;
                n.h = null;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[4]);
            String string = uRL != null ? z[0] : z[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',');
            String string2 = e2 != null ? z[0] : z[1];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(',').append(n2).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final void a(int n2, int n3, int n4, byte by, int n5, int n6, int n7, int n8, int[] nArray, int[] nArray2, int n9, int n10, int n11, int n12, int n13, int n14) {
        boolean bl = client.vh;
        try {
            int n15;
            int n16;
            block51: {
                block50: {
                    ++a;
                    if (n2 <= 0) {
                        return;
                    }
                    n16 = 0;
                    n15 = 0;
                    if (0 != n3) {
                        n16 = n10 / n3 << -1240379354;
                        n15 = n14 / n3 << 1095567590;
                    }
                    n8 <<= 2;
                    if (0 > n16) break block50;
                    if (n16 <= 4032) break block51;
                    n16 = 4032;
                    if (!bl) break block51;
                }
                n16 = 0;
            }
            if (by != 25) {
                return;
            }
            int n17 = n2;
            do {
                int n18 = ~n17;
                int n19 = -1;
                block18: while (true) {
                    block53: {
                        block52: {
                            if (n18 >= n19) return;
                            n5 = n15;
                            n11 = n16;
                            n10 += n6;
                            n14 += n12;
                            if (bl) return;
                            if ((n3 += n7) != 0) {
                                n15 = n14 / n3 << -1050475738;
                                n16 = n10 / n3 << -700638746;
                            }
                            if (~n16 <= -1) break block52;
                            n16 = 0;
                            if (!bl) break block53;
                        }
                        if (~n16 < -4033) {
                            n16 = 4032;
                        }
                    }
                    int n20 = n15 + -n5 >> -490767996;
                    int n21 = -n11 + n16 >> 461459556;
                    n11 += 0xC0000 & n13;
                    int n22 = n13 >> 1111279860;
                    n13 += n8;
                    if (-17 >= ~n17) {
                        n4 = nArray[(n11 >> 1781768774) + (0xFC0 & n5)] >>> n22;
                        if (0 != n4) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n4 = nArray[((n11 += n21) >> 1983493318) + ((n5 += n20) & 0xFC0)] >>> n22;
                        if (-1 != ~n4) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n4 = nArray[((n11 += n21) >> 1627062566) + (0xFC0 & (n5 += n20))] >>> n22;
                        if (0 != n4) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n4 = nArray[(0xFC0 & (n5 += n20)) + ((n11 += n21) >> 387291942)] >>> n22;
                        if (0 != n4) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n11 += n21;
                        n22 = n13 >> -1634170220;
                        n11 = (0xC0000 & n13) + (0xFFF & n11);
                        n13 += n8;
                        n4 = nArray[(n11 >> -291291162) + (0xFC0 & (n5 += n20))] >>> n22;
                        if (-1 != ~n4) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n4 = nArray[((n5 += n20) & 0xFC0) + ((n11 += n21) >> 1451268166)] >>> n22;
                        if (n4 != 0) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n4 = nArray[((n5 += n20) & 0xFC0) + ((n11 += n21) >> 258323942)] >>> n22;
                        if (-1 != ~n4) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n4 = nArray[((n5 += n20) & 0xFC0) + ((n11 += n21) >> -1744130106)] >>> n22;
                        if (0 != n4) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n11 += n21;
                        n11 = (n11 & 0xFFF) + (n13 & 0xC0000);
                        n22 = n13 >> -1353915596;
                        n4 = nArray[(n11 >> 1242019238) + ((n5 += n20) & 0xFC0)] >>> n22;
                        if (n4 != 0) {
                            nArray2[n9] = n4;
                        }
                        n13 += n8;
                        ++n9;
                        n4 = nArray[((n11 += n21) >> -980660250) + (0xFC0 & (n5 += n20))] >>> n22;
                        if (0 != n4) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n4 = nArray[((n11 += n21) >> -1023624666) + (0xFC0 & (n5 += n20))] >>> n22;
                        if (~n4 != -1) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n4 = nArray[((n5 += n20) & 0xFC0) + ((n11 += n21) >> -530730842)] >>> n22;
                        if (~n4 != -1) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n11 += n21;
                        n22 = n13 >> -352222028;
                        n11 = (n13 & 0xC0000) + (n11 & 0xFFF);
                        n13 += n8;
                        n4 = nArray[((n5 += n20) & 0xFC0) - -(n11 >> -2119184058)] >>> n22;
                        if (~n4 != -1) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n4 = nArray[((n5 += n20) & 0xFC0) - -((n11 += n21) >> 1682015462)] >>> n22;
                        if (0 != n4) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n4 = nArray[((n5 += n20) & 0xFC0) + ((n11 += n21) >> 971517030)] >>> n22;
                        if (-1 != ~n4) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n4 = nArray[(0xFC0 & (n5 += n20)) - -((n11 += n21) >> -761317434)] >>> n22;
                        if (-1 != ~n4) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        if (!bl) break;
                    }
                    for (int i2 = 0; n17 > i2; ++i2) {
                        n18 = -1;
                        n4 = nArray[(n11 >> -1996690362) + (0xFC0 & n5)] >>> n22;
                        n19 = ~n4;
                        if (bl) continue block18;
                        if (n18 != n19) {
                            nArray2[n9] = n4;
                        }
                        ++n9;
                        n11 += n21;
                        n5 += n20;
                        if (3 != (3 & i2)) continue;
                        n22 = n13 >> 898327988;
                        n11 = (0xFFF & n11) + (n13 & 0xC0000);
                        n13 += n8;
                        if (!bl) continue;
                    }
                    break;
                }
                n17 -= 16;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(z[8]).append(n2).append(',').append(n3).append(',').append(n4).append(',').append(by).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(',').append(n8).append(',').append(nArray != null ? z[0] : z[1]).append(',');
            if (nArray2 != null) {
                string = z[0];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n9).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(')').toString());
            }
            string = z[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n9).append(',').append(n10).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(')').toString());
        }
    }

    static final void a(aa aa2, byte by) {
        try {
            fb.a = aa2;
            ++d;
            int n2 = -87 % ((-31 - by) / 41);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[2]);
            String string = aa2 != null ? z[0] : z[1];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(by).append(')').toString());
        }
    }

    /*
     * Unable to fully structure code
     */
    static final void a(int var0, int var1_1, int var2_2, int var3_3, int var4_4, int var5_5, int var6_6, int var7_7, int var8_8, int var9_9, int[] var10_10, int var11_11, int var12_12, int[] var13_13, byte var14_14) {
        block54: {
            var21_15 = client.vh;
            try {
                block49: {
                    block56: {
                        block48: {
                            block47: {
                                block46: {
                                    block45: {
                                        block44: {
                                            ++cb.b;
                                            if (0 >= var11_11) {
                                                return;
                                            }
                                            var15_16 = 0;
                                            var16_18 = 0;
                                            if (var14_14 <= 97) {
                                                cb.a(-65, -47, -42, (byte)-16, 62, 50, -59, -91, null, null, 71, -91, -16, -29, 110, 81);
                                            }
                                            var19_19 = 0;
                                            if (0 != var7_7) {
                                                var3_3 = var1_1 / var7_7 << -987682713;
                                                var6_6 = var2_2 / var7_7 << -1456039769;
                                            }
                                            var7_7 += var12_12;
                                            if (-1 < ~var6_6) break block44;
                                            if (~var6_6 >= -16257) break block45;
                                            var6_6 = 16256;
                                            if (!var21_15) break block45;
                                        }
                                        var6_6 = 0;
                                    }
                                    var1_1 += var5_5;
                                    var2_2 += var8_8;
                                    if (~var7_7 != -1) break block46;
                                    break block47;
                                }
                                var15_16 = var2_2 / var7_7 << 1406785287;
                                var16_18 = var1_1 / var7_7 << -129167545;
                            }
                            if (-1 < ~var15_16) break block56;
                            if (-16257 > ~var15_16) break block48;
                            break block49;
                            catch (RuntimeException v3) {
                                throw v3;
                            }
                        }
                        var15_16 = 16256;
                        if (!var21_15) break block49;
                    }
                    var15_16 = 0;
                }
                var17_20 = var15_16 + -var6_6 >> -62706076;
                var18_21 = -var3_3 + var16_18 >> 750561924;
                var20_22 = var11_11 >> -1736383548;
                while (~var20_22 < -1) {
                    block53: {
                        block57: {
                            block52: {
                                block51: {
                                    block50: {
                                        var19_19 = var4_4 >> -719426313;
                                        var6_6 += var4_4 & 0x600000;
                                        var4_4 += var9_9;
                                        var13_13[var0++] = ib.a(var13_13[var0] >> -1882151871, 0x7F7F7F) + (var10_10[(var6_6 >> -864795129) + ib.a(var3_3, 16256)] >>> var19_19);
                                        var13_13[var0++] = ib.a(var13_13[var0] >> 1149821121, 0x7F7F7F) + (var10_10[((var6_6 += var17_20) >> -1630381273) + ib.a(16256, var3_3 += var18_21)] >>> var19_19);
                                        var13_13[var0++] = (var10_10[ib.a(16256, var3_3 += var18_21) - -((var6_6 += var17_20) >> 1709041255)] >>> var19_19) - -(ib.a(0xFEFEFE, var13_13[var0]) >> -783642367);
                                        var13_13[var0++] = (ib.a(0xFEFEFE, var13_13[var0]) >> 2114203393) + (var10_10[ib.a(var3_3 += var18_21, 16256) + ((var6_6 += var17_20) >> 1285028711)] >>> var19_19);
                                        var6_6 += var17_20;
                                        var19_19 = var4_4 >> 967448183;
                                        var6_6 = (var6_6 & 16383) + (var4_4 & 0x600000);
                                        var4_4 += var9_9;
                                        var13_13[var0++] = ib.a(var13_13[var0] >> 237363553, 0x7F7F7F) + (var10_10[ib.a(16256, var3_3 += var18_21) - -(var6_6 >> 1720769863)] >>> var19_19);
                                        var13_13[var0++] = (var10_10[((var6_6 += var17_20) >> -1353166233) + ib.a(var3_3 += var18_21, 16256)] >>> var19_19) - -ib.a(var13_13[var0] >> 464826369, 0x7F7F7F);
                                        var13_13[var0++] = (ib.a(var13_13[var0], 0xFEFEFF) >> 76839841) + (var10_10[ib.a(var3_3 += var18_21, 16256) - -((var6_6 += var17_20) >> 2006644519)] >>> var19_19);
                                        var13_13[var0++] = (var10_10[((var6_6 += var17_20) >> -587801977) + ib.a(16256, var3_3 += var18_21)] >>> var19_19) - -(ib.a(var13_13[var0], 0xFEFEFF) >> -1787059871);
                                        var6_6 += var17_20;
                                        var6_6 = (16383 & var6_6) + (var4_4 & 0x600000);
                                        var19_19 = var4_4 >> 1022075575;
                                        var13_13[var0++] = (ib.a(0xFEFEFF, var13_13[var0]) >> 263459617) + (var10_10[(var6_6 >> -1125486105) + ib.a(var3_3 += var18_21, 16256)] >>> var19_19);
                                        var13_13[var0++] = ib.a(var13_13[var0] >> 254571777, 0x7F7F7F) + (var10_10[((var6_6 += var17_20) >> 201079751) + ib.a(16256, var3_3 += var18_21)] >>> var19_19);
                                        var13_13[var0++] = (var10_10[ib.a(var3_3 += var18_21, 16256) - -((var6_6 += var17_20) >> 1856596775)] >>> var19_19) + ib.a(0x7F7F7F, var13_13[var0] >> -2129743583);
                                        var13_13[var0++] = (ib.a(0xFEFEFF, var13_13[var0]) >> 345902369) + (var10_10[ib.a(16256, var3_3 += var18_21) - -((var6_6 += var17_20) >> 1601471175)] >>> var19_19);
                                        var6_6 += var17_20;
                                        var6_6 = (var6_6 & 16383) + ((var4_4 += var9_9) & 0x600000);
                                        var19_19 = var4_4 >> -1943261385;
                                        var13_13[var0++] = ib.a(0x7F7F7F, var13_13[var0] >> 1923875073) + (var10_10[(var6_6 >> 965333095) + ib.a(var3_3 += var18_21, 16256)] >>> var19_19);
                                        var4_4 += var9_9;
                                        var13_13[var0++] = ib.a(var13_13[var0] >> 481724481, 0x7F7F7F) + (var10_10[((var6_6 += var17_20) >> 1596705383) + ib.a(16256, var3_3 += var18_21)] >>> var19_19);
                                        var13_13[var0++] = (var10_10[ib.a(var3_3 += var18_21, 16256) - -((var6_6 += var17_20) >> 1419911623)] >>> var19_19) + ib.a(var13_13[var0] >> 905849729, 0x7F7F7F);
                                        var13_13[var0++] = ib.a(var13_13[var0] >> 1680585121, 0x7F7F7F) + (var10_10[((var6_6 += var17_20) >> 1937899527) + ib.a(16256, var3_3 += var18_21)] >>> var19_19);
                                        var7_7 += var12_12;
                                        var1_1 += var5_5;
                                        var2_2 += var8_8;
                                        var3_3 = var16_18;
                                        var6_6 = var15_16;
                                        v5 = -1;
                                        v6 = ~var7_7;
                                        if (!var21_15) {
                                            if (v5 != v6) break block50;
                                            break block51;
                                        }
                                        ** GOTO lbl132
                                    }
                                    var16_18 = var1_1 / var7_7 << 1649606631;
                                    var15_16 = var2_2 / var7_7 << 1163846599;
                                }
                                if (var15_16 < 0) break block57;
                                if (-16257 > ~var15_16) break block52;
                                break block53;
                                catch (RuntimeException v8) {
                                    throw v8;
                                }
                            }
                            var15_16 = 16256;
                            if (!var21_15) break block53;
                        }
                        var15_16 = 0;
                    }
                    var18_21 = var16_18 - var3_3 >> 915302724;
                    var17_20 = -var6_6 + var15_16 >> 1435337316;
                    --var20_22;
                    if (!var21_15) continue;
                }
                var20_22 = 0;
                do {
                    block55: {
                        block58: {
                            v5 = ~var20_22;
                            v6 = ~(var11_11 & 15);
lbl132:
                            // 3 sources

                            if (v5 <= v6) break block54;
                            if (var21_15) break block54;
                            break block58;
                            catch (RuntimeException v10) {
                                throw v10;
                            }
                        }
                        if (-1 != ~(var20_22 & 3)) break block55;
                        var6_6 = (var4_4 & 0x600000) + (var6_6 & 16383);
                        var19_19 = var4_4 >> -1240422601;
                        var4_4 += var9_9;
                    }
                    var13_13[var0++] = (var10_10[ib.a(var3_3, 16256) - -(var6_6 >> 1938838919)] >>> var19_19) + (ib.a(var13_13[var0], 0xFEFEFE) >> -616036735);
                    var6_6 += var17_20;
                    var3_3 += var18_21;
                    ++var20_22;
                } while (!var21_15);
            }
            catch (RuntimeException var15_17) {
                v12 = var15_17;
                v13 = new StringBuilder().append(cb.z[7]).append(var0).append(',').append(var1_1).append(',').append(var2_2).append(',').append(var3_3).append(',').append(var4_4).append(',').append(var5_5).append(',').append(var6_6).append(',').append(var7_7).append(',').append(var8_8).append(',').append(var9_9).append(',');
                v14 = var10_10 != null ? cb.z[0] : cb.z[1];
                v16 = v13.append(v14).append(',').append(var11_11).append(',').append(var12_12).append(',');
                v17 = var13_13 != null ? cb.z[0] : cb.z[1];
                throw i.a(v12, v16.append(v17).append(',').append(var14_14).append(')').toString());
            }
        }
    }

    static final void a(byte by, Object[] objectArray, int[] nArray) {
        try {
            if (by != -70) {
                return;
            }
            ub.a(nArray, (byte)-128, 0, nArray.length + -1, objectArray);
            ++g;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[9]).append(by).append(',');
            String string = objectArray != null ? z[0] : z[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',');
            String string2 = nArray != null ? z[0] : z[1];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(')').toString());
        }
    }

    static {
        z = new String[]{cb.z(cb.z("j`[\u0018W")), cb.z(cb.z("\u007f;\u0019Z")), cb.z(cb.z("r,[u\u0002")), cb.z(cb.z("r!\u001bBO\u007f:\u0016DIb")), cb.z(cb.z("r,[w\u0002")), cb.z(cb.z("R&\u0010UAx \u0012\u0016L~<UXOfn\u0016YDe+\u001bB")), cb.z(cb.z("X \u0003WFx*UuxRn\u001cX\nR\u001c6\u0016Iy+\u0016]\nw'\u0019S")), cb.z(cb.z("r,[s\u0002")), cb.z(cb.z("r,[t\u0002")), cb.z(cb.z("r,[r\u0002"))};
        c = new String[200];
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
                        n4 = 17;
                        break;
                    }
                    case 1: {
                        n4 = 78;
                        break;
                    }
                    case 2: {
                        n4 = 117;
                        break;
                    }
                    case 3: {
                        n4 = 54;
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

