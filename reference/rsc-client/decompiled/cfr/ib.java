/*
 * Decompiled with CFR 0.152.
 */
import java.io.IOException;
import java.net.URL;

class ib {
    ib a;
    static URL c;
    static int[] d;
    ib e;
    static int b;
    static int f;
    private static final String[] z;

    static int a(int n2, int n3) {
        try {
            return n2 & n3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[8] + n2 + ',' + n3 + ')');
        }
    }

    final void a(int n2) {
        try {
            ++f;
            if (n2 != -27331) {
                return;
            }
            if (null == this.e) {
                return;
            }
            this.e.a = this.a;
            this.a.e = this.e;
            this.a = null;
            this.e = null;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[7] + n2 + ')');
        }
    }

    protected ib() {
    }

    static final byte[] a(int n2, String string, int n3, int n4) throws IOException {
        boolean bl = client.vh;
        ++b;
        if (null != la.g[n4]) {
            return la.g[n4];
        }
        if (n2 > -73) {
            return null;
        }
        try {
            block31: {
                Object object;
                block28: {
                    block30: {
                        block29: {
                            block27: {
                                nb.q = n3;
                                o.l = string;
                                if (m.e != null) break block27;
                                break block28;
                            }
                            object = m.e.a(9395, n4);
                            if (null != object) break block29;
                            break block28;
                        }
                        if (mb.a(object, ((byte[])object).length, 0) == tb.l[n4]) break block30;
                        break block28;
                    }
                    la.g[n4] = k.a(128, true, object);
                    return la.g[n4];
                }
                object = new URL(c, z[5] + n4 + "_" + Long.toHexString(tb.l[n4]));
                byte[] byArray = null;
                int n5 = 0;
                while (~n5 > -4) {
                    block32: {
                        byArray = da.a((URL)object, true, true);
                        if (bl) break block31;
                        if (~mb.a(byArray, byArray.length, 0) != ~tb.l[n4] && !bl) break block32;
                        try {
                            if (null != m.e) {
                                m.e.a(n4, byArray.length, -97, byArray);
                            }
                            la.g[n4] = k.a(128, true, byArray);
                            return la.g[n4];
                        }
                        catch (IOException iOException) {
                            block33: {
                                if (~n5 == -3) break block33;
                                break block32;
                            }
                            throw iOException;
                        }
                    }
                    ++n5;
                    if (!bl) continue;
                }
                if (null != byArray) {
                    StringBuilder stringBuilder = new StringBuilder(z[4] + n4 + z[3] + tb.l[n4]);
                    stringBuilder.append(z[2] + byArray.length);
                    int n6 = 0;
                    while (~byArray.length < ~n6) {
                        if (-6 >= ~n6) break;
                        stringBuilder.append(" " + byArray[n6]);
                        ++n6;
                        if (!bl) continue;
                        break;
                    }
                    throw new IOException(stringBuilder.toString());
                }
            }
            throw new IOException(z[4] + n4 + z[3] + tb.l[n4]);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[0]).append(n2).append(',');
            String string2 = string != null ? z[1] : z[6];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(',').append(n3).append(',').append(n4).append(')').toString());
        }
    }

    static {
        z = new String[]{ib.z(ib.z("8\u00108\u001d#y")), ib.z(ib.z("*\\8|\u001f")), ib.z(ib.z("q\u001es<_")), ib.z(ib.z("kRu \u0001l")), ib.z(ib.z("\u0012\u001dc>\u0006?Ubr\u0006>\u0005x>\r0\u001664\u000b=\u00176q")), ib.z(ib.z("2\u001dx&\u0007?\u0006")), ib.z(ib.z("?\u0007z>")), ib.z(ib.z("8\u00108\u0002#y")), ib.z(ib.z("8\u00108\u0003#y"))};
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x62);
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
                        n4 = 81;
                        break;
                    }
                    case 1: {
                        n4 = 114;
                        break;
                    }
                    case 2: {
                        n4 = 22;
                        break;
                    }
                    case 3: {
                        n4 = 82;
                        break;
                    }
                    default: {
                        n4 = 98;
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

