/*
 * Decompiled with CFR 0.152.
 */
import java.awt.Component;
import java.awt.Image;
import java.awt.image.IndexColorModel;

final class pa {
    static c b;
    static int[] g;
    static int[] f;
    private static byte[] e;
    static int[] a;
    static c k;
    static int i;
    static int c;
    static long h;
    static int[] j;
    static int[] d;
    private static final String[] z;

    static final Image a(int n2, Component component, byte[] byArray) {
        try {
            int n3;
            k.o = byArray[12] + byArray[13] * 256;
            da.bb = byArray[14] + 256 * byArray[15];
            ++i;
            byte[] byArray2 = new byte[256];
            byte[] byArray3 = new byte[256];
            byte[] byArray4 = new byte[256];
            int n4 = 0;
            while (-257 < ~n4) {
                byArray2[n4] = byArray[20 + n4 * 3];
                byArray3[n4] = byArray[3 * n4 + 19];
                byArray4[n4] = byArray[3 * n4 + 18];
                ++n4;
            }
            m.d = new IndexColorModel(8, 256, byArray2, byArray3, byArray4);
            byte[] byArray5 = new byte[k.o * da.bb];
            int n5 = 0;
            int n6 = -1 + da.bb;
            while (-1 >= ~n6) {
                for (n3 = 0; k.o > n3; ++n3) {
                    byArray5[n5++] = byArray[k.o * n6 + 786 + n3];
                }
                --n6;
            }
            Image image = component.createImage(da.db);
            n3 = 113 / ((n2 - 10) / 53);
            lb.a(true, byArray5);
            component.prepareImage(image, da.db);
            lb.a(true, byArray5);
            component.prepareImage(image, da.db);
            lb.a(true, byArray5);
            component.prepareImage(image, da.db);
            return image;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[0]).append(n2).append(',');
            String string = component != null ? z[2] : z[1];
            StringBuilder stringBuilder2 = stringBuilder.append(string).append(',');
            String string2 = byArray != null ? z[2] : z[1];
            throw i.a(runtimeException2, stringBuilder2.append(string2).append(')').toString());
        }
    }

    static final byte[] a(int n2) {
        try {
            ++c;
            byte[] byArray = new byte[256];
            int n3 = -128;
            if (n2 > -125) {
                pa.a(-7, null, null);
            }
            while (n3 < 127) {
                int n4;
                block9: {
                    block8: {
                        int n5 = n3;
                        int n6 = (n5 ^= 0xFFFFFFFF) & 0x80;
                        int n7 = n5 >> -1437861628 & 7;
                        int n8 = 0xF & n5;
                        n8 |= 0x10;
                        n8 = 1 + (n8 << -1014531647);
                        n4 = n8 << n7 + 2;
                        n4 = -132 + n4;
                        if (~n6 != -1) break block8;
                        break block9;
                    }
                    n4 = -n4;
                }
                byArray[ib.a((int)n3, (int)255)] = (byte)(n4 / 256);
                ++n3;
            }
            return byArray;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[3] + n2 + ')');
        }
    }

    static {
        int n2;
        z = new String[]{pa.z(pa.z("J_\f\t%")), pa.z(pa.z("TKN'")), pa.z(pa.z("A\u0010\fep")), pa.z(pa.z("J_\f\n%"))};
        e = new byte[64];
        g = new int[100];
        a = new int[512];
        j = new int[2048];
        d = new int[256];
        for (n2 = 0; 256 > n2; ++n2) {
            pa.a[n2] = (int)(32768.0 * Math.sin(0.02454369 * (double)n2));
            pa.a[256 + n2] = (int)(32768.0 * Math.cos((double)n2 * 0.02454369));
        }
        n2 = 0;
        while (-1025 < ~n2) {
            pa.j[n2] = (int)(Math.sin((double)n2 * 0.00613592315) * 32768.0);
            pa.j[n2 - -1024] = (int)(Math.cos((double)n2 * 0.00613592315) * 32768.0);
            ++n2;
        }
        n2 = 0;
        while (-11 < ~n2) {
            pa.e[n2] = (byte)(48 + n2);
            ++n2;
        }
        n2 = 0;
        while (~n2 > -27) {
            pa.e[n2 - -10] = (byte)(n2 + 65);
            ++n2;
        }
        for (n2 = 0; 26 > n2; ++n2) {
            pa.e[n2 - -36] = (byte)(97 + n2);
        }
        pa.e[63] = 36;
        pa.e[62] = -93;
        n2 = 0;
        while (-11 < ~n2) {
            pa.d[n2 + 48] = n2;
            ++n2;
        }
        n2 = 0;
        while (-27 < ~n2) {
            pa.d[n2 + 65] = n2 - -10;
            ++n2;
        }
        for (n2 = 0; n2 < 26; ++n2) {
            pa.d[n2 + 97] = 36 + n2;
        }
        pa.d[36] = 63;
        pa.d[163] = 62;
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
                        n4 = 58;
                        break;
                    }
                    case 1: {
                        n4 = 62;
                        break;
                    }
                    case 2: {
                        n4 = 34;
                        break;
                    }
                    case 3: {
                        n4 = 75;
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

