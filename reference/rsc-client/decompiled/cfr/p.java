/*
 * Decompiled with CFR 0.152.
 */
final class p {
    static boolean d;
    static String[] e;
    static int b;
    static String[] c;
    static String[] a;
    static int f;
    private static final String[] z;

    static final synchronized long a(int n2) {
        try {
            ++b;
            if (n2 != 0) {
                return -57L;
            }
            long l2 = System.currentTimeMillis();
            if ((client.ze ^ 0xFFFFFFFFFFFFFFFFL) < (l2 ^ 0xFFFFFFFFFFFFFFFFL)) {
                wb.w += client.ze + -l2;
            }
            client.ze = l2;
            return wb.w + l2;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[0] + n2 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final void a(int n2, int n3, int n4, int n5, int n6, int[] nArray, int n7, int n8, int n9, int n10, int[] nArray2, int n11, int n12, int n13, int n14) {
        boolean bl = client.vh;
        try {
            int n15;
            int n16;
            block28: {
                block27: {
                    ++f;
                    if (0 >= n14) {
                        return;
                    }
                    n16 = 0;
                    n15 = 0;
                    if (n12 != 0) {
                        n15 = n5 / n12 << -258111322;
                        n16 = n9 / n12 << 637317126;
                    }
                    n2 <<= 2;
                    if (~n16 > -1) break block27;
                    if (n16 <= 4032) break block28;
                    n16 = 4032;
                    if (!bl) break block28;
                }
                n16 = 0;
            }
            if (n3 != 1121159302) {
                p.a(-69, 127, -20, -29, -78, null, 16, 2, -77, -5, null, 113, -57, 68, -87);
            }
            int n17 = n14;
            do {
                int n18 = 0;
                int n19 = n17;
                block12: while (true) {
                    block30: {
                        block29: {
                            if (n18 >= n19) return;
                            n12 += n13;
                            n9 += n6;
                            n5 += n4;
                            n10 = n16;
                            n8 = n15;
                            if (bl) return;
                            if (-1 != ~n12) {
                                n16 = n9 / n12 << 25779686;
                                n15 = n5 / n12 << 1121159302;
                            }
                            if (-1 >= ~n16) break block29;
                            n16 = 0;
                            if (!bl) break block30;
                        }
                        if (~n16 < -4033) {
                            n16 = 4032;
                        }
                    }
                    int n20 = n15 + -n8 >> 1397627332;
                    int n21 = -n10 + n16 >> 1542826468;
                    int n22 = n7 >> -1668610924;
                    n10 += 0xC0000 & n7;
                    n7 += n2;
                    if (~n17 <= -17) {
                        nArray2[n11++] = nArray[ib.a(n8, 4032) - -(n10 >> -1148525818)] >>> n22;
                        nArray2[n11++] = nArray[((n10 += n21) >> 1034190278) + ib.a(n8 += n20, 4032)] >>> n22;
                        nArray2[n11++] = nArray[((n10 += n21) >> -385010618) + ib.a(4032, n8 += n20)] >>> n22;
                        nArray2[n11++] = nArray[((n10 += n21) >> 747209702) + ib.a(4032, n8 += n20)] >>> n22;
                        n10 += n21;
                        n22 = n7 >> 597207284;
                        n10 = (n7 & 0xC0000) + (0xFFF & n10);
                        n7 += n2;
                        nArray2[n11++] = nArray[ib.a(4032, n8 += n20) + (n10 >> 831423910)] >>> n22;
                        nArray2[n11++] = nArray[ib.a(n8 += n20, 4032) - -((n10 += n21) >> -512409978)] >>> n22;
                        nArray2[n11++] = nArray[ib.a(n8 += n20, 4032) + ((n10 += n21) >> -783757370)] >>> n22;
                        nArray2[n11++] = nArray[((n10 += n21) >> -129948154) + ib.a(4032, n8 += n20)] >>> n22;
                        n10 += n21;
                        n22 = n7 >> 92466196;
                        n10 = (0xC0000 & n7) + (0xFFF & n10);
                        n7 += n2;
                        nArray2[n11++] = nArray[ib.a(n8 += n20, 4032) - -(n10 >> -1989449594)] >>> n22;
                        nArray2[n11++] = nArray[((n10 += n21) >> -76155226) + ib.a(n8 += n20, 4032)] >>> n22;
                        nArray2[n11++] = nArray[ib.a(n8 += n20, 4032) - -((n10 += n21) >> -158732986)] >>> n22;
                        nArray2[n11++] = nArray[ib.a(4032, n8 += n20) + ((n10 += n21) >> 1960099526)] >>> n22;
                        n10 += n21;
                        n22 = n7 >> 1740031764;
                        n10 = (0xFFF & n10) - -(n7 & 0xC0000);
                        n7 += n2;
                        nArray2[n11++] = nArray[ib.a(n8 += n20, 4032) + (n10 >> -1366214458)] >>> n22;
                        nArray2[n11++] = nArray[((n10 += n21) >> -1971655962) + ib.a(n8 += n20, 4032)] >>> n22;
                        nArray2[n11++] = nArray[((n10 += n21) >> -674692218) + ib.a(n8 += n20, 4032)] >>> n22;
                        nArray2[n11++] = nArray[ib.a(4032, n8 += n20) - -((n10 += n21) >> -1886734874)] >>> n22;
                        if (!bl) break;
                    }
                    for (int i2 = 0; i2 < n17; ++i2) {
                        nArray2[n11++] = nArray[(n10 >> -1102323802) + ib.a(4032, n8)] >>> n22;
                        n8 += n20;
                        n10 += n21;
                        n18 = 3;
                        n19 = 3 & i2;
                        if (bl) continue block12;
                        if (n18 != n19) continue;
                        n22 = n7 >> -667502188;
                        n10 = (n7 & 0xC0000) + (0xFFF & n10);
                        n7 += n2;
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
            StringBuilder stringBuilder = new StringBuilder().append(z[3]).append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',').append(n6).append(',').append(nArray != null ? z[2] : z[1]).append(',').append(n7).append(',').append(n8).append(',').append(n9).append(',').append(n10).append(',');
            if (nArray2 != null) {
                string = z[2];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(')').toString());
            }
            string = z[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n11).append(',').append(n12).append(',').append(n13).append(',').append(n14).append(')').toString());
        }
    }

    static {
        z = new String[]{p.z(p.z("/ 3\t")), p.z(p.z("1{\u001eM")), p.z(p.z("$ \\\u000f\u000f")), p.z(p.z("/ 0\t"))};
        d = true;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x72);
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
                        n4 = 95;
                        break;
                    }
                    case 1: {
                        n4 = 14;
                        break;
                    }
                    case 2: {
                        n4 = 114;
                        break;
                    }
                    case 3: {
                        n4 = 33;
                        break;
                    }
                    default: {
                        n4 = 114;
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

