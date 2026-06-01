/*
 * Decompiled with CFR 0.152.
 */
import java.awt.image.ImageConsumer;

final class u {
    static int e;
    static int f;
    static int[] a;
    static ImageConsumer d;
    static int c;
    static String[] b;
    static int g;
    private static final String[] z;

    static final int a(int n2, tb tb2, String string) {
        try {
            if (n2 <= 10) {
                a = null;
            }
            ++f;
            int n3 = tb2.w;
            byte[] byArray = h.a(string, (byte)30);
            tb2.b(byArray.length, (byte)-88);
            tb2.w += fb.a.a(byArray.length, tb2.F, tb2.w, byArray, 0, 119);
            return tb2.w - n3;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[3]).append(n2).append(',');
            String string2 = tb2 != null ? z[1] : z[2];
            StringBuilder stringBuilder2 = stringBuilder.append(string2).append(',');
            String string3 = string != null ? z[1] : z[2];
            throw i.a(runtimeException2, stringBuilder2.append(string3).append(')').toString());
        }
    }

    static final i a(boolean n2, int n3) {
        boolean bl = client.vh;
        try {
            int n4;
            block8: {
                ++c;
                i[] iArray = gb.a(69);
                for (int i2 = 0; iArray.length > i2; ++i2) {
                    i i3 = iArray[i2];
                    n4 = i3.a;
                    if (bl) break block8;
                    if (n4 != n3) continue;
                    return i3;
                }
                n4 = n2;
            }
            if (n4 != 0) {
                g = -2;
            }
            return null;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[0] + (n2 != 0) + ',' + n3 + ')');
        }
    }

    static final void a(int n2, long l2) {
        try {
            try {
                if (n2 != 0) {
                    return;
                }
                Thread.sleep(l2);
            }
            catch (InterruptedException interruptedException) {
                // empty catch block
            }
            ++e;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[4] + n2 + ',' + l2 + ')');
        }
    }

    static {
        z = new String[]{u.z(u.z("\u0005#{\u001a")), u.z(u.z("\u000b#\u0016\u001c/")), u.z(u.z("\u001exT^")), u.z(u.z("\u0005#z\u001a")), u.z(u.z("\u0005#y\u001a"))};
        g = 0;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x52);
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
                        n4 = 112;
                        break;
                    }
                    case 1: {
                        n4 = 13;
                        break;
                    }
                    case 2: {
                        n4 = 56;
                        break;
                    }
                    case 3: {
                        n4 = 50;
                        break;
                    }
                    default: {
                        n4 = 82;
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

