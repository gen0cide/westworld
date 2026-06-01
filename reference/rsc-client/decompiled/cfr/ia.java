/*
 * Decompiled with CFR 0.152.
 */
final class ia {
    static int e;
    static int b;
    static String[] a;
    static int[] d;
    static int h;
    static int f;
    static int i;
    static String[] g;
    static int c;
    private static final String[] z;

    static final void a(int n2, int n3, int[] nArray, int n4, int n5, int[] nArray2, int n6, byte by) {
        block15: {
            boolean bl = client.vh;
            try {
                int n7;
                int n8;
                block14: {
                    ++e;
                    if (0 <= n6) {
                        return;
                    }
                    n3 = nArray[(0xFFB9 & n4) >> -1668975128];
                    n4 += (n2 <<= 1);
                    n8 = n6 / 8;
                    int n9 = 69 % ((by - 27) / 45);
                    for (n7 = n8; 0 > n7; ++n7) {
                        nArray2[n5++] = n3;
                        nArray2[n5++] = n3;
                        n3 = nArray[(0xFFDF & n4) >> 2132179464];
                        nArray2[n5++] = n3;
                        nArray2[n5++] = n3;
                        n3 = nArray[((n4 += n2) & 0xFF6F) >> -1202367672];
                        nArray2[n5++] = n3;
                        nArray2[n5++] = n3;
                        n3 = nArray[(0xFF16 & (n4 += n2)) >> 1020025064];
                        nArray2[n5++] = n3;
                        nArray2[n5++] = n3;
                        n3 = nArray[((n4 += n2) & 0xFF6A) >> -31505560];
                        n4 += n2;
                        if (!bl) {
                            if (!bl) continue;
                            break;
                        }
                        break block14;
                    }
                    n8 = -(n6 % 8);
                }
                n7 = 0;
                while (~n7 > ~n8) {
                    block16: {
                        nArray2[n5++] = n3;
                        if (bl) break block15;
                        if (1 != (n7 & 1)) break block16;
                        n3 = nArray[n4 >> 867708808 & 0xFF];
                        n4 += n2;
                    }
                    ++n7;
                    if (!bl) continue;
                    break;
                }
            }
            catch (RuntimeException runtimeException) {
                RuntimeException runtimeException2 = runtimeException;
                StringBuilder stringBuilder = new StringBuilder().append(z[2]).append(n2).append(',').append(n3).append(',');
                String string = nArray != null ? z[0] : z[1];
                StringBuilder stringBuilder2 = stringBuilder.append(string).append(',').append(n4).append(',').append(n5).append(',');
                String string2 = nArray2 != null ? z[0] : z[1];
                throw i.a(runtimeException2, stringBuilder2.append(string2).append(',').append(n6).append(',').append(by).append(')').toString());
            }
        }
    }

    static final String a(tb tb2, boolean bl) {
        try {
            if (bl) {
                a = null;
            }
            ++c;
            return client.a(0, tb2, Short.MAX_VALUE);
        }
        catch (RuntimeException runtimeException) {
            RuntimeException runtimeException2 = runtimeException;
            StringBuilder stringBuilder = new StringBuilder().append(z[3]);
            String string = tb2 != null ? z[0] : z[1];
            throw i.a(runtimeException2, stringBuilder.append(string).append(',').append(bl).append(')').toString());
        }
    }

    /*
     * Enabled aggressive block sorting
     * Enabled unnecessary exception pruning
     * Enabled aggressive exception aggregation
     */
    static final boolean a(v v2, byte by) {
        try {
            int n2 = 3 / ((-39 - by) / 54);
            ++f;
            if (da.O == v2) return true;
            if (ga.c == v2) return true;
            if (ta.f == v2) return true;
            if (v2 == eb.d) return true;
            if (gb.n == v2) return true;
            return false;
        }
        catch (RuntimeException runtimeException) {
            String string;
            StringBuilder stringBuilder = new StringBuilder().append(z[4]);
            if (v2 != null) {
                string = z[0];
                throw i.a(runtimeException, stringBuilder.append(string).append(',').append(by).append(')').toString());
            }
            string = z[1];
            throw i.a(runtimeException, stringBuilder.append(string).append(',').append(by).append(')').toString());
        }
    }

    static {
        z = new String[]{ia.z(ia.z("3Wd\\*")), ia.z(ia.z("&\f&\u001e")), ia.z(ia.z("!\u0018d0\u007f")), ia.z(ia.z("!\u0018d1\u007f")), ia.z(ia.z("!\u0018d3\u007f"))};
        b = 0;
        d = new int[256];
        a = new String[100];
        i = 0;
        g = new String[100];
    }

    private static char[] z(String string) {
        char[] cArray = string.toCharArray();
        if (cArray.length < 2) {
            cArray = cArray;
            cArray[0] = (char)(cArray[0] ^ 0x57);
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
                        n4 = 72;
                        break;
                    }
                    case 1: {
                        n4 = 121;
                        break;
                    }
                    case 2: {
                        n4 = 74;
                        break;
                    }
                    case 3: {
                        n4 = 114;
                        break;
                    }
                    default: {
                        n4 = 87;
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

