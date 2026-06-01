/*
 * Decompiled with CFR 0.152.
 */
import java.awt.image.ColorModel;
import java.io.IOException;
import java.net.Socket;

abstract class m {
    static byte[][] b;
    static int[] g;
    static int k;
    String h;
    static ob e;
    static int[] i;
    int f;
    static int a;
    static ColorModel d;
    static int c;
    static int j;
    private static final String[] z;

    static final int a(boolean bl, int n2, byte[] byArray) {
        try {
            if (!bl) {
                b = null;
            }
            ++a;
            return (byArray[n2 + 3] & 0xFF) + ((byArray[n2 - -2] << -1719048248 & 0xFF00) + ((0xFF & byArray[n2]) << -523457256)) - -((0xFF & byArray[1 + n2]) << -728162096);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[2]).append(bl).append(',').append(n2).append(',');
            String string = byArray != null ? z[1] : z[0];
            throw i.a(runtimeException2, stringBuilder.append(string).append(')').toString());
        }
    }

    final Socket a(boolean bl) throws IOException {
        try {
            ++k;
            if (bl) {
                i = null;
            }
            return new Socket(this.h, this.f);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, z[8] + bl + ')');
        }
    }

    /*
     * Unable to fully structure code
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final void a(byte[] var0, byte var1_1, boolean var2_2) {
        var6_3 = client.vh;
        try {
            block353: {
                block352: {
                    block351: {
                        block349: {
                            block347: {
                                block346: {
                                    block345: {
                                        block344: {
                                            block343: {
                                                block342: {
                                                    block341: {
                                                        block340: {
                                                            block339: {
                                                                block338: {
                                                                    block337: {
                                                                        block336: {
                                                                            block335: {
                                                                                block334: {
                                                                                    block333: {
                                                                                        block332: {
                                                                                            block331: {
                                                                                                block330: {
                                                                                                    block329: {
                                                                                                        block328: {
                                                                                                            block327: {
                                                                                                                block326: {
                                                                                                                    block325: {
                                                                                                                        block324: {
                                                                                                                            block323: {
                                                                                                                                block322: {
                                                                                                                                    block321: {
                                                                                                                                        block320: {
                                                                                                                                            block319: {
                                                                                                                                                block318: {
                                                                                                                                                    block317: {
                                                                                                                                                        block316: {
                                                                                                                                                            block315: {
                                                                                                                                                                block314: {
                                                                                                                                                                    block313: {
                                                                                                                                                                        block312: {
                                                                                                                                                                            block311: {
                                                                                                                                                                                block310: {
                                                                                                                                                                                    block309: {
                                                                                                                                                                                        block308: {
                                                                                                                                                                                            block307: {
                                                                                                                                                                                                block306: {
                                                                                                                                                                                                    block305: {
                                                                                                                                                                                                        block304: {
                                                                                                                                                                                                            block303: {
                                                                                                                                                                                                                block302: {
                                                                                                                                                                                                                    block301: {
                                                                                                                                                                                                                        block300: {
                                                                                                                                                                                                                            block299: {
                                                                                                                                                                                                                                block298: {
                                                                                                                                                                                                                                    block297: {
                                                                                                                                                                                                                                        block296: {
                                                                                                                                                                                                                                            block295: {
                                                                                                                                                                                                                                                block294: {
                                                                                                                                                                                                                                                    block293: {
                                                                                                                                                                                                                                                        block292: {
                                                                                                                                                                                                                                                            block291: {
                                                                                                                                                                                                                                                                block290: {
                                                                                                                                                                                                                                                                    block289: {
                                                                                                                                                                                                                                                                        block288: {
                                                                                                                                                                                                                                                                            block287: {
                                                                                                                                                                                                                                                                                block286: {
                                                                                                                                                                                                                                                                                    block285: {
                                                                                                                                                                                                                                                                                        block284: {
                                                                                                                                                                                                                                                                                            block283: {
                                                                                                                                                                                                                                                                                                block282: {
                                                                                                                                                                                                                                                                                                    kb.d = na.a(m.z[7], 0, var0, -124);
                                                                                                                                                                                                                                                                                                    ++m.c;
                                                                                                                                                                                                                                                                                                    jb.p = 0;
                                                                                                                                                                                                                                                                                                    b.v = na.a(m.z[4], 0, var0, -125);
                                                                                                                                                                                                                                                                                                    ka.b = 0;
                                                                                                                                                                                                                                                                                                    gb.p = t.a(65525);
                                                                                                                                                                                                                                                                                                    fa.e = new int[gb.p];
                                                                                                                                                                                                                                                                                                    ac.x = new String[gb.p];
                                                                                                                                                                                                                                                                                                    ka.c = new int[gb.p];
                                                                                                                                                                                                                                                                                                    h.c = new int[gb.p];
                                                                                                                                                                                                                                                                                                    lb.ac = new String[gb.p];
                                                                                                                                                                                                                                                                                                    ua.Bb = new int[gb.p];
                                                                                                                                                                                                                                                                                                    kb.b = new int[gb.p];
                                                                                                                                                                                                                                                                                                    mb.k = new int[gb.p];
                                                                                                                                                                                                                                                                                                    kb.c = new int[gb.p];
                                                                                                                                                                                                                                                                                                    ga.b = new String[gb.p];
                                                                                                                                                                                                                                                                                                    gb.s = new int[gb.p];
                                                                                                                                                                                                                                                                                                    for (var3_4 = 0; gb.p > var3_4; ++var3_4) {
                                                                                                                                                                                                                                                                                                        ac.x[var3_4] = o.a((byte)38);
                                                                                                                                                                                                                                                                                                        if (!var6_3) {
                                                                                                                                                                                                                                                                                                            if (!var6_3) continue;
                                                                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                                                                        break block282;
                                                                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                                                                    var3_4 = 0;
                                                                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                                                                while (~var3_4 > ~gb.p) {
                                                                                                                                                                                                                                                                                                    ga.b[var3_4] = o.a((byte)38);
                                                                                                                                                                                                                                                                                                    ++var3_4;
                                                                                                                                                                                                                                                                                                    if (!var6_3) {
                                                                                                                                                                                                                                                                                                        if (!var6_3) continue;
                                                                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                                                                    break block283;
                                                                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                                                                var3_4 = 0;
                                                                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                                                                            while (gb.p > var3_4) {
                                                                                                                                                                                                                                                                                                lb.ac[var3_4] = o.a((byte)38);
                                                                                                                                                                                                                                                                                                ++var3_4;
                                                                                                                                                                                                                                                                                                if (!var6_3) {
                                                                                                                                                                                                                                                                                                    if (!var6_3) continue;
                                                                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                                                                break block284;
                                                                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                                                                            var3_4 = 0;
                                                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                                                        while (gb.p > var3_4) {
                                                                                                                                                                                                                                                                                            ua.Bb[var3_4] = t.a(65525);
                                                                                                                                                                                                                                                                                            v0 = ~mb.l;
                                                                                                                                                                                                                                                                                            v1 = ~(ua.Bb[var3_4] + 1);
                                                                                                                                                                                                                                                                                            if (!var6_3) {
                                                                                                                                                                                                                                                                                                if (v0 > v1) {
                                                                                                                                                                                                                                                                                                    mb.l = ua.Bb[var3_4] + 1;
                                                                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                                                                ++var3_4;
                                                                                                                                                                                                                                                                                                if (!var6_3) continue;
                                                                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                                                                            ** GOTO lbl57
                                                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                                                        var3_4 = 0;
                                                                                                                                                                                                                                                                                        do {
                                                                                                                                                                                                                                                                                            v0 = ~var3_4;
                                                                                                                                                                                                                                                                                            v1 = ~gb.p;
lbl57:
                                                                                                                                                                                                                                                                                            // 2 sources

                                                                                                                                                                                                                                                                                            if (v0 <= v1) break;
                                                                                                                                                                                                                                                                                            kb.b[var3_4] = ub.a((byte)-105);
                                                                                                                                                                                                                                                                                            ++var3_4;
                                                                                                                                                                                                                                                                                            if (var6_3) break block285;
                                                                                                                                                                                                                                                                                        } while (!var6_3);
                                                                                                                                                                                                                                                                                        var3_4 = 0;
                                                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                                                    while (gb.p > var3_4) {
                                                                                                                                                                                                                                                                                        fa.e[var3_4] = v.a(-30504);
                                                                                                                                                                                                                                                                                        ++var3_4;
                                                                                                                                                                                                                                                                                        if (!var6_3) {
                                                                                                                                                                                                                                                                                            if (!var6_3) continue;
                                                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                                                        break block286;
                                                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                                                    var3_4 = 0;
                                                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                                                while (gb.p > var3_4) {
                                                                                                                                                                                                                                                                                    gb.s[var3_4] = v.a(-30504);
                                                                                                                                                                                                                                                                                    ++var3_4;
                                                                                                                                                                                                                                                                                    if (!var6_3) {
                                                                                                                                                                                                                                                                                        if (!var6_3) continue;
                                                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                                                    break block287;
                                                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                                                var3_4 = 0;
                                                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                                                            while (~gb.p < ~var3_4) {
                                                                                                                                                                                                                                                                                mb.k[var3_4] = t.a(65525);
                                                                                                                                                                                                                                                                                ++var3_4;
                                                                                                                                                                                                                                                                                if (!var6_3) {
                                                                                                                                                                                                                                                                                    if (!var6_3) continue;
                                                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                                                break block288;
                                                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                                                            var3_4 = 0;
                                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                                        while (var3_4 < gb.p) {
                                                                                                                                                                                                                                                                            h.c[var3_4] = ub.a((byte)-105);
                                                                                                                                                                                                                                                                            ++var3_4;
                                                                                                                                                                                                                                                                            if (!var6_3) {
                                                                                                                                                                                                                                                                                if (!var6_3) continue;
                                                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                                                            break block289;
                                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                                        var3_4 = 0;
                                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                                    while (gb.p > var3_4) {
                                                                                                                                                                                                                                                                        kb.c[var3_4] = v.a(-30504);
                                                                                                                                                                                                                                                                        ++var3_4;
                                                                                                                                                                                                                                                                        if (!var6_3) {
                                                                                                                                                                                                                                                                            if (!var6_3) continue;
                                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                                        break block290;
                                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                                    var3_4 = 0;
                                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                                while (gb.p > var3_4) {
                                                                                                                                                                                                                                                                    ka.c[var3_4] = v.a(-30504);
                                                                                                                                                                                                                                                                    ++var3_4;
                                                                                                                                                                                                                                                                    if (!var6_3) {
                                                                                                                                                                                                                                                                        if (!var6_3) continue;
                                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                                    break block291;
                                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                                var3_4 = 0;
                                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                                            while (var3_4 < gb.p) {
                                                                                                                                                                                                                                                                v2 = 1;
                                                                                                                                                                                                                                                                v3 = var2_2;
                                                                                                                                                                                                                                                                if (!var6_3) {
                                                                                                                                                                                                                                                                    if (v2 != v3 && -2 == ~ka.c[var3_4]) {
                                                                                                                                                                                                                                                                        ac.x[var3_4] = m.z[6];
                                                                                                                                                                                                                                                                        ga.b[var3_4] = m.z[5];
                                                                                                                                                                                                                                                                        kb.b[var3_4] = 0;
                                                                                                                                                                                                                                                                        lb.ac[var3_4] = "";
                                                                                                                                                                                                                                                                        gb.s[0] = 0;
                                                                                                                                                                                                                                                                        mb.k[var3_4] = 0;
                                                                                                                                                                                                                                                                        kb.c[var3_4] = 1;
                                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                                    ++var3_4;
                                                                                                                                                                                                                                                                    if (!var6_3) continue;
                                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                                break block292;
                                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                                            la.d = t.a(65525);
                                                                                                                                                                                                                                                            fb.d = new int[la.d];
                                                                                                                                                                                                                                                            b.h = new int[la.d];
                                                                                                                                                                                                                                                            jb.k = new int[la.d];
                                                                                                                                                                                                                                                            ob.h = new int[la.d];
                                                                                                                                                                                                                                                            la.a = new int[la.d];
                                                                                                                                                                                                                                                            m.g = new int[la.d];
                                                                                                                                                                                                                                                            v.e = new int[la.d];
                                                                                                                                                                                                                                                            o.a = new int[la.d];
                                                                                                                                                                                                                                                            ba.ac = new String[la.d];
                                                                                                                                                                                                                                                            fb.c = new int[la.d];
                                                                                                                                                                                                                                                            p.e = new String[la.d];
                                                                                                                                                                                                                                                            da.T = new int[la.d];
                                                                                                                                                                                                                                                            e.Mb = new String[la.d];
                                                                                                                                                                                                                                                            na.a = new int[la.d];
                                                                                                                                                                                                                                                            db.j = new int[la.d];
                                                                                                                                                                                                                                                            v2 = la.d;
                                                                                                                                                                                                                                                            v3 = 12;
                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                        qb.d = new int[v2][v3];
                                                                                                                                                                                                                                                        eb.b = new int[la.d];
                                                                                                                                                                                                                                                        ua.Ab = new int[la.d];
                                                                                                                                                                                                                                                        var3_4 = 0;
                                                                                                                                                                                                                                                        while (~la.d < ~var3_4) {
                                                                                                                                                                                                                                                            e.Mb[var3_4] = o.a((byte)38);
                                                                                                                                                                                                                                                            ++var3_4;
                                                                                                                                                                                                                                                            if (!var6_3) {
                                                                                                                                                                                                                                                                if (!var6_3) continue;
                                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                                            break block293;
                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                        var3_4 = 0;
                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                    while (~var3_4 > ~la.d) {
                                                                                                                                                                                                                                                        ba.ac[var3_4] = o.a((byte)38);
                                                                                                                                                                                                                                                        ++var3_4;
                                                                                                                                                                                                                                                        if (!var6_3) {
                                                                                                                                                                                                                                                            if (!var6_3) continue;
                                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                                        break block294;
                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                    var3_4 = 0;
                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                while (la.d > var3_4) {
                                                                                                                                                                                                                                                    la.a[var3_4] = v.a(-30504);
                                                                                                                                                                                                                                                    ++var3_4;
                                                                                                                                                                                                                                                    if (!var6_3) {
                                                                                                                                                                                                                                                        if (!var6_3) continue;
                                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                                    break block295;
                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                var3_4 = 0;
                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                            while (~la.d < ~var3_4) {
                                                                                                                                                                                                                                                eb.b[var3_4] = v.a(-30504);
                                                                                                                                                                                                                                                ++var3_4;
                                                                                                                                                                                                                                                if (!var6_3) {
                                                                                                                                                                                                                                                    if (!var6_3) continue;
                                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                                break block296;
                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                            var3_4 = 0;
                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                        while (~la.d < ~var3_4) {
                                                                                                                                                                                                                                            fb.d[var3_4] = v.a(-30504);
                                                                                                                                                                                                                                            ++var3_4;
                                                                                                                                                                                                                                            if (!var6_3) {
                                                                                                                                                                                                                                                if (!var6_3) continue;
                                                                                                                                                                                                                                            }
                                                                                                                                                                                                                                            break block297;
                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                        var3_4 = 0;
                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                    while (~var3_4 > ~la.d) {
                                                                                                                                                                                                                                        jb.k[var3_4] = v.a(-30504);
                                                                                                                                                                                                                                        ++var3_4;
                                                                                                                                                                                                                                        if (!var6_3) {
                                                                                                                                                                                                                                            if (!var6_3) continue;
                                                                                                                                                                                                                                        }
                                                                                                                                                                                                                                        break block298;
                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                    var3_4 = 0;
                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                while (la.d > var3_4) {
                                                                                                                                                                                                                                    o.a[var3_4] = v.a(-30504);
                                                                                                                                                                                                                                    ++var3_4;
                                                                                                                                                                                                                                    if (!var6_3) {
                                                                                                                                                                                                                                        if (!var6_3) continue;
                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                    break block299;
                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                var3_4 = 0;
                                                                                                                                                                                                                            }
                                                                                                                                                                                                                            block152: while (true) {
                                                                                                                                                                                                                                v4 = ~la.d;
                                                                                                                                                                                                                                v5 = ~var3_4;
                                                                                                                                                                                                                                block153: while (v4 < v5) {
                                                                                                                                                                                                                                    v6 = 0;
                                                                                                                                                                                                                                    if (var6_3) break block300;
                                                                                                                                                                                                                                    for (var4_6 = v175411; var4_6 < 12; ++var4_6) {
                                                                                                                                                                                                                                        qb.d[var3_4][var4_6] = v.a(-30504);
                                                                                                                                                                                                                                        v4 = qb.d[var3_4][var4_6];
                                                                                                                                                                                                                                        v5 = 255;
                                                                                                                                                                                                                                        if (var6_3) continue block153;
                                                                                                                                                                                                                                        if (v4 != v5) continue;
                                                                                                                                                                                                                                        qb.d[var3_4][var4_6] = -1;
                                                                                                                                                                                                                                        if (!var6_3) continue;
                                                                                                                                                                                                                                    }
                                                                                                                                                                                                                                    ++var3_4;
                                                                                                                                                                                                                                    if (!var6_3) continue block152;
                                                                                                                                                                                                                                }
                                                                                                                                                                                                                                break;
                                                                                                                                                                                                                            }
                                                                                                                                                                                                                            v6 = var3_4 = 0;
                                                                                                                                                                                                                        }
                                                                                                                                                                                                                        while (var3_4 < la.d) {
                                                                                                                                                                                                                            da.T[var3_4] = ub.a((byte)-105);
                                                                                                                                                                                                                            ++var3_4;
                                                                                                                                                                                                                            if (!var6_3) {
                                                                                                                                                                                                                                if (!var6_3) continue;
                                                                                                                                                                                                                            }
                                                                                                                                                                                                                            break block301;
                                                                                                                                                                                                                        }
                                                                                                                                                                                                                        var3_4 = 0;
                                                                                                                                                                                                                    }
                                                                                                                                                                                                                    while (~var3_4 > ~la.d) {
                                                                                                                                                                                                                        m.g[var3_4] = ub.a((byte)-105);
                                                                                                                                                                                                                        ++var3_4;
                                                                                                                                                                                                                        if (!var6_3) {
                                                                                                                                                                                                                            if (!var6_3) continue;
                                                                                                                                                                                                                        }
                                                                                                                                                                                                                        break block302;
                                                                                                                                                                                                                    }
                                                                                                                                                                                                                    var3_4 = 0;
                                                                                                                                                                                                                }
                                                                                                                                                                                                                while (~var3_4 > ~la.d) {
                                                                                                                                                                                                                    ua.Ab[var3_4] = ub.a((byte)-105);
                                                                                                                                                                                                                    ++var3_4;
                                                                                                                                                                                                                    if (!var6_3) {
                                                                                                                                                                                                                        if (!var6_3) continue;
                                                                                                                                                                                                                    }
                                                                                                                                                                                                                    break block303;
                                                                                                                                                                                                                }
                                                                                                                                                                                                                if (var1_1 < 10) {
                                                                                                                                                                                                                    return;
                                                                                                                                                                                                                }
                                                                                                                                                                                                            }
                                                                                                                                                                                                            for (var3_4 = 0; la.d > var3_4; ++var3_4) {
                                                                                                                                                                                                                v.e[var3_4] = ub.a((byte)-105);
                                                                                                                                                                                                                if (!var6_3) {
                                                                                                                                                                                                                    if (!var6_3) continue;
                                                                                                                                                                                                                }
                                                                                                                                                                                                                break block304;
                                                                                                                                                                                                            }
                                                                                                                                                                                                            var3_4 = 0;
                                                                                                                                                                                                        }
                                                                                                                                                                                                        while (~var3_4 > ~la.d) {
                                                                                                                                                                                                            fb.c[var3_4] = t.a(65525);
                                                                                                                                                                                                            ++var3_4;
                                                                                                                                                                                                            if (!var6_3) {
                                                                                                                                                                                                                if (!var6_3) continue;
                                                                                                                                                                                                            }
                                                                                                                                                                                                            break block305;
                                                                                                                                                                                                        }
                                                                                                                                                                                                        var3_4 = 0;
                                                                                                                                                                                                    }
                                                                                                                                                                                                    while (~var3_4 > ~la.d) {
                                                                                                                                                                                                        b.h[var3_4] = t.a(65525);
                                                                                                                                                                                                        ++var3_4;
                                                                                                                                                                                                        if (!var6_3) {
                                                                                                                                                                                                            if (!var6_3) continue;
                                                                                                                                                                                                        }
                                                                                                                                                                                                        break block306;
                                                                                                                                                                                                    }
                                                                                                                                                                                                    var3_4 = 0;
                                                                                                                                                                                                }
                                                                                                                                                                                                while (~var3_4 > ~la.d) {
                                                                                                                                                                                                    ob.h[var3_4] = v.a(-30504);
                                                                                                                                                                                                    ++var3_4;
                                                                                                                                                                                                    if (!var6_3) {
                                                                                                                                                                                                        if (!var6_3) continue;
                                                                                                                                                                                                    }
                                                                                                                                                                                                    break block307;
                                                                                                                                                                                                }
                                                                                                                                                                                                var3_4 = 0;
                                                                                                                                                                                            }
                                                                                                                                                                                            while (~var3_4 > ~la.d) {
                                                                                                                                                                                                na.a[var3_4] = v.a(-30504);
                                                                                                                                                                                                ++var3_4;
                                                                                                                                                                                                if (!var6_3) {
                                                                                                                                                                                                    if (!var6_3) continue;
                                                                                                                                                                                                }
                                                                                                                                                                                                break block308;
                                                                                                                                                                                            }
                                                                                                                                                                                            var3_4 = 0;
                                                                                                                                                                                        }
                                                                                                                                                                                        while (var3_4 < la.d) {
                                                                                                                                                                                            db.j[var3_4] = v.a(-30504);
                                                                                                                                                                                            ++var3_4;
                                                                                                                                                                                            if (!var6_3) {
                                                                                                                                                                                                if (!var6_3) continue;
                                                                                                                                                                                            }
                                                                                                                                                                                            break block309;
                                                                                                                                                                                        }
                                                                                                                                                                                        var3_4 = 0;
                                                                                                                                                                                    }
                                                                                                                                                                                    while (var3_4 < la.d) {
                                                                                                                                                                                        p.e[var3_4] = o.a((byte)38);
                                                                                                                                                                                        ++var3_4;
                                                                                                                                                                                        if (!var6_3) {
                                                                                                                                                                                            if (!var6_3) continue;
                                                                                                                                                                                        }
                                                                                                                                                                                        break block310;
                                                                                                                                                                                    }
                                                                                                                                                                                    jb.o = t.a(65525);
                                                                                                                                                                                    p.c = new String[jb.o];
                                                                                                                                                                                    mb.g = new String[jb.o];
                                                                                                                                                                                }
                                                                                                                                                                                for (var3_4 = 0; var3_4 < jb.o; ++var3_4) {
                                                                                                                                                                                    mb.g[var3_4] = o.a((byte)38);
                                                                                                                                                                                    if (!var6_3) {
                                                                                                                                                                                        if (!var6_3) continue;
                                                                                                                                                                                    }
                                                                                                                                                                                    break block311;
                                                                                                                                                                                }
                                                                                                                                                                                var3_4 = 0;
                                                                                                                                                                            }
                                                                                                                                                                            while (jb.o > var3_4) {
                                                                                                                                                                                p.c[var3_4] = o.a((byte)38);
                                                                                                                                                                                ++var3_4;
                                                                                                                                                                                if (!var6_3) {
                                                                                                                                                                                    if (!var6_3) continue;
                                                                                                                                                                                }
                                                                                                                                                                                break block312;
                                                                                                                                                                            }
                                                                                                                                                                            na.e = t.a(65525);
                                                                                                                                                                            aa.c = new int[na.e];
                                                                                                                                                                            cb.e = new String[na.e];
                                                                                                                                                                            nb.d = new int[na.e];
                                                                                                                                                                            n.m = new int[na.e];
                                                                                                                                                                            w.g = new int[na.e];
                                                                                                                                                                            db.l = new int[na.e];
                                                                                                                                                                        }
                                                                                                                                                                        for (var3_4 = 0; var3_4 < na.e; ++var3_4) {
                                                                                                                                                                            cb.e[var3_4] = o.a((byte)38);
                                                                                                                                                                            if (!var6_3) {
                                                                                                                                                                                if (!var6_3) continue;
                                                                                                                                                                            }
                                                                                                                                                                            break block313;
                                                                                                                                                                        }
                                                                                                                                                                        var3_4 = 0;
                                                                                                                                                                    }
                                                                                                                                                                    while (~var3_4 > ~na.e) {
                                                                                                                                                                        db.l[var3_4] = ub.a((byte)-105);
                                                                                                                                                                        ++var3_4;
                                                                                                                                                                        if (!var6_3) {
                                                                                                                                                                            if (!var6_3) continue;
                                                                                                                                                                        }
                                                                                                                                                                        break block314;
                                                                                                                                                                    }
                                                                                                                                                                    var3_4 = 0;
                                                                                                                                                                }
                                                                                                                                                                while (~na.e < ~var3_4) {
                                                                                                                                                                    n.m[var3_4] = v.a(-30504);
                                                                                                                                                                    ++var3_4;
                                                                                                                                                                    if (!var6_3) {
                                                                                                                                                                        if (!var6_3) continue;
                                                                                                                                                                    }
                                                                                                                                                                    break block315;
                                                                                                                                                                }
                                                                                                                                                                var3_4 = 0;
                                                                                                                                                            }
                                                                                                                                                            while (var3_4 < na.e) {
                                                                                                                                                                nb.d[var3_4] = v.a(-30504);
                                                                                                                                                                ++var3_4;
                                                                                                                                                                if (!var6_3) {
                                                                                                                                                                    if (!var6_3) continue;
                                                                                                                                                                }
                                                                                                                                                                break block316;
                                                                                                                                                            }
                                                                                                                                                            var3_4 = 0;
                                                                                                                                                        }
                                                                                                                                                        while (~var3_4 > ~na.e) {
                                                                                                                                                            aa.c[var3_4] = v.a(-30504);
                                                                                                                                                            ++var3_4;
                                                                                                                                                            if (!var6_3) {
                                                                                                                                                                if (!var6_3) continue;
                                                                                                                                                            }
                                                                                                                                                            break block317;
                                                                                                                                                        }
                                                                                                                                                        var3_4 = 0;
                                                                                                                                                    }
                                                                                                                                                    while (na.e > var3_4) {
                                                                                                                                                        w.g[var3_4] = v.a(-30504);
                                                                                                                                                        ++var3_4;
                                                                                                                                                        if (!var6_3) {
                                                                                                                                                            if (!var6_3) continue;
                                                                                                                                                        }
                                                                                                                                                        break block318;
                                                                                                                                                    }
                                                                                                                                                    ua.Db = t.a(65525);
                                                                                                                                                    mb.a = new int[ua.Db];
                                                                                                                                                    ub.g = new int[ua.Db];
                                                                                                                                                    la.f = new String[ua.Db];
                                                                                                                                                    l.a = new String[ua.Db];
                                                                                                                                                    f.f = new int[ua.Db];
                                                                                                                                                    p.a = new String[ua.Db];
                                                                                                                                                    s.f = new String[ua.Db];
                                                                                                                                                    fb.f = new int[ua.Db];
                                                                                                                                                    h.b = new int[ua.Db];
                                                                                                                                                }
                                                                                                                                                var3_4 = 0;
                                                                                                                                                while (~ua.Db < ~var3_4) {
                                                                                                                                                    l.a[var3_4] = o.a((byte)38);
                                                                                                                                                    ++var3_4;
                                                                                                                                                    if (!var6_3) {
                                                                                                                                                        if (!var6_3) continue;
                                                                                                                                                    }
                                                                                                                                                    break block319;
                                                                                                                                                }
                                                                                                                                                var3_4 = 0;
                                                                                                                                            }
                                                                                                                                            while (var3_4 < ua.Db) {
                                                                                                                                                la.f[var3_4] = o.a((byte)38);
                                                                                                                                                ++var3_4;
                                                                                                                                                if (!var6_3) {
                                                                                                                                                    if (!var6_3) continue;
                                                                                                                                                }
                                                                                                                                                break block320;
                                                                                                                                            }
                                                                                                                                            var3_4 = 0;
                                                                                                                                        }
                                                                                                                                        while (~var3_4 > ~ua.Db) {
                                                                                                                                            s.f[var3_4] = o.a((byte)38);
                                                                                                                                            ++var3_4;
                                                                                                                                            if (!var6_3) {
                                                                                                                                                if (!var6_3) continue;
                                                                                                                                            }
                                                                                                                                            break block321;
                                                                                                                                        }
                                                                                                                                        var3_4 = 0;
                                                                                                                                    }
                                                                                                                                    while (~ua.Db < ~var3_4) {
                                                                                                                                        p.a[var3_4] = o.a((byte)38);
                                                                                                                                        ++var3_4;
                                                                                                                                        if (!var6_3) {
                                                                                                                                            if (!var6_3) continue;
                                                                                                                                        }
                                                                                                                                        break block322;
                                                                                                                                    }
                                                                                                                                    var3_4 = 0;
                                                                                                                                }
                                                                                                                                while (var3_4 < ua.Db) {
                                                                                                                                    fb.f[var3_4] = ca.a((byte)91, o.a((byte)38));
                                                                                                                                    ++var3_4;
                                                                                                                                    if (!var6_3) {
                                                                                                                                        if (!var6_3) continue;
                                                                                                                                    }
                                                                                                                                    break block323;
                                                                                                                                }
                                                                                                                                var3_4 = 0;
                                                                                                                            }
                                                                                                                            while (~var3_4 > ~ua.Db) {
                                                                                                                                f.f[var3_4] = v.a(-30504);
                                                                                                                                ++var3_4;
                                                                                                                                if (!var6_3) {
                                                                                                                                    if (!var6_3) continue;
                                                                                                                                }
                                                                                                                                break block324;
                                                                                                                            }
                                                                                                                            var3_4 = 0;
                                                                                                                        }
                                                                                                                        while (~ua.Db < ~var3_4) {
                                                                                                                            ub.g[var3_4] = v.a(-30504);
                                                                                                                            ++var3_4;
                                                                                                                            if (!var6_3) {
                                                                                                                                if (!var6_3) continue;
                                                                                                                            }
                                                                                                                            break block325;
                                                                                                                        }
                                                                                                                        var3_4 = 0;
                                                                                                                    }
                                                                                                                    while (~ua.Db < ~var3_4) {
                                                                                                                        mb.a[var3_4] = v.a(-30504);
                                                                                                                        ++var3_4;
                                                                                                                        if (!var6_3) {
                                                                                                                            if (!var6_3) continue;
                                                                                                                        }
                                                                                                                        break block326;
                                                                                                                    }
                                                                                                                    var3_4 = 0;
                                                                                                                }
                                                                                                                while (~var3_4 > ~ua.Db) {
                                                                                                                    h.b[var3_4] = v.a(-30504);
                                                                                                                    ++var3_4;
                                                                                                                    if (!var6_3) {
                                                                                                                        if (!var6_3) continue;
                                                                                                                    }
                                                                                                                    break block327;
                                                                                                                }
                                                                                                                h.a = t.a(65525);
                                                                                                                u.b = new String[h.a];
                                                                                                                u.a = new int[h.a];
                                                                                                                v.a = new int[h.a];
                                                                                                                f.e = new String[h.a];
                                                                                                                client.Jk = new int[h.a];
                                                                                                                lb.Tb = new int[h.a];
                                                                                                                ub.b = new String[h.a];
                                                                                                                ta.r = new String[h.a];
                                                                                                                ib.d = new int[h.a];
                                                                                                            }
                                                                                                            for (var3_4 = 0; h.a > var3_4; ++var3_4) {
                                                                                                                ta.r[var3_4] = o.a((byte)38);
                                                                                                                if (!var6_3) {
                                                                                                                    if (!var6_3) continue;
                                                                                                                }
                                                                                                                break block328;
                                                                                                            }
                                                                                                            var3_4 = 0;
                                                                                                        }
                                                                                                        while (~var3_4 > ~h.a) {
                                                                                                            ub.b[var3_4] = o.a((byte)38);
                                                                                                            ++var3_4;
                                                                                                            if (!var6_3) {
                                                                                                                if (!var6_3) continue;
                                                                                                            }
                                                                                                            break block329;
                                                                                                        }
                                                                                                        var3_4 = 0;
                                                                                                    }
                                                                                                    while (~h.a < ~var3_4) {
                                                                                                        u.b[var3_4] = o.a((byte)38);
                                                                                                        ++var3_4;
                                                                                                        if (!var6_3) {
                                                                                                            if (!var6_3) continue;
                                                                                                        }
                                                                                                        break block330;
                                                                                                    }
                                                                                                    var3_4 = 0;
                                                                                                }
                                                                                                while (~h.a < ~var3_4) {
                                                                                                    f.e[var3_4] = o.a((byte)38);
                                                                                                    ++var3_4;
                                                                                                    if (!var6_3) {
                                                                                                        if (!var6_3) continue;
                                                                                                    }
                                                                                                    break block331;
                                                                                                }
                                                                                                var3_4 = 0;
                                                                                            }
                                                                                            while (var3_4 < h.a) {
                                                                                                ib.d[var3_4] = t.a(65525);
                                                                                                ++var3_4;
                                                                                                if (!var6_3) {
                                                                                                    if (!var6_3) continue;
                                                                                                }
                                                                                                break block332;
                                                                                            }
                                                                                            var3_4 = 0;
                                                                                        }
                                                                                        while (~h.a < ~var3_4) {
                                                                                            v.a[var3_4] = ub.a((byte)-105);
                                                                                            ++var3_4;
                                                                                            if (!var6_3) {
                                                                                                if (!var6_3) continue;
                                                                                            }
                                                                                            break block333;
                                                                                        }
                                                                                        var3_4 = 0;
                                                                                    }
                                                                                    while (var3_4 < h.a) {
                                                                                        client.Jk[var3_4] = ub.a((byte)-105);
                                                                                        ++var3_4;
                                                                                        if (!var6_3) {
                                                                                            if (!var6_3) continue;
                                                                                        }
                                                                                        break block334;
                                                                                    }
                                                                                    var3_4 = 0;
                                                                                }
                                                                                while (~h.a < ~var3_4) {
                                                                                    u.a[var3_4] = v.a(-30504);
                                                                                    ++var3_4;
                                                                                    if (!var6_3) {
                                                                                        if (!var6_3) continue;
                                                                                    }
                                                                                    break block335;
                                                                                }
                                                                                var3_4 = 0;
                                                                            }
                                                                            while (~var3_4 > ~h.a) {
                                                                                lb.Tb[var3_4] = v.a(-30504);
                                                                                ++var3_4;
                                                                                if (!var6_3) {
                                                                                    if (!var6_3) continue;
                                                                                }
                                                                                break block336;
                                                                            }
                                                                            v.h = t.a(65525);
                                                                            d.g = new int[v.h];
                                                                            i.g = new int[v.h];
                                                                        }
                                                                        var3_4 = 0;
                                                                        while (~var3_4 > ~v.h) {
                                                                            i.g[var3_4] = v.a(-30504);
                                                                            ++var3_4;
                                                                            if (!var6_3) {
                                                                                if (!var6_3) continue;
                                                                            }
                                                                            break block337;
                                                                        }
                                                                        var3_4 = 0;
                                                                    }
                                                                    while (~v.h < ~var3_4) {
                                                                        d.g[var3_4] = v.a(-30504);
                                                                        ++var3_4;
                                                                        if (!var6_3) {
                                                                            if (!var6_3) continue;
                                                                        }
                                                                        break block338;
                                                                    }
                                                                    ia.h = t.a(65525);
                                                                    da.N = new int[ia.h];
                                                                    ac.l = new int[ia.h];
                                                                    qa.K = new int[ia.h];
                                                                }
                                                                for (var3_4 = 0; ia.h > var3_4; ++var3_4) {
                                                                    qa.K[var3_4] = ub.a((byte)-105);
                                                                    if (!var6_3) {
                                                                        if (!var6_3) continue;
                                                                    }
                                                                    break block339;
                                                                }
                                                                var3_4 = 0;
                                                            }
                                                            while (~var3_4 > ~ia.h) {
                                                                da.N[var3_4] = v.a(-30504);
                                                                ++var3_4;
                                                                if (!var6_3) {
                                                                    if (!var6_3) continue;
                                                                }
                                                                break block340;
                                                            }
                                                            var3_4 = 0;
                                                        }
                                                        while (var3_4 < ia.h) {
                                                            ac.l[var3_4] = v.a(-30504);
                                                            ++var3_4;
                                                            if (!var6_3) {
                                                                if (!var6_3) continue;
                                                            }
                                                            break block341;
                                                        }
                                                        n.c = t.a(65525);
                                                        fa.b = t.a(65525);
                                                        oa.d = new int[fa.b][];
                                                        qb.e = new int[fa.b];
                                                        da.J = new int[fa.b][];
                                                        ja.L = new String[fa.b];
                                                        oa.a = new String[fa.b];
                                                        pa.f = new int[fa.b];
                                                        o.p = new int[fa.b];
                                                    }
                                                    var3_4 = 0;
                                                    while (~var3_4 > ~fa.b) {
                                                        ja.L[var3_4] = o.a((byte)38);
                                                        ++var3_4;
                                                        if (!var6_3) {
                                                            if (!var6_3) continue;
                                                        }
                                                        break block342;
                                                    }
                                                    var3_4 = 0;
                                                }
                                                while (fa.b > var3_4) {
                                                    oa.a[var3_4] = o.a((byte)38);
                                                    ++var3_4;
                                                    if (!var6_3) {
                                                        if (!var6_3) continue;
                                                    }
                                                    break block343;
                                                }
                                                var3_4 = 0;
                                            }
                                            while (fa.b > var3_4) {
                                                pa.f[var3_4] = v.a(-30504);
                                                ++var3_4;
                                                if (!var6_3) {
                                                    if (!var6_3) continue;
                                                }
                                                break block344;
                                            }
                                            var3_4 = 0;
                                        }
                                        while (var3_4 < fa.b) {
                                            o.p[var3_4] = v.a(-30504);
                                            ++var3_4;
                                            if (!var6_3) {
                                                if (!var6_3) continue;
                                            }
                                            break block345;
                                        }
                                        var3_4 = 0;
                                    }
                                    while (~fa.b < ~var3_4) {
                                        qb.e[var3_4] = v.a(-30504);
                                        ++var3_4;
                                        if (!var6_3) {
                                            if (!var6_3) continue;
                                        }
                                        break block346;
                                    }
                                    var3_4 = 0;
                                }
                                while (var3_4 < fa.b) {
                                    block348: {
                                        var4_6 = v.a(-30504);
                                        oa.d[var3_4] = new int[var4_6];
                                        v7 = 0;
                                        if (var6_3) break block347;
                                        var5_7 = v7;
                                        while (~var4_6 < ~var5_7) {
                                            oa.d[var3_4][var5_7] = t.a(65525);
                                            ++var5_7;
                                            if (!var6_3) {
                                                if (!var6_3) continue;
                                            }
                                            break block348;
                                        }
                                        ++var3_4;
                                    }
                                    if (!var6_3) continue;
                                }
                                v7 = var3_4 = 0;
                            }
                            while (var3_4 < fa.b) {
                                block350: {
                                    var4_6 = v.a(-30504);
                                    da.J[var3_4] = new int[var4_6];
                                    v8 = 0;
                                    if (var6_3) break block349;
                                    var5_7 = v8;
                                    while (~var5_7 > ~var4_6) {
                                        da.J[var3_4][var5_7] = v.a(-30504);
                                        ++var5_7;
                                        if (!var6_3) {
                                            if (!var6_3) continue;
                                        }
                                        break block350;
                                    }
                                    ++var3_4;
                                }
                                if (!var6_3) continue;
                            }
                            t.g = t.a(65525);
                            t.h = new String[t.g];
                            ca.B = new int[t.g];
                            fa.c = new int[t.g];
                            h.e = new String[t.g];
                            v8 = var3_4 = 0;
                        }
                        while (var3_4 < t.g) {
                            t.h[var3_4] = o.a((byte)38);
                            ++var3_4;
                            if (!var6_3) {
                                if (!var6_3) continue;
                            }
                            break block351;
                        }
                        var3_4 = 0;
                    }
                    while (var3_4 < t.g) {
                        h.e[var3_4] = o.a((byte)38);
                        ++var3_4;
                        if (!var6_3) {
                            if (!var6_3) continue;
                        }
                        break block352;
                    }
                    var3_4 = 0;
                }
                while (t.g > var3_4) {
                    ca.B[var3_4] = v.a(-30504);
                    ++var3_4;
                    if (!var6_3) {
                        if (!var6_3) continue;
                    }
                    break block353;
                }
                var3_4 = 0;
            }
            while (t.g > var3_4) {
                fa.c[var3_4] = v.a(-30504);
                ++var3_4;
                if (var6_3 != false) return;
                if (!var6_3) continue;
            }
            b.v = null;
            kb.d = null;
            return;
        }
        catch (RuntimeException var3_5) {
            v9 = new StringBuilder().append(m.z[3]);
            if (var0 != null) {
                v10 = m.z[1];
                throw i.a(var3_5, v9.append(v10).append(',').append(var1_1).append(',').append(var2_2).append(')').toString());
            }
            v10 = m.z[0];
            throw i.a(var3_5, v9.append(v10).append(',').append(var1_1).append(',').append(var2_2).append(')').toString());
        }
    }

    m() {
    }

    abstract Socket a(byte var1) throws IOException;

    static {
        z = new String[]{m.z(m.z(".aKU")), m.z(m.z(";:\t\u0017#")), m.z(m.z("-:f\u0011")), m.z(m.z("-:e\u0011")), m.z(m.z(")zS\\9%f\t]?4")), m.z(m.z("\u0019{R\u00190%qC\u0019*/4E\\~!4J\\3\"qU\u0019*/4RJ;``OP-`{ES;#`")), m.z(m.z("\rqJ[;2g\u0007V<*qDM")), m.z(m.z("3`UP0':CX*")), m.z(m.z("-:d\u0011"))};
        b = new byte[50][];
        i = new int[256];
        e = null;
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x5E);
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
                        n4 = 64;
                        break;
                    }
                    case 1: {
                        n4 = 20;
                        break;
                    }
                    case 2: {
                        n4 = 39;
                        break;
                    }
                    case 3: {
                        n4 = 57;
                        break;
                    }
                    default: {
                        n4 = 94;
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

