/*
 * Decompiled with CFR 0.152.
 */
final class ac {
    byte[] j;
    int L;
    int E;
    int G;
    byte[] q;
    boolean[] n;
    static String[] z;
    int[][] J = new int[6][258];
    static String[] x;
    int[] r;
    int h;
    int F;
    byte[] i;
    static int C;
    byte[] d;
    byte[] s;
    int f;
    byte[] A;
    int c;
    int[] D;
    static int[] l;
    int o = 0;
    int y;
    int p;
    int[][] t;
    int[] w;
    int a = 0;
    static char[] I;
    int b;
    static int k;
    int[] m;
    int e;
    int H;
    boolean[] v;
    int K;
    int[][] u;
    byte g;
    byte[][] B;
    private static final String[] M;

    static final void a(ib ib2, byte by, ib ib3) {
        try {
            if (null != ib2.e) {
                ib2.a(-27331);
            }
            ++k;
            if (by != 34) {
                x = null;
            }
            ib2.a = ib3;
            ib2.e = ib3.e;
            ib2.e.a = ib2;
            ib2.a.e = ib2;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(M[2]);
            String string = ib2 != null ? M[3] : M[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',').append(by).append(',');
            String string2 = ib3 != null ? M[3] : M[1];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final char a(char c2, int n2) {
        boolean bl = client.vh;
        try {
            if (n2 != -194) {
                ac.a('\uff86', 87);
            }
            ++C;
            char c3 = c2;
            if ('\uffffffdf' == ~c3) {
                if (!bl) return '_';
            }
            if ('\u00a0' == c3) return '_';
            if ('_' == c3) return '_';
            if ('\uffffffd2' == ~c3) {
                if (!bl) return '_';
            }
            if (~c3 == '\uffffffa4') {
                if (!bl) return c2;
            }
            if (-94 == ~c3) return c2;
            if ('#' == c3) {
                if (!bl) return c2;
            }
            if ('\u00e0' == c3) {
                if (!bl) return 'a';
            }
            if ('\u00e1' == c3) {
                if (!bl) return 'a';
            }
            if (~c3 == '\uffffff1d') {
                if (!bl) return 'a';
            }
            if (c3 == '\u00e4') {
                if (!bl) return 'a';
            }
            if (c3 == '\u00e3') return 'a';
            if (c3 == '\u00c0') {
                if (!bl) return 'a';
            }
            if ('\uffffff3e' == ~c3) {
                if (!bl) return 'a';
            }
            if ('\u00c2' == c3) return 'a';
            if (c3 == '\u00c4') return 'a';
            if ('\u00c3' == c3) {
                if (!bl) return 'a';
            }
            if (c3 == '\u00e8') return 'e';
            if ('\uffffff16' == ~c3) {
                if (!bl) return 'e';
            }
            if ('\u00ea' == c3) return 'e';
            if ('\u00eb' == c3) return 'e';
            if (~c3 == -201) return 'e';
            if ('\u00c9' == c3) return 'e';
            if (c3 == '\u00ca') return 'e';
            if (~c3 == -204) return 'e';
            if (c3 == '\u00ed') return 'i';
            if (c3 == '\u00ee') {
                if (!bl) return 'i';
            }
            if (~c3 == -240) return 'i';
            if (~c3 == '\uffffff32') {
                if (!bl) return 'i';
            }
            if (~c3 == '\uffffff31') {
                if (!bl) return 'i';
            }
            if (~c3 == -208) return 'i';
            if ('\uffffff0d' == ~c3) {
                if (!bl) return 'o';
            }
            if (-244 == ~c3) return 'o';
            if ('\uffffff0b' == ~c3) {
                if (!bl) return 'o';
            }
            if (-247 == ~c3) return 'o';
            if (c3 == '\u00f5') {
                if (!bl) return 'o';
            }
            if (c3 == '\u00d2') return 'o';
            if (~c3 == '\uffffff2c') {
                if (!bl) return 'o';
            }
            if (c3 == '\u00d4') return 'o';
            if (~c3 == -215) return 'o';
            if (-214 == ~c3) return 'o';
            if ('\u00f9' == c3) {
                if (!bl) return 'u';
            }
            if ('\uffffff05' == ~c3) {
                if (!bl) return 'u';
            }
            if ('\uffffff04' == ~c3) {
                if (!bl) return 'u';
            }
            if ('\u00fc' == c3) return 'u';
            if (~c3 == -218) return 'u';
            if (c3 == '\u00da') return 'u';
            if (c3 == '\u00db') {
                if (!bl) return 'u';
            }
            if ('\u00dc' == c3) return 'u';
            if (-232 == ~c3) return 'c';
            if ('\u00c7' == c3) return 'c';
            if ('\u00ff' == c3) {
                if (!bl) return 'y';
            }
            if ('\ufffffe87' == ~c3) {
                if (!bl) return 'y';
            }
            if (c3 == '\u00f1') {
                if (!bl) return 'n';
            }
            if ('\u00d1' == c3) {
                if (!bl) return 'n';
            }
            if (-224 == ~c3) return 'b';
            return Character.toLowerCase(c2);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, M[0] + c2 + ',' + n2 + ')');
        }
    }

    ac() {
        this.j = new byte[18002];
        this.n = new boolean[256];
        this.m = new int[256];
        this.r = new int[16];
        this.s = new byte[18002];
        this.A = new byte[4096];
        this.w = new int[257];
        this.u = new int[6][258];
        this.D = new int[6];
        this.d = new byte[256];
        this.v = new boolean[16];
        this.B = new byte[6][258];
        this.t = new int[6][258];
    }

    static {
        M = new String[]{ac.z(ac.z("h_og%")), ac.z(ac.z("gI-J")), ac.z(ac.z("h_od%")), ac.z(ac.z("r\u0012o\bp"))};
        z = new String[200];
        I = new char[]{'[', ']', '#'};
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0xD);
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
                        n4 = 9;
                        break;
                    }
                    case 1: {
                        n4 = 60;
                        break;
                    }
                    case 2: {
                        n4 = 65;
                        break;
                    }
                    case 3: {
                        n4 = 38;
                        break;
                    }
                    default: {
                        n4 = 13;
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

