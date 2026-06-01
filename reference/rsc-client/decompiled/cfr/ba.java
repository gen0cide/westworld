/*
 * Decompiled with CFR 0.152.
 */
import java.awt.Component;
import java.awt.event.ActionEvent;

final class ba
extends ua {
    client dc;
    static int bc;
    static int Zb;
    static int[] cc;
    static String[] Yb;
    static int ec;
    static String[] ac;
    private static final String[] fc;

    static final String e(int n2, int n3) {
        try {
            int n4 = 109 % ((n2 - -8) / 63);
            ++bc;
            return (n3 >> 2086116408 & 0xFF) + "." + (0xFF & n3 >> 968554992) + "." + (n3 >> 1312213096 & 0xFF) + "." + (n3 & 0xFF);
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, fc[5] + n2 + ',' + n3 + ')');
        }
    }

    ba(int n2, int n3, int n4, Component component) {
        super(n2, n3, n4, component);
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    @Override
    final void a(int n2, int n3, int n4, int n5, int n6, int n7, byte by, int n8) {
        boolean bl = client.vh;
        try {
            block15: {
                block12: {
                    block13: {
                        block16: {
                            block14: {
                                ++ec;
                                if (n3 >= 50000) break block12;
                                if (40000 <= n3) break block13;
                                if (n3 < 20000) break block14;
                                this.dc.a(n6, n8, 105, n4, n2, -20000 + n3, n7, n5);
                                if (!bl) break block15;
                            }
                            if (~n3 <= -5001) break block16;
                            super.f(n5, n6, n4, n7, 5924, n3);
                            if (!bl) break block15;
                        }
                        this.dc.b(n8, n7, by ^ 9, n2, n5, n6, n4, -5000 + n3);
                        if (!bl) break block15;
                    }
                    this.dc.b(n4, n8, n5, n3 - 40000, n7, -122, n6);
                    if (!bl) break block15;
                }
                this.dc.a(n8, n5, n6, n4, n3 - 50000, n7, 2);
            }
            if (by == 29) return;
            ac = null;
            return;
        }
        catch (RuntimeException runtimeException) {
            throw i.a(runtimeException, fc[4] + n2 + ',' + n3 + ',' + n4 + ',' + n5 + ',' + n6 + ',' + n7 + ',' + by + ',' + n8 + ')');
        }
    }

    /*
     * Unable to fully structure code
     */
    static final void a(Object var0, int var1_1, c var2_2) {
        var4_3 = client.vh;
        try {
            ++ba.Zb;
            if (null == var2_2.n) {
                return;
            }
            var3_4 = 0;
            while (-51 < ~var3_4) {
                v0 = var2_2.n.peekEvent();
                if (!var4_3) {
                    if (v0 == null) break;
                }
                ** GOTO lbl27
                mb.a(11200, 1L);
                ++var3_4;
                if (!var4_3) continue;
                break;
            }
            if (var1_1 != 1) {
                return;
            }
            try {
                v0 = var0;
lbl27:
                // 2 sources

                if (v0 != null) {
                    var2_2.n.postEvent(new ActionEvent(var0, 1001, ba.fc[0]));
                }
            }
            catch (Exception var3_5) {}
        }
        catch (RuntimeException var3_6) {
            v3 = var3_6;
            v4 = new StringBuilder().append(ba.fc[3]);
            v5 = var0 != null ? ba.fc[2] : ba.fc[1];
            v7 = v4.append(v5).append(',').append(var1_1).append(',');
            v8 = var2_2 != null ? ba.fc[2] : ba.fc[1];
            throw i.a(v3, v7.append(v8).append(')').toString());
        }
    }

    static {
        fc = new String[]{ba.z(ba.z("\t\u0013\u0007\u001c{")), ba.z(ba.z("\u0003\u0013\u0006\u001d")), ba.z(ba.z("\u0016HD_\u007f")), ba.z(ba.z("\u000f\u0007D0*")), ba.z(ba.z("\u000f\u0007D3*")), ba.z(ba.z("\u000f\u0007D2*"))};
        cc = new int[2048];
        Yb = new String[100];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 2);
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
                        n4 = 109;
                        break;
                    }
                    case 1: {
                        n4 = 102;
                        break;
                    }
                    case 2: {
                        n4 = 106;
                        break;
                    }
                    case 3: {
                        n4 = 113;
                        break;
                    }
                    default: {
                        n4 = 2;
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

