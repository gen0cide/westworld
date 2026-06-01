/*
 * Decompiled with CFR 0.152.
 */
final class v {
    static int c;
    int i;
    static int[] a;
    static int f;
    static int h;
    static int d;
    static int b;
    static int[] g;
    static int[] e;
    private static final String[] z;

    static final byte[] a(byte[] byArray, int n2, int n3, int n4) {
        boolean bl = client.vh;
        try {
            byte[] byArray2;
            block9: {
                ++c;
                if (n3 != -98) {
                    a = null;
                }
                byte[] byArray3 = new byte[n2];
                for (int i2 = 0; i2 < n2; ++i2) {
                    byArray2 = byArray3;
                    if (!bl) {
                        byArray2[i2] = qa.l[ib.a(byArray[n4 + i2], 255)];
                        if (!bl) continue;
                        break;
                    }
                    break block9;
                }
                byArray2 = byArray3;
            }
            return byArray2;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[3]);
            String string = byArray != null ? z[2] : z[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(')').toString());
        }
    }

    static final int a(int n2) {
        try {
            ++d;
            int n3 = b.v[ka.b] & 0xFF;
            ++ka.b;
            if (n2 != -30504) {
                v.a(113);
            }
            return n3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[4] + n2 + ')');
        }
    }

    public final String toString() {
        try {
            ++f;
            throw new IllegalStateException();
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[5]);
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final boolean a(char c2, int n2) {
        try {
            ++b;
            if (n2 <= 111) {
                v.a(null, 51, 127, 27);
            }
            if (~c2 <= '\uffffffcf') {
                if (~c2 >= -58) return true;
            }
            if (c2 >= 'A') {
                if (~c2 >= -91) return true;
            }
            if (-98 < ~c2) return false;
            if (-123 > ~c2) return false;
            return true;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[6] + c2 + ',' + n2 + ')');
        }
    }

    v(String string, String string2, String string3, int n2) {
        try {
            this.i = n2;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[1]);
            String string4 = string != null ? z[2] : z[0];
            StringBuilder stringBuilder2 = stringBuilder.append(string4).append(',');
            String string5 = string2 != null ? z[2] : z[0];
            StringBuilder stringBuilder3 = stringBuilder2.append(string5).append(',');
            String string6 = string3 != null ? z[2] : z[0];
            throw i.a(runtimeException2, stringBuilder3.append(string6).append(',').append(n2).append(')').toString());
        }
    }

    static {
        z = new String[]{v.z(v.z("\"tv%")), v.z(v.z(":/& /%u$a")), v.z(v.z("7/4g<")), v.z(v.z(":/Ya")), v.z(v.z(":/[a")), v.z(v.z(":/n&\u00128ss'&d(")), v.z(v.z(":/Xa"))};
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x41);
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
                        n4 = 76;
                        break;
                    }
                    case 1: {
                        n4 = 1;
                        break;
                    }
                    case 2: {
                        n4 = 26;
                        break;
                    }
                    case 3: {
                        n4 = 73;
                        break;
                    }
                    default: {
                        n4 = 65;
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

