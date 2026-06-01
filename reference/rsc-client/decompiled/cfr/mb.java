/*
 * Decompiled with CFR 0.152.
 */
import java.io.DataInputStream;
import java.net.URL;

final class mb {
    static int c;
    private static int b;
    static int[] i;
    static int[] a;
    static int d;
    static int e;
    static int[] k;
    static int l;
    static String[] g;
    static int h;
    static int f;
    static int j;
    private static final String[] z;

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final synchronized byte[] a(int n2, byte by) {
        boolean bl = client.vh;
        try {
            int n3;
            ++j;
            if (n2 == 100 && ~n.b < -1) {
                byte[] byArray = ob.j[--n.b];
                ob.j[n.b] = null;
                return byArray;
            }
            if (5000 == n2 && ~s.d < -1) {
                byte[] byArray = e.kb[--s.d];
                e.kb[s.d] = null;
                return byArray;
            }
            if (by > -97) {
                return null;
            }
            if (-30001 == ~n2 && ~b < -1) {
                byte[] byArray = ca.tb[--b];
                ca.tb[mb.b] = null;
                return byArray;
            }
            if (null != t.n) {
                int n4 = 0;
                while (~e.wb.length < ~n4) {
                    n3 = ~e.wb[n4];
                    if (bl) return new byte[n3];
                    if (n3 == ~n2 && 0 < v.g[n4]) {
                        int n5 = n4;
                        int n6 = v.g[n5] - 1;
                        v.g[n5] = n6;
                        byte[] byArray = t.n[n4][n6];
                        t.n[n4][v.g[n4]] = null;
                        return byArray;
                    }
                    ++n4;
                    if (!bl) continue;
                }
            } else {
                n3 = n2;
            }
            return new byte[n3];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[1] + n2 + ',' + by + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final void a(int n2, long l2) {
        try {
            ++h;
            if (n2 != 11200) {
                g = null;
            }
            if (0L >= l2) return;
            if (l2 % 10L == 0L) {
                u.a(n2 + -11200, -1L + l2);
                u.a(0, 1L);
                if (!client.vh) return;
            }
            u.a(0, l2);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[0] + n2 + ',' + l2 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final String a(int n2, int n3) {
        boolean bl = client.vh;
        try {
            String string;
            block10: {
                ++f;
                string = "" + n2;
                int n4 = -3 + string.length();
                while (-1 > ~n4) {
                    string = string.substring(0, n4) + "," + string.substring(n4);
                    n4 -= 3;
                    if (!bl) {
                        if (!bl) continue;
                    }
                    break block10;
                }
                if (n3 != 131071) {
                    mb.a(null, -74, 53);
                }
            }
            if (-9 > ~string.length()) {
                string = z[24] + string.substring(0, string.length() + -8) + z[22] + string + ")";
                if (!bl) return string;
            }
            if (~string.length() < -5) return z[21] + string.substring(0, -4 + string.length()) + z[20] + string + ")";
            return string;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[23] + n2 + ',' + n3 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final String a(String string, String string2, boolean bl, int n2) {
        boolean bl2 = client.vh;
        try {
            int n3;
            ++d;
            if (!bl) {
                l = 90;
            }
            if (0 != (n3 = n2) || bl2) {
                if (n3 == 1) {
                    if (string2 == null) return string;
                    if (string2.length() != 0) return string2 + z[6] + string;
                    return string;
                }
                if (n3 == 2 && !bl2) {
                    if (null == string2) return string;
                    if (-1 == ~string2.length()) return string;
                    return z[5] + string2 + z[7] + string;
                }
                if (n3 == 3 && !bl2) {
                    if (null == string2) return string;
                    if (~string2.length() == -1) return string;
                    return string2 + z[7] + string;
                }
                if (~n3 == -5 && !bl2) {
                    if (null == string2) return string;
                    if (~string2.length() != -1) return string2 + z[7] + string;
                    return string;
                }
                if (~n3 == -6) return string;
                if (6 == n3) {
                    if (!bl2) return string2 + z[8];
                }
                if (-8 != ~n3) return "";
                if (!bl2) {
                    if (null == string2) return string;
                    if (~string2.length() != -1) return string2 + z[7] + string;
                    return string;
                }
            }
            if (string2 == null) return string;
            if (-1 == ~string2.length()) return string;
            return string2 + z[7] + string;
        }
        catch (RuntimeException runtimeException) {
            String string3;
            StringBuilder stringBuilder = new StringBuilder().append(z[9]).append(string != null ? z[4] : z[3]).append(',');
            if (string2 != null) {
                string3 = z[4];
                throw i.a(runtimeException, stringBuilder.append(string3).append(',').append(bl).append(',').append(n2).append(')').toString());
            }
            string3 = z[3];
            throw i.a(runtimeException, stringBuilder.append(string3).append(',').append(bl).append(',').append(n2).append(')').toString());
        }
    }

    static final void a(int n2, Throwable throwable, String string) {
        block22: {
            boolean bl = client.vh;
            ++e;
            try {
                String string2;
                block21: {
                    block20: {
                        block18: {
                            block19: {
                                block17: {
                                    block16: {
                                        string2 = "";
                                        if (n2 != 0x1FFFFF) {
                                            mb.a(null, null, true, 27);
                                        }
                                        if (throwable != null) break block16;
                                        break block17;
                                    }
                                    string2 = gb.a(false, throwable);
                                }
                                if (null == string) break block18;
                                if (throwable == null) break block19;
                                string2 = string2 + z[10];
                            }
                            string2 = string2 + string;
                        }
                        n.a((byte)-93, string2);
                        string2 = jb.a(true, z[15], ":", string2);
                        string2 = jb.a(true, z[12], "@", string2);
                        string2 = jb.a(true, z[16], "&", string2);
                        string2 = jb.a(true, z[13], "#", string2);
                        if (null == l.b) break block20;
                        break block21;
                    }
                    return;
                }
                g g2 = pa.b.a((byte)74, new URL(l.b.getCodeBase(), z[11] + db.d + z[19] + (ka.a != null ? ka.a : "" + pa.h) + z[17] + c.q + z[18] + c.k + z[14] + string2));
                while (~g2.b == -1) {
                    mb.a(11200, 1L);
                    if (!bl) {
                        if (!bl) continue;
                    }
                    break block22;
                }
                if (-2 == ~g2.b) {
                    DataInputStream dataInputStream = (DataInputStream)g2.d;
                    dataInputStream.read();
                    dataInputStream.close();
                }
            }
            catch (Exception exception) {}
        }
    }

    static final int a(byte[] byArray, int n2, int n3) {
        try {
            if (n3 != 0) {
                return 6;
            }
            ++c;
            return w.a(n2, -49, byArray, 0);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[2]);
            String string = byArray != null ? z[4] : z[3];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(')').toString());
        }
    }

    static {
        z = new String[]{mb.z(mb.z("U<HS\u001e")), mb.z(mb.z("U<HV\u001e")), mb.z(mb.z("U<HP\u001e")), mb.z(mb.z("V+\ny")), mb.z(mb.z("CpH;K")), mb.z(mb.z("a1\u00135B]2\n5")), mb.z(mb.z("\u0018*\u0003yZK~\u001fzC\u0002~")), mb.z(mb.z("\u0002~")), mb.z(mb.z("\u0018)\u000ff^]-FaY\u0018*\u0014tR]~\u0011|BP~\u001fzC\u0016")), mb.z(mb.z("U<HQ\u001e")), mb.z(mb.z("\u0018\"F")), mb.z(mb.z("[2\u000fpXL;\u0014gYJp\u0011f\t[c")), mb.z(mb.z("\u001djV")), mb.z(mb.z("\u001dlU")), mb.z(mb.z("\u001e;[")), mb.z(mb.z("\u001dm\u0007")), mb.z(mb.z("\u001dlP")), mb.z(mb.z("\u001e(W(")), mb.z(mb.z("\u001e(T(")), mb.z(mb.z("\u001e+[")), mb.z(mb.z("s~&b^Q\u001eN")), mb.z(mb.z("x=\u001ftv")), mb.z(mb.z("\u00183\u000fyZQ1\b5vO6\u000fU\u001e")), mb.z(mb.z("U<HW\u001e")), mb.z(mb.z("x9\u0014pv"))};
        i = new int[]{0, 1, 3, 7, 15, 31, 63, 127, 255, 511, 1023, 2047, 4095, 8191, 16383, Short.MAX_VALUE, 65535, 131071, 262143, 524287, 1048575, 0x1FFFFF, 0x3FFFFF, 0x7FFFFF, 0xFFFFFF, 0x1FFFFFF, 0x3FFFFFF, 0x7FFFFFF, 0xFFFFFFF, 0x1FFFFFFF, 0x3FFFFFFF, Integer.MAX_VALUE, -1};
        b = 0;
        l = 0;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x36);
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
                        n4 = 56;
                        break;
                    }
                    case 1: {
                        n4 = 94;
                        break;
                    }
                    case 2: {
                        n4 = 102;
                        break;
                    }
                    case 3: {
                        n4 = 21;
                        break;
                    }
                    default: {
                        n4 = 54;
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

