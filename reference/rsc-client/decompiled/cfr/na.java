/*
 * Decompiled with CFR 0.152.
 */
final class na {
    static int[] a;
    static String[] d;
    static int b;
    static int c;
    static int e;
    private static final String[] z;

    static final byte[] a(String string, int n2, byte[] byArray, int n3) {
        try {
            if (n3 > -117) {
                a = null;
            }
            ++c;
            return t.a(byArray, -127, n2, string, null);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[1]);
            String string2 = string != null ? z[2] : z[0];
            StringBuilder stringBuilder2 = stringBuilder.append(string2).append(',').append(n2).append(',');
            String string3 = byArray != null ? z[2] : z[0];
            throw i.a(runtimeException2, stringBuilder2.append(string3).append(',').append(n3).append(')').toString());
        }
    }

    static final m a(int n2, int n3, String string) {
        try {
            ++b;
            gb gb2 = new gb();
            gb2.h = string;
            gb2.f = n3;
            if (n2 != 4718) {
                return null;
            }
            return gb2;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[3]).append(n2).append(',').append(n3).append(',');
            String string2 = string != null ? z[2] : z[0];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(')').toString());
        }
    }

    static {
        z = new String[]{na.z(na.z("_.h:")), na.z(na.z("_:*\u0017G")), na.z(na.z("Ju*x\u0012")), na.z(na.z("_:*\u0014G"))};
        d = new String[]{na.z(na.z("q,l?/")), na.z(na.z("q8}7/")), na.z(na.z("q8}7/")), na.z(na.z("q,l?/")), na.z(na.z("q\"a:/")), na.z(na.z("q8}7/")), na.z(na.z("q,l?/")), na.z(na.z("q,l?/"))};
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x6F);
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
                        n4 = 49;
                        break;
                    }
                    case 1: {
                        n4 = 91;
                        break;
                    }
                    case 2: {
                        n4 = 4;
                        break;
                    }
                    case 3: {
                        n4 = 86;
                        break;
                    }
                    default: {
                        n4 = 111;
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

