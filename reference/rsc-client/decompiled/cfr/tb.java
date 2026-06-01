/*
 * Decompiled with CFR 0.152.
 */
import java.math.BigInteger;

class tb
extends ib {
    static int B;
    static int v;
    static int r;
    static int D;
    static int x;
    static int g;
    static int m;
    static int o;
    static int i;
    byte[] F;
    static int A;
    static int q;
    int w;
    static int n;
    static int E;
    static int k;
    static int[] l;
    static int s;
    static int C;
    static int u;
    static int j;
    static int z;
    static int p;
    static int h;
    static int y;
    static int t;
    private static final String[] G;

    final void a(String string, int n2) {
        try {
            block7: {
                block6: {
                    ++x;
                    int n3 = string.indexOf(0);
                    if (0 <= n3) break block6;
                    break block7;
                }
                throw new IllegalArgumentException("");
            }
            this.F[this.w++] = 0;
            this.w += i.a(string.length(), this.w, 0, string, (byte)-118, this.F);
            int n4 = 53 / ((n2 - 45) / 55);
            this.F[this.w++] = 0;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[17]);
            String string2 = string != null ? G[3] : G[4];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n2).append(')').toString());
        }
    }

    final void a(int n2, byte by) {
        try {
            this.F[this.w++] = (byte)(n2 >> -592188912);
            ++q;
            this.F[this.w++] = (byte)(n2 >> -1114664312);
            if (by != -13) {
                return;
            }
            this.F[this.w++] = (byte)n2;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[12] + n2 + ',' + by + ')');
        }
    }

    final boolean e(int n2) {
        try {
            this.w -= 4;
            ++t;
            if (n2 != -422797528) {
                return false;
            }
            int n3 = w.a(this.w, 107, this.F, 0);
            int n4 = this.b(-129);
            return n4 == n3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[19] + n2 + ')');
        }
    }

    private final void a(boolean bl, int n2, int n3, byte[] byArray) {
        boolean bl2 = client.vh;
        try {
            block11: {
                int n4 = n2;
                while (~(n3 + n2) < ~n4) {
                    byArray[n4] = this.F[this.w++];
                    ++n4;
                    if (!bl2) {
                        if (!bl2) continue;
                        break;
                    }
                    break block11;
                }
                ++i;
            }
            if (bl) {
                this.a(null, -94, null);
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[16]).append(bl).append(',').append(n2).append(',').append(n3).append(',');
            String string = byArray != null ? G[3] : G[4];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    final void a(int n2, int n3, int n4, byte[] byArray) {
        block11: {
            boolean bl = client.vh;
            try {
                int n5 = n2;
                while (~(n4 + n2) < ~n5) {
                    this.F[this.w++] = byArray[n5];
                    ++n5;
                    if (!bl) {
                        if (!bl) continue;
                        break;
                    }
                    break block11;
                }
                if (n3 >= -120) {
                    l = null;
                }
                ++D;
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(G[5]).append(n2).append(',').append(n3).append(',').append(n4).append(',');
                String string = byArray != null ? G[3] : G[4];
                throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
            }
        }
    }

    final void c(int n2, int n3) {
        try {
            this.F[this.w++] = (byte)n2;
            int n4 = 121 / ((-5 - n3) / 32);
            ++E;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[0] + n2 + ',' + n3 + ')');
        }
    }

    final void a(byte by, String string) {
        try {
            block10: {
                block9: {
                    ++A;
                    int n2 = string.indexOf(0);
                    if (~n2 <= -1) break block9;
                    break block10;
                }
                throw new IllegalArgumentException("");
            }
            this.w += i.a(string.length(), this.w, 0, string, (byte)-112, this.F);
            this.F[this.w++] = 0;
            if (by != -39) {
                this.h(-74);
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[15]).append(by).append(',');
            String string2 = string != null ? G[3] : G[4];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(')').toString());
        }
    }

    final void e(int n2, int n3) {
        try {
            ++C;
            this.F[this.w++] = (byte)(n3 >> 436303880);
            if (n2 != 393) {
                return;
            }
            this.F[this.w++] = (byte)n3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[11] + n2 + ',' + n3 + ')');
        }
    }

    final byte h(int n2) {
        try {
            if (n2 != 20869) {
                return 113;
            }
            ++v;
            return this.F[this.w++];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[22] + n2 + ')');
        }
    }

    final int a(byte by) {
        try {
            ++p;
            if (by != 104) {
                this.b(111, (byte)-26);
            }
            return this.F[this.w++] & 0xFF;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[18] + by + ')');
        }
    }

    final void a(byte by, int n2, int[] nArray, int n3) {
        block10: {
            boolean bl = client.vh;
            try {
                ++m;
                int n4 = this.w;
                if (by != 87) {
                    return;
                }
                this.w = n2;
                int n5 = (-n2 + n3) / 8;
                int n6 = 0;
                while (n6 < n5) {
                    block11: {
                        int n7 = this.b(-129);
                        int n8 = this.b(-129);
                        int n9 = 0;
                        int n10 = -1640531527;
                        if (bl) break block10;
                        int n11 = 32;
                        while (~n11-- < -1) {
                            n8 += (n7 += (n8 << 1853481540 ^ n8 >>> -1249842747) + n8 ^ n9 + nArray[n9 & 3]) + (n7 >>> -820868603 ^ n7 << 683776932) ^ (n9 += n10) - -nArray[(0x1BE9 & n9) >>> 2036143115];
                            if (!bl) {
                                if (!bl) continue;
                                break;
                            }
                            break block11;
                        }
                        this.w -= 8;
                        this.b(-422797528, n7);
                        this.b(-422797528, n8);
                        ++n6;
                    }
                    if (!bl) continue;
                }
                this.w = n4;
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(G[25]).append(by).append(',').append(n2).append(',');
                String string = nArray != null ? G[3] : G[4];
                throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n3).append(')').toString());
            }
        }
    }

    final void d(int n2, int n3) {
        try {
            this.F[-2 + (-n2 + this.w)] = (byte)(n2 >> 2065078440);
            ++k;
            this.F[this.w + (-n2 + -1)] = (byte)n2;
            if (n3 != 1) {
                this.a(null, 53);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[23] + n2 + ',' + n3 + ')');
        }
    }

    final int b(byte by) {
        try {
            if (by != 68) {
                return 53;
            }
            ++y;
            int n2 = 0xFF & this.F[this.w];
            if (-129 < ~n2) {
                return this.a((byte)104);
            }
            return Short.MIN_VALUE + this.f(255);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[13] + by + ')');
        }
    }

    final int a(boolean bl) {
        try {
            this.w += 2;
            ++u;
            if (bl) {
                return -8;
            }
            int n2 = (0xFF & this.F[this.w - 1]) + (this.F[-2 + this.w] << -1500474744 & 0xFF00);
            if (Short.MIN_VALUE > ~n2) {
                n2 -= 65536;
            }
            return n2;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[20] + bl + ')');
        }
    }

    final int c(int n2) {
        try {
            ++n;
            if (n2 != 103) {
                return 72;
            }
            if (0 > this.F[this.w]) {
                return Integer.MAX_VALUE & this.b(-129);
            }
            return this.f(n2 + 152);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[1] + n2 + ')');
        }
    }

    final String c(byte by) {
        try {
            block10: {
                block9: {
                    ++o;
                    if (by != -44) {
                        this.d(-84);
                    }
                    byte by2 = this.F[this.w++];
                    if (by2 != 0) break block9;
                    break block10;
                }
                throw new IllegalStateException("");
            }
            int n2 = this.w;
            while (this.F[this.w++] != 0) {
            }
            int n3 = -1 + (this.w - n2);
            if (-1 == ~n3) {
                return "";
            }
            return ga.a(n3, by + -68, n2, this.F);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[7] + by + ')');
        }
    }

    final void b(int n2, byte by) {
        try {
            if (by > -62) {
                return;
            }
            ++h;
            if (0 <= n2 && 128 > n2) {
                this.c(n2, 43);
                return;
            }
            if (n2 >= 0 && n2 < 32768) {
                this.e(393, 32768 - -n2);
                return;
            }
            throw new IllegalArgumentException();
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[8] + n2 + ',' + by + ')');
        }
    }

    final void b(int n2, int n3) {
        try {
            ++s;
            this.F[this.w++] = (byte)(n3 >> -2105201640);
            if (n2 != -422797528) {
                this.c(-62, 1);
            }
            this.F[this.w++] = (byte)(n3 >> -952226864);
            this.F[this.w++] = (byte)(n3 >> -422797528);
            this.F[this.w++] = (byte)n3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[21] + n2 + ',' + n3 + ')');
        }
    }

    final int f(int n2) {
        try {
            ++B;
            if (n2 != 255) {
                this.a(null, 71, null);
            }
            this.w += 2;
            return ((this.F[-2 + this.w] & 0xFF) << -958656888) - -(0xFF & this.F[-1 + this.w]);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[14] + n2 + ')');
        }
    }

    final int b(int n2) {
        try {
            this.w += 4;
            if (n2 != -129) {
                return 124;
            }
            ++r;
            return (this.F[this.w - 3] << 1172488496 & 0xFF0000) + (this.F[this.w + -4] << 2040727736 & 0xFF000000) - (-(0xFF00 & this.F[this.w + -2] << -1377058840) - (this.F[-1 + this.w] & 0xFF));
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[24] + n2 + ')');
        }
    }

    final long g(int n2) {
        try {
            if (n2 != 0) {
                return -13L;
            }
            ++g;
            long l2 = (long)this.b(-129) & 0xFFFFFFFFL;
            long l3 = (long)this.b(-129) & 0xFFFFFFFFL;
            return (l2 << 1382465952) - -l3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[2] + n2 + ')');
        }
    }

    tb(int n2) {
        try {
            this.F = mb.a(n2, (byte)-104);
            this.w = 0;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[10] + n2 + ')');
        }
    }

    final byte[] d(int n2) {
        boolean bl = client.vh;
        try {
            byte[] byArray;
            block6: {
                ++z;
                byte[] byArray2 = new byte[this.w];
                int n3 = n2;
                while (~n3 > ~this.w) {
                    byArray = byArray2;
                    if (!bl) {
                        byArray[n3] = this.F[n3];
                        ++n3;
                        if (!bl) continue;
                        break;
                    }
                    break block6;
                }
                byArray = byArray2;
            }
            return byArray;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, G[6] + n2 + ')');
        }
    }

    tb(byte[] byArray) {
        try {
            this.F = byArray;
            this.w = 0;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[10]);
            String string = byArray != null ? G[3] : G[4];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    final void a(BigInteger bigInteger, int n2, BigInteger bigInteger2) {
        try {
            int n3 = -98 / ((n2 - 6) / 52);
            ++j;
            int n4 = this.w;
            this.w = 0;
            byte[] byArray = new byte[n4];
            this.a(false, 0, n4, byArray);
            BigInteger bigInteger3 = new BigInteger(byArray);
            BigInteger bigInteger4 = bigInteger3.modPow(bigInteger2, bigInteger);
            byte[] byArray2 = bigInteger4.toByteArray();
            this.w = 0;
            this.e(393, byArray2.length);
            this.a(0, -127, byArray2.length, byArray2);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(G[9]);
            String string = bigInteger != null ? G[3] : G[4];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',').append(n2).append(',');
            String string2 = bigInteger2 != null ? G[3] : G[4];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(')').toString());
        }
    }

    static {
        G = new String[]{tb.z(tb.z("#.09R")), tb.z(tb.z("#.0%R")), tb.z(tb.z("#.0!R")), tb.z(tb.z(",b0\\\u0007")), tb.z(tb.z("99r\u001e")), tb.z(tb.z("#.03;\u007f")), tb.z(tb.z("#.0&R")), tb.z(tb.z("#.0?R")), tb.z(tb.z("#.0:R")), tb.z(tb.z("#.0$R")), tb.z(tb.z("#.0N\u00139%jLR")), tb.z(tb.z("#.04R")), tb.z(tb.z("#.0;R")), tb.z(tb.z("#.0\"R")), tb.z(tb.z("#.05R")), tb.z(tb.z("#.08R")), tb.z(tb.z("#.0<R")), tb.z(tb.z("#.0=R")), tb.z(tb.z("#.07R")), tb.z(tb.z("#.00;\u007f")), tb.z(tb.z("#.06;\u007f")), tb.z(tb.z("#.0#R")), tb.z(tb.z("#.0>R")), tb.z(tb.z("#.0 R")), tb.z(tb.z("#.0'R")), tb.z(tb.z("#.01;\u007f"))};
        l = new int[12];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x7A);
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
                        n4 = 87;
                        break;
                    }
                    case 1: {
                        n4 = 76;
                        break;
                    }
                    case 2: {
                        n4 = 30;
                        break;
                    }
                    case 3: {
                        n4 = 114;
                        break;
                    }
                    default: {
                        n4 = 122;
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

