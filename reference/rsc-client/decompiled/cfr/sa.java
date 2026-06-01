/*
 * Decompiled with CFR 0.152.
 */
import java.awt.Component;

class sa {
    static int t;
    private static eb n;
    private long u = p.a(0);
    int[] j;
    private static int o;
    static boolean i;
    private boolean b = false;
    private int r = 32;
    private va a;
    private int g;
    private int q;
    private int c = 0;
    private int k;
    private va[] s = new va[8];
    private long e = 0L;
    private int p = 0;
    private int m = 0;
    private va[] d = new va[8];
    private long f = 0L;
    private int h = 0;
    private boolean l = true;

    int b() throws Exception {
        return this.k;
    }

    void b(int n2) throws Exception {
    }

    void c() throws Exception {
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final synchronized void d() {
        if (n != null) {
            boolean bl = true;
            for (int n2 = 0; n2 < 2; ++n2) {
                if (sa.n.f[n2] == this) {
                    sa.n.f[n2] = null;
                }
                if (sa.n.f[n2] == null) continue;
                bl = false;
            }
            if (bl) {
                sa.n.a = true;
                while (sa.n.i) {
                    mb.a(11200, 50L);
                }
                n = null;
            }
        }
        this.e();
        this.j = null;
        this.b = true;
    }

    void e() {
    }

    private final void a(va va2, int n2) {
        block3: {
            block2: {
                int n3 = n2 >> 5;
                va va3 = this.s[n3];
                if (va3 != null) break block2;
                this.d[n3] = va2;
                break block3;
            }
            va3.j = va2;
        }
        this.s[n3] = va2;
        va2.i = n2;
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final sa a(c c2, Component component, int n2, int n3) {
        if (t == 0) {
            throw new IllegalStateException();
        }
        if (n2 < 0 || n2 >= 2) {
            throw new IllegalArgumentException();
        }
        if (n3 < 256) {
            n3 = 256;
        }
        try {
            pb pb2 = new pb();
            pb2.j = new int[256 * (i ? 2 : 1)];
            pb2.q = n3;
            ((sa)pb2).a(component);
            pb2.k = (n3 & 0xFFFFFC00) + 1024;
            if (pb2.k > 16384) {
                pb2.k = 16384;
            }
            ((sa)pb2).b(pb2.k);
            if (o > 0 && n == null) {
                n = new eb();
                sa.n.g = c2;
                c2.a(true, n, o);
            }
            if (n != null) {
                if (sa.n.f[n2] != null) {
                    throw new IllegalArgumentException();
                }
                sa.n.f[n2] = pb2;
            }
            return pb2;
        }
        catch (Throwable throwable) {
            return new sa();
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    final synchronized void a() {
        if (this.b) {
            return;
        }
        long l2 = p.a(0);
        try {
            if (l2 > this.u + 6000L) {
                this.u = l2 - 6000L;
            }
            while (l2 > this.u + 5000L) {
                this.a(256);
                this.u += (long)(256000 / t);
                l2 = p.a(0);
            }
        }
        catch (Exception exception) {
            this.u = l2;
        }
        if (this.j == null) {
            return;
        }
        try {
            int n3;
            int n2;
            if (this.e != 0L) {
                if (l2 < this.e) {
                    return;
                }
                this.b(this.k);
                this.e = 0L;
                this.l = true;
            }
            if (this.h - (n2 = this.b()) > this.m) {
                this.m = this.h - n2;
            }
            if ((n3 = this.q + this.g) + 256 > 16384) {
                n3 = 16128;
            }
            if (n3 + 256 > this.k) {
                this.k += 1024;
                if (this.k > 16384) {
                    this.k = 16384;
                }
                this.e();
                this.b(this.k);
                n2 = 0;
                this.l = true;
                if (n3 + 256 > this.k) {
                    n3 = this.k - 256;
                    this.g = n3 - this.q;
                }
            }
            while (n2 < n3) {
                this.a(this.j, 256);
                this.c();
                n2 += 256;
            }
            if (l2 > this.f) {
                if (!this.l) {
                    if (this.m == 0 && this.c == 0) {
                        this.e();
                        this.e = l2 + 2000L;
                        return;
                    }
                    this.g = Math.min(this.c, this.m);
                    this.c = this.m;
                } else {
                    this.l = false;
                }
                this.m = 0;
                this.f = l2 + 2000L;
            }
            this.h = n2;
            return;
        }
        catch (Exception exception) {
            this.e();
            this.e = l2 + 2000L;
        }
    }

    final synchronized void a(va va2) {
        this.a = va2;
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final void a(int n2, boolean bl, int n3) {
        if (n2 >= 8000 && n2 <= 48000) {
            t = n2;
            i = bl;
            o = n3;
            return;
        }
        throw new IllegalArgumentException();
    }

    private final void a(int[] nArray, int n2) {
        block36: {
            va va2;
            int n3;
            int n4 = n2;
            if (i) {
                n4 <<= 1;
            }
            ab.a(nArray, 0, n4);
            this.p -= n2;
            if (this.a == null || this.p > 0) break block36;
            this.p += t >> 4;
            sa.b(this.a);
            this.a(this.a, this.a.c());
            int n5 = 0;
            int n6 = 255;
            int n7 = 7;
            block22: while (true) {
                int n8;
                int n9;
                block43: {
                    block37: {
                        if (n6 == 0) break;
                        if (n7 >= 0) break block37;
                        n9 = n7 & 3;
                        n8 = -(n7 >> 2);
                        break block43;
                    }
                    n9 = n7;
                    n8 = 0;
                }
                n3 = n6 >>> n9 & 0x11111111;
                while (true) {
                    block39: {
                        block38: {
                            if (n3 == 0) {
                                break;
                            }
                            if ((n3 & 1) != 0) break block38;
                            break block39;
                        }
                        n6 &= ~(1 << n9);
                        va2 = null;
                        va va3 = this.d[n9];
                        while (va3 != null) {
                            va va4;
                            block42: {
                                block41: {
                                    bb bb2;
                                    block40: {
                                        bb2 = va3.h;
                                        if (bb2 == null || bb2.g <= n8) break block40;
                                        n6 |= 1 << n9;
                                        va2 = va3;
                                        va3 = va3.j;
                                        continue;
                                    }
                                    va3.g = true;
                                    int n10 = va3.d();
                                    n5 += n10;
                                    if (bb2 != null) {
                                        bb2.g += n10;
                                    }
                                    if (n5 >= this.r) {
                                        break block22;
                                    }
                                    va va5 = va3.b();
                                    if (va5 != null) {
                                        int n11 = va3.i;
                                        while (va5 != null) {
                                            this.a(va5, n11 * va5.c() >> 8);
                                            va5 = va3.a();
                                        }
                                    }
                                    va4 = va3.j;
                                    va3.j = null;
                                    if (va2 != null) break block41;
                                    this.d[n9] = va4;
                                    break block42;
                                }
                                va2.j = va4;
                            }
                            if (va4 == null) {
                                this.s[n9] = va2;
                            }
                            va3 = va4;
                        }
                    }
                    n9 += 4;
                    ++n8;
                    n3 >>>= 4;
                }
                --n7;
            }
            for (n7 = 0; n7 < 8; ++n7) {
                va va6 = this.d[n7];
                va[] vaArray = this.d;
                n3 = n7;
                this.s[n7] = null;
                vaArray[n3] = null;
                while (va6 != null) {
                    va2 = va6.j;
                    va6.j = null;
                    va6 = va2;
                }
            }
        }
        if (this.p < 0) {
            this.p = 0;
        }
        if (this.a != null) {
            this.a.b(nArray, 0, n2);
        }
        this.u = p.a(0);
    }

    private final void a(int n2) {
        this.p -= n2;
        if (this.p < 0) {
            this.p = 0;
        }
        if (this.a != null) {
            this.a.b(n2);
        }
    }

    private static final void b(va va2) {
        va2.g = false;
        if (va2.h != null) {
            va2.h.g = 0;
        }
        va va3 = va2.b();
        while (va3 != null) {
            sa.b(va3);
            va3 = va2.a();
        }
    }

    void a(Component component) throws Exception {
    }

    sa() {
    }
}

