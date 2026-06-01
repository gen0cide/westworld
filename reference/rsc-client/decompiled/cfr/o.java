/*
 * Decompiled with CFR 0.152.
 */
final class o {
    static int b;
    static int j;
    static int o;
    static int h;
    private int i;
    static int[] a;
    private int d;
    static String[] g;
    private int[] n;
    private int[] k;
    static int[] p;
    static int c;
    private int m;
    static int e;
    private int f;
    static String l;
    private static final String[] z;

    final int c(int n2) {
        try {
            if (-1 == ~this.i--) {
                this.b(-110);
                this.i = 255;
            }
            ++b;
            if (n2 > -67) {
                this.d = 32;
            }
            return this.k[this.i];
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[7] + n2 + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    private final void b(int n2) {
        boolean bl = client.vh;
        try {
            ++h;
            this.d += ++this.f;
            if (n2 > -100) {
                g = null;
            }
            int n3 = 0;
            do {
                int n4;
                int n5;
                block16: {
                    block17: {
                        int n6;
                        block18: {
                            block15: {
                                if (-257 >= ~n3) return;
                                n5 = this.n[n3];
                                n6 = n3 & 3;
                                if (bl) return;
                                if (n6 != 0) break block15;
                                this.m ^= this.m << -402254995;
                                if (!bl) break block16;
                            }
                            if (-2 == ~n6) break block17;
                            if (-3 != ~n6) break block18;
                            this.m ^= this.m << 2019129250;
                            if (!bl) break block16;
                        }
                        if (n6 != 3) break block16;
                        this.m ^= this.m >>> -350927312;
                        if (!bl) break block16;
                    }
                    this.m ^= this.m >>> -585344026;
                }
                this.m += this.n[0xFF & 128 + n3];
                this.n[n3] = n4 = this.m + this.n[ib.a(1020, n5) >> -1542190526] - -this.d;
                this.k[n3] = this.d = this.n[ib.a(255, n4 >> 725943080 >> -16506142)] + n5;
                ++n3;
            } while (!bl);
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[6] + n2 + ')');
        }
    }

    private final void a(int n2) {
        block19: {
            boolean bl = client.vh;
            try {
                int n3;
                int n4;
                int n5;
                int n6;
                int n7;
                int n8;
                int n9;
                int n10;
                int n11;
                block18: {
                    block17: {
                        ++c;
                        n11 = -1640531527;
                        n10 = -1640531527;
                        n9 = -1640531527;
                        n8 = -1640531527;
                        n7 = -1640531527;
                        if (n2 != -2) {
                            this.a(-15);
                        }
                        n6 = -1640531527;
                        n5 = -1640531527;
                        n4 = -1640531527;
                        n3 = 0;
                        while (~n3 > -5) {
                            n4 += (n10 ^= n6 << -1205206741);
                            n6 += n11;
                            n7 += (n6 ^= n11 >>> -890549438);
                            n11 += n4;
                            n11 ^= n4 << 2101945832;
                            n4 += n7;
                            n4 ^= n7 >>> 581272400;
                            n7 += (n5 += n11);
                            n9 += (n7 ^= n5 << -1718537238);
                            n5 += (n8 += n4);
                            n5 ^= n8 >>> 1908708324;
                            n8 += n9;
                            n8 ^= n9 << 297382696;
                            n9 += (n10 += n5);
                            n9 ^= n10 >>> 1751843753;
                            n10 += (n6 += n8);
                            n11 += n9;
                            ++n3;
                            if (!bl) {
                                if (!bl) continue;
                                break;
                            }
                            break block17;
                        }
                        n3 = 0;
                    }
                    while (-257 < ~n3) {
                        n10 += this.k[n3];
                        n7 += this.k[4 + n3];
                        n9 += this.k[n3 + 7];
                        n5 += this.k[n3 - -5];
                        n4 += this.k[3 + n3];
                        n8 += this.k[6 + n3];
                        n6 += (n11 += this.k[2 + n3]);
                        n6 ^= n11 >>> 186681986;
                        n11 += (n4 += (n10 ^= (n6 += this.k[n3 + 1]) << -536269493));
                        n11 ^= n4 << -1065768376;
                        n4 += (n7 += n6);
                        n8 += (n4 ^= n7 >>> 21869104);
                        n7 += (n5 += n11);
                        n9 += (n7 ^= n5 << -1346760726);
                        n5 += n8;
                        n10 += (n5 ^= n8 >>> -949961308);
                        n8 += n9;
                        n6 += (n8 ^= n9 << 1673244968);
                        n9 += n10;
                        n11 += (n9 ^= n10 >>> 409602633);
                        n10 += n6;
                        this.n[n3] = n10;
                        this.n[n3 - -1] = n6;
                        this.n[2 + n3] = n11;
                        this.n[3 + n3] = n4;
                        this.n[4 + n3] = n7;
                        this.n[n3 - -5] = n5;
                        this.n[n3 - -6] = n8;
                        this.n[n3 - -7] = n9;
                        n3 += 8;
                        if (!bl) {
                            if (!bl) continue;
                            break;
                        }
                        break block18;
                    }
                    n3 = 0;
                }
                while (~n3 > -257) {
                    n8 += this.n[6 + n3];
                    n6 += this.n[n3 + 1];
                    n9 += this.n[7 + n3];
                    n7 += this.n[4 + n3];
                    n5 += this.n[5 + n3];
                    n4 += this.n[n3 + 3];
                    n10 += this.n[n3];
                    n10 ^= n6 << -1294322773;
                    n6 += (n11 += this.n[n3 - -2]);
                    n6 ^= n11 >>> -117514910;
                    n11 += (n4 += n10);
                    n5 += (n11 ^= n4 << -1087924184);
                    n4 += (n7 += n6);
                    n4 ^= n7 >>> 1052530928;
                    n7 += n5;
                    n7 ^= n5 << -1448878102;
                    n5 += (n8 += n4);
                    n5 ^= n8 >>> 819293412;
                    n8 += (n9 += n7);
                    n8 ^= n9 << -1355920056;
                    n9 += (n10 += n5);
                    n9 ^= n10 >>> -2145912983;
                    n10 += (n6 += n8);
                    n11 += n9;
                    this.n[n3] = n10;
                    this.n[n3 - -1] = n6;
                    this.n[n3 - -2] = n11;
                    this.n[3 + n3] = n4;
                    this.n[n3 + 4] = n7;
                    this.n[n3 + 5] = n5;
                    this.n[n3 - -6] = n8;
                    this.n[7 + n3] = n9;
                    n3 += 8;
                    if (!bl) {
                        if (!bl) continue;
                        break;
                    }
                    break block19;
                }
                this.b(-105);
                this.i = 256;
            }
            catch (RuntimeException runtimeException) {
                throw i.a(runtimeException, z[2] + n2 + ')');
            }
        }
    }

    static final int a(int n2, int n3, int n4, int n5) {
        try {
            if (n3 != 9570) {
                o.a((byte)56);
            }
            ++e;
            return (n2 << 905616656) + ((n5 << 81348360) - -n4);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[1] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ')');
        }
    }

