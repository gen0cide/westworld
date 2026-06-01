/*
 * Decompiled with CFR 0.152.
 */
final class ra
extends va {
    private db n = new db();
    private db l = new db();
    private int m = 0;
    private int k = -1;

    @Override
    final va b() {
        return (va)this.n.a((byte)35);
    }

    final void a(vb vb2, int n2, int n3) {
        this.a(sb.a(vb2, n2, n3));
    }

    private final void e() {
        if (this.m > 0) {
            hb hb2 = (hb)this.l.a((byte)-87);
            while (hb2 != null) {
                hb2.g -= this.m;
                hb2 = (hb)this.l.b((byte)111);
            }
            this.k -= this.m;
            this.m = 0;
        }
    }

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     */
    @Override
    final synchronized void b(int[] nArray, int n2, int n3) {
        do {
            hb hb2;
            if (this.k < 0) {
                this.c(nArray, n2, n3);
                return;
            }
            if (this.m + n3 < this.k) {
                this.m += n3;
                this.c(nArray, n2, n3);
                return;
            }
            int n4 = this.k - this.m;
            this.c(nArray, n2, n4);
            n2 += n4;
            n3 -= n4;
            this.m += n4;
            this.e();
            hb hb3 = hb2 = (hb)this.l.a((byte)106);
            synchronized (hb3) {
                int n5 = hb2.a(this);
                if (n5 < 0) {
                    hb2.g = 0;
                    this.a(hb2);
                } else {
                    hb2.g = n5;
                    this.a(hb2.a, hb2);
                }
            }
        } while (n3 != 0);
    }

    private final void a(ib ib2, hb hb2) {
        while (ib2 != this.l.k && ((hb)ib2).g <= hb2.g) {
            ib2 = ib2.a;
        }
        ac.a(hb2, (byte)34, ib2);
        this.k = ((hb)this.l.k.a).g;
    }

    @Override
    final int d() {
        return 0;
    }

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     */
    @Override
    final synchronized void b(int n2) {
        do {
            hb hb2;
            if (this.k < 0) {
                this.c(n2);
                return;
            }
            if (this.m + n2 < this.k) {
                this.m += n2;
                this.c(n2);
                return;
            }
            int n3 = this.k - this.m;
            this.c(n3);
            n2 -= n3;
            this.m += n3;
            this.e();
            hb hb3 = hb2 = (hb)this.l.a((byte)-80);
            synchronized (hb3) {
                int n4 = hb2.a(this);
                if (n4 < 0) {
                    hb2.g = 0;
                    this.a(hb2);
                } else {
                    hb2.g = n4;
                    this.a(hb2.a, hb2);
                }
            }
        } while (n2 != 0);
    }

    private final void c(int[] nArray, int n2, int n3) {
        va va2 = (va)this.n.a((byte)-88);
        while (va2 != null) {
            va2.a(nArray, n2, n3);
            va2 = (va)this.n.b((byte)122);
        }
    }

    @Override
    final va a() {
        return (va)this.n.b((byte)115);
    }

    private final synchronized void a(va va2) {
        this.n.a(va2, false);
    }

    private final void a(hb hb2) {
        hb2.a(-27331);
        hb2.a();
        ib ib2 = this.l.k.a;
        this.k = ib2 == this.l.k ? -1 : ((hb)ib2).g;
    }

    private final void c(int n2) {
        va va2 = (va)this.n.a((byte)-123);
        while (va2 != null) {
            va2.b(n2);
            va2 = (va)this.n.b((byte)80);
        }
    }
}

