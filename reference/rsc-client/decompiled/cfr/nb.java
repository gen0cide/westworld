/*
 * Decompiled with CFR 0.152.
 */
import java.io.BufferedInputStream;
import java.io.EOFException;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.URL;

final class nb {
    private byte[] n;
    static int m;
    private long h = -1L;
    private long l;
    static int g;
    static int p;
    private byte[] a;
    private long x;
    static int e;
    private d b;
    static int[] d;
    private int t;
    static int k;
    private int i = 0;
    static int y;
    static int r;
    static String[] u;
    static int z;
    static URL s;
    static int w;
    static char[] f;
    private long o;
    private long A;
    private long j = -1L;
    static int c;
    static int q;
    static int v;
    private static final String[] B;

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void a(int n2) throws IOException {
        boolean bl = client.vh;
        try {
            block22: {
                long l2;
                long l3;
                block26: {
                    block25: {
                        block24: {
                            block23: {
                                if (this.h == -1L) break block22;
                                if ((this.x ^ 0xFFFFFFFFFFFFFFFFL) != (this.h ^ 0xFFFFFFFFFFFFFFFFL)) {
                                    this.b.a(0, this.h);
                                    this.x = this.h;
                                }
                                this.b.b(this.a, this.i, 1, 0);
                                this.x += (long)this.i;
                                if ((this.A ^ 0xFFFFFFFFFFFFFFFFL) > (this.x ^ 0xFFFFFFFFFFFFFFFFL)) {
                                    this.A = this.x;
                                }
                                l3 = -1L;
                                l2 = -1L;
                                if (this.h >= this.j && (this.h ^ 0xFFFFFFFFFFFFFFFFL) > (this.j + (long)this.t ^ 0xFFFFFFFFFFFFFFFFL)) break block23;
                                if ((this.j ^ 0xFFFFFFFFFFFFFFFFL) > (this.h ^ 0xFFFFFFFFFFFFFFFFL) || this.j >= (long)this.i + this.h) break block24;
                                l3 = this.j;
                                if (!bl) break block24;
                            }
                            l3 = this.h;
                        }
                        if ((this.j ^ 0xFFFFFFFFFFFFFFFFL) <= (this.h + (long)this.i ^ 0xFFFFFFFFFFFFFFFFL) || ((long)this.i + this.h ^ 0xFFFFFFFFFFFFFFFFL) < (this.j + (long)this.t ^ 0xFFFFFFFFFFFFFFFFL)) break block25;
                        l2 = this.h + (long)this.i;
                        if (!bl) break block26;
                    }
                    if ((long)this.t + this.j > this.h && this.j - -((long)this.t) <= (long)this.i + this.h) {
                        l2 = this.j - -((long)this.t);
                    }
                }
                if (-1L < l3 && (l3 ^ 0xFFFFFFFFFFFFFFFFL) > (l2 ^ 0xFFFFFFFFFFFFFFFFL)) {
                    int n3 = (int)(l2 + -l3);
                    ab.a(this.a, (int)(l3 + -this.h), this.n, (int)(-this.j + l3), n3);
                }
                this.i = 0;
                this.h = -1L;
            }
            ++v;
            if (n2 == -14779) return;
            this.n = null;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, B[9] + n2 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final void a(boolean bl, int n2, int n3, byte[] byArray) throws IOException {
        boolean bl2 = client.vh;
        try {
            int n4;
            int n5;
            block35: {
                block36: {
                    ++z;
                    try {
                        long l2;
                        long l3;
                        int n6;
                        long l4;
                        block42: {
                            block41: {
                                int n7;
                                block40: {
                                    block39: {
                                        int n8;
                                        block38: {
                                            block37: {
                                                if (~(n2 + n3) < ~byArray.length) {
                                                    throw new ArrayIndexOutOfBoundsException(n3 - -n2 - byArray.length);
                                                }
                                                if (this.h != -1L && this.h <= this.l && this.h - -((long)this.i) >= (long)n2 + this.l) {
                                                    ab.a(this.a, (int)(this.l - this.h), byArray, n3, n2);
                                                    this.l += (long)n2;
                                                    return;
                                                }
                                                l4 = this.l;
                                                n6 = n3;
                                                n7 = n2;
                                                if ((this.l ^ 0xFFFFFFFFFFFFFFFFL) <= (this.j ^ 0xFFFFFFFFFFFFFFFFL) && (this.j + (long)this.t ^ 0xFFFFFFFFFFFFFFFFL) < (this.l ^ 0xFFFFFFFFFFFFFFFFL)) {
                                                    n8 = (int)(this.j + (-this.l + (long)this.t));
                                                    if (~n2 > ~n8) {
                                                        n8 = n2;
                                                    }
                                                    ab.a(this.n, (int)(this.l - this.j), byArray, n3, n8);
                                                    n2 -= n8;
                                                    n3 += n8;
                                                    this.l += (long)n8;
                                                }
                                                if (!bl) {
                                                    return;
                                                }
                                                if (n2 > this.n.length) break block37;
                                                if (0 >= n2) break block38;
                                                this.b((byte)34);
                                                n8 = n2;
                                                if (~this.t > ~n8) {
                                                    n8 = this.t;
                                                }
                                                ab.a(this.n, 0, byArray, n3, n8);
                                                n3 += n8;
                                                n2 -= n8;
                                                this.l += (long)n8;
                                                if (!bl2) break block38;
                                            }
                                            this.b.a(0, this.l);
                                            this.x = this.l;
                                            while (-1 > ~n2) {
                                                n8 = this.b.a(byArray, n2, n3, -1);
                                                n5 = 0;
                                                n4 = ~n8;
                                                if (!bl2) {
                                                    if (n5 == n4 && !bl2) break;
                                                    n3 += n8;
                                                    n2 -= n8;
                                                    this.l += (long)n8;
                                                    this.x += (long)n8;
                                                    if (!bl2) continue;
                                                }
                                                break block35;
                                            }
                                        }
                                        if (this.h == -1L) break block36;
                                        if ((this.l ^ 0xFFFFFFFFFFFFFFFFL) > (this.h ^ 0xFFFFFFFFFFFFFFFFL) && 0 < n2) {
                                            n8 = (int)(this.h - this.l) + n3;
                                            if (n8 > n3 - -n2) {
                                                n8 = n3 - -n2;
                                            }
                                            while (~n8 < ~n3) {
                                                byArray[n3++] = 0;
                                                --n2;
                                                ++this.l;
                                                if (!bl2) {
                                                    if (!bl2) continue;
                                                }
                                                break block36;
                                            }
                                        }
                                        l3 = -1L;
                                        if ((l4 ^ 0xFFFFFFFFFFFFFFFFL) < (this.h ^ 0xFFFFFFFFFFFFFFFFL) || (l4 + (long)n7 ^ 0xFFFFFFFFFFFFFFFFL) >= (this.h ^ 0xFFFFFFFFFFFFFFFFL)) break block39;
                                        l3 = this.h;
                                        if (!bl2) break block40;
                                    }
                                    if ((l4 ^ 0xFFFFFFFFFFFFFFFFL) <= (this.h ^ 0xFFFFFFFFFFFFFFFFL) && l4 < this.h + (long)this.i) {
                                        l3 = l4;
                                    }
                                }
                                l2 = -1L;
                                if (l4 < this.h + (long)this.i && this.h - -((long)this.i) <= l4 + (long)n7) break block41;
                                if (l4 - -((long)n7) <= this.h || ((long)this.i + this.h ^ 0xFFFFFFFFFFFFFFFFL) > ((long)n7 + l4 ^ 0xFFFFFFFFFFFFFFFFL)) break block42;
                                l2 = l4 + (long)n7;
                                if (!bl2) break block42;
                            }
                            l2 = this.h + (long)this.i;
                        }
                        if (l3 <= -1L || (l3 ^ 0xFFFFFFFFFFFFFFFFL) <= (l2 ^ 0xFFFFFFFFFFFFFFFFL)) break block36;
                        int n9 = (int)(l2 + -l3);
                        ab.a(this.a, (int)(-this.h + l3), byArray, n6 - -((int)(l3 + -l4)), n9);
                        if ((this.l ^ 0xFFFFFFFFFFFFFFFFL) > (l2 ^ 0xFFFFFFFFFFFFFFFFL)) {
                            n2 = (int)((long)n2 - (l2 - this.l));
                            this.l = l2;
                        }
                    }
                    catch (IOException iOException) {
                        this.x = -1L;
                        throw iOException;
                    }
                }
                n5 = 0;
                n4 = n2;
            }
            if (n5 < n4) throw new EOFException();
            return;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(B[6]).append(bl).append(',').append(n2).append(',').append(n3).append(',');
            if (byArray != null) {
                string = B[2];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = B[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    final void a(byte by, byte[] byArray) throws IOException {
        try {
            int n2 = -12 % ((by - -22) / 54);
            this.a(true, byArray.length, 0, byArray);
            ++p;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(B[4]).append(by).append(',');
            String string = byArray != null ? B[2] : B[1];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    final long a(byte by) {
        try {
            if (by != -111) {
                d = null;
            }
            ++r;
            return this.o;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, B[0] + by + ')');
        }
    }

    static final int a(int n2, byte by) {
        try {
            ++e;
            if (n2 != 255) {
                nb.a(-35, (byte)126);
            }
            return by & 0xFF;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, B[8] + n2 + ',' + by + ')');
        }
    }

    static final InputStream a(boolean bl, String string) throws IOException {
        if (!bl) {
            return null;
        }
        try {
            InputStream inputStream;
            block12: {
                block11: {
                    ++c;
                    if (null == s) break block11;
                    URL uRL = new URL(s, string);
                    inputStream = uRL.openStream();
                    if (!client.vh) break block12;
                }
                inputStream = new BufferedInputStream(new FileInputStream(string));
            }
            return inputStream;
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(B[5]).append(bl).append(',');
            String string2 = string != null ? B[2] : B[1];
            throw i.a(runtimeException2, stringBuilder.append(string2).append(')').toString());
        }
    }

    final void a(int n2, int n3, int n4, byte[] byArray) throws IOException {
        boolean bl = client.vh;
        ++w;
        if (n3 > -80) {
            return;
        }
        try {
            try {
                block29: {
                    block30: {
                        long l2;
                        long l3;
                        block35: {
                            block34: {
                                block33: {
                                    block32: {
                                        if (this.l + (long)n2 > this.o) {
                                            this.o = (long)n2 + this.l;
                                        }
                                        if (this.h != -1L && ((this.l ^ 0xFFFFFFFFFFFFFFFFL) > (this.h ^ 0xFFFFFFFFFFFFFFFFL) || (long)this.i + this.h < this.l)) {
                                            this.a(-14779);
                                        }
                                        if (0L != (this.h ^ 0xFFFFFFFFFFFFFFFFL) && (long)n2 + this.l > (long)this.a.length + this.h) {
                                            int n5 = (int)(-this.l - -this.h + (long)this.a.length);
                                            ab.a(byArray, n4, this.a, (int)(-this.h + this.l), n5);
                                            n2 -= n5;
                                            n4 += n5;
                                            this.l += (long)n5;
                                            this.i = this.a.length;
                                            this.a(-14779);
                                        }
                                        if (~n2 >= ~this.a.length) break block29;
                                        if ((this.x ^ 0xFFFFFFFFFFFFFFFFL) != (this.l ^ 0xFFFFFFFFFFFFFFFFL)) {
                                            this.b.a(0, this.l);
                                            this.x = this.l;
                                        }
                                        this.b.b(byArray, n2, 1, n4);
                                        this.x += (long)n2;
                                        if (this.x > this.A) {
                                            this.A = this.x;
                                        }
                                        l3 = -1L;
                                        l2 = -1L;
                                        if (this.j > this.l || this.l >= this.j + (long)this.t) break block32;
                                        l3 = this.l;
                                        if (!bl) break block33;
                                    }
                                    if (this.l <= this.j && this.l - -((long)n2) > this.j) {
                                        l3 = this.j;
                                    }
                                }
                                if (this.j >= this.l - -((long)n2) || (long)this.t + this.j < this.l - -((long)n2)) break block34;
                                l2 = this.l + (long)n2;
                                if (!bl) break block35;
                            }
                            if ((this.l ^ 0xFFFFFFFFFFFFFFFFL) > ((long)this.t + this.j ^ 0xFFFFFFFFFFFFFFFFL) && (this.j - -((long)this.t) ^ 0xFFFFFFFFFFFFFFFFL) >= (this.l - -((long)n2) ^ 0xFFFFFFFFFFFFFFFFL)) {
                                l2 = this.j - -((long)this.t);
                            }
                        }
                        if (0L <= (l3 ^ 0xFFFFFFFFFFFFFFFFL) || (l2 ^ 0xFFFFFFFFFFFFFFFFL) >= (l3 ^ 0xFFFFFFFFFFFFFFFFL)) break block30;
                        int n6 = (int)(l2 - l3);
                        ab.a(byArray, (int)(-this.l + l3 + (long)n4), this.n, (int)(-this.j + l3), n6);
                    }
                    this.l += (long)n2;
                    return;
                }
                if (n2 > 0) {
                    if (this.h == -1L) {
                        this.h = this.l;
                    }
                    ab.a(byArray, n4, this.a, (int)(-this.h + this.l), n2);
                    this.l += (long)n2;
                    if (((long)this.i ^ 0xFFFFFFFFFFFFFFFFL) > (this.l + -this.h ^ 0xFFFFFFFFFFFFFFFFL)) {
                        this.i = (int)(this.l + -this.h);
                    }
                    return;
                }
            }
            catch (IOException iOException) {
                this.x = -1L;
                throw iOException;
            }
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(B[10]).append(n2).append(',').append(n3).append(',').append(n4).append(',');
            String string = byArray != null ? B[2] : B[1];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    final void a(long l2, int n2) throws IOException {
        try {
            ++m;
            if ((l2 ^ 0xFFFFFFFFFFFFFFFFL) > -1L) {
                throw new IOException();
            }
            this.l = l2;
            int n3 = -39 / ((n2 - -66) / 55);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, B[11] + l2 + ',' + n2 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void b(byte n2) throws IOException {
        boolean bl = client.vh;
        try {
            int n3;
            int n4;
            block10: {
                this.t = 0;
                ++k;
                if ((this.l ^ 0xFFFFFFFFFFFFFFFFL) != (this.x ^ 0xFFFFFFFFFFFFFFFFL)) {
                    this.b.a(0, this.l);
                    this.x = this.l;
                }
                this.j = this.l;
                while (~this.t > ~this.n.length) {
                    int n5 = -this.t + this.n.length;
                    n4 = ~n5;
                    n3 = -200000001;
                    if (!bl) {
                        int n6;
                        if (n4 < n3) {
                            n5 = 200000000;
                        }
                        if (0 == ~(n6 = this.b.a(this.n, n5, this.t, -1)) && !bl) break;
                        this.x += (long)n6;
                        this.t += n6;
                        if (!bl) continue;
                    }
                    break block10;
                }
                n4 = n2;
                n3 = 14;
            }
            if (n4 > n3) return;
            this.b = null;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, B[7] + n2 + ')');
        }
    }

    nb(d d2, int n2, int n3) throws IOException {
        try {
            this.b = d2;
            this.o = this.A = d2.a((byte)47);
            this.a = new byte[n3];
            this.l = 0L;
            this.n = new byte[n2];
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(B[3]);
            String string = d2 != null ? B[2] : B[1];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(n2).append(',').append(n3).append(')').toString());
        }
    }

    static {
        B = new String[]{nb.z(nb.z("M]4s\u0002")), nb.z(nb.z("MJvV")), nb.z(nb.z("X\u00114\u0014W")), nb.z(nb.z("M]4\u0006CMVn\u0004\u0002")), nb.z(nb.z("M]4~\u0002")), nb.z(nb.z("M]4|\u0002")), nb.z(nb.z("M]4r\u0002")), nb.z(nb.z("M]4x\u0002")), nb.z(nb.z("M]4}\u0002")), nb.z(nb.z("M]4{\u0002")), nb.z(nb.z("M]4\u007f\u0002")), nb.z(nb.z("M]4y\u0002"))};
        g = 0;
        s = null;
        u = new String[]{nb.z(nb.z("wFj_\nWW\u007f\u001aDVRx_X\u0003P|\u001aCWZwI\nWP:IOOS:[DG\u001fjHOPL:_DWZh"))};
        f = new char[]{'\u20ac', '\u0000', '\u201a', '\u0192', '\u201e', '\u2026', '\u2020', '\u2021', '\u02c6', '\u2030', '\u0160', '\u2039', '\u0152', '\u0000', '\u017d', '\u0000', '\u0000', '\u2018', '\u2019', '\u201c', '\u201d', '\u2022', '\u2013', '\u2014', '\u02dc', '\u2122', '\u0161', '\u203a', '\u0153', '\u0000', '\u017e', '\u0178'};
        q = 0;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x2A);
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
                        n4 = 35;
                        break;
                    }
                    case 1: {
                        n4 = 63;
                        break;
                    }
                    case 2: {
                        n4 = 26;
                        break;
                    }
                    case 3: {
                        n4 = 58;
                        break;
                    }
                    default: {
                        n4 = 42;
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

