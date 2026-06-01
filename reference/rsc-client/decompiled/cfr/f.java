/*
 * Decompiled with CFR 0.152.
 */
import java.io.IOException;

final class f {
    static int[] f;
    static int d;
    static int a;
    static String[] e;
    static String[] c;
    static i b;
    private static final String[] z;

    /*
     * Unable to fully structure code
     */
    static final boolean a(char var0, int var1_1) {
        var5_2 = client.vh;
        try {
            block12: {
                ++f.a;
                if (Character.isISOControl(var0)) {
                    return false;
                }
                if (v.a(var0, 115)) {
                    return true;
                }
                var2_3 = ga.a;
                var3_5 = var1_1;
                while (~var3_5 > ~var2_3.length) {
                    block11: {
                        var4_6 = var2_3[var3_5];
                        v0 = ~var0;
                        v1 = ~var4_6;
                        if (!var5_2) {
                            if (v0 != v1) break block11;
                        }
                        ** GOTO lbl30
                        return true;
                    }
                    ++var3_5;
                    if (!var5_2) continue;
                }
                var2_3 = ac.I;
                var3_5 = 0;
                do {
                    block13: {
                        v0 = var2_3.length;
                        v1 = var3_5;
lbl30:
                        // 2 sources

                        if (v0 <= v1) break;
                        var4_6 = var2_3[var3_5];
                        v3 = var4_6;
                        if (var5_2) break block12;
                        if (v3 != var0) break block13;
                        return true;
                    }
                    ++var3_5;
                } while (!var5_2);
                v3 = '\u0000';
            }
            return (boolean)v3;
        }
        catch (RuntimeException var2_4) {
            throw i.a(var2_4, f.z[0] + var0 + ',' + var1_1 + ')');
        }
    }

    static final void a(int n2, tb tb2) {
        block22: {
            boolean bl = client.vh;
            try {
                byte[] byArray;
                block20: {
                    block19: {
                        ++d;
                        byArray = new byte[24];
                        if (b.q != null) break block19;
                        break block20;
                    }
                    try {
                        int n3;
                        int n4;
                        block21: {
                            int n5;
                            b.q.a(0L, n2 + -22592);
                            b.q.a((byte)-123, byArray);
                            for (n5 = 0; n5 < 24; ++n5) {
                                n4 = -1;
                                n3 = ~byArray[n5];
                                if (bl) break block21;
                                if (n4 == n3) continue;
                                if (!bl) break;
                                if (!bl) continue;
                                break;
                            }
                            n4 = -25;
                            n3 = ~n5;
                        }
                        if (n4 >= n3) {
                            throw new IOException();
                        }
                    }
                    catch (Exception exception) {
                        int n6 = 0;
                        while (-25 < ~n6) {
                            byArray[n6] = -1;
                            ++n6;
                            if (!bl) {
                                if (!bl) continue;
                                break;
                            }
                            break block22;
                        }
                    }
                }
                if (n2 != 22607) {
                    return;
                }
                tb2.a(0, -126, 24, byArray);
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(z[2]).append(n2).append(',');
                String string = tb2 != null ? z[3] : z[1];
                throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
            }
        }
    }

    static {
        z = new String[]{f.z(f.z("UNLh")), f.z(f.z("]\u0015a,")), f.z(f.z("UNOh")), f.z(f.z("HN#nm"))};
        c = new String[]{f.z(f.z("c\fh!cV@h.dV\u0012-4xV@c5}Q\u0005\u007f`\u007fU@d4u^\u0013-4\u007f\u0013\u0004h0\u007f@\ty")), f.z(f.z("R\u000ei``A\u0005~30V\u000ey%b"))};
        b = new i(f.z(f.z("d)]")), 2);
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x10);
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
                        n4 = 51;
                        break;
                    }
                    case 1: {
                        n4 = 96;
                        break;
                    }
                    case 2: {
                        n4 = 13;
                        break;
                    }
                    case 3: {
                        n4 = 64;
                        break;
                    }
                    default: {
                        n4 = 16;
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

