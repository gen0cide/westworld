/*
 * Decompiled with CFR 0.152.
 */
final class oa {
    static int[][] d;
    static int e;
    static String[] c;
    static String[] a;
    static int b;
    private static final String[] z;

    static final int a(String string, byte by, byte[] byArray) {
        boolean bl = client.vh;
        try {
            int n2;
            block15: {
                int n3;
                int n4;
                int n5;
                block14: {
                    ++e;
                    n5 = d.a(0, (byte)48, byArray);
                    n4 = 0;
                    string = string.toUpperCase();
                    n3 = 0;
                    while (~string.length() < ~n3) {
                        n4 = -32 + string.charAt(n3) + 61 * n4;
                        ++n3;
                        if (!bl) {
                            if (!bl) continue;
                            break;
                        }
                        break block14;
                    }
                    n3 = 2 + n5 * 10;
                }
                if (by != 68) {
                    return 71;
                }
                int n6 = 0;
                while (~n6 > ~n5) {
                    int n7;
                    block16: {
                        int n8 = (byArray[5 + 10 * n6] & 0xFF) + ((0xFF & byArray[4 + 10 * n6]) * 256 + (0xFF & byArray[2 + 10 * n6]) * 0x1000000 + 65536 * (byArray[10 * n6 + 3] & 0xFF));
                        n7 = (byArray[10 * n6 + 11] & 0xFF) + 256 * (byArray[n6 * 10 - -10] & 0xFF) + (0xFF & byArray[9 + 10 * n6]) * 65536;
                        n2 = ~n4;
                        if (bl) break block15;
                        if (n2 != ~n8) break block16;
                        return n3;
                    }
                    n3 += n7;
                    ++n6;
                    if (!bl) continue;
                }
                n2 = 0;
            }
            return n2;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[0]);
            String string2 = string != null ? z[2] : z[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string2).append(',').append(by).append(',');
            String string3 = byArray != null ? z[2] : z[1];
            throw i.a(runtimeException2, stringBuilder2.append(string3).append(')').toString());
        }
    }

    static {
        z = new String[]{oa.z(oa.z("x\fmz\u000f")), oa.z(oa.z("y\u0018/W")), oa.z(oa.z("lCm\u0015Z"))};
        c = new String[]{oa.z(oa.z("R\u00037^U7\u00036VEr\u001fcTA7\u00047^JdM7T\u0007d\u0019\"PB7\f-_\u0007g\u001f&HT7\b-OBe"))};
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x27);
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
                        n4 = 109;
                        break;
                    }
                    case 2: {
                        n4 = 67;
                        break;
                    }
                    case 3: {
                        n4 = 59;
                        break;
                    }
                    default: {
                        n4 = 39;
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

