/*
 * Decompiled with CFR 0.152.
 */
import java.io.EOFException;
import java.io.File;
import java.io.IOException;
import java.io.RandomAccessFile;

final class d {
    static int n;
    static int e;
    static int l;
    static int k;
    static int d;
    static int i;
    static int f;
    static int[] g;
    static int a;
    private RandomAccessFile j;
    private long c;
    static e h;
    private long b;
    static String[] m;
    private static final String[] z;

    private final void a(int n2) throws IOException {
        try {
            ++n;
            if (null != this.j) {
                this.j.close();
                this.j = null;
            }
            if (n2 != 25291) {
                d.a(62, (byte)14, null);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[9] + n2 + ')');
        }
    }

    final long a(byte by) throws IOException {
        try {
            if (by != 47) {
                this.c = -52L;
            }
            ++i;
            return this.j.length();
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[8] + by + ')');
        }
    }

    protected final void finalize() throws Throwable {
        try {
            ++a;
            if (null != this.j) {
                System.out.println("");
                this.a(25291);
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[6]);
        }
    }

    static int a(int n2, int n3) {
        try {
            return n2 | n3;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[3] + n2 + ',' + n3 + ')');
        }
    }

    final void b(byte[] byArray, int n2, int n3, int n4) throws IOException {
        try {
            ++d;
            if ((this.b ^ 0xFFFFFFFFFFFFFFFFL) > (this.c + (long)n2 ^ 0xFFFFFFFFFFFFFFFFL)) {
                this.j.seek(this.b);
                this.j.write(1);
                throw new EOFException();
            }
            this.j.write(byArray, n4, n2);
            this.c += (long)n2;
            if (n3 != 1) {
                d.a(63, (byte)-101, null);
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[7]);
            String string = byArray != null ? z[0] : z[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(')').toString());
        }
    }

    static final int a(int n2, byte by, byte[] byArray) {
        try {
            if (by < 4) {
                h = null;
            }
            ++f;
            return (byArray[1 + n2] & 0xFF) + ((0xFF & byArray[n2]) << 922410888);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[1]).append(n2).append(',').append(by).append(',');
            String string = byArray != null ? z[0] : z[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    final void a(int n2, long l2) throws IOException {
        if (n2 != 0) {
            return;
        }
        try {
            this.j.seek(l2);
            ++k;
            this.c = l2;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[4] + n2 + ',' + l2 + ')');
        }
    }

    final int a(byte[] byArray, int n2, int n3, int n4) throws IOException {
        try {
            int n5;
            block7: {
                block6: {
                    ++e;
                    n5 = this.j.read(byArray, n3, n2);
                    if (~n5 < n4) break block6;
                    break block7;
                }
                this.c += (long)n5;
            }
            return n5;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[10]);
            String string = byArray != null ? z[0] : z[2];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(',').append(n4).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    d(File file, String string, long l2) throws IOException {
        try {
            if (0L == (l2 ^ 0xFFFFFFFFFFFFFFFFL)) {
                l2 = Long.MAX_VALUE;
            }
            if ((l2 ^ 0xFFFFFFFFFFFFFFFFL) > (file.length() ^ 0xFFFFFFFFFFFFFFFFL)) {
                file.delete();
            }
            this.j = new RandomAccessFile(file, string);
            this.b = l2;
            this.c = 0L;
            int n2 = this.j.read();
            if (0 != ~n2 && !string.equals("r")) {
                this.j.seek(0L);
                this.j.write(n2);
            }
            this.j.seek(0L);
            return;
        }
        catch (RuntimeException runtimeException) {
            String string2;
            StringBuilder stringBuilder = new StringBuilder().append(z[5]).append(file != null ? z[0] : z[2]).append(',');
            if (string != null) {
                string2 = z[0];
                throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(l2).append(')').toString());
            }
            string2 = z[2];
            throw i.a(runtimeException, stringBuilder.append(string2).append(',').append(l2).append(')').toString());
        }
    }

    static {
        z = new String[]{d.z(d.z("\u0003\u001aP L")), d.z(d.z("\u001c\u001a:&")), d.z(d.z("\u0016A\u0012b")), d.z(d.z("\u001c\u001a<&")), d.z(d.z("\u001c\u001a8&")), d.z(d.z("\u001c\u001aBg_\u0011@@&")), d.z(d.z("\u001c\u001a\u0018g_\u0019X\u0017tTP\u001d")), d.z(d.z("\u001c\u001a9&")), d.z(d.z("\u001c\u001a?&")), d.z(d.z("\u001c\u001a;&")), d.z(d.z("\u001c\u001a=&"))};
        l = 0;
        m = new String[]{d.z(d.z("(X\u001boB\u001d\u0014\u001b`E\u001dF^zY\u001d\u0014\u0010{\\\u001aQ\f.^\u001e\u0014\u0017zT\u0015G^z^XC\u0017zY\u001cF\u001fy")), d.z(d.z("\u0019Z\u001a.A\nQ\r}\u0011\u001dZ\nkC"))};
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x31);
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
                        n4 = 120;
                        break;
                    }
                    case 1: {
                        n4 = 52;
                        break;
                    }
                    case 2: {
                        n4 = 126;
                        break;
                    }
                    case 3: {
                        n4 = 14;
                        break;
                    }
                    default: {
                        n4 = 49;
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

