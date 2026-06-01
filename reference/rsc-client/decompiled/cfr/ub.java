/*
 * Decompiled with CFR 0.152.
 */
final class ub {
    static String[] c;
    static int f;
    static String[] b;
    static int e;
    static int[] g;
    static int d;
    static String[] a;
    private static final String[] z;

    static final void a(int[] nArray, byte by, int n2, int n3, Object[] objectArray) {
        boolean bl = client.vh;
        try {
            int n4;
            int n5;
            block13: {
                if (by > -82) {
                    return;
                }
                if (~n2 > ~n3) {
                    int n6 = (n3 + n2) / 2;
                    int n7 = n2;
                    int n8 = nArray[n6];
                    nArray[n6] = nArray[n3];
                    nArray[n3] = n8;
                    Object object = objectArray[n6];
                    objectArray[n6] = objectArray[n3];
                    objectArray[n3] = object;
                    int n9 = Integer.MIN_VALUE == ~n8 ? 0 : 1;
                    int n10 = n9;
                    for (int i2 = n2; i2 < n3; ++i2) {
                        n5 = (i2 & n10) + n8;
                        n4 = nArray[i2];
                        if (bl) break block13;
                        if (n5 <= n4) continue;
                        int n11 = nArray[i2];
                        nArray[i2] = nArray[n7];
                        nArray[n7] = n11;
                        Object object2 = objectArray[i2];
                        objectArray[i2] = objectArray[n7];
                        objectArray[n7++] = object2;
                        if (!bl) continue;
                    }
                    nArray[n3] = nArray[n7];
                    nArray[n7] = n8;
                    objectArray[n3] = objectArray[n7];
                    objectArray[n7] = object;
                    ub.a(nArray, (byte)-124, n2, n7 + -1, objectArray);
                    ub.a(nArray, (byte)-123, 1 + n7, n3, objectArray);
                }
                n5 = e;
                n4 = 1;
            }
            e = n5 + n4;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[2]);
            String string = nArray != null ? z[3] : z[4];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',').append(by).append(',').append(n2).append(',').append(n3).append(',');
            String string2 = objectArray != null ? z[3] : z[4];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(')').toString());
        }
    }

    static final v a(int n2, byte by) {
        boolean bl = client.vh;
        try {
            ++d;
            if (by != 24) {
                c = null;
            }
            v[] vArray = i.a(by + -735);
            int n3 = 0;
            while (~vArray.length < ~n3) {
                v v2 = vArray[n3];
                if (n2 == v2.i) {
                    return v2;
                }
                ++n3;
                if (!bl) continue;
            }
            return null;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[0] + n2 + ',' + by + ')');
        }
    }

    static final int a(byte by) {
        try {
            ++f;
            int n2 = m.a(true, ka.b, b.v);
            if (~n2 < -100000000) {
                n2 = -n2 + 99999999;
            }
            if (by != -105) {
                g = null;
            }
            ka.b += 4;
            return n2;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[1] + by + ')');
        }
    }

    static {
        z = new String[]{ub.z(ub.z("\u0017\u000e\rJP")), ub.z(ub.z("\u0017\u000e\rKP")), ub.z(ub.z("\u0017\u000e\rIP")), ub.z(ub.z("\u0019B\r&\u0005")), ub.z(ub.z("\f\u0019Od"))};
        c = new String[5000];
        a = new String[100];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x78);
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
                        n4 = 98;
                        break;
                    }
                    case 1: {
                        n4 = 108;
                        break;
                    }
                    case 2: {
                        n4 = 35;
                        break;
                    }
                    case 3: {
                        n4 = 8;
                        break;
                    }
                    default: {
                        n4 = 120;
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

