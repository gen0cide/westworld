/*
 * Decompiled with CFR 0.152.
 */
final class w {
    int l;
    int r;
    int m;
    int p = -1;
    int b;
    int j;
    boolean c = false;
    int q;
    int i;
    int e;
    static int n;
    static int d;
    int k;
    int t;
    int s;
    static int[] g;
    int f = 0;
    int h;
    static int a;
    int u;
    ca o;
    private static final String[] z;

    /*
     * Unable to fully structure code
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final String a(CharSequence var0, byte var1_1) {
        var9_2 = client.vh;
        try {
            block24: {
                block23: {
                    ++w.a;
                    if (var1_1 <= 47) {
                        return null;
                    }
                    if (var0 == null) {
                        return null;
                    }
                    var2_3 = 0;
                    var3_5 = var0.length();
                    while (~var2_3 > ~var3_5) {
                        v0 = (int)db.a(32, var0.charAt(var2_3));
                        if (!var9_2) {
                            if (v0 == 0) break;
                            ++var2_3;
                            if (!var9_2) continue;
                        }
                        ** GOTO lbl19
                    }
                    do {
                        v0 = var3_5;
lbl19:
                        // 2 sources

                        if (v0 <= var2_3) break;
                        v1 = (int)db.a(32, var0.charAt(-1 + var3_5));
                        if (var9_2) break block23;
                        if (v1 == 0) break;
                        --var3_5;
                    } while (!var9_2);
                    v1 = var3_5 - var2_3;
                }
                if (1 > (var4_6 = v1)) return null;
                if (-13 > ~var4_6) return null;
                var5_7 = new StringBuilder(var4_6);
                var6_8 = var2_3;
                while (~var3_5 < ~var6_8) {
                    var7_9 = var0.charAt(var6_8);
                    v2 = (int)f.a(var7_9, 0);
                    if (!var9_2) {
                        if ((v2 != 0 || var9_2) && ~(var8_10 = ac.a(var7_9, -194)) != '\uffffffff') {
                            var5_7.append(var8_10);
                        }
                        ++var6_8;
                        if (!var9_2) continue;
                    }
                    break block24;
                }
                v2 = 0;
            }
            if (v2 == var5_7.length()) return null;
            return var5_7.toString();
        }
        catch (RuntimeException var2_4) {
            v3 = new StringBuilder().append(w.z[4]);
            if (var0 != null) {
                v4 = w.z[1];
                throw i.a(var2_4, v3.append(v4).append(',').append(var1_1).append(')').toString());
            }
            v4 = w.z[2];
            throw i.a(var2_4, v3.append(v4).append(',').append(var1_1).append(')').toString());
        }
    }

    static final int a(int n2, int n3, byte[] byArray, int n4) {
        boolean bl = client.vh;
        try {
            int n5;
            block8: {
                int n6;
                ++n;
                n5 = -1;
                for (n6 = n4; n2 > n6; ++n6) {
                    n5 = wb.q[(byArray[n6] ^ n5) & 0xFF] ^ n5 >>> 1628768296;
                    if (!bl) {
                        if (!bl) continue;
                        break;
                    }
                    break block8;
                }
                n6 = 123 / ((n3 - 23) / 63);
                n5 ^= 0xFFFFFFFF;
            }
            return n5;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[3]).append(n2).append(',').append(n3).append(',');
            String string = byArray != null ? z[1] : z[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n4).append(')').toString());
        }
    }

    static final int a(byte[] byArray, int n2, int n3) {
        try {
            int n4;
            block8: {
                block7: {
                    ++d;
                    if (n2 != -1) {
                        return 71;
                    }
                    n4 = 256 * nb.a(255, byArray[n3]) + nb.a(255, byArray[1 + n3]);
                    if (Short.MIN_VALUE > ~n4) break block7;
                    break block8;
                }
                n4 -= 65536;
            }
            return n4;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[0]);
            String string = byArray != null ? z[1] : z[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(')').toString());
        }
    }

    w() {
    }

    static {
        z = new String[]{w.z(w.z("\u0011S8B")), w.z(w.z("\u001dSTDf")), w.z(w.z("\b\b\u0016\u0006")), w.z(w.z("\u0011S9B")), w.z(w.z("\u0011S;B"))};
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
                        n4 = 102;
                        break;
                    }
                    case 1: {
                        n4 = 125;
                        break;
                    }
                    case 2: {
                        n4 = 122;
                        break;
                    }
                    case 3: {
                        n4 = 106;
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

