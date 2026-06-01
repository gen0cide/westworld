/*
 * Decompiled with CFR 0.152.
 */
final class eb
implements Runnable {
    volatile boolean a = false;
    volatile sa[] f = new sa[2];
    static int h;
    c g;
    volatile boolean i = false;
    static i e;
    static v d;
    static int[] b;
    private static char[] c;
    private static final String z;

    /*
     * WARNING - Removed try catching itself - possible behaviour change.
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    @Override
    public final void run() {
        boolean bl = client.vh;
        try {
            this.i = true;
            ++h;
            try {
                block9: do {
                    if (this.a) return;
                    if (bl) return;
                    int n2 = 0;
                    while (~n2 > -3) {
                        sa sa2 = this.f[n2];
                        if (bl) continue block9;
                        if (sa2 != null) {
                            sa2.a();
                        }
                        ++n2;
                        if (!bl) continue;
                    }
                    mb.a(11200, 10L);
                    ba.a(null, 1, this.g);
                } while (!bl);
                return;
            }
            catch (Exception exception) {
                mb.a(0x1FFFFF, exception, null);
                return;
            }
            finally {
                this.i = false;
            }
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z);
        }
    }

    eb() {
    }

    static {
        z = eb.z(eb.z("X[F~}S\u0011A"));
        e = new i(eb.z(eb.z("qp>I")), 0);
        d = new v(eb.z(eb.z("jm!")), eb.z(eb.z("R_\u000eekX")), eb.z(eb.z("bN\u001ce")), 5);
        c = new char[256];
        for (int i2 = 0; 256 > i2; ++i2) {
            eb.c[i2] = (char)i2;
        }
        eb.c[45] = 45;
        eb.c[59] = 59;
        eb.c[42] = 42;
        eb.c[124] = 124;
        eb.c[43] = 43;
        eb.c[33] = 33;
        eb.c[34] = 34;
        eb.c[47] = 47;
        eb.c[46] = 46;
        eb.c[61] = 61;
        eb.c[92] = 92;
        eb.c[44] = 44;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 8);
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
                        n4 = 61;
                        break;
                    }
                    case 1: {
                        n4 = 57;
                        break;
                    }
                    case 2: {
                        n4 = 104;
                        break;
                    }
                    case 3: {
                        n4 = 12;
                        break;
                    }
                    default: {
                        n4 = 8;
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

