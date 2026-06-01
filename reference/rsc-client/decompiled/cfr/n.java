/*
 * Decompiled with CFR 0.152.
 */
final class n {
    static int i;
    int e;
    static int[] a;
    int k;
    static nb h;
    static String[] f;
    int d;
    int l;
    static int[] j;
    static int g;
    static int[] m;
    static int c;
    static int b;
    private static final String[] z;

    static final void a(byte by, String string) {
        if (by != -93) {
            f = null;
        }
        ++i;
        System.out.println(z[1] + jb.a(true, "\n", z[0], string));
    }

    n() {
    }

    static {
        z = new String[]{n.z(n.z("Y\u0004\u000b")), n.z(n.z("9F\u0018(qF\u0014"))};
        h = null;
        f = new String[]{n.z(n.z("9Z\u001e\"q\\Z\u001f*a\u0019FJ(e\\]\u001e\"n\u000f\u0014\u001e(#\u000eQ\u0007(u\u0019\u0014\u000b)g\\D\u0018\"p\u000f\u0014\u000f)w\u0019F"))};
        String string = n.z(n.z("=v)\u0003F:s\"\u000eI7x'\tL,e8\u0014W)b=\u001fZ&U\b$g\u0019R\r/j\u0016_\u0006*m\u0013D\u001b5p\bA\u001c0{\u0005NZv1O\u0000_q4D\rKe\u00a0X\u00114a)T\u001dG\u0018>Wo\u0011\u001a~G\u000eM\u0007 \u0002\u0018Vi=S\u000b6;#"));
        a = new int[256];
        for (int i2 = 0; i2 < 256; ++i2) {
            int n2 = string.indexOf(i2);
            if (~n2 == 0) {
                n2 = 74;
            }
            n.a[i2] = 9 * n2;
        }
        g = 0;
        j = new int[100];
        c = 0;
        b = 0;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 3);
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
                        n4 = 124;
                        break;
                    }
                    case 1: {
                        n4 = 52;
                        break;
                    }
                    case 2: {
                        n4 = 106;
                        break;
                    }
                    case 3: {
                        n4 = 71;
                        break;
                    }
                    default: {
                        n4 = 3;
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

