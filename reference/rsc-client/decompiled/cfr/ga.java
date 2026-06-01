/*
 * Decompiled with CFR 0.152.
 */
final class ga {
    static v c;
    static int d;
    static char[] a;
    static String[] b;
    private static final String[] z;

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final String a(int n2, int n3, int n4, byte[] byArray) {
        boolean bl = client.vh;
        try {
            int n5;
            int n6;
            int n7;
            int n8;
            char[] cArray;
            block13: {
                ++d;
                cArray = new char[n2];
                n8 = 0;
                n7 = 0;
                while (~n2 < ~n7) {
                    int n9 = 0xFF & byArray[n4 - -n7];
                    n6 = 0;
                    n5 = n9;
                    if (!bl) {
                        if (n6 != n5 || bl) {
                            if (~n9 <= -129 && ~n9 > -161) {
                                int n10 = nb.f[n9 - 128];
                                if (~n10 == -1) {
                                    n10 = 63;
                                }
                                n9 = n10;
                            }
                            cArray[n8++] = (char)n9;
                        }
                        ++n7;
                        if (!bl) continue;
                    }
                    break block13;
                }
                n6 = -103;
                n5 = (-63 - n3) / 49;
            }
            n7 = n6 / n5;
            return new String(cArray, 0, n8);
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(z[2]).append(n2).append(',').append(n3).append(',').append(n4).append(',');
            if (byArray != null) {
                string = z[1];
                throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
            }
            string = z[0];
            throw i.a(runtimeException, stringBuilder.append(string).append(')').toString());
        }
    }

    static {
        z = new String[]{ga.z(ga.z("<eak")), ga.z(ga.z(")>#)i")), ga.z(ga.z("5q#F<"))};
        c = new v(ga.z(ga.z("\u0005D\\F")), ga.z(ga.z("=vknw7")), ga.z(ga.z("\ral")), 2);
        a = new char[]{' ', '\u00a0', '_', '-', '\u00e0', '\u00e1', '\u00e2', '\u00e4', '\u00e3', '\u00c0', '\u00c1', '\u00c2', '\u00c4', '\u00c3', '\u00e8', '\u00e9', '\u00ea', '\u00eb', '\u00c8', '\u00c9', '\u00ca', '\u00cb', '\u00ed', '\u00ee', '\u00ef', '\u00cd', '\u00ce', '\u00cf', '\u00f2', '\u00f3', '\u00f4', '\u00f6', '\u00f5', '\u00d2', '\u00d3', '\u00d4', '\u00d6', '\u00d5', '\u00f9', '\u00fa', '\u00fb', '\u00fc', '\u00d9', '\u00da', '\u00db', '\u00dc', '\u00e7', '\u00c7', '\u00ff', '\u0178', '\u00f1', '\u00d1', '\u00df'};
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x14);
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
                        n4 = 82;
                        break;
                    }
                    case 1: {
                        n4 = 16;
                        break;
                    }
                    case 2: {
                        n4 = 13;
                        break;
                    }
                    case 3: {
                        n4 = 7;
                        break;
                    }
                    default: {
                        n4 = 20;
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

