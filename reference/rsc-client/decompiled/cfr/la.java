/*
 * Decompiled with CFR 0.152.
 */
final class la
extends RuntimeException {
    Throwable e;
    String h;
    static v b = new v(la.z(la.z("[\fDi<")), "", la.z(la.z("{,dI\u001c")), 4);
    static int[] a;
    static int d;
    static byte[] c;
    static byte[][] g;
    static String[] f;

    la(Throwable throwable, String string) {
        this.e = throwable;
        this.h = string;
    }

    static {
        c = new byte[520];
        g = new byte[12][];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x70);
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
                        n4 = 23;
                        break;
                    }
                    case 1: {
                        n4 = 67;
                        break;
                    }
                    case 2: {
                        n4 = 7;
                        break;
                    }
                    case 3: {
                        n4 = 40;
                        break;
                    }
                    default: {
                        n4 = 112;
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

