/*
 * Decompiled with CFR 0.152.
 */
import java.math.BigInteger;

final class ja
extends tb {
    static BigInteger K;
    static int I;
    static int[] N;
    static int J;
    static String[] L;
    static int G;
    private int M;
    static int H;
    private static final String[] O;

    final int k(int n2) {
        try {
            if (n2 != -31874) {
                return 40;
            }
            ++H;
            return this.M;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, O[1] + n2 + ')');
        }
    }

    final void i(int n2) {
        try {
            this.M = 8 * this.w;
            ++G;
            if (n2 != -2231) {
                this.k(-48);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, O[3] + n2 + ')');
        }
    }

    final int f(int n2, int n3) {
        boolean bl = client.vh;
        try {
            int n4;
            block14: {
                int n5;
                int n6;
                block13: {
                    ++J;
                    n6 = this.M >> -606412605;
                    n5 = -(this.M & 7) + 8;
                    if (n2 >= -67) {
                        K = null;
                    }
                    n4 = 0;
                    this.M += n3;
                    while (~n5 > ~n3) {
                        n4 += (mb.i[n5] & this.F[n6++]) << -n5 + n3;
                        n3 -= n5;
                        n5 = 8;
                        if (!bl) {
                            if (!bl) continue;
                            break;
                        }
                        break block13;
                    }
                    if (~n5 == ~n3) break block13;
                    n4 += this.F[n6] >> -n3 + n5 & mb.i[n3];
                    if (!bl) break block14;
                }
                n4 += this.F[n6] & mb.i[n5];
            }
            return n4;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, O[2] + n2 + ',' + n3 + ')');
        }
    }

    ja(int n2) {
        super(n2);
    }

    final void j(int n2) {
        try {
            if (n2 != 25505) {
                this.f(12, -68);
            }
            this.w = (7 + this.M) / 8;
            ++I;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, O[0] + n2 + ')');
        }
    }

    static {
        O = new String[]{ja.z(ja.z("ZS5`n")), ja.z(ja.z("ZS5en")), ja.z(ja.z("ZS5gn")), ja.z(ja.z("ZS5fn"))};
        N = new int[100];
        K = new BigInteger(ja.z(ja.z("SS\"\u0011v\u0004\u0005)E#\t\u0005-\u0011w\b\u0007yBt\t\u0002}Bs\u0004S#\u0016uR\u0003\u007f\u0016\u007fR\u0006-@%\u0003Q}\u0012q\u0006\u0000+\u0017$R\n,\u0015#VS)\u0013~T\u000bx\u0010\u007fU\u0003-@#VQ.\u0017 V\u0006,\u001du\u0000\u0007*\u0016u\u0004\u0007/\u0011v\u0005\u0002#\u0016 \u0004\u0005+\u0014$\u0000Vz\u0017~\u0001\u0002/\u0013 \u0005\u0003y\u001cq\u0002T\"F$UWz\u0012s\u0003T)\u0015 T\u0000/\u001c'\u0001\u0002}Bs\u0002\u0001\"Fu\u0000\u0000(\u0010'TV(\u0011\u007f\u0001\u0001xFp\u0000\u0004#@u\u0001\u0004~@\"\u0004\u0003#\u0012w\u0001\u0001(\u0010'U\u0002/\u0013 SV\"E%R\u0005y\u0014%\u0001\u0001y\u0017v\u0003\u000b(Et\u0006\u0000+\u0010\"S\n.\u0015~\u0003W+E\u007f\u0005\u0007.\u0011%\u0000\u0003yA#\b\u0002+\u0010r\u0000W\"\u0013rRP\"Fr\u0004\u0003}\u0010p\u0004T/\u0014s\u0007")), 16);
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x46);
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
                        n4 = 48;
                        break;
                    }
                    case 1: {
                        n4 = 50;
                        break;
                    }
                    case 2: {
                        n4 = 27;
                        break;
                    }
                    case 3: {
                        n4 = 36;
                        break;
                    }
                    default: {
                        n4 = 70;
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