    static final String a(byte by) {
        boolean bl = client.vh;
        try {
            String string;
            block7: {
                ++o;
                if (by != 38) {
                    o.a(67, 106, -48, 111);
                }
                string = "";
                while (-1 != ~kb.d[jb.p]) {
                    string = string + (char)kb.d[jb.p++];
                    if (!bl) {
                        if (!bl) continue;
                        break;
                    }
                    break block7;
                }
                ++jb.p;
            }
            return string;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[0] + by + ')');
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final int a(int n2, int n3, int n4, int n5, boolean bl, int n6, int n7, int n8) {
        boolean bl2 = client.vh;
        try {
            int n9;
            int n10;
            int n11;
            block12: {
                ++j;
                n11 = 0;
                int n12 = 0;
                while (~n12 > ~n6) {
                    int n13;
                    int n14;
                    block14: {
                        block13: {
                            n10 = n8;
                            n9 = n3;
                            if (bl2) break block12;
                            n14 = n10 * (n9 + ((!bl ? -n12 : n12) - n7));
                            if (n14 < -100) break block13;
                            if (-101 <= ~n14) break block14;
                            n14 = 100;
                            if (!bl2) break block14;
                        }
                        n14 = -100;
                    }
                    if (-11 < ~(n13 = n4 - -n14)) {
                        n13 = 10;
                    }
                    n11 += n2 * n13 / 100;
                    ++n12;
                    if (!bl2) continue;
                }
                n10 = n5;
                n9 = -30910;
            }
            if (n10 != n9) {
                o.a((byte)106);
            }
            return n11;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[8] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + bl + ',' + n6 + ',' + n7 + ',' + n8 + ')');
        }
    }

    o(int[] nArray) {
        block8: {
            boolean bl = client.vh;
            try {
                this.k = new int[256];
                this.n = new int[256];
                int n2 = 0;
                while (~nArray.length < ~n2) {
                    this.k[n2] = nArray[n2];
                    ++n2;
                    if (!bl) {
                        if (!bl) continue;
                        break;
                    }
                    break block8;
                }
                this.a(-2);
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(z[5]);
                String string = nArray != null ? z[3] : z[4];
                throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
            }
        }
    }

    static {
        z = new String[]{o.z(o.z("SeY;")), o.z(o.z("SeZ;")), o.z(o.z("Se^;")), o.z(o.z("Ge5=\u000e")), o.z(o.z("R>w\u007f")), o.z(o.z("Se'z\u001dU?%;")), o.z(o.z("Se_;")), o.z(o.z("SeX;")), o.z(o.z("Se];"))};
        g = new String[]{o.z(o.z("}9~3\nS>;`\u0006N.;j\u001cIklz\u0000Tko|SO rcSH#~3\u0007I?ta\u001a]'")), o.z(o.z("]%\u007f3\u0007Y'~c\u001cN?;g\u001c\u001c\u0007n~\u0011N\"\u007ft\u0016\u0003"))};
        l = "";
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x73);
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
                        n4 = 60;
                        break;
                    }
                    case 1: {
                        n4 = 75;
                        break;
                    }
                    case 2: {
                        n4 = 27;
                        break;
                    }
                    case 3: {
                        n4 = 19;
                        break;
                    }
                    default: {
                        n4 = 115;
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

