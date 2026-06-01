/*
 * Decompiled with CFR 0.152.
 */
import java.io.DataInputStream;
import java.io.EOFException;
import java.io.IOException;
import java.io.InputStream;

final class ta {
    int t;
    static v f;
    int y;
    int s = -1;
    int[] m = new int[12];
    int[] F = new int[10];
    int D;
    String n;
    int p;
    int q;
    int z = 0;
    String C;
    int b;
    int x;
    int[] k = new int[10];
    int j;
    static int l;
    static String[] r;
    int u = 0;
    int A;
    int h = 0;
    int J = 0;
    static int v;
    int H;
    int e;
    int i;
    static int g;
    int o;
    String c;
    int I = 0;
    int B = 0;
    int a = 0;
    int w = 0;
    int G = 0;
    int E = 0;
    int K;
    int d = 0;
    private static final String[] L;

    static final void a(String string, int n2, byte[] byArray, int n3) throws IOException {
        try {
            ++v;
            if (n2 != -19675) {
                f = null;
            }
            InputStream inputStream = nb.a(true, string);
            DataInputStream dataInputStream = new DataInputStream(inputStream);
            try {
                dataInputStream.readFully(byArray, 0, n3);
            }
            catch (EOFException eOFException) {
                // empty catch block
            }
            dataInputStream.close();
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(L[1]);
            String string2 = string != null ? L[3] : L[2];
            StringBuilder stringBuilder2 = stringBuilder.append(string2).append(',').append(n2).append(',');
            String string3 = byArray != null ? L[3] : L[2];
            throw i.a(runtimeException2, stringBuilder2.append(string3).append(',').append(n3).append(')').toString());
        }
    }

    static final int a(int n2, byte by) {
        try {
            int n3 = 110 / ((65 - by) / 44);
            ++l;
            n2 = (0x55555555 & n2 >>> -2016269343) + (0x55555555 & n2);
            n2 = ((n2 & 0xCCCCCCCC) >>> 209493378) + (0x33333333 & n2);
            n2 = n2 - -(n2 >>> 1603286532) & 0xF0F0F0F;
            n2 += n2 >>> -843073400;
            n2 += n2 >>> -761392816;
            return 0xFF & n2;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, L[0] + n2 + ',' + by + ')');
        }
    }

    ta() {
    }

    static {
        L = new String[]{ta.z(ta.z("?>o\u0018\u0015")), ta.z(ta.z("?>o\u001b\u0015")), ta.z(ta.z("%*-6")), ta.z(ta.z("0qot@"))};
        f = new v(ta.z(ta.z("\u001c\u000b\u0016\u0013m")), ta.z(ta.z("$9'3^.")), ta.z(ta.z("\u0014((*")), 3);
        g = 176;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x3D);
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
                        n4 = 75;
                        break;
                    }
                    case 1: {
                        n4 = 95;
                        break;
                    }
                    case 2: {
                        n4 = 65;
                        break;
                    }
                    case 3: {
                        n4 = 90;
                        break;
                    }
                    default: {
                        n4 = 61;
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

