/*
 * Decompiled with CFR 0.152.
 */
final class t {
    int e;
    int i;
    static int f;
    int j;
    int l;
    int m;
    static int c;
    String b;
    static int a;
    String p = null;
    String o = null;
    int d;
    static int k;
    static int g;
    static String[] h;
    static byte[][][] n;
    private static final String[] z;

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final void a(int n2, int n3, int n4, int[] nArray, int[] nArray2, int n5, int n6, int n7) {
        boolean bl = client.vh;
        try {
            int n8;
            int n9;
            block15: {
                ++a;
                if (n4 >= 0) {
                    return;
                }
                n2 = nArray2[0xFF & n5 >> -1455207192];
                n5 += (n3 <<= 2);
                n8 = n9 = n4 / 16;
                while (-1 < ~n8) {
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    n2 = nArray2[(n5 & 0xFFF5) >> 59579400];
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    n2 = nArray2[(0xFFBD & (n5 += n3)) >> 1966207912];
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    n2 = nArray2[0xFF & (n5 += n3) >> 1256567432];
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    nArray[n6++] = n2;
                    n2 = nArray2[(0xFF32 & (n5 += n3)) >> 1042063560];
                    n5 += n3;
                    ++n8;
                    if (!bl) {
                        if (!bl) continue;
                    }
                    break block15;
                }
                n9 = -(n4 % 16);
            }
            if (n7 != 418609192) {
                t.a(76, -63, -59, null, null, 124, -66, -99);
            }
            n8 = 0;
            do {
                if (n9 <= n8) return;
                nArray[n6++] = n2;
                if (bl) return;
                if (~(3 & n8) == -4) {
                    n2 = nArray2[0xFF & n5 >> 418609192];
                    n5 += n3;
                }
                ++n8;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(z[3]).append(n2).append(',').append(n3).append(',').append(n4).append(',').append(nArray != null ? z[0] : z[1]).append(',');
            if (nArray2 != null) {
                string = z[0];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(')').toString());
            }
            string = z[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(n5).append(',').append(n6).append(',').append(n7).append(')').toString());
        }
    }

    static final byte[] a(byte[] byArray, int n2, int n3, String string, byte[] byArray2) {
        boolean bl = client.vh;
        try {
            int n4;
            int n5;
            int n6;
            int n7;
            block20: {
                ++k;
                n7 = (0xFF & byArray[1]) + (byArray[0] & 0xFF) * 256;
                string = string.toUpperCase();
                n6 = 0;
                n5 = 0;
                while (~string.length() < ~n5) {
                    n4 = 61 * n6 + (string.charAt(n5) - 32);
                    if (!bl) {
                        n6 = n4;
                        ++n5;
                        if (!bl) continue;
                    }
                    break block20;
                }
                n4 = 74 % ((n2 - -74) / 49);
            }
            int n8 = n4;
            n5 = 2 + 10 * n7;
            for (n8 = 0; n8 < n7; ++n8) {
                int n9;
                block22: {
                    byte[] byArray3;
                    block25: {
                        block24: {
                            int n10;
                            block23: {
                                block21: {
                                    int n11 = (0xFF & byArray[5 + n8 * 10]) + ((0xFF & byArray[n8 * 10 + 4]) * 256 + (byArray[3 + 10 * n8] & 0xFF) * 65536 + (0xFF & byArray[2 + 10 * n8]) * 0x1000000);
                                    n10 = (byArray[10 * n8 + 7] & 0xFF) * 256 + ((byArray[n8 * 10 + 6] & 0xFF) * 65536 + (0xFF & byArray[n8 * 10 + 8]));
                                    n9 = (byArray[10 + 10 * n8] & 0xFF) * 256 + 65536 * (0xFF & byArray[n8 * 10 + 9]) + (byArray[10 * n8 + 11] & 0xFF);
                                    if (n11 == n6) break block21;
                                    break block22;
                                }
                                if (null == byArray2) {
                                    byArray2 = new byte[n3 + n10];
                                }
                                if (n9 == n10) break block23;
                                ea.a(byArray2, n10, byArray, n9, n5);
                                if (!bl) break block24;
                            }
                            for (int i2 = 0; i2 < n10; ++i2) {
                                byArray3 = byArray2;
                                if (!bl) {
                                    byArray3[i2] = byArray[i2 + n5];
                                    if (!bl) continue;
                                    break;
                                }
                                break block25;
                            }
                        }
                        byArray3 = byArray2;
                    }
                    return byArray3;
                }
                n5 += n9;
                if (!bl) continue;
            }
            return null;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[4]);
            String string2 = byArray != null ? z[0] : z[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string2).append(',').append(n2).append(',').append(n3).append(',');
            String string3 = string != null ? z[0] : z[1];
            StringBuilder stringBuilder3 = stringBuilder2.append(string3).append(',');
            String string4 = byArray2 != null ? z[0] : z[1];
            throw i.a(runtimeException2, stringBuilder3.append(string4).append(')').toString());
        }
    }

    static final int a(int n2) {
        try {
            if (n2 != 65525) {
                h = null;
            }
            ++c;
            int n3 = d.a(ka.b, (byte)100, b.v);
            ka.b += 2;
            return n3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[5] + n2 + ')');
        }
    }

    final void a(String string, int n2, int n3, int n4, int n5, String string2, int n6, int n7, String string3, String string4, int n8, String string5) {
        try {
            ++f;
            this.p = string;
            this.l = n2;
            this.m = n4;
            if (n6 <= 69) {
                return;
            }
            this.d = n7;
            this.e = n3;
            this.o = string3;
            this.i = n5;
            this.j = n8;
            this.b = string5;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[2]);
            String string6 = string != null ? z[0] : z[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string6).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(',').append(n5).append(',');
            String string7 = string2 != null ? z[0] : z[1];
            StringBuilder stringBuilder3 = stringBuilder2.append(string7).append(',').append(n6).append(',').append(n7).append(',');
            String string8 = string3 != null ? z[0] : z[1];
            StringBuilder stringBuilder4 = stringBuilder3.append(string8).append(',');
            String string9 = string4 != null ? z[0] : z[1];
            StringBuilder stringBuilder5 = stringBuilder4.append(string9).append(',').append(n8).append(',');
            String string10 = string5 != null ? z[0] : z[1];
            throw i.a(runtimeException2, stringBuilder5.append(string10).append(')').toString());
        }
    }

    t() {
    }

    static {
        z = new String[]{t.z(t.z("9}_\u0015\t")), t.z(t.z(",&\u001dW")), t.z(t.z("6}3\u0013")), t.z(t.z("6}2\u0013")), t.z(t.z("6}5\u0013")), t.z(t.z("6}0\u0013"))};
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x74);
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
                        n4 = 66;
                        break;
                    }
                    case 1: {
                        n4 = 83;
                        break;
                    }
                    case 2: {
                        n4 = 113;
                        break;
                    }
                    case 3: {
                        n4 = 59;
                        break;
                    }
                    default: {
                        n4 = 116;
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

